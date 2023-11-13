package v1

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetCreationTimeOrderedNodes(t *testing.T) {
	time1 := v1.NewTime(time.Now().UTC().Add(time.Duration(5) * time.Hour))
	time2 := v1.NewTime(time.Now().UTC().Add(time.Duration(10) * time.Hour))
	time3 := v1.NewTime(time.Now().UTC().Add(time.Duration(15) * time.Hour))
	time4 := v1.NewTime(time.Now().UTC().Add(time.Duration(20) * time.Hour))

	cluster := &NifiCluster{
		Spec: NifiClusterSpec{
			Nodes: []Node{
				{Id: 2, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 3, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 4, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 5, NodeConfigGroup: "other-group", Labels: map[string]string{"other_group": "true"}},
			},
		},
		Status: NifiClusterStatus{
			NodesState: map[string]NodeState{
				"2": {
					CreationTime: &time1,
				},
				"3": {
					CreationTime: &time3,
				},
				"4": {
					CreationTime: &time2,
				},
				"5": {
					CreationTime: &time4,
				},
			},
		},
	}

	nodeList := cluster.GetCreationTimeOrderedNodes()

	if len(nodeList) != 4 {
		t.Errorf("Incorrect node list: %v+", nodeList)
	}
	if nodeList[0].Id != 2 || nodeList[1].Id != 4 || nodeList[2].Id != 3 || nodeList[3].Id != 5 {
		t.Errorf("Incorrect node list: %v+", nodeList)
	}
}

func TestListenersConfig(t *testing.T) {
	cluster := &NifiCluster{
		Spec: NifiClusterSpec{
			Nodes: []Node{
				{Id: 2, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 3, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 4, NodeConfigGroup: "scale-group", Labels: map[string]string{"scale_me": "true"}},
				{Id: 5, NodeConfigGroup: "other-group", Labels: map[string]string{"other_group": "true"}},
			},
			ListenersConfig: &ListenersConfig{
				InternalListeners: []InternalListenerConfig{
					{
						Name:     "foo",
					},
				},
			},
		},
	}

	// assert blank by default
	if cluster.Spec.ListenersConfig.InternalListeners[0].Protocol != "" {
		t.Errorf("incorrect protocol")
	}
	// set protocol
	cluster.Spec.ListenersConfig.InternalListeners[0].Protocol = corev1.ProtocolUDP
	if cluster.Spec.ListenersConfig.InternalListeners[0].Protocol != corev1.ProtocolUDP {
		t.Errorf("incorrect protocol. Should have been UDP")
	}
}
