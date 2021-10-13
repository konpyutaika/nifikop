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
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the registy client informations
	regCliEntity, rsp, body, err := client.ControllerApi.GetRegistryClient(context, id)

	if err := errorGetOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &regCliEntity, nil
}

func (n *nifiClient) CreateRegistryClient(entity nigoapi.RegistryClientEntity) (*nigoapi.RegistryClientEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the registry client
	regCliEntity, rsp, body, err := client.ControllerApi.CreateRegistryClient(context, entity)
	if err := errorCreateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &regCliEntity, nil
}

func (n *nifiClient) UpdateRegistryClient(entity nigoapi.RegistryClientEntity) (*nigoapi.RegistryClientEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the registry client
	regCliEntity, rsp, body, err := client.ControllerApi.UpdateRegistryClient(context, entity.Id, entity)
	if err := errorUpdateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &regCliEntity, nil
}

func (n *nifiClient) RemoveRegistryClient(entity nigoapi.RegistryClientEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the registry client
	_, rsp, body, err := client.ControllerApi.DeleteRegistryClient(context, entity.Id,
		&nigoapi.ControllerApiDeleteRegistryClientOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, body, err)
}
