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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
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

func TestRemovedNodeStorageAndRestart(t *testing.T) {
	for _, tc := range []struct {
		name                          string
		shutdown, podPresent, reclaim bool
		step                          v1.ActionStep
	}{
		{name: "scale to zero", shutdown: true, podPresent: true, step: v1.ConnectStatus},
		{name: "scale to zero during offload", shutdown: true, podPresent: true, step: v1.OffloadStatus},
		{name: "evicted after offload", reclaim: true, step: v1.OffloadStatus},
		{name: "offloaded with pod", podPresent: true, reclaim: true, step: v1.OffloadStatus},
		{name: "restart after pod removal requested", step: v1.RemovePodAction},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cluster := &v1.NifiCluster{
				ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "tenant", UID: "cluster"},
				Spec:       v1.NifiClusterSpec{Nodes: []v1.Node{{Id: 0}}},
				Status:     v1.NifiClusterStatus{NodesState: map[string]v1.NodeState{"1": {PodIsReady: true, GracefulActionState: v1.GracefulActionState{State: v1.GracefulDownscaleRunning, ActionStep: tc.step}}}},
			}
			if tc.shutdown {
				cluster.Spec.Nodes = nil
			}
			newClaim := func(name, node, policy string, owner types.UID) *corev1.PersistentVolumeClaim {
				ref := templates.ClusterOwnerReference(cluster)
				ref.UID = owner
				return &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: cluster.Namespace, UID: types.UID(name),
					Labels:          map[string]string{"app": "nifi", "nifi_cr": cluster.Name, "nodeId": node, nifiutil.NifiDataVolumeMountKey: "true", nifiutil.NifiVolumeReclaimPolicyKey: policy},
					OwnerReferences: []metav1.OwnerReference{ref}}}
			}
			data := newClaim("data", "1", "Delete", cluster.UID)
			retained := newClaim("retained", "1", "Retain", cluster.UID)
			foreign := newClaim("foreign", "1", "Delete", "other-cluster")
			otherNode := newClaim("other-node", "0", "Delete", cluster.UID)
			secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "example-config-1", Namespace: cluster.Namespace}}
			service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "example-1-node", Namespace: cluster.Namespace}}
			objects := []client.Object{cluster, data, retained, foreign, otherNode, secret, service}
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "example-1-node", Namespace: cluster.Namespace, UID: "pod",
				Labels: map[string]string{"app": "nifi", "nifi_cr": cluster.Name, "nodeId": "1"}},
				Spec: corev1.PodSpec{Volumes: []corev1.Volume{{Name: "repo", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: data.Name}}}}}}
			if tc.podPresent {
				objects = append(objects, pod)
			}
			scheme := runtime.NewScheme()
			require.NoError(t, v1.AddToScheme(scheme))
			require.NoError(t, corev1.AddToScheme(scheme))
			k8s := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(cluster).WithObjects(objects...).Build()
			key := client.ObjectKeyFromObject(cluster)
			for range 3 {
				require.NoError(t, k8s.Get(ctx, key, cluster))
				r := New(k8s, k8s, scheme, cluster, *cluster.Status.DeepCopy())
				require.NoError(t, r.reconcileNifiPodDelete(*zap.NewNop()))
			}
			require.NoError(t, k8s.Get(ctx, key, cluster))
			state := cluster.Status.NodesState["1"]
			require.False(t, state.PodIsReady)
			require.Equal(t, v1.RemovePodStatus, state.GracefulActionState.ActionStep)
			want := v1.GracefulDownscaleRunning
			if tc.shutdown {
				want = v1.GracefulDownscaleSucceeded
			}
			require.Equal(t, want, state.GracefulActionState.State)
			for _, obj := range []client.Object{pod, secret, service} {
				require.True(t, apierrors.IsNotFound(k8s.Get(ctx, client.ObjectKeyFromObject(obj), obj)), obj.GetName())
			}
			for _, claim := range []*corev1.PersistentVolumeClaim{retained, foreign, otherNode} {
				require.NoError(t, k8s.Get(ctx, client.ObjectKeyFromObject(claim), claim))
			}
			err := k8s.Get(ctx, client.ObjectKeyFromObject(data), data)
			if tc.reclaim {
				require.True(t, apierrors.IsNotFound(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPendingPodCancellationRechecksScheduling(t *testing.T) {
	for name, schedulerWins := range map[string]bool{"unbound": false, "bind during deletion": true} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			cluster := &v1.NifiCluster{ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "tenant", UID: "cluster"},
				Spec:   v1.NifiClusterSpec{Nodes: []v1.Node{{Id: 0}}},
				Status: v1.NifiClusterStatus{NodesState: map[string]v1.NodeState{"1": {GracefulActionState: v1.GracefulActionState{State: v1.GracefulUpscaleRunning, ActionStep: v1.ConnectNodeAction}}}}}
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pending", Namespace: cluster.Namespace, UID: "first-pod",
				Labels: map[string]string{"app": "nifi", "nifi_cr": cluster.Name, "nodeId": "1"}}, Status: corev1.PodStatus{Phase: corev1.PodPending}}
			scheme := runtime.NewScheme()
			require.NoError(t, v1.AddToScheme(scheme))
			require.NoError(t, corev1.AddToScheme(scheme))
			k8s := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(cluster).WithObjects(cluster, pod).WithInterceptorFuncs(interceptor.Funcs{
				Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
					if p, ok := obj.(*corev1.Pod); ok {
						pre := (&client.DeleteOptions{}).ApplyOptions(opts).Preconditions
						require.NotNil(t, pre)
						require.NotNil(t, pre.UID)
						require.NotNil(t, pre.ResourceVersion)
						require.Equal(t, p.UID, *pre.UID)
						require.Equal(t, p.ResourceVersion, *pre.ResourceVersion)
						if schedulerWins {
							bound := p.DeepCopy()
							bound.Spec.NodeName = "worker"
							require.NoError(t, c.Update(ctx, bound))
						}
					}
					return c.Delete(ctx, obj, opts...)
				},
			}).Build()
			key := client.ObjectKeyFromObject(cluster)
			run := func() error {
				t.Helper()
				require.NoError(t, k8s.Get(ctx, key, cluster))
				return New(k8s, k8s, scheme, cluster, *cluster.Status.DeepCopy()).reconcileNifiPodDelete(*zap.NewNop())
			}
			require.NoError(t, run())
			require.Equal(t, v1.GracefulActionState{State: v1.GracefulDownscaleRequired, ActionStep: v1.CancelPendingPodAction}, cluster.Status.NodesState["1"].GracefulActionState)
			require.NoError(t, k8s.Get(ctx, client.ObjectKeyFromObject(pod), pod))
			// The task controller confirmed this node is absent from the NiFi membership list.
			cluster.Status.NodesState["1"] = v1.NodeState{GracefulActionState: v1.GracefulActionState{State: v1.GracefulDownscaleRunning, ActionStep: v1.CancelPendingPodAction}}
			require.NoError(t, k8s.Status().Update(ctx, cluster))
			err := run()
			if schedulerWins {
				require.True(t, apierrors.IsConflict(err))
				require.NoError(t, run())
				require.Equal(t, v1.GracefulActionState{State: v1.GracefulDownscaleRequired}, cluster.Status.NodesState["1"].GracefulActionState)
				require.NoError(t, k8s.Get(ctx, client.ObjectKeyFromObject(pod), pod))
				require.Equal(t, "worker", pod.Spec.NodeName)
			} else {
				require.NoError(t, err)
				require.NoError(t, run())
				require.True(t, apierrors.IsNotFound(k8s.Get(ctx, client.ObjectKeyFromObject(pod), pod)))
				require.Equal(t, v1.RemovePodStatus, cluster.Status.NodesState["1"].GracefulActionState.ActionStep)
			}
		})
	}
}

func TestStaleOffloadCannotReclaimClaimsAfterShutdownRequested(t *testing.T) {
	ctx := context.Background()
	cluster := &v1.NifiCluster{
		ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "tenant", UID: "cluster", Generation: 1},
		Spec:       v1.NifiClusterSpec{Nodes: []v1.Node{{Id: 0}}},
		Status: v1.NifiClusterStatus{NodesState: map[string]v1.NodeState{
			"1": {GracefulActionState: v1.GracefulActionState{State: v1.GracefulDownscaleRunning, ActionStep: v1.OffloadStatus}},
		}},
	}
	latest := cluster.DeepCopy()
	latest.Generation++
	latest.Spec.Nodes = nil
	claim := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
		Name: "repo", Namespace: cluster.Namespace, UID: "repo",
		OwnerReferences: []metav1.OwnerReference{templates.ClusterOwnerReference(cluster)},
		Labels:          map[string]string{"nifi_cr": cluster.Name, "nodeId": "1", nifiutil.NifiDataVolumeMountKey: "true", nifiutil.NifiVolumeReclaimPolicyKey: "Delete"},
	}}
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, v1.AddToScheme(scheme))
	k8s := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(cluster).WithObjects(latest, claim).Build()
	r := New(k8s, k8s, scheme, cluster, *cluster.Status.DeepCopy())
	require.True(t, apierrors.IsConflict(r.reconcileNifiPodDelete(*zap.NewNop())))
	require.NoError(t, k8s.Get(ctx, client.ObjectKeyFromObject(claim), claim))
}
