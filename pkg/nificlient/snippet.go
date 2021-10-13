package nificlient

import nigoapi "github.com/erdrix/nigoapi/pkg/nifi"

func (n *nifiClient) CreateSnippet(entity nigoapi.SnippetEntity) (*nigoapi.SnippetEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the snippet
	snippetEntity, rsp, body, err := client.SnippetsApi.CreateSnippet(context, entity)
	if err := errorCreateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &snippetEntity, nil
}

func (n *nifiClient) UpdateSnippet(entity nigoapi.SnippetEntity) (*nigoapi.SnippetEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the snippet
	snippetEntity, rsp, body, err := client.SnippetsApi.UpdateSnippet(context, entity.Snippet.Id, entity)
	if err := errorUpdateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &snippetEntity, nil
}
