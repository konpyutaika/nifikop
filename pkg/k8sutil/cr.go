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
	"emperror.dev/errors"
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

// AddNewNodeToCr modifies the CR and adds a new node
func AddNewNodeToCr(node v1alpha1.Node, crName, namespace string, client runtimeClient.Client) error {
	cr, err := Cr(crName, namespace, client)
	if err != nil {
		return err
	}
	cr.Spec.Nodes = append(cr.Spec.Nodes, node)

	return updateCr(cr, client)
}

// RemoveNodeFromCr modifies the CR and removes the given node from the cluster
func RemoveNodeFromCr(nodeId, crName, namespace string, client runtimeClient.Client) error {

	cr, err := Cr(crName, namespace, client)
	if err != nil {
		return err
	}

	tmpNodes := cr.Spec.Nodes[:0]
	for _, node := range cr.Spec.Nodes {
		if strconv.Itoa(int(node.Id)) != nodeId {
			tmpNodes = append(tmpNodes, node)
		}
	}
	cr.Spec.Nodes = tmpNodes
	return updateCr(cr, client)
}

// AddPvToSpecificNode adds a new PV to a specific node
func AddPvToSpecificNode(nodeId, crName, namespace string, storageConfig *v1alpha1.StorageConfig, client runtimeClient.Client) error {
	cr, err := Cr(crName, namespace, client)
	if err != nil {
		return err
	}
	tempConfigs := cr.Spec.Nodes[:0]
	for _, node := range cr.Spec.Nodes {
		if strconv.Itoa(int(node.Id)) == nodeId {
			node.NodeConfig.StorageConfigs = append(node.NodeConfig.StorageConfigs, *storageConfig)
		}
		tempConfigs = append(tempConfigs, node)
	}

	cr.Spec.Nodes = tempConfigs
	return updateCr(cr, client)
}

// Cr returns the given cr object
func Cr(name, namespace string, client runtimeClient.Client) (*v1alpha1.NifiCluster, error) {
	cr := &v1alpha1.NifiCluster{}

	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, cr)
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "could not get cr from k8s", "crName", name, "namespace", namespace)
	}
	return cr, nil
}

func updateCr(cr *v1alpha1.NifiCluster, client runtimeClient.Client) error {
	typeMeta := cr.TypeMeta
	err := client.Update(context.TODO(), cr)
	if err != nil {
		return err
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cr.TypeMeta = typeMeta
	return nil
}

// UpdateCrWithRollingUpgrade modifies CR status
func UpdateCrWithRollingUpgrade(errorCount int, cr *v1alpha1.NifiCluster, client runtimeClient.Client) error {

	cr.Status.RollingUpgrade.ErrorCount = errorCount
	return updateCr(cr, client)
}
