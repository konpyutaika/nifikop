package pki

import (
	"context"
	"crypto/tls"
	"fmt"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/pki/certmanagerpki"
	"github.com/konpyutaika/nifikop/pkg/pki/operatorpki"
	"github.com/konpyutaika/nifikop/pkg/util/pki"
)

// MockBackend is used for mocking during testing.
var MockBackend = v1.PKIBackend("mock")

// defaultBackend is used for clusters that do not name a PKI backend. It stays
// cert-manager unless the operator is started with -cert-manager-enabled=false, which
// is how a cluster without the cert-manager CRDs selects the operator's own backend.
var defaultBackend = v1.PKIBackendCertManager

// SetDefaultBackend sets the PKI backend used when a NifiCluster leaves
// spec.listenersConfig.sslSecrets.pkiBackend empty. It is called once at startup.
func SetDefaultBackend(backend v1.PKIBackend) {
	defaultBackend = backend
}

// GetPKIManager returns a PKI/User manager interface for a given cluster.
//
// Only an empty backend takes the operator's default. A backend that is named but not
// implemented - "vault" passes schema validation - returns a manager that refuses every
// operation, rather than quietly signing the cluster's certificates with a different
// trust implementation from the one it asked for.
func GetPKIManager(client client.Client, cluster *v1.NifiCluster) pki.Manager {
	var backend v1.PKIBackend
	if sslSecrets := cluster.GetSSLSecrets(); sslSecrets != nil {
		backend = sslSecrets.PKIBackend
	}
	if backend == "" {
		backend = defaultBackend
	}

	switch backend {
	// Use cert-manager for pki backend
	case v1.PKIBackendCertManager:
		return certmanagerpki.New(client, cluster)

	// Use the operator itself as the CA, for clusters without cert-manager
	case v1.PKIBackendOperator:
		return operatorpki.New(client, cluster)

	// TODO: Add vault
	// Use vault for pki backend
	/*case v1alpha1.PKIBackendVault:
	return vaultpki.New(client, cluster)*/

	// Return mock backend for testing - cannot be triggered by CR due to enum in api schema
	case MockBackend:
		return newMockPKIManager(client, cluster)

	default:
		return &unsupportedPKIManager{backend: backend}
	}
}

// unsupportedPKIManager stands in for a backend the operator does not implement.
type unsupportedPKIManager struct {
	backend v1.PKIBackend
}

func (u *unsupportedPKIManager) err() error {
	return errorfactory.New(errorfactory.InternalError{},
		fmt.Errorf("PKI backend %q is not implemented", u.backend),
		"unsupported sslSecrets.pkiBackend")
}

func (u *unsupportedPKIManager) ReconcilePKI(context.Context, zap.Logger, *runtime.Scheme, []string) error {
	return u.err()
}

// FinalizePKI succeeds so that a cluster naming an unsupported backend can still be
// deleted: refusing here would leave its finalizer in place for good, and a backend
// that was never implemented provisioned nothing to clean up.
func (u *unsupportedPKIManager) FinalizePKI(context.Context, zap.Logger) error {
	return nil
}

func (u *unsupportedPKIManager) ReconcileUserCertificate(context.Context, zap.Logger, *v1.NifiUser,
	*runtime.Scheme,
) (*pki.UserCertificate, error) {
	return nil, u.err()
}

func (u *unsupportedPKIManager) FinalizeUserCertificate(context.Context, *v1.NifiUser) error {
	return nil
}

func (u *unsupportedPKIManager) GetControllerTLSConfig() (*tls.Config, error) {
	return nil, u.err()
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

func (m *mockPKIManager) ReconcileUserCertificate(ctx context.Context, logger zap.Logger, user *v1.NifiUser, scheme *runtime.Scheme) (*pki.UserCertificate, error) {
	return &pki.UserCertificate{}, nil
}

func (m *mockPKIManager) FinalizeUserCertificate(ctx context.Context, user *v1.NifiUser) error {
	return nil
}

func (m *mockPKIManager) GetControllerTLSConfig() (*tls.Config, error) {
	return &tls.Config{}, nil
}
