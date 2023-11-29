package nificlient

import (
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config/common"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

var (
	nodesId = map[int32]string{0: "12334456", 1: "12334456", 2: "12334456"}
)

type mockNiFiClient struct {
	NifiClient
	opts *clientconfig.NifiConfig

	newClient func(*nigoapi.Configuration) *nigoapi.APIClient
	failOpts  bool
}

func newMockOpts() *clientconfig.NifiConfig {
	return &clientconfig.NifiConfig{}
}

func newMockHttpClient(c *nigoapi.Configuration) *nigoapi.APIClient {
	client := nigoapi.NewAPIClient(c)
	httpmock.Activate()
	return client
}

func newMockClient() *nifiClient {
	return &nifiClient{
		log:       zap.NewNop(),
		opts:      newMockOpts(),
		newClient: newMockHttpClient,
	}
}

func NewMockNiFiClient() *nifiClient {
	return &nifiClient{
		log:       zap.NewNop(),
		opts:      newMockOpts(),
		newClient: newMockHttpClient,
	}
}

func NewMockNiFiClientFailOps() *mockNiFiClient {
	return &mockNiFiClient{
		opts:      newMockOpts(),
		newClient: newMockHttpClient,
		failOpts:  true,
	}
}

func MockGetClusterResponse(cluster *v1.NifiCluster, empty bool) map[string]interface{} {
	if empty {
		return make(map[string]interface{})
	}
	return map[string]interface{}{
		"cluster": map[string]interface{}{
			"nodes": []nigoapi.NodeDto{
				{
					NodeId:  nodesId[0],
					Address: nifiutil.GenerateRequestNiFiNodeHostnameFromCluster(0, cluster),
					ApiPort: httpContainerPort,
					Status:  string(v1.ConnectStatus),
				},
				{
					NodeId:  nodesId[1],
					Address: nifiutil.GenerateRequestNiFiNodeHostnameFromCluster(1, cluster),
					ApiPort: httpContainerPort,
					Status:  string(v1.DisconnectStatus),
				},
				{
					NodeId:  nodesId[2],
					Address: nifiutil.GenerateRequestNiFiNodeHostnameFromCluster(2, cluster),
					ApiPort: httpContainerPort,
					Status:  string(v1.OffloadStatus),
				},
			},
		},
	}
}

func MockGetNodeResponse(nodeId int32, cluster *v1.NifiCluster) interface{} {
	nodes := map[int32]map[string]interface{}{
		0: {
			"node": nigoapi.NodeDto{
				NodeId:  nodesId[0],
				Address: nifiutil.GenerateRequestNiFiNodeHostnameFromCluster(0, cluster),
				ApiPort: httpContainerPort,
				Status:  string(v1.ConnectStatus),
			},
		},
		1: {
			"node": nigoapi.NodeDto{
				NodeId:  nodesId[1],
				Address: nifiutil.GenerateRequestNiFiNodeHostnameFromCluster(1, cluster),
				ApiPort: httpContainerPort,
				Status:  string(v1.ConnectStatus),
			},
		},
		2: {
			"node": nigoapi.NodeDto{
				NodeId:  nodesId[2],
				Address: nifiutil.GenerateRequestNiFiNodeHostnameFromCluster(2, cluster),
				ApiPort: httpContainerPort,
				Status:  string(v1.ConnectStatus),
			},
		},
	}

	return nodes[nodeId]
}

func testClusterMock(t *testing.T) *v1.NifiCluster {
	t.Helper()
	cluster := &v1.NifiCluster{}

	cluster.Name = clusterName
	cluster.Namespace = clusterNamespace
	cluster.Spec = v1.NifiClusterSpec{}
	cluster.Spec.ListenersConfig = &v1.ListenersConfig{}

	cluster.Spec.Nodes = []v1.Node{
		{Id: 0},
		{Id: 1},
		{Id: 2},
	}

	cluster.Spec.ListenersConfig.InternalListeners = []v1.InternalListenerConfig{
		{Type: "http", ContainerPort: httpContainerPort},
		{Type: "cluster", ContainerPort: 8083},
		{Type: "s2s", ContainerPort: 8085},
	}
	return cluster
}

func configFromCluster(cluster *v1.NifiCluster) (*clientconfig.NifiConfig, error) {
	conf := common.ClusterConfig(cluster)
	return conf, nil
}
