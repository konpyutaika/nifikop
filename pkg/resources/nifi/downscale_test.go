package nifi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

func TestOrphanedUpscaleDrainsBeforePodDeletion(t *testing.T) {
	for _, step := range []v1.ActionStep{v1.ConnectNodeAction, v1.ConnectStatus} {
		t.Run(string(step), func(t *testing.T) {
			ctx := context.Background()
			cluster := &v1.NifiCluster{
				ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "tenant"},
				Spec:       v1.NifiClusterSpec{Nodes: []v1.Node{{Id: 0}}},
				Status: v1.NifiClusterStatus{NodesState: map[string]v1.NodeState{
					"3": {GracefulActionState: v1.GracefulActionState{State: v1.GracefulUpscaleRunning, ActionStep: step}},
				}},
			}
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: cluster.Name + "-3-node", Namespace: cluster.Namespace, Labels: map[string]string{"app": "nifi", "nifi_cr": cluster.Name, "nodeId": "3"}},
				Status:     corev1.PodStatus{Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}},
			}
			s := runtime.NewScheme()
			require.NoError(t, v1.AddToScheme(s))
			require.NoError(t, corev1.AddToScheme(s))
			c := fake.NewClientBuilder().WithScheme(s).WithStatusSubresource(cluster).WithObjects(cluster, pod).Build()
			reconcile := func() {
				t.Helper()
				require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(cluster), cluster))
				r := New(c, c, s, cluster, *cluster.Status.DeepCopy())
				require.NoError(t, r.reconcileNifiPodDelete(*zap.NewNop()))
				require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(cluster), cluster))
			}
			reconcile()
			reconcile()
			require.True(t, cluster.Status.NodesState["3"].PodIsReady)
			require.Equal(t, v1.GracefulActionState{State: v1.GracefulDownscaleRequired}, cluster.Status.NodesState["3"].GracefulActionState)
			key := client.ObjectKeyFromObject(pod)
			require.NoError(t, c.Get(ctx, key, pod))
			state := cluster.Status.NodesState["3"]
			state.GracefulActionState = v1.GracefulActionState{State: v1.GracefulDownscaleRunning, ActionStep: v1.OffloadStatus}
			cluster.Status.NodesState["3"] = state
			require.NoError(t, c.Status().Update(ctx, cluster))
			reconcile()
			require.True(t, apierrors.IsNotFound(c.Get(ctx, key, pod)))
			reconcile()
			require.False(t, cluster.Status.NodesState["3"].PodIsReady)
			require.Equal(t, v1.RemovePodStatus, cluster.Status.NodesState["3"].GracefulActionState.ActionStep)
			require.Equal(t, v1.GracefulDownscaleRunning, cluster.Status.NodesState["3"].GracefulActionState.State)
		})
	}
}
