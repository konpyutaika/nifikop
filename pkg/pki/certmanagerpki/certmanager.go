package certmanagerpki

import (
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/util/pki"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CertManager interface {
	pki.Manager
}

// certManager implements a PKIManager using cert-manager as the backend
type certManager struct {
	client  client.Client
	cluster *v1alpha1.NifiCluster
}

func New(client client.Client, cluster *v1alpha1.NifiCluster) CertManager {
	return &certManager{client: client, cluster: cluster}
}
