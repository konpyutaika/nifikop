package nifi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
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

func TestPVCNeedsSpecUpdateVolumeAttributesClassName(t *testing.T) {
	vac := "rt-content-4000-default"
	gp3 := "gp3"
	current := &corev1.PersistentVolumeClaim{
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &gp3,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOncePod},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: mustParseQuantity("100Gi"),
				},
			},
		},
	}
	desired := current.DeepCopy()
	desired.Spec.VolumeAttributesClassName = &vac

	assert.False(t, pvcNeedsSpecUpdate(current, current))
	assert.True(t, pvcNeedsSpecUpdate(current, desired))

	currentWithVAC := desired.DeepCopy()
	desiredWithoutVAC := current.DeepCopy()
	assert.True(t, pvcNeedsSpecUpdate(currentWithVAC, desiredWithoutVAC))
	assert.False(t, pvcNeedsSpecUpdate(desiredWithoutVAC, desiredWithoutVAC))

	empty := ""
	currentWithEmptyVAC := current.DeepCopy()
	currentWithEmptyVAC.Spec.VolumeAttributesClassName = &empty
	assert.False(t, pvcNeedsSpecUpdate(currentWithEmptyVAC, desiredWithoutVAC))
}

func TestPVCForUpdateStripsLastAppliedAnnotation(t *testing.T) {
	current := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				patch.LastAppliedConfig: "stale",
				"mountPath":             "/opt/nifi/content_repository",
			},
		},
	}
	desired := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				patch.LastAppliedConfig: "should-not-copy",
				"storageName":           "content-repository",
			},
		},
	}

	updated := pvcForUpdate(current, desired)
	_, hasLastApplied := updated.Annotations[patch.LastAppliedConfig]
	assert.False(t, hasLastApplied)
	assert.Equal(t, "content-repository", updated.Annotations["storageName"])
	assert.Equal(t, "/opt/nifi/content_repository", updated.Annotations["mountPath"])
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

func TestPVCForUpdateRemovesVolumeAttributesClassName(t *testing.T) {
	vac := "gold"
	current := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-pvc",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName:                "pv-bound",
			VolumeAttributesClassName: &vac,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: mustParseQuantity("10Gi"),
				},
			},
		},
	}
	desired := &corev1.PersistentVolumeClaim{
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: mustParseQuantity("10Gi"),
				},
			},
		},
	}

	updated := pvcForUpdate(current, desired)
	assert.Equal(t, "pv-bound", updated.Spec.VolumeName)
	requireVolumeAttributesClassNameCleared(t, updated.Spec.VolumeAttributesClassName)

	payload, err := json.Marshal(updated)
	require.NoError(t, err)
	assert.Contains(t, string(payload), `"volumeAttributesClassName":""`)
}

func requireVolumeAttributesClassNameCleared(t *testing.T, name *string) {
	t.Helper()
	if name == nil {
		return
	}
	assert.Empty(t, *name)
}

func mustParseQuantity(s string) resource.Quantity {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		panic(err)
	}
	return q
}
