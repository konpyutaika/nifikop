package autoscale

import (
	"github.com/konpyutaika/nifikop/api/v1alpha1"
)

type HorizontalDownscaleStrategy interface {
	Type() v1alpha1.ClusterScalingStrategy

	// ScaleDown takes the current set of nodes, removes numNodesToRemove per the strategy, and
	// returns the set of nodes that should be removed from the set of currentNodes
	ScaleDown(currentNodes []v1alpha1.Node, numNodesToRemove int32) (nodesToRemove []v1alpha1.Node, err error)
}

type HorizontalUpscaleStrategy interface {
	Type() v1alpha1.ClusterScalingStrategy

	// ScaleUp takes the current set of nodes, removes numNodesToRemove per the strategy, and
	// returns the set of nodes that should be removed from the set of currentNodes
	ScaleUp(currentNodes []v1alpha1.Node, numNodesToAdd int32) (newNodes []v1alpha1.Node, err error)
}

// LIFO downscale strategy
// Nodes are added by monotonically increasing nodeId, so LIFO is simply a strategy where the highest ID nodes are removed first.
type LIFOHorizontalDownscaleStrategy struct{}

// ScaleDown takes the current set of nodes, removes numNodesToRemove per the strategy, and
// returns the set of nodes that should be removed from the set of currentNodes
func (lifo *LIFOHorizontalDownscaleStrategy) ScaleDown(currentNodes []v1alpha1.Node, numNodesToRemove int32) (nodesToRemove []v1alpha1.Node, err error) {
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
// A simple cluster upscale operation is simply appending a node to the existing node set
type SimpleHorizontalUpscaleStrategy struct {
	NifiNodeGroupAutoscaler *v1alpha1.NifiNodeGroupAutoscaler
	MaxNodeId               int32
}

func (simple *SimpleHorizontalUpscaleStrategy) Type() v1alpha1.ClusterScalingStrategy {
	return v1alpha1.SimpleClusterUpscaleStrategy
}

// ScaleUp takes the current set of nodes, removes numNodesToRemove per the strategy, and
// returns the set of nodes that should be removed from the set of currentNodes
func (simple *SimpleHorizontalUpscaleStrategy) ScaleUp(currentNodes []v1alpha1.Node, numNodesToAdd int32) (newNodes []v1alpha1.Node, err error) {
	if numNodesToAdd == 0 {
		return newNodes, nil
	}
	autoscalingNodeLabels, err := simple.NifiNodeGroupAutoscaler.Spec.NifiNodeGroupSelectorAsMap()
	if err != nil {
		return nil, err
	}

	for i := int32(0); i < numNodesToAdd; i++ {
		newNodes = append(newNodes, v1alpha1.Node{
			Id:              simple.MaxNodeId + i,
			NodeConfigGroup: simple.NifiNodeGroupAutoscaler.Spec.NodeConfigGroupId,
			ReadOnlyConfig:  simple.NifiNodeGroupAutoscaler.Spec.ReadOnlyConfig,
			Labels:          autoscalingNodeLabels,
		})
	}
	return
}
