package parametercontext

import (
	"strings"

	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	corev1 "k8s.io/api/core/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("parametercontext-method")

func ExistParameterContext(parameterContext *v1.NifiParameterContext, config *clientconfig.NifiConfig) (bool, error) {
	if parameterContext.Status.Id == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetParameterContext(parameterContext.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get parameter-context"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

func FindParameterContextByName(parameterContext *v1.NifiParameterContext, config *clientconfig.NifiConfig) (*v1.NifiParameterContextStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	entities, err := nClient.GetParameterContexts()
	if err := clientwrappers.ErrorGetOperation(log, err, "Get parameter-contexts"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil, nil
		}
		return nil, err
	}

	for _, entity := range entities {
		if parameterContext.GetName() == entity.Component.Name {
			return &v1.NifiParameterContextStatus{
				Id:      entity.Id,
				Version: *entity.Revision.Version,
			}, nil
		}
	}

	return nil, nil
}

func CreateParameterContext(
	parameterContext *v1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	parameterContextRefs []*v1.NifiParameterContext,
	config *clientconfig.NifiConfig) (*v1.NifiParameterContextStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.ParameterContextEntity{}
	updateParameterContextEntity(parameterContext, parameterSecrets, parameterContextRefs, &scratchEntity)

	entity, err := nClient.CreateParameterContext(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Create parameter-context"); err != nil {
		return nil, err
	}

	parameterContext.Status.Id = entity.Id
	parameterContext.Status.Version = *entity.Revision.Version

	return &parameterContext.Status, nil
}

func SyncParameterContext(
	parameterContext *v1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	parameterContextRefs []*v1.NifiParameterContext,
	config *clientconfig.NifiConfig) (*v1.NifiParameterContextStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.GetParameterContext(parameterContext.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get parameter-context"); err != nil {
		return nil, err
	}

	latestUpdateRequest := parameterContext.Status.LatestUpdateRequest
	if latestUpdateRequest != nil && !latestUpdateRequest.Complete && !latestUpdateRequest.NotFound {
		updateRequest, err := nClient.GetParameterContextUpdateRequest(parameterContext.Status.Id, latestUpdateRequest.Id)
		if updateRequest != nil {
			parameterContext.Status.LatestUpdateRequest = updateRequest2Status(updateRequest)
		}

		if err := clientwrappers.ErrorGetOperation(log, err, "Get update-request"); err != nificlient.ErrNifiClusterReturned404 {
			if err != nil {
				return &parameterContext.Status, err
			}
			return &parameterContext.Status, errorfactory.NifiParameterContextUpdateRequestRunning{}
		}

		if err == nificlient.ErrNifiClusterReturned404 {
			parameterContext.Status.LatestUpdateRequest.NotFoundRetryCount += 1
			if parameterContext.Status.LatestUpdateRequest.NotFoundRetryCount >= 3 {
				parameterContext.Status.LatestUpdateRequest.NotFound = true
			}
			return &parameterContext.Status, errorfactory.NifiParameterContextUpdateRequestNotFound{}
		}
	}

	if !parameterContextIsSync(parameterContext, parameterSecrets, parameterContextRefs, entity) {
		entity.Component.Parameters = updateRequestPrepare(parameterContext, parameterSecrets, parameterContextRefs, entity)

		updateRequest, err := nClient.CreateParameterContextUpdateRequest(entity.Id, *entity)
		if err := clientwrappers.ErrorCreateOperation(log, err, "Create parameter-context update-request"); err != nil {
			return nil, err
		}

		parameterContext.Status.LatestUpdateRequest =
			updateRequest2Status(updateRequest)
		return &parameterContext.Status, errorfactory.NifiParameterContextUpdateRequestRunning{}
	}

	var status *v1.NifiParameterContextStatus
	if parameterContext.Status.Version != *entity.Revision.Version || parameterContext.Status.Id != entity.Id {
		status = &parameterContext.Status
		status.Version = *entity.Revision.Version
		status.Id = entity.Id
	}

	return status, nil
}

func RemoveParameterContext(
	parameterContext *v1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	parameterContextRefs []*v1.NifiParameterContext,
	config *clientconfig.NifiConfig) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	entity, err := nClient.GetParameterContext(parameterContext.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Failed to fetch parameter-context for removal: "+parameterContext.Name); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil
		}
		return err
	}

	updateParameterContextEntity(parameterContext, parameterSecrets, parameterContextRefs, entity)
	err = nClient.RemoveParameterContext(*entity)

	return clientwrappers.ErrorRemoveOperation(log, err, "Failed to remove parameter-context "+parameterContext.Name)
}

func parameterContextIsSync(
	parameterContext *v1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	parameterContextRefs []*v1.NifiParameterContext,
	entity *nigoapi.ParameterContextEntity) bool {
	e := nigoapi.ParameterContextEntity{}
	updateParameterContextEntity(parameterContext, parameterSecrets, parameterContextRefs, &e)

	if len(e.Component.Parameters) != len(entity.Component.Parameters) {
		return false
	}

	for _, expected := range e.Component.Parameters {
		notFound := true
		for _, param := range entity.Component.Parameters {
			if expected.Parameter.Name == param.Parameter.Name {
				notFound = false

				if (!param.Parameter.Sensitive &&
					!((expected.Parameter.Value == nil && param.Parameter.Value == nil) ||
						((expected.Parameter.Value != nil && param.Parameter.Value != nil) &&
							(*expected.Parameter.Value == *param.Parameter.Value)))) ||
					!((expected.Parameter.Description == nil && param.Parameter.Description == nil) ||
						((expected.Parameter.Description != nil && param.Parameter.Description != nil) &&
							(*expected.Parameter.Description == *param.Parameter.Description))) {
					return false
				}
			}
		}
		if notFound {
			return false
		}
	}

	if len(e.Component.InheritedParameterContexts) != len(entity.Component.InheritedParameterContexts) {
		return false
	}

	for idx, expected := range e.Component.InheritedParameterContexts {
		if expected.Id != entity.Component.InheritedParameterContexts[idx].Id {
			return false
		}
	}

	return e.Component.Description == entity.Component.Description && e.Component.Name == entity.Component.Name
}

func updateRequestPrepare(
	parameterContext *v1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	parameterContextRefs []*v1.NifiParameterContext,
	entity *nigoapi.ParameterContextEntity) []nigoapi.ParameterEntity {
	tmp := entity.Component.Parameters
	updateParameterContextEntity(parameterContext, parameterSecrets, parameterContextRefs, entity)

	// List all parameter to remove
	var toRemove []string
	for _, toFind := range tmp {
		notFound := true
		for _, p := range entity.Component.Parameters {
			if p.Parameter.Name == toFind.Parameter.Name {
				notFound = false
				break
			}
		}

		if notFound {
			toRemove = append(toRemove, toFind.Parameter.Name)
		}
	}

	// List all parameter to upsert
	parameters := make([]nigoapi.ParameterEntity, 0)
	for _, expected := range entity.Component.Parameters {
		notFound := true
		for _, param := range tmp {
			if expected.Parameter.Name == param.Parameter.Name {
				notFound = false
				if (!param.Parameter.Sensitive &&
					!((expected.Parameter.Value == nil && param.Parameter.Value == nil) ||
						((expected.Parameter.Value != nil && param.Parameter.Value != nil) &&
							(*expected.Parameter.Value == *param.Parameter.Value)))) ||
					!((expected.Parameter.Description == nil && param.Parameter.Description == nil) ||
						((expected.Parameter.Description != nil && param.Parameter.Description != nil) &&
							(*expected.Parameter.Description == *param.Parameter.Description))) {
					notFound = false
					if expected.Parameter.Value == nil && param.Parameter.Value != nil {
						toRemove = append(toRemove, expected.Parameter.Name)
						break
					}
					parameters = append(parameters, expected)
					break
				}
			}
		}
		if notFound {
			parameters = append(parameters, expected)
		}
	}

	for _, name := range toRemove {
		parameters = append(parameters, nigoapi.ParameterEntity{
			Parameter: &nigoapi.ParameterDto{
				Name: name,
			},
		})
	}

	return parameters
}

func updateParameterContextEntity(
	parameterContext *v1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	parameterContextRefs []*v1.NifiParameterContext,
	entity *nigoapi.ParameterContextEntity) {
	var defaultVersion int64 = 0
	if entity == nil {
		entity = &nigoapi.ParameterContextEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.ParameterContextDto{}
	}

	parameters := make([]nigoapi.ParameterEntity, 0)

	emptyString := ""
	for _, secret := range parameterSecrets {
		for k, v := range secret.Data {
			value := string(v)
			parameters = append(parameters, nigoapi.ParameterEntity{
				Parameter: &nigoapi.ParameterDto{
					Name:        k,
					Description: &emptyString,
					Sensitive:   true,
					Value:       &value,
				},
			})
		}
	}

	for _, parameter := range parameterContext.Spec.Parameters {
		desc := strings.TrimSpace(parameter.Description)
		parameters = append(parameters, nigoapi.ParameterEntity{
			Parameter: &nigoapi.ParameterDto{
				Name:        parameter.Name,
				Description: &desc,
				Sensitive:   parameter.Sensitive,
				Value:       parameter.Value,
			},
		})
	}

	inheritedParameterContexts := make([]nigoapi.ParameterContextReferenceEntity, 0)
	for _, parameterContextRef := range parameterContextRefs {
		inheritedParameterContexts = append(inheritedParameterContexts,
			nigoapi.ParameterContextReferenceEntity{
				Id: parameterContextRef.Status.Id,
				Component: &nigoapi.ParameterContextReferenceDto{
					Id:   parameterContextRef.Status.Id,
					Name: parameterContextRef.Name,
				},
			})
	}

	entity.Component.Name = parameterContext.Name
	entity.Component.Description = parameterContext.Spec.Description
	entity.Component.Parameters = parameters
	entity.Component.InheritedParameterContexts = inheritedParameterContexts
}

func updateRequest2Status(updateRequest *nigoapi.ParameterContextUpdateRequestEntity) *v1.ParameterContextUpdateRequest {
	ur := updateRequest.Request
	return &v1.ParameterContextUpdateRequest{
		Id:                 ur.RequestId,
		Uri:                ur.Uri,
		SubmissionTime:     ur.SubmissionTime,
		LastUpdated:        ur.LastUpdated,
		Complete:           ur.Complete,
		FailureReason:      ur.FailureReason,
		PercentCompleted:   ur.PercentCompleted,
		State:              ur.State,
		NotFound:           false,
		NotFoundRetryCount: 0,
	}
}
