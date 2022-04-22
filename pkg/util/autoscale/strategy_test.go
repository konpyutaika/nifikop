package autoscale

import (
	"testing"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
)

func TestLIFORemoveAllNodes(t *testing.T) {
	lifo := LIFOHorizontalDownscaleStrategy{}

	nodes := []v1alpha1.Node{
		{Id: 2, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
		{Id: 3, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
		{Id: 4, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
	}
	// this also verifies if you remove more nodes than the current set, that it just returns an empty list.
	nodesToRemove, err := lifo.ScaleDown(nodes, 100)

	if err != nil {
		t.Error("Should not have encountered an error")
	}

	if len(nodesToRemove) != 0 {
		t.Error("nodesToRemove should have been empty")
	}
}

func TestLIFORemoveSomeNodes(t *testing.T) {
	lifo := LIFOHorizontalDownscaleStrategy{}

	nodes := []v1alpha1.Node{}
	expectedNode := v1alpha1.Node{
		Id: 1,
	}
	nodes = append(nodes,
		expectedNode,
		v1alpha1.Node{
			Id: 2,
		},
		v1alpha1.Node{
			Id: 3,
		},
	)
	nodesToRemove, err := lifo.ScaleDown(nodes, 2)

	if err != nil {
		t.Error("Should not have encountered an error")
	}
	if len(nodesToRemove) != 2 {
		t.Errorf("Did not remove correct number of nodes: %v+", nodesToRemove)
	}
	if nodesToRemove[0].Id != 2 && nodesToRemove[0].Id != 3 {
		t.Errorf("Incorrect results. Nodes: %v+", nodesToRemove)
	}
}

func TestLIFORemoveOneNode(t *testing.T) {
	lifo := LIFOHorizontalDownscaleStrategy{}

	nodes := []v1alpha1.Node{}
	nodes = append(nodes,
		v1alpha1.Node{
			Id: 1,
		},
		v1alpha1.Node{
			Id: 2,
		},
		v1alpha1.Node{
			Id: 3,
		},
	)
	nodesToRemove, err := lifo.ScaleDown(nodes, 1)

	if err != nil {
		t.Error("Should not have encountered an error")
	}
	if len(nodesToRemove) != 1 {
		t.Errorf("Did not remove correct number of nodes: %v+", nodesToRemove)
	}

	if nodesToRemove[0].Id != 3 {
		t.Errorf("Incorrect results. Nodes: %v+", nodesToRemove)
	}
}

func TestLIFORemoveNoNodes(t *testing.T) {
	lifo := LIFOHorizontalDownscaleStrategy{}

	nodes := make([]v1alpha1.Node, 1)
	nodes = append(nodes, v1alpha1.Node{
		Id: 1,
	})
	nodesToRemove, err := lifo.ScaleDown(nodes, 0)

	if err != nil {
		t.Error("Should not have encountered an error")
	}

	if len(nodesToRemove) != 0 {
		t.Error("nodesToRemove should have been empty")
	}
}

func TestSimpleAddOneNode(t *testing.T) {
	simple := SimpleHorizontalUpscaleStrategy{
		NifiNodeGroupAutoscaler: &v1alpha1.NifiNodeGroupAutoscaler{},
		MaxNodeId:               0,
	}

	nodes := make([]v1alpha1.Node, 1)
	nodes = append(nodes, v1alpha1.Node{
		Id: 1,
	})
	nodesToAdd, err := simple.ScaleUp(nodes, 2)

	if err != nil {
		t.Error("Should not have encountered an error")
	}

	if len(nodesToAdd) != 2 {
		t.Error("nodesToAdd should have been 2")
	}
	if nodesToAdd[0].Id != 0 || nodesToAdd[1].Id != 1 {
		t.Errorf("nodesToAdd Ids are not correct: %v+", nodesToAdd)
	}
}

func TestSimpleAddNoNodes(t *testing.T) {
	simple := SimpleHorizontalUpscaleStrategy{
		NifiNodeGroupAutoscaler: &v1alpha1.NifiNodeGroupAutoscaler{},
		MaxNodeId:               0,
	}

	nodes := make([]v1alpha1.Node, 1)
	nodes = append(nodes, v1alpha1.Node{
		Id: 1,
	})
	nodesToAdd, err := simple.ScaleUp(nodes, 0)

	if err != nil {
		t.Error("Should not have encountered an error")
	}

	if len(nodesToAdd) != 0 {
		t.Errorf("nodesToAdd should have been empty: %v+", nodesToAdd)
	}
}
