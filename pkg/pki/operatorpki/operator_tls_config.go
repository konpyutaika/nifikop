package operatorpki

import (
	"crypto/tls"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/pki/tlsconfig"
)

// GetControllerTLSConfig builds the mTLS config the operator uses to call the NiFi API,
// from the controller user's secret. The secret is read by name only, so it does not
// matter which backend wrote it.
//
// It refuses configuration the backend does not support, as issuance does: with issuerRef
// set, authenticating as the controller with a certificate from the local CA is the same
// trust mismatch issuance refuses to create.
func (o *operatorPKI) GetControllerTLSConfig() (*tls.Config, error) {
	if err := o.supported(); err != nil {
		return nil, err
	}
	return tlsconfig.FromSecret(o.client, v1.SecretReference{
		Namespace: o.cluster.Namespace,
		Name:      o.cluster.GetNifiControllerUserIdentity(),
	})
}
