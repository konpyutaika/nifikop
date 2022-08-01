package nificlient

import (
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
)

func (n *nifiClient) GetConnection(id string) (*nigoapi.ConnectionEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the connection informations
	connectionEntity, rsp, body, err := client.ConnectionsApi.GetConnection(context, id)
	if err := errorGetOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &connectionEntity, nil
}
