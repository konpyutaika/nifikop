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
	"fmt"
	"github.com/Orange-OpenSource/nifikop/pkg/util/clientconfig"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	httpContainerPort int32 = 80
	succeededNodeId   int32 = 4

	clusterName      = "test-cluster"
	clusterNamespace = "test-namespace"
)

type mockClient struct {
	client.Client
}

var (
	nodeURITemplate = fmt.Sprintf("%s-%s-node.%s.svc.cluster.local:%s",
		clusterName, "%d", clusterNamespace, "%d")
	nifiURITemplate = "cluster-all-node.namespace.svc.cluster.local:%d"
)

func TestNew(t *testing.T) {
	opts := newMockOpts()
	if client := New(opts); client == nil {
		t.Error("Expected new client, got nil")
	}
}

func TestBuild(t *testing.T) {
	assert := assert.New(t)
	client := newMockClient()

	client.opts.NodesURI = make(map[int32]clientconfig.NodeUri)
	client.opts.NodesURI[1] = clientconfig.NodeUri{
		HostListener: fmt.Sprintf(nodeURITemplate, 1, httpContainerPort),
		RequestHost:  fmt.Sprintf(nodeURITemplate, 1, httpContainerPort),
	}
	client.opts.NodeURITemplate = nodeURITemplate
	client.opts.NifiURI = fmt.Sprintf(nifiURITemplate, httpContainerPort)

	url := "http://" + fmt.Sprintf(nodeURITemplate, 1, httpContainerPort) + "/nifi-api/controller/cluster"

	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				200,
				map[string]interface{}{
					"cluster": map[string]interface{}{
						"nodes": []interface{}{
							[]nigoapi.NodeDto{
								{
									NodeId:  "1234556",
									Address: fmt.Sprintf(nodeURITemplate, 1, httpContainerPort),
									ApiPort: httpContainerPort,
									Status:  string(v1alpha1.ConnectStatus),
								},
							},
						},
					},
				})
		})

	err := client.Build()
	assert.Nil(err)

	httpmock.DeactivateAndReset()

	err = client.Build()
	assert.IsType(errorfactory.NodesUnreachable{}, err)
}
