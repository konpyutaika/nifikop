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

package k8sutil

import (
	"context"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/apis/nifi/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// LookupNifiCluster returns the running cluster instance based on its name and namespace
func LookupNifiCluster(client runtimeClient.Client, clusterName, clusterNamespace string) (cluster *v1alpha1.NifiCluster, err error) {
	cluster = &v1alpha1.NifiCluster{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: clusterName, Namespace: clusterNamespace}, cluster)
	return
}
