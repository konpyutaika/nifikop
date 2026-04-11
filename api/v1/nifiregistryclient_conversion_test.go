package v1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v2alpha1 "github.com/konpyutaika/nifikop/api/v2alpha1"
)

func TestNifiRegistryClientConversion(t *testing.T) {
	rc := createNifiRegistryClient()

	// convert v1 to v2alpha1
	v2rc := &v2alpha1.NifiRegistryClient{}
	rc.ConvertTo(v2rc)
	assertRegistryClientsEqual(rc, v2rc, t)

	// convert v2alpha1 to v1
	newClient := &NifiRegistryClient{}
	newClient.ConvertFrom(v2rc)
	assertRegistryClientsEqual(newClient, v2rc, t)
}

func assertRegistryClientsEqual(rc *NifiRegistryClient, v2rc *v2alpha1.NifiRegistryClient, t *testing.T) {
	t.Helper()
	if rc.ObjectMeta.Name != v2rc.ObjectMeta.Name ||
		rc.ObjectMeta.Namespace != v2rc.ObjectMeta.Namespace {
		t.Error("object metas not equal")
	}
	if rc.Spec.ClusterRef.Name != v2rc.Spec.ClusterRef.Name ||
		rc.Spec.ClusterRef.Namespace != v2rc.Spec.ClusterRef.Namespace {
		t.Error("cluster refs not equal")
	}
	if rc.Spec.Description != v2rc.Spec.Description {
		t.Error("description not equal")
	}
	if v2rc.Spec.Type != v2alpha1.RegistryClientType {
		t.Error("type not equal to registry")
	}
	if v2rc.Spec.RegistryClientConfig == nil || rc.Spec.Uri != v2rc.Spec.RegistryClientConfig.Uri {
		t.Error("uri not equal")
	}
	if rc.Status.Id != v2rc.Status.Id {
		t.Error("status IDs not equal")
	}
	if rc.Status.Version != v2rc.Status.Version {
		t.Error("version not equal")
	}
}

func createNifiRegistryClient() *NifiRegistryClient {
	return &NifiRegistryClient{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "registryClient",
			Namespace: "namespace",
		},
		Spec: NifiRegistryClientSpec{
			Uri:         "registry.uri",
			Description: "description",
			ClusterRef: ClusterReference{
				Name:      "cluster",
				Namespace: "namespace",
			},
		},
		Status: NifiRegistryClientStatus{
			Id:      "id",
			Version: 6,
		},
	}
}
