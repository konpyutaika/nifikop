package nificlient

import (
	"strconv"

	"github.com/antihax/optional"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

func (n *nifiClient) GetProcessGroup(id string) (*nigoapi.ProcessGroupEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the process group informations
	pGEntity, rsp, body, err := client.ProcessGroupsApi.GetProcessGroup(context, id)
	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &pGEntity, nil
}

func (n *nifiClient) CreateProcessGroup(
	entity nigoapi.ProcessGroupEntity,
	pgParentId string) (*nigoapi.ProcessGroupEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the versioned process group
	pgEntity, rsp, body, err := client.ProcessGroupsApi.CreateProcessGroup(
		context,
		entity,
		pgParentId,
		&nigoapi.ProcessGroupsApiCreateProcessGroupOpts{ParameterContextHandlingStrategy: optional.NewString("KEEP_EXISTING")})
	if err := errorCreateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &pgEntity, nil
}

func (n *nifiClient) UpdateProcessGroup(entity nigoapi.ProcessGroupEntity) (*nigoapi.ProcessGroupEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Remove VersionControlInformation to avoid Cannot set Version Control Info because process group is already under version control
	entity.Component.VersionControlInformation = nil

	// Request on Nifi Rest API to update the versioned process group
	pgEntity, rsp, body, err := client.ProcessGroupsApi.UpdateProcessGroup(context, entity, entity.Id)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &pgEntity, nil
}

func (n *nifiClient) RemoveProcessGroup(entity nigoapi.ProcessGroupEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the versioned process group
	_, rsp, body, err := client.ProcessGroupsApi.RemoveProcessGroup(
		context,
		entity.Id,
		&nigoapi.ProcessGroupsApiRemoveProcessGroupOpts{
			Version: optional.NewInterface(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, body, err, n.log)
}

func (n *nifiClient) CreateConnection(entity nigoapi.ConnectionEntity) (*nigoapi.ConnectionEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create a connection
	conEntity, rsp, body, err := client.ProcessGroupsApi.CreateConnection(context, entity, entity.Component.ParentGroupId)
	if err := errorCreateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &conEntity, nil
}
