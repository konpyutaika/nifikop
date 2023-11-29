package nificlient

import (
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

func (n *nifiClient) GetDropRequest(connectionId, id string) (*nigoapi.DropRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the drop request information
	dropRequest, rsp, body, err := client.FlowfileQueuesApi.GetDropRequest(context, connectionId, id)
	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &dropRequest, nil
}

func (n *nifiClient) CreateDropRequest(connectionId string) (*nigoapi.DropRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the drop Request
	entity, rsp, body, err := client.FlowfileQueuesApi.CreateDropRequest(context, connectionId)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &entity, nil
}

// TODO : when last supported will be NiFi 1.12.X
// func (n *nifiClient) CreateDropRequest(pgId string)(*nigoapi.ProcessGroupEntity, error) {
//	// Get nigoapi client, favoring the one associated to the coordinator node.
//	client, context := n.privilegeCoordinatorClient()
//	if client == nil {
//		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
//		return nil, ErrNoNodeClientsAvailable
//	}
//
//	// Request on Nifi Rest API to create the registry client
//	entity, rsp, err := client.ProcessGroupsApi.CreateEmptyAllConnectionsRequest(context, pgId)
//	if err := errorCreateOperation(rsp, err); err != nil {
//		return nil, err
//	}
//
//	return &entity, nil
//}
