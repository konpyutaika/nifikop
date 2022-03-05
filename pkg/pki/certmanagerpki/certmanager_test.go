package certmanagerpki

import (
	"reflect"
	"testing"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type mockClient struct {
	client.Client
}

func newMockCluster() *v1alpha1.NifiCluster {
	cluster := &v1alpha1.NifiCluster{}
	cluster.Name = "test"
	cluster.Namespace = "test-namespace"
	cluster.Spec = v1alpha1.NifiClusterSpec{}
	cluster.Spec.ListenersConfig = &v1alpha1.ListenersConfig{}
	cluster.Spec.ListenersConfig.InternalListeners = []v1alpha1.InternalListenerConfig{
		{ContainerPort: 9092},
	}
	cluster.Spec.ListenersConfig.SSLSecrets = &v1alpha1.SSLSecrets{
		TLSSecretName: "test-controller",
		PKIBackend:    v1alpha1.PKIBackendCertManager,
		Create:        true,
	}

	cluster.Spec.Nodes = []v1alpha1.Node{
		{Id: 0},
		{Id: 1},
		{Id: 2},
	}
	return cluster
}

func newMock(cluster *v1alpha1.NifiCluster) *certManager {
	certv1.AddToScheme(scheme.Scheme)
	v1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)
	return &certManager{
		cluster: cluster,
		client:  fake.NewFakeClientWithScheme(scheme.Scheme),
	}
}

func TestNew(t *testing.T) {
	pkiManager := New(&mockClient{}, newMockCluster())
	if reflect.TypeOf(pkiManager) != reflect.TypeOf(&certManager{}) {
		t.Error("Expected new certmanager from New, got:", reflect.TypeOf(pkiManager))
	}
}
