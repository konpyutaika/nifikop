package nificlient

import nigoapi "github.com/erdrix/nigoapi/pkg/nifi"

func (n *nifiClient) CreateSnippet(entity nigoapi.SnippetEntity) (*nigoapi.SnippetEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the snippet
	snippetEntity, rsp, err := client.SnippetsApi.CreateSnippet(nil, entity)
	if err := errorCreateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &snippetEntity, nil
}

func (n *nifiClient) UpdateSnippet(entity nigoapi.SnippetEntity) (*nigoapi.SnippetEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the snippet
	snippetEntity, rsp, err := client.SnippetsApi.UpdateSnippet(nil, entity.Snippet.Id, entity)
	if err := errorUpdateOperation(rsp, err); err != nil {
		return nil, err
	}

	return &snippetEntity, nil
}
