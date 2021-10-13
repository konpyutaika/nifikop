package basic

import (
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/util/clientconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Basic interface {
	clientconfig.Manager
}

type basic struct {
	client     client.Client
	clusterRef v1alpha1.ClusterReference
}

func New(client client.Client, clusterRef v1alpha1.ClusterReference) Basic {
	return &basic{clusterRef: clusterRef, client: client}
}
