package controller

import (
	"context"
	"testing"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

const (
	hibernationTestCluster   = "hibernate-test"
	hibernationTestNamespace = "nifi"
)

func hibernationTestObjects() []client.Object {
	cluster := &v1.NifiCluster{
		ObjectMeta: metav1.ObjectMeta{Name: hibernationTestCluster, Namespace: hibernationTestNamespace},
		Spec: v1.NifiClusterSpec{
			Hibernate:       true,
			ListenersConfig: &v1.ListenersConfig{},
			Nodes:           []v1.Node{{Id: 0}, {Id: 1}},
		},
		Status: v1.NifiClusterStatus{
			State: v1.NifiClusterRunning,
			NodesState: map[string]v1.NodeState{
				"0": {PodIsReady: true, ConfigurationState: v1.ConfigInSync, GracefulActionState: v1.GracefulActionState{State: v1.GracefulUpscaleSucceeded}},
				"1": {PodIsReady: true, ConfigurationState: v1.ConfigInSync, GracefulActionState: v1.GracefulActionState{State: v1.GracefulUpscaleSucceeded}},
			},
		},
	}

	objs := make([]client.Object, 0, 9)
	objs = append(objs, cluster)
	for _, id := range []string{"0", "1"} {
		labels := nifiutil.LabelsForNifi(hibernationTestCluster)
		labels["nodeId"] = id
		objs = append(objs,
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: hibernationTestCluster + "-" + id, Namespace: hibernationTestNamespace, Labels: labels}},
			// Data PVC with the Delete reclaim policy: removing the node from spec.nodes would delete it.
			&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
				Name:      hibernationTestCluster + "-" + id + "-data",
				Namespace: hibernationTestNamespace,
				Labels: map[string]string{
					"nifi_cr":                           hibernationTestCluster,
					"nodeId":                            id,
					nifiutil.NifiDataVolumeMountKey:     "true",
					nifiutil.NifiVolumeReclaimPolicyKey: string(corev1.PersistentVolumeReclaimDelete),
				},
			}},
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: hibernationTestCluster + "-config-" + id, Namespace: hibernationTestNamespace, Labels: labels}},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: hibernationTestCluster + "-" + id, Namespace: hibernationTestNamespace, Labels: labels}},
		)
	}
	return objs
}

// hibernationTestListeners returns the listeners the usual reconcile needs to build the node pods.
func hibernationTestListeners() *v1.ListenersConfig {
	return &v1.ListenersConfig{InternalListeners: []v1.InternalListenerConfig{
		{Type: "http", Name: "http", ContainerPort: 8080},
		{Type: "cluster", Name: "cluster", ContainerPort: 6007},
		{Type: "s2s", Name: "s2s", ContainerPort: 10000},
	}}
}

func newHibernationTestReconciler(t *testing.T, objs ...client.Object) *NifiClusterReconciler {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		WithStatusSubresource(&v1.NifiCluster{}).
		Build()
	logger := zap.NewNop()
	return &NifiClusterReconciler{
		Client:       c,
		DirectClient: c,
		Log:          *logger,
		Scheme:       scheme,
		Recorder:     record.NewFakeRecorder(100),
		RequeueIntervals: map[string]int{
			"CLUSTER_TASK_RUNNING_REQUEUE_INTERVAL":   1,
			"CLUSTER_TASK_NO_NODE_INTERVAL":           1,
			"CLUSTER_TASK_NOT_READY_REQUEUE_INTERVAL": 1,
		},
	}
}

func reconcileHibernationTestCluster(t *testing.T, r *NifiClusterReconciler) *v1.NifiCluster {
	t.Helper()
	key := types.NamespacedName{Name: hibernationTestCluster, Namespace: hibernationTestNamespace}
	if _, err := r.Reconcile(context.TODO(), reconcile.Request{NamespacedName: key}); err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}
	cluster := &v1.NifiCluster{}
	if err := r.Client.Get(context.TODO(), key, cluster); err != nil {
		t.Fatal(err)
	}
	return cluster
}

func countObjects(t *testing.T, c client.Client, list client.ObjectList) int {
	t.Helper()
	if err := c.List(context.TODO(), list, client.InNamespace(hibernationTestNamespace)); err != nil {
		t.Fatal(err)
	}
	switch l := list.(type) {
	case *corev1.PodList:
		return len(l.Items)
	case *corev1.PersistentVolumeClaimList:
		return len(l.Items)
	case *corev1.SecretList:
		return len(l.Items)
	case *corev1.ServiceList:
		return len(l.Items)
	}
	t.Fatalf("unsupported list type %T", list)
	return 0
}

func TestHibernationStopsPodsAndKeepsState(t *testing.T) {
	r := newHibernationTestReconciler(t, hibernationTestObjects()...)

	// First pass deletes the pods.
	cluster := reconcileHibernationTestCluster(t, r)
	if cluster.Status.State != v1.NifiClusterHibernating {
		t.Errorf("expected state %s, got %s", v1.NifiClusterHibernating, cluster.Status.State)
	}

	// Second pass sees no pod left.
	cluster = reconcileHibernationTestCluster(t, r)
	if cluster.Status.State != v1.NifiClusterHibernated {
		t.Errorf("expected state %s, got %s", v1.NifiClusterHibernated, cluster.Status.State)
	}

	if n := countObjects(t, r.Client, &corev1.PodList{}); n != 0 {
		t.Errorf("expected all pods deleted, %d left", n)
	}
	if n := countObjects(t, r.Client, &corev1.PersistentVolumeClaimList{}); n != 2 {
		t.Errorf("expected 2 PVCs kept despite the Delete reclaim policy, got %d", n)
	}
	if n := countObjects(t, r.Client, &corev1.SecretList{}); n != 2 {
		t.Errorf("expected 2 node config secrets kept, got %d", n)
	}
	if n := countObjects(t, r.Client, &corev1.ServiceList{}); n != 2 {
		t.Errorf("expected 2 node services kept, got %d", n)
	}

	if len(cluster.Status.NodesState) != 2 {
		t.Fatalf("expected node status kept for 2 nodes, got %d", len(cluster.Status.NodesState))
	}
	for id, state := range cluster.Status.NodesState {
		if state.PodIsReady {
			t.Errorf("expected node %s marked not ready", id)
		}
		if state.GracefulActionState.State != v1.GracefulUpscaleSucceeded {
			t.Errorf("expected node %s graceful state unchanged, got %s", id, state.GracefulActionState.State)
		}
	}
	if cluster.IsReady() {
		t.Error("expected hibernated cluster not to be ready")
	}
	if len(cluster.GetFinalizers()) == 0 {
		t.Error("expected finalizers on hibernated cluster")
	}
}

func TestHibernationWithoutNodes(t *testing.T) {
	cluster := &v1.NifiCluster{
		ObjectMeta: metav1.ObjectMeta{Name: hibernationTestCluster, Namespace: hibernationTestNamespace},
		Spec:       v1.NifiClusterSpec{Hibernate: true, ListenersConfig: &v1.ListenersConfig{}},
	}
	r := newHibernationTestReconciler(t, cluster)

	got := reconcileHibernationTestCluster(t, r)
	if got.Status.State != v1.NifiClusterHibernated {
		t.Errorf("expected state %s, got %s", v1.NifiClusterHibernated, got.Status.State)
	}
}

func TestWakeUpFromHibernation(t *testing.T) {
	for _, state := range []v1.ClusterState{v1.NifiClusterHibernating, v1.NifiClusterHibernated} {
		cluster := &v1.NifiCluster{
			ObjectMeta: metav1.ObjectMeta{Name: hibernationTestCluster, Namespace: hibernationTestNamespace},
			Status: v1.NifiClusterStatus{
				State:      state,
				NodesState: map[string]v1.NodeState{"0": {InitClusterNode: v1.IsInitClusterNode}},
			},
		}
		r := newHibernationTestReconciler(t, cluster)
		current := cluster.DeepCopy()

		if err := r.wakeUpFromHibernation(cluster, current); err != nil {
			t.Fatalf("wake up failed: %v", err)
		}
		got := &v1.NifiCluster{}
		if err := r.Client.Get(context.TODO(), client.ObjectKeyFromObject(cluster), got); err != nil {
			t.Fatal(err)
		}
		if got.Status.State != v1.NifiClusterInitialized {
			t.Errorf("from %s: expected state %s, got %s", state, v1.NifiClusterInitialized, got.Status.State)
		}
		if got.Status.State.IsReady() {
			t.Errorf("from %s: expected woken cluster not ready until nodes are up", state)
		}
	}
}

func TestWakeUpClusterCreatedHibernated(t *testing.T) {
	cluster := &v1.NifiCluster{
		ObjectMeta: metav1.ObjectMeta{Name: hibernationTestCluster, Namespace: hibernationTestNamespace},
		Spec: v1.NifiClusterSpec{
			Hibernate:       true,
			ListenersConfig: hibernationTestListeners(),
			Nodes:           []v1.Node{{Id: 0, NodeConfig: &v1.NodeConfig{}}, {Id: 1, NodeConfig: &v1.NodeConfig{}}},
		},
	}
	r := newHibernationTestReconciler(t, cluster)

	got := reconcileHibernationTestCluster(t, r)
	if got.Status.State != v1.NifiClusterHibernated {
		t.Fatalf("expected state %s, got %s", v1.NifiClusterHibernated, got.Status.State)
	}

	got.Spec.Hibernate = false
	if err := r.Client.Update(context.TODO(), got); err != nil {
		t.Fatal(err)
	}

	// The rest of the reconcile needs a real cluster: only the initialization step is checked here.
	key := client.ObjectKeyFromObject(cluster)
	_, _ = r.Reconcile(context.TODO(), reconcile.Request{NamespacedName: key})
	if err := r.Client.Get(context.TODO(), key, got); err != nil {
		t.Fatal(err)
	}

	if got.Status.State == v1.NifiClusterHibernated || got.Status.State == v1.NifiClusterInitializing {
		t.Errorf("expected cluster initialized after first wake-up, got state %s", got.Status.State)
	}
	for _, id := range []string{"0", "1"} {
		if !got.Status.NodesState[id].InitClusterNode {
			t.Errorf("expected node %s marked as init cluster node", id)
		}
	}
}

func TestNodesBlockingHibernation(t *testing.T) {
	done := v1.NodeState{GracefulActionState: v1.GracefulActionState{State: v1.GracefulUpscaleSucceeded}}
	tests := []struct {
		name     string
		state    v1.ClusterState
		nodes    map[string]v1.NodeState
		expected []string
	}{
		{name: "no node change", state: v1.NifiClusterRunning, nodes: map[string]v1.NodeState{"0": done, "1": done}, expected: []string{}},
		{name: "no node status", state: "", nodes: nil, expected: []string{}},
		{name: "downscale offloaded", state: v1.NifiClusterRunning, nodes: map[string]v1.NodeState{"0": done, "1": done,
			"2": {GracefulActionState: v1.GracefulActionState{State: v1.GracefulDownscaleRunning, ActionStep: v1.OffloadStatus}}},
			expected: []string{"2"}},
		{name: "removed node not yet downscaling", state: v1.NifiClusterRunning, nodes: map[string]v1.NodeState{"0": done, "1": done, "2": done},
			expected: []string{"2"}},
		{name: "upscale required", state: v1.NifiClusterRunning, nodes: map[string]v1.NodeState{"0": done,
			"1": {GracefulActionState: v1.GracefulActionState{State: v1.GracefulUpscaleRequired}}},
			expected: []string{"1"}},
		{name: "hibernation already started", state: v1.NifiClusterHibernating, nodes: map[string]v1.NodeState{"0": done, "1": done,
			"2": {GracefulActionState: v1.GracefulActionState{State: v1.GracefulDownscaleRunning}}},
			expected: nil},
	}
	for _, tt := range tests {
		cluster := &v1.NifiCluster{
			Spec:   v1.NifiClusterSpec{Hibernate: true, Nodes: []v1.Node{{Id: 0}, {Id: 1}}},
			Status: v1.NifiClusterStatus{State: tt.state, NodesState: tt.nodes},
		}
		got := nodesBlockingHibernation(cluster)
		if len(got) != len(tt.expected) {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, got)
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, got)
			}
		}
	}
}

func TestHibernationWaitsForDownscale(t *testing.T) {
	objs := hibernationTestObjects()
	cluster := objs[0].(*v1.NifiCluster)
	// Node 1 was removed from spec.nodes and its downscale is offloading.
	cluster.Spec.Nodes = []v1.Node{{Id: 0, NodeConfig: &v1.NodeConfig{}}}
	cluster.Spec.ListenersConfig = hibernationTestListeners()
	cluster.Status.NodesState["1"] = v1.NodeState{PodIsReady: true, ConfigurationState: v1.ConfigInSync,
		GracefulActionState: v1.GracefulActionState{State: v1.GracefulDownscaleRunning, ActionStep: v1.OffloadStatus}}
	r := newHibernationTestReconciler(t, objs...)

	// The usual reconcile runs instead of hibernation; it needs a real cluster, so only hibernation is checked.
	key := client.ObjectKeyFromObject(cluster)
	_, _ = r.Reconcile(context.TODO(), reconcile.Request{NamespacedName: key})
	got := &v1.NifiCluster{}
	if err := r.Client.Get(context.TODO(), key, got); err != nil {
		t.Fatal(err)
	}
	if got.Status.State.IsHibernation() {
		t.Errorf("expected hibernation to wait for the downscale, got state %s", got.Status.State)
	}
	if !got.Status.NodesState["0"].PodIsReady {
		t.Error("expected node 0 not stopped while the downscale runs")
	}
}

func TestHibernationWithoutListenersConfig(t *testing.T) {
	cluster := &v1.NifiCluster{
		ObjectMeta: metav1.ObjectMeta{Name: hibernationTestCluster, Namespace: hibernationTestNamespace},
		Spec:       v1.NifiClusterSpec{Hibernate: true},
	}
	r := newHibernationTestReconciler(t, cluster)

	got := reconcileHibernationTestCluster(t, r)
	if got.Status.State != v1.NifiClusterHibernated {
		t.Errorf("expected state %s, got %s", v1.NifiClusterHibernated, got.Status.State)
	}
}
