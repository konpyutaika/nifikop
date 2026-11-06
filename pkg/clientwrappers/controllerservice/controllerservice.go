package controllerservice

import (
	"errors"
	"fmt"

	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("controllerservice-method")

// ErrTransitioning is returned when NiFi is still applying a run state change (ENABLING or
// DISABLING). The caller should requeue and retry once the service settles.
var ErrTransitioning = errors.New("controller service run state transition in progress")

// ErrReferenced is returned when a controller service cannot be removed because other
// components still reference it.
var ErrReferenced = errors.New("controller service is still referenced by other components")

// StandardWebClientServiceProperties are the NiFi property names of StandardWebClientServiceProvider.
var StandardWebClientServiceProperties = struct {
	ConnectTimeout string
	ReadTimeout    string
	WriteTimeout   string
}{
	ConnectTimeout: "Connect Timeout",
	ReadTimeout:    "Read Timeout",
	WriteTimeout:   "Write Timeout",
}

// ExistControllerService reports whether the controller service recorded in the status still
// exists in NiFi.
func ExistControllerService(controllerService *v1alpha1.NifiControllerService, config *clientconfig.NifiConfig) (bool, error) {
	if controllerService.Status.Id == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetControllerService(controllerService.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get controller-service"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

// CreateControllerService creates the controller service in NiFi. It deliberately does not enable
// it: the id must reach the status before any further call, otherwise a failed enable would make
// the next reconciliation create a duplicate. SyncControllerService enables it on the next step.
func CreateControllerService(controllerService *v1alpha1.NifiControllerService,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiControllerServiceStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.ControllerServiceEntity{}
	updateControllerServiceEntity(controllerService, &scratchEntity)

	entity, err := nClient.CreateControllerService(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Failed to create controller-service "+controllerService.Name); err != nil {
		return nil, err
	}

	status := extractStatus(entity)
	return &status, nil
}

// SyncControllerService drives the NiFi controller service towards the desired configuration
// and the ENABLED state, one step per call. NiFi refuses configuration changes while a service
// is enabled, so an out-of-sync enabled service is first disabled, then updated and re-enabled on
// the following calls. The returned status carries the observed run state so the caller can decide
// to requeue while a transition is in progress.
func SyncControllerService(controllerService *v1alpha1.NifiControllerService,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiControllerServiceStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.GetControllerService(controllerService.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get controller-service"); err != nil {
		return nil, err
	}

	switch planSync(entity.Component.State, controllerServiceIsSync(controllerService, entity)) {
	case actionDisable:
		entity, err = nClient.UpdateControllerServiceRunStatus(*entity, string(v1alpha1.ControllerServiceStateDisabled))
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Disable controller-service"); err != nil {
			return nil, err
		}
	case actionUpdateAndEnable:
		updateControllerServiceEntity(controllerService, entity)
		entity, err = nClient.UpdateControllerService(*entity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update controller-service"); err != nil {
			return nil, err
		}
		fallthrough
	case actionEnable:
		entity, err = nClient.UpdateControllerServiceRunStatus(*entity, string(v1alpha1.ControllerServiceStateEnabled))
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Enable controller-service"); err != nil {
			return nil, err
		}
	case actionNone:
	}

	status := extractStatus(entity)
	return &status, nil
}

// RemoveControllerService deletes the controller service from NiFi. It refuses to delete a service
// that is still referenced (ErrReferenced), disables an enabled service first and reports
// ErrTransitioning until NiFi has settled, so the caller can requeue.
func RemoveControllerService(controllerService *v1alpha1.NifiControllerService,
	config *clientconfig.NifiConfig) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	entity, err := nClient.GetControllerService(controllerService.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get controller-service"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil
		}
		return err
	}

	references, err := nClient.GetControllerServiceReferences(entity.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get controller-service references"); err != nil {
		return err
	}

	switch planRemove(entity.Component.State, len(references.ControllerServiceReferencingComponents)) {
	case actionRefuse:
		return fmt.Errorf("%w: %s", ErrReferenced, referenceNames(references))
	case actionDisable:
		if _, err = nClient.UpdateControllerServiceRunStatus(*entity, string(v1alpha1.ControllerServiceStateDisabled)); err != nil {
			return clientwrappers.ErrorUpdateOperation(log, err, "Disable controller-service")
		}
		return ErrTransitioning
	case actionWait:
		return ErrTransitioning
	case actionRemove:
		err = nClient.RemoveControllerService(*entity)
		return clientwrappers.ErrorRemoveOperation(log, err, "Remove controller-service")
	}
	return nil
}

type action int

const (
	actionNone action = iota
	actionEnable
	actionDisable
	actionUpdateAndEnable
	actionWait
	actionRefuse
	actionRemove
)

// planSync picks the next lifecycle step from the observed run state and whether the
// configuration matches the desired one. Transitions in progress are left alone.
func planSync(state string, inSync bool) action {
	switch v1alpha1.NifiControllerServiceState(state) {
	case v1alpha1.ControllerServiceStateEnabling, v1alpha1.ControllerServiceStateDisabling:
		return actionNone
	case v1alpha1.ControllerServiceStateEnabled:
		if inSync {
			return actionNone
		}
		return actionDisable
	default: // DISABLED, or a state this operator does not know about
		if inSync {
			return actionEnable
		}
		return actionUpdateAndEnable
	}
}

// planRemove picks the next deletion step: refuse while referenced, disable first when enabled,
// wait during a transition, otherwise remove.
func planRemove(state string, referenceCount int) action {
	if referenceCount > 0 {
		return actionRefuse
	}
	switch v1alpha1.NifiControllerServiceState(state) {
	case v1alpha1.ControllerServiceStateEnabled:
		return actionDisable
	case v1alpha1.ControllerServiceStateEnabling, v1alpha1.ControllerServiceStateDisabling:
		return actionWait
	default:
		return actionRemove
	}
}

func referenceNames(references *nigoapi.ControllerServiceReferencingComponentsEntity) string {
	names := ""
	for i, ref := range references.ControllerServiceReferencingComponents {
		if i > 0 {
			names += ", "
		}
		if ref.Component != nil {
			names += fmt.Sprintf("%s (%s)", ref.Component.Name, ref.Component.ReferenceType)
		} else {
			names += ref.Id
		}
	}
	return names
}

func extractStatus(entity *nigoapi.ControllerServiceEntity) v1alpha1.NifiControllerServiceStatus {
	status := v1alpha1.NifiControllerServiceStatus{Id: entity.Id}
	if entity.Revision != nil && entity.Revision.Version != nil {
		status.Version = *entity.Revision.Version
	}
	if entity.Component != nil {
		status.State = v1alpha1.NifiControllerServiceState(entity.Component.State)
	}
	return status
}

// controllerServiceIsSync compares the desired configuration with the NiFi entity. Only the
// properties this operator manages are compared, so properties set directly in NiFi are left alone.
func controllerServiceIsSync(controllerService *v1alpha1.NifiControllerService, entity *nigoapi.ControllerServiceEntity) bool {
	if entity.Component == nil {
		return false
	}
	if controllerService.Name != entity.Component.Name ||
		controllerService.Spec.GetType() != entity.Component.Type_ ||
		controllerService.Spec.Description != entity.Component.Comments {
		return false
	}

	switch controllerService.Spec.Type {
	case v1alpha1.StandardWebClientServiceProviderType:
		return controllerServiceIsSync_StandardWebClientServiceProvider(controllerService.Spec.StandardWebClientServiceProviderSpec, entity)
	}
	return true
}

func controllerServiceIsSync_StandardWebClientServiceProvider(cfg *v1alpha1.StandardWebClientServiceProviderSpec, entity *nigoapi.ControllerServiceEntity) bool {
	if cfg == nil {
		return true
	}
	return cfg.ConnectTimeout == entity.Component.Properties[StandardWebClientServiceProperties.ConnectTimeout] &&
		cfg.ReadTimeout == entity.Component.Properties[StandardWebClientServiceProperties.ReadTimeout] &&
		cfg.WriteTimeout == entity.Component.Properties[StandardWebClientServiceProperties.WriteTimeout]
}

// updateControllerServiceEntity writes the desired configuration into the entity. The properties
// map only carries the managed properties: NiFi merges the map it receives, so unmanaged
// properties keep their value.
func updateControllerServiceEntity(controllerService *v1alpha1.NifiControllerService, entity *nigoapi.ControllerServiceEntity) {
	var defaultVersion int64 = 0

	if entity.Revision == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.ControllerServiceDto{
			Type_: controllerService.Spec.GetType(),
		}
	}

	entity.Component.Name = controllerService.Name
	entity.Component.Comments = controllerService.Spec.Description
	entity.Component.Properties = make(map[string]string)

	switch controllerService.Spec.Type {
	case v1alpha1.StandardWebClientServiceProviderType:
		updateEntity_StandardWebClientServiceProvider(controllerService.Spec.StandardWebClientServiceProviderSpec, entity)
	}
}

func updateEntity_StandardWebClientServiceProvider(spec *v1alpha1.StandardWebClientServiceProviderSpec, entity *nigoapi.ControllerServiceEntity) {
	if spec == nil {
		return
	}
	entity.Component.Properties[StandardWebClientServiceProperties.ConnectTimeout] = spec.ConnectTimeout
	entity.Component.Properties[StandardWebClientServiceProperties.ReadTimeout] = spec.ReadTimeout
	entity.Component.Properties[StandardWebClientServiceProperties.WriteTimeout] = spec.WriteTimeout
}
