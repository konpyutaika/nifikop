package nifi

//
// import (
//	"fmt"
//	"strings"
//	"testing"
//
//	"github.com/konpyutaika/nifikop/api/v1"
//	"github.com/stretchr/testify/assert"
//)
//
// const (
//	httpContainerPort int32 = 443
//
//	clusterName      = "test-cluster"
//	clusterNamespace = "test-namespace"
//
//	localClusterDomain       = "cluster.local"
//	demoClusterDomain        = "demo.local"
//	externalDNSClusterDomain = "external.dns.com"
//)
//
// func TestParseStringToInt32(t *testing.T) {
//	assert := assert.New(t)
//
//	var expectedValue int32 = 12
//	parsed, _ := ParseStringToInt32("12")
//
//	assert.Equal(expectedValue, parsed)
//}
//
// func testClusterLocal(t *testing.T) *v1.NifiCluster {
//	t.Helper()
//	cluster := &v1.NifiCluster{}
//
//	cluster.Name = clusterName
//	cluster.Namespace = clusterNamespace
//	cluster.Spec = v1.NifiClusterSpec{}
//
//	cluster.Spec.Nodes = []v1.Node{
//		{Id: 0},
//		{Id: 1},
//		{Id: 2},
//	}
//
//	cluster.Spec.ListenersConfig.InternalListeners = []v1.InternalListenerConfig{
//		{Type: "https", ContainerPort: httpContainerPort},
//		{Type: "http", ContainerPort: 8080},
//		{Type: "cluster", ContainerPort: 8083},
//		{Type: "s2s", ContainerPort: 8085},
//	}
//	return cluster
//}
//
// func testClusterDemo(t *testing.T) *v1.NifiCluster {
//	t.Helper()
//	cluster := testClusterLocal(t)
//	cluster.Spec.ListenersConfig.ClusterDomain = demoClusterDomain
//	return cluster
//}
//
// func testClusterExternalDNS(t *testing.T) *v1.NifiCluster {
//	t.Helper()
//	cluster := testClusterLocal(t)
//	cluster.Spec.ListenersConfig.ClusterDomain = externalDNSClusterDomain
//	cluster.Spec.ListenersConfig.UseExternalDNS = true
//	return cluster
//}
//
// func TestGenerateNiFiAddressFromCluster(t *testing.T) {
//
//	testNiFiAddressFromCluster(t, testClusterLocal(t), localClusterDomain, false)
//	testNiFiAddressFromCluster(t, testClusterDemo(t), demoClusterDomain, false)
//	testNiFiAddressFromCluster(t, testClusterExternalDNS(t), externalDNSClusterDomain, true)
//}
//
// func testNiFiAddressFromCluster(t *testing.T,
//	cluster *v1.NifiCluster, expectedClusterDomain string, expectedUseExternalDNS bool) {
//
//	assert := assert.New(t)
//
//	// Test headless service
//	cluster.Spec.Service.HeadlessEnabled = true
//	assert.Equal(
//		fmt.Sprintf("%s:%d", ComputeAllNodeServiceHostname(clusterName, clusterNamespace,
//			true, expectedClusterDomain, expectedUseExternalDNS),
//			httpContainerPort),
//		GenerateNiFiAddressFromCluster(cluster))
//
//	// Test all nodes service
//	cluster.Spec.Service.HeadlessEnabled = false
//	assert.Equal(
//		fmt.Sprintf("%s:%d", ComputeAllNodeServiceHostname(clusterName, clusterNamespace,
//			false, expectedClusterDomain, expectedUseExternalDNS),
//			httpContainerPort),
//		GenerateNiFiAddressFromCluster(cluster))
//}
//
// func TestComputeNiFiAddress(t *testing.T) {
//
//	cluster := testClusterLocal(t)
//	testNiFiAddress(t,
//		cluster.Name,
//		cluster.Namespace,
//		cluster.Spec.Service.HeadlessEnabled,
//		cluster.Spec.ListenersConfig.GetClusterDomain(),
//		cluster.Spec.ListenersConfig.UseExternalDNS,
//		cluster.Spec.ListenersConfig.InternalListeners,
//		localClusterDomain, false)
//
//	cluster = testClusterDemo(t)
//	testNiFiAddress(t,
//		cluster.Name,
//		cluster.Namespace,
//		cluster.Spec.Service.HeadlessEnabled,
//		cluster.Spec.ListenersConfig.GetClusterDomain(),
//		cluster.Spec.ListenersConfig.UseExternalDNS,
//		cluster.Spec.ListenersConfig.InternalListeners,
//		demoClusterDomain, false)
//
//	cluster = testClusterExternalDNS(t)
//	testNiFiAddress(t,
//		cluster.Name,
//		cluster.Namespace,
//		cluster.Spec.Service.HeadlessEnabled,
//		cluster.Spec.ListenersConfig.GetClusterDomain(),
//		cluster.Spec.ListenersConfig.UseExternalDNS,
//		cluster.Spec.ListenersConfig.InternalListeners,
//		externalDNSClusterDomain, true)
//
//}
//
// func testNiFiAddress(t *testing.T,
//	clusterName, namespace string,
//	headlessServiceEnabled bool,
//	clusterDomain string,
//	useExternalDNS bool,
//	internalListeners []v1.InternalListenerConfig,
//	expectedClusterDomain string, expectedUseExternalDNS bool) {
//
//	assert := assert.New(t)
//
//	assert.Equal(
//		fmt.Sprintf("%s:%d", ComputeAllNodeServiceHostname(
//			clusterName, clusterNamespace, headlessServiceEnabled,
//			expectedClusterDomain, expectedUseExternalDNS), httpContainerPort),
//		ComputeNiFiAddress(clusterName, namespace, headlessServiceEnabled, clusterDomain, useExternalDNS, internalListeners))
//
//}
//
// func TestComputeAllNodeServiceHostname(t *testing.T) {
//	cluster := testClusterLocal(t)
//	testComputeAllNodeServiceHostname(t,
//		cluster.Name,
//		cluster.Namespace,
//		cluster.Spec.Service.HeadlessEnabled,
//		cluster.Spec.ListenersConfig.GetClusterDomain(),
//		cluster.Spec.ListenersConfig.UseExternalDNS,
//		localClusterDomain, false)
//
//	cluster = testClusterDemo(t)
//	testComputeAllNodeServiceHostname(t,
//		cluster.Name,
//		cluster.Namespace,
//		cluster.Spec.Service.HeadlessEnabled,
//		cluster.Spec.ListenersConfig.GetClusterDomain(),
//		cluster.Spec.ListenersConfig.UseExternalDNS,
//		demoClusterDomain, false)
//
//	cluster = testClusterExternalDNS(t)
//	testComputeAllNodeServiceHostname(t,
//		cluster.Name,
//		cluster.Namespace,
//		cluster.Spec.Service.HeadlessEnabled,
//		cluster.Spec.ListenersConfig.GetClusterDomain(),
//		cluster.Spec.ListenersConfig.UseExternalDNS,
//		externalDNSClusterDomain, true)
//}
//
// func testComputeAllNodeServiceHostname(t *testing.T,
//	clusterName, namespace string,
//	headlessServiceEnabled bool,
//	clusterDomain string,
//	useExternalDNS bool,
//	ExpectedClusterDomain string, ExpectedUseExternalDNS bool) {
//
//	assert := assert.New(t)
//
//	expectedHeadless, expectedAllNodes := computeNiFiHostnames(ExpectedClusterDomain, ExpectedUseExternalDNS)
//
//	if headlessServiceEnabled {
//		assert.Equal(
//			expectedHeadless,
//			ComputeAllNodeServiceHostname(clusterName, namespace, headlessServiceEnabled, clusterDomain, useExternalDNS))
//	} else {
//		assert.Equal(
//			expectedAllNodes,
//			ComputeAllNodeServiceHostname(clusterName, namespace, headlessServiceEnabled, clusterDomain, useExternalDNS))
//	}
//}
//
//func computeNiFiHostnames(clusterDomain string, useExternalDNS bool) (string, string) {
//	svc := ""
//	if !useExternalDNS {
//		svc = fmt.Sprintf(".%s.svc", clusterNamespace)
//	}
//
//	return fmt.Sprintf("%s-headless%s.%s", clusterName, svc, clusterDomain),
//		fmt.Sprintf("%s-all-node%s.%s", clusterName, svc, clusterDomain)
//}
//
//func TestComputeAllNodeServiceNameFull(t *testing.T) {
//	assert := assert.New(t)
//
//	cluster := testClusterLocal(t)
//	cluster.Spec.Service.HeadlessEnabled = true
//	assert.Equal(fmt.Sprintf("%s.%s.svc", ComputeAllNodeServiceName(clusterName, true), clusterNamespace),
//		ComputeAllNodeServiceNameFull(cluster.Name, cluster.Namespace,
//			cluster.Spec.Service.HeadlessEnabled, cluster.Spec.ListenersConfig.UseExternalDNS))
//
//	cluster = testClusterExternalDNS(t)
//	cluster.Spec.Service.HeadlessEnabled = true
//	assert.Equal(ComputeAllNodeServiceName(clusterName, true),
//		ComputeAllNodeServiceNameFull(cluster.Name, cluster.Namespace,
//			cluster.Spec.Service.HeadlessEnabled, cluster.Spec.ListenersConfig.UseExternalDNS))
//}
//
//func TestComputeAllNodeServiceNameNs(t *testing.T) {
//	assert := assert.New(t)
//
//	cluster := testClusterLocal(t)
//	cluster.Spec.Service.HeadlessEnabled = true
//	assert.Equal(fmt.Sprintf("%s.%s", ComputeAllNodeServiceName(clusterName, true), clusterNamespace),
//		ComputeAllNodeServiceNameNs(cluster.Name, cluster.Namespace, cluster.Spec.Service.HeadlessEnabled))
//
//}
//
//func TestGenerateNodeAddressFromCluster(t *testing.T) {
//	assert := assert.New(t)
//
//	clusters := []*v1.NifiCluster{
//		testClusterLocal(t),
//		testClusterDemo(t),
//		testClusterExternalDNS(t),
//	}
//
//	for _, cluster := range clusters {
//		for _, node := range cluster.Spec.Nodes {
//			nifiAddress := GenerateNiFiAddressFromCluster(cluster)
//			if !cluster.Spec.Service.HeadlessEnabled {
//				nifiAddress = strings.SplitAfterN(nifiAddress, ".", 2)[1]
//			}
//			assert.Equal(
//				fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id), nifiAddress),
//				GenerateNodeAddressFromCluster(node.Id, cluster))
//		}
//	}
//}
//
//func TestComputeNodeAddress(t *testing.T) {
//	assert := assert.New(t)
//
//	clusters := []*v1.NifiCluster{
//		testClusterLocal(t),
//		testClusterDemo(t),
//		testClusterExternalDNS(t),
//	}
//
//	for _, cluster := range clusters {
//		for _, node := range cluster.Spec.Nodes {
//			nifiAddress := ComputeNiFiAddress(cluster.Name,
//				cluster.Namespace,
//				cluster.Spec.Service.HeadlessEnabled,
//				cluster.Spec.ListenersConfig.GetClusterDomain(),
//				cluster.Spec.ListenersConfig.UseExternalDNS,
//				cluster.Spec.ListenersConfig.InternalListeners)
//
//			if !cluster.Spec.Service.HeadlessEnabled {
//				nifiAddress = strings.SplitAfterN(nifiAddress, ".", 2)[1]
//			}
//
//			assert.Equal(
//				fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id), nifiAddress),
//				ComputeNodeAddress(node.Id, cluster.Name,
//					cluster.Namespace,
//					cluster.Spec.Service.HeadlessEnabled,
//					cluster.Spec.ListenersConfig.GetClusterDomain(),
//					cluster.Spec.ListenersConfig.UseExternalDNS,
//					cluster.Spec.ListenersConfig.InternalListeners))
//		}
//	}
//}
//
//func TestComputeNodeHostname(t *testing.T) {
//	assert := assert.New(t)
//
//	clusters := []*v1.NifiCluster{
//		testClusterLocal(t),
//		testClusterDemo(t),
//		testClusterExternalDNS(t),
//	}
//
//	for _, cluster := range clusters {
//		for _, node := range cluster.Spec.Nodes {
//			nifiAddress := ComputeAllNodeServiceHostname(cluster.Name,
//				cluster.Namespace,
//				cluster.Spec.Service.HeadlessEnabled,
//				cluster.Spec.ListenersConfig.GetClusterDomain(),
//				cluster.Spec.ListenersConfig.UseExternalDNS)
//			if !cluster.Spec.Service.HeadlessEnabled {
//				nifiAddress = strings.SplitAfterN(nifiAddress, ".", 2)[1]
//			}
//
//			assert.Equal(
//				fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id), nifiAddress),
//				ComputeNodeHostname(node.Id, cluster.Name,
//					cluster.Namespace,
//					cluster.Spec.Service.HeadlessEnabled,
//					cluster.Spec.ListenersConfig.GetClusterDomain(),
//					cluster.Spec.ListenersConfig.UseExternalDNS))
//		}
//	}
//}
//
//func TestComputeNodeServiceNameFull(t *testing.T) {
//	assert := assert.New(t)
//
//	clusters := []*v1.NifiCluster{
//		testClusterLocal(t),
//		testClusterDemo(t),
//		testClusterExternalDNS(t),
//	}
//
//	for _, cluster := range clusters {
//		for _, node := range cluster.Spec.Nodes {
//			nifiAddress := ComputeAllNodeServiceNameFull(cluster.Name,
//				cluster.Namespace,
//				cluster.Spec.Service.HeadlessEnabled,
//				cluster.Spec.ListenersConfig.UseExternalDNS)
//			if !cluster.Spec.Service.HeadlessEnabled && !cluster.Spec.ListenersConfig.UseExternalDNS {
//				nifiAddress = strings.SplitAfterN(nifiAddress, ".", 2)[1]
//			}
//
//			toTest := ComputeNodeServiceNameFull(node.Id, cluster.Name,
//				cluster.Namespace,
//				cluster.Spec.Service.HeadlessEnabled,
//				cluster.Spec.ListenersConfig.UseExternalDNS)
//
//			if cluster.Spec.ListenersConfig.UseExternalDNS {
//				assert.Equal(fmt.Sprintf(NodeNameTemplate, clusterName, node.Id), toTest)
//			} else {
//				assert.Equal(
//					fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id), nifiAddress),
//					toTest)
//			}
//		}
//	}
//}
//
//func TestComputeNodeServiceNameNs(t *testing.T) {
//	assert := assert.New(t)
//
//	clusters := []*v1.NifiCluster{
//		testClusterLocal(t),
//		testClusterDemo(t),
//		testClusterExternalDNS(t),
//	}
//
//	for _, cluster := range clusters {
//		for _, node := range cluster.Spec.Nodes {
//
//			nifiAddress := ComputeAllNodeServiceNameNs(cluster.Name,
//				cluster.Namespace,
//				cluster.Spec.Service.HeadlessEnabled)
//			if !cluster.Spec.Service.HeadlessEnabled {
//				nifiAddress = strings.SplitAfterN(nifiAddress, ".", 2)[1]
//			}
//
//			assert.Equal(
//				fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id),
//					nifiAddress),
//				ComputeNodeServiceNameNs(node.Id, cluster.Name,
//					cluster.Namespace,
//					cluster.Spec.Service.HeadlessEnabled))
//		}
//	}
//}
//
//func TestComputeNodeServiceName(t *testing.T) {
//	assert := assert.New(t)
//
//	clusters := []*v1.NifiCluster{
//		testClusterLocal(t),
//		testClusterDemo(t),
//		testClusterExternalDNS(t),
//	}
//
//	for _, cluster := range clusters {
//		for _, node := range cluster.Spec.Nodes {
//
//			nifiAddress := ComputeAllNodeServiceName(cluster.Name,
//				cluster.Spec.Service.HeadlessEnabled)
//
//			toTest := ComputeNodeServiceName(node.Id, cluster.Name,
//				cluster.Spec.Service.HeadlessEnabled)
//
//			if !cluster.Spec.Service.HeadlessEnabled {
//				assert.Equal(
//					fmt.Sprintf(NodeNameTemplate, clusterName, node.Id), toTest)
//			} else {
//				assert.Equal(
//					fmt.Sprintf("%s.%s", fmt.Sprintf(NodeNameTemplate, clusterName, node.Id),
//						nifiAddress),
//					toTest)
//			}
//
//		}
//	}
//}
//
//func TestComputeHostname(t *testing.T) {
//	/*cluster := testCluster(t)
//
//	for _, node := range cluster.Spec.Nodes {
//		headlessAddress := ComputeHostname(true, node.Id, cluster.Name, cluster.Namespace)
//		expectedAddress := fmt.Sprintf("%s.test-cluster-headless.test-namespace.svc.cluster.local", fmt.Sprintf(templates.NodeNameTemplate, "test-cluster", node.Id))
//		if !reflect.DeepEqual(headlessAddress, expectedAddress) {
//			t.Errorf("Expected %+v\nGot %+v", expectedAddress, headlessAddress)
//		}
//
//		allNodeAddress := ComputeHostname(false, node.Id, cluster.Name, cluster.Namespace)
//		expectedAddress = fmt.Sprintf("%s.test-namespace.svc.cluster.local", fmt.Sprintf(templates.NodeNameTemplate, "test-cluster", node.Id))
//		if !reflect.DeepEqual(allNodeAddress, expectedAddress) {
//			t.Errorf("Expected %+v\nGot %+v", expectedAddress, allNodeAddress)
//		}
//	}*/
//}
//
//func TestInternalListenerForComm(t *testing.T) {
//	assert := assert.New(t)
//
//	internalListeners := testClusterLocal(t).Spec.ListenersConfig.InternalListeners
//	assert.Equal(v1.InternalListenerConfig{ContainerPort: httpContainerPort, Type: "https"},
//		InternalListenerForComm(internalListeners))
//}
//
//func TestDetermineInternalListenerForComm(t *testing.T) {
//	assert := assert.New(t)
//
//	internalListeners := testClusterLocal(t).Spec.ListenersConfig.InternalListeners
//	assert.Equal(httpContainerPort,
//		internalListeners[determineInternalListenerForComm(internalListeners)].ContainerPort)
//
//}
