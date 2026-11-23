package k8sutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	nifi "github.com/konpyutaika/nifikop/api/v1"
)

func TestErrorIs(t *testing.T) {
	err := &controllerutil.AlreadyOwnedError{
		Object: &v1.ObjectMeta{},
		Owner: v1.OwnerReference{
			Name: "test",
		},
	}

	is := IsAlreadyOwnedError(err)
	if !is {
		t.Errorf("Error is not AlreadyOwnedError")
	}
}

func TestStaleUpscaleCannotOverwriteDownscale(t *testing.T) {
	ctx := context.Background()
	cluster := &nifi.NifiCluster{ObjectMeta: v1.ObjectMeta{Name: "cluster", Namespace: "tenant"},
		Status: nifi.NifiClusterStatus{NodesState: map[string]nifi.NodeState{"3": {GracefulActionState: nifi.GracefulActionState{State: nifi.GracefulUpscaleRunning, ActionStep: nifi.ConnectStatus}}}}}
	s := runtime.NewScheme()
	require.NoError(t, nifi.AddToScheme(s))
	c := fake.NewClientBuilder().WithScheme(s).WithStatusSubresource(cluster).WithObjects(cluster).Build()
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(cluster), cluster))
	stale := cluster.DeepCopy()
	status := *cluster.Status.DeepCopy()
	require.NoError(t, UpdateNodeStatus(c, []string{"3"}, cluster, status, nifi.GracefulActionState{State: nifi.GracefulDownscaleRequired}, *zap.NewNop()))
	err := UpdateNodeStatus(c, []string{"3"}, stale, status, nifi.GracefulActionState{State: nifi.GracefulUpscaleSucceeded}, *zap.NewNop())
	require.True(t, apierrors.IsConflict(err))
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(cluster), cluster))
	require.Equal(t, nifi.GracefulDownscaleRequired, cluster.Status.NodesState["3"].GracefulActionState.State)
}
