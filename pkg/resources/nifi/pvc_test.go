package nifi

import (
	"testing"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
)

func TestPVC(t *testing.T) {
	r := resources.Reconciler{
		NifiCluster: &v1.NifiCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster",
				Namespace: "namespace",
			},
			Spec: v1.NifiClusterSpec{
			},
		},
	}
	rec := Reconciler{
		Reconciler: r,
	}
	storage := v1.StorageConfig{
		Name: "storage",
		MountPath: "/path",
		Metadata: v1.Metadata{
			Labels: map[string]string{
				"label": "value",
			},
			Annotations: map[string]string{
				"annotation": "value",
			},
		},
		PVCSpec: &corev1.PersistentVolumeClaimSpec{},
	}
	pvc := rec.pvc(0, storage, *zap.NewNop())
	p := pvc.(*corev1.PersistentVolumeClaim)

	// ensure the PVC has the specified metadata
	assert.Equal(t, "value", p.ObjectMeta.Annotations["annotation"])
	assert.Equal(t, "value", p.ObjectMeta.Labels["label"])
}