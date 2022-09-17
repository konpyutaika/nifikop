package nificlient

import (
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

func (n *nifiClient) CreateVersionUpdateRequest(pgId string, entity nigoapi.VersionControlInformationEntity) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the version update request
	request, rsp, body, err := client.VersionsApi.InitiateVersionControlUpdate(context, pgId, entity)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &request, nil
}

func (n *nifiClient) GetVersionUpdateRequest(id string) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the update request information
	request, rsp, body, err := client.VersionsApi.GetUpdateRequest(context, id)
	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &request, nil
}

func (n *nifiClient) CreateVersionRevertRequest(pgId string, entity nigoapi.VersionControlInformationEntity) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the version revert request
	request, rsp, body, err := client.VersionsApi.InitiateRevertFlowVersion(context, pgId, entity)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &request, nil
}

func (n *nifiClient) GetVersionRevertRequest(id string) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the revert request information
	request, rsp, body, err := client.VersionsApi.GetRevertRequest(context, id)
	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &request, nil
}
