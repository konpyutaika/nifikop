package nifi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

func TestDesiredNodesStableForClusterAPIs(t *testing.T) {
	t.Parallel()

	baseCluster := func() *v1.NifiCluster {
		return &v1.NifiCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster",
				Namespace: "namespace",
			},
			Spec: v1.NifiClusterSpec{
				Nodes: []v1.Node{
					{Id: 0},
					{Id: 1},
				},
			},
			Status: v1.NifiClusterStatus{
				NodesState: map[string]v1.NodeState{
					"0": {
						ConfigurationState: v1.ConfigInSync,
						PodIsReady:         true,
						GracefulActionState: v1.GracefulActionState{
							State: v1.GracefulUpscaleSucceeded,
						},
					},
					"1": {
						ConfigurationState: v1.ConfigInSync,
						PodIsReady:         true,
						GracefulActionState: v1.GracefulActionState{
							State: v1.GracefulUpscaleSucceeded,
						},
					},
				},
			},
		}
	}

	tests := []struct {
		name           string
		mutate         func(cluster *v1.NifiCluster)
		want           bool
		wantReasonPart string
	}{
		{
			name: "runs when all desired nodes are ready and succeeded",
			want: true,
		},
		{
			name: "skips when status missing",
			mutate: func(cluster *v1.NifiCluster) {
				delete(cluster.Status.NodesState, "1")
			},
			wantReasonPart: "status missing",
		},
		{
			name: "skips when pod is not ready",
			mutate: func(cluster *v1.NifiCluster) {
				state := cluster.Status.NodesState["0"]
				state.PodIsReady = false
				cluster.Status.NodesState["0"] = state
			},
			wantReasonPart: "pod not ready",
		},
		{
			name: "skips when config is not in sync",
			mutate: func(cluster *v1.NifiCluster) {
				state := cluster.Status.NodesState["0"]
				state.ConfigurationState = v1.ConfigOutOfSync
				cluster.Status.NodesState["0"] = state
			},
			wantReasonPart: "configurationState",
		},
		{
			name: "skips when graceful action still running",
			mutate: func(cluster *v1.NifiCluster) {
				state := cluster.Status.NodesState["1"]
				state.GracefulActionState.State = v1.GracefulUpscaleRunning
				cluster.Status.NodesState["1"] = state
			},
			wantReasonPart: "gracefulState",
		},
		{
			name: "ignores extra non desired node status entries",
			mutate: func(cluster *v1.NifiCluster) {
				cluster.Status.NodesState["2"] = v1.NodeState{
					ConfigurationState: v1.ConfigOutOfSync,
					PodIsReady:         false,
					GracefulActionState: v1.GracefulActionState{
						State: v1.GracefulDownscaleRunning,
					},
				}
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := baseCluster()
			if tt.mutate != nil {
				tt.mutate(cluster)
			}

			got, reason := desiredNodesStableForClusterAPIs(cluster)

			assert.Equal(t, tt.want, got)
			if tt.wantReasonPart != "" {
				assert.Contains(t, reason, tt.wantReasonPart)
			} else {
				assert.Empty(t, reason)
			}
		})
	}
}
