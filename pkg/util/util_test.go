package util

import (
	"reflect"
	"testing"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
)

func TestSubtractNodes(t *testing.T) {
	sourceList := []v1alpha1.Node{
		{
			Id: 1,
		},
		{
			Id: 2,
		},
		{
			Id: 3,
		},
	}

	nodesToSubtract := []v1alpha1.Node{
		{
			Id: 3,
		},
	}

	// subtract 1 node
	if results := SubtractNodes(sourceList, nodesToSubtract); len(results) != 2 {
		t.Error("There should be two nodes remaining")
	}

	// subtract empty list
	if results := SubtractNodes(sourceList, []v1alpha1.Node{}); len(results) != 3 {
		t.Error("there should be 3 results")
	}

	// subtract all nodes
	if results := SubtractNodes(sourceList, sourceList); len(results) != 0 {
		t.Error("There should be two nodes remaining")
	}
}

func TestComputeNewNodeIds(t *testing.T) {
	nodeList := []v1alpha1.Node{
		{
			Id: 1,
		},
		{
			Id: 2,
		},
		{
			Id: 5,
		},
	}

	// add more nodes than size of input node list
	newNodeIds := ComputeNewNodeIds(nodeList, 5)
	if len(newNodeIds) != 5 {
		t.Errorf("There should be 5 new nodes. %v+", newNodeIds)
	}
	if !reflect.DeepEqual(newNodeIds, []int32{0, 3, 4, 6, 7}) {
		t.Errorf("lists are not equal. %v+", newNodeIds)
	}

	// add less nodes than size of input node list
	newNodeIds = ComputeNewNodeIds(nodeList, 2)

	if len(newNodeIds) != 2 {
		t.Errorf("There should be 2 new nodes. %v+", newNodeIds)
	}
	if !reflect.DeepEqual(newNodeIds, []int32{0, 3}) {
		t.Errorf("lists are not equal. %v+", newNodeIds)
	}

	// add same number of nodes than size of input node list
	newNodeIds = ComputeNewNodeIds(nodeList, 3)
	if len(newNodeIds) != 3 {
		t.Errorf("There should be 3 new nodes. %v+", newNodeIds)
	}
	if !reflect.DeepEqual(newNodeIds, []int32{0, 3, 4}) {
		t.Errorf("lists are not equal. %v+", newNodeIds)
	}

	// add zero new nodes
	newNodeIds = ComputeNewNodeIds(nodeList, 0)
	if len(newNodeIds) != 0 {
		t.Errorf("There should be 0 new nodes. %v+", newNodeIds)
	}
}
