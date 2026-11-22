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
	user      *nigoapi.UserEntity
	grants    map[string][]nigoapi.TenantEntity
	policyErr error
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
func (n *scalingNiFi) GetClusterNode(int32) (*nigoapi.NodeEntity, error) {
	return &nigoapi.NodeEntity{Node: &nigoapi.NodeDto{Status: string(v1.ConnectStatus)}}, nil
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
