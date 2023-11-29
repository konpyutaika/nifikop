package certmanagerpki

import (
	"reflect"
	"testing"

	certv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

type mockClient struct {
	client.Client
}

func newMockCluster() *v1.NifiCluster {
	cluster := &v1.NifiCluster{}
	cluster.Name = "test"
	cluster.Namespace = "test-namespace"
	cluster.Spec = v1.NifiClusterSpec{}
	cluster.Spec.ListenersConfig = &v1.ListenersConfig{}
	cluster.Spec.ListenersConfig.InternalListeners = []v1.InternalListenerConfig{
		{ContainerPort: 9092},
	}
	cluster.Spec.ListenersConfig.SSLSecrets = &v1.SSLSecrets{
		TLSSecretName: "test-controller",
		PKIBackend:    v1.PKIBackendCertManager,
		Create:        true,
	}

	cluster.Spec.Nodes = []v1.Node{
		{Id: 0},
		{Id: 1},
		{Id: 2},
	}
	return cluster
}

func newMock(cluster *v1.NifiCluster) *certManager {
	certv1.AddToScheme(scheme.Scheme)
	v1.SchemeBuilder.AddToScheme(scheme.Scheme)
	return &certManager{
		cluster: cluster,
		client:  fake.NewClientBuilder().WithScheme(scheme.Scheme).Build(),
	}
}

func TestNew(t *testing.T) {
	pkiManager := New(&mockClient{}, newMockCluster())
	if reflect.TypeOf(pkiManager) != reflect.TypeOf(&certManager{}) {
		t.Error("Expected new certmanager from New, got:", reflect.TypeOf(pkiManager))
	}
}
