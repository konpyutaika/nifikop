package operatorpki

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"reflect"
	"time"

	keystore "github.com/pavel-v-chernykh/keystore-go"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	certutil "github.com/konpyutaika/nifikop/pkg/util/cert"
	pkicommon "github.com/konpyutaika/nifikop/pkg/util/pki"
)

const (
	// leafCertDuration is how long an issued node or user certificate is valid for.
	// It is longer than cert-manager's 90 day default because an operator that is
	// scaled down, or a cluster that is left alone between releases, must not come
	// back to expired TLS, which on a cluster nobody is watching means an outage.
	leafCertDuration = 8760 * time.Hour // 1 year
	// renewAtFraction is how far through its own lifetime a certificate is reissued,
	// matching cert-manager's default of two thirds. It is a fraction rather than a fixed
	// window so that a certificate capped short by its CA is not treated as permanently
	// due for renewal, which would reissue it on every reconcile.
	renewAtFraction = 2.0 / 3.0
	// leafKeySize matches the key size the cert-manager backend requests.
	leafKeySize = 4096
	// jksPasswordLength matches certutil's own generated password length.
	jksPasswordLength = 16
)

// FinalizeUserCertificate is a no-op: user secrets carry a controller reference to the
// NifiUser, so Kubernetes garbage collects them. This mirrors the cert-manager backend.
func (o *operatorPKI) FinalizeUserCertificate(ctx context.Context, user *v1.NifiUser) error {
	return nil
}

// ReconcileUserCertificate issues, renews and stores a user certificate signed by the
// cluster CA, and is idempotent: an existing certificate that is still valid for the
// requested identity is returned untouched.
func (o *operatorPKI) ReconcileUserCertificate(ctx context.Context, logger zap.Logger, user *v1.NifiUser,
	scheme *runtime.Scheme,
) (*pkicommon.UserCertificate, error) {
	if err := o.supported(); err != nil {
		return nil, err
	}
	if err := o.checkUserSecretName(user); err != nil {
		return nil, err
	}

	ca, err := o.signingCA(ctx, scheme)
	if err != nil {
		return nil, err
	}

	existing := &corev1.Secret{}
	err = o.client.Get(ctx, types.NamespacedName{Name: user.Spec.SecretName, Namespace: user.Namespace}, existing)
	switch {
	case err == nil:
		// Refused before anything is read from it or written to it: a secret another
		// object controls is neither this user's certificate nor this backend's to change.
		if err := refuseForeign(user, existing, scheme); err != nil {
			return nil, err
		}
		if o.reusable(existing, user, ca) {
			return o.keep(ctx, user, existing, scheme)
		}
	case apierrors.IsNotFound(err):
		existing = nil
	default:
		return nil, errorfactory.New(errorfactory.APIFailure{}, err, "failed looking up user secret")
	}

	logger.Info("Issuing certificate from the operator-managed CA",
		zap.String("clusterName", o.cluster.Name),
		zap.String("user", user.GetName()))

	data, err := o.issue(user, ca, existingPassword(existing))
	if err != nil {
		return nil, err
	}

	secret := existing.DeepCopy()
	if secret == nil {
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      user.Spec.SecretName,
				Namespace: user.Namespace,
			},
		}
	}
	secret.Data = data
	secret.Labels = pkicommon.LabelsForNifiPKI(o.cluster.Name)

	if err := claim(user, secret, scheme); err != nil {
		return nil, err
	}

	if existing == nil {
		if err := o.client.Create(ctx, secret); err != nil {
			return nil, errorfactory.New(errorfactory.APIFailure{}, err, "could not create user secret")
		}
	} else if err := o.client.Update(ctx, secret); err != nil {
		return nil, errorfactory.New(errorfactory.APIFailure{}, err, "could not update user secret")
	}

	return &pkicommon.UserCertificate{
		CA:          data[v1.CoreCACertKey],
		Certificate: data[corev1.TLSCertKey],
		Key:         data[corev1.TLSPrivateKeyKey],
	}, nil
}

// signingCA returns the CA to issue from. With a user-supplied CA, the user's
// secret is authoritative on every issuance, not only when the cluster reconciles: read
// only the copy, and a CA the user has withdrawn or rotated keeps signing for as
// long as the copy survives. Otherwise the CA is read without being created - minting
// one is the cluster controller's job, and two controllers racing to do it would not do.
func (o *operatorPKI) signingCA(ctx context.Context, scheme *runtime.Scheme) (*certificateAuthority, error) {
	if !o.cluster.GetSSLSecrets().Create {
		return o.userProvidedCA(ctx, scheme)
	}
	return o.getCA(ctx, scheme)
}

// checkUserSecretName refuses the two names a user certificate must never be written to.
// Both hold a CA: writing a leaf certificate over either replaces the CA's private key
// with the leaf's, destroying the CA. Checked by name, independently of ownership, since
// a user's own secret normally has no owner to compare against.
func (o *operatorPKI) checkUserSecretName(user *v1.NifiUser) error {
	name := user.Spec.SecretName
	if name == o.caSecretName() {
		return errorfactory.New(errorfactory.InternalError{},
			fmt.Errorf("user %s secretName %q is the cluster CA secret", user.GetName(), name),
			"user certificate secret collides with the cluster CA")
	}
	if sslSecrets := o.cluster.GetSSLSecrets(); !sslSecrets.Create && name == sslSecrets.TLSSecretName {
		return errorfactory.New(errorfactory.InternalError{},
			fmt.Errorf("user %s secretName %q is the CA secret supplied in sslSecrets.tlsSecretName",
				user.GetName(), name),
			"user certificate secret collides with the provided CA secret")
	}
	return nil
}

// keep returns a reusable certificate, bringing the secret around it up to date without
// reissuing: ownership by the NifiUser, so the secret is garbage collected with it, and
// no keystores once the user has stopped asking for them. It writes only if something
// changed, so a settled cluster generates no writes.
func (o *operatorPKI) keep(ctx context.Context, user *v1.NifiUser, existing *corev1.Secret,
	scheme *runtime.Scheme,
) (*pkicommon.UserCertificate, error) {
	updated := existing.DeepCopy()
	if !user.Spec.IncludeJKS {
		delete(updated.Data, v1.TLSJKSKeyStore)
		delete(updated.Data, v1.TLSJKSTrustStore)
		delete(updated.Data, v1.PasswordKey)
	}
	if err := claim(user, updated, scheme); err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(updated.Data, existing.Data) ||
		!reflect.DeepEqual(updated.OwnerReferences, existing.OwnerReferences) {
		if err := o.client.Update(ctx, updated); err != nil {
			return nil, errorfactory.New(errorfactory.APIFailure{}, err, "could not update user secret")
		}
	}
	return userCertificateFromSecret(updated), nil
}

// issue signs a new certificate for the user and returns the full secret payload.
func (o *operatorPKI) issue(user *v1.NifiUser, ca *certificateAuthority, password []byte) (map[string][]byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, leafKeySize)
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not generate user key")
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not generate user serial")
	}

	spiffeID, err := url.Parse(o.spiffeID(user))
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not build user SPIFFE id")
	}

	now := time.Now()
	notAfter := now.Add(leafCertDuration)
	if notAfter.After(ca.Certificate.NotAfter) {
		// A certificate must not outlive the CA that signed it: past that point it fails
		// validation anyway, and some clients reject the chain sooner.
		notAfter = ca.Certificate.NotAfter
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		// The common name is the NiFi identity this certificate authenticates as, so it
		// is used exactly as requested. Keeping it under 64 bytes is the caller's job,
		// via nodeUserIdentityTemplate or a shorter cluster name.
		Subject:               subjectFor(user),
		DNSNames:              user.Spec.DNSNames,
		URIs:                  []*url.URL{spiffeID},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           requiredExtKeyUsages,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, ca.Certificate, &key.PublicKey, ca.PrivateKey)
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not sign user certificate")
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not encode user key")
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	// tls.crt carries the leaf followed by its issuer, which is what cert-manager's CA
	// issuer produces and what the controller TLS config falls back to parsing.
	chainPEM := append(append([]byte{}, certPEM...), ca.CertPEM...)

	data := map[string][]byte{
		corev1.TLSCertKey:       chainPEM,
		corev1.TLSPrivateKeyKey: keyPEM,
		v1.CoreCACertKey:        ca.CertPEM,
	}

	if !user.Spec.IncludeJKS {
		return data, nil
	}

	if len(password) == 0 {
		password = certutil.GeneratePass(jksPasswordLength)
	}

	keyStore, err := certutil.GenerateJKSKeyStore(certPEM, keyPEM, ca.CertPEM, password)
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not generate user keystore")
	}
	trustStore, err := certutil.GenerateJKSTrustStore(ca.CertPEM, password)
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not generate user truststore")
	}

	data[v1.TLSJKSKeyStore] = keyStore
	data[v1.TLSJKSTrustStore] = trustStore
	data[v1.PasswordKey] = password

	return data, nil
}

// reusable reports whether the certificate already in the secret still satisfies the
// user's request and does not need renewing. Every property issue sets is checked here
// too: a certificate that differs in any of them - a SAN that should have been dropped,
// a different subject - is reissued rather than kept for the rest of its life.
func (o *operatorPKI) reusable(secret *corev1.Secret, user *v1.NifiUser, ca *certificateAuthority) bool {
	required := []string{corev1.TLSCertKey, corev1.TLSPrivateKeyKey, v1.CoreCACertKey}
	if user.Spec.IncludeJKS {
		required = append(required, v1.TLSJKSKeyStore, v1.TLSJKSTrustStore, v1.PasswordKey)
	}
	for _, key := range required {
		if len(secret.Data[key]) == 0 {
			return false
		}
	}

	// The certificate and key have to actually pair up: a secret that was edited or
	// half-written otherwise stays broken forever, because every later reconcile sees
	// the same populated fields and reuses them.
	if _, err := tls.X509KeyPair(secret.Data[corev1.TLSCertKey], secret.Data[corev1.TLSPrivateKeyKey]); err != nil {
		return false
	}

	// A rotated CA invalidates everything it signed.
	if !bytes.Equal(secret.Data[v1.CoreCACertKey], ca.CertPEM) {
		return false
	}

	cert, err := certutil.DecodeCertificate(secret.Data[corev1.TLSCertKey])
	if err != nil {
		return false
	}

	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) || needsRenewal(cert) {
		return false
	}

	// The full subject, not only the common name: it is the identity NiFi authorises.
	if cert.Subject.String() != subjectFor(user).String() {
		return false
	}

	// Compared as sets, so that a SAN removed from the request is removed from the
	// certificate too, as cert-manager does, rather than staying valid until expiry.
	if !sameStringSet(cert.DNSNames, user.Spec.DNSNames) {
		return false
	}

	uris := make([]string, 0, len(cert.URIs))
	for _, uri := range cert.URIs {
		uris = append(uris, uri.String())
	}
	if !sameStringSet(uris, []string{o.spiffeID(user)}) {
		return false
	}

	if !hasExtKeyUsages(cert, requiredExtKeyUsages...) {
		return false
	}

	if user.Spec.IncludeJKS && !keyStoresMatch(secret, cert, ca) {
		return false
	}

	// Confirm the certificate really was signed by the CA we hold.
	return cert.CheckSignatureFrom(ca.Certificate) == nil
}

// existingPassword returns the JKS password already in a secret, so regenerating the
// stores does not change a password NiFi has been started with.
func existingPassword(secret *corev1.Secret) []byte {
	if secret == nil {
		return nil
	}
	return secret.Data[v1.PasswordKey]
}

func userCertificateFromSecret(secret *corev1.Secret) *pkicommon.UserCertificate {
	return &pkicommon.UserCertificate{
		CA:          secret.Data[v1.CoreCACertKey],
		Certificate: secret.Data[corev1.TLSCertKey],
		Key:         secret.Data[corev1.TLSPrivateKeyKey],
	}
}

// keyStoresMatch reports whether the stores in a secret open with the password beside
// them AND actually hold the credentials they are supposed to. Checking only that they
// decode would accept an empty keystore, or one still holding superseded credentials,
// either of which leaves NiFi with unusable TLS despite valid looking PEM fields.
func keyStoresMatch(secret *corev1.Secret, leaf *x509.Certificate, ca *certificateAuthority) bool {
	password := secret.Data[v1.PasswordKey]

	keyStore, err := keystore.Decode(bytes.NewReader(secret.Data[v1.TLSJKSKeyStore]), password)
	if err != nil {
		return false
	}
	entry, ok := keyStore[certutil.JKSKeyAlias].(*keystore.PrivateKeyEntry)
	if !ok || len(entry.PrivKey) == 0 || len(entry.CertChain) == 0 {
		return false
	}
	if !bytes.Equal(entry.CertChain[0].Content, leaf.Raw) {
		return false
	}
	// The chain NiFi presents must be exactly what issue builds: the leaf, then every
	// certificate of the CA bundle in order (a supplied CA may carry its own issuers).
	caChain, err := certutil.DecodeCertificateChain(ca.CertPEM)
	if err != nil || len(entry.CertChain) != 1+len(caChain) {
		return false
	}
	for i, cert := range caChain {
		if !bytes.Equal(entry.CertChain[i+1].Content, cert.Raw) {
			return false
		}
	}
	privateKey, err := parsePrivateKeyDER(entry.PrivKey)
	if err != nil {
		return false
	}
	leafKey, ok := leaf.PublicKey.(interface{ Equal(crypto.PublicKey) bool })
	if !ok || !leafKey.Equal(privateKey.Public()) {
		return false
	}

	trustStore, err := keystore.Decode(bytes.NewReader(secret.Data[v1.TLSJKSTrustStore]), password)
	if err != nil {
		return false
	}
	trusted, ok := trustStore[certutil.JKSCAAlias].(*keystore.TrustedCertificateEntry)
	if !ok {
		return false
	}
	return bytes.Equal(trusted.Certificate.Content, ca.Certificate.Raw)
}

// needsRenewal reports whether a certificate is far enough through its own lifetime to
// be reissued.
func needsRenewal(cert *x509.Certificate) bool {
	lifetime := cert.NotAfter.Sub(cert.NotBefore)
	if lifetime <= 0 {
		return true
	}
	return time.Now().After(cert.NotBefore.Add(time.Duration(float64(lifetime) * renewAtFraction)))
}

// requiredExtKeyUsages are the extended key usages every issued certificate carries:
// nodes and the controller each act as both client and server.
var requiredExtKeyUsages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}

// subjectFor is the subject issued to a user. The common name is the NiFi identity the
// certificate authenticates as, so it is used exactly as requested; keeping it under 64
// bytes is the caller's job, via nodeUserIdentityTemplate or a shorter cluster name.
func subjectFor(user *v1.NifiUser) pkix.Name {
	return pkix.Name{CommonName: user.GetName()}
}

// spiffeID is the URI SAN issued to a user, matching the one the cert-manager backend
// puts on its certificates.
func (o *operatorPKI) spiffeID(user *v1.NifiUser) string {
	return fmt.Sprintf(pkicommon.SpiffeIdTemplate, o.cluster.Name, user.GetNamespace(), user.GetName())
}

// hasExtKeyUsages reports whether a certificate carries every one of the given usages.
func hasExtKeyUsages(cert *x509.Certificate, usages ...x509.ExtKeyUsage) bool {
	for _, want := range usages {
		found := false
		for _, have := range cert.ExtKeyUsage {
			if have == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// sameStringSet reports whether two slices hold the same values, ignoring order and
// duplicates.
func sameStringSet(a, b []string) bool {
	set := func(values []string) map[string]struct{} {
		out := make(map[string]struct{}, len(values))
		for _, v := range values {
			out[v] = struct{}{}
		}
		return out
	}
	sa, sb := set(a), set(b)
	if len(sa) != len(sb) {
		return false
	}
	for v := range sa {
		if _, ok := sb[v]; !ok {
			return false
		}
	}
	return true
}
