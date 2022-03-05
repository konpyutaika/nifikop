package nificlient

import (
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
)

func (n *nifiClient) GetControllerConfig() (*nigoapi.ControllerConfigurationEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the reporting task informations

	out, rsp, body, err := client.ControllerApi.GetControllerConfig(context)

	if err := errorGetOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &out, nil
}

func (n *nifiClient) UpdateControllerConfig(entity nigoapi.ControllerConfigurationEntity) (*nigoapi.ControllerConfigurationEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the reporting task
	out, rsp, body, err := client.ControllerApi.UpdateControllerConfig(context, entity)
	if err := errorUpdateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &out, nil
}
