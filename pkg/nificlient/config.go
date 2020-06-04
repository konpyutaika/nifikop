// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nificlient

import (
	"crypto/tls"
	"fmt"

	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/apis/nifi/v1alpha1"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/pki"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/resources/templates"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/util"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/util/nifi"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nifiDefaultTimeout = int64(5)
	serviceHostnameTemplate = "%s.%s.svc.cluster.local:%d"
)

// NifiConfig are the options to creating a new ClusterAdmin client
type NifiConfig struct {
	nodeURITemplate string
	NodesURI 		map[int32]string
	NifiURI 		string
	UseSSL    		bool
	TLSConfig 		*tls.Config

	OperationTimeout int64
}

// ClusterConfig creates connection options from a NifiCluster CR
func ClusterConfig(client client.Client, cluster *v1alpha1.NifiCluster) (*NifiConfig, error) {
	conf := &NifiConfig{}
	conf.nodeURITemplate = generateNodesURITemplate(cluster)
	conf.NodesURI = generateNodesAddress(cluster)
	conf.NifiURI = GenerateNifiAddress(cluster)
	conf.OperationTimeout = nifiDefaultTimeout
	if cluster.Spec.ListenersConfig.SSLSecrets != nil && useSSL(cluster) {
		tlsConfig, err := pki.GetPKIManager(client, cluster).GetControllerTLSConfig()
		if err != nil {
			return conf, err
		}
		conf.UseSSL = true
		conf.TLSConfig = tlsConfig
	}
	return conf, nil
}

func useSSL(cluster *v1alpha1.NifiCluster) bool {
	//return cluster.Spec.ListenersConfig.InternalListeners[determineInternalListenerForInnerCom(cluster.Spec.ListenersConfig.InternalListeners)].Type != "plaintext"
	return cluster.Spec.ClusterSecure
}

func determineInternalListenerForComm(internalListeners []v1alpha1.InternalListenerConfig) int {
	var httpsServerPortId int
	var httpServerPortId int
	for id, iListener := range internalListeners {
		if iListener.Type == v1alpha1.HttpsListenerType {
			httpsServerPortId = id
		} else if iListener.Type == v1alpha1.HttpListenerType {
			httpServerPortId = id
		}
	}
	if &httpsServerPortId != nil {
		return httpsServerPortId
	}
	return httpServerPortId
}

func generateNodesAddress(cluster *v1alpha1.NifiCluster) map[int32]string {
	addresses := make(map[int32]string)

	for nId, state := range cluster.Status.NodesState {
		if !(state.GracefulActionState.State.IsRunningState() || state.GracefulActionState.State.IsRequiredState()) && state.GracefulActionState.ActionStep != v1alpha1.RemoveStatus   {
			addresses[util.ConvertStringToInt32(nId)] = GenerateNodeAddress(cluster, util.ConvertStringToInt32(nId))
		}
	}
	return addresses
}

func GenerateNodeAddress(cluster *v1alpha1.NifiCluster, nodeId int32) string {
		return fmt.Sprintf(generateNodesURITemplate(cluster), nodeId)
}

func generateNodesURITemplate(cluster *v1alpha1.NifiCluster) string {

	nodeNameTemplate := fmt.Sprintf(templates.PrefixNodeNameTemplate, cluster.Name) + templates.RootNodeNameTemplate + templates.SuffixNodeNameTemplate

	return nodeNameTemplate + fmt.Sprintf(".%s",
		GenerateNifiAddress(cluster),
	)
}

func GenerateNifiAddress(cluster *v1alpha1.NifiCluster) string {

	if cluster.Spec.HeadlessServiceEnabled {
		return fmt.Sprintf(serviceHostnameTemplate,
			fmt.Sprintf(nifi.HeadlessServiceTemplate, cluster.Name),
			cluster.Namespace,
			cluster.Spec.ListenersConfig.InternalListeners[determineInternalListenerForComm(cluster.Spec.ListenersConfig.InternalListeners)].ContainerPort,
			)
	}

	return fmt.Sprintf(serviceHostnameTemplate,
		fmt.Sprintf(nifi.AllNodeServiceTemplate, cluster.Name),
		cluster.Namespace,
		cluster.Spec.ListenersConfig.InternalListeners[determineInternalListenerForComm(cluster.Spec.ListenersConfig.InternalListeners)].ContainerPort,
		)
}
