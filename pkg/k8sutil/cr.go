package k8sutil

import (
	"context"
	"strconv"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konpyutaika/nifikop/api/v1"
)

// AddNewNodeToCr modifies the CR and adds a new node.
func AddNewNodeToCr(node v1.Node, crName, namespace string, client runtimeClient.Client) error {
	cr, err := Cr(crName, namespace, client)
	if err != nil {
		return err
	}
	cr.Spec.Nodes = append(cr.Spec.Nodes, node)

	return updateCr(cr, client)
}

// RemoveNodeFromCr modifies the CR and removes the given node from the cluster.
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

// AddPvToSpecificNode adds a new PV to a specific node.
func AddPvToSpecificNode(nodeId, crName, namespace string, storageConfig *v1.StorageConfig, client runtimeClient.Client) error {
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

// Cr returns the given cr object.
func Cr(name, namespace string, client runtimeClient.Client) (*v1.NifiCluster, error) {
	cr := &v1.NifiCluster{}

	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, cr)
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "could not get cr from k8s", "crName", name, "namespace", namespace)
	}
	return cr, nil
}

func updateCr(cr *v1.NifiCluster, client runtimeClient.Client) error {
	typeMeta := cr.TypeMeta
	err := client.Update(context.TODO(), cr)
	if err != nil {
		return err
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cr.TypeMeta = typeMeta
	return nil
}

// UpdateCrWithRollingUpgrade modifies CR status.
func UpdateCrWithRollingUpgrade(errorCount int, cr *v1.NifiCluster, client runtimeClient.Client) error {
	cr.Status.RollingUpgrade.ErrorCount = errorCount
	return updateCr(cr, client)
}
