package nificlient

import nigoapi "github.com/erdrix/nigoapi/pkg/nifi"

func (n *nifiClient) UpdateOutputPortRunStatus(id string, entity nigoapi.PortRunStatusEntity) (*nigoapi.ProcessorEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the output port run status
	processor, rsp, body, err := client.OutputPortsApi.UpdateRunStatus(context, id, entity)
	if err := errorUpdateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &processor, nil
}

func (n *nifiClient) GetOutputPort(id string) (*nigoapi.PortEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the output port informations
	port, rsp, body, err := client.OutputPortsApi.GetOutputPort(context, id)
	if err := errorUpdateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &port, nil
}
