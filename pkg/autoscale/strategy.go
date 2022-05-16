package autoscale

import (
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type HorizontalDownscaleStrategy interface {
	Type() v1alpha1.ClusterScalingStrategy

	// returns the set of "numNodesToRemove" nodes that should be removed from the cluster
	ScaleDown(numNodesToRemove int32) (nodesToRemove []v1alpha1.Node, err error)
}

type HorizontalUpscaleStrategy interface {
	Type() v1alpha1.ClusterScalingStrategy

	// returns the set of "numNodesToAdd" nodes that should be added to the cluster
	ScaleUp(numNodesToAdd int32) (newNodes []v1alpha1.Node, err error)
}

// LIFO downscale strategy
// Nodes are added by monotonically increasing nodeId, so LIFO is simply a strategy where the highest ID nodes are removed first.
type LIFOHorizontalDownscaleStrategy struct {
	NifiCluster             *v1alpha1.NifiCluster
	NifiNodeGroupAutoscaler *v1alpha1.NifiNodeGroupAutoscaler
}

// returns the set of "numNodesToRemove" nodes that should be removed from the cluster
func (lifo *LIFOHorizontalDownscaleStrategy) ScaleDown(numNodesToRemove int32) (nodesToRemove []v1alpha1.Node, err error) {
	// we use the creation time-ordered nodes here so that we can remove the last nodes added to the cluster
	currentNodes, err := getManagedNodes(lifo.NifiNodeGroupAutoscaler, lifo.NifiCluster.GetCreationTimeOrderedNodes())
	if err != nil {
		return nil, err
	}
	numberOfCurrentNodes := int32(len(currentNodes))
	if numNodesToRemove >= numberOfCurrentNodes || numNodesToRemove == 0 {
		return []v1alpha1.Node{}, nil
	}

	nodesToRemove = []v1alpha1.Node{}
	nodesToRemove = append(nodesToRemove, currentNodes[numberOfCurrentNodes-numNodesToRemove:]...)

	// the last <numNodesToRemove> are the nodes which need to be removed
	return nodesToRemove, nil
}

func (lifo *LIFOHorizontalDownscaleStrategy) Type() v1alpha1.ClusterScalingStrategy {
	return v1alpha1.LIFOClusterDownscaleStrategy
}

// Simple upscale strategy
// A simple cluster upscale operation is simply adding a node to the existing node set
type SimpleHorizontalUpscaleStrategy struct {
	NifiCluster             *v1alpha1.NifiCluster
	NifiNodeGroupAutoscaler *v1alpha1.NifiNodeGroupAutoscaler
}

func (simple *SimpleHorizontalUpscaleStrategy) Type() v1alpha1.ClusterScalingStrategy {
	return v1alpha1.SimpleClusterUpscaleStrategy
}

// returns the set of "numNodesToAdd" nodes that should be added to the cluster
func (simple *SimpleHorizontalUpscaleStrategy) ScaleUp(numNodesToAdd int32) (newNodes []v1alpha1.Node, err error) {
	if numNodesToAdd == 0 {
		return newNodes, nil
	}
	autoscalingNodeLabels, err := simple.NifiNodeGroupAutoscaler.Spec.NifiNodeGroupSelectorAsMap()
	if err != nil {
		return nil, err
	}

	// when computing new node IDs, we consider the entire cluster so that we don't inadvertntly re-use existing IDs
	newNodeIds := util.ComputeNewNodeIds(simple.NifiCluster.Spec.Nodes, numNodesToAdd)

	for _, id := range newNodeIds {
		newNodes = append(newNodes, v1alpha1.Node{
			Id:              id,
			NodeConfigGroup: simple.NifiNodeGroupAutoscaler.Spec.NodeConfigGroupId,
			ReadOnlyConfig:  simple.NifiNodeGroupAutoscaler.Spec.ReadOnlyConfig,
			Labels:          autoscalingNodeLabels,
			NodeConfig:      simple.NifiNodeGroupAutoscaler.Spec.NodeConfig,
		})
	}
	return
}

// filter the set of provided nodes by the autoscaler's node selector
func getManagedNodes(autoscaler *v1alpha1.NifiNodeGroupAutoscaler, nodes []v1alpha1.Node) (managedNodes []v1alpha1.Node, err error) {
	selector, err := metav1.LabelSelectorAsSelector(autoscaler.Spec.NodeLabelsSelector)
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if selector.Matches(labels.Set(node.Labels)) {
			managedNodes = append(managedNodes, node)
		}
	}
	return managedNodes, nil
}
