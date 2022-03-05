package nificlient

import (
	"strconv"

	"github.com/antihax/optional"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
)

func (n *nifiClient) GetUserGroups() ([]nigoapi.UserGroupEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the user groups informations
	userGroupsEntity, rsp, body, err := client.TenantsApi.GetUserGroups(context)

	if err := errorGetOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return userGroupsEntity.UserGroups, nil
}

func (n *nifiClient) GetUserGroup(id string) (*nigoapi.UserGroupEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the user groups informations
	userGroupEntity, rsp, body, err := client.TenantsApi.GetUserGroup(context, id)

	if err := errorGetOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &userGroupEntity, nil
}

func (n *nifiClient) CreateUserGroup(entity nigoapi.UserGroupEntity) (*nigoapi.UserGroupEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the user group
	userGroupEntity, rsp, body, err := client.TenantsApi.CreateUserGroup(context, entity)
	if err := errorCreateOperation(rsp, body, err); err != nil {
		return nil, err
	}
	return &userGroupEntity, nil
}

func (n *nifiClient) UpdateUserGroup(entity nigoapi.UserGroupEntity) (*nigoapi.UserGroupEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the user group
	userGroupEntity, rsp, body, err := client.TenantsApi.UpdateUserGroup(context, entity.Id, entity)
	if err := errorUpdateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &userGroupEntity, nil
}

func (n *nifiClient) RemoveUserGroup(entity nigoapi.UserGroupEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the user group
	_, rsp, body, err := client.TenantsApi.RemoveUserGroup(context, entity.Id,
		&nigoapi.TenantsApiRemoveUserGroupOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, body, err)
}
