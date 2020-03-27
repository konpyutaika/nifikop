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

	"github.com/erdrix/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/erdrix/nifikop/pkg/pki"
	"github.com/erdrix/nifikop/pkg/util/nifi"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const nifiDefaultTimeout = int64(5)

// NifiConfig are the options to creating a new ClusterAdmin client
type NifiConfig struct {
	NifiURI string
	UseSSL    bool
	TLSConfig *tls.Config

	OperationTimeout int64
}

// ClusterConfig creates connection options from a NifiCluster CR
func ClusterConfig(client client.Client, cluster *v1alpha1.NifiCluster) (*NifiConfig, error) {
	conf := &NifiConfig{}
	conf.NifiURI = generateNifiAddress(cluster)
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

func determineInternalListenerForInnerCom(internalListeners []v1alpha1.InternalListenerConfig) int {
	for id, val := range internalListeners {
		if val.Type == v1alpha1.ClusterListenerType {
			return id
		}
	}
	return 0
}

func generateNifiAddress(cluster *v1alpha1.NifiCluster) string {
	if cluster.Spec.HeadlessServiceEnabled {
		return fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			fmt.Sprintf(nifi.HeadlessServiceTemplate, cluster.Name),
			cluster.Namespace,
			cluster.Spec.ListenersConfig.InternalListeners[determineInternalListenerForInnerCom(cluster.Spec.ListenersConfig.InternalListeners)].ContainerPort,
		)
	}
	return fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		fmt.Sprintf(nifi.AllNodeServiceTemplate, cluster.Name),
		cluster.Namespace,
		cluster.Spec.ListenersConfig.InternalListeners[determineInternalListenerForInnerCom(cluster.Spec.ListenersConfig.InternalListeners)].ContainerPort,
	)
}
