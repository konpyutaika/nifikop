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

func (n *nifiClient) GetAccessPolicy(action, resource string) (*nigoapi.AccessPolicyEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the access policy informations

	for true {
		if resource[0:1] == "/" {
			resource = resource[1:]
			continue
		}
		break
	}

	accessPolicyEntity, rsp, err := client.PoliciesApi.GetAccessPolicyForResource(nil, action, resource)

	if err := errorGetOperation(rsp, err); err != nil {
		return nil, err
	}

	return &accessPolicyEntity, nil
}

func (n *nifiClient) CreateAccessPolicy(entity nigoapi.AccessPolicyEntity) (*nigoapi.AccessPolicyEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the access policy
	accessPolicyEntity, rsp, err := client.PoliciesApi.CreateAccessPolicy(nil, entity)
	if err := errorCreateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &accessPolicyEntity, nil
}

func (n *nifiClient) UpdateAccessPolicy(entity nigoapi.AccessPolicyEntity) (*nigoapi.AccessPolicyEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the access policy
	accessPolicyEntity, rsp, err := client.PoliciesApi.UpdateAccessPolicy(nil, entity.Id, entity)
	if err := errorUpdateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &accessPolicyEntity, nil
}

func (n *nifiClient) RemoveAccessPolicy(entity nigoapi.AccessPolicyEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the registry client
	_, rsp, err := client.PoliciesApi.RemoveAccessPolicy(nil, entity.Id,
		&nigoapi.PoliciesApiRemoveAccessPolicyOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, err)
}