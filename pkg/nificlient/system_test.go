package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/konpyutaika/nifikop/api/v1"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

func TestDescribeCluster(t *testing.T) {
	assert := assert.New(t)

	clusterEntity, err := testDescribeCluster(t, 200)
	assert.Nil(err)
	assert.NotNil(clusterEntity)

	clusterEntity, err = testDescribeCluster(t, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(clusterEntity)

	clusterEntity, err = testDescribeCluster(t, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(clusterEntity)
}

func testDescribeCluster(t *testing.T, status int) (*nigoapi.ClusterEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, "/controller/cluster")
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockGetClusterResponse(cluster, false))
		})

	return client.DescribeCluster()
}

func TestGetClusterNode(t *testing.T) {
	assert := assert.New(t)

	cluster := testClusterMock(t)

	for _, node := range cluster.Spec.Nodes {
		nodeEntity, err := testGetClusterNode(t, cluster, node.Id, 200)
		assert.Nil(err)
		assert.NotNil(nodeEntity)
	}

	nodeEntity, err := testGetClusterNode(t, cluster, 10, 200)
	assert.IsType(ErrNifiClusterNodeNotFound, err)
	assert.Nil(nodeEntity)

	nodeEntity, err = testGetClusterNode(t, cluster, 0, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(nodeEntity)
}

func testGetClusterNode(t *testing.T, cluster *v1.NifiCluster, nodeId int32, status int) (*nigoapi.NodeEntity, error) {
	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller/cluster/nodes/%s", nodesId[nodeId]))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockGetNodeResponse(nodeId, cluster))
		})

	return client.GetClusterNode(nodeId)
}

func TestDisconnectClusterNode(t *testing.T) {
	assert := assert.New(t)

	cluster := testClusterMock(t)

	for _, node := range cluster.Spec.Nodes {
		nodeEntity, err := testDisconnectClusterNode(t, cluster, node.Id, 200)
		assert.Nil(err)
		assert.NotNil(nodeEntity)
	}

	nodeEntity, err := testDisconnectClusterNode(t, cluster, 10, 200)
	assert.IsType(ErrNifiClusterNodeNotFound, err)
	assert.Nil(nodeEntity)

	nodeEntity, err = testDisconnectClusterNode(t, cluster, 1, 500)
	assert.Nil(err)
	assert.NotNil(nodeEntity)

	nodeEntity, err = testDisconnectClusterNode(t, cluster, 0, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(nodeEntity)
}

func testDisconnectClusterNode(t *testing.T, cluster *v1.NifiCluster, nodeId int32, status int) (*nigoapi.NodeEntity, error) {
	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller/cluster/nodes/%s", nodesId[nodeId]))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockGetNodeResponse(nodeId, cluster))
		})

	return client.DisconnectClusterNode(nodeId)
}

func TestConnectClusterNode(t *testing.T) {
	assert := assert.New(t)

	cluster := testClusterMock(t)

	for _, node := range cluster.Spec.Nodes {
		nodeEntity, err := testConnectClusterNode(t, cluster, node.Id, 200)
		assert.Nil(err)
		assert.NotNil(nodeEntity)
	}

	nodeEntity, err := testConnectClusterNode(t, cluster, 10, 200)
	assert.IsType(ErrNifiClusterNodeNotFound, err)
	assert.Nil(nodeEntity)

	nodeEntity, err = testConnectClusterNode(t, cluster, 0, 500)
	assert.Nil(err)
	assert.NotNil(nodeEntity)

	nodeEntity, err = testConnectClusterNode(t, cluster, 1, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(nodeEntity)
}

func testConnectClusterNode(t *testing.T, cluster *v1.NifiCluster, nodeId int32, status int) (*nigoapi.NodeEntity, error) {
	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller/cluster/nodes/%s", nodesId[nodeId]))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockGetNodeResponse(nodeId, cluster))
		})

	return client.ConnectClusterNode(nodeId)
}

func TestOffloadClusterNode(t *testing.T) {
	assert := assert.New(t)

	cluster := testClusterMock(t)

	for _, node := range cluster.Spec.Nodes {
		nodeEntity, err := testOffloadClusterNode(t, cluster, node.Id, 200)
		assert.Nil(err)
		assert.NotNil(nodeEntity)
	}

	nodeEntity, err := testOffloadClusterNode(t, cluster, 10, 200)
	assert.IsType(ErrNifiClusterNodeNotFound, err)
	assert.Nil(nodeEntity)

	nodeEntity, err = testOffloadClusterNode(t, cluster, 2, 500)
	assert.Nil(err)
	assert.NotNil(nodeEntity)

	nodeEntity, err = testOffloadClusterNode(t, cluster, 1, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(nodeEntity)
}

func testOffloadClusterNode(t *testing.T, cluster *v1.NifiCluster, nodeId int32, status int) (*nigoapi.NodeEntity, error) {
	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller/cluster/nodes/%s", nodesId[nodeId]))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockGetNodeResponse(nodeId, cluster))
		})

	return client.OffloadClusterNode(nodeId)
}

func TestRemoveClusterNode(t *testing.T) {
	assert := assert.New(t)

	cluster := testClusterMock(t)

	for _, node := range cluster.Spec.Nodes {
		err := testRemoveClusterNode(t, cluster, node.Id, 200)
		assert.Nil(err)
	}

	err := testRemoveClusterNode(t, cluster, 10, 404)
	assert.IsType(ErrNifiClusterNodeNotFound, err)

	err = testRemoveClusterNode(t, cluster, 1, 404)
	assert.Nil(err)

	err = testRemoveClusterNode(t, cluster, 1, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
}

func testRemoveClusterNode(t *testing.T, cluster *v1.NifiCluster, nodeId int32, status int) error {
	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller/cluster/nodes/%s", nodesId[nodeId]))
	httpmock.RegisterResponder(http.MethodDelete, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockGetNodeResponse(nodeId, cluster))
		})

	return client.RemoveClusterNode(nodeId)
}

func TestRemoveClusterNodeFromClusterNodeId(t *testing.T) {
	assert := assert.New(t)

	cluster := testClusterMock(t)

	for _, node := range cluster.Spec.Nodes {
		err := testRemoveClusterNode(t, cluster, node.Id, 200)
		assert.Nil(err)
	}

	err := testRemoveClusterNodeFromClusterNodeId(t, cluster, 10, 404)
	assert.Nil(err)

	err = testRemoveClusterNodeFromClusterNodeId(t, cluster, 1, 404)
	assert.Nil(err)

	err = testRemoveClusterNodeFromClusterNodeId(t, cluster, 1, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
}

func testRemoveClusterNodeFromClusterNodeId(t *testing.T, cluster *v1.NifiCluster, nodeId int32, status int) error {
	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller/cluster/nodes/%s", nodesId[nodeId]))
	httpmock.RegisterResponder(http.MethodDelete, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockGetNodeResponse(nodeId, cluster))
		})

	return client.RemoveClusterNodeFromClusterNodeId(nodesId[nodeId])
}

func testClientFromCluster(cluster *v1.NifiCluster, empty bool) (NifiClient, error) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	url := nifiAddress(cluster, "/controller/cluster")

	cfg, _ := configFromCluster(cluster)
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				200,
				MockGetClusterResponse(cluster, empty))
		})

	cli, err := NewFromConfig(cfg, zap.NewNop())
	return cli, err
}

func nifiAddress(cluster *v1.NifiCluster, path string) string {
	return fmt.Sprintf("http://%s/nifi-api%s", nifiutil.GenerateRequestNiFiAllNodeAddressFromCluster(cluster), path)
}
