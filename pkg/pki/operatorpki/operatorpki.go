// Package operatorpki implements a PKI backend that issues the cluster's
// certificates from the operator itself, with no dependency on cert-manager.
//
// It exists for clusters where the cert-manager CRDs are not installed and
// cannot be: the cert-manager backend fails there before any node pod is
// created, because reconciling an Issuer returns "no matches for kind".
//
// The secrets it writes are byte-for-byte compatible with the ones the
// cert-manager backend produces, so the rest of the operator, and NiFi itself,
// are unaware of which backend issued them.
package operatorpki

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/util/pki"
)

type OperatorPKI interface {
	pki.Manager
}

// operatorPKI implements a PKIManager using the operator itself as the CA.
type operatorPKI struct {
	client  client.Client
	cluster *v1.NifiCluster
}

func New(client client.Client, cluster *v1.NifiCluster) OperatorPKI {
	return &operatorPKI{client: client, cluster: cluster}
}
