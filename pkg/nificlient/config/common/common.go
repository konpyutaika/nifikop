package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/pki"
	"github.com/konpyutaika/nifikop/pkg/pki/certmanagerpki"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	"github.com/konpyutaika/nifikop/pkg/util/nifi"
)

func TlsConfig(client client.Client, cluster *v1.NifiCluster) (config *tls.Config, err error) {
	if cluster.IsExternal() {
		return certmanagerpki.GetControllerTLSConfigFromSecret(client, cluster.Spec.SecretRef)
	}

	return pki.GetPKIManager(client, cluster).GetControllerTLSConfig()
}

func ClusterConfig(cluster *v1.NifiCluster) *clientconfig.NifiConfig {
	if cluster.IsExternal() {
		return externalClusterConfig(cluster)
	}

	return internalClusterConfig(cluster)
}

func externalClusterConfig(cluster *v1.NifiCluster) *clientconfig.NifiConfig {
	conf := &clientconfig.NifiConfig{}
	ref := cluster.Spec
	nodesURI := generateNodesAddressFromTemplate(ref.Nodes, ref.NodeURITemplate)

	conf.RootProcessGroupId = ref.RootProcessGroupId
	conf.NodeURITemplate = ref.NodeURITemplate
	conf.NodesURI = nodesURI
	conf.NifiURI = ref.NifiURI
	conf.OperationTimeout = clientconfig.NifiDefaultTimeout
	conf.NodesContext = make(map[int32]context.Context)
	conf.ProxyUrl = ref.ProxyUrl
	conf.UseSSL = true

	return conf
}

func internalClusterConfig(cluster *v1.NifiCluster) *clientconfig.NifiConfig {
	conf := &clientconfig.NifiConfig{}
	conf.RootProcessGroupId = cluster.Status.RootProcessGroupId
	conf.NodeURITemplate = generateNodesURITemplate(cluster)
	conf.NodesURI = generateNodesAddress(cluster)
	conf.NifiURI = nifi.GenerateRequestNiFiAllNodeAddressFromCluster(cluster)
	conf.OperationTimeout = clientconfig.NifiDefaultTimeout
	conf.NodesContext = make(map[int32]context.Context)
	conf.UseSSL = cluster.Spec.ListenersConfig.SSLSecrets != nil && UseSSL(cluster)
	return conf
}

func generateNodesAddress(cluster *v1.NifiCluster) map[int32]clientconfig.NodeUri {
	addresses := make(map[int32]clientconfig.NodeUri)

	for nId, state := range cluster.Status.NodesState {
		if !(state.GracefulActionState.State.IsRunningState() || state.GracefulActionState.State.IsRequiredState()) && state.GracefulActionState.ActionStep != v1.RemoveStatus {
			addresses[util.ConvertStringToInt32(nId)] = clientconfig.NodeUri{
				HostListener: nifi.GenerateHostListenerNodeAddressFromCluster(util.ConvertStringToInt32(nId), cluster),
				RequestHost:  nifi.GenerateRequestNiFiNodeAddressFromCluster(util.ConvertStringToInt32(nId), cluster),
			}
		}
	}
	return addresses
}

func generateNodesURITemplate(cluster *v1.NifiCluster) string {
	nodeNameTemplate :=
		fmt.Sprintf(nifi.PrefixNodeNameTemplate, cluster.Name) +
			nifi.RootNodeNameTemplate +
			nifi.SuffixNodeNameTemplate

	return nodeNameTemplate + fmt.Sprintf(".%s",
		strings.SplitAfterN(nifi.GenerateRequestNiFiNodeAddressFromCluster(0, cluster), ".", 2)[1],
	)
}

func generateNodesAddressFromTemplate(nodes []v1.Node, template string) map[int32]clientconfig.NodeUri {
	addresses := make(map[int32]clientconfig.NodeUri)

	for _, node := range nodes {
		addresses[node.Id] = clientconfig.NodeUri{
			HostListener: fmt.Sprintf(template, node.Id),
			RequestHost:  fmt.Sprintf(template, node.Id),
		}
	}
	return addresses
}

func UseSSL(cluster *v1.NifiCluster) bool {
	return cluster.Spec.ListenersConfig.SSLSecrets != nil
}
