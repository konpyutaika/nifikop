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

func (n *nifiClient) GetUsers() ([]nigoapi.UserEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the users informations
	usersEntity, rsp, err := client.TenantsApi.GetUsers(nil)

	if err := errorGetOperation(rsp, err); err != nil {
		return nil, err
	}

	return usersEntity.Users, nil
}

func (n *nifiClient) GetUser(id string) (*nigoapi.UserEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the user informations
	userEntity, rsp, err := client.TenantsApi.GetUser(nil, id)

	if err := errorGetOperation(rsp, err); err != nil {
		return nil, err
	}

	return &userEntity, nil
}

func (n *nifiClient) CreateUser(entity nigoapi.UserEntity) (*nigoapi.UserEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the user
	userEntity, rsp, err := client.TenantsApi.CreateUser(nil, entity)
	if err := errorCreateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &userEntity, nil
}

func (n *nifiClient) UpdateUser(entity nigoapi.UserEntity) (*nigoapi.UserEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the user
	userEntity, rsp, err := client.TenantsApi.UpdateUser(nil, entity.Id, entity)
	if err := errorUpdateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &userEntity, nil
}

func (n *nifiClient) RemoveUser(entity nigoapi.UserEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the user
	_, rsp, err := client.TenantsApi.RemoveUser(nil, entity.Id,
		&nigoapi.TenantsApiRemoveUserOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, err)
}
