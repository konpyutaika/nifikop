package operatorpki

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"reflect"
	"testing"
	"time"

	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	keystore "github.com/pavel-v-chernykh/keystore-go"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	certutil "github.com/konpyutaika/nifikop/pkg/util/cert"
	pkicommon "github.com/konpyutaika/nifikop/pkg/util/pki"
)

var log = zap.NewNop()

func newMockCluster() *v1.NifiCluster {
	cluster := &v1.NifiCluster{}
	cluster.Name = "test"
	cluster.Namespace = "test-namespace"
	cluster.UID = "test-cluster-uid"
	cluster.Spec = v1.NifiClusterSpec{}
	cluster.Spec.ListenersConfig = &v1.ListenersConfig{}
	cluster.Spec.ListenersConfig.InternalListeners = []v1.InternalListenerConfig{
		{ContainerPort: 9092},
	}
	cluster.Spec.ListenersConfig.SSLSecrets = &v1.SSLSecrets{
		// Not "test-controller": that is also the controller user's default certificate
		// secret, so under create:false issuing the controller certificate would write a
		// leaf over the user's CA - which the backend now refuses, and which the
		// fixture copied from the cert-manager tests was silently doing.
		TLSSecretName: "test-provided-ca",
		Create:        true,
	}
	cluster.Spec.Nodes = []v1.Node{{Id: 0}, {Id: 1}, {Id: 2}}
	return cluster
}

func newMock(cluster *v1.NifiCluster, objs ...client.Object) *operatorPKI {
	v1.SchemeBuilder.AddToScheme(scheme.Scheme)
	return &operatorPKI{
		cluster: cluster,
		client:  fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objs...).Build(),
	}
}

func TestNew(t *testing.T) {
	if got := New(nil, newMockCluster()); reflect.TypeOf(got) != reflect.TypeOf(&operatorPKI{}) {
		t.Error("Expected an operatorPKI from New, got:", reflect.TypeOf(got))
	}
}

// TestReconcilePKICreatesCAAndUsers is the core guarantee: the backend brings up a PKI
// without any cert-manager object existing.
func TestReconcilePKICreatesCAAndUsers(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	ca := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: "test-ca-certificate", Namespace: cluster.Namespace}, ca); err != nil {
		t.Fatal("Expected the CA secret to exist, got:", err)
	}
	for _, key := range []string{v1.CoreCACertKey, corev1.TLSCertKey, corev1.TLSPrivateKeyKey} {
		if len(ca.Data[key]) == 0 {
			t.Error("Expected CA secret to contain", key)
		}
	}

	cert, err := certutil.DecodeCertificate(ca.Data[corev1.TLSCertKey])
	if err != nil {
		t.Fatal("Expected a parseable CA certificate, got:", err)
	}
	if !cert.IsCA {
		t.Error("Expected the generated certificate to be a CA")
	}

	users := &v1.NifiUserList{}
	if err := manager.client.List(ctx, users); err != nil {
		t.Fatal("Expected to list users, got:", err)
	}
	// One controller user plus one per node.
	if len(users.Items) != len(cluster.Spec.Nodes)+1 {
		t.Error("Expected", len(cluster.Spec.Nodes)+1, "users, got:", len(users.Items))
	}
}

// TestReconcilePKIIsIdempotent guards against the CA being replaced on every pass, which
// would invalidate every certificate already issued from it.
func TestReconcilePKIIsIdempotent(t *testing.T) {
	manager := newMock(newMockCluster())
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	first, err := manager.getCA(ctx, scheme.Scheme)
	if err != nil {
		t.Fatal("Expected the CA to exist, got:", err)
	}

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	second, err := manager.getCA(ctx, scheme.Scheme)
	if err != nil {
		t.Fatal("Expected the CA to exist, got:", err)
	}

	if !bytes.Equal(first.CertPEM, second.CertPEM) {
		t.Error("Expected the CA to be stable across reconciles")
	}
}

func TestReconcileUserCertificate(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	user := pkicommon.ControllerUserForCluster(cluster)
	cert, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme)
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if len(cert.Certificate) == 0 || len(cert.Key) == 0 || len(cert.CA) == 0 {
		t.Fatal("Expected a fully populated user certificate")
	}

	secret := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}

	// NiFi and the operator between them read all six of these.
	for _, key := range []string{
		corev1.TLSCertKey, corev1.TLSPrivateKeyKey, v1.CoreCACertKey,
		v1.TLSJKSKeyStore, v1.TLSJKSTrustStore, v1.PasswordKey,
	} {
		if len(secret.Data[key]) == 0 {
			t.Error("Expected user secret to contain", key)
		}
	}

	leaf, err := certutil.DecodeCertificate(secret.Data[corev1.TLSCertKey])
	if err != nil {
		t.Fatal("Expected a parseable certificate, got:", err)
	}
	if leaf.Subject.CommonName != user.GetName() {
		t.Error("Expected common name", user.GetName(), "got:", leaf.Subject.CommonName)
	}

	ca, err := manager.getCA(ctx, scheme.Scheme)
	if err != nil {
		t.Fatal("Expected the CA to exist, got:", err)
	}
	if err := leaf.CheckSignatureFrom(ca.Certificate); err != nil {
		t.Error("Expected the certificate to be signed by the cluster CA, got:", err)
	}
	if err := verifyExtKeyUsage(leaf); err != nil {
		t.Error(err)
	}
}

// TestReconcileUserCertificateReusesMaterial checks a reconcile does not hand NiFi a new
// keystore, and a new password, every time it runs.
func TestReconcileUserCertificateReusesMaterial(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	user := pkicommon.ControllerUserForCluster(cluster)
	first, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme)
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	second, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme)
	if err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	if !bytes.Equal(first.Certificate, second.Certificate) {
		t.Error("Expected the certificate to be reused rather than reissued")
	}
}

// TestUserKeyStoresOpenWithStoredPassword is what NiFi actually does at startup.
func TestUserKeyStoresOpenWithStoredPassword(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	secret := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	password := secret.Data[v1.PasswordKey]

	ks, err := keystore.Decode(bytes.NewReader(secret.Data[v1.TLSJKSKeyStore]), password)
	if err != nil {
		t.Fatal("Expected the keystore to open with the stored password, got:", err)
	}
	entry, ok := ks[certutil.JKSKeyAlias].(*keystore.PrivateKeyEntry)
	if !ok {
		t.Fatal("Expected a private key entry in the keystore")
	}
	if len(entry.CertChain) < 2 {
		t.Error("Expected the keystore chain to hold the certificate and its issuer, got:", len(entry.CertChain))
	}

	ts, err := keystore.Decode(bytes.NewReader(secret.Data[v1.TLSJKSTrustStore]), password)
	if err != nil {
		t.Fatal("Expected the truststore to open with the stored password, got:", err)
	}
	if _, ok := ts[certutil.JKSCAAlias].(*keystore.TrustedCertificateEntry); !ok {
		t.Error("Expected the CA to be trusted in the truststore")
	}
}

func TestGetControllerTLSConfig(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if _, err := manager.ReconcileUserCertificate(ctx, *log,
		pkicommon.ControllerUserForCluster(cluster), scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	config, err := manager.GetControllerTLSConfig()
	if err != nil {
		t.Fatal("Expected a TLS config, got:", err)
	}
	if len(config.Certificates) != 1 {
		t.Error("Expected one client certificate, got:", len(config.Certificates))
	}
	if config.RootCAs == nil {
		t.Error("Expected the cluster CA to be trusted")
	}
}

// TestUserProvidedCA covers sslSecrets.create=false, where the user supplies the
// signing CA in the documented caCert/caKey form.
func TestUserProvidedCA(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false

	caCert, caKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}

	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName,
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
	}

	manager := newMock(cluster, provided)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	ca, err := manager.getCA(ctx, scheme.Scheme)
	if err != nil {
		t.Fatal("Expected the CA to exist, got:", err)
	}
	if !bytes.Equal(ca.CertPEM, caCert) {
		t.Error("Expected the provided CA to be used as-is")
	}

	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	// The user's secret is the source of truth and is never written to.
	kept := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: provided.Name, Namespace: cluster.Namespace}, kept); err != nil {
		t.Fatal("Expected the user's secret to exist, got:", err)
	}
	if !reflect.DeepEqual(kept.Data, provided.Data) || len(kept.OwnerReferences) != 0 {
		t.Error("Expected the user's CA secret to be left exactly as supplied")
	}
}

// TestUserProvidedCARequiresMaterial makes the failure legible rather than a nil panic.
func TestUserProvidedCARequiresMaterial(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false

	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName,
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{},
	}

	manager := newMock(cluster, provided)
	if err := manager.ReconcilePKI(context.Background(), *log, scheme.Scheme, []string{}); err == nil {
		t.Error("Expected an error when the provided secret holds no CA material")
	}
}

func TestFinalizePKI(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if err := manager.FinalizePKI(ctx, *log); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	ca := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: "test-ca-certificate", Namespace: cluster.Namespace}, ca); err == nil {
		t.Error("Expected the CA secret to be removed")
	}
}

// TestFinalizePKIKeepsProvidedCA: a CA the user supplied is not ours to delete.
func TestFinalizePKIKeepsProvidedCA(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false

	caCert, caKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName,
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
	}

	manager := newMock(cluster, provided)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if err := manager.FinalizePKI(ctx, *log); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	kept := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: provided.Name, Namespace: cluster.Namespace}, kept); err != nil {
		t.Error("Expected the provided secret to be left alone, got:", err)
	}
}

func verifyExtKeyUsage(cert *x509.Certificate) error {
	var client, server bool
	for _, usage := range cert.ExtKeyUsage {
		switch usage {
		case x509.ExtKeyUsageClientAuth:
			client = true
		case x509.ExtKeyUsageServerAuth:
			server = true
		}
	}
	if !client || !server {
		return errNotBothUsages
	}
	return nil
}

var errNotBothUsages = errors.New("expected the certificate to allow both client and server auth")

// generateTestCA builds a CA in the caCert/caKey PEM form a user would supply.
func generateTestCA() (certPEM, keyPEM []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "provided-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), nil
}

// TestUnreadableCAIsNotReplaced: overwriting a CA the operator cannot parse would
// invalidate every certificate already issued from it, so it must be reported instead.
func TestUnreadableCAIsNotReplaced(t *testing.T) {
	cluster := newMockCluster()
	corrupt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ca-certificate", Namespace: cluster.Namespace},
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte("not a certificate"),
			corev1.TLSPrivateKeyKey: []byte("not a key"),
		},
	}
	manager := newMock(cluster, corrupt)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err == nil {
		t.Error("Expected an error rather than a silently replaced CA")
	}

	kept := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: corrupt.Name, Namespace: cluster.Namespace}, kept); err != nil {
		t.Fatal("Expected the secret to still exist, got:", err)
	}
	if !bytes.Equal(kept.Data[corev1.TLSCertKey], corrupt.Data[corev1.TLSCertKey]) {
		t.Error("Expected the unreadable CA material to be left untouched")
	}
}

// TestProvidedECDSACA: a user-supplied CA is not necessarily RSA, and an unchecked type
// assertion would take the operator down rather than reject it.
func TestProvidedECDSACA(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false

	caCert, caKey, err := generateTestECDSACA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName,
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
	}

	manager := newMock(cluster, provided)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected an ECDSA CA to be usable, got:", err)
	}

	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if _, err := manager.GetControllerTLSConfig(); err != nil {
		t.Error("Expected a usable TLS config, got:", err)
	}
}

// TestMismatchedKeyIsReissued: a secret whose fields are all populated but whose key does
// not match its certificate would otherwise be reused forever.
func TestMismatchedKeyIsReissued(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	secret := &corev1.Secret{}
	name := types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}
	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}

	_, strangerKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build replacement material:", err)
	}
	secret.Data[corev1.TLSPrivateKeyKey] = strangerKey
	if err := manager.client.Update(ctx, secret); err != nil {
		t.Fatal("Could not seed the broken secret:", err)
	}

	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if _, err := manager.GetControllerTLSConfig(); err != nil {
		t.Error("Expected the broken material to be reissued, got:", err)
	}
}

// TestCorruptKeyStoreIsReissued: NiFi will not start if its keystore cannot be opened
// with the password beside it.
func TestCorruptKeyStoreIsReissued(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	secret := &corev1.Secret{}
	name := types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}
	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	secret.Data[v1.TLSJKSKeyStore] = []byte("not a keystore")
	if err := manager.client.Update(ctx, secret); err != nil {
		t.Fatal("Could not seed the broken secret:", err)
	}

	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	if _, err := keystore.Decode(bytes.NewReader(secret.Data[v1.TLSJKSKeyStore]),
		secret.Data[v1.PasswordKey]); err != nil {
		t.Error("Expected the keystore to be rebuilt, got:", err)
	}
}

// TestReusedSecretKeepsOwnership: FinalizeUserCertificate relies on garbage collection,
// which only happens if the NifiUser owns the secret - including on the reuse path.
func TestReusedSecretKeepsOwnership(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	secret := &corev1.Secret{}
	name := types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}
	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	secret.OwnerReferences = nil
	if err := manager.client.Update(ctx, secret); err != nil {
		t.Fatal("Could not strip the owner reference:", err)
	}

	// The material is still valid, so this takes the reuse path.
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	if len(secret.OwnerReferences) == 0 {
		t.Error("Expected the reuse path to restore the NifiUser controller reference")
	}
}

// generateTestECDSACA builds a non-RSA CA in the caCert/caKey PEM form.
func generateTestECDSACA() (certPEM, keyPEM []byte, err error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "provided-ecdsa-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), nil
}

// TestSteadyStateDoesNotWrite: ReconcilePKI and ReconcileUserCertificate run on every
// requeue for every node, so a settled cluster must not generate API writes.
func TestSteadyStateDoesNotWrite(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	versions := func() (string, string) {
		ca := &corev1.Secret{}
		if err := manager.client.Get(ctx,
			types.NamespacedName{Name: "test-ca-certificate", Namespace: cluster.Namespace}, ca); err != nil {
			t.Fatal("Expected the CA secret to exist, got:", err)
		}
		secret := &corev1.Secret{}
		if err := manager.client.Get(ctx,
			types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}, secret); err != nil {
			t.Fatal("Expected the user secret to exist, got:", err)
		}
		return ca.ResourceVersion, secret.ResourceVersion
	}

	caBefore, userBefore := versions()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	caAfter, userAfter := versions()
	if caBefore != caAfter {
		t.Error("Expected the CA secret to be left alone on a settled reconcile")
	}
	if userBefore != userAfter {
		t.Error("Expected the user secret to be left alone on a settled reconcile")
	}
}

// TestAdoptsExistingCertManagerCA: a cluster moving off cert-manager already has a CA
// secret in this name and shape. Reusing it means the certificates cert-manager issued
// stay valid, instead of every node being re-signed by a brand new CA at once.
func TestAdoptsExistingCertManagerCA(t *testing.T) {
	cluster := newMockCluster()

	caCert, caKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	// The shape cert-manager leaves behind for its CA Certificate.
	inherited := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ca-certificate", Namespace: cluster.Namespace},
		Data: map[string][]byte{
			corev1.TLSCertKey:       caCert,
			corev1.TLSPrivateKeyKey: caKey,
			v1.CoreCACertKey:        caCert,
		},
	}

	manager := newMock(cluster, inherited)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	ca, err := manager.getCA(ctx, scheme.Scheme)
	if err != nil {
		t.Fatal("Expected the CA to exist, got:", err)
	}
	if !bytes.Equal(ca.CertPEM, caCert) {
		t.Error("Expected the existing CA to be adopted rather than replaced")
	}

	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	secret := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	leaf, err := certutil.DecodeCertificate(secret.Data[corev1.TLSCertKey])
	if err != nil {
		t.Fatal("Expected a parseable certificate, got:", err)
	}
	if err := leaf.CheckSignatureFrom(ca.Certificate); err != nil {
		t.Error("Expected certificates to chain to the adopted CA, got:", err)
	}
}

// TestLeafNeverOutlivesCA: with a short-lived CA - an older cert-manager install used a
// 90 day one - a fixed leaf duration would issue certificates that are invalid for most
// of their stated life.
func TestLeafNeverOutlivesCA(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false

	caCert, caKey, err := generateTestCA() // valid for 24 hours
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName,
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
	}

	manager := newMock(cluster, provided)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected a short-lived CA to still be usable, got:", err)
	}

	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	secret := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	leaf, err := certutil.DecodeCertificate(secret.Data[corev1.TLSCertKey])
	if err != nil {
		t.Fatal("Expected a parseable certificate, got:", err)
	}
	ca, err := certutil.DecodeCertificate(caCert)
	if err != nil {
		t.Fatal("Expected a parseable CA, got:", err)
	}

	if leaf.NotAfter.After(ca.NotAfter) {
		t.Errorf("Expected the certificate to expire with its CA at %s, got %s", ca.NotAfter, leaf.NotAfter)
	}
}

// TestExpiredCADoesNotIssue: the NifiUser controller reaches the CA through getCA, not
// ensureCA, so a guard in only one of them lets expired material keep signing.
func TestExpiredCADoesNotIssue(t *testing.T) {
	cluster := newMockCluster()

	caCert, caKey, err := generateTestCAWithLifetime(-48*time.Hour, -24*time.Hour)
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	expired := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ca-certificate", Namespace: cluster.Namespace},
		Data: map[string][]byte{
			corev1.TLSCertKey:       caCert,
			corev1.TLSPrivateKeyKey: caKey,
			v1.CoreCACertKey:        caCert,
		},
	}

	manager := newMock(cluster, expired)
	ctx := context.Background()
	user := pkicommon.ControllerUserForCluster(cluster)

	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err == nil {
		t.Error("Expected an expired CA to refuse to issue")
	}

	secret := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}, secret); err == nil {
		t.Error("Expected no certificate to be written from an expired CA")
	}
}

// TestExpiredProvidedCARejected: the same check has to apply to a CA a user supplies.
func TestExpiredProvidedCARejected(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false

	caCert, caKey, err := generateTestCAWithLifetime(-48*time.Hour, -24*time.Hour)
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName,
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
	}

	manager := newMock(cluster, provided)
	if err := manager.ReconcilePKI(context.Background(), *log, scheme.Scheme, []string{}); err == nil {
		t.Error("Expected an expired supplied CA to be rejected")
	}
}

// TestEmptyKeyStoreIsReissued: an empty keystore decodes perfectly well, so a
// decode-only check would keep it and leave NiFi unable to start.
func TestEmptyKeyStoreIsReissued(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	secret := &corev1.Secret{}
	name := types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}
	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}

	var empty bytes.Buffer
	if err := keystore.Encode(&empty, keystore.KeyStore{}, secret.Data[v1.PasswordKey]); err != nil {
		t.Fatal("Could not build an empty keystore:", err)
	}
	secret.Data[v1.TLSJKSKeyStore] = empty.Bytes()
	if err := manager.client.Update(ctx, secret); err != nil {
		t.Fatal("Could not seed the empty keystore:", err)
	}

	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	rebuilt, err := keystore.Decode(bytes.NewReader(secret.Data[v1.TLSJKSKeyStore]), secret.Data[v1.PasswordKey])
	if err != nil {
		t.Fatal("Expected a readable keystore, got:", err)
	}
	entry, ok := rebuilt[certutil.JKSKeyAlias].(*keystore.PrivateKeyEntry)
	if !ok || len(entry.PrivKey) == 0 {
		t.Error("Expected the keystore to be rebuilt with its private key entry")
	}
}

// TestStaleKeyStoreIsReissued: stores holding superseded credentials decode too.
func TestStaleKeyStoreIsReissued(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	secret := &corev1.Secret{}
	name := types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}
	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	original := append([]byte{}, secret.Data[v1.TLSJKSKeyStore]...)

	// A keystore built from unrelated credentials, under the same password.
	strangerCert, strangerKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build replacement material:", err)
	}
	stale, err := certutil.GenerateJKSKeyStore(strangerCert, strangerKey, strangerCert, secret.Data[v1.PasswordKey])
	if err != nil {
		t.Fatal("Could not build a stale keystore:", err)
	}
	secret.Data[v1.TLSJKSKeyStore] = stale
	if err := manager.client.Update(ctx, secret); err != nil {
		t.Fatal("Could not seed the stale keystore:", err)
	}

	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	if bytes.Equal(secret.Data[v1.TLSJKSKeyStore], stale) {
		t.Error("Expected the stale keystore to be replaced")
	}
	if bytes.Equal(secret.Data[v1.TLSJKSKeyStore], original) {
		t.Log("keystore rebuilt (byte-identical rebuild is acceptable)")
	}
}

// generateTestCAWithLifetime builds a CA valid over an explicit window, offsets relative
// to now.
func generateTestCAWithLifetime(notBefore, notAfter time.Duration) (certPEM, keyPEM []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(notBefore),
		NotAfter:              time.Now().Add(notAfter),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), nil
}

// ---- Fix 1: issuerRef ----

// TestIssuerRefRejected: issuerRef names a cert-manager Issuer. Swapping it for a CA this
// backend generates would change the cluster's trust root without anyone choosing it.
func TestIssuerRefRejected(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.IssuerRef = &cmmeta.ObjectReference{
		Name: "company-pki", Kind: "ClusterIssuer",
	}
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err == nil {
		t.Error("Expected issuerRef to be refused")
	}
	ca := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: "test-ca-certificate", Namespace: cluster.Namespace}, ca); err == nil {
		t.Error("Expected no local CA to be generated in place of the requested issuer")
	}

	// The NifiUser controller reaches issuance without ReconcilePKI, so it must refuse too.
	if _, err := manager.ReconcileUserCertificate(ctx, *log,
		pkicommon.ControllerUserForCluster(cluster), scheme.Scheme); err == nil {
		t.Error("Expected certificate issuance to refuse issuerRef as well")
	}
}

// ---- Fix 2: no listeners config ----

// TestFinalizePKIWithoutListeners: listenersConfig is optional, and finalizing must not
// panic for a cluster that leaves it, or its sslSecrets, out.
func TestFinalizePKIWithoutListeners(t *testing.T) {
	for name, mutate := range map[string]func(*v1.NifiCluster){
		"no sslSecrets":      func(c *v1.NifiCluster) { c.Spec.ListenersConfig.SSLSecrets = nil },
		"no listenersConfig": func(c *v1.NifiCluster) { c.Spec.ListenersConfig = nil },
	} {
		t.Run(name, func(t *testing.T) {
			cluster := newMockCluster()
			mutate(cluster)
			if err := newMock(cluster).FinalizePKI(context.Background(), *log); err != nil {
				t.Error("Expected no error, got:", err)
			}
		})
	}
}

// ---- Fix 3: the mirrored CA ----

// TestFinalizePKIRemovesProvidedCAMirror: the copy of a user-supplied CA holds its private key,
// so it must go with the cluster - while the user's own secret stays.
func TestFinalizePKIRemovesProvidedCAMirror(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false

	caCert, caKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName,
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
	}
	manager := newMock(cluster, provided)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	mirrorName := types.NamespacedName{Name: "test-ca-certificate", Namespace: cluster.Namespace}
	mirror := &corev1.Secret{}
	if err := manager.client.Get(ctx, mirrorName, mirror); err != nil {
		t.Fatal("Expected the CA mirror to exist, got:", err)
	}
	if !metav1.IsControlledBy(mirror, cluster) {
		t.Error("Expected the CA mirror to be owned by the NifiCluster, so it is garbage collected with it")
	}

	if err := manager.FinalizePKI(ctx, *log); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if err := manager.client.Get(ctx, mirrorName, mirror); err == nil {
		t.Error("Expected the CA mirror, and the private key in it, to be removed")
	}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: provided.Name, Namespace: cluster.Namespace}, &corev1.Secret{}); err != nil {
		t.Error("Expected the user's own secret to be left alone, got:", err)
	}
}

// TestFinalizePKILeavesUnownedCASecret: only a secret this cluster controls is deleted.
func TestFinalizePKILeavesUnownedCASecret(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false

	caCert, caKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName,
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
	}
	controller := true
	foreign := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ca-certificate",
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "v1", Kind: "ConfigMap", Name: "someone-else", UID: "other-uid", Controller: &controller,
			}},
		},
	}
	manager := newMock(cluster, provided, foreign)
	ctx := context.Background()

	if err := manager.FinalizePKI(ctx, *log); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: foreign.Name, Namespace: cluster.Namespace}, &corev1.Secret{}); err != nil {
		t.Error("Expected a secret controlled by something else to be left alone, got:", err)
	}
}

// TestProvidedSecretNameCollisionRejected: a user's secret given the operator's CA
// name would be overwritten by the mirror, and then deleted with the cluster.
func TestProvidedSecretNameCollisionRejected(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false
	cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName = "test-ca-certificate"

	caCert, caKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ca-certificate", Namespace: cluster.Namespace},
		Data:       map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
	}
	manager := newMock(cluster, provided)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err == nil {
		t.Error("Expected the colliding secret name to be refused")
	}
	kept := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: provided.Name, Namespace: cluster.Namespace}, kept); err != nil {
		t.Fatal("Expected the user's secret to still exist, got:", err)
	}
	if len(kept.Data) != 2 || len(kept.OwnerReferences) != 0 {
		t.Error("Expected the user's secret to be left exactly as it was")
	}
}

// TestProvidedCASteadyStateDoesNotWrite: the mirror is refreshed on every reconcile, so
// an unchanged CA must not produce a write.
func TestProvidedCASteadyStateDoesNotWrite(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false

	caCert, caKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName,
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
	}
	manager := newMock(cluster, provided)
	ctx := context.Background()
	mirrorName := types.NamespacedName{Name: "test-ca-certificate", Namespace: cluster.Namespace}

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	before := &corev1.Secret{}
	if err := manager.client.Get(ctx, mirrorName, before); err != nil {
		t.Fatal("Expected the CA mirror to exist, got:", err)
	}
	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	after := &corev1.Secret{}
	if err := manager.client.Get(ctx, mirrorName, after); err != nil {
		t.Fatal("Expected the CA mirror to exist, got:", err)
	}
	if before.ResourceVersion != after.ResourceVersion {
		t.Error("Expected an unchanged CA mirror not to be rewritten")
	}
}

// ---- Fix 4: the full requested identity ----

// TestRemovedSANTriggersReissue: a SAN dropped from the request must leave the
// certificate too, rather than stay valid until expiry.
func TestRemovedSANTriggersReissue(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	user := pkicommon.ControllerUserForCluster(cluster)
	user.Spec.DNSNames = []string{"a.example", "b.example"}
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	user.Spec.DNSNames = []string{"a.example"}
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	secret := &corev1.Secret{}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	leaf, err := certutil.DecodeCertificate(secret.Data[corev1.TLSCertKey])
	if err != nil {
		t.Fatal("Expected a parseable certificate, got:", err)
	}
	if len(leaf.DNSNames) != 1 || leaf.DNSNames[0] != "a.example" {
		t.Error("Expected the removed SAN to be dropped from the certificate, got:", leaf.DNSNames)
	}
}

// TestIdentityMismatchIsReissued: a certificate differing from what would be issued in any
// identity property is replaced. The unmodified case is the control: it proves the
// crafted secret is otherwise reusable, so the other cases fail on the property under test.
func TestIdentityMismatchIsReissued(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*x509.Certificate)
		reissue bool
	}{
		{"unmodified", func(*x509.Certificate) {}, false},
		{"wrong SPIFFE id", func(c *x509.Certificate) {
			u, _ := url.Parse("spiffe://elsewhere/ns/x/nifiuser/y")
			c.URIs = []*url.URL{u}
		}, true},
		{"missing server auth", func(c *x509.Certificate) {
			c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		}, true},
		{"extra subject field", func(c *x509.Certificate) {
			c.Subject.Organization = []string{"someone-else"}
		}, true},
		{"not yet valid", func(c *x509.Certificate) {
			c.NotBefore = time.Now().Add(time.Hour)
			c.NotAfter = time.Now().Add(48 * time.Hour)
		}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cluster := newMockCluster()
			manager := newMock(cluster)
			ctx := context.Background()

			if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
				t.Fatal("Expected no error, got:", err)
			}
			user := pkicommon.ControllerUserForCluster(cluster)
			user.Spec.IncludeJKS = false // isolates the identity checks from the keystore ones

			crafted := craftUserSecret(t, manager, user, tc.mutate)
			if err := manager.client.Create(ctx, crafted); err != nil {
				t.Fatal("Could not seed the crafted secret:", err)
			}

			if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
				t.Fatal("Expected no error, got:", err)
			}

			secret := &corev1.Secret{}
			if err := manager.client.Get(ctx,
				types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}, secret); err != nil {
				t.Fatal("Expected the user secret to exist, got:", err)
			}
			reissued := !bytes.Equal(secret.Data[corev1.TLSCertKey], crafted.Data[corev1.TLSCertKey])
			if reissued != tc.reissue {
				t.Errorf("Expected reissue=%v, got %v", tc.reissue, reissued)
			}
		})
	}
}

// ---- Fix 7: key usage ----

func TestProvidedCAKeyUsage(t *testing.T) {
	cases := []struct {
		name     string
		usage    x509.KeyUsage
		accepted bool
	}{
		{"cert sign", x509.KeyUsageCertSign | x509.KeyUsageCRLSign, true},
		// RFC 5280 lets a CA omit the extension; Go reports that as zero.
		{"no key usage extension", 0, true},
		{"present without cert sign", x509.KeyUsageDigitalSignature, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cluster := newMockCluster()
			cluster.Spec.ListenersConfig.SSLSecrets.Create = false

			caCert, caKey, err := generateCustomCA(func(c *x509.Certificate) { c.KeyUsage = tc.usage })
			if err != nil {
				t.Fatal("Could not build a test CA:", err)
			}
			provided := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName,
					Namespace: cluster.Namespace,
				},
				Data: map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
			}
			err = newMock(cluster, provided).ReconcilePKI(context.Background(), *log, scheme.Scheme, []string{})
			if accepted := err == nil; accepted != tc.accepted {
				t.Errorf("Expected accepted=%v, got error: %v", tc.accepted, err)
			}
		})
	}
}

// ---- Fix 8: the keystore's private key ----

// TestKeyStoreWithWrongPrivateKeyIsReissued: the right certificate with the wrong key in
// the keystore passes every check that only looks at the certificate.
func TestKeyStoreWithWrongPrivateKeyIsReissued(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	secret := &corev1.Secret{}
	name := types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}
	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}

	chain, err := certutil.DecodeCertificateChain(secret.Data[corev1.TLSCertKey])
	if err != nil {
		t.Fatal("Expected a parseable chain, got:", err)
	}
	leafPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: chain[0].Raw})
	_, strangerKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build replacement material:", err)
	}
	tampered, err := certutil.GenerateJKSKeyStore(leafPEM, strangerKey,
		secret.Data[v1.CoreCACertKey], secret.Data[v1.PasswordKey])
	if err != nil {
		t.Fatal("Could not build the tampered keystore:", err)
	}
	secret.Data[v1.TLSJKSKeyStore] = tampered
	if err := manager.client.Update(ctx, secret); err != nil {
		t.Fatal("Could not seed the tampered keystore:", err)
	}

	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	if bytes.Equal(secret.Data[v1.TLSJKSKeyStore], tampered) {
		t.Error("Expected a keystore holding the wrong private key to be rebuilt")
	}
}

// ---- helpers ----

// craftUserSecret signs a certificate exactly as the backend would, applies mutate, and
// returns it as a user secret, so a single property can be varied in isolation.
func craftUserSecret(t *testing.T, manager *operatorPKI, user *v1.NifiUser,
	mutate func(*x509.Certificate),
) *corev1.Secret {
	t.Helper()
	ca, err := manager.getCA(context.Background(), scheme.Scheme)
	if err != nil {
		t.Fatal("Expected the CA to exist, got:", err)
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	spiffe, err := url.Parse(manager.spiffeID(user))
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               subjectFor(user),
		DNSNames:              user.Spec.DNSNames,
		URIs:                  []*url.URL{spiffe},
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           requiredExtKeyUsages,
		BasicConstraintsValid: true,
	}
	mutate(template)
	der, err := x509.CreateCertificate(rand.Reader, template, ca.Certificate, &key.PublicKey, ca.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: user.Spec.SecretName, Namespace: user.Namespace},
		Data: map[string][]byte{
			corev1.TLSCertKey:       append(append([]byte{}, certPEM...), ca.CertPEM...),
			corev1.TLSPrivateKeyKey: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
			v1.CoreCACertKey:        ca.CertPEM,
		},
	}
}

// generateCustomCA builds a self-signed CA with mutate applied to its template.
func generateCustomCA(mutate func(*x509.Certificate)) (certPEM, keyPEM []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: "custom-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	mutate(template)
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), nil
}

// ---- Ownership: a secret another object controls is never used, changed or deleted ----

// foreignController marks a secret as controlled by an object that is not ours.
func foreignController() []metav1.OwnerReference {
	controller := true
	return []metav1.OwnerReference{{
		APIVersion: "v1", Kind: "ConfigMap", Name: "someone-else", UID: "other-uid", Controller: &controller,
	}}
}

// assertUntouched fails unless the secret is exactly as it was seeded: same data, labels,
// owners, and no write at all.
func assertUntouched(t *testing.T, manager *operatorPKI, seeded *corev1.Secret) {
	t.Helper()
	got := &corev1.Secret{}
	if err := manager.client.Get(context.Background(),
		types.NamespacedName{Name: seeded.Name, Namespace: seeded.Namespace}, got); err != nil {
		t.Fatal("Expected the secret to still exist, got:", err)
	}
	if got.ResourceVersion != seeded.ResourceVersion ||
		!reflect.DeepEqual(got.Data, seeded.Data) ||
		!reflect.DeepEqual(got.Labels, seeded.Labels) ||
		!reflect.DeepEqual(got.OwnerReferences, seeded.OwnerReferences) {
		t.Error("Expected a secret controlled by another object to be left byte for byte unchanged")
	}
}

// seed creates a secret and returns it as stored, resource version included.
func seed(t *testing.T, manager *operatorPKI, secret *corev1.Secret) *corev1.Secret {
	t.Helper()
	if err := manager.client.Create(context.Background(), secret); err != nil {
		t.Fatal("Could not seed secret:", err)
	}
	stored := &corev1.Secret{}
	if err := manager.client.Get(context.Background(),
		types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, stored); err != nil {
		t.Fatal("Could not read back seeded secret:", err)
	}
	return stored
}

// TestForeignCAMirrorUntouched: under create:false the mirror is written on every
// reconcile, so a foreign secret at that name must be refused, not overwritten.
func TestForeignCAMirrorUntouched(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false
	caCert, caKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	manager := newMock(cluster, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName, Namespace: cluster.Namespace},
		Data:       map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
	})
	foreign := seed(t, manager, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ca-certificate", Namespace: cluster.Namespace, OwnerReferences: foreignController(),
			Labels: map[string]string{"owner": "someone-else"},
		},
		Data: map[string][]byte{"theirs": []byte("do not touch")},
	})

	if err := manager.ReconcilePKI(context.Background(), *log, scheme.Scheme, []string{}); err == nil {
		t.Error("Expected a foreign-controlled CA secret to be refused")
	}
	assertUntouched(t, manager, foreign)
}

// TestForeignCAAdoptionRefused: a CA another object controls is someone else's key, on
// both signing paths - the cluster controller's and the NifiUser controller's.
func TestForeignCAAdoptionRefused(t *testing.T) {
	cluster := newMockCluster()
	caCert, caKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	manager := newMock(cluster)
	foreign := seed(t, manager, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ca-certificate", Namespace: cluster.Namespace, OwnerReferences: foreignController(),
		},
		Data: map[string][]byte{corev1.TLSCertKey: caCert, corev1.TLSPrivateKeyKey: caKey, v1.CoreCACertKey: caCert},
	})
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err == nil {
		t.Error("Expected the cluster controller to refuse a foreign-controlled CA")
	}
	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err == nil {
		t.Error("Expected the NifiUser controller to refuse to sign with a foreign-controlled CA")
	}
	if err := manager.client.Get(ctx,
		types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}, &corev1.Secret{}); err == nil {
		t.Error("Expected no certificate to be issued from a foreign-controlled CA")
	}
	assertUntouched(t, manager, foreign)
}

// TestForeignUserSecretUntouched covers both user paths: a secret that would be reissued,
// and one that would otherwise be reused as this user's certificate.
func TestForeignUserSecretUntouched(t *testing.T) {
	for _, reusable := range []bool{false, true} {
		t.Run(fmt.Sprintf("reusable=%v", reusable), func(t *testing.T) {
			cluster := newMockCluster()
			manager := newMock(cluster)
			ctx := context.Background()
			if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
				t.Fatal("Expected no error, got:", err)
			}
			user := pkicommon.ControllerUserForCluster(cluster)
			user.Spec.IncludeJKS = false

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: user.Spec.SecretName, Namespace: cluster.Namespace},
				Data:       map[string][]byte{"theirs": []byte("do not touch")},
			}
			if reusable {
				secret = craftUserSecret(t, manager, user, func(*x509.Certificate) {})
			}
			secret.OwnerReferences = foreignController()
			foreign := seed(t, manager, secret)

			if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err == nil {
				t.Error("Expected a foreign-controlled user secret to be refused")
			}
			assertUntouched(t, manager, foreign)
		})
	}
}

// TestFinalizePKIDeletesOnlyOwnedSecrets: every secret lives at a predictable name, so
// only one still carrying the owner this backend gave it is deleted.
func TestFinalizePKIDeletesOnlyOwnedSecrets(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	controllerUser := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, controllerUser, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	nodeSecret := func(id int32) string { return fmt.Sprintf(pkicommon.NodeServerCertTemplate, cluster.Name, id) }
	controller := true
	// Node 0: ours, controlled by its NifiUser. Node 1: unowned. Node 2: someone else's.
	seed(t, manager, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name: nodeSecret(0), Namespace: cluster.Namespace,
		OwnerReferences: []metav1.OwnerReference{{
			APIVersion: v1.GroupVersion.String(), Kind: "NifiUser",
			Name: pkicommon.GetNodeUserName(cluster, 0), UID: "node-user-uid", Controller: &controller,
		}},
	}})
	unowned := seed(t, manager, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: nodeSecret(1), Namespace: cluster.Namespace}})
	foreign := seed(t, manager, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name: nodeSecret(2), Namespace: cluster.Namespace, OwnerReferences: foreignController(),
	}})

	if err := manager.FinalizePKI(ctx, *log); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	for _, gone := range []string{"test-ca-certificate", controllerUser.Spec.SecretName, nodeSecret(0)} {
		if err := manager.client.Get(ctx, types.NamespacedName{Name: gone, Namespace: cluster.Namespace}, &corev1.Secret{}); err == nil {
			t.Error("Expected owned secret to be deleted:", gone)
		}
	}
	assertUntouched(t, manager, unowned)
	assertUntouched(t, manager, foreign)
}

// TestFinalizePKIRemovesAdoptedCA: a CA left by cert-manager carries no owner. Adoption
// has to take ownership, or ownership-checked teardown would leave its key behind.
func TestFinalizePKIRemovesAdoptedCA(t *testing.T) {
	cluster := newMockCluster()
	caCert, caKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	manager := newMock(cluster, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ca-certificate", Namespace: cluster.Namespace},
		Data:       map[string][]byte{corev1.TLSCertKey: caCert, corev1.TLSPrivateKeyKey: caKey, v1.CoreCACertKey: caCert},
	})
	ctx := context.Background()
	name := types.NamespacedName{Name: "test-ca-certificate", Namespace: cluster.Namespace}

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	adopted := &corev1.Secret{}
	if err := manager.client.Get(ctx, name, adopted); err != nil {
		t.Fatal("Expected the CA to exist, got:", err)
	}
	if !metav1.IsControlledBy(adopted, cluster) {
		t.Fatal("Expected adoption to make the NifiCluster the CA's controller")
	}

	// Taking ownership is a single write, not one per reconcile.
	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	again := &corev1.Secret{}
	if err := manager.client.Get(ctx, name, again); err != nil {
		t.Fatal("Expected the CA to exist, got:", err)
	}
	if again.ResourceVersion != adopted.ResourceVersion {
		t.Error("Expected a settled adopted CA not to be rewritten")
	}

	if err := manager.FinalizePKI(ctx, *log); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if err := manager.client.Get(ctx, name, &corev1.Secret{}); err == nil {
		t.Error("Expected the adopted CA, and its key, to be removed with the cluster")
	}
}

// ---- A user-supplied CA is authoritative on every issuance ----

func providedCAFixture(t *testing.T) (*v1.NifiCluster, *operatorPKI, *corev1.Secret) {
	t.Helper()
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false
	caCert, caKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName, Namespace: cluster.Namespace},
		Data:       map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
	}
	manager := newMock(cluster, provided)
	if err := manager.ReconcilePKI(context.Background(), *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if _, err := manager.ReconcileUserCertificate(context.Background(), *log,
		pkicommon.ControllerUserForCluster(cluster), scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	return cluster, manager, provided
}

// TestWithdrawnProvidedCAStopsIssuance: a user deleting their CA secret is withdrawing
// it. The operator's copy must not let the NifiUser controller carry on regardless.
func TestWithdrawnProvidedCAStopsIssuance(t *testing.T) {
	cluster, manager, provided := providedCAFixture(t)
	ctx := context.Background()

	if err := manager.client.Delete(ctx, provided); err != nil {
		t.Fatal("Could not withdraw the CA:", err)
	}
	// Only the NifiUser path runs: the cluster has not reconciled since.
	if _, err := manager.ReconcileUserCertificate(ctx, *log,
		pkicommon.ControllerUserForCluster(cluster), scheme.Scheme); err == nil {
		t.Error("Expected issuance to stop once the provided CA is withdrawn, even with a valid certificate on hand")
	}
}

// TestRotatedProvidedCAReachesUserPath: a rotation is picked up by the NifiUser
// controller directly, not only after the next cluster reconcile.
func TestRotatedProvidedCAReachesUserPath(t *testing.T) {
	cluster, manager, provided := providedCAFixture(t)
	ctx := context.Background()

	newCert, newKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a test CA:", err)
	}
	current := &corev1.Secret{}
	if err := manager.client.Get(ctx, types.NamespacedName{Name: provided.Name, Namespace: cluster.Namespace}, current); err != nil {
		t.Fatal(err)
	}
	current.Data = map[string][]byte{v1.CACertKey: newCert, v1.CAPrivateKeyKey: newKey}
	if err := manager.client.Update(ctx, current); err != nil {
		t.Fatal("Could not rotate the CA:", err)
	}

	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	secret := &corev1.Secret{}
	if err := manager.client.Get(ctx, types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}, secret); err != nil {
		t.Fatal(err)
	}
	leaf, err := certutil.DecodeCertificate(secret.Data[corev1.TLSCertKey])
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := certutil.DecodeCertificate(newCert)
	if err != nil {
		t.Fatal(err)
	}
	if err := leaf.CheckSignatureFrom(rotated); err != nil {
		t.Error("Expected the certificate to be reissued from the rotated CA, got:", err)
	}
}

// ---- Reserved secret names ----

// TestUserSecretNameReserved: both CA secrets are refused as user certificate secrets,
// since writing a leaf over either destroys the CA key. The user's secret name is
// only reserved while it actually holds the CA, under create:false.
func TestUserSecretNameReserved(t *testing.T) {
	cases := []struct {
		name    string
		create  bool
		secret  func(c *v1.NifiCluster) string
		refused bool
	}{
		{"cluster CA secret", true, func(*v1.NifiCluster) string { return "test-ca-certificate" }, true},
		{"provided CA secret", false, func(c *v1.NifiCluster) string { return c.Spec.ListenersConfig.SSLSecrets.TLSSecretName }, true},
		{"tlsSecretName while the CA is generated", true, func(c *v1.NifiCluster) string { return c.Spec.ListenersConfig.SSLSecrets.TLSSecretName }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cluster := newMockCluster()
			cluster.Spec.ListenersConfig.SSLSecrets.Create = tc.create
			var objs []client.Object
			if !tc.create {
				caCert, caKey, err := generateTestCA()
				if err != nil {
					t.Fatal(err)
				}
				objs = append(objs, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName, Namespace: cluster.Namespace},
					Data:       map[string][]byte{v1.CACertKey: caCert, v1.CAPrivateKeyKey: caKey},
				})
			}
			manager := newMock(cluster, objs...)
			ctx := context.Background()
			if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
				t.Fatal("Expected no error, got:", err)
			}

			target := &corev1.Secret{}
			targetName := types.NamespacedName{Name: tc.secret(cluster), Namespace: cluster.Namespace}
			before, _ := func() (*corev1.Secret, error) {
				err := manager.client.Get(ctx, targetName, target)
				return target.DeepCopy(), err
			}()

			user := pkicommon.ControllerUserForCluster(cluster)
			user.Spec.SecretName = tc.secret(cluster)
			_, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme)
			if refused := err != nil; refused != tc.refused {
				t.Fatalf("Expected refused=%v, got error: %v", tc.refused, err)
			}
			if tc.refused {
				after := &corev1.Secret{}
				if err := manager.client.Get(ctx, targetName, after); err != nil {
					t.Fatal("Expected the CA secret to still exist, got:", err)
				}
				if !reflect.DeepEqual(after.Data, before.Data) {
					t.Error("Expected the CA secret's contents to be left alone")
				}
			}
		})
	}
}

// ---- includeJKS turned off ----

// TestDisablingJKSStripsStores: once a user stops asking for keystores, reuse drops them
// rather than keeping them forever - without reissuing the certificate.
func TestDisablingJKSStripsStores(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()
	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	name := types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}
	before := &corev1.Secret{}
	if err := manager.client.Get(ctx, name, before); err != nil {
		t.Fatal(err)
	}

	user.Spec.IncludeJKS = false
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	after := &corev1.Secret{}
	if err := manager.client.Get(ctx, name, after); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{v1.TLSJKSKeyStore, v1.TLSJKSTrustStore, v1.PasswordKey} {
		if _, ok := after.Data[key]; ok {
			t.Error("Expected keystore material to be removed:", key)
		}
	}
	if !bytes.Equal(after.Data[corev1.TLSCertKey], before.Data[corev1.TLSCertKey]) {
		t.Error("Expected the certificate itself to be kept, not reissued")
	}

	// And settled after that.
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	settled := &corev1.Secret{}
	if err := manager.client.Get(ctx, name, settled); err != nil {
		t.Fatal(err)
	}
	if settled.ResourceVersion != after.ResourceVersion {
		t.Error("Expected no further writes once the stores are gone")
	}
}

// ---- Controller TLS config follows the same support rules ----

func TestControllerTLSConfigRefusesIssuerRef(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()
	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if _, err := manager.ReconcileUserCertificate(ctx, *log, pkicommon.ControllerUserForCluster(cluster), scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if _, err := manager.GetControllerTLSConfig(); err != nil {
		t.Fatal("Expected a TLS config before the change, got:", err)
	}

	cluster.Spec.ListenersConfig.SSLSecrets.IssuerRef = &cmmeta.ObjectReference{Name: "company-pki", Kind: "ClusterIssuer"}
	if _, err := manager.GetControllerTLSConfig(); err == nil {
		t.Error("Expected the controller to stop authenticating with local-CA material once issuerRef is set")
	}
}

// TestFinalizePKISkipsSecretChangedAfterCheck: teardown deletes only the object whose
// ownership it checked. A secret rewritten between that check and the delete is refused,
// survives the pass, and is checked again on the next one.
func TestFinalizePKISkipsSecretChangedAfterCheck(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	caName := types.NamespacedName{Name: "test-ca-certificate", Namespace: cluster.Namespace}

	rewritten := false
	manager.client = interceptor.NewClient(manager.client.(client.WithWatch), interceptor.Funcs{
		Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
			if obj.GetName() == caName.Name && !rewritten {
				rewritten = true
				// Another writer updates the secret after FinalizePKI has read it.
				current := &corev1.Secret{}
				if err := c.Get(ctx, caName, current); err != nil {
					return err
				}
				current.Labels = map[string]string{"rewritten": "true"}
				if err := c.Update(ctx, current); err != nil {
					return err
				}
			}
			return c.Delete(ctx, obj, opts...)
		},
	})

	if err := manager.FinalizePKI(ctx, *log); !apierrors.IsConflict(err) {
		t.Fatal("Expected the delete to be refused as a conflict, got:", err)
	}
	if err := manager.client.Get(ctx, caName, &corev1.Secret{}); err != nil {
		t.Fatal("Expected the rewritten secret to survive, got:", err)
	}

	// The next pass reads it again, still finds it owned, and removes it.
	if err := manager.FinalizePKI(ctx, *log); err != nil {
		t.Fatal("Expected no error on the next pass, got:", err)
	}
	if err := manager.client.Get(ctx, caName, &corev1.Secret{}); !apierrors.IsNotFound(err) {
		t.Error("Expected the secret to be deleted on the next pass, got:", err)
	}
}

// TestFinalizePKISendsDeletePreconditions: both preconditions reach the API server. The
// fake client enforces only the resourceVersion one, so the UID one is checked as sent.
func TestFinalizePKISendsDeletePreconditions(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	caName := types.NamespacedName{Name: "test-ca-certificate", Namespace: cluster.Namespace}
	ca := &corev1.Secret{}
	if err := manager.client.Get(ctx, caName, ca); err != nil {
		t.Fatal("Expected the CA secret to exist, got:", err)
	}
	// The fake client assigns no UIDs; give the secret one, as an API server would have.
	ca.UID = "ca-secret-uid"
	if err := manager.client.Update(ctx, ca); err != nil {
		t.Fatal("Could not set the secret's UID:", err)
	}
	if err := manager.client.Get(ctx, caName, ca); err != nil {
		t.Fatal("Expected the CA secret to exist, got:", err)
	}

	var sent *client.DeleteOptions
	manager.client = interceptor.NewClient(manager.client.(client.WithWatch), interceptor.Funcs{
		Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
			if obj.GetName() == caName.Name {
				sent = (&client.DeleteOptions{}).ApplyOptions(opts)
			}
			return c.Delete(ctx, obj, opts...)
		},
	})

	if err := manager.FinalizePKI(ctx, *log); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if sent == nil || sent.Preconditions == nil {
		t.Fatal("Expected the CA secret to be deleted with preconditions")
	}
	if uid := sent.Preconditions.UID; uid == nil || *uid != ca.UID {
		t.Error("Expected a UID precondition matching the checked secret, got:", uid)
	}
	if rv := sent.Preconditions.ResourceVersion; rv == nil || *rv != ca.ResourceVersion {
		t.Error("Expected a resourceVersion precondition matching the checked secret, got:", rv)
	}
}

// TestProvidedCABundleIsStable: a supplied CA that carries its own issuer yields the
// chain leaf, intermediate, root - and that chain is reused, not reissued every pass.
func TestProvidedCABundleIsStable(t *testing.T) {
	cluster := newMockCluster()
	cluster.Spec.ListenersConfig.SSLSecrets.Create = false

	rootCert, rootKey, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build a root CA:", err)
	}
	intermediateCert, intermediateKey, err := generateTestIntermediate(rootCert, rootKey)
	if err != nil {
		t.Fatal("Could not build an intermediate CA:", err)
	}
	bundle := append(append([]byte{}, intermediateCert...), rootCert...)
	provided := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName,
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{v1.CACertKey: bundle, v1.CAPrivateKeyKey: intermediateKey},
	}

	manager := newMock(cluster, provided)
	ctx := context.Background()
	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	user := pkicommon.ControllerUserForCluster(cluster)
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}

	secret := &corev1.Secret{}
	name := types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}
	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	ks, err := keystore.Decode(bytes.NewReader(secret.Data[v1.TLSJKSKeyStore]), secret.Data[v1.PasswordKey])
	if err != nil {
		t.Fatal("Expected the keystore to open, got:", err)
	}
	entry, ok := ks[certutil.JKSKeyAlias].(*keystore.PrivateKeyEntry)
	if !ok {
		t.Fatal("Expected a private key entry in the keystore")
	}
	want, err := certutil.DecodeCertificateChain(bundle)
	if err != nil {
		t.Fatal("Could not decode the bundle:", err)
	}
	if len(entry.CertChain) != 3 ||
		!bytes.Equal(entry.CertChain[1].Content, want[0].Raw) || !bytes.Equal(entry.CertChain[2].Content, want[1].Raw) {
		t.Fatal("Expected the keystore chain to be leaf, intermediate, root; got entries:", len(entry.CertChain))
	}

	before := secret.ResourceVersion
	if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
		t.Fatal("Expected no error, got:", err)
	}
	if err := manager.client.Get(ctx, name, secret); err != nil {
		t.Fatal("Expected the user secret to exist, got:", err)
	}
	if secret.ResourceVersion != before {
		t.Error("Expected a valid bundled chain to be reused, but the secret was rewritten")
	}
}

// generateTestIntermediate builds a CA signed by the given root.
func generateTestIntermediate(rootCertPEM, rootKeyPEM []byte) (certPEM, keyPEM []byte, err error) {
	root, err := certutil.DecodeCertificate(rootCertPEM)
	if err != nil {
		return nil, nil, err
	}
	rootKey, err := parsePrivateKey(rootKeyPEM)
	if err != nil {
		return nil, nil, err
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "test-intermediate-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(5 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, root, &key.PublicKey, rootKey)
	if err != nil {
		return nil, nil, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), nil
}

// TestKeyStoreWithWrongChainIsReissued: a keystore whose chain is anything other than the
// leaf followed by this CA is rebuilt, even when the leaf and its key still match.
func TestKeyStoreWithWrongChainIsReissued(t *testing.T) {
	strangerCA, _, err := generateTestCA()
	if err != nil {
		t.Fatal("Could not build an unrelated CA:", err)
	}
	for name, chainCA := range map[string][]byte{
		"foreign issuer": strangerCA,
		"leaf only":      nil,
	} {
		t.Run(name, func(t *testing.T) {
			cluster := newMockCluster()
			manager := newMock(cluster)
			ctx := context.Background()

			if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
				t.Fatal("Expected no error, got:", err)
			}
			user := pkicommon.ControllerUserForCluster(cluster)
			if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
				t.Fatal("Expected no error, got:", err)
			}

			secret := &corev1.Secret{}
			name := types.NamespacedName{Name: user.Spec.SecretName, Namespace: cluster.Namespace}
			if err := manager.client.Get(ctx, name, secret); err != nil {
				t.Fatal("Expected the user secret to exist, got:", err)
			}
			caBlock, _ := pem.Decode(secret.Data[v1.CoreCACertKey])
			leafBlock, _ := pem.Decode(secret.Data[corev1.TLSCertKey])
			if caBlock == nil || leafBlock == nil {
				t.Fatal("Expected PEM certificates in the user secret")
			}

			// Same leaf and key, under the same password, but the wrong chain behind them.
			tampered, err := certutil.GenerateJKSKeyStore(pem.EncodeToMemory(leafBlock),
				secret.Data[corev1.TLSPrivateKeyKey], chainCA, secret.Data[v1.PasswordKey])
			if err != nil {
				t.Fatal("Could not build the tampered keystore:", err)
			}
			secret.Data[v1.TLSJKSKeyStore] = tampered
			if err := manager.client.Update(ctx, secret); err != nil {
				t.Fatal("Could not seed the tampered keystore:", err)
			}

			if _, err := manager.ReconcileUserCertificate(ctx, *log, user, scheme.Scheme); err != nil {
				t.Fatal("Expected no error, got:", err)
			}
			if err := manager.client.Get(ctx, name, secret); err != nil {
				t.Fatal("Expected the user secret to exist, got:", err)
			}
			ks, err := keystore.Decode(bytes.NewReader(secret.Data[v1.TLSJKSKeyStore]), secret.Data[v1.PasswordKey])
			if err != nil {
				t.Fatal("Expected the rebuilt keystore to open, got:", err)
			}
			entry, ok := ks[certutil.JKSKeyAlias].(*keystore.PrivateKeyEntry)
			if !ok {
				t.Fatal("Expected a private key entry in the keystore")
			}
			if len(entry.CertChain) != 2 || !bytes.Equal(entry.CertChain[1].Content, caBlock.Bytes) {
				t.Error("Expected the keystore chain to be rebuilt as the leaf followed by the cluster CA")
			}
		})
	}
}
