package nificlient

import (
	"github.com/antihax/optional"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

func (n *nifiClient) CreateAccessTokenUsingBasicAuth(username, password string, nodeId int32) (*string, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	// @TODO : force the targeted host, or recreate token for all nodes
	client := n.nodeClient[nodeId]
	context := n.opts.NodesContext[nodeId]
	if client == nil {
		n.log.Error("Error during creating node client",
			zap.Int32("nodeId", nodeId),
			zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the reporting task informations
	_, rsp, body, err := client.AccessApi.CreateAccessToken(context, &nigoapi.AccessApiCreateAccessTokenOpts{
		Username: optional.NewString(username),
		Password: optional.NewString(password),
	})

	if err := errorCreateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}
	return body, nil
}
