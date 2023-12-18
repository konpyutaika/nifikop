package config

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config/basic"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config/tls"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var MockClientConfig = v1.ClientConfigType("mock")

func GetClientConfigManager(client client.Client, clusterRef v1.ClusterReference) clientconfig.Manager {
	cluster, _ := k8sutil.LookupNifiCluster(client, clusterRef.Name, clusterRef.Namespace)
	switch cluster.GetClientType() {
	case v1.ClientConfigTLS:
		return tls.New(client, clusterRef)
	case v1.ClientConfigBasic:
		return basic.New(client, clusterRef)
	case MockClientConfig:
		return NewMockClientConfig(client, clusterRef)
	default:
		return tls.New(client, clusterRef)
	}
}

// Mock types and functions.
type mockClientConfig struct {
	clientconfig.Manager
	client     client.Client
	clusterRef v1.ClusterReference
}

func NewMockClientConfig(client client.Client, clusterRef v1.ClusterReference) clientconfig.Manager {
	return &mockClientConfig{client: client, clusterRef: clusterRef}
}

func (n *mockClientConfig) BuildConfig() (*clientconfig.NifiConfig, error) {
	return nil, nil
}

func (n *mockClientConfig) BuildConnect() (cluster clientconfig.ClusterConnect, err error) {
	return
}

func (n *mockClientConfig) IsExternal() bool {
	return true
}
