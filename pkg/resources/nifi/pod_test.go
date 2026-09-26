package nifi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources"
)

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

func TestPodHostUsers(t *testing.T) {
	falseVal := false
	trueVal := true

	tests := []struct {
		name      string
		hostUsers *bool
	}{
		{name: "unset leaves hostUsers absent", hostUsers: nil},
		{name: "false is propagated", hostUsers: &falseVal},
		{name: "true is propagated", hostUsers: &trueVal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newPodTestReconciler()
			nodeConfig := &v1.NodeConfig{HostUsers: tt.hostUsers}
			node := v1.Node{Id: 0}

			obj := r.pod(node, nodeConfig, []corev1.PersistentVolumeClaim{}, zap.Logger{})
			pod, ok := obj.(*corev1.Pod)
			assert.True(t, ok, "pod() should return a *corev1.Pod")

			if tt.hostUsers == nil {
				assert.Nil(t, pod.Spec.HostUsers)
			} else {
				assert.NotNil(t, pod.Spec.HostUsers)
				assert.Equal(t, *tt.hostUsers, *pod.Spec.HostUsers)
			}
		})
	}
}
