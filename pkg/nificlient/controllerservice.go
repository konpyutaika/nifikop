package nificlient

import (
	"strconv"

	"github.com/antihax/optional"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

// GetControllerService retrieves a controller service by id from the NiFi cluster.
func (n *nifiClient) GetControllerService(id string) (*nigoapi.ControllerServiceEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the controller service informations
	out, rsp, body, err := client.ControllerServicesApi.GetControllerService(context, id, nil)
	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}

// CreateControllerService creates a controller-level controller service, the scope visible to
// reporting tasks and flow registry clients.
func (n *nifiClient) CreateControllerService(entity nigoapi.ControllerServiceEntity) (*nigoapi.ControllerServiceEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the controller service
	out, rsp, body, err := client.ControllerApi.CreateControllerService(context, entity)
	if err := errorCreateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}

// UpdateControllerService updates the configuration of an existing controller service.
func (n *nifiClient) UpdateControllerService(entity nigoapi.ControllerServiceEntity) (*nigoapi.ControllerServiceEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the controller service
	out, rsp, body, err := client.ControllerServicesApi.UpdateControllerService(context, entity, entity.Id)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}

// RemoveControllerService deletes a controller service, passing the revision version required by
// NiFi optimistic locking.
func (n *nifiClient) RemoveControllerService(entity nigoapi.ControllerServiceEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the controller service
	_, rsp, body, err := client.ControllerServicesApi.RemoveControllerService(context, entity.Id,
		&nigoapi.ControllerServicesApiRemoveControllerServiceOpts{
			Version: optional.NewInterface(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, body, err, n.log)
}

// UpdateControllerServiceRunStatus requests a state transition (ENABLED or DISABLED) for a
// controller service. NiFi applies the transition asynchronously: the returned entity reflects
// the state at the time of the call, so callers must poll until the target state is reached.
func (n *nifiClient) UpdateControllerServiceRunStatus(entity nigoapi.ControllerServiceEntity, state string) (*nigoapi.ControllerServiceEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to change the controller service run status
	out, rsp, body, err := client.ControllerServicesApi.UpdateRunStatus2(context,
		nigoapi.ControllerServiceRunStatusEntity{
			Revision: entity.Revision,
			State:    state,
		}, entity.Id)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}

// GetControllerServiceReferences lists the components referencing a controller service.
func (n *nifiClient) GetControllerServiceReferences(id string) (*nigoapi.ControllerServiceReferencingComponentsEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the controller service references
	out, rsp, body, err := client.ControllerServicesApi.GetControllerServiceReferences(context, id)
	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}
