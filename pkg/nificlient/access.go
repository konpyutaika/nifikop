package nificlient

import (
	"github.com/antihax/optional"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
)

func (n *nifiClient) CreateAccessTokenUsingBasicAuth(username, password string, nodeId int32) (*string, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	// @TODO : force the targeted host, or recreate token for all nodes
	client := n.nodeClient[nodeId]
	context := n.opts.NodesContext[nodeId]
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the reporting task informations
	_, rsp, body, err := client.AccessApi.CreateAccessToken(context, &nigoapi.AccessApiCreateAccessTokenOpts{
		Username: optional.NewString(username),
		Password: optional.NewString(password),
	})

	if err := errorCreateOperation(rsp, body, err); err != nil {
		return nil, err
	}
	return body, nil
}
