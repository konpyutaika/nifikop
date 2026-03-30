package nifi

import (
	"context"
	"testing"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources"
)

func TestReconcileNifiPodDoesNotDeleteNotReadyPodWhenPatchIsEmpty(t *testing.T) {
	t.Parallel()

	testScheme := runtime.NewScheme()
	require.NoError(t, scheme.AddToScheme(testScheme))
	require.NoError(t, v1.SchemeBuilder.AddToScheme(testScheme))

	currentPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-0-nodeabcde",
			Namespace: "namespace",
			Labels: map[string]string{
				"nifi_cr": "cluster",
				"nodeId":  "0",
			},
			Annotations: map[string]string{
				podServerCertHashAnnotation: "serverhash0000",
				podClientCertHashAnnotation: "clienthash0000",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nifi",
					Image: "apache/nifi:1.28.0",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "nifi",
					Ready: false,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					},
				},
			},
		},
	}
	desiredPod := currentPod.DeepCopy()

	require.NoError(t, patch.DefaultAnnotator.SetLastAppliedAnnotation(currentPod))
	require.NoError(t, patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPod))

	fakeClient := fake.NewClientBuilder().
		WithScheme(testScheme).
		WithObjects(currentPod.DeepCopy()).
		Build()

	cluster := &v1.NifiCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster",
			Namespace: "namespace",
		},
		Spec: v1.NifiClusterSpec{
			Nodes: []v1.Node{
				{Id: 0},
			},
		},
		Status: v1.NifiClusterStatus{
			NodesState: map[string]v1.NodeState{
				"0": {
					ConfigurationState: v1.ConfigOutOfSync,
					PodIsReady:         false,
					GracefulActionState: v1.GracefulActionState{
						State:      v1.GracefulUpscaleRunning,
						ActionStep: v1.ConnectStatus,
					},
				},
			},
		},
	}

	reconciler := &Reconciler{
		Reconciler: resources.Reconciler{
			Client:                   fakeClient,
			DirectClient:             fakeClient,
			NifiCluster:              cluster,
			NifiClusterCurrentStatus: cluster.Status,
		},
	}

	err, ready := reconciler.reconcileNifiPod(*zap.NewNop(), desiredPod)
	require.NoError(t, err)
	assert.False(t, ready)

	podList := &corev1.PodList{}
	require.NoError(t, fakeClient.List(context.Background(), podList, ctrlclient.InNamespace("namespace")))
	require.Len(t, podList.Items, 1)
	assert.Equal(t, currentPod.Name, podList.Items[0].Name)
}

func TestShouldDelayNoDiffConnectingPodRecycle(t *testing.T) {
	t.Parallel()

	cluster := &v1.NifiCluster{
		Status: v1.NifiClusterStatus{
			NodesState: map[string]v1.NodeState{
				"0": {
					GracefulActionState: v1.GracefulActionState{
						State:      v1.GracefulUpscaleRunning,
						ActionStep: v1.ConnectStatus,
					},
				},
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"nodeId": "0"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "nifi",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{ExitCode: 1},
					},
				},
			},
		},
	}

	delay, reason := shouldDelayNoDiffConnectingPodRecycle(cluster, pod)
	assert.True(t, delay)
	assert.Contains(t, reason, "still connecting")
}

func TestShouldNotDelayNoDiffConnectingPodRecycleForFailedPod(t *testing.T) {
	t.Parallel()

	cluster := &v1.NifiCluster{
		Status: v1.NifiClusterStatus{
			NodesState: map[string]v1.NodeState{
				"0": {
					GracefulActionState: v1.GracefulActionState{
						State:      v1.GracefulUpscaleRunning,
						ActionStep: v1.ConnectStatus,
					},
				},
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"nodeId": "0"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "nifi",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{ExitCode: 1},
					},
				},
			},
		},
	}

	delay, reason := shouldDelayNoDiffConnectingPodRecycle(cluster, pod)
	assert.False(t, delay)
	assert.Empty(t, reason)
}
