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

func (n *nifiClient) GetReportingTask(id string) (*nigoapi.ReportingTaskEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the reporting task informations
	out, rsp, body, err := client.ReportingTasksApi.GetReportingTask(context, id)

	if err := errorGetOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &out, nil
}

func (n *nifiClient) CreateReportingTask(entity nigoapi.ReportingTaskEntity) (*nigoapi.ReportingTaskEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the reporting task
	out, rsp, body, err := client.ControllerApi.CreateReportingTask(context, entity)
	if err := errorCreateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &out, nil
}

func (n *nifiClient) UpdateReportingTask(entity nigoapi.ReportingTaskEntity) (*nigoapi.ReportingTaskEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the reporting task
	out, rsp, body, err := client.ReportingTasksApi.UpdateReportingTask(context, entity.Id, entity)
	if err := errorUpdateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &out, nil
}

func (n *nifiClient) UpdateRunStatusReportingTask(id string, entity nigoapi.ReportingTaskRunStatusEntity) (*nigoapi.ReportingTaskEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the reporting task
	out, rsp, body, err := client.ReportingTasksApi.UpdateRunStatus(context, id, entity)
	if err := errorUpdateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &out, nil
}

func (n *nifiClient) RemoveReportingTask(entity nigoapi.ReportingTaskEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the reporting task
	_, rsp, body, err := client.ReportingTasksApi.RemoveReportingTask(context, entity.Id,
		&nigoapi.ReportingTasksApiRemoveReportingTaskOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, body, err)
}
