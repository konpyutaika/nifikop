package nifi

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

var reconcilePVCScheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(reconcilePVCScheme))
	utilruntime.Must(v1.AddToScheme(reconcilePVCScheme))
}

func testClusterForPVCReconcile() *v1.NifiCluster {
	return &v1.NifiCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "nifi.konpyutaika.com/v1",
			Kind:       "NifiCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster",
			Namespace: "namespace",
			UID:       "cluster-uid",
		},
	}
}

func storageConfigWithSize(size string) v1.StorageConfig {
	return v1.StorageConfig{
		Name:          "data",
		MountPath:     "/opt/nifi/nifi-current/logs",
		ReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
		PVCSpec: &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}
}

func existingNodePVC(cluster *v1.NifiCluster, nodeID int32, storage v1.StorageConfig, size string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d-%s-storage", cluster.Name, nodeID, storage.Name),
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"app":                               "nifi",
				"nifi_cr":                           cluster.Name,
				"nodeId":                            fmt.Sprintf("%d", nodeID),
				"storageName":                       storage.Name,
				nifiutil.NifiVolumeReclaimPolicyKey: string(storage.ReclaimPolicy),
				nifiutil.NifiDataVolumeMountKey:     "true",
			},
			Annotations: map[string]string{
				"mountPath":   storage.MountPath,
				"storageName": storage.Name,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}
}

func newReconcilerForPVC(cluster *v1.NifiCluster, pvcs ...client.Object) (*Reconciler, client.Client) {
	objs := []client.Object{cluster}
	objs = append(objs, pvcs...)
	c := fake.NewClientBuilder().WithScheme(reconcilePVCScheme).WithObjects(objs...).Build()
	rec := &Reconciler{
		Reconciler: resources.Reconciler{
			Client:      c,
			NifiCluster: cluster,
		},
	}
	return rec, c
}

func pvcStorageRequest(t *testing.T, c client.Client, name string) resource.Quantity {
	pvc := &corev1.PersistentVolumeClaim{}
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Namespace: "namespace", Name: name}, pvc))
	return pvc.Spec.Resources.Requests[corev1.ResourceStorage]
}

func TestReconcileNifiPVC_IncreasesStorageFromNifiClusterSpec(t *testing.T) {
	cluster := testClusterForPVCReconcile()
	storage := storageConfigWithSize("20Gi")
	existing := existingNodePVC(cluster, 0, storage, "10Gi")

	rec, c := newReconcilerForPVC(cluster, existing)
	desired := rec.pvc(0, storage, *zap.NewNop()).(*corev1.PersistentVolumeClaim)

	err := rec.reconcileNifiPVC(0, *zap.NewNop(), desired)
	require.NoError(t, err)

	want := resource.MustParse("20Gi")
	assert.Equal(t, 0, want.Cmp(pvcStorageRequest(t, c, existing.Name)))
}

func TestReconcileNifiPVC_KeepsLargerManualPVCWhenSpecUnchanged(t *testing.T) {
	cluster := testClusterForPVCReconcile()
	storage := storageConfigWithSize("10Gi")
	existing := existingNodePVC(cluster, 0, storage, "20Gi")

	rec, c := newReconcilerForPVC(cluster, existing)
	desired := rec.pvc(0, storage, *zap.NewNop()).(*corev1.PersistentVolumeClaim)

	err := rec.reconcileNifiPVC(0, *zap.NewNop(), desired)
	require.NoError(t, err)

	want := resource.MustParse("20Gi")
	assert.Equal(t, 0, want.Cmp(pvcStorageRequest(t, c, existing.Name)))

	err = rec.reconcileNifiPVC(0, *zap.NewNop(), desired)
	require.NoError(t, err)

	assert.Equal(t, 0, want.Cmp(pvcStorageRequest(t, c, existing.Name)))
}
