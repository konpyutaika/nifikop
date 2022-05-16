package autoscale

import (
	"testing"
	"time"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var lifo = LIFOHorizontalDownscaleStrategy{
	NifiNodeGroupAutoscaler: &v1alpha1.NifiNodeGroupAutoscaler{
		Spec: v1alpha1.NifiNodeGroupAutoscalerSpec{
			NodeLabelsSelector: &v1.LabelSelector{
				MatchLabels: map[string]string{"scale_me": "true"},
			},
		},
	},
	NifiCluster: &v1alpha1.NifiCluster{
		Spec: v1alpha1.NifiClusterSpec{
			Nodes: []v1alpha1.Node{
				{Id: 2, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 3, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 4, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 5, NodeConfigGroup: "other-group", Labels: map[string]string{"other_group": "true"}},
			},
		},
		Status: v1alpha1.NifiClusterStatus{
			NodesState: map[string]v1alpha1.NodeState{
				"2": v1alpha1.NodeState{
					CreationTime: v1.NewTime(time.Now().UTC().Add(time.Duration(5) * time.Hour)),
				},
				"3": v1alpha1.NodeState{
					CreationTime: v1.NewTime(time.Now().UTC().Add(time.Duration(10) * time.Hour)),
				},
				"4": v1alpha1.NodeState{
					CreationTime: v1.NewTime(time.Now().UTC().Add(time.Duration(15) * time.Hour)),
				},
				"5": v1alpha1.NodeState{
					CreationTime: v1.NewTime(time.Now().UTC().Add(time.Duration(20) * time.Hour)),
				},
			},
		},
	},
}

var simple = SimpleHorizontalUpscaleStrategy{
	NifiNodeGroupAutoscaler: &v1alpha1.NifiNodeGroupAutoscaler{
		Spec: v1alpha1.NifiNodeGroupAutoscalerSpec{
			NodeLabelsSelector: &v1.LabelSelector{
				MatchLabels: map[string]string{"scale_me": "true"},
			},
		},
	},
	NifiCluster: &v1alpha1.NifiCluster{
		Spec: v1alpha1.NifiClusterSpec{
			Nodes: []v1alpha1.Node{
				{Id: 2, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 3, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 4, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 5, NodeConfigGroup: "other-group", Labels: map[string]string{"other_group": "true"}},
			},
		},
		Status: v1alpha1.NifiClusterStatus{
			NodesState: map[string]v1alpha1.NodeState{
				"2": v1alpha1.NodeState{
					CreationTime: v1.NewTime(time.Now().UTC().Add(time.Duration(5) * time.Hour)),
				},
				"3": v1alpha1.NodeState{
					CreationTime: v1.NewTime(time.Now().UTC().Add(time.Duration(10) * time.Hour)),
				},
				"4": v1alpha1.NodeState{
					CreationTime: v1.NewTime(time.Now().UTC().Add(time.Duration(15) * time.Hour)),
				},
				"5": v1alpha1.NodeState{
					CreationTime: v1.NewTime(time.Now().UTC().Add(time.Duration(20) * time.Hour)),
				},
			},
		},
	},
}

func TestLIFORemoveAllNodes(t *testing.T) {
	// this also verifies if you remove more nodes than the current set, that it just returns an empty list.
	nodesToRemove, err := lifo.ScaleDown(100)

	if err != nil {
		t.Error("Should not have encountered an error")
	}

	if len(nodesToRemove) != 0 {
		t.Error("nodesToRemove should have been empty")
	}
}

func TestLIFORemoveSomeNodes(t *testing.T) {
	nodesToRemove, err := lifo.ScaleDown(2)

	if err != nil {
		t.Error("Should not have encountered an error")
	}
	if len(nodesToRemove) != 2 {
		t.Errorf("Did not remove correct number of nodes: %v+", nodesToRemove)
	}
	if nodesToRemove[0].Id != 3 && nodesToRemove[0].Id != 4 {
		t.Errorf("Incorrect results. Nodes: %v+", nodesToRemove)
	}
}

func TestLIFORemoveOneNode(t *testing.T) {
	nodesToRemove, err := lifo.ScaleDown(1)

	if err != nil {
		t.Error("Should not have encountered an error")
	}
	if len(nodesToRemove) != 1 {
		t.Errorf("Did not remove correct number of nodes: %v+", nodesToRemove)
	}

	if nodesToRemove[0].Id != 4 {
		t.Errorf("Incorrect results. Nodes: %v+", nodesToRemove)
	}
}

func TestLIFORemoveNoNodes(t *testing.T) {
	nodesToRemove, err := lifo.ScaleDown(0)

	if err != nil {
		t.Error("Should not have encountered an error")
	}
	if len(nodesToRemove) != 0 {
		t.Error("nodesToRemove should have been empty")
	}
}

func TestSimpleAddNodes(t *testing.T) {
	// 3 is enough to add nodes while considering other node config groups. nodes 0, 1, and 6 should be added.
	nodesToAdd, err := simple.ScaleUp(3)

	if err != nil {
		t.Error("Should not have encountered an error")
	}

	if len(nodesToAdd) != 3 {
		t.Error("nodesToAdd should have been 2")
	}

	if nodesToAdd[0].Id != 0 || nodesToAdd[1].Id != 1 || nodesToAdd[2].Id != 6 {
		t.Errorf("nodesToAdd Ids are not correct: %v+", nodesToAdd)
	}
}

func TestSimpleAddNoNodes(t *testing.T) {
	nodesToAdd, err := simple.ScaleUp(0)

	if err != nil {
		t.Error("Should not have encountered an error")
	}

	if len(nodesToAdd) != 0 {
		t.Errorf("nodesToAdd should have been empty: %v+", nodesToAdd)
	}
}
