// Copyright 2020 Orange SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package apis

package nificlient

import (
	"strconv"

	"github.com/antihax/optional"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
)

func (n *nifiClient) GetUserGroups() ([]nigoapi.UserGroupEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the user groups informations
	userGroupsEntity, rsp, err := client.TenantsApi.GetUserGroups(nil)

	if err := errorGetOperation(rsp, err); err != nil {
		return nil, err
	}

	return userGroupsEntity.UserGroups, nil
}

func (n *nifiClient) GetUserGroup(id string) (*nigoapi.UserGroupEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the user groups informations
	userGroupEntity, rsp, err := client.TenantsApi.GetUserGroup(nil, id)

	if err := errorGetOperation(rsp, err); err != nil {
		return nil, err
	}

	return &userGroupEntity, nil
}

func (n *nifiClient) CreateUserGroup(entity nigoapi.UserGroupEntity) (*nigoapi.UserGroupEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the user group
	userGroupEntity, rsp, err := client.TenantsApi.CreateUserGroup(nil, entity)
	if err := errorCreateOperation(rsp, err); err != nil {
		return nil, err
	}
	return &userGroupEntity, nil
}

func (n *nifiClient) UpdateUserGroup(entity nigoapi.UserGroupEntity) (*nigoapi.UserGroupEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the user group
	userGroupEntity, rsp, err := client.TenantsApi.UpdateUserGroup(nil, entity.Id, entity)
	if err := errorUpdateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &userGroupEntity, nil
}

func (n *nifiClient) RemoveUserGroup(entity nigoapi.UserGroupEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the user group
	_, rsp, err := client.TenantsApi.RemoveUserGroup(nil, entity.Id,
		&nigoapi.TenantsApiRemoveUserGroupOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, err)
}
