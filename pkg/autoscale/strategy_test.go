package autoscale

import (
	"reflect"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v12 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
)

var (
	time1 = v1.NewTime(time.Now().UTC().Add(time.Duration(5) * time.Hour))
	time2 = v1.NewTime(time.Now().UTC().Add(time.Duration(10) * time.Hour))
	time3 = v1.NewTime(time.Now().UTC().Add(time.Duration(15) * time.Hour))
	time4 = v1.NewTime(time.Now().UTC().Add(time.Duration(20) * time.Hour))
)

var lifo = LIFOHorizontalDownscaleStrategy{
	NifiNodeGroupAutoscaler: &v1alpha1.NifiNodeGroupAutoscaler{
		Spec: v1alpha1.NifiNodeGroupAutoscalerSpec{
			NodeLabelsSelector: &v1.LabelSelector{
				MatchLabels: map[string]string{"scale_me": "true"},
			},
		},
	},
	NifiCluster: &v12.NifiCluster{
		Spec: v12.NifiClusterSpec{
			Nodes: []v12.Node{
				{Id: 2, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 3, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 4, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 5, NodeConfigGroup: "other-group", Labels: map[string]string{"other_group": "true"}},
				{Id: 6, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
			},
		},
		Status: v12.NifiClusterStatus{
			NodesState: map[string]v12.NodeState{
				"2": {
					CreationTime: &time1,
				},
				"3": {
					CreationTime: &time2,
				},
				"4": {
					CreationTime: &time3,
				},
				"5": {
					CreationTime: &time4,
				},
				"6": {
					CreationTime: nil,
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
	NifiCluster: &v12.NifiCluster{
		Spec: v12.NifiClusterSpec{
			Nodes: []v12.Node{
				{Id: 2, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 3, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 4, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 5, NodeConfigGroup: "other-group", Labels: map[string]string{"other_group": "true"}},
				{Id: 6, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
			},
		},
		Status: v12.NifiClusterStatus{
			NodesState: map[string]v12.NodeState{
				"2": {
					CreationTime: &time1,
				},
				"3": {
					CreationTime: &time2,
				},
				"4": {
					CreationTime: &time3,
				},
				"5": {
					CreationTime: &time4,
				},
				"6": {
					CreationTime: nil,
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
	if nodesToRemove[0].Id != 4 || nodesToRemove[1].Id != 6 {
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

	if nodesToRemove[0].Id != 6 {
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

	if nodesToAdd[0].Id != 0 || nodesToAdd[1].Id != 1 || nodesToAdd[2].Id != 7 {
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

func TestComputeNewNodeIds(t *testing.T) {
	nodeList := []v12.Node{
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
