package basic

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

type Basic interface {
	clientconfig.Manager
}

type basic struct {
	client     client.Client
	clusterRef v1.ClusterReference
}

func New(client client.Client, clusterRef v1.ClusterReference) Basic {
	return &basic{clusterRef: clusterRef, client: client}
}
