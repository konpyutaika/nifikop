package certmanagerpki

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	certutil "github.com/konpyutaika/nifikop/pkg/util/cert"
	pkicommon "github.com/konpyutaika/nifikop/pkg/util/pki"
)

func newMockUser() *v1.NifiUser {
	user := &v1.NifiUser{}
	user.Name = "test-user"
	user.Namespace = "test-namespace"
	user.Spec = v1.NifiUserSpec{SecretName: "test-secret", IncludeJKS: true}
	return user
}

func newMockNodeUser(nodeID int32, cluster *v1.NifiCluster) *v1.NifiUser {
	user := newMockUser()
	user.Name = pkicommon.GetNodeUserName(cluster, nodeID)
	user.Spec.SecretName = fmt.Sprintf(pkicommon.NodeServerCertTemplate, cluster.Name, nodeID)
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

func TestClusterCertificateForUserLeavesRotationPolicyUnset(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	cert := manager.clusterCertificateForUser(newMockNodeUser(0, cluster), scheme.Scheme)

	if cert.Spec.PrivateKey == nil {
		t.Fatal("expected private key settings to be present")
	}

	if cert.Spec.PrivateKey.RotationPolicy != "" {
		t.Fatalf("expected empty rotation policy, got %q", cert.Spec.PrivateKey.RotationPolicy)
	}
}

func TestReconcileUserCertificate(t *testing.T) {
	manager := newMock(newMockCluster())
	ctx := context.Background()

	manager.client.Create(context.TODO(), newMockUser())
	if _, err := manager.ReconcileUserCertificate(ctx, *log, newMockUser(), scheme.Scheme); err == nil {
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
	if _, err := manager.ReconcileUserCertificate(ctx, *log, newMockUser(), scheme.Scheme); err != nil {
		t.Error("Expected no error, got:", err)
	}

	// Test error conditions
	manager = newMock(newMockCluster())
	manager.client.Create(context.TODO(), newMockUser())
	manager.client.Create(context.TODO(), manager.clusterCertificateForUser(newMockUser(), scheme.Scheme))
	if _, err := manager.ReconcileUserCertificate(ctx, *log, newMockUser(), scheme.Scheme); err == nil {
		t.Error("Expected  error, got nil")
	}
}
