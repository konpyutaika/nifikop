package pki

import (
	"context"
	"reflect"
	"testing"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

var log zap.Logger

type mockClient struct {
	client.Client
}

func newMockCluster() *v1.NifiCluster {
	cluster := &v1.NifiCluster{}
	cluster.Name = "test"
	cluster.Namespace = "test"
	cluster.Spec = v1.NifiClusterSpec{}
	cluster.Spec.ListenersConfig = &v1.ListenersConfig{}
	cluster.Spec.ListenersConfig.InternalListeners = []v1.InternalListenerConfig{
		{ContainerPort: 80},
	}
	cluster.Spec.ListenersConfig.SSLSecrets = &v1.SSLSecrets{
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

	if _, err = mock.ReconcileUserCertificate(ctx, log, &v1.NifiUser{}, scheme.Scheme); err != nil {
		t.Error("Expected nil error got:", err)
	}

	if err = mock.FinalizeUserCertificate(ctx, &v1.NifiUser{}); err != nil {
		t.Error("Expected nil error got:", err)
	}

	if _, err = mock.GetControllerTLSConfig(); err != nil {
		t.Error("Expected nil error got:", err)
	}

	// Test other getters
	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1.PKIBackendCertManager
	certmanager := GetPKIManager(&mockClient{}, cluster)
	pkiType := reflect.TypeOf(certmanager).String()
	expected := "*certmanagerpki.certManager"
	if pkiType != expected {
		t.Error("Expected:", expected, "got:", pkiType)
	}

	// Default should be cert-manager also
	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1.PKIBackend("")
	certmanager = GetPKIManager(&mockClient{}, cluster)
	pkiType = reflect.TypeOf(certmanager).String()
	expected = "*certmanagerpki.certManager"
	if pkiType != expected {
		t.Error("Expected:", expected, "got:", pkiType)
	}

	/* TODO: Add Vault
	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1alpha1.PKIBackendVault
	certmanager = GetPKIManager(&mockClient{}, cluster)
	pkiType = reflect.TypeOf(certmanager).String()
	expected = "*vaultpki.vaultPKI"
	if pkiType != expected {
		t.Error("Expected:", expected, "got:", pkiType)
	}*/
}

// TestDefaultBackendFollowsOperatorFlag covers the selection the nifikop chart drives:
// with -cert-manager-enabled=false the operator signs certificates itself, and a cluster
// that names cert-manager explicitly still gets cert-manager.
func TestDefaultBackendFollowsOperatorFlag(t *testing.T) {
	original := defaultBackend
	defer SetDefaultBackend(original)

	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1.PKIBackend("")

	SetDefaultBackend(v1.PKIBackendOperator)
	manager := GetPKIManager(&mockClient{}, cluster)
	if got, expected := reflect.TypeOf(manager).String(), "*operatorpki.operatorPKI"; got != expected {
		t.Error("Expected:", expected, "got:", got)
	}

	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1.PKIBackendCertManager
	manager = GetPKIManager(&mockClient{}, cluster)
	if got, expected := reflect.TypeOf(manager).String(), "*certmanagerpki.certManager"; got != expected {
		t.Error("Expected an explicit backend to win, expected:", expected, "got:", got)
	}

	SetDefaultBackend(v1.PKIBackendCertManager)
	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1.PKIBackend("")
	manager = GetPKIManager(&mockClient{}, cluster)
	if got, expected := reflect.TypeOf(manager).String(), "*certmanagerpki.certManager"; got != expected {
		t.Error("Expected:", expected, "got:", got)
	}
}

// nop is a usable logger: the package-level log is a zero value, which panics on use.
var nop = *zap.NewNop()

// TestFinalizeWithoutSSLSecrets: listenersConfig is optional, so a cluster can reach
// teardown with no TLS configuration at all. Both backends have to finalize it without
// panicking, which is what leaves a cluster's finalizer stuck.
func TestFinalizeWithoutSSLSecrets(t *testing.T) {
	original := defaultBackend
	defer SetDefaultBackend(original)

	for _, backend := range []v1.PKIBackend{v1.PKIBackendCertManager, v1.PKIBackendOperator} {
		for name, mutate := range map[string]func(*v1.NifiCluster){
			"no sslSecrets":      func(c *v1.NifiCluster) { c.Spec.ListenersConfig.SSLSecrets = nil },
			"no listenersConfig": func(c *v1.NifiCluster) { c.Spec.ListenersConfig = nil },
		} {
			t.Run(string(backend)+"/"+name, func(t *testing.T) {
				SetDefaultBackend(backend)
				cluster := newMockCluster()
				mutate(cluster)

				manager := GetPKIManager(&mockClient{}, cluster)
				if err := manager.FinalizePKI(context.Background(), nop); err != nil {
					t.Error("Expected no error, got:", err)
				}
			})
		}
	}
}

// TestUnsupportedBackendRefuses: a named backend the operator does not implement must not
// quietly become a different trust implementation. "vault" passes schema validation.
func TestUnsupportedBackendRefuses(t *testing.T) {
	original := defaultBackend
	defer SetDefaultBackend(original)

	for _, fallback := range []v1.PKIBackend{v1.PKIBackendCertManager, v1.PKIBackendOperator} {
		t.Run("default "+string(fallback), func(t *testing.T) {
			SetDefaultBackend(fallback)
			cluster := newMockCluster()
			cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1.PKIBackend("vault")

			manager := GetPKIManager(&mockClient{}, cluster)
			if got := reflect.TypeOf(manager).String(); got != "*pki.unsupportedPKIManager" {
				t.Fatal("Expected an unsupported backend not to fall back, got:", got)
			}

			ctx := context.Background()
			if err := manager.ReconcilePKI(ctx, nop, scheme.Scheme, []string{}); err == nil {
				t.Error("Expected ReconcilePKI to refuse")
			}
			if _, err := manager.ReconcileUserCertificate(ctx, nop, &v1.NifiUser{}, scheme.Scheme); err == nil {
				t.Error("Expected ReconcileUserCertificate to refuse")
			}
			if _, err := manager.GetControllerTLSConfig(); err == nil {
				t.Error("Expected GetControllerTLSConfig to refuse")
			}
			// Deletion must still be possible, or the cluster's finalizer is stuck for good.
			if err := manager.FinalizePKI(ctx, nop); err != nil {
				t.Error("Expected FinalizePKI to let the cluster be deleted, got:", err)
			}
			if err := manager.FinalizeUserCertificate(ctx, &v1.NifiUser{}); err != nil {
				t.Error("Expected FinalizeUserCertificate to let the user be deleted, got:", err)
			}
		})
	}
}
