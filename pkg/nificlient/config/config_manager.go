package config

import (
	"context"

	"go.uber.org/zap"
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

// BuildConfig returns an empty configuration: a reconciler under test is expected to stub the
// calls that would use it to reach NiFi.
func (n *mockClientConfig) BuildConfig() (*clientconfig.NifiConfig, error) {
	return &clientconfig.NifiConfig{NodesContext: map[int32]context.Context{}}, nil
}

// BuildConnect returns a cluster that is always ready, so reconcilers can be exercised with a
// fake Kubernetes client and no NiFi behind it.
func (n *mockClientConfig) BuildConnect() (cluster clientconfig.ClusterConnect, err error) {
	return &mockClusterConnect{name: n.clusterRef.Name}, nil
}

func (n *mockClientConfig) IsExternal() bool {
	return true
}

type mockClusterConnect struct {
	name string
}

func (c *mockClusterConnect) IsInternal() bool            { return false }
func (c *mockClusterConnect) IsExternal() bool            { return true }
func (c *mockClusterConnect) ClusterLabelString() string  { return c.name }
func (c *mockClusterConnect) IsReady(log zap.Logger) bool { return true }
func (c *mockClusterConnect) Id() string                  { return c.name }
