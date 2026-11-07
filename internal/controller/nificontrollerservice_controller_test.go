package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/controllerservice"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

const (
	csTestNamespace = "nifikop"
	csTestCluster   = "nc"
	csTestName      = "web-client"
	csTestId        = "16cfd2ec-0174-1000-0000-00004b9b35cc"
)

// csFixture wires a NifiControllerService reconciler to a fake Kubernetes client, a NifiCluster
// using the mock client config, and stubbed NiFi operations.
type csFixture struct {
	t          *testing.T
	client     client.Client
	recorder   *record.FakeRecorder
	reconciler *NifiControllerServiceReconciler
	ops        *controllerServiceOps
	calls      []string
}

func newCsFixture(t *testing.T, cs *v1alpha1.NifiControllerService) *csFixture {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, v1.AddToScheme(scheme))
	require.NoError(t, v1alpha1.AddToScheme(scheme))

	cluster := &v1.NifiCluster{
		ObjectMeta: metav1.ObjectMeta{Name: csTestCluster, Namespace: csTestNamespace},
		Spec:       v1.NifiClusterSpec{Type: v1.ExternalCluster, ClientType: config.MockClientConfig},
	}

	f := &csFixture{t: t, recorder: record.NewFakeRecorder(100)}
	f.client = fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cluster, cs).
		WithStatusSubresource(&v1alpha1.NifiControllerService{}).
		Build()
	f.ops = &controllerServiceOps{
		exist: func(*v1alpha1.NifiControllerService, *clientconfig.NifiConfig) (bool, error) {
			f.calls = append(f.calls, "exist")
			return false, nil
		},
		create: func(*v1alpha1.NifiControllerService, *clientconfig.NifiConfig) (*v1alpha1.NifiControllerServiceStatus, error) {
			f.calls = append(f.calls, "create")
			return &v1alpha1.NifiControllerServiceStatus{Id: csTestId, Version: 0, State: v1alpha1.ControllerServiceStateDisabled}, nil
		},
		sync: func(*v1alpha1.NifiControllerService, *clientconfig.NifiConfig) (*v1alpha1.NifiControllerServiceStatus, error) {
			f.calls = append(f.calls, "sync")
			return &v1alpha1.NifiControllerServiceStatus{Id: csTestId, Version: 1, State: v1alpha1.ControllerServiceStateEnabled}, nil
		},
		remove: func(*v1alpha1.NifiControllerService, *clientconfig.NifiConfig) error {
			f.calls = append(f.calls, "remove")
			return nil
		},
	}
	f.reconciler = &NifiControllerServiceReconciler{
		Client:          f.client,
		Log:             *zap.NewNop(),
		Scheme:          scheme,
		Recorder:        f.recorder,
		RequeueInterval: 15,
		RequeueOffset:   0,
		ops:             f.ops,
	}
	return f
}

func (f *csFixture) reconcile() (ctrl.Result, error) {
	f.t.Helper()
	return f.reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: csTestName, Namespace: csTestNamespace},
	})
}

func (f *csFixture) get() (*v1alpha1.NifiControllerService, error) {
	cs := &v1alpha1.NifiControllerService{}
	err := f.client.Get(context.Background(), types.NamespacedName{Name: csTestName, Namespace: csTestNamespace}, cs)
	return cs, err
}

// events drains the recorder and returns the event reasons seen so far.
func (f *csFixture) events() []string {
	var reasons []string
	for {
		select {
		case e := <-f.recorder.Events:
			// FakeRecorder formats events as "<type> <reason> <message>".
			reasons = append(reasons, strings.SplitN(e, " ", 3)[1])
		default:
			return reasons
		}
	}
}

func newCs() *v1alpha1.NifiControllerService {
	return &v1alpha1.NifiControllerService{
		ObjectMeta: metav1.ObjectMeta{Name: csTestName, Namespace: csTestNamespace},
		Spec: v1alpha1.NifiControllerServiceSpec{
			ClusterRef: v1.ClusterReference{Name: csTestCluster, Namespace: csTestNamespace},
			Type:       v1alpha1.StandardWebClientServiceProviderType,
			StandardWebClientServiceProviderSpec: &v1alpha1.StandardWebClientServiceProviderSpec{
				ConnectTimeout: "10 secs", ReadTimeout: "10 secs", WriteTimeout: "10 secs",
			},
		},
	}
}

func TestControllerServiceReconcile_CreateThenWaitForEnabled(t *testing.T) {
	f := newCsFixture(t, newCs())
	f.ops.sync = func(*v1alpha1.NifiControllerService, *clientconfig.NifiConfig) (*v1alpha1.NifiControllerServiceStatus, error) {
		f.calls = append(f.calls, "sync")
		return &v1alpha1.NifiControllerServiceStatus{Id: csTestId, Version: 1, State: v1alpha1.ControllerServiceStateEnabling}, nil
	}

	result, err := f.reconcile()
	require.NoError(t, err)

	assert.Equal(t, []string{"exist", "create", "sync"}, f.calls)
	// A service that is not ENABLED yet is revisited quickly, not at the regular interval.
	assert.Equal(t, controllerServiceTransitionRequeue, result.RequeueAfter)

	cs, err := f.get()
	require.NoError(t, err)
	assert.Equal(t, csTestId, cs.Status.Id)
	assert.Equal(t, int64(1), cs.Status.Version)
	assert.Equal(t, v1alpha1.ControllerServiceStateEnabling, cs.Status.State)
	assert.Contains(t, cs.GetFinalizers(), controllerServiceFinalizer)
	assert.Equal(t, csTestCluster, cs.GetLabels()[ClusterRefLabel])
	assert.Contains(t, f.events(), "RunStateTransition")
}

func TestControllerServiceReconcile_SteadyState(t *testing.T) {
	cs := newCs()
	cs.Status = v1alpha1.NifiControllerServiceStatus{Id: csTestId, Version: 1, State: v1alpha1.ControllerServiceStateEnabled}
	f := newCsFixture(t, cs)
	f.ops.exist = func(*v1alpha1.NifiControllerService, *clientconfig.NifiConfig) (bool, error) {
		f.calls = append(f.calls, "exist")
		return true, nil
	}

	result, err := f.reconcile()
	require.NoError(t, err)

	assert.Equal(t, []string{"exist", "sync"}, f.calls, "an existing service is synced, never recreated")
	assert.Equal(t, 15*time.Second, result.RequeueAfter)
	assert.Contains(t, f.events(), "Reconciled")
}

func TestControllerServiceReconcile_SyncFailure(t *testing.T) {
	f := newCsFixture(t, newCs())
	f.ops.sync = func(*v1alpha1.NifiControllerService, *clientconfig.NifiConfig) (*v1alpha1.NifiControllerServiceStatus, error) {
		return nil, errors.New("nifi unavailable")
	}

	_, err := f.reconcile()
	require.Error(t, err)

	// The create step already ran and its id must have been persisted before the failure.
	cs, err := f.get()
	require.NoError(t, err)
	assert.Equal(t, csTestId, cs.Status.Id)
	assert.Contains(t, f.events(), "SynchronizingFailed")
}

func deletingCs() *v1alpha1.NifiControllerService {
	cs := newCs()
	cs.SetFinalizers([]string{controllerServiceFinalizer})
	cs.Status = v1alpha1.NifiControllerServiceStatus{Id: csTestId, Version: 1, State: v1alpha1.ControllerServiceStateEnabled}
	return cs
}

func markForDeletion(t *testing.T, f *csFixture) {
	t.Helper()
	cs, err := f.get()
	require.NoError(t, err)
	// With a finalizer present the fake client sets the deletion timestamp instead of deleting.
	require.NoError(t, f.client.Delete(context.Background(), cs))
	cs, err = f.get()
	require.NoError(t, err)
	require.NotNil(t, cs.DeletionTimestamp)
}

func TestControllerServiceReconcile_DeleteRefusedWhileReferenced(t *testing.T) {
	f := newCsFixture(t, deletingCs())
	markForDeletion(t, f)
	f.ops.remove = func(*v1alpha1.NifiControllerService, *clientconfig.NifiConfig) error {
		return fmt.Errorf("%w: bitbucket-flows (FlowRegistryClient)", controllerservice.ErrReferenced)
	}

	result, err := f.reconcile()
	require.NoError(t, err, "a referenced service is an expected condition, not an error")
	assert.Equal(t, 15*time.Second, result.RequeueAfter)

	cs, err := f.get()
	require.NoError(t, err)
	assert.Contains(t, cs.GetFinalizers(), controllerServiceFinalizer, "the finalizer must stay until the references are gone")
	assert.Contains(t, f.events(), "StillReferenced")
}

func TestControllerServiceReconcile_DeleteWaitsForDisable(t *testing.T) {
	f := newCsFixture(t, deletingCs())
	markForDeletion(t, f)
	f.ops.remove = func(*v1alpha1.NifiControllerService, *clientconfig.NifiConfig) error {
		return controllerservice.ErrTransitioning
	}

	result, err := f.reconcile()
	require.NoError(t, err)
	assert.Equal(t, controllerServiceTransitionRequeue, result.RequeueAfter)

	cs, err := f.get()
	require.NoError(t, err)
	assert.Contains(t, cs.GetFinalizers(), controllerServiceFinalizer)
	assert.Contains(t, f.events(), "RunStateTransition")
}

func TestControllerServiceReconcile_DeleteRemovesFinalizer(t *testing.T) {
	f := newCsFixture(t, deletingCs())
	markForDeletion(t, f)

	result, err := f.reconcile()
	require.NoError(t, err)
	assert.Equal(t, []string{"remove"}, f.calls)
	assert.False(t, result.Requeue)
	assert.Zero(t, result.RequeueAfter)

	// Once the finalizer is gone the fake client completes the deletion.
	_, err = f.get()
	assert.True(t, apierrors.IsNotFound(err), "expected the resource to be gone, got: %v", err)
}

func TestControllerServiceReconcile_RemoveFailureIsAnError(t *testing.T) {
	f := newCsFixture(t, deletingCs())
	markForDeletion(t, f)
	f.ops.remove = func(*v1alpha1.NifiControllerService, *clientconfig.NifiConfig) error {
		return errors.New("nifi unavailable")
	}

	_, err := f.reconcile()
	require.Error(t, err)

	cs, err := f.get()
	require.NoError(t, err)
	assert.Contains(t, cs.GetFinalizers(), controllerServiceFinalizer)
}
