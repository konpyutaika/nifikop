package certmanagerpki

import (
	"crypto/tls"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/pki/tlsconfig"
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

// GetControllerTLSConfigFromSecret is retained for callers outside this package; the
// implementation is shared with the other PKI backends.
func GetControllerTLSConfigFromSecret(client client.Client, ref v1.SecretReference) (*tls.Config, error) {
	return tlsconfig.FromSecret(client, ref)
}
