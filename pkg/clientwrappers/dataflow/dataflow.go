package dataflow

import (
	"strings"

	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("dataflow-method")

// DataflowExist check if the NifiDataflow exist on NiFi Cluster.
func DataflowExist(flow *v1.NifiDataflow, config *clientconfig.NifiConfig) (bool, error) {
	log.Debug("Checking existence of dataflow",
		zap.String("clusterName", flow.Spec.ClusterRef.Name),
		zap.String("flowId", flow.Spec.FlowId),
		zap.String("dataflow", flow.Name))

	if flow.Status.ProcessGroupID == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	flowEntity, err := nClient.GetFlow(flow.Status.ProcessGroupID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return flowEntity != nil, nil
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

func GetDataflowInformation(flow *v1.NifiDataflow, config *clientconfig.NifiConfig) (*nigoapi.ProcessGroupFlowEntity, error) {
	if flow.Status.ProcessGroupID == "" {
		return nil, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	flowEntity, err := nClient.GetFlow(flow.Status.ProcessGroupID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil, nil
		}
		return nil, err
	}

	return flowEntity, nil
}

// CreateDataflow will deploy the NifiDataflow on NiFi Cluster.
func CreateDataflow(flow *v1.NifiDataflow, config *clientconfig.NifiConfig,
	registry *v1.NifiRegistryClient) (*v1.NifiDataflowStatus, error) {
	log.Debug("Creating dataflow",
		zap.String("clusterName", flow.Spec.ClusterRef.Name),
		zap.String("flowId", flow.Spec.FlowId),
		zap.String("dataflow", flow.Name))

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.ProcessGroupEntity{}
	updateProcessGroupEntity(flow, registry, config, &scratchEntity)

	entity, err := nClient.CreateProcessGroup(scratchEntity, flow.Spec.GetParentProcessGroupID(config.RootProcessGroupId))

	if err := clientwrappers.ErrorCreateOperation(log, err, "Create process-group"); err != nil {
		return nil, err
	}

	flow.Status.ProcessGroupID = entity.Id
	return &flow.Status, nil
}

// IsDataflowUnscheduled control if the deployed dataflow has unscheduled controller services and components.
func IsDataflowUnscheduled(flow *v1.NifiDataflow, config *clientconfig.NifiConfig) (bool, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	// Check all controller services are enabled
	csEntities, err := nClient.GetFlowControllerServices(flow.Status.ProcessGroupID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow controller services"); err != nil {
		return false, err
	}
	for _, csEntity := range csEntities.ControllerServices {
		if csEntity.Status.RunStatus != "ENABLED" &&
			!(flow.Spec.SkipInvalidControllerService && csEntity.Status.ValidationStatus == "INVALID") {
			return true, nil
		}
	}

	// Check all components are ok
	processGroups, _, _, _, _ := listComponents(config, flow.Status.ProcessGroupID)
	pGEntity, err := nClient.GetProcessGroup(flow.Status.ProcessGroupID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		return false, err
	}
	processGroups = append(processGroups, *pGEntity)

	for _, pgEntity := range processGroups {
		if pgEntity.StoppedCount > 0 || (!flow.Spec.SkipInvalidComponent && pgEntity.InvalidCount > 0) {
			return true, nil
		}
	}

	return false, nil
}

// ScheduleDataflow will schedule the controller services and components of the NifiDataflow.
func ScheduleDataflow(flow *v1.NifiDataflow, config *clientconfig.NifiConfig) error {
	log.Debug("Scheduling dataflow",
		zap.String("clusterName", flow.Spec.ClusterRef.Name),
		zap.String("flowId", flow.Spec.FlowId),
		zap.String("dataflow", flow.Name))

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	// Schedule controller services
	_, err = nClient.UpdateFlowControllerServices(nigoapi.ActivateControllerServicesEntity{
		Id:    flow.Status.ProcessGroupID,
		State: "ENABLED",
	})
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Schedule flow's controller services"); err != nil {
		return err
	}

	// Check all controller services are enabled
	csEntities, err := nClient.GetFlowControllerServices(flow.Status.ProcessGroupID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow controller services"); err != nil {
		return err
	}
	for _, csEntity := range csEntities.ControllerServices {
		if csEntity.Status.RunStatus != "ENABLED" &&
			!(flow.Spec.SkipInvalidControllerService && csEntity.Status.ValidationStatus == "INVALID") {
			return errorfactory.NifiFlowControllerServiceScheduling{}
		}
	}

	// Schedule flow
	_, err = nClient.UpdateFlowProcessGroup(nigoapi.ScheduleComponentsEntity{
		Id:    flow.Status.ProcessGroupID,
		State: "RUNNING",
	})
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Schedule flow"); err != nil {
		return err
	}

	// Check all components are ok
	processGroups, _, _, _, _ := listComponents(config, flow.Status.ProcessGroupID)
	pGEntity, err := nClient.GetProcessGroup(flow.Status.ProcessGroupID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		return err
	}
	processGroups = append(processGroups, *pGEntity)

	for _, pgEntity := range processGroups {
		if pgEntity.StoppedCount > 0 || (!flow.Spec.SkipInvalidComponent && pgEntity.InvalidCount > 0) {
			return errorfactory.NifiFlowScheduling{}
		}
	}

	return nil
}

// IsOutOfSyncDataflow control if the deployed dataflow is out of sync with the NifiDataflow resource.
func IsOutOfSyncDataflow(
	flow *v1.NifiDataflow,
	config *clientconfig.NifiConfig,
	registry *v1.NifiRegistryClient,
	parameterContext *v1.NifiParameterContext) (bool, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	pGEntity, err := nClient.GetProcessGroup(flow.Status.ProcessGroupID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		return false, err
	}

	processGroups, _, _, _, err := listComponents(config, flow.Status.ProcessGroupID)
	if err != nil {
		return false, err
	}
	processGroups = append(processGroups, *pGEntity)

	return isParameterContextChanged(parameterContext, processGroups) ||
		isVersioningChanged(flow, registry, pGEntity) || !isVersionSync(flow, pGEntity) || localChanged(pGEntity) ||
		isParentProcessGroupChanged(flow, config, pGEntity) || isNameChanged(flow, pGEntity) || isPostionChanged(flow, pGEntity), nil
}

func isParameterContextChanged(
	parameterContext *v1.NifiParameterContext,
	processGroups []nigoapi.ProcessGroupEntity) bool {
	for _, processGroup := range processGroups {
		pgParameterContext := processGroup.ParameterContext

		if pgParameterContext == nil && parameterContext == nil {
			continue
		}

		if (pgParameterContext == nil && parameterContext != nil) ||
			(pgParameterContext != nil && parameterContext == nil) ||
			processGroup.ParameterContext.Id != parameterContext.Status.Id {
			return true
		}
	}
	return false
}

func isParentProcessGroupChanged(
	flow *v1.NifiDataflow,
	config *clientconfig.NifiConfig,
	pgFlowEntity *nigoapi.ProcessGroupEntity) bool {
	return flow.Spec.GetParentProcessGroupID(config.RootProcessGroupId) != pgFlowEntity.Component.ParentGroupId
}

func isNameChanged(flow *v1.NifiDataflow, pgFlowEntity *nigoapi.ProcessGroupEntity) bool {
	return flow.Name != pgFlowEntity.Component.Name
}

// isVersionSync check if the flow version is out of sync.
func isVersionSync(flow *v1.NifiDataflow, pgFlowEntity *nigoapi.ProcessGroupEntity) bool {
	return *flow.Spec.FlowVersion == pgFlowEntity.Component.VersionControlInformation.Version
}

func localChanged(pgFlowEntity *nigoapi.ProcessGroupEntity) bool {
	return strings.Contains(pgFlowEntity.Component.VersionControlInformation.State, "LOCALLY_MODIFIED")
}

// isVersioningChanged check if the versioning configuration is out of sync on process group.
func isVersioningChanged(
	flow *v1.NifiDataflow,
	registry *v1.NifiRegistryClient,
	pgFlowEntity *nigoapi.ProcessGroupEntity) bool {
	return pgFlowEntity.Component.VersionControlInformation == nil ||
		flow.Spec.FlowId != pgFlowEntity.Component.VersionControlInformation.FlowId ||
		flow.Spec.BucketId != pgFlowEntity.Component.VersionControlInformation.BucketId ||
		registry.Status.Id != pgFlowEntity.Component.VersionControlInformation.RegistryId
}

// isPostionChanged check if the position of the process group is out of sync.
func isPostionChanged(flow *v1.NifiDataflow, pgFlowEntity *nigoapi.ProcessGroupEntity) bool {
	return flow.Spec.FlowPosition != nil &&
		((flow.Spec.FlowPosition.X != nil && float64(flow.Spec.FlowPosition.GetX()) != pgFlowEntity.Component.Position.X) ||
			(flow.Spec.FlowPosition.Y != nil && float64(flow.Spec.FlowPosition.GetY()) != pgFlowEntity.Component.Position.Y))
}

// SyncDataflow implements the logic to sync a NifiDataflow with the deployed flow.
func SyncDataflow(
	flow *v1.NifiDataflow,
	config *clientconfig.NifiConfig,
	registry *v1.NifiRegistryClient,
	parameterContext *v1.NifiParameterContext) (*v1.NifiDataflowStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	pGEntity, err := nClient.GetProcessGroup(flow.Status.ProcessGroupID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		return nil, err
	}

	processGroups, _, _, _, err := listComponents(config, flow.Status.ProcessGroupID)
	if err != nil {
		return nil, err
	}

	processGroups = append(processGroups, *pGEntity)
	if isParameterContextChanged(parameterContext, processGroups) {
		// unschedule dataflow
		if err := UnscheduleDataflow(flow, config); err != nil {
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
		return &flow.Status, errorfactory.NifiFlowSyncing{}
	}

	if isVersioningChanged(flow, registry, pGEntity) {
		return RemoveDataflow(flow, config)
	}

	if isNameChanged(flow, pGEntity) || isPostionChanged(flow, pGEntity) {
		pGEntity.Component.ParentGroupId = flow.Spec.GetParentProcessGroupID(config.RootProcessGroupId)
		pGEntity.Component.Name = flow.Name

		var xPos, yPos float64
		if flow.Spec.FlowPosition == nil || flow.Spec.FlowPosition.X == nil {
			xPos = pGEntity.Component.Position.X
		} else {
			xPos = float64(flow.Spec.FlowPosition.GetX())
		}

		if flow.Spec.FlowPosition == nil || flow.Spec.FlowPosition.Y == nil {
			yPos = pGEntity.Component.Position.Y
		} else {
			yPos = float64(flow.Spec.FlowPosition.GetY())
		}

		pGEntity.Component.Position = &nigoapi.PositionDto{
			X: xPos,
			Y: yPos,
		}

		_, err := nClient.UpdateProcessGroup(*pGEntity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Stop flow"); err != nil {
			return nil, err
		}
		return &flow.Status, errorfactory.NifiFlowSyncing{}
	}

	if isParentProcessGroupChanged(flow, config, pGEntity) {
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
				ParentGroupId: flow.Spec.GetParentProcessGroupID(config.RootProcessGroupId),
			},
		})
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update snippet"); err != nil {
			return nil, err
		}
		return &flow.Status, errorfactory.NifiFlowSyncing{}
	}

	latestUpdateRequest := flow.Status.LatestUpdateRequest
	if latestUpdateRequest != nil && !latestUpdateRequest.Complete && !latestUpdateRequest.NotFound {
		var t v1.DataflowUpdateRequestType
		var err error
		var updateRequest *nigoapi.VersionedFlowUpdateRequestEntity
		if latestUpdateRequest.Type == v1.UpdateRequestType {
			t = v1.UpdateRequestType
			updateRequest, err = nClient.GetVersionUpdateRequest(latestUpdateRequest.Id)
		} else {
			t = v1.RevertRequestType
			updateRequest, err = nClient.GetVersionRevertRequest(latestUpdateRequest.Id)
		}
		if updateRequest != nil {
			flow.Status.LatestUpdateRequest =
				updateRequest2Status(updateRequest, t)
		}

		if err := clientwrappers.ErrorGetOperation(log, err, "Get version-request"); err != nificlient.ErrNifiClusterReturned404 ||
			(updateRequest != nil && updateRequest.Request != nil && !updateRequest.Request.Complete) {
			if err != nil {
				return &flow.Status, err
			}
			return &flow.Status, errorfactory.NifiFlowUpdateRequestRunning{}
		}

		if err == nificlient.ErrNifiClusterReturned404 {
			flow.Status.LatestUpdateRequest.NotFoundRetryCount += 1
			if flow.Status.LatestUpdateRequest.NotFoundRetryCount >= 3 {
				flow.Status.LatestUpdateRequest.NotFound = true
			}
			return &flow.Status, errorfactory.NifiFlowUpdateRequestNotFound{}
		}
	}

	isOutOfSink, err := IsOutOfSyncDataflow(flow, config, registry, parameterContext)
	if err != nil {
		return &flow.Status, err
	}
	if isOutOfSink {
		status, err := prepareUpdatePG(flow, config)
		if err != nil {
			return status, err
		}
		flow.Status = *status

		if err := UnscheduleDataflow(flow, config); err != nil {
			return &flow.Status, err
		}
	}

	pGEntity, err = nClient.GetProcessGroup(flow.Status.ProcessGroupID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process group"); err != nil {
		return nil, err
	}

	if localChanged(pGEntity) {
		vInfo := pGEntity.Component.VersionControlInformation
		updateRequest, err := nClient.CreateVersionRevertRequest(
			flow.Status.ProcessGroupID,
			nigoapi.VersionControlInformationEntity{
				ProcessGroupRevision: pGEntity.Revision,
				VersionControlInformation: &nigoapi.VersionControlInformationDto{
					GroupId:    pGEntity.Id,
					RegistryId: vInfo.RegistryId,
					BucketId:   vInfo.BucketId,
					FlowId:     vInfo.FlowId,
					Version:    vInfo.Version,
				},
			},
		)
		if err := clientwrappers.ErrorCreateOperation(log, err, "Create version revert-request"); err != nil {
			return nil, err
		}

		flow.Status.LatestUpdateRequest =
			updateRequest2Status(updateRequest, v1.RevertRequestType)
		return &flow.Status, errorfactory.NifiFlowUpdateRequestRunning{}
	}

	if !isVersionSync(flow, pGEntity) {
		updateRequest, err := nClient.CreateVersionUpdateRequest(
			flow.Status.ProcessGroupID,
			nigoapi.VersionControlInformationEntity{
				ProcessGroupRevision: pGEntity.Revision,
				VersionControlInformation: &nigoapi.VersionControlInformationDto{
					GroupId:    pGEntity.Id,
					RegistryId: registry.Status.Id,
					BucketId:   flow.Spec.BucketId,
					FlowId:     flow.Spec.FlowId,
					Version:    *flow.Spec.FlowVersion,
				},
			},
		)
		if err := clientwrappers.ErrorCreateOperation(log, err, "Create version update-request"); err != nil {
			return nil, err
		}

		flow.Status.LatestUpdateRequest =
			updateRequest2Status(updateRequest, v1.UpdateRequestType)
		return &flow.Status, errorfactory.NifiFlowUpdateRequestRunning{}
	}

	return &flow.Status, nil
}

// prepareUpdatePG ensure drain or drop logic.
func prepareUpdatePG(flow *v1.NifiDataflow, config *clientconfig.NifiConfig) (*v1.NifiDataflowStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	if flow.Spec.UpdateStrategy == v1.DropStrategy {
		// unschedule processors
		_, err := nClient.UpdateFlowProcessGroup(nigoapi.ScheduleComponentsEntity{
			Id:    flow.Status.ProcessGroupID,
			State: "STOPPED",
		})
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Stop flow"); err != nil {
			return nil, err
		}

		//
		if flow.Status.LatestDropRequest != nil && !flow.Status.LatestDropRequest.Finished && !flow.Status.LatestDropRequest.NotFound {
			dropRequest, err :=
				nClient.GetDropRequest(flow.Status.LatestDropRequest.ConnectionId, flow.Status.LatestDropRequest.Id)
			if err := clientwrappers.ErrorGetOperation(log, err, "Get drop-request"); err != nificlient.ErrNifiClusterReturned404 {
				if err != nil {
					return nil, err
				}

				flow.Status.LatestDropRequest =
					dropRequest2Status(flow.Status.LatestDropRequest.ConnectionId, dropRequest)
				if !dropRequest.DropRequest.Finished {
					return &flow.Status, errorfactory.NifiConnectionDropping{}
				}
			}

			if err == nificlient.ErrNifiClusterReturned404 {
				flow.Status.LatestDropRequest.NotFoundRetryCount += 1
				if flow.Status.LatestDropRequest.NotFoundRetryCount >= 3 {
					flow.Status.LatestDropRequest.NotFound = true
				}
				return &flow.Status, errorfactory.NifiConnectionDropRequestNotFound{}
			}
		}

		// Drop all events in connections
		_, _, connections, _, err := listComponents(config, flow.Status.ProcessGroupID)
		if err := clientwrappers.ErrorGetOperation(log, err, "Get recursively flow components"); err != nil {
			return nil, err
		}
		for _, connection := range connections {
			if connection.Status.AggregateSnapshot.FlowFilesQueued != 0 {
				dropRequest, err := nClient.CreateDropRequest(connection.Id)
				if err := clientwrappers.ErrorUpdateOperation(log, err, "Create drop-request"); err != nil {
					return nil, err
				}

				flow.Status.LatestDropRequest =
					dropRequest2Status(connection.Id, dropRequest)

				return &flow.Status, errorfactory.NifiConnectionDropping{}
			}
		}
	} else {
		// Check all components are ok
		flowEntity, err := nClient.GetFlow(flow.Spec.GetParentProcessGroupID(config.RootProcessGroupId))
		if err := clientwrappers.ErrorGetOperation(log, err, "Get flow"); err != nil {
			return nil, err
		}

		pgEntity := processGroupFromFlow(flowEntity, flow)
		if pgEntity == nil {
			return nil, errorfactory.NifiFlowDraining{}
		}

		// If flow is not fully drained
		if pgEntity.Status.AggregateSnapshot.FlowFilesQueued != 0 {
			_, processors, connections, inputPorts, err := listComponents(config, flow.Status.ProcessGroupID)
			if err := clientwrappers.ErrorGetOperation(log, err, "Get recursively flow components"); err != nil {
				return nil, err
			}

			// Unlist all processors with input connections
			for _, connection := range connections {
				processors = removeProcessor(processors, connection.DestinationId)
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

	return &flow.Status, nil
}

func RemoveDataflow(flow *v1.NifiDataflow, config *clientconfig.NifiConfig) (*v1.NifiDataflowStatus, error) {
	log.Debug("Removing dataflow",
		zap.String("clusterName", flow.Spec.ClusterRef.Name),
		zap.String("flowId", flow.Spec.FlowId),
		zap.String("dataflow", flow.Name))

	// Prepare Dataflow
	status, err := prepareUpdatePG(flow, config)
	if err != nil {
		return status, err
	}
	flow.Status = *status

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	if err := UnscheduleDataflow(flow, config); err != nil {
		return &flow.Status, err
	}

	pGEntity, err := nClient.GetProcessGroup(flow.Status.ProcessGroupID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil, nil
		}
		return &flow.Status, err
	}

	err = nClient.RemoveProcessGroup(*pGEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Remove process-group"); err != nil {
		return &flow.Status, err
	}

	return nil, nil
}

func UnscheduleDataflow(flow *v1.NifiDataflow, config *clientconfig.NifiConfig) error {
	log.Debug("Unscheduling dataflow",
		zap.String("clusterName", flow.Spec.ClusterRef.Name),
		zap.String("flowId", flow.Spec.FlowId),
		zap.String("dataflow", flow.Name))

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	// UnSchedule flow
	_, err = nClient.UpdateFlowProcessGroup(nigoapi.ScheduleComponentsEntity{
		Id:    flow.Status.ProcessGroupID,
		State: "STOPPED",
	})
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Unschedule flow"); err != nil {
		return err
	}

	// Schedule controller services
	_, err = nClient.UpdateFlowControllerServices(nigoapi.ActivateControllerServicesEntity{
		Id:    flow.Status.ProcessGroupID,
		State: "DISABLED",
	})
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Unschedule flow's controller services"); err != nil {
		return err
	}

	// Check all controller services are enabled
	csEntities, err := nClient.GetFlowControllerServices(flow.Status.ProcessGroupID)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow controller services"); err != nil {
		return err
	}
	for _, csEntity := range csEntities.ControllerServices {
		if csEntity.Status.RunStatus != "DISABLED" &&
			!(flow.Spec.SkipInvalidControllerService && csEntity.Status.ValidationStatus == "INVALID") {
			return errorfactory.NifiFlowControllerServiceScheduling{}
		}
	}

	// Check all components are ok
	flowEntity, err := nClient.GetFlow(flow.Spec.GetParentProcessGroupID(config.RootProcessGroupId))
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow"); err != nil {
		return err
	}

	pgEntity := processGroupFromFlow(flowEntity, flow)
	if pgEntity == nil {
		return errorfactory.NifiFlowScheduling{}
	}

	if pgEntity.RunningCount > 0 {
		return errorfactory.NifiFlowScheduling{}
	}

	return nil
}

// processGroupFromFlow convert a ProcessGroupFlowEntity to NifiDataflow.
func processGroupFromFlow(
	flowEntity *nigoapi.ProcessGroupFlowEntity,
	flow *v1.NifiDataflow) *nigoapi.ProcessGroupEntity {
	for _, entity := range flowEntity.ProcessGroupFlow.Flow.ProcessGroups {
		if entity.Id == flow.Status.ProcessGroupID {
			return &entity
		}
	}

	return nil
}

// // inputConnectionFromFlow retrieve all input connection from a list of input ports
// func inputConnectionFromFlow(flowEntity *nigoapi.ProcessGroupFlowEntity,
// 	inputPorts []nigoapi.PortEntity) []nigoapi.ConnectionEntity {
// 	var connections []nigoapi.ConnectionEntity

// 	for _, connection := range flowEntity.ProcessGroupFlow.Flow.Connections {
// 		for _, inputPort := range inputPorts {
// 			if connection.DestinationId == inputPort.Id {
// 				connections = append(connections, connection)
// 			}
// 		}
// 	}

// 	return connections
// }

// // hasInputConnectionsActive will determine if a flow has input connections that are still active
// func hasInputConnectionsActive(flow *v1.NifiDataflow, config *clientconfig.NifiConfig) (*bool, error) {
// 	var connections []nigoapi.ConnectionEntity

// 	nClient, err := common.NewClusterConnection(log, config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	flowEntity, err := nClient.GetFlow(flow.Status.ProcessGroupID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	parentFlowEntity, err := nClient.GetFlow(flowEntity.ProcessGroupFlow.ParentGroupId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	connections = inputConnectionFromFlow(parentFlowEntity, flowEntity.ProcessGroupFlow.Flow.InputPorts)

// 	var hasConnectionActive bool = false
// 	for _, inputConnection := range connections {
// 		if inputConnection.Status.AggregateSnapshot.FlowFilesQueued > 0 || inputConnection.Component.Source.Running {
// 			hasConnectionActive = true
// 		}
// 	}

// 	return &hasConnectionActive, nil
// }

// listComponents will get all ProcessGroups, Processors, Connections and Ports recursively.
func listComponents(config *clientconfig.NifiConfig,
	processGroupID string) ([]nigoapi.ProcessGroupEntity, []nigoapi.ProcessorEntity, []nigoapi.ConnectionEntity, []nigoapi.PortEntity, error) {
	var processGroups []nigoapi.ProcessGroupEntity
	var processors []nigoapi.ProcessorEntity
	var connections []nigoapi.ConnectionEntity
	var inputPorts []nigoapi.PortEntity

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return processGroups, processors, connections, inputPorts, err
	}

	flowEntity, err := nClient.GetFlow(processGroupID)
	if err != nil {
		return processGroups, processors, connections, inputPorts, err
	}
	flow := flowEntity.ProcessGroupFlow.Flow

	processGroups = flow.ProcessGroups
	processors = flow.Processors
	connections = flow.Connections
	inputPorts = flow.InputPorts

	for _, pg := range flow.ProcessGroups {
		childPG, childP, childC, childI, err := listComponents(config, pg.Id)
		if err != nil {
			return processGroups, processors, connections, inputPorts, err
		}
		processGroups = append(processGroups, childPG...)
		processors = append(processors, childP...)
		connections = append(connections, childC...)
		inputPorts = append(inputPorts, childI...)
	}

	return processGroups, processors, connections, inputPorts, nil
}

func dropRequest2Status(connectionId string, dropRequest *nigoapi.DropRequestEntity) *v1.DropRequest {
	dr := dropRequest.DropRequest
	return &v1.DropRequest{
		ConnectionId:       connectionId,
		Id:                 dr.Id,
		Uri:                dr.Uri,
		LastUpdated:        dr.LastUpdated,
		Finished:           dr.Finished,
		FailureReason:      dr.FailureReason,
		PercentCompleted:   dr.PercentCompleted,
		CurrentCount:       dr.CurrentCount,
		CurrentSize:        dr.CurrentSize,
		Current:            dr.Current,
		OriginalCount:      dr.OriginalCount,
		OriginalSize:       dr.OriginalSize,
		Original:           dr.Original,
		DroppedCount:       dr.DroppedCount,
		DroppedSize:        dr.DroppedSize,
		Dropped:            dr.Dropped,
		State:              dr.State,
		NotFound:           false,
		NotFoundRetryCount: 0,
	}
}

func updateRequest2Status(updateRequest *nigoapi.VersionedFlowUpdateRequestEntity,
	updateType v1.DataflowUpdateRequestType) *v1.UpdateRequest {
	ur := updateRequest.Request
	return &v1.UpdateRequest{
		Type:               updateType,
		Id:                 ur.RequestId,
		Uri:                ur.Uri,
		LastUpdated:        ur.LastUpdated,
		Complete:           ur.Complete,
		FailureReason:      ur.FailureReason,
		PercentCompleted:   ur.PercentCompleted,
		State:              ur.State,
		NotFound:           false,
		NotFoundRetryCount: 0,
	}
}

func updateProcessGroupEntity(
	flow *v1.NifiDataflow,
	registry *v1.NifiRegistryClient,
	config *clientconfig.NifiConfig,
	entity *nigoapi.ProcessGroupEntity) {
	stringFactory := func() string { return "" }

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

	entity.Component.Name = flow.Name
	entity.Component.ParentGroupId = flow.Spec.GetParentProcessGroupID(config.RootProcessGroupId)

	var xPos, yPos float64
	if entity.Component.Position != nil {
		xPos = entity.Component.Position.X
		yPos = entity.Component.Position.Y
	} else {
		if flow.Spec.FlowPosition == nil || flow.Spec.FlowPosition.X == nil {
			xPos = float64(1)
		} else {
			xPos = float64(flow.Spec.FlowPosition.GetX())
		}

		if flow.Spec.FlowPosition == nil || flow.Spec.FlowPosition.Y == nil {
			yPos = float64(1)
		} else {
			yPos = float64(flow.Spec.FlowPosition.GetY())
		}
	}

	entity.Component.Position = &nigoapi.PositionDto{
		X: xPos,
		Y: yPos,
	}

	entity.Component.VersionControlInformation = &nigoapi.VersionControlInformationDto{
		GroupId:          stringFactory(),
		RegistryName:     stringFactory(),
		BucketName:       stringFactory(),
		FlowName:         stringFactory(),
		FlowDescription:  stringFactory(),
		State:            stringFactory(),
		StateExplanation: stringFactory(),
		RegistryId:       registry.Status.Id,
		BucketId:         flow.Spec.BucketId,
		FlowId:           flow.Spec.FlowId,
		Version:          *flow.Spec.FlowVersion,
	}
}

func removeProcessor(processors []nigoapi.ProcessorEntity, toRemoveId string) []nigoapi.ProcessorEntity {
	var tmp []nigoapi.ProcessorEntity

	for _, processor := range processors {
		if processor.Id != toRemoveId {
			tmp = append(tmp, processor)
		}
	}

	return tmp
}

// Check if a dataflow contains flowfile.
func IsDataflowEmpty(flow *v1.NifiDataflow, config *clientconfig.NifiConfig) (bool, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	flowEntity, err := nClient.GetFlow(flow.Spec.GetParentProcessGroupID(config.RootProcessGroupId))
	if err := clientwrappers.ErrorGetOperation(log, err, "Get flow"); err != nil {
		return false, err
	}

	pgEntity := processGroupFromFlow(flowEntity, flow)
	if pgEntity == nil {
		return false, errorfactory.NifiFlowDraining{}
	}

	return pgEntity.Status.AggregateSnapshot.FlowFilesQueued == 0, nil
}
