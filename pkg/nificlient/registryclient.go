package nificlient

import (
	"strconv"

	"github.com/antihax/optional"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

func (n *nifiClient) GetRegistryClient(id string) (*nigoapi.FlowRegistryClientEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the registry client informations
	regCliEntity, rsp, body, err := client.ControllerApi.GetFlowRegistryClient(context, id)

	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &regCliEntity, nil
}

func (n *nifiClient) CreateRegistryClient(entity nigoapi.FlowRegistryClientEntity) (*nigoapi.FlowRegistryClientEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the registry client
	regCliEntity, rsp, body, err := client.ControllerApi.CreateFlowRegistryClient(context, entity)
	if err := errorCreateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &regCliEntity, nil
}

func (n *nifiClient) UpdateRegistryClient(entity nigoapi.FlowRegistryClientEntity) (*nigoapi.FlowRegistryClientEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the registry client
	regCliEntity, rsp, body, err := client.ControllerApi.UpdateFlowRegistryClient(context, entity.Id, entity)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &regCliEntity, nil
}

func (n *nifiClient) RemoveRegistryClient(entity nigoapi.FlowRegistryClientEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the registry client
	_, rsp, body, err := client.ControllerApi.DeleteFlowRegistryClient(context, entity.Id,
		&nigoapi.ControllerApiDeleteFlowRegistryClientOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, body, err, n.log)
}
