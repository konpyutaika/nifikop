package autoscale

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/util"
)

type HorizontalDownscaleStrategy interface {
	Type() v1alpha1.ClusterScalingStrategy

	// returns the set of "numNodesToRemove" nodes that should be removed from the cluster
	ScaleDown(numNodesToRemove int32) (nodesToRemove []v1.Node, err error)
}

type HorizontalUpscaleStrategy interface {
	Type() v1alpha1.ClusterScalingStrategy

	// returns the set of "numNodesToAdd" nodes that should be added to the cluster
	ScaleUp(numNodesToAdd int32) (newNodes []v1.Node, err error)
}

// LIFO downscale strategy
// Nodes are added by monotonically increasing nodeId, so LIFO is simply a strategy where the highest ID nodes are removed first.
type LIFOHorizontalDownscaleStrategy struct {
	NifiCluster             *v1.NifiCluster
	NifiNodeGroupAutoscaler *v1alpha1.NifiNodeGroupAutoscaler
}

// returns the set of "numNodesToRemove" nodes that should be removed from the cluster.
func (lifo *LIFOHorizontalDownscaleStrategy) ScaleDown(numNodesToRemove int32) (nodesToRemove []v1.Node, err error) {
	// we use the creation time-ordered nodes here so that we can remove the last nodes added to the cluster
	currentNodes, err := getManagedNodes(lifo.NifiNodeGroupAutoscaler, lifo.NifiCluster.GetCreationTimeOrderedNodes())
	if err != nil {
		return nil, err
	}
	numberOfCurrentNodes := int32(len(currentNodes))
	if numNodesToRemove > numberOfCurrentNodes || numNodesToRemove == 0 {
		return []v1.Node{}, nil
	}

	nodesToRemove = []v1.Node{}
	nodesToRemove = append(nodesToRemove, currentNodes[numberOfCurrentNodes-numNodesToRemove:]...)

	// the last <numNodesToRemove> are the nodes which need to be removed
	return nodesToRemove, nil
}

func (lifo *LIFOHorizontalDownscaleStrategy) Type() v1alpha1.ClusterScalingStrategy {
	return v1alpha1.LIFOClusterDownscaleStrategy
}

// Simple upscale strategy
// A simple cluster upscale operation is simply adding a node to the existing node set.
type SimpleHorizontalUpscaleStrategy struct {
	NifiCluster             *v1.NifiCluster
	NifiNodeGroupAutoscaler *v1alpha1.NifiNodeGroupAutoscaler
}

func (simple *SimpleHorizontalUpscaleStrategy) Type() v1alpha1.ClusterScalingStrategy {
	return v1alpha1.SimpleClusterUpscaleStrategy
}

// returns the set of "numNodesToAdd" nodes that should be added to the cluster.
func (simple *SimpleHorizontalUpscaleStrategy) ScaleUp(numNodesToAdd int32) (newNodes []v1.Node, err error) {
	if numNodesToAdd == 0 {
		return newNodes, nil
	}
	autoscalingNodeLabels, err := simple.NifiNodeGroupAutoscaler.Spec.NifiNodeGroupSelectorAsMap()
	if err != nil {
		return nil, err
	}

	// when computing new node IDs, we consider the entire cluster so that we don't inadvertntly re-use existing IDs
	newNodeIds := ComputeNewNodeIds(simple.NifiCluster.Spec.Nodes, numNodesToAdd)

	for _, id := range newNodeIds {
		newNodes = append(newNodes, v1.Node{
			Id:              id,
			NodeConfigGroup: simple.NifiNodeGroupAutoscaler.Spec.NodeConfigGroupId,
			ReadOnlyConfig:  simple.NifiNodeGroupAutoscaler.Spec.ReadOnlyConfig,
			Labels:          autoscalingNodeLabels,
			NodeConfig:      simple.NifiNodeGroupAutoscaler.Spec.NodeConfig,
		})
	}
	return
}

// filter the set of provided nodes by the autoscaler's node selector.
func getManagedNodes(autoscaler *v1alpha1.NifiNodeGroupAutoscaler, nodes []v1.Node) (managedNodes []v1.Node, err error) {
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

// New nodes are assigned an Id in the following manner:
//
// - Assigned node Ids will always be a non-negative integer starting with zero
//
// - extract and sort the node Ids in the provided node list
//
// - iterate through the node Id list starting with zero. For any unassigned node Id, assign it
//
// - return the list of assigned node Ids.
func ComputeNewNodeIds(nodes []v1.Node, numNewNodes int32) []int32 {
	nodeIdList := util.NodesToIdList(nodes)
	sort.Slice(nodeIdList, func(i, j int) bool {
		return nodeIdList[i] < nodeIdList[j]
	})

	newNodeIds := []int32{}
	index := int32(0)

	// assign IDs in any gaps in the existing node list, starting with zero
	for i := int32(0); len(nodeIdList) > 0 && i < nodeIdList[len(nodeIdList)-1] && int32(len(newNodeIds)) < numNewNodes; i++ {
		if nodeIdList[index] == i {
			index++
		} else {
			newNodeIds = append(newNodeIds, i)
		}
	}

	// retrieve the max id to start from it
	max, err := util.MaxSlice32(nodeIdList)
	if err != nil {
		max = -1
	}
	// add any remaining nodes needed
	remainder := numNewNodes - int32(len(newNodeIds))
	for j := int32(max + 1); j <= remainder+max; j++ {
		newNodeIds = append(newNodeIds, j)
	}
	return newNodeIds
}
