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

func (n *nifiClient) GetRegistryClient(id string) (*nigoapi.RegistryClientEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the registy client informations
	regCliEntity, rsp, err := client.ControllerApi.GetRegistryClient(nil, id)

	if err := errorGetOperation(rsp, err); err != nil {
		return nil, err
	}

	return &regCliEntity, nil
}

func (n *nifiClient) CreateRegistryClient(entity nigoapi.RegistryClientEntity) (*nigoapi.RegistryClientEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the registry client
	regCliEntity, rsp, err := client.ControllerApi.CreateRegistryClient(nil, entity)
	if err := errorCreateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &regCliEntity, nil
}

func (n *nifiClient) UpdateRegistryClient(entity nigoapi.RegistryClientEntity) (*nigoapi.RegistryClientEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the registry client
	regCliEntity, rsp, err := client.ControllerApi.UpdateRegistryClient(nil, entity.Id, entity)
	if err := errorUpdateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &regCliEntity, nil
}

func (n *nifiClient) RemoveRegistryClient(entity nigoapi.RegistryClientEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the registry client
	_, rsp, err := client.ControllerApi.DeleteRegistryClient(nil, entity.Id,
		&nigoapi.ControllerApiDeleteRegistryClientOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, err)
}
