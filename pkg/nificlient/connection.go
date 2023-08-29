package nificlient

import (
	"strconv"

	"github.com/antihax/optional"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

func (n *nifiClient) GetConnection(id string) (*nigoapi.ConnectionEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the connection informations
	connectionEntity, rsp, body, err := client.ConnectionsApi.GetConnection(context, id)
	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &connectionEntity, nil
}

func (n *nifiClient) UpdateConnection(entity nigoapi.ConnectionEntity) (*nigoapi.ConnectionEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the connection informations
	connectionEntity, rsp, body, err := client.ConnectionsApi.UpdateConnection(context, entity.Id, entity)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &connectionEntity, nil
}

func (n *nifiClient) DeleteConnection(entity nigoapi.ConnectionEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to delete the connection
	_, rsp, body, err := client.ConnectionsApi.DeleteConnection(
		context,
		entity.Id,
		&nigoapi.ConnectionsApiDeleteConnectionOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, body, err, n.log)
}
