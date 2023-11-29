package certmanagerpki

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	certutil "github.com/konpyutaika/nifikop/pkg/util/cert"
)

func newMockUser() *v1.NifiUser {
	user := &v1.NifiUser{}
	user.Name = "test-user"
	user.Namespace = "test-namespace"
	user.Spec = v1.NifiUserSpec{SecretName: "test-secret", IncludeJKS: true}
	return user
}

func newMockUserSecret() *corev1.Secret {
	secret := &corev1.Secret{}
	secret.Name = "test-secret"
	secret.Namespace = "test-namespace"
	cert, key, _, _ := certutil.GenerateTestCert()
	secret.Data = map[string][]byte{
		corev1.TLSCertKey:       cert,
		corev1.TLSPrivateKeyKey: key,
		v1.TLSJKSKeyStore:       []byte("testkeystore"),
		v1.PasswordKey:          []byte("testpassword"),
		v1.TLSJKSTrustStore:     []byte("testtruststore"),
		v1.CoreCACertKey:        cert,
	}
	return secret
}

func TestFinalizeUserCertificate(t *testing.T) {
	manager := newMock(newMockCluster())
	if err := manager.FinalizeUserCertificate(context.Background(), &v1.NifiUser{}); err != nil {
		t.Error("Expected no error, got:", err)
	}
}

func TestReconcileUserCertificate(t *testing.T) {
	manager := newMock(newMockCluster())
	ctx := context.Background()

	manager.client.Create(context.TODO(), newMockUser())
	if _, err := manager.ReconcileUserCertificate(ctx, newMockUser(), scheme.Scheme); err == nil {
		t.Error("Expected resource not ready error, got nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errorfactory.ResourceNotReady{}) {
		t.Error("Expected resource not ready error, got:", reflect.TypeOf(err))
	}
	if err := manager.client.Delete(context.TODO(), newMockUserSecret()); err != nil {
		t.Error("could not delete test secret")
	}
	if err := manager.client.Create(context.TODO(), newMockUserSecret()); err != nil {
		t.Error("could not update test secret")
	}
	if _, err := manager.ReconcileUserCertificate(ctx, newMockUser(), scheme.Scheme); err != nil {
		t.Error("Expected no error, got:", err)
	}

	// Test error conditions
	manager = newMock(newMockCluster())
	manager.client.Create(context.TODO(), newMockUser())
	manager.client.Create(context.TODO(), manager.clusterCertificateForUser(newMockUser(), scheme.Scheme))
	if _, err := manager.ReconcileUserCertificate(ctx, newMockUser(), scheme.Scheme); err == nil {
		t.Error("Expected  error, got nil")
	}
}
