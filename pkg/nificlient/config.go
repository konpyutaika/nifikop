// Copyright 2020 Orange SA
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
// limitations under the License.package apis

package nificlient

import (
	"crypto/tls"
	"fmt"

	"github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/pki"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	"github.com/Orange-OpenSource/nifikop/pkg/util/nifi"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nifiDefaultTimeout = int64(5)
)

// NifiConfig are the options to creating a new ClusterAdmin client
type NifiConfig struct {
	nodeURITemplate string
	NodesURI        map[int32]string
	NifiURI         string
	UseSSL          bool
	TLSConfig       *tls.Config

	OperationTimeout int64
}

// ClusterConfig creates connection options from a NifiCluster CR
func ClusterConfig(client client.Client, cluster *v1alpha1.NifiCluster) (*NifiConfig, error) {

	conf := &NifiConfig{}

	conf.nodeURITemplate = generateNodesURITemplate(cluster)
	conf.NodesURI = generateNodesAddress(cluster)
	conf.NifiURI = nifi.GenerateNiFiAddressFromCluster(cluster)
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
	return cluster.Spec.ClusterSecure
}

func generateNodesAddress(cluster *v1alpha1.NifiCluster) map[int32]string {
	addresses := make(map[int32]string)

	for nId, state := range cluster.Status.NodesState {
		if !(state.GracefulActionState.State.IsRunningState() || state.GracefulActionState.State.IsRequiredState()) && state.GracefulActionState.ActionStep != v1alpha1.RemoveStatus {
			addresses[util.ConvertStringToInt32(nId)] = nifi.GenerateNodeAddressFromCluster(util.ConvertStringToInt32(nId), cluster)
		}
	}
	return addresses
}

func generateNodesURITemplate(cluster *v1alpha1.NifiCluster) string {
	nodeNameTemplate :=
		fmt.Sprintf(nifi.PrefixNodeNameTemplate, cluster.Name) +
			nifi.RootNodeNameTemplate +
			nifi.SuffixNodeNameTemplate

	return nodeNameTemplate + fmt.Sprintf(".%s",
		nifi.GenerateNiFiAddressFromCluster(cluster),
	)
}
