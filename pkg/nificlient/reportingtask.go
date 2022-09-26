package nificlient

import (
	"strconv"

	"github.com/antihax/optional"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

func (n *nifiClient) GetReportingTask(id string) (*nigoapi.ReportingTaskEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the reporting task informations
	out, rsp, body, err := client.ReportingTasksApi.GetReportingTask(context, id)

	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}

func (n *nifiClient) CreateReportingTask(entity nigoapi.ReportingTaskEntity) (*nigoapi.ReportingTaskEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the reporting task
	out, rsp, body, err := client.ControllerApi.CreateReportingTask(context, entity)
	if err := errorCreateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}

func (n *nifiClient) UpdateReportingTask(entity nigoapi.ReportingTaskEntity) (*nigoapi.ReportingTaskEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the reporting task
	out, rsp, body, err := client.ReportingTasksApi.UpdateReportingTask(context, entity.Id, entity)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}

func (n *nifiClient) UpdateRunStatusReportingTask(id string, entity nigoapi.ReportingTaskRunStatusEntity) (*nigoapi.ReportingTaskEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the reporting task
	out, rsp, body, err := client.ReportingTasksApi.UpdateRunStatus(context, id, entity)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}

func (n *nifiClient) RemoveReportingTask(entity nigoapi.ReportingTaskEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the reporting task
	_, rsp, body, err := client.ReportingTasksApi.RemoveReportingTask(context, entity.Id,
		&nigoapi.ReportingTasksApiRemoveReportingTaskOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, body, err, n.log)
}
