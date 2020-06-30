package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/apis/nifi/v1alpha1"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/errorfactory"
	nifiutil "gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/util/nifi"
)

var (
	nodeURITemplate = fmt.Sprintf("%s-%s-node.%s-all-node.%s.svc.cluster.local:%s",
		clusterName, "%d", clusterName, clusterNamespace, "%d")
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

	client.opts.NodesURI        = make(map[int32]string)
	client.opts.NodesURI[1]     = fmt.Sprintf(nodeURITemplate,1, httpContainerPort)
	client.opts.nodeURITemplate = nodeURITemplate
	client.opts.NifiURI         = fmt.Sprintf(nifiURITemplate, httpContainerPort)

	url := "http://" + fmt.Sprintf(nodeURITemplate,1, httpContainerPort) + "/nifi-api/controller/cluster"

	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				200,
				map[string]interface{}{
					"cluster": map[string]interface{}{
						"nodes": []interface{}{
							[]nigoapi.NodeDto {
								{
									NodeId:  "1234556",
									Address: fmt.Sprintf(nodeURITemplate,1, httpContainerPort),
									ApiPort: httpContainerPort,
									Status:  string(v1alpha1.ConnectStatus),
								},
							},
						},
					},
				})
		})

	err := client.Build();
	assert.Nil(err)

	httpmock.DeactivateAndReset()

	err = client.Build()
	assert.IsType(errorfactory.NodesUnreachable{}, err)
}

func TestNewFromCluster(t *testing.T) {
	httpmock.Activate()
	assert := assert.New(t)

	cluster := testClusterMock(t)

	url := fmt.Sprintf("http://%s/nifi-api/controller/cluster", nifiutil.GenerateNiFiAddressFromCluster(cluster))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				200,
				MockGetClusterResponse(cluster))
		})

	_, err :=  NewFromCluster(mockClient{}, cluster)
	assert.Nil(err)

	httpmock.DeactivateAndReset()
	_, err =  NewFromCluster(mockClient{}, cluster)
	assert.IsType(errorfactory.NodesUnreachable{}, err)

}