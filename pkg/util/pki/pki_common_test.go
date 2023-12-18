package pki

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	"github.com/konpyutaika/nifikop/pkg/util"
	certutil "github.com/konpyutaika/nifikop/pkg/util/cert"
)

func testCluster(t *testing.T) *v1.NifiCluster {
	t.Helper()
	cluster := &v1.NifiCluster{}
	cluster.Name = "test-cluster"
	cluster.Namespace = "test-namespace"
	cluster.Spec = v1.NifiClusterSpec{}
	cluster.Spec.ListenersConfig = &v1.ListenersConfig{}

	cluster.Spec.Nodes = []v1.Node{
		{Id: 0},
		{Id: 1},
		{Id: 2},
	}
	return cluster
}

func TestDN(t *testing.T) {
	cert, _, expected, err := certutil.GenerateTestCert()
	if err != nil {
		t.Fatal("failed to generate certificate for testing:", err)
	}
	userCert := &UserCertificate{
		Certificate: cert,
	}
	dn := userCert.DN()
	if dn != expected {
		t.Error("Expected:", expected, "got:", dn)
	}
}

func TestLabelsForNifiPKI(t *testing.T) {
	expected := map[string]string{
		"app":         "nifi",
		"nifi_issuer": fmt.Sprintf(NodeIssuerTemplate, "test"),
	}
	got := LabelsForNifiPKI("test")
	if !reflect.DeepEqual(got, expected) {
		t.Error("Expected:", expected, "got:", got)
	}
}

func TestGetInternalDNSNames(t *testing.T) {
	assert := assert.New(t)
	cluster := testCluster(t)

	for _, node := range cluster.Spec.Nodes {
		cluster.Spec.Service.HeadlessEnabled = true
		headlessNames := GetInternalDNSNames(cluster, node.Id)
		expected := []string{
			"test-cluster-headless.test-namespace.svc.cluster.local",
			fmt.Sprintf("test-cluster-%d-node.test-cluster-headless.test-namespace.svc.cluster.local", node.Id),
			"test-cluster-headless.test-namespace.svc",
			fmt.Sprintf("test-cluster-%d-node.test-cluster-headless.test-namespace.svc", node.Id),
			"test-cluster-headless.test-namespace",
			fmt.Sprintf("test-cluster-%d-node.test-cluster-headless.test-namespace", node.Id),
			"test-cluster-headless",
			fmt.Sprintf("test-cluster-%d-node.test-cluster-headless", node.Id),
			fmt.Sprintf("test-cluster-%d-node", node.Id),
		}
		if !reflect.DeepEqual(expected, headlessNames) {
			t.Error("Expected:", expected, "got:", headlessNames)
		}

		cluster.Spec.Service.HeadlessEnabled = false
		allNodeNames := GetInternalDNSNames(cluster, node.Id)
		expected = []string{
			"test-cluster-all-node.test-namespace.svc.cluster.local",
			fmt.Sprintf("test-cluster-%d-node.test-namespace.svc.cluster.local", node.Id),
			"test-cluster-all-node.test-namespace.svc",
			fmt.Sprintf("test-cluster-%d-node.test-namespace.svc", node.Id),
			"test-cluster-all-node.test-namespace",
			fmt.Sprintf("test-cluster-%d-node.test-namespace", node.Id),
			"test-cluster-all-node",
			fmt.Sprintf("test-cluster-%d-node", node.Id),
			fmt.Sprintf("test-cluster-%d-node.test-cluster-all-node.test-namespace.svc.cluster.local", node.Id),
		}
		if !reflect.DeepEqual(expected, allNodeNames) {
			t.Error("Expected:", expected, "got:", allNodeNames)
		}
	}

	cluster.Spec.ListenersConfig.UseExternalDNS = true
	for _, node := range cluster.Spec.Nodes {
		names := GetInternalDNSNames(cluster, node.Id)
		assert.Equal(2, len(names))
		expected := []string{
			"test-cluster-all-node.cluster.local",
			fmt.Sprintf("test-cluster-%d-node.cluster.local", node.Id),
		}

		assert.Equal(expected, names)
	}
}

func TestNodeUsersForCluster(t *testing.T) {
	cluster := testCluster(t)
	users := NodeUsersForCluster(cluster, []string{})

	for _, node := range cluster.Spec.Nodes {
		expected := &v1.NifiUser{
			ObjectMeta: templates.ObjectMeta(GetNodeUserName(cluster, node.Id), LabelsForNifiPKI(cluster.Name), cluster),
			Spec: v1.NifiUserSpec{
				SecretName: fmt.Sprintf(NodeServerCertTemplate, cluster.Name, node.Id),
				DNSNames:   GetInternalDNSNames(cluster, node.Id),
				IncludeJKS: true,
				ClusterRef: v1.ClusterReference{
					Name:      cluster.Name,
					Namespace: cluster.Namespace,
				},
				AccessPolicies: []v1.AccessPolicy{
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ProxyAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.ProxyAccessPolicyResource},
				},
			},
		}
		if !util.NifiUserSliceContains(users, expected) {
			t.Errorf("Expected %+v\ninto %+v", expected, users)
		}
	}
}

func TestControllerUserForCluster(t *testing.T) {
	cluster := testCluster(t)
	user := ControllerUserForCluster(cluster)
	/*	nodeControllerName := fmt.Sprintf(NodeControllerFQDNTemplate,
		cluster.GetNifiControllerUserIdentity(),
		cluster.Namespace,
		cluster.Spec.ListenersConfig.GetClusterDomain())*/

	expected := &v1.NifiUser{
		ObjectMeta: templates.ObjectMeta(
			cluster.GetNifiControllerUserIdentity(),
			LabelsForNifiPKI(cluster.Name), cluster,
		),
		Spec: v1.NifiUserSpec{
			DNSNames:   []string{cluster.GetNifiControllerUserIdentity()},
			SecretName: cluster.GetNifiControllerUserIdentity(),
			IncludeJKS: true,
			ClusterRef: v1.ClusterReference{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
			AccessPolicies: []v1.AccessPolicy{
				{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.FlowAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.FlowAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ControllerAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.ControllerAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.RestrictedComponentsAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.RestrictedComponentsAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.PoliciesAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.PoliciesAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.TenantsAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.TenantsAccessPolicyResource},
			},
		},
	}

	if !reflect.DeepEqual(user, expected) {
		t.Errorf("Expected %+v\nGot %+v", expected, user)
	}
}
