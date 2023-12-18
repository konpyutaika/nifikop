package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/pki"
)

const (
	httpContainerPort int32 = 443
	succeededNodeId   int32 = 4

	clusterName      = "test-cluster"
	clusterNamespace = "test-namespace"
)

func testCluster(t *testing.T) *v1.NifiCluster {
	t.Helper()
	cluster := &v1.NifiCluster{}

	cluster.Name = clusterName
	cluster.Namespace = clusterNamespace
	cluster.Spec = v1.NifiClusterSpec{}
	cluster.Spec.ListenersConfig = &v1.ListenersConfig{}

	cluster.Status.NodesState = make(map[string]v1.NodeState)
	cluster.Status.NodesState["1"] = v1.NodeState{
		GracefulActionState: v1.GracefulActionState{
			State: v1.GracefulDownscaleRunning,
		},
	}

	cluster.Status.NodesState["2"] = v1.NodeState{
		GracefulActionState: v1.GracefulActionState{
			State: v1.GracefulUpscaleRequired,
		},
	}

	cluster.Status.NodesState["3"] = v1.NodeState{
		GracefulActionState: v1.GracefulActionState{
			ActionStep: v1.RemoveStatus,
		},
	}

	cluster.Status.NodesState[fmt.Sprint(succeededNodeId)] = v1.NodeState{
		GracefulActionState: v1.GracefulActionState{
			State: v1.GracefulDownscaleSucceeded,
		},
	}

	cluster.Spec.ListenersConfig.InternalListeners = []v1.InternalListenerConfig{
		{Type: "https", ContainerPort: httpContainerPort},
		{Type: "http", ContainerPort: 8080},
		{Type: "cluster", ContainerPort: 8083},
		{Type: "s2s", ContainerPort: 8085},
		{Type: "load-balance", ContainerPort: 6342},
	}
	return cluster
}

func testSecuredCluster(t *testing.T) *v1.NifiCluster {
	cluster := testCluster(t)
	cluster.Spec.ListenersConfig.SSLSecrets = &v1.SSLSecrets{
		PKIBackend: pki.MockBackend,
	}

	return cluster
}

func TestClusterConfig(t *testing.T) {
	cluster := testCluster(t)
	testClusterConfig(t, cluster, false)
	cluster = testSecuredCluster(t)
	testClusterConfig(t, cluster, true)
}

func testClusterConfig(t *testing.T, cluster *v1.NifiCluster, expectedUseSSL bool) {
	assert := assert.New(t)
	conf := ClusterConfig(cluster)
	assert.Equal(expectedUseSSL, conf.UseSSL)

	// if expectedUseSSL {
	//	assert.NotNil(conf.TLSConfig)
	// } else {
	//	assert.Nil(conf.TLSConfig)
	//}

	assert.Equal(
		fmt.Sprintf("%s-%s-node.%s.svc.cluster.local:%d",
			clusterName, "%d", clusterNamespace, httpContainerPort),
		conf.NodeURITemplate)

	assert.Equal(1, len(conf.NodesURI))
	assert.NotNil(conf.NodesURI[succeededNodeId])
	assert.Equal(
		fmt.Sprintf("%s-%d-node.%s.svc.cluster.local:%d",
			clusterName, succeededNodeId, clusterNamespace, httpContainerPort),
		conf.NodesURI[succeededNodeId].RequestHost)

	assert.Equal(
		fmt.Sprintf("%s-all-node.%s.svc.cluster.local:%d",
			clusterName, clusterNamespace, httpContainerPort),
		conf.NifiURI)
}

func TestUseSSL(t *testing.T) {
	assert := assert.New(t)

	cluster := testCluster(t)
	assert.Equal(false, UseSSL(cluster))
	cluster = testSecuredCluster(t)
	assert.Equal(true, UseSSL(cluster))
}

func TestGenerateNodesAddress(t *testing.T) {
	assert := assert.New(t)

	cluster := testCluster(t)
	nodesURI := generateNodesAddress(cluster)

	assert.Equal(1, len(nodesURI))
	assert.NotNil(nodesURI[succeededNodeId])
	assert.Equal(
		fmt.Sprintf("%s-%d-node.%s.svc.cluster.local:%d",
			clusterName, succeededNodeId, clusterNamespace, httpContainerPort),
		nodesURI[succeededNodeId].RequestHost)
}

func TestGenerateNodesURITemplate(t *testing.T) {
	assert := assert.New(t)

	cluster := testCluster(t)

	assert.Equal(
		fmt.Sprintf("%s-%s-node.%s.svc.cluster.local:%d",
			clusterName, "%d", clusterNamespace, httpContainerPort),
		generateNodesURITemplate(cluster))
}
