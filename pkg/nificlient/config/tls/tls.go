package tls

import (
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/util/clientconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Tls interface {
	clientconfig.Manager
}

type tls struct {
	client     client.Client
	clusterRef v1alpha1.ClusterReference
}

func New(client client.Client, clusterRef v1alpha1.ClusterReference) Tls {
	return &tls{clusterRef: clusterRef, client: client}
}
