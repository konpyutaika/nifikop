package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

const expectedNodeHostname = "nifi-cluster-1-node.nifi-cluster-headless.nifi.svc.cluster.local"

func generateListenerConfig(listeners []v1.InternalListenerConfig) string {
	return GenerateListenerSpecificConfig(
		&v1.ListenersConfig{InternalListeners: listeners},
		1,
		"nifi",
		"nifi-cluster",
		"cluster.local",
		false,
		"%s-headless",
		*zap.NewNop())
}

func TestGenerateListenerSpecificConfigLoadBalancePort(t *testing.T) {
	config := generateListenerConfig([]v1.InternalListenerConfig{
		{Type: v1.ClusterListenerType, Name: "cluster", ContainerPort: 6007},
		{Type: v1.LoadBalanceListenerType, Name: "load-balance", ContainerPort: 7342},
	})

	// NiFi only reads nifi.cluster.load.balance.port; a non-default port proves the value is honoured
	assert.Contains(t, config, "nifi.cluster.load.balance.port=7342\n")
	assert.NotContains(t, config, "nifi.cluster.node.load.balance.port")
	assert.Contains(t, config, "nifi.cluster.node.protocol.port=6007\n")
}

func TestGenerateListenerSpecificConfigLoadBalancePortUnset(t *testing.T) {
	config := generateListenerConfig([]v1.InternalListenerConfig{
		{Type: v1.ClusterListenerType, Name: "cluster", ContainerPort: 6007},
	})

	// A blank value makes NiFi fall back to its default load balance port
	assert.Contains(t, config, "nifi.cluster.load.balance.port=\n")
}

func TestGenerateListenerSpecificConfigLoadBalanceHost(t *testing.T) {
	config := generateListenerConfig([]v1.InternalListenerConfig{
		{Type: v1.ClusterListenerType, Name: "cluster", ContainerPort: 6007},
		{Type: v1.LoadBalanceListenerType, Name: "load-balance", ContainerPort: 6342},
	})

	// NiFi 2.12 TLS endpoint identification requires both addresses to be DNS names matching the node certificate
	assert.Contains(t, config, "nifi.cluster.node.address="+expectedNodeHostname+"\n")
	assert.Contains(t, config, "nifi.cluster.load.balance.host="+expectedNodeHostname+"\n")
}
