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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/pki"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	httpContainerPort int32 = 443
	succeededNodeId   int32 = 4

	clusterName      = "test-cluster"
	clusterNamespace = "test-namespace"
)

type mockClient struct {
	client.Client
}

func testCluster(t *testing.T) *v1alpha1.NifiCluster {
	t.Helper()
	cluster := &v1alpha1.NifiCluster{}

	cluster.Name = clusterName
	cluster.Namespace = clusterNamespace
	cluster.Spec = v1alpha1.NifiClusterSpec{}

	cluster.Status.NodesState = make(map[string]v1alpha1.NodeState)
	cluster.Status.NodesState["1"] = v1alpha1.NodeState{
		GracefulActionState: v1alpha1.GracefulActionState{
			State: v1alpha1.GracefulDownscaleRunning,
		},
	}

	cluster.Status.NodesState["2"] = v1alpha1.NodeState{
		GracefulActionState: v1alpha1.GracefulActionState{
			State: v1alpha1.GracefulUpscaleRequired,
		},
	}

	cluster.Status.NodesState["3"] = v1alpha1.NodeState{
		GracefulActionState: v1alpha1.GracefulActionState{
			ActionStep: v1alpha1.RemoveStatus,
		},
	}

	cluster.Status.NodesState[fmt.Sprint(succeededNodeId)] = v1alpha1.NodeState{
		GracefulActionState: v1alpha1.GracefulActionState{
			State: v1alpha1.GracefulDownscaleSucceeded,
		},
	}

	cluster.Spec.ListenersConfig.InternalListeners = []v1alpha1.InternalListenerConfig{
		{Type: "https", ContainerPort: httpContainerPort},
		{Type: "http", ContainerPort: 8080},
		{Type: "cluster", ContainerPort: 8083},
		{Type: "s2s", ContainerPort: 8085},
	}
	return cluster
}

func testSecuredCluster(t *testing.T) *v1alpha1.NifiCluster {
	cluster := testCluster(t)
	cluster.Spec.ListenersConfig.SSLSecrets = &v1alpha1.SSLSecrets{
		PKIBackend: pki.MockBackend,
	}

	cluster.Spec.ClusterSecure = true
	return cluster
}

func TestClusterConfig(t *testing.T) {
	cluster := testCluster(t)
	testClusterConfig(t, cluster, false)
	cluster = testSecuredCluster(t)
	testClusterConfig(t, cluster, true)
}

func testClusterConfig(t *testing.T, cluster *v1alpha1.NifiCluster, expectedUseSSL bool) {
	assert := assert.New(t)
	conf, err := ClusterConfig(mockClient{}, cluster)
	assert.Nil(err)
	assert.Equal(expectedUseSSL, conf.UseSSL)

	if expectedUseSSL {
		assert.NotNil(conf.TLSConfig)
	} else {
		assert.Nil(conf.TLSConfig)
	}

	assert.Equal(
		fmt.Sprintf("%s-%s-node.%s-all-node.%s.svc.cluster.local:%d",
			clusterName, "%d", clusterName, clusterNamespace, httpContainerPort),
		conf.nodeURITemplate)

	assert.Equal(1, len(conf.NodesURI))
	assert.NotNil(conf.NodesURI[succeededNodeId])
	assert.Equal(
		fmt.Sprintf("%s-%d-node.%s-all-node.%s.svc.cluster.local:%d",
			clusterName, succeededNodeId, clusterName, clusterNamespace, httpContainerPort),
		conf.NodesURI[succeededNodeId])

	assert.Equal(
		fmt.Sprintf("%s-all-node.%s.svc.cluster.local:%d",
			clusterName, clusterNamespace, httpContainerPort),
		conf.NifiURI)
}

func TestUseSSL(t *testing.T) {
	assert := assert.New(t)

	cluster := testCluster(t)
	assert.Equal(false, useSSL(cluster))
	cluster = testSecuredCluster(t)
	assert.Equal(true, useSSL(cluster))
}

func TestGenerateNodesAddress(t *testing.T) {
	assert := assert.New(t)

	cluster := testCluster(t)
	nodesURI := generateNodesAddress(cluster)

	assert.Equal(1, len(nodesURI))
	assert.NotNil(nodesURI[succeededNodeId])
	assert.Equal(
		fmt.Sprintf("%s-%d-node.%s-all-node.%s.svc.cluster.local:%d",
			clusterName, succeededNodeId, clusterName, clusterNamespace, httpContainerPort),
		nodesURI[succeededNodeId])
}

func TestGenerateNodesURITemplate(t *testing.T) {
	assert := assert.New(t)

	cluster := testCluster(t)

	assert.Equal(
		fmt.Sprintf("%s-%s-node.%s-all-node.%s.svc.cluster.local:%d",
			clusterName, "%d", clusterName, clusterNamespace, httpContainerPort),
		generateNodesURITemplate(cluster))
}
