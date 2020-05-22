package nifi

import (
	"fmt"
	"reflect"
	"testing"

	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/apis/nifi/v1alpha1"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/resources/templates"
)

func testCluster(t *testing.T) *v1alpha1.NifiCluster {
	t.Helper()
	cluster := &v1alpha1.NifiCluster{}
	cluster.Name = "test-cluster"
	cluster.Namespace = "test-namespace"
	cluster.Spec = v1alpha1.NifiClusterSpec{}

	cluster.Spec.Nodes = []v1alpha1.Node{
		{Id: 0},
		{Id: 1},
		{Id: 2},
	}
	return cluster
}


func TestComputeHostname(t *testing.T) {
	cluster := testCluster(t)

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
	}
}