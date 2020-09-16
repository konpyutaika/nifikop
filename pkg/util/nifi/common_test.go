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

package nifi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
)

const (
	httpContainerPort int32 = 443

	clusterName      = "test-cluster"
	clusterNamespace = "test-namespace"

	localClusterDomain       = "cluster.local"
	demoClusterDomain        = "demo.local"
	externalDNSClusterDomain = "external.dns.com"
)

func TestParseStringToInt32(t *testing.T) {
	assert := assert.New(t)

	var expectedValue int32 = 12
	parsed, _ := ParseStringToInt32("12")

	assert.Equal(expectedValue, parsed)
}

func testClusterLocal(t *testing.T) *v1alpha1.NifiCluster {
	t.Helper()
	cluster := &v1alpha1.NifiCluster{}

	cluster.Name = clusterName
	cluster.Namespace = clusterNamespace
	cluster.Spec = v1alpha1.NifiClusterSpec{}

	cluster.Spec.Nodes = []v1alpha1.Node{
		{Id: 0},
		{Id: 1},
		{Id: 2},
	}

	cluster.Spec.ListenersConfig.InternalListeners = []v1alpha1.InternalListenerConfig{
		{Type: "https", ContainerPort: httpContainerPort},
		{Type: "http", ContainerPort: 8080},
		{Type: "cluster", ContainerPort: 8083},
		{Type: "s2s", ContainerPort: 8085},
	}
	return cluster
}

func testClusterDemo(t *testing.T) *v1alpha1.NifiCluster {
	t.Helper()
	cluster := testClusterLocal(t)
	cluster.Spec.ListenersConfig.ClusterDomain = demoClusterDomain
	return cluster
}

func testClusterExternalDNS(t *testing.T) *v1alpha1.NifiCluster {
	t.Helper()
	cluster := testClusterLocal(t)
	cluster.Spec.ListenersConfig.ClusterDomain = externalDNSClusterDomain
	cluster.Spec.ListenersConfig.UseExternalDNS = true
	return cluster
}

func TestGenerateNiFiAddressFromCluster(t *testing.T) {

	testNiFiAddressFromCluster(t, testClusterLocal(t), localClusterDomain, false)
	testNiFiAddressFromCluster(t, testClusterDemo(t), demoClusterDomain, false)
	testNiFiAddressFromCluster(t, testClusterExternalDNS(t), externalDNSClusterDomain, true)
}

func testNiFiAddressFromCluster(t *testing.T,
	cluster *v1alpha1.NifiCluster, expectedClusterDomain string, expectedUseExternalDNS bool) {

	assert := assert.New(t)

	// Test headless service
	cluster.Spec.Service.HeadlessEnabled = true
	assert.Equal(
		fmt.Sprintf("%s:%d", ComputeNiFiHostname(clusterName, clusterNamespace,
			true, expectedClusterDomain, expectedUseExternalDNS),
			httpContainerPort),
		GenerateNiFiAddressFromCluster(cluster))

	// Test all nodes service
	cluster.Spec.Service.HeadlessEnabled = false
	assert.Equal(
		fmt.Sprintf("%s:%d", ComputeNiFiHostname(clusterName, clusterNamespace,
			false, expectedClusterDomain, expectedUseExternalDNS),
			httpContainerPort),
		GenerateNiFiAddressFromCluster(cluster))
}

func TestComputeNiFiAddress(t *testing.T) {

	cluster := testClusterLocal(t)
	testNiFiAddress(t,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.ListenersConfig.InternalListeners,
		localClusterDomain, false)

	cluster = testClusterDemo(t)
	testNiFiAddress(t,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.ListenersConfig.InternalListeners,
		demoClusterDomain, false)

	cluster = testClusterExternalDNS(t)
	testNiFiAddress(t,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.ListenersConfig.InternalListeners,
		externalDNSClusterDomain, true)

}

func testNiFiAddress(t *testing.T,
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool,
	internalListeners []v1alpha1.InternalListenerConfig,
	expectedClusterDomain string, expectedUseExternalDNS bool) {

	assert := assert.New(t)

	assert.Equal(
		fmt.Sprintf("%s:%d", ComputeNiFiHostname(
			clusterName, clusterNamespace, headlessServiceEnabled,
			expectedClusterDomain, expectedUseExternalDNS), httpContainerPort),
		ComputeNiFiAddress(clusterName, namespace, headlessServiceEnabled, clusterDomain, useExternalDNS, internalListeners))

}

func TestComputeNiFiHostname(t *testing.T) {
	cluster := testClusterLocal(t)
	testComputeNiFiHostname(t,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		localClusterDomain, false)

	cluster = testClusterDemo(t)
	testComputeNiFiHostname(t,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		demoClusterDomain, false)

	cluster = testClusterExternalDNS(t)
	testComputeNiFiHostname(t,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		externalDNSClusterDomain, true)
}

func testComputeNiFiHostname(t *testing.T,
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool,
	ExpectedClusterDomain string, ExpectedUseExternalDNS bool) {

	assert := assert.New(t)

	expectedHeadless, expectedAllNodes := computeNiFiHostnames(ExpectedClusterDomain, ExpectedUseExternalDNS)

	if headlessServiceEnabled {
		assert.Equal(
			expectedHeadless,
			ComputeNiFiHostname(clusterName, namespace, headlessServiceEnabled, clusterDomain, useExternalDNS))
	} else {
		assert.Equal(
			expectedAllNodes,
			ComputeNiFiHostname(clusterName, namespace, headlessServiceEnabled, clusterDomain, useExternalDNS))
	}
}

func computeNiFiHostnames(clusterDomain string, useExternalDNS bool) (string, string) {
	svc := ""
	if !useExternalDNS {
		svc = fmt.Sprintf(".%s.svc", clusterNamespace)
	}

	return fmt.Sprintf("%s-headless%s.%s", clusterName, svc, clusterDomain),
		fmt.Sprintf("%s-all-node%s.%s", clusterName, svc, clusterDomain)
}

func TestComputeServiceNameFull(t *testing.T) {
	assert := assert.New(t)

	cluster := testClusterLocal(t)
	cluster.Spec.Service.HeadlessEnabled = true
	assert.Equal(fmt.Sprintf("%s.%s.svc", ComputeServiceName(clusterName, true, ), clusterNamespace),
		ComputeServiceNameFull(cluster.Name, cluster.Namespace,
			cluster.Spec.Service.HeadlessEnabled, cluster.Spec.ListenersConfig.UseExternalDNS))

	cluster = testClusterExternalDNS(t)
	cluster.Spec.Service.HeadlessEnabled = true
	assert.Equal(ComputeServiceName(clusterName, true),
		ComputeServiceNameFull(cluster.Name, cluster.Namespace,
			cluster.Spec.Service.HeadlessEnabled, cluster.Spec.ListenersConfig.UseExternalDNS))
}

func TestComputeServiceNameWithNamespace(t *testing.T) {
	assert := assert.New(t)

	cluster := testClusterLocal(t)
	cluster.Spec.Service.HeadlessEnabled = true
	assert.Equal(fmt.Sprintf("%s.%s", ComputeServiceName(clusterName, true, ), clusterNamespace),
		ComputeServiceNameWithNamespace(cluster.Name, cluster.Namespace,
			cluster.Spec.Service.HeadlessEnabled, cluster.Spec.ListenersConfig.UseExternalDNS))

	cluster = testClusterExternalDNS(t)
	cluster.Spec.Service.HeadlessEnabled = true
	assert.Equal(ComputeServiceName(clusterName, true),
		ComputeServiceNameWithNamespace(cluster.Name, cluster.Namespace,
			cluster.Spec.Service.HeadlessEnabled, cluster.Spec.ListenersConfig.UseExternalDNS))
}

func TestGenerateNodeAddressFromCluster(t *testing.T) {
	assert := assert.New(t)

	clusters := []*v1alpha1.NifiCluster{
		testClusterLocal(t),
		testClusterDemo(t),
		testClusterExternalDNS(t),
	}

	for _, cluster := range clusters {
		for _, node := range cluster.Spec.Nodes {
			assert.Equal(
				fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id),
					GenerateNiFiAddressFromCluster(cluster)),
				GenerateNodeAddressFromCluster(node.Id, cluster))
		}
	}
}

func TestComputeNodeAddress(t *testing.T) {
	assert := assert.New(t)

	clusters := []*v1alpha1.NifiCluster{
		testClusterLocal(t),
		testClusterDemo(t),
		testClusterExternalDNS(t),
	}

	for _, cluster := range clusters {
		for _, node := range cluster.Spec.Nodes {
			assert.Equal(
				fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id),
					ComputeNiFiAddress(cluster.Name,
						cluster.Namespace,
						cluster.Spec.Service.HeadlessEnabled,
						cluster.Spec.ListenersConfig.GetClusterDomain(),
						cluster.Spec.ListenersConfig.UseExternalDNS,
						cluster.Spec.ListenersConfig.InternalListeners)),
				ComputeNodeAddress(node.Id, cluster.Name,
					cluster.Namespace,
					cluster.Spec.Service.HeadlessEnabled,
					cluster.Spec.ListenersConfig.GetClusterDomain(),
					cluster.Spec.ListenersConfig.UseExternalDNS,
					cluster.Spec.ListenersConfig.InternalListeners))
		}
	}
}

func TestComputeNodeHostname(t *testing.T) {
	assert := assert.New(t)

	clusters := []*v1alpha1.NifiCluster{
		testClusterLocal(t),
		testClusterDemo(t),
		testClusterExternalDNS(t),
	}

	for _, cluster := range clusters {
		for _, node := range cluster.Spec.Nodes {
			assert.Equal(
				fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id),
					ComputeNiFiHostname(cluster.Name,
						cluster.Namespace,
						cluster.Spec.Service.HeadlessEnabled,
						cluster.Spec.ListenersConfig.GetClusterDomain(),
						cluster.Spec.ListenersConfig.UseExternalDNS)),
				ComputeNodeHostname(node.Id, cluster.Name,
					cluster.Namespace,
					cluster.Spec.Service.HeadlessEnabled,
					cluster.Spec.ListenersConfig.GetClusterDomain(),
					cluster.Spec.ListenersConfig.UseExternalDNS))
		}
	}
}

func TestComputeNodeServiceNameFull(t *testing.T) {
	assert := assert.New(t)

	clusters := []*v1alpha1.NifiCluster{
		testClusterLocal(t),
		testClusterDemo(t),
		testClusterExternalDNS(t),
	}

	for _, cluster := range clusters {
		for _, node := range cluster.Spec.Nodes {
			assert.Equal(
				fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id),
					ComputeServiceNameFull(cluster.Name,
						cluster.Namespace,
						cluster.Spec.Service.HeadlessEnabled,
						cluster.Spec.ListenersConfig.UseExternalDNS)),
				ComputeNodeServiceNameFull(node.Id, cluster.Name,
					cluster.Namespace,
					cluster.Spec.Service.HeadlessEnabled,
					cluster.Spec.ListenersConfig.UseExternalDNS))
		}
	}
}

func TestComputeNodeServiceNameNs(t *testing.T) {
	assert := assert.New(t)

	clusters := []*v1alpha1.NifiCluster{
		testClusterLocal(t),
		testClusterDemo(t),
		testClusterExternalDNS(t),
	}

	for _, cluster := range clusters {
		for _, node := range cluster.Spec.Nodes {
			assert.Equal(
				fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id),
					ComputeServiceNameWithNamespace(cluster.Name,
						cluster.Namespace,
						cluster.Spec.Service.HeadlessEnabled,
						cluster.Spec.ListenersConfig.UseExternalDNS)),
				ComputeNodeServiceNameNs(node.Id, cluster.Name,
					cluster.Namespace,
					cluster.Spec.Service.HeadlessEnabled,
					cluster.Spec.ListenersConfig.UseExternalDNS))
		}
	}
}

func TestComputeNodeServiceName(t *testing.T) {
	assert := assert.New(t)

	clusters := []*v1alpha1.NifiCluster{
		testClusterLocal(t),
		testClusterDemo(t),
		testClusterExternalDNS(t),
	}

	for _, cluster := range clusters {
		for _, node := range cluster.Spec.Nodes {
			assert.Equal(
				fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id),
					ComputeServiceName(cluster.Name,
						cluster.Spec.Service.HeadlessEnabled)),
				ComputeNodeServiceName(node.Id, cluster.Name,
					cluster.Spec.Service.HeadlessEnabled, ))
		}
	}
}

func TestComputeHostname(t *testing.T) {
	/*cluster := testCluster(t)

	for _, node := range cluster.Spec.Nodes {
		headlessAddress := ComputeHostname(true, node.Id, cluster.Name, cluster.Namespace)
		expectedAddress := fmt.Sprintf("%s.test-cluster-headless.test-namespace.svc.cluster.local", fmt.Sprintf(templates.NodeNameTemplate, "test-cluster", node.Id))
		if !reflect.DeepEqual(headlessAddress, expectedAddress) {
			t.Errorf("Expected %+v\nGot %+v", expectedAddress, headlessAddress)
		}

		allNodeAddress := ComputeHostname(false, node.Id, cluster.Name, cluster.Namespace)
		expectedAddress = fmt.Sprintf("%s.test-namespace.svc.cluster.local", fmt.Sprintf(templates.NodeNameTemplate, "test-cluster", node.Id))
		if !reflect.DeepEqual(allNodeAddress, expectedAddress) {
			t.Errorf("Expected %+v\nGot %+v", expectedAddress, allNodeAddress)
		}
	}*/
}

func TestInternalListenerForComm(t *testing.T) {
	assert := assert.New(t)

	internalListeners := testClusterLocal(t).Spec.ListenersConfig.InternalListeners
	assert.Equal(v1alpha1.InternalListenerConfig{ContainerPort: httpContainerPort, Type: "https"},
		InternalListenerForComm(internalListeners))
}

func TestDetermineInternalListenerForComm(t *testing.T) {
	assert := assert.New(t)

	internalListeners := testClusterLocal(t).Spec.ListenersConfig.InternalListeners
	assert.Equal(httpContainerPort,
		internalListeners[determineInternalListenerForComm(internalListeners)].ContainerPort)

}
