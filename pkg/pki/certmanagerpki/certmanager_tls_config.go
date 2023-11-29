package certmanagerpki

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
)

// GetControllerTLSConfig creates a TLS config from the user secret created for
// cruise control and manager operations.
func (c *certManager) GetControllerTLSConfig() (config *tls.Config, err error) {
	config, err = GetControllerTLSConfigFromSecret(c.client, v1.SecretReference{
		Namespace: c.cluster.Namespace,
		Name:      c.cluster.GetNifiControllerUserIdentity(),
	})
	return
}

func GetControllerTLSConfigFromSecret(client client.Client, ref v1.SecretReference) (config *tls.Config, err error) {
	config = &tls.Config{}
	tlsKeys := &corev1.Secret{}
	err = client.Get(context.TODO(),
		types.NamespacedName{
			Namespace: ref.Namespace,
			Name:      ref.Name,
		},
		tlsKeys,
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = errorfactory.New(errorfactory.ResourceNotReady{}, err, "controller secret not found")
		}
		return
	}
	clientCert := tlsKeys.Data[corev1.TLSCertKey]
	clientKey := tlsKeys.Data[corev1.TLSPrivateKeyKey]
	caCert := tlsKeys.Data[v1.CoreCACertKey]

	if len(caCert) == 0 {
		certs := strings.SplitAfter(string(clientCert), "-----END CERTIFICATE-----")
		clientCert = []byte(certs[0])
		caCert = []byte(certs[len(certs)-1])
		if len(certs) == 3 {
			caCert = []byte(certs[len(certs)-2])
		}
	}

	x509ClientCert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		err = errorfactory.New(errorfactory.InternalError{}, err, "could not decode controller certificate")
		return
	}

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(caCert)

	config.Certificates = []tls.Certificate{x509ClientCert}
	config.RootCAs = rootCAs

	return
}
