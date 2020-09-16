package nificlient

import nigoapi "github.com/erdrix/nigoapi/pkg/nifi"

func (n *nifiClient) CreateVersionUpdateRequest(pgId string, entity nigoapi.VersionControlInformationEntity) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the version update request
	request, rsp, err := client.VersionsApi.InitiateVersionControlUpdate(nil, pgId, entity)
	if err := errorUpdateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &request, nil
}

func (n *nifiClient) GetVersionUpdateRequest(id string) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the update request information
	request, rsp, err := client.VersionsApi.GetUpdateRequest(nil, id)
	if err := errorGetOperation(rsp, err); err != nil {
		return nil, err
	}

	return &request, nil
}

func (n *nifiClient) CreateVersionRevertRequest(pgId string, entity nigoapi.VersionControlInformationEntity) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the version revert request
	request, rsp, err := client.VersionsApi.InitiateRevertFlowVersion(nil, pgId, entity)
	if err := errorUpdateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &request, nil
}

func (n *nifiClient) GetVersionRevertRequest(id string) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the revert request information
	request, rsp, err := client.VersionsApi.GetRevertRequest(nil, id)
	if err := errorGetOperation(rsp, err); err != nil {
		return nil, err
	}

	return &request, nil
}
