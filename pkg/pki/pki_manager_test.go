package pki

import (
	"context"
	"reflect"
	"testing"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log zap.Logger

type mockClient struct {
	client.Client
}

func newMockCluster() *v1alpha1.NifiCluster {
	cluster := &v1alpha1.NifiCluster{}
	cluster.Name = "test"
	cluster.Namespace = "test"
	cluster.Spec = v1alpha1.NifiClusterSpec{}
	cluster.Spec.ListenersConfig = &v1alpha1.ListenersConfig{}
	cluster.Spec.ListenersConfig.InternalListeners = []v1alpha1.InternalListenerConfig{
		{ContainerPort: 80},
	}
	cluster.Spec.ListenersConfig.SSLSecrets = &v1alpha1.SSLSecrets{
		PKIBackend: MockBackend,
	}
	return cluster
}

func TestGetPKIManager(t *testing.T) {
	cluster := newMockCluster()
	mock := GetPKIManager(&mockClient{}, cluster)
	if reflect.TypeOf(mock) != reflect.TypeOf(&mockPKIManager{}) {
		t.Error("Expected mock client got:", reflect.TypeOf(mock))
	}
	ctx := context.Background()

	// Test mock functions
	var err error
	if err = mock.ReconcilePKI(ctx, log, scheme.Scheme, []string{}); err != nil {
		t.Error("Expected nil error got:", err)
	}

	if err = mock.FinalizePKI(ctx, log); err != nil {
		t.Error("Expected nil error got:", err)
	}

	if _, err = mock.ReconcileUserCertificate(ctx, &v1alpha1.NifiUser{}, scheme.Scheme); err != nil {
		t.Error("Expected nil error got:", err)
	}

	if err = mock.FinalizeUserCertificate(ctx, &v1alpha1.NifiUser{}); err != nil {
		t.Error("Expected nil error got:", err)
	}

	if _, err = mock.GetControllerTLSConfig(); err != nil {
		t.Error("Expected nil error got:", err)
	}

	// Test other getters
	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1alpha1.PKIBackendCertManager
	certmanager := GetPKIManager(&mockClient{}, cluster)
	pkiType := reflect.TypeOf(certmanager).String()
	expected := "*certmanagerpki.certManager"
	if pkiType != expected {
		t.Error("Expected:", expected, "got:", pkiType)
	}

	// Default should be cert-manager also
	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1alpha1.PKIBackend("")
	certmanager = GetPKIManager(&mockClient{}, cluster)
	pkiType = reflect.TypeOf(certmanager).String()
	expected = "*certmanagerpki.certManager"
	if pkiType != expected {
		t.Error("Expected:", expected, "got:", pkiType)
	}

	/* TODO : Add Vault
	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1alpha1.PKIBackendVault
	certmanager = GetPKIManager(&mockClient{}, cluster)
	pkiType = reflect.TypeOf(certmanager).String()
	expected = "*vaultpki.vaultPKI"
	if pkiType != expected {
		t.Error("Expected:", expected, "got:", pkiType)
	}*/

}
