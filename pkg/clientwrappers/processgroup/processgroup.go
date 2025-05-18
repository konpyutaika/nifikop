package processgroup

import (
	"encoding/json"
	"fmt"

	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"

	v1alpha1 "github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("processgroup-method")

func ExistProcessGroup(resource *v1alpha1.NifiResource, config *clientconfig.NifiConfig) (bool, error) {
	if resource.Status.Id == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetProcessGroup(resource.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process-group"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

func CreateProcessGroup(resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.ProcessGroupEntity{}
	updateProcessGroupEntity(resource, &scratchEntity)

	entity, err := nClient.CreateProcessGroup(scratchEntity, resource.Spec.GetParentProcessGroupID(config.RootProcessGroupId))
	if err := clientwrappers.ErrorCreateOperation(log, err, "Failed to create resource "+resource.Name); err != nil {
		return nil, err
	}

	return &v1alpha1.NifiResourceStatus{
		Id:      entity.Id,
		Version: *entity.Revision.Version,
	}, nil
}

func SyncProcessGroup(resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.GetProcessGroup(resource.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process-group"); err != nil {
		return nil, err
	}

	if isParentProcessGroupChanged(resource, config, entity) {
		snippet, err := nClient.CreateSnippet(nigoapi.SnippetEntity{
			Snippet: &nigoapi.SnippetDto{
				ParentGroupId: entity.Component.ParentGroupId,
				ProcessGroups: map[string]nigoapi.RevisionDto{entity.Id: *entity.Revision},
			},
		})
		if err := clientwrappers.ErrorCreateOperation(log, err, "Create snippet"); err != nil {
			return nil, err
		}

		_, err = nClient.UpdateSnippet(nigoapi.SnippetEntity{
			Snippet: &nigoapi.SnippetDto{
				Id:            snippet.Snippet.Id,
				ParentGroupId: resource.Spec.GetParentProcessGroupID(config.RootProcessGroupId),
			},
		})
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update snippet"); err != nil {
			return nil, err
		}
		return &resource.Status, errorfactory.NifiParentProcessGroupSyncing{}
	}

	isSync, err := processGroupIsSync(resource, entity)
	if err != nil {
		return nil, err
	}

	if !isSync {
		if err := updateProcessGroupEntity(resource, entity); err != nil {
			return nil, err
		}
		entity, err = nClient.UpdateProcessGroup(*entity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update process-group"); err != nil {
			return nil, err
		}
	}

	status := resource.Status
	status.Version = *entity.Revision.Version
	status.Id = entity.Id

	return &status, nil
}

func RemoveProcessGroup(resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	entity, err := nClient.GetProcessGroup(resource.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get resource"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil
		}
		return err
	}

	if err := updateProcessGroupEntity(resource, entity); err != nil {
		return err
	}

	err = nClient.RemoveProcessGroup(*entity)

	return clientwrappers.ErrorRemoveOperation(log, err, "Remove resource")
}

func isParentProcessGroupChanged(
	resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig,
	entity *nigoapi.ProcessGroupEntity) bool {
	return resource.Spec.GetParentProcessGroupID(config.RootProcessGroupId) != entity.Component.ParentGroupId
}

func processGroupIsSync(resource *v1alpha1.NifiResource, entity *nigoapi.ProcessGroupEntity) (bool, error) {
	config, err := resource.Spec.GetConfiguration()
	if err != nil {
		return false, err
	}

	if comments, ok := config["comments"].(string); ok {
		if entity.Component.Comments != comments {
			return false, nil
		}
	}
	if defaultBackPressureDataSizeThreshold, ok := config["defaultBackPressureDataSizeThreshold"].(string); ok {
		if entity.Component.DefaultBackPressureDataSizeThreshold != defaultBackPressureDataSizeThreshold {
			return false, nil
		}
	}
	if defaultBackPressureObjectThreshold, ok := config["defaultBackPressureObjectThreshold"].(int64); ok {
		if entity.Component.DefaultBackPressureObjectThreshold != defaultBackPressureObjectThreshold {
			return false, nil
		}
	}
	if defaultFlowFileExpiration, ok := config["defaultFlowFileExpiration"].(string); ok {
		if entity.Component.DefaultFlowFileExpiration != defaultFlowFileExpiration {
			return false, nil
		}
	}
	if executionEngine, ok := config["executionEngine"].(string); ok {
		if entity.Component.ExecutionEngine != executionEngine {
			return false, nil
		}
	}
	if flowfileConcurrency, ok := config["flowfileConcurrency"].(string); ok {
		if entity.Component.FlowfileConcurrency != flowfileConcurrency {
			return false, nil
		}
	}
	if flowfileOutboundPolicy, ok := config["flowfileOutboundPolicy"].(string); ok {
		if entity.Component.FlowfileOutboundPolicy != flowfileOutboundPolicy {
			return false, nil
		}
	}
	if logFileSuffix, ok := config["logFileSuffix"].(string); ok {
		if entity.Component.LogFileSuffix != logFileSuffix {
			return false, nil
		}
	}
	if positionRaw, ok := config["position"].(map[string]interface{}); ok {
		positionJSON, err := json.Marshal(positionRaw)
		if err != nil {
			return false, err
		}

		var position nigoapi.PositionDto
		if err := json.Unmarshal(positionJSON, &position); err != nil {
			return false, err
		}

		if entity.Component.Position == nil ||
			entity.Component.Position.X != position.X ||
			entity.Component.Position.Y != position.Y {
			return false, nil
		}
	}

	return resource.GetDisplayName() == entity.Component.Name, nil
}

func updateProcessGroupEntity(resource *v1alpha1.NifiResource, entity *nigoapi.ProcessGroupEntity) error {
	config, err := resource.Spec.GetConfiguration()
	if err != nil {
		return err
	}

	var defaultVersion int64 = 0

	if entity == nil {
		entity = &nigoapi.ProcessGroupEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.ProcessGroupDto{}
	}

	entity.Component.Name = resource.GetDisplayName()

	if val, ok := config["comments"].(string); ok {
		entity.Component.Comments = val
	}
	if val, ok := config["defaultBackPressureDataSizeThreshold"].(string); ok {
		entity.Component.DefaultBackPressureDataSizeThreshold = val
	}
	if val, ok := config["defaultBackPressureObjectThreshold"].(int64); ok {
		entity.Component.DefaultBackPressureObjectThreshold = val
	}
	if val, ok := config["defaultFlowFileExpiration"].(string); ok {
		entity.Component.DefaultFlowFileExpiration = val
	}
	if val, ok := config["executionEngine"].(string); ok {
		entity.Component.ExecutionEngine = val
	}
	if val, ok := config["flowfileConcurrency"].(string); ok {
		entity.Component.FlowfileConcurrency = val
	}
	if val, ok := config["flowfileOutboundPolicy"].(string); ok {
		entity.Component.FlowfileOutboundPolicy = val
	}
	if val, ok := config["logFileSuffix"].(string); ok {
		entity.Component.LogFileSuffix = val
	}
	if posRaw, ok := config["position"].(map[string]interface{}); ok {
		posJSON, err := json.Marshal(posRaw)
		if err != nil {
			return fmt.Errorf("error marshaling position: %w", err)
		}

		var position nigoapi.PositionDto
		if err := json.Unmarshal(posJSON, &position); err != nil {
			return fmt.Errorf("error unmarshaling position: %w", err)
		}

		entity.Component.Position = &position
	}

	return nil
}
