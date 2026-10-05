package operatorpki

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"time"

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
	// caCertDuration matches the validity the cert-manager backend asks for, so that
	// switching backend does not change how often the CA rolls. A long-lived CA keeps
	// the key stable so issued leaves keep chaining to it.
	caCertDuration = 87600 * time.Hour // ~10 years
	// caKeySize is the RSA key size for the CA, matching the leaf size cert-manager uses.
	caKeySize = 4096
	// maxCommonNameBytes is the X.509 upper bound on a common name (RFC 5280 ub-common-name).
	maxCommonNameBytes = 64
)

// certificateAuthority is the signing material backing every certificate this backend
// issues. The key is held as a crypto.Signer so that a CA supplied by a user can be
// RSA or ECDSA; keys this backend generates itself are always RSA.
type certificateAuthority struct {
	Certificate *x509.Certificate
	PrivateKey  crypto.Signer
	CertPEM     []byte
	KeyPEM      []byte
}

// caSecretName is the secret this backend keeps its CA in. It matches the name the
// cert-manager backend gives the secret behind its CA Certificate, so the two agree
// on where the cluster CA lives.
func (o *operatorPKI) caSecretName() string {
	return fmt.Sprintf(pkicommon.NodeCACertTemplate, o.cluster.Name)
}

// ensureCA returns the cluster CA, generating and persisting one if needed.
//
// When sslSecrets.create is false the CA is taken from the user-provided secret at
// sslSecrets.tlsSecretName, which must hold caCert and caKey - the same contract the
// cert-manager backend documents.
func (o *operatorPKI) ensureCA(ctx context.Context, scheme *runtime.Scheme) (*certificateAuthority, error) {
	if !o.cluster.GetSSLSecrets().Create {
		return o.userProvidedCA(ctx, scheme)
	}

	existing := &corev1.Secret{}
	err := o.client.Get(ctx, types.NamespacedName{Name: o.caSecretName(), Namespace: o.cluster.Namespace}, existing)
	switch {
	case err == nil:
		// A CA secret another object controls is someone else's key: never sign with it.
		if err := refuseForeign(o.cluster, existing, scheme); err != nil {
			return nil, err
		}
		ca, parseErr := caFromSecret(existing)
		if parseErr != nil {
			// Replacing the CA here would invalidate every certificate issued from it,
			// so an unreadable secret is surfaced rather than overwritten.
			return nil, errorfactory.New(errorfactory.InternalError{}, parseErr,
				fmt.Sprintf("secret %s exists but does not hold readable CA material; it must be repaired or deleted deliberately",
					o.caSecretName()))
		}
		if err := ca.usable(); err != nil {
			// Replacing it here would re-sign every node at once with no overlapping
			// trust, breaking cluster communication mid-transition. An expired CA is a
			// deliberate operation, not something to paper over.
			return nil, errorfactory.New(errorfactory.InternalError{},
				fmt.Errorf("CA in secret %s cannot be used: %w", o.caSecretName(), err),
				"cluster CA must be rotated deliberately")
		}
		// An adopted CA - one cert-manager left behind - carries no owner. Taking
		// ownership is what lets FinalizePKI remove it, and the key in it, with the
		// cluster, now that FinalizePKI only deletes what the cluster controls.
		if err := o.claimExisting(ctx, existing, scheme); err != nil {
			return nil, err
		}
		return ca, nil
	case apierrors.IsNotFound(err):
		return o.generateCA(ctx, nil, scheme)
	default:
		return nil, errorfactory.New(errorfactory.APIFailure{}, err, "could not look up CA secret")
	}
}

// getCA returns the cluster CA without creating one, for paths that must not mint
// material of their own.
func (o *operatorPKI) getCA(ctx context.Context, scheme *runtime.Scheme) (*certificateAuthority, error) {
	secret := &corev1.Secret{}
	err := o.client.Get(ctx, types.NamespacedName{Name: o.caSecretName(), Namespace: o.cluster.Namespace}, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errorfactory.New(errorfactory.ResourceNotReady{}, err, "cluster CA not created yet")
		}
		return nil, errorfactory.New(errorfactory.APIFailure{}, err, "could not look up CA secret")
	}
	if err := refuseForeign(o.cluster, secret, scheme); err != nil {
		return nil, err
	}
	ca, err := caFromSecret(secret)
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not read cluster CA")
	}
	if err := ca.usable(); err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "cluster CA cannot sign certificates")
	}
	return ca, nil
}

// userProvidedCA loads a CA supplied by the user and mirrors it into the operator's
// CA secret, so issuance has a single place to read from either way. The mirror holds
// the CA private key and is owned by the NifiCluster; the user's own secret is
// never written to and never deleted.
func (o *operatorPKI) userProvidedCA(ctx context.Context, scheme *runtime.Scheme) (*certificateAuthority, error) {
	sourceName := o.cluster.GetSSLSecrets().TLSSecretName
	if sourceName == o.caSecretName() {
		// The mirror would be written over the user's own secret, and then deleted
		// with the cluster. Neither is acceptable for something the user owns.
		return nil, errorfactory.New(errorfactory.InternalError{},
			fmt.Errorf("sslSecrets.tlsSecretName must not be %q: that name holds the operator's copy of the CA",
				o.caSecretName()),
			"provided tls secret name collides with the operator CA secret")
	}

	provided := &corev1.Secret{}
	err := o.client.Get(ctx,
		types.NamespacedName{Name: sourceName, Namespace: o.cluster.Namespace},
		provided)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errorfactory.New(errorfactory.ResourceNotReady{}, err, "could not find provided tls secret")
		}
		return nil, errorfactory.New(errorfactory.APIFailure{}, err, "could not lookup provided tls secret")
	}

	certPEM := provided.Data[v1.CACertKey]
	keyPEM := provided.Data[v1.CAPrivateKeyKey]
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return nil, errorfactory.New(errorfactory.ResourceNotReady{},
			fmt.Errorf("secret %s must contain %s and %s", provided.Name, v1.CACertKey, v1.CAPrivateKeyKey),
			"provided tls secret is missing CA material")
	}

	ca, err := parseCA(certPEM, keyPEM)
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not read provided CA material")
	}
	if err := ca.usable(); err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{},
			fmt.Errorf("CA in %s cannot be used: %w", provided.Name, err),
			"provided CA certificate cannot sign")
	}

	if err := o.persistCA(ctx, ca, nil, scheme); err != nil {
		return nil, err
	}
	return ca, nil
}

// generateCA mints a new self-signed CA and stores it.
func (o *operatorPKI) generateCA(ctx context.Context, existing *corev1.Secret,
	scheme *runtime.Scheme,
) (*certificateAuthority, error) {
	key, err := rsa.GenerateKey(rand.Reader, caKeySize)
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not generate CA key")
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not generate CA serial")
	}

	commonName := truncateCommonName(fmt.Sprintf(pkicommon.CAFQDNTemplate,
		o.cluster.Name, o.cluster.Namespace, o.cluster.Spec.ListenersConfig.GetClusterDomain()))

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(caCertDuration),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not create CA certificate")
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not parse generated CA certificate")
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, errorfactory.New(errorfactory.InternalError{}, err, "could not encode CA key")
	}

	ca := &certificateAuthority{
		Certificate: cert,
		PrivateKey:  key,
		CertPEM:     pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		KeyPEM:      pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
	}

	if err := o.persistCA(ctx, ca, existing, scheme); err != nil {
		return nil, err
	}
	return ca, nil
}

// persistCA writes the CA into its secret, in the same shape the cert-manager backend
// leaves behind so anything reading the CA secret works with either backend. The secret
// is owned by the NifiCluster, so the CA private key in it is garbage collected with
// the cluster rather than left behind.
func (o *operatorPKI) persistCA(ctx context.Context, ca *certificateAuthority, existing *corev1.Secret,
	scheme *runtime.Scheme,
) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.caSecretName(),
			Namespace: o.cluster.Namespace,
			Labels:    pkicommon.LabelsForNifiPKI(o.cluster.Name),
		},
		Data: map[string][]byte{
			v1.CoreCACertKey:        ca.CertPEM,
			corev1.TLSCertKey:       ca.CertPEM,
			corev1.TLSPrivateKeyKey: ca.KeyPEM,
		},
	}

	if existing == nil {
		current := &corev1.Secret{}
		err := o.client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, current)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return errorfactory.New(errorfactory.APIFailure{}, err, "could not look up CA secret")
			}
			if err := claim(o.cluster, secret, scheme); err != nil {
				return err
			}
			if err := o.client.Create(ctx, secret); err != nil {
				return errorfactory.New(errorfactory.APIFailure{}, err, "could not create CA secret")
			}
			return nil
		}
		existing = current
	}

	updated := existing.DeepCopy()
	if updated.Data == nil {
		updated.Data = map[string][]byte{}
	}
	for k, v := range secret.Data {
		updated.Data[k] = v
	}
	updated.Labels = secret.Labels
	if err := claim(o.cluster, updated, scheme); err != nil {
		return err
	}

	// ReconcilePKI runs on every cluster reconcile, so writing unconditionally would be a
	// write per requeue interval for a secret that almost never changes.
	if reflect.DeepEqual(updated.Data, existing.Data) && reflect.DeepEqual(updated.Labels, existing.Labels) &&
		reflect.DeepEqual(updated.OwnerReferences, existing.OwnerReferences) {
		return nil
	}

	if err := o.client.Update(ctx, updated); err != nil {
		return errorfactory.New(errorfactory.APIFailure{}, err, "could not update CA secret")
	}
	return nil
}

// caFromSecret reads CA material previously written by persistCA.
func caFromSecret(secret *corev1.Secret) (*certificateAuthority, error) {
	certPEM := secret.Data[corev1.TLSCertKey]
	if len(certPEM) == 0 {
		certPEM = secret.Data[v1.CoreCACertKey]
	}
	return parseCA(certPEM, secret.Data[corev1.TLSPrivateKeyKey])
}

func parseCA(certPEM, keyPEM []byte) (*certificateAuthority, error) {
	cert, err := certutil.DecodeCertificate(certPEM)
	if err != nil {
		return nil, err
	}

	key, err := parsePrivateKey(keyPEM)
	if err != nil {
		return nil, err
	}

	comparable, ok := cert.PublicKey.(interface{ Equal(crypto.PublicKey) bool })
	if !ok {
		return nil, fmt.Errorf("unsupported CA public key type %T", cert.PublicKey)
	}
	if !comparable.Equal(key.Public()) {
		return nil, errors.New("CA private key does not match the CA certificate")
	}

	return &certificateAuthority{
		Certificate: cert,
		PrivateKey:  key,
		CertPEM:     certPEM,
		KeyPEM:      keyPEM,
	}, nil
}

// parsePrivateKey reads a PEM encoded private key in any of the encodings a user
// might supply. certutil.DecodeKey is deliberately not used: it assumes RSA.
func parsePrivateKey(keyPEM []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, errors.New("failed to decode PEM data")
	}
	return parsePrivateKeyDER(block.Bytes)
}

// parsePrivateKeyDER reads a DER encoded private key in PKCS1, SEC1 or PKCS8 form.
func parsePrivateKeyDER(der []byte) (crypto.Signer, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, err
	}
	signer, ok := parsed.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("unsupported private key type %T", parsed)
	}
	return signer, nil
}

// usable reports whether the CA is fit to sign with right now. Every path that loads a
// CA checks this: the cluster controller reaches one through ensureCA and the NifiUser
// controller through getCA, so a check in only one of them lets the other issue
// certificates from an expired CA.
func (ca *certificateAuthority) usable() error {
	if !ca.Certificate.IsCA {
		return errors.New("certificate is not a CA and cannot sign")
	}
	// RFC 5280 lets a CA certificate omit the keyUsage extension entirely, which Go
	// reports as zero, so only a present extension that leaves out certificate signing
	// is refused. Such a CA would issue certificates that fail chain validation later.
	if ca.Certificate.KeyUsage != 0 && ca.Certificate.KeyUsage&x509.KeyUsageCertSign == 0 {
		return errors.New("CA certificate's key usage does not permit certificate signing")
	}
	now := time.Now()
	if now.Before(ca.Certificate.NotBefore) {
		return fmt.Errorf("CA is not valid until %s", ca.Certificate.NotBefore)
	}
	if now.After(ca.Certificate.NotAfter) {
		return fmt.Errorf("CA expired at %s", ca.Certificate.NotAfter)
	}
	return nil
}

// truncateCommonName keeps a generated common name inside the X.509 limit. It is only
// applied to the CA, whose name nothing authorises against - node and user common names
// are identities and are left exactly as requested.
func truncateCommonName(name string) string {
	if len(name) <= maxCommonNameBytes {
		return name
	}
	return name[:maxCommonNameBytes]
}

// claimExisting takes ownership of an existing CA secret, writing only if that changes
// anything so that a settled cluster generates no writes.
func (o *operatorPKI) claimExisting(ctx context.Context, existing *corev1.Secret, scheme *runtime.Scheme) error {
	owned := existing.DeepCopy()
	if err := claim(o.cluster, owned, scheme); err != nil {
		return err
	}
	if reflect.DeepEqual(owned.OwnerReferences, existing.OwnerReferences) {
		return nil
	}
	if err := o.client.Update(ctx, owned); err != nil {
		return errorfactory.New(errorfactory.APIFailure{}, err, "could not take ownership of CA secret")
	}
	return nil
}
