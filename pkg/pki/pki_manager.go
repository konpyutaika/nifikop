package pki

import (
	"context"
	"crypto/tls"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/pki/certmanagerpki"
	"github.com/konpyutaika/nifikop/pkg/util/pki"
)

// MockBackend is used for mocking during testing.
var MockBackend = v1.PKIBackend("mock")

// GetPKIManager returns a PKI/User manager interface for a given cluster.
func GetPKIManager(client client.Client, cluster *v1.NifiCluster) pki.Manager {
	switch cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend {
	// Use cert-manager for pki backend
	case v1.PKIBackendCertManager:
		return certmanagerpki.New(client, cluster)

	// TODO : Add vault
	// Use vault for pki backend
	/*case v1alpha1.PKIBackendVault:
	return vaultpki.New(client, cluster)*/

	// Return mock backend for testing - cannot be triggered by CR due to enum in api schema
	case MockBackend:
		return newMockPKIManager(client, cluster)

	// Default use cert-manager - state explicitly for clarity and to make compiler happy
	default:
		return certmanagerpki.New(client, cluster)
	}
}

// Mock types and functions

type mockPKIManager struct {
	pki.Manager
	client  client.Client
	cluster *v1.NifiCluster
}

func newMockPKIManager(client client.Client, cluster *v1.NifiCluster) pki.Manager {
	return &mockPKIManager{client: client, cluster: cluster}
}

func (m *mockPKIManager) ReconcilePKI(ctx context.Context, logger zap.Logger, scheme *runtime.Scheme, externalHostnames []string) error {
	return nil
}

func (m *mockPKIManager) FinalizePKI(ctx context.Context, logger zap.Logger) error {
	return nil
}

func (m *mockPKIManager) ReconcileUserCertificate(ctx context.Context, user *v1.NifiUser, scheme *runtime.Scheme) (*pki.UserCertificate, error) {
	return &pki.UserCertificate{}, nil
}

func (m *mockPKIManager) FinalizeUserCertificate(ctx context.Context, user *v1.NifiUser) error {
	return nil
}

func (m *mockPKIManager) GetControllerTLSConfig() (*tls.Config, error) {
	return &tls.Config{}, nil
}
