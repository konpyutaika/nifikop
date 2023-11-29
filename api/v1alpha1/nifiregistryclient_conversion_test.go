package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

func TestNifiRegistryClientConversion(t *testing.T) {
	rc := createNifiRegistryClient()

	// convert v1alpha1 to v1
	v1rc := &v1.NifiRegistryClient{}
	rc.ConvertTo(v1rc)
	assertRegistryClientsEqual(rc, v1rc, t)

	// convert v1 to v1alpha1
	newClient := &NifiRegistryClient{}
	newClient.ConvertFrom(v1rc)
	assertRegistryClientsEqual(newClient, v1rc, t)
}

func assertRegistryClientsEqual(rc *NifiRegistryClient, v1rc *v1.NifiRegistryClient, t *testing.T) {
	if rc.ObjectMeta.Name != v1rc.ObjectMeta.Name ||
		rc.ObjectMeta.Namespace != v1rc.ObjectMeta.Namespace {
		t.Error("object metas not equal")
	}
	if rc.Spec.ClusterRef.Name != v1rc.Spec.ClusterRef.Name ||
		rc.Spec.ClusterRef.Namespace != v1rc.Spec.ClusterRef.Namespace {
		t.Error("cluster refs not equal")
	}
	if rc.Spec.Description != v1rc.Spec.Description {
		t.Error("descriptions not equal")
	}
	if rc.Spec.Uri != v1rc.Spec.Uri {
		t.Error("Uris not equal")
	}
	if rc.Status.Id != v1rc.Status.Id {
		t.Error("status IDs not equal")
	}
	if rc.Status.Version != v1rc.Status.Version {
		t.Error("versions not equal")
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
