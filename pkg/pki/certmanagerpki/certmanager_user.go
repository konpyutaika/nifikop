package certmanagerpki

import (
	"context"
	"errors"
	"fmt"

	certv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	certutil "github.com/konpyutaika/nifikop/pkg/util/cert"
	pkicommon "github.com/konpyutaika/nifikop/pkg/util/pki"
)

// FinalizeUserCertificate for cert-manager backend auto returns because controller references handle cleanup.
func (c *certManager) FinalizeUserCertificate(ctx context.Context, user *v1.NifiUser) (err error) {
	return
}

// ReconcileUserCertificate ensures a certificate/secret combination using cert-manager.
func (c *certManager) ReconcileUserCertificate(ctx context.Context, user *v1.NifiUser, scheme *runtime.Scheme) (*pkicommon.UserCertificate, error) {
	var err error
	var secret *corev1.Secret
	// See if we have an existing certificate for this user already
	_, err = c.getUserCertificate(ctx, user)

	if err != nil && apierrors.IsNotFound(err) {
		// the certificate does not exist, let's make one
		// check if jks is required and create password for it
		if user.Spec.IncludeJKS {
			if err := c.injectJKSPassword(ctx, user); err != nil {
				return nil, err
			}
		}
		cert := c.clusterCertificateForUser(user, scheme)
		if err = c.client.Create(ctx, cert); err != nil {
			return nil, errorfactory.New(errorfactory.APIFailure{}, err, "could not create user certificate")
		}
	} else if err != nil {
		// API failure, requeue
		return nil, errorfactory.New(errorfactory.APIFailure{}, err, "failed looking up user certificate")
	}

	// Get the secret created from the certificate
	secret, err = c.getUserSecret(ctx, user)
	if err != nil {
		return nil, err
	}

	// Ensure controller reference on user secret
	if err = c.ensureControllerReference(ctx, user, secret, scheme); err != nil {
		return nil, err
	}

	return &pkicommon.UserCertificate{
		CA:          secret.Data[v1.CoreCACertKey],
		Certificate: secret.Data[corev1.TLSCertKey],
		Key:         secret.Data[corev1.TLSPrivateKeyKey],
	}, nil
}

// injectJKSPassword ensures that a secret contains JKS password when requested.
func (c *certManager) injectJKSPassword(ctx context.Context, user *v1.NifiUser) error {
	var err error
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      user.Spec.SecretName,
			Namespace: user.Namespace,
		},
		Data: map[string][]byte{},
	}
	secret, err = certutil.EnsureSecretPassJKS(secret)
	if err != nil {
		return errorfactory.New(errorfactory.InternalError{}, err, "could not inject secret with jks password")
	}
	if err = c.client.Create(ctx, secret); err != nil {
		return errorfactory.New(errorfactory.APIFailure{}, err, "could not create secret with jks password")
	}

	return nil
}

// ensureControllerReference ensures that a NifiUser owns a given Secret.
func (c *certManager) ensureControllerReference(ctx context.Context, user *v1.NifiUser, secret *corev1.Secret, scheme *runtime.Scheme) error {
	err := controllerutil.SetControllerReference(user, secret, scheme)
	if err != nil && !k8sutil.IsAlreadyOwnedError(err) {
		return errorfactory.New(errorfactory.InternalError{}, err, "error checking controller reference on user secret")
	} else if err == nil {
		if err = c.client.Update(ctx, secret); err != nil {
			return errorfactory.New(errorfactory.APIFailure{}, err, "could not update secret with controller reference")
		}
	}
	return nil
}

// getUserCertificate fetches the cert-manager Certificate for a user.
func (c *certManager) getUserCertificate(ctx context.Context, user *v1.NifiUser) (*certv1.Certificate, error) {
	cert := &certv1.Certificate{}
	err := c.client.Get(ctx, types.NamespacedName{Name: user.Name, Namespace: user.Namespace}, cert)
	return cert, err
}

// getUserSecret fetches the secret created from a cert-manager Certificate for a user.
func (c *certManager) getUserSecret(ctx context.Context, user *v1.NifiUser) (secret *corev1.Secret, err error) {
	secret = &corev1.Secret{}
	err = c.client.Get(ctx, types.NamespacedName{Name: user.Spec.SecretName, Namespace: user.Namespace}, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return secret, errorfactory.New(errorfactory.ResourceNotReady{}, err, "user secret not ready")
		}
		return secret, errorfactory.New(errorfactory.APIFailure{}, err, "failed to get user secret")
	}
	if user.Spec.IncludeJKS {
		if len(secret.Data) != 6 {
			return secret, errorfactory.New(errorfactory.ResourceNotReady{}, err, "user secret not populated yet")
		}
	} else {
		if len(secret.Data) != 3 {
			return secret, errorfactory.New(errorfactory.ResourceNotReady{}, err, "user secret not populated yet")
		}
	}

	for _, v := range secret.Data {
		if len(v) == 0 {
			return secret, errorfactory.New(errorfactory.ResourceNotReady{},
				errors.New("not all secret value populated"), "secret is not ready")
		}
	}

	return secret, nil
}

// clusterCertificateForUser generates a Certificate object for a NifiUser.
func (c *certManager) clusterCertificateForUser(user *v1.NifiUser, scheme *runtime.Scheme) *certv1.Certificate {
	caName, caKind, caGroup := c.getCA()
	cert := &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      user.GetName(),
			Namespace: user.GetNamespace(),
		},
		Spec: certv1.CertificateSpec{
			SecretName: user.Spec.SecretName,
			CommonName: user.GetName(),
			URIs:       []string{fmt.Sprintf(pkicommon.SpiffeIdTemplate, c.cluster.Name, user.GetNamespace(), user.GetName())},
			Usages:     []certv1.KeyUsage{certv1.UsageClientAuth, certv1.UsageServerAuth},
			IssuerRef: certmeta.ObjectReference{
				Name:  caName,
				Kind:  caKind,
				Group: caGroup,
			},
			PrivateKey: &certv1.CertificatePrivateKey{
				Encoding:  certv1.PKCS8,
				Algorithm: certv1.RSAKeyAlgorithm,
				Size:      4096,
			},
		},
	}
	if user.Spec.IncludeJKS {
		cert.Spec.Keystores = &certv1.CertificateKeystores{
			JKS: &certv1.JKSKeystore{
				Create: true,
				PasswordSecretRef: certmeta.SecretKeySelector{
					LocalObjectReference: certmeta.LocalObjectReference{
						Name: user.Spec.SecretName,
					},
					Key: v1.PasswordKey,
				},
			},
		}
	}
	if user.Spec.DNSNames != nil && len(user.Spec.DNSNames) > 0 {
		cert.Spec.DNSNames = user.Spec.DNSNames
	}
	controllerutil.SetControllerReference(user, cert, scheme)
	return cert
}

// getCA returns the CA name/kind/group for the NifiCluster.
func (c *certManager) getCA() (caName, caKind, caGroup string) {
	caKind = certv1.IssuerKind
	issuerRef := c.cluster.Spec.ListenersConfig.SSLSecrets.IssuerRef
	if issuerRef != nil {
		caName = issuerRef.Name
		caKind = issuerRef.Kind
		caGroup = issuerRef.Group
	} else {
		if c.cluster.Spec.ListenersConfig.SSLSecrets.ClusterScoped {
			caKind = certv1.ClusterIssuerKind
		}
		caName = fmt.Sprintf(pkicommon.NodeIssuerTemplate, c.cluster.Name)
	}
	// TODO: Do we need to ensure this Issuer is exist?
	return
}
