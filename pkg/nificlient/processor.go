package nificlient

import nigoapi "github.com/erdrix/nigoapi/pkg/nifi"

func (n *nifiClient) UpdateProcessorRunStatus(
	id string,
	entity nigoapi.ProcessorRunStatusEntity) (*nigoapi.ProcessorEntity, error) {

	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the processor run status
	processor, rsp, body, err := client.ProcessorsApi.UpdateRunStatus(context, id, entity)
	if err := errorUpdateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &processor, nil
}
