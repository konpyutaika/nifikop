package processgroup

import (
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/dataflow"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("processgroup-method")

// ProcessGroupExist check if the NifiResource (ProcessGroup) exist on NiFi Cluster.
func ProcessGroupExist(processGroup *v1alpha1.NifiResource, config *clientconfig.NifiConfig) (bool, error) {
	log.Debug("Checking existence of process group",
		zap.String("clusterName", processGroup.Spec.ClusterRef.Name),
		zap.String("processGroup", processGroup.Name))

	if processGroup.Status.UUID == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	processGroupEntity, err := nClient.GetProcessGroup(processGroup.Status.UUID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return processGroupEntity != nil, nil
}

func RootProcessGroup(config *clientconfig.NifiConfig) (string, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return "", err
	}

	rootPg, err := nClient.GetFlow("root")
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return "", nil
		}
		return "", err
	}

	return rootPg.ProcessGroupFlow.Id, nil
}

func GetProcessGroupInformation(processGroup *v1alpha1.NifiResource, config *clientconfig.NifiConfig) (*nigoapi.ProcessGroupFlowEntity, error) {
	if processGroup.Status.UUID == "" {
		return nil, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	flowEntity, err := nClient.GetFlow(processGroup.Status.UUID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil, nil
		}
		return nil, err
	}

	return flowEntity, nil
}

// CreateProcessGroup will deploy the NifiDataflow on NiFi Cluster.
func CreateProcessGroup(resource *v1alpha1.NifiResource, parameterContext *v1.NifiParameterContext, parentProcessGroup *v1alpha1.NifiResource, config *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	log.Debug("Creating process group",
		zap.String("clusterName", resource.Spec.ClusterRef.Name),
		zap.String("processGroup", resource.Name))

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.ProcessGroupEntity{}
	err = updateProcessGroupEntity(resource, parameterContext, parentProcessGroup, config, &scratchEntity)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.CreateProcessGroup(scratchEntity, resource.Spec.GetParentProcessGroupID(config.RootProcessGroupId, parentProcessGroup))

	if err := clientwrappers.ErrorCreateOperation(log, err, "Create process-group"); err != nil {
		return nil, err
	}

	resource.Status.UUID = entity.Id
	return &resource.Status, nil
}

// IsProcessGroupUnscheduled control if the deployed process group has unscheduled controller services and components.
func IsProcessGroupUnscheduled(processGroup *v1alpha1.NifiResource, config *clientconfig.NifiConfig) (bool, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	// Check all controller services are enabled
	csEntities, err := nClient.GetFlowControllerServices(processGroup.Status.UUID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow controller services"); err != nil {
		return false, err
	}

	// Extracted from Configuraion
	resourceConfig, err := processGroup.Spec.GetConfiguration()
	if err != nil {
		return false, err
	}

	skipInvalidControllerService, _ := resourceConfig["skipInvalidControllerService"].(bool)
	for _, csEntity := range csEntities.ControllerServices {
		if csEntity.Status.RunStatus != "ENABLED" &&
			!(skipInvalidControllerService && csEntity.Status.ValidationStatus == "INVALID") {
			return true, nil
		}
	}

	// Check all components are ok
	processGroups, _, _, _, _ := dataflow.ListComponents(config, processGroup.Status.UUID)
	pGEntity, err := nClient.GetProcessGroup(processGroup.Status.UUID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		return false, err
	}
	processGroups = append(processGroups, *pGEntity)

	skipInvalidComponent, _ := resourceConfig["skipInvalidComponent"].(bool)
	for _, pgEntity := range processGroups {
		if pgEntity.StoppedCount > 0 || (!skipInvalidComponent && pgEntity.InvalidCount > 0) {
			return true, nil
		}
	}

	return false, nil
}

// ScheduleProcessGroup will schedule the controller services and components of the NifiResource
func ScheduleProcessGroup(processGroup *v1alpha1.NifiResource, config *clientconfig.NifiConfig) error {
	log.Debug("Scheduling process group",
		zap.String("clusterName", processGroup.Spec.ClusterRef.Name),
		zap.String("processGroup", processGroup.Name))

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	// Schedule controller services
	_, err = nClient.UpdateFlowControllerServices(nigoapi.ActivateControllerServicesEntity{
		Id:    processGroup.Status.UUID,
		State: "ENABLED",
	})
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Schedule flow's controller services"); err != nil {
		return err
	}

	// Check all controller services are enabled
	csEntities, err := nClient.GetFlowControllerServices(processGroup.Status.UUID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow controller services"); err != nil {
		return err
	}

	// Extracted from Configuraion
	resourceConfig, err := processGroup.Spec.GetConfiguration()
	if err != nil {
		return err
	}

	skipInvalidControllerService, _ := resourceConfig["skipInvalidControllerService"].(bool)
	for _, csEntity := range csEntities.ControllerServices {
		if csEntity.Status.RunStatus != "ENABLED" &&
			!(skipInvalidControllerService && csEntity.Status.ValidationStatus == "INVALID") {
			return errorfactory.NifiFlowControllerServiceScheduling{}
		}
	}

	// Schedule flow
	_, err = nClient.UpdateFlowProcessGroup(nigoapi.ScheduleComponentsEntity{
		Id:    processGroup.Status.UUID,
		State: "RUNNING",
	})
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Schedule flow"); err != nil {
		return err
	}

	// Check all components are ok
	processGroups, _, _, _, _ := dataflow.ListComponents(config, processGroup.Status.UUID)
	pGEntity, err := nClient.GetProcessGroup(processGroup.Status.UUID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		return err
	}
	processGroups = append(processGroups, *pGEntity)

	skipInvalidComponent, _ := resourceConfig["skipInvalidComponent"].(bool)
	for _, pgEntity := range processGroups {
		if pgEntity.StoppedCount > 0 || (!skipInvalidComponent && pgEntity.InvalidCount > 0) {
			return errorfactory.NifiFlowScheduling{}
		}
	}

	return nil
}

// IsOutOfSyncResource control if the deployed dataflow is out of sync with the NifiResource resource.
func IsOutOfSyncResource(
	processGroup *v1alpha1.NifiResource,
	parentProcessGroup *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig,
	parameterContext *v1.NifiParameterContext) (bool, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	pGEntity, err := nClient.GetProcessGroup(processGroup.Status.UUID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		return false, err
	}

	processGroups, _, _, _, err := dataflow.ListComponents(config, processGroup.Status.UUID)
	if err != nil {
		return false, err
	}
	processGroups = append(processGroups, *pGEntity)

	return dataflow.IsParameterContextChanged(parameterContext, processGroups) ||
		isParentProcessGroupChanged(processGroup, parentProcessGroup, config, pGEntity) || isNameChanged(processGroup, pGEntity) || isPostionChanged(processGroup, pGEntity), nil
}

func SyncProcessGroup(
	resource *v1alpha1.NifiResource,
	parameterContext *v1.NifiParameterContext,
	parentProcessGroup *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	// Extracted from Configuraion
	resourceConfig, err := resource.Spec.GetConfiguration()
	if err != nil {
		return nil, err
	}

	pGEntity, err := nClient.GetProcessGroup(resource.Status.UUID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		return nil, err
	}

	processGroups, _, _, _, err := dataflow.ListComponents(config, resource.Status.UUID)
	if err != nil {
		return nil, err
	}

	processGroups = append(processGroups, *pGEntity)

	if dataflow.IsParameterContextChanged(parameterContext, processGroups) {
		// unschedule process group
		if err := UnscheduleProcessGroup(resource, parentProcessGroup, config); err != nil {
			return nil, err
		}

		for _, pg := range processGroups {
			if parameterContext == nil {
				pg.Component.ParameterContext = &nigoapi.ParameterContextReferenceEntity{}
			} else {
				pg.Component.ParameterContext = &nigoapi.ParameterContextReferenceEntity{
					Id: parameterContext.Status.Id,
				}
			}

			_, err := nClient.UpdateProcessGroup(pg)
			if err := clientwrappers.ErrorUpdateOperation(log, err, "Set parameter-context"); err != nil {
				return nil, err
			}
		}

		return &resource.Status, errorfactory.NifiFlowSyncing{}
	}

	if isNameChanged(resource, pGEntity) || isPostionChanged(resource, pGEntity) {
		pGEntity.Component.ParentGroupId = resource.Spec.GetParentProcessGroupID(config.RootProcessGroupId, parentProcessGroup)
		pGEntity.Component.Name = resource.Spec.Name

		var xPos, yPos float64
		xPos, yPos = getPositionFromConfiguration(resourceConfig)

		pGEntity.Component.Position = &nigoapi.PositionDto{
			X: xPos,
			Y: yPos,
		}

		_, err := nClient.UpdateProcessGroup(*pGEntity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Stop flow"); err != nil {
			return nil, err
		}
		return &resource.Status, errorfactory.NifiFlowSyncing{}
	}

	// TODO HERE
	if isParentProcessGroupChanged(resource, parentProcessGroup, config, pGEntity) {
		snippet, err := nClient.CreateSnippet(nigoapi.SnippetEntity{
			Snippet: &nigoapi.SnippetDto{
				ParentGroupId: pGEntity.Component.ParentGroupId,
				ProcessGroups: map[string]nigoapi.RevisionDto{pGEntity.Id: *pGEntity.Revision},
			},
		})
		if err := clientwrappers.ErrorCreateOperation(log, err, "Create snippet"); err != nil {
			return nil, err
		}

		_, err = nClient.UpdateSnippet(nigoapi.SnippetEntity{
			Snippet: &nigoapi.SnippetDto{
				Id:            snippet.Snippet.Id,
				ParentGroupId: resource.Spec.GetParentProcessGroupID(config.RootProcessGroupId, parentProcessGroup),
			},
		})
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update snippet"); err != nil {
			return nil, err
		}
		return &resource.Status, errorfactory.NifiFlowSyncing{}
	}

	isOutOfSink, err := IsOutOfSyncResource(resource, parentProcessGroup, config, parameterContext)
	if err != nil {
		return &resource.Status, err
	}
	if isOutOfSink {
		status, err := prepareUpdatePG(resource, parentProcessGroup, config)
		if err != nil {
			return status, err
		}
		resource.Status = *status

		if err := UnscheduleProcessGroup(resource, parentProcessGroup, config); err != nil {
			return &resource.Status, err
		}
	}

	return &resource.Status, nil

}

func updateProcessGroupEntity(
	processGroup *v1alpha1.NifiResource,
	parameterContext *v1.NifiParameterContext,
	parentProcessGroup *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig,
	entity *nigoapi.ProcessGroupEntity) error {

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

	entity.Component.Name = processGroup.Spec.Name
	entity.Component.ParentGroupId = processGroup.Spec.GetParentProcessGroupID(config.RootProcessGroupId, parentProcessGroup)
	entity.Component.Comments = processGroup.Spec.Comments

	var xPos, yPos float64
	if entity.Component.Position != nil {
		xPos = entity.Component.Position.X
		yPos = entity.Component.Position.Y
	}

	// Extracted from Configuraion
	resourceConfig, err := processGroup.Spec.GetConfiguration()
	if err != nil {
		return err
	}

	if resourceConfig != nil {

		if val, ok := resourceConfig["defaultBackPressureDataSizeThreshold"].(string); ok {
			entity.Component.DefaultBackPressureDataSizeThreshold = val
		}
		if val, ok := resourceConfig["defaultBackPressureObjectThreshold"].(int64); ok {
			entity.Component.DefaultBackPressureObjectThreshold = val
		}
		if val, ok := resourceConfig["defaultFlowFileExpiration"].(string); ok {
			entity.Component.DefaultFlowFileExpiration = val
		}
		if val, ok := resourceConfig["executionEngine"].(string); ok {
			entity.Component.ExecutionEngine = val
		}
		if val, ok := resourceConfig["flowfileConcurrency"].(string); ok {
			entity.Component.FlowfileConcurrency = val
		}
		if val, ok := resourceConfig["flowfileOutboundPolicy"].(string); ok {
			entity.Component.FlowfileOutboundPolicy = val
		}
		if val, ok := resourceConfig["logFileSuffix"].(string); ok {
			entity.Component.LogFileSuffix = val
		}

		xPos, yPos = getPositionFromConfiguration(resourceConfig)

		if val, ok := resourceConfig["statelessFlowTimeout"].(string); ok {
			entity.Component.StatelessFlowTimeout = val
		}
		if val, ok := resourceConfig["maxConcurrentTasks"].(int32); ok {
			entity.Component.MaxConcurrentTasks = val
		}
	}

	if parameterContext != nil {
		entity.Component.ParameterContext.Id = parameterContext.Status.Id
	}

	entity.Component.Position = &nigoapi.PositionDto{
		X: xPos,
		Y: yPos,
	}

	return nil
}

func RemoveProcessGroup(resource *v1alpha1.NifiResource, parentProcessGroup *v1alpha1.NifiResource, config *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	log.Debug("Removing processGroup",
		zap.String("clusterName", resource.Spec.ClusterRef.Name),
		zap.String("processGroup", resource.Name))

	status, err := prepareUpdatePG(resource, parentProcessGroup, config)
	if err != nil {
		return status, err
	}
	resource.Status = *status

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	if err := UnscheduleProcessGroup(resource, parentProcessGroup, config); err != nil {
		return &resource.Status, err
	}

	pGEntity, err := nClient.GetProcessGroup(resource.Status.UUID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil, nil
		}
		return &resource.Status, err
	}

	err = nClient.RemoveProcessGroup(*pGEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Remove process-group"); err != nil {
		return &resource.Status, err
	}

	return nil, nil
}

func UnscheduleProcessGroup(resource *v1alpha1.NifiResource, parentProcessGroup *v1alpha1.NifiResource, config *clientconfig.NifiConfig) error {
	log.Debug("Unscheduling process group",
		zap.String("clusterName", resource.Spec.ClusterRef.Name),
		zap.String("processGroup", resource.Name))

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	// UnSchedule flow
	_, err = nClient.UpdateFlowProcessGroup(nigoapi.ScheduleComponentsEntity{
		Id:    resource.Status.UUID,
		State: "STOPPED",
	})
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Unschedule process group"); err != nil {
		return err
	}

	// Schedule controller services
	_, err = nClient.UpdateFlowControllerServices(nigoapi.ActivateControllerServicesEntity{
		Id:    resource.Status.UUID,
		State: "DISABLED",
	})
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Unschedule process groups's controller services"); err != nil {
		return err
	}

	// Check all controller services are enabled
	csEntities, err := nClient.GetFlowControllerServices(resource.Status.UUID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group controller services"); err != nil {
		return err
	}

	resourceConfig, _ := resource.Spec.GetConfiguration()
	for _, csEntity := range csEntities.ControllerServices {
		if csEntity.Status.RunStatus != "DISABLED" && // TODO the skipInvalid
			!(resourceConfig["skipInvalidControllerService"].(bool) && csEntity.Status.ValidationStatus == "INVALID") {
			return errorfactory.NifiFlowControllerServiceScheduling{}
		}
	}

	// Check all components are ok
	flowEntity, err := nClient.GetFlow(resource.Spec.GetParentProcessGroupID(config.RootProcessGroupId, parentProcessGroup))
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		return err
	}

	pgEntity := processGroupFromResource(flowEntity, resource)
	if pgEntity == nil {
		return errorfactory.NifiFlowScheduling{}
	}

	if pgEntity.RunningCount > 0 {
		return errorfactory.NifiFlowScheduling{}
	}

	return nil
}

func isParentProcessGroupChanged(
	flow *v1alpha1.NifiResource,
	parentProcessGroup *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig,
	pgFlowEntity *nigoapi.ProcessGroupEntity) bool {
	return flow.Spec.GetParentProcessGroupID(config.RootProcessGroupId, parentProcessGroup) != pgFlowEntity.Component.ParentGroupId
}

func isNameChanged(resource *v1alpha1.NifiResource, pgFlowEntity *nigoapi.ProcessGroupEntity) bool {
	return resource.Spec.Name != pgFlowEntity.Component.Name
}

func isPostionChanged(processGroup *v1alpha1.NifiResource, pgFlowEntity *nigoapi.ProcessGroupEntity) bool {
	// Extracted from Configuraion
	resourceConfig, err := processGroup.Spec.GetConfiguration()
	if err == nil && resourceConfig != nil {
		if val, ok := resourceConfig["position"].(map[string]interface{}); ok {
			xVal, xok := val["posX"]
			yVal, yok := val["posY"]

			if !xok || !yok || xVal == nil || yVal == nil || xVal == 0 || yVal == 0 {
				return false
			} else {
				return xVal.(float64) != pgFlowEntity.Component.Position.X || yVal.(float64) != pgFlowEntity.Component.Position.Y
			}
		}
	}
	return false
}

// prepareUpdatePG ensure drain or drop logic.
func prepareUpdatePG(processGroup *v1alpha1.NifiResource, parentProcessGroup *v1alpha1.NifiResource, config *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	// Extracted from Configuraion
	resourceConfig, err := processGroup.Spec.GetConfiguration()
	if err != nil {
		return nil, err
	}

	if resourceConfig["updateStrategy"] == v1.DropStrategy {
		// unschedule processors
		_, err := nClient.UpdateFlowProcessGroup(nigoapi.ScheduleComponentsEntity{
			Id:    processGroup.Status.UUID,
			State: "STOPPED",
		})
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Stop flow"); err != nil {
			return nil, err
		}

		//
		if processGroup.Status.LatestDropRequest != nil && !processGroup.Status.LatestDropRequest.Finished && !processGroup.Status.LatestDropRequest.NotFound {
			dropRequest, err :=
				nClient.GetDropRequest(processGroup.Status.LatestDropRequest.ConnectionId, processGroup.Status.LatestDropRequest.Id)
			if err := clientwrappers.ErrorGetOperation(log, err, "Get drop-request"); err != nificlient.ErrNifiClusterReturned404 {
				if err != nil {
					return nil, err
				}

				processGroup.Status.LatestDropRequest =
					dataflow.DropRequest2Status(processGroup.Status.LatestDropRequest.ConnectionId, dropRequest)
				if !dropRequest.DropRequest.Finished {
					return &processGroup.Status, errorfactory.NifiConnectionDropping{}
				}
			}

			if err == nificlient.ErrNifiClusterReturned404 {
				processGroup.Status.LatestDropRequest.NotFoundRetryCount += 1
				if processGroup.Status.LatestDropRequest.NotFoundRetryCount >= 3 {
					processGroup.Status.LatestDropRequest.NotFound = true
				}
				return &processGroup.Status, errorfactory.NifiConnectionDropRequestNotFound{}
			}
		}

		// Drop all events in connections
		_, _, connections, _, err := dataflow.ListComponents(config, processGroup.Status.UUID)
		if err := clientwrappers.ErrorGetOperation(log, err, "Get recursively flow components"); err != nil {
			return nil, err
		}
		for _, connection := range connections {
			if connection.Status.AggregateSnapshot.FlowFilesQueued != 0 {
				dropRequest, err := nClient.CreateDropRequest(connection.Id)
				if err := clientwrappers.ErrorUpdateOperation(log, err, "Create drop-request"); err != nil {
					return nil, err
				}

				processGroup.Status.LatestDropRequest =
					dataflow.DropRequest2Status(connection.Id, dropRequest)

				return &processGroup.Status, errorfactory.NifiConnectionDropping{}
			}
		}
	} else {
		// Check all components are ok
		flowEntity, err := nClient.GetFlow(processGroup.Spec.GetParentProcessGroupID(config.RootProcessGroupId, parentProcessGroup))
		if err := clientwrappers.ErrorGetOperation(log, err, "Get flow"); err != nil {
			return nil, err
		}

		pgEntity := processGroupFromResource(flowEntity, processGroup)
		if pgEntity == nil {
			return nil, errorfactory.NifiFlowDraining{}
		}

		// If flow is not fully drained
		if pgEntity.Status.AggregateSnapshot.FlowFilesQueued != 0 {
			_, processors, connections, inputPorts, err := dataflow.ListComponents(config, processGroup.Status.UUID)
			if err := clientwrappers.ErrorGetOperation(log, err, "Get recursively flow components"); err != nil {
				return nil, err
			}

			// Unlist all processors with input connections
			for _, connection := range connections {
				processors = dataflow.RemoveProcessor(processors, connection.DestinationId)
			}

			// Stop all input processor
			for _, processor := range processors {
				if processor.Status.RunStatus == "Running" {
					_, err := nClient.UpdateProcessorRunStatus(processor.Id, nigoapi.ProcessorRunStatusEntity{
						Revision: processor.Revision,
						State:    "STOPPED",
					})
					if err := clientwrappers.ErrorUpdateOperation(log, err, "Stop processor"); err != nil {
						return nil, err
					}
				}
			}

			// Stop all input remote
			for _, inputPort := range inputPorts {
				if inputPort.AllowRemoteAccess && inputPort.Status.RunStatus == "Running" {
					_, err := nClient.UpdateInputPortRunStatus(inputPort.Id, nigoapi.PortRunStatusEntity{
						Revision: inputPort.Revision,
						State:    "STOPPED",
					})
					if err := clientwrappers.ErrorUpdateOperation(log, err, "Stop remote input-port"); err != nil {
						return nil, err
					}
				}
			}
			return nil, errorfactory.NifiFlowDraining{}
		}
	}

	return &processGroup.Status, nil
}

// processGroupFromFlow convert a ProcessGroupFlowEntity to NifiDataflow.
func processGroupFromResource(
	flowEntity *nigoapi.ProcessGroupFlowEntity,
	resource *v1alpha1.NifiResource) *nigoapi.ProcessGroupEntity {
	for _, entity := range flowEntity.ProcessGroupFlow.Flow.ProcessGroups {
		if entity.Id == resource.Status.UUID {
			return &entity
		}
	}

	return nil
}

func getPositionFromConfiguration(resourceConfig map[string]interface{}) (float64, float64) {
	var xPos, yPos float64
	if val, ok := resourceConfig["position"].(map[string]interface{}); ok {
		xVal, ok := val["posX"]
		if !ok || xVal == nil || xVal == 0 {
			xPos = float64(1)
		} else {
			xPos = xVal.(float64)
		}

		yVal, ok := val["posY"]
		if !ok || yVal == nil || yVal == 0 {
			yPos = float64(1)
		} else {
			yPos = yVal.(float64)
		}
	} else {
		xPos = float64(1)
		yPos = float64(1)
	}
	return xPos, yPos
}

// Check if a process group contains flowfile.
func IsProcessGroupEmpty(resource *v1alpha1.NifiResource, parentProcessGroup *v1alpha1.NifiResource, config *clientconfig.NifiConfig) (bool, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	flowEntity, err := nClient.GetFlow(resource.Spec.GetParentProcessGroupID(config.RootProcessGroupId, parentProcessGroup))
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow"); err != nil {
		return false, err
	}

	pgEntity := processGroupFromResource(flowEntity, resource)
	if pgEntity == nil {
		return false, errorfactory.NifiFlowDraining{}
	}

	return pgEntity.Status.AggregateSnapshot.FlowFilesQueued == 0, nil
}
