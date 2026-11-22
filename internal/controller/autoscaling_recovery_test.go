package controller

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
	pkicommon "github.com/konpyutaika/nifikop/pkg/util/pki"
)

func scalingCluster() *v1.NifiCluster {
	template := "example-node-%d"
	return &v1.NifiCluster{
		TypeMeta:   metav1.TypeMeta{APIVersion: v1.GroupVersion.String(), Kind: "NifiCluster"},
		ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "tenant", UID: "cluster-uid"},
		Spec: v1.NifiClusterSpec{
			NodeUserIdentityTemplate: &template,
			ListenersConfig:          &v1.ListenersConfig{InternalListeners: []v1.InternalListenerConfig{{Type: v1.HttpListenerType, ContainerPort: 8080}}},
			Nodes:                    []v1.Node{{Id: 0, Labels: map[string]string{"pool": "a"}}, {Id: 1, Labels: map[string]string{"pool": "a"}}},
		},
		Status: v1.NifiClusterStatus{
			State: v1.NifiClusterRunning,
			NodesState: map[string]v1.NodeState{
				"0": {GracefulActionState: v1.GracefulActionState{State: v1.GracefulUpscaleSucceeded}, ConfigurationState: v1.ConfigInSync, PodIsReady: true},
				"1": {GracefulActionState: v1.GracefulActionState{State: v1.GracefulUpscaleRunning, ActionStep: v1.ConnectStatus}, ConfigurationState: v1.ConfigInSync},
			},
		},
	}
}

func scalingClient(t *testing.T, objects ...client.Object) client.Client {
	t.Helper()
	s := runtime.NewScheme()
	require.NoError(t, v1.AddToScheme(s))
	require.NoError(t, v1alpha1.AddToScheme(s))
	require.NoError(t, corev1.AddToScheme(s))
	return fake.NewClientBuilder().WithScheme(s).WithStatusSubresource(&v1.NifiCluster{}, &v1.NifiUser{}, &v1alpha1.NifiNodeGroupAutoscaler{}).WithObjects(objects...).Build()
}

type scalingNiFi struct {
	nificlient.NifiClient
	user          *nigoapi.UserEntity
	grants        map[string][]nigoapi.TenantEntity
	policyErr     error
	membership    map[int32]string
	membershipErr error
	disconnects   []int32
	offloads      []int32
	removals      []int32
}

func (n *scalingNiFi) GetUsers() ([]nigoapi.UserEntity, error) { return nil, nil }
func (n *scalingNiFi) CreateUser(user nigoapi.UserEntity) (*nigoapi.UserEntity, error) {
	user.Id = "joining-user"
	n.user = &user
	return n.user, nil
}
func (n *scalingNiFi) GetUser(string) (*nigoapi.UserEntity, error) { return n.user, nil }
func (n *scalingNiFi) GetAccessPolicy(action, resource string) (*nigoapi.AccessPolicyEntity, error) {
	return &nigoapi.AccessPolicyEntity{Component: &nigoapi.AccessPolicyDto{Action: action, Resource: resource}}, nil
}
func (n *scalingNiFi) UpdateAccessPolicy(policy nigoapi.AccessPolicyEntity) (*nigoapi.AccessPolicyEntity, error) {
	if n.policyErr != nil {
		return nil, n.policyErr
	}
	n.grants[policy.Component.Action+policy.Component.Resource] = policy.Component.Users
	return &policy, nil
}
func (n *scalingNiFi) GetClusterNode(id int32) (*nigoapi.NodeEntity, error) {
	if n.membershipErr != nil {
		return nil, n.membershipErr
	}
	status := string(v1.ConnectStatus)
	if n.membership != nil {
		var ok bool
		if status, ok = n.membership[id]; !ok {
			return nil, nificlient.ErrNifiClusterNodeNotFound
		}
	}
	return &nigoapi.NodeEntity{Node: &nigoapi.NodeDto{Status: status}}, nil
}
func (n *scalingNiFi) DisconnectClusterNode(id int32) (*nigoapi.NodeEntity, error) {
	n.disconnects = append(n.disconnects, id)
	return n.GetClusterNode(id)
}
func (n *scalingNiFi) OffloadClusterNode(id int32) (*nigoapi.NodeEntity, error) {
	n.offloads = append(n.offloads, id)
	return n.GetClusterNode(id)
}
func (n *scalingNiFi) RemoveClusterNode(id int32) error {
	n.removals = append(n.removals, id)
	delete(n.membership, id)
	return nil
}

func useScalingNiFi(t *testing.T, n *scalingNiFi) {
	prev := common.NewNifiFromConfig
	t.Cleanup(func() { common.NewNifiFromConfig = prev })
	common.NewNifiFromConfig = func(*clientconfig.NifiConfig, *zap.Logger) (nificlient.NifiClient, error) { return n, nil }
}

func TestJoiningNodeProxyPolicies(t *testing.T) {
	ctx := context.Background()
	cluster := scalingCluster()
	user := pkicommon.NodeUsersForCluster(cluster, nil)[1]
	createCert := false // Exercise the gate after certificate reconciliation.
	user.Spec.CreateCert = &createCert
	k8s := scalingClient(t, cluster, user)
	nifi := &scalingNiFi{policyErr: errors.New("policy revision conflict"), grants: make(map[string][]nigoapi.TenantEntity)}
	useScalingNiFi(t, nifi)
	r := NifiUserReconciler{Client: k8s, Log: *zap.NewNop(), Recorder: record.NewFakeRecorder(32)}
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(user)}
	_, err := r.Reconcile(ctx, req)
	require.ErrorIs(t, err, nifi.policyErr)
	nifi.policyErr = nil
	_, err = r.Reconcile(ctx, req)
	require.NoError(t, err)
	require.Equal(t, user.Name, nifi.user.Component.Identity)
	for _, policy := range []string{"read/proxy", "write/proxy"} {
		require.Equal(t, []nigoapi.TenantEntity{{Id: "joining-user"}}, nifi.grants[policy])
	}
	require.NoError(t, k8s.Get(ctx, req.NamespacedName, user))
	require.Equal(t, "joining-user", user.Status.Id)
	require.False(t, cluster.IsReady())

	cluster.Spec.Nodes = cluster.Spec.Nodes[:1]
	require.True(t, isClusterNodeUser(cluster, user), "legacy orphan identity")
	require.False(t, isClusterNodeUser(cluster, pkicommon.ControllerUserForCluster(cluster)))
	user.OwnerReferences[0].UID = "other-cluster"
	require.False(t, isClusterNodeUser(cluster, user))
}

func TestOrdinaryUserWaitsForClusterReadiness(t *testing.T) {
	ctx := context.Background()
	cluster := scalingCluster()
	user := pkicommon.NodeUsersForCluster(cluster, nil)[1]
	user.OwnerReferences = nil
	createCert := false
	user.Spec.CreateCert = &createCert
	k8s := scalingClient(t, cluster, user)
	recorder := record.NewFakeRecorder(8)
	r := NifiUserReconciler{Client: k8s, Log: *zap.NewNop(), Recorder: recorder}
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(user)}
	result, err := r.Reconcile(ctx, req)
	require.NoError(t, err)
	require.Positive(t, result.RequeueAfter)
	require.NoError(t, k8s.Get(ctx, req.NamespacedName, user))
	const blockedSince = "nifi.konpyutaika.com/cluster-not-ready-since"
	require.NotEmpty(t, user.Annotations[blockedSince])
	require.NotEmpty(t, recorder.Events)
	require.Contains(t, <-recorder.Events, "Normal ReferenceClusterNotReady")
	user.Annotations[blockedSince] = time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	require.NoError(t, k8s.Update(ctx, user))
	_, err = r.Reconcile(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, recorder.Events)
	require.Contains(t, <-recorder.Events, "Warning ReferenceClusterNotReady")
}

func TestConnectedTaskCompletesBeforePodReadiness(t *testing.T) {
	for _, step := range []v1.ActionStep{v1.ConnectNodeAction, v1.ConnectStatus} {
		t.Run(string(step), func(t *testing.T) {
			ctx := context.Background()
			cluster := scalingCluster()
			state := cluster.Status.NodesState["1"]
			state.GracefulActionState.ActionStep = step
			state.GracefulActionState.TaskStarted = time.Now().UTC().Add(-time.Hour).Format(nifiutil.TimeStampLayout)
			cluster.Status.NodesState["1"] = state
			k8s := scalingClient(t, cluster)
			useScalingNiFi(t, &scalingNiFi{})
			recorder := record.NewFakeRecorder(8)
			r := NifiClusterTaskReconciler{Client: k8s, Log: *zap.NewNop(), Recorder: recorder}
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cluster)}
			_, err := r.Reconcile(ctx, req)
			require.NoError(t, err)
			require.NoError(t, k8s.Get(ctx, req.NamespacedName, cluster))
			require.Equal(t, v1.GracefulActionState{State: v1.GracefulUpscaleSucceeded}, cluster.Status.NodesState["1"].GracefulActionState)
			require.False(t, cluster.Status.NodesState["1"].PodIsReady)
			require.NotEmpty(t, recorder.Events)
			require.Contains(t, <-recorder.Events, "Warning GracefulTaskOverdue")
		})
	}
}

func TestAutoscalerWaitsForJoiningAndDrainingNodes(t *testing.T) {
	ctx := context.Background()
	cluster := scalingCluster()
	group := &v1alpha1.NifiNodeGroupAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: "scale", Namespace: "controllers"},
		Spec: v1alpha1.NifiNodeGroupAutoscalerSpec{Replicas: 1, ClusterRef: v1.ClusterReference{Name: cluster.Name, Namespace: cluster.Namespace},
			NodeLabelsSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"pool": "a"}}}}
	pod := func(id int, pool, owner string) *corev1.Pod {
		return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%d", owner, id), Namespace: cluster.Namespace,
			Labels: map[string]string{"app": "nifi", "nifi_cr": owner, "pool": pool, "nodeId": fmt.Sprint(id)}}}
	}
	pod1 := pod(1, "a", cluster.Name)
	k8s := scalingClient(t, cluster, group, pod(0, "a", cluster.Name), pod1, pod(2, "b", cluster.Name), pod(4, "a", "other"))
	r := NifiNodeGroupAutoscalerReconciler{Client: k8s, APIReader: k8s, Log: *zap.NewNop(), Recorder: record.NewFakeRecorder(20)}
	key := client.ObjectKeyFromObject(group)
	reconcile := func() {
		t.Helper()
		_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})
		require.NoError(t, err)
		require.NoError(t, k8s.Get(ctx, key, group))
		require.NoError(t, k8s.Get(ctx, client.ObjectKeyFromObject(cluster), cluster))
	}
	reconcile()
	require.Len(t, cluster.Spec.Nodes, 2)
	require.Equal(t, v1alpha1.AutoscalerStateOutOfSync, group.Status.State)
	require.EqualValues(t, 2, group.Status.Replicas)
	require.Contains(t, string(group.Status.Selector), "nifi_cr="+cluster.Name)

	state := cluster.Status.NodesState["1"]
	state.GracefulActionState.State = v1.GracefulUpscaleSucceeded
	cluster.Status.NodesState["1"] = state
	require.NoError(t, k8s.Status().Update(ctx, cluster))
	reconcile()
	require.Len(t, cluster.Spec.Nodes, 1)
	require.Equal(t, v1alpha1.AutoscalerStateOutOfSync, group.Status.State)

	group.Spec.Replicas = 2
	require.NoError(t, k8s.Update(ctx, group))
	reconcile()
	require.Len(t, cluster.Spec.Nodes, 1, "draining replica still occupies the requested capacity")
	require.Equal(t, v1alpha1.AutoscalerStateOutOfSync, group.Status.State)
	require.NoError(t, k8s.Delete(ctx, pod1))
	reconcile()
	require.Len(t, cluster.Spec.Nodes, 2)
	require.EqualValues(t, 3, cluster.Spec.Nodes[1].Id, "IDs held by status and another group's pod are reserved")
	require.Equal(t, v1alpha1.AutoscalerStateOutOfSync, group.Status.State)
	require.NoError(t, k8s.Create(ctx, pod(3, "a", cluster.Name)))
	reconcile()
	require.Equal(t, v1alpha1.AutoscalerStateInSync, group.Status.State)
	require.EqualValues(t, 2, group.Status.Replicas)
}

func TestWaitingJoinAllowsDrainAndStatusCleanup(t *testing.T) {
	ctx := context.Background()
	cluster := scalingCluster()
	started := time.Now().UTC().Format(nifiutil.TimeStampLayout)
	cluster.Status.NodesState["1"] = v1.NodeState{GracefulActionState: v1.GracefulActionState{
		State: v1.GracefulUpscaleRunning, ActionStep: v1.ConnectNodeAction, TaskStarted: started,
	}}
	cluster.Status.NodesState["2"] = v1.NodeState{GracefulActionState: v1.GracefulActionState{State: v1.GracefulDownscaleRequired}}
	cluster.Status.NodesState["3"] = v1.NodeState{GracefulActionState: v1.GracefulActionState{State: v1.GracefulDownscaleSucceeded}}
	k8s := scalingClient(t, cluster)
	nifi := &scalingNiFi{membership: map[int32]string{2: string(v1.DisconnectStatus)}}
	useScalingNiFi(t, nifi)
	r := NifiClusterTaskReconciler{Client: k8s, Log: *zap.NewNop(), Recorder: record.NewFakeRecorder(8)}
	key := client.ObjectKeyFromObject(cluster)
	for _, step := range []v1.ActionStep{v1.DisconnectNodeAction, v1.DisconnectStatus, v1.OffloadNodeAction, v1.OffloadStatus} {
		result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})
		require.NoError(t, err)
		require.Positive(t, result.RequeueAfter)
		require.NoError(t, k8s.Get(ctx, key, cluster))
		require.Equal(t, step, cluster.Status.NodesState["2"].GracefulActionState.ActionStep)
		require.Equal(t, v1.ConnectNodeAction, cluster.Status.NodesState["1"].GracefulActionState.ActionStep)
		require.NotContains(t, cluster.Status.NodesState, "3")
		if step == v1.OffloadNodeAction {
			nifi.membership[2] = string(v1.OffloadStatus)
		}
	}
	require.Equal(t, []int32{2}, nifi.disconnects)
	require.Equal(t, []int32{2}, nifi.offloads)

	cluster.Status.NodesState["2"] = v1.NodeState{GracefulActionState: v1.GracefulActionState{State: v1.GracefulDownscaleRunning, ActionStep: v1.RemovePodStatus}}
	require.NoError(t, k8s.Status().Update(ctx, cluster))
	for range 3 {
		_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})
		require.NoError(t, err)
	}
	require.NoError(t, k8s.Get(ctx, key, cluster))
	require.NotContains(t, cluster.Status.NodesState, "2")
	require.Equal(t, []int32{2}, nifi.removals)
}

func TestDownscaleRollbackReconnectDeadline(t *testing.T) {
	now := time.Now().UTC()
	for _, tc := range []struct {
		name, started string
		expired       bool
	}{
		{"within deadline", now.Format(nifiutil.TimeStampLayout), false},
		{"expired", now.Add(-time.Hour).Format(nifiutil.TimeStampLayout), true},
		{"legacy missing start", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cluster := scalingCluster()
			cluster.Spec.Nodes = cluster.Spec.Nodes[:1]
			action := v1.GracefulActionState{State: v1.GracefulDownscaleRequired, ActionStep: v1.ConnectNodeAction, TaskStarted: tc.started}
			cluster.Status.NodesState["1"] = v1.NodeState{GracefulActionState: action}
			k8s := scalingClient(t, cluster)
			nifi := &scalingNiFi{membership: map[int32]string{1: string(v1.ConnectNodeAction)}}
			useScalingNiFi(t, nifi)
			recorder := record.NewFakeRecorder(8)
			r := NifiClusterTaskReconciler{Client: k8s, Log: *zap.NewNop(), Recorder: recorder}
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cluster)}
			_, err := r.Reconcile(ctx, req)
			require.NoError(t, err)
			require.NoError(t, k8s.Get(ctx, req.NamespacedName, cluster))
			require.Empty(t, nifi.disconnects)
			if !tc.expired {
				require.Equal(t, action, cluster.Status.NodesState["1"].GracefulActionState)
				return
			}
			require.Equal(t, v1.GracefulActionState{State: v1.GracefulDownscaleRequired}, cluster.Status.NodesState["1"].GracefulActionState)
			require.NotEmpty(t, recorder.Events)
			require.Contains(t, <-recorder.Events, "Warning RollbackReconnectExpired")
			_, err = r.Reconcile(ctx, req)
			require.NoError(t, err)
			require.NoError(t, k8s.Get(ctx, req.NamespacedName, cluster))
			require.Equal(t, v1.DisconnectNodeAction, cluster.Status.NodesState["1"].GracefulActionState.ActionStep)
			require.Equal(t, []int32{1}, nifi.disconnects)
		})
	}
}

func TestCancelPendingJoinChecksMembership(t *testing.T) {
	for _, tc := range []struct {
		name      string
		members   map[int32]string
		err       error
		wantState v1.State
		wantStep  v1.ActionStep
	}{
		{"absent", map[int32]string{}, nil, v1.GracefulDownscaleRunning, v1.CancelPendingPodAction},
		{"registered", map[int32]string{1: string(v1.ConnectStatus)}, nil, v1.GracefulDownscaleRunning, v1.DisconnectNodeAction},
		{"unreachable", nil, errors.New("connection refused"), v1.GracefulDownscaleRequired, v1.CancelPendingPodAction},
		{"HTTP 404", nil, nificlient.ErrNifiClusterReturned404, v1.GracefulDownscaleRequired, v1.CancelPendingPodAction},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cluster := scalingCluster()
			cluster.Spec.Nodes = cluster.Spec.Nodes[:1]
			cluster.Status.NodesState["1"] = v1.NodeState{GracefulActionState: v1.GracefulActionState{State: v1.GracefulDownscaleRequired, ActionStep: v1.CancelPendingPodAction}}
			k8s := scalingClient(t, cluster)
			useScalingNiFi(t, &scalingNiFi{membership: tc.members, membershipErr: tc.err})
			r := NifiClusterTaskReconciler{Client: k8s, Log: *zap.NewNop(), Recorder: record.NewFakeRecorder(8)}
			key := client.ObjectKeyFromObject(cluster)
			_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})
			require.NoError(t, err)
			require.NoError(t, k8s.Get(ctx, key, cluster))
			action := cluster.Status.NodesState["1"].GracefulActionState
			require.Equal(t, tc.wantState, action.State)
			require.Equal(t, tc.wantStep, action.ActionStep)
		})
	}
}

func TestAutoscalerCancelsOnlyUnscheduledJoiningCandidates(t *testing.T) {
	for _, nodeName := range []string{"", "worker"} {
		t.Run("assigned="+nodeName, func(t *testing.T) {
			ctx := context.Background()
			cluster := scalingCluster()
			cluster.Status.NodesState["1"] = v1.NodeState{GracefulActionState: v1.GracefulActionState{State: v1.GracefulUpscaleRunning, ActionStep: v1.ConnectNodeAction}}
			group := &v1alpha1.NifiNodeGroupAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: "scale", Namespace: cluster.Namespace},
				Spec: v1alpha1.NifiNodeGroupAutoscalerSpec{Replicas: 0, ClusterRef: v1.ClusterReference{Name: cluster.Name}, NodeLabelsSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"pool": "a"}}}}
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "joining", Namespace: cluster.Namespace,
				Labels: map[string]string{"app": "nifi", "nifi_cr": cluster.Name, "pool": "a", "nodeId": "1"}},
				Spec: corev1.PodSpec{NodeName: nodeName}, Status: corev1.PodStatus{Phase: corev1.PodPending}}
			k8s := scalingClient(t, cluster, group, pod)
			r := NifiNodeGroupAutoscalerReconciler{Client: k8s, APIReader: k8s, Log: *zap.NewNop(), Recorder: record.NewFakeRecorder(8)}
			_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(group)})
			require.NoError(t, err)
			require.NoError(t, k8s.Get(ctx, client.ObjectKeyFromObject(cluster), cluster))
			want := 2
			if nodeName == "" {
				want = 1
			}
			require.Len(t, cluster.Spec.Nodes, want)
			require.EqualValues(t, 0, cluster.Spec.Nodes[0].Id)
		})
	}
}
