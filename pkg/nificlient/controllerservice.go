package nificlient

import (
	"strconv"

	"github.com/antihax/optional"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

func (n *nifiClient) GetControllerService(id string) (*nigoapi.ControllerServiceEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}
	var opts *nigoapi.ControllerServicesApiGetControllerServiceOpts = nil
	// Request on Nifi Rest API to get the controller service informations
	out, rsp, body, err := client.ControllerServicesApi.GetControllerService(context, id, opts)

	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}

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

func (n *nifiClient) UpdateControllerService(entity nigoapi.ControllerServiceEntity) (*nigoapi.ControllerServiceEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the registry client
	out, rsp, body, err := client.ControllerServicesApi.UpdateControllerService(context, entity, entity.Id)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}

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
