package nifi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources"
)

func TestPVC(t *testing.T) {
	r := resources.Reconciler{
		NifiCluster: &v1.NifiCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster",
				Namespace: "namespace",
			},
			Spec: v1.NifiClusterSpec{},
		},
	}
	rec := Reconciler{
		Reconciler: r,
	}
	storage := v1.StorageConfig{
		Name:      "storage",
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

func TestStorageConfigPVCExists(t *testing.T) {
	r := resources.Reconciler{
		NifiCluster: &v1.NifiCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster",
				Namespace: "namespace",
			},
			Spec: v1.NifiClusterSpec{},
		},
	}
	rec := Reconciler{
		Reconciler: r,
	}

	pvcList := []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pvc1",
				Annotations: map[string]string{
					"storageName": "FOO",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pvc2",
				Annotations: map[string]string{
					"storageName": "TARGET",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pvc3",
				Annotations: map[string]string{
					"storageName": "BAR",
				},
			},
		},
	}

	exists, pvc := rec.storageConfigPVCExists(pvcList, "TARGET")
	assert.True(t, exists)
	assert.Equal(t, pvc, &pvcList[1])

	exists, pvc = rec.storageConfigPVCExists(pvcList, "NONEXISTENT")
	assert.False(t, exists)
	assert.Nil(t, pvc)
}

func TestPVCForUpdatePreservesVolumeNameAndAppliesSpec(t *testing.T) {
	vac := "gold"
	current := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-pvc",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName: "pv-bound",
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: mustParseQuantity("10Gi"),
				},
			},
		},
	}
	desired := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      map[string]string{"nodeId": "0"},
			Annotations: map[string]string{"storageName": "data"},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeAttributesClassName: &vac,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: mustParseQuantity("20Gi"),
				},
			},
		},
	}

	updated := pvcForUpdate(current, desired)
	assert.Equal(t, "pv-bound", updated.Spec.VolumeName)
	assert.Equal(t, &vac, updated.Spec.VolumeAttributesClassName)
	assert.Equal(t, mustParseQuantity("20Gi"), updated.Spec.Resources.Requests[corev1.ResourceStorage])
	assert.Equal(t, "0", updated.Labels["nodeId"])
}

func mustParseQuantity(s string) resource.Quantity {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		panic(err)
	}
	return q
}
