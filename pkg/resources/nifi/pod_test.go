package nifi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources"
)

func ptrTo[T any](v T) *T { return &v }

func newPodTestReconciler() *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			NifiCluster: &v1.NifiCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cluster",
					Namespace: "namespace",
				},
				Spec: v1.NifiClusterSpec{
					ListenersConfig: &v1.ListenersConfig{
						InternalListeners: []v1.InternalListenerConfig{
							{Type: v1.HttpListenerType, Name: "http", ContainerPort: 8080},
							{Type: v1.ClusterListenerType, Name: "cluster", ContainerPort: 6007},
						},
					},
				},
			},
		},
	}
}

func TestPodWithoutOverridesIsUnchanged(t *testing.T) {
	r := newPodTestReconciler()
	node := v1.Node{Id: 0}

	obj, err := r.pod(node, &v1.NodeConfig{}, []corev1.PersistentVolumeClaim{}, zap.Logger{})
	require.NoError(t, err)
	pod, ok := obj.(*corev1.Pod)
	require.True(t, ok)
	assert.Nil(t, pod.Spec.HostUsers)
	assert.Equal(t, int64(1000), *pod.Spec.SecurityContext.RunAsUser)
	assert.Equal(t, corev1.RestartPolicyNever, pod.Spec.RestartPolicy)
}

func TestPodOverridesAreAppliedAfterProxyFields(t *testing.T) {
	r := newPodTestReconciler()
	node := v1.Node{Id: 0}
	nodeConfig := &v1.NodeConfig{
		RunAsUser:         ptrTo[int64](1000),
		PriorityClassName: ptrTo("from-proxy"),
		PodOverrides: &corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{"example.com/foo": "bar"},
			},
			Spec: corev1.PodSpec{
				HostUsers:         ptrTo(false),
				PriorityClassName: "from-override",
				SecurityContext:   &corev1.PodSecurityContext{RunAsUser: ptrTo[int64](1000650000)},
				Hostname:          "ignored",
				Containers: []corev1.Container{
					{Name: ContainerName, SecurityContext: &corev1.SecurityContext{ReadOnlyRootFilesystem: ptrTo(true)}},
					{Name: "sidecar", Image: "example/sidecar:1"},
				},
			},
		},
	}

	obj, err := r.pod(node, nodeConfig, []corev1.PersistentVolumeClaim{}, zap.Logger{})
	require.NoError(t, err)
	pod := obj.(*corev1.Pod)

	assert.False(t, *pod.Spec.HostUsers)
	assert.Equal(t, "from-override", pod.Spec.PriorityClassName)
	assert.Equal(t, int64(1000650000), *pod.Spec.SecurityContext.RunAsUser)
	assert.Equal(t, "bar", pod.Annotations["example.com/foo"])
	assert.Equal(t, "cluster-0-node", pod.Spec.Hostname)

	require.Len(t, pod.Spec.Containers, 2)
	nifi := pod.Spec.Containers[0]
	assert.Equal(t, ContainerName, nifi.Name)
	assert.NotEmpty(t, nifi.Command, "operator-generated command must survive the merge")
	assert.NotEmpty(t, nifi.Ports)
	assert.True(t, *nifi.SecurityContext.ReadOnlyRootFilesystem)
	assert.Equal(t, "sidecar", pod.Spec.Containers[1].Name)
}

func TestPodOverridesErrorIsReturned(t *testing.T) {
	r := newPodTestReconciler()
	nodeConfig := &v1.NodeConfig{
		PodOverrides: &corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Image: "unnamed"}}},
		},
	}

	obj, err := r.pod(v1.Node{Id: 0}, nodeConfig, []corev1.PersistentVolumeClaim{}, zap.Logger{})
	require.Error(t, err)
	assert.Nil(t, obj)
}
