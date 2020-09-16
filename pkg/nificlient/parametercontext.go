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

func (n *nifiClient) GetParameterContext(id string) (*nigoapi.ParameterContextEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the parameter context informations
	pcEntity, rsp, err := client.ParameterContextsApi.GetParameterContext(nil, id)
	if err := errorGetOperation(rsp, err); err != nil {
		return nil, err
	}

	return &pcEntity, nil
}

func (n *nifiClient) CreateParameterContext(entity nigoapi.ParameterContextEntity) (*nigoapi.ParameterContextEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the parameter context
	pcEntity, rsp, err := client.ParameterContextsApi.CreateParameterContext(nil, entity)
	if err := errorCreateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &pcEntity, nil
}

func (n *nifiClient) RemoveParameterContext(entity nigoapi.ParameterContextEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the parameter context
	_, rsp, err := client.ParameterContextsApi.DeleteParameterContext(nil, entity.Id,
		&nigoapi.ParameterContextsApiDeleteParameterContextOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, err)
}

func (n *nifiClient) CreateParameterContextUpdateRequest(contextId string, entity nigoapi.ParameterContextEntity) (*nigoapi.ParameterContextUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the parameter context update request
	request, rsp, err := client.ParameterContextsApi.SubmitParameterContextUpdate(nil, contextId, entity)
	if err := errorUpdateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &request, nil
}

func (n *nifiClient) GetParameterContextUpdateRequest(contextId, id string) (*nigoapi.ParameterContextUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the parameter context update request information
	request, rsp, err := client.ParameterContextsApi.GetParameterContextUpdate(nil, contextId, id)
	if err := errorGetOperation(rsp, err); err != nil {
		return nil, err
	}

	return &request, nil
}
