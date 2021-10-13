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
