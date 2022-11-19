package basic

import (
	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
