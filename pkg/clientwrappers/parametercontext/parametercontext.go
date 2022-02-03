package parametercontext

import (
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers"
	"github.com/Orange-OpenSource/nifikop/pkg/common"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/Orange-OpenSource/nifikop/pkg/nificlient"
	"github.com/Orange-OpenSource/nifikop/pkg/util/clientconfig"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("parametercontext-method")

func ExistParameterContext(parameterContext *v1alpha1.NifiParameterContext, config *clientconfig.NifiConfig) (bool, error) {

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

func CreateParameterContext(parameterContext *v1alpha1.NifiParameterContext, parameterSecrets []*corev1.Secret,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiParameterContextStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.ParameterContextEntity{}
	updateParameterContextEntity(parameterContext, parameterSecrets, &scratchEntity)

	entity, err := nClient.CreateParameterContext(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Create parameter-context"); err != nil {
		return nil, err
	}

	parameterContext.Status.Id = entity.Id
	parameterContext.Status.Version = *entity.Revision.Version

	return &parameterContext.Status, nil
}

func SyncParameterContext(parameterContext *v1alpha1.NifiParameterContext, parameterSecrets []*corev1.Secret,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiParameterContextStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.GetParameterContext(parameterContext.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get parameter-context"); err != nil {
		return nil, err
	}

	latestUpdateRequest := parameterContext.Status.LatestUpdateRequest
	if latestUpdateRequest != nil && !latestUpdateRequest.Complete {
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
	}

	if !parameterContextIsSync(parameterContext, parameterSecrets, entity) {

		entity.Component.Parameters = updateRequestPrepare(parameterContext, parameterSecrets, entity)

		updateRequest, err := nClient.CreateParameterContextUpdateRequest(entity.Id, *entity)
		if err := clientwrappers.ErrorCreateOperation(log, err, "Create parameter-context update-request"); err != nil {
			return nil, err
		}

		parameterContext.Status.LatestUpdateRequest =
			updateRequest2Status(updateRequest)
		return &parameterContext.Status, errorfactory.NifiParameterContextUpdateRequestRunning{}
	}

	var status *v1alpha1.NifiParameterContextStatus
	if parameterContext.Status.Version != *entity.Revision.Version || parameterContext.Status.Id != entity.Id {
		status := &parameterContext.Status
		status.Version = *entity.Revision.Version
		status.Id = entity.Id
	}

	return status, nil
}

func RemoveParameterContext(parameterContext *v1alpha1.NifiParameterContext, parameterSecrets []*corev1.Secret,
	config *clientconfig.NifiConfig) error {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	entity, err := nClient.GetParameterContext(parameterContext.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get parameter-context"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil
		}
		return err
	}

	updateParameterContextEntity(parameterContext, parameterSecrets, entity)
	err = nClient.RemoveParameterContext(*entity)

	return clientwrappers.ErrorRemoveOperation(log, err, "Remove parameter-context")
}

func parameterContextIsSync(
	parameterContext *v1alpha1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	entity *nigoapi.ParameterContextEntity) bool {

	e := nigoapi.ParameterContextEntity{}
	updateParameterContextEntity(parameterContext, parameterSecrets, &e)

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
					!((expected.Parameter.Description == nil || *expected.Parameter.Description == "") &&
						((param.Parameter.Description == nil || *param.Parameter.Description == "") ||
							(expected.Parameter.Description == param.Parameter.Description))) {

					return false
				}
			}
		}
		if notFound {
			return false
		}
	}

	return e.Component.Description == entity.Component.Description && e.Component.Name == entity.Component.Name
}

func updateRequestPrepare(
	parameterContext *v1alpha1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	entity *nigoapi.ParameterContextEntity) []nigoapi.ParameterEntity {

	tmp := entity.Component.Parameters
	updateParameterContextEntity(parameterContext, parameterSecrets, entity)

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
	var parameters []nigoapi.ParameterEntity
	for _, expected := range entity.Component.Parameters {
		notFound := true
		for _, param := range tmp {
			if expected.Parameter.Name == param.Parameter.Name {
				notFound = false
				if (!param.Parameter.Sensitive &&
					!((expected.Parameter.Value == nil && param.Parameter.Value == nil) ||
						((expected.Parameter.Value != nil && param.Parameter.Value != nil) &&
							(*expected.Parameter.Value == *param.Parameter.Value)))) ||
					!((expected.Parameter.Description == nil || *expected.Parameter.Description == "") &&
						((param.Parameter.Description == nil || *param.Parameter.Description == "") ||
							(expected.Parameter.Description == param.Parameter.Description))) {
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

func updateParameterContextEntity(parameterContext *v1alpha1.NifiParameterContext, parameterSecrets []*corev1.Secret, entity *nigoapi.ParameterContextEntity) {

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
		desc := parameter.Description
		parameters = append(parameters, nigoapi.ParameterEntity{
			Parameter: &nigoapi.ParameterDto{
				Name:        parameter.Name,
				Description: &desc,
				Sensitive:   parameter.Sensitive,
				Value:       parameter.Value,
			},
		})
	}
	entity.Component.Name = parameterContext.Name
	entity.Component.Description = parameterContext.Spec.Description
	entity.Component.Parameters = parameters
}

func updateRequest2Status(updateRequest *nigoapi.ParameterContextUpdateRequestEntity) *v1alpha1.ParameterContextUpdateRequest {
	ur := updateRequest.Request
	return &v1alpha1.ParameterContextUpdateRequest{
		Id:               ur.RequestId,
		Uri:              ur.Uri,
		SubmissionTime:   ur.SubmissionTime,
		LastUpdated:      ur.LastUpdated,
		Complete:         ur.Complete,
		FailureReason:    ur.FailureReason,
		PercentCompleted: ur.PercentCompleted,
		State:            ur.State,
	}
}
