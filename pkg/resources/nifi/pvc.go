package nifi

import (
	"context"
	"fmt"
	"sort"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	"github.com/konpyutaika/nifikop/pkg/util"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

func (r *Reconciler) pvc(id int32, storage v1.StorageConfig, log zap.Logger) runtime.Object {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: templates.ObjectMetaWithGeneratedNameAndAnnotations(
			// name
			fmt.Sprintf(templates.NodeStorageTemplate, r.NifiCluster.Name, id, storage.Name),
			// labels
			util.MergeLabels(
				nifiutil.LabelsForNifi(r.NifiCluster.Name),
				storage.Metadata.Labels,
				map[string]string{
					"nodeId":                            fmt.Sprintf("%d", id),
					"storageName":                       storage.Name,
					nifiutil.NifiVolumeReclaimPolicyKey: string(storage.ReclaimPolicy),
					nifiutil.NifiDataVolumeMountKey:     "true",
				},
			),
			// annotations
			util.MergeAnnotations(
				storage.Metadata.Annotations,
				map[string]string{"mountPath": storage.MountPath, "storageName": storage.Name},
			),
			r.NifiCluster),
		Spec: *storage.PVCSpec,
	}
}

// returns true and the PVC if it exists. Else false and nil, respectively.
func (r *Reconciler) storageConfigPVCExists(pvcs []corev1.PersistentVolumeClaim, storageName string) (bool, *corev1.PersistentVolumeClaim) {
	for _, pvc := range pvcs {
		if pvc.Annotations["storageName"] == storageName {
			return true, &pvc
		}
	}
	return false, nil
}

// normalizePVCDesiredForReconcile aligns identity metadata on desired with the live PVC so
// objectmatcher comparisons are stable (GenerateName vs Name).
func normalizePVCDesiredForReconcile(current, desired *corev1.PersistentVolumeClaim) {
	desired.Name = current.Name
	desired.Namespace = current.Namespace
	desired.GenerateName = ""
	desired.UID = ""
	desired.ResourceVersion = ""
	desired.CreationTimestamp = metav1.Time{}
}

// preparePVCForDiff returns a copy of desired with the live last-applied annotation, so
// objectmatcher does not treat operator bookkeeping as perpetual drift.
func preparePVCForDiff(current, desired *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	forDiff := desired.DeepCopy()
	if current.Annotations == nil {
		return forDiff
	}
	lastApplied, ok := current.Annotations[patch.LastAppliedConfig]
	if !ok {
		return forDiff
	}
	if forDiff.Annotations == nil {
		forDiff.Annotations = map[string]string{}
	}
	forDiff.Annotations[patch.LastAppliedConfig] = lastApplied
	return forDiff
}

// pvcNeedsSpecUpdate compares reconcilable PVC spec fields against the live object.
// This avoids false "in sync" results from banzaicloud three-way merge when
// last-applied records fields the apiserver did not persist on spec.
func pvcNeedsSpecUpdate(current, desired *corev1.PersistentVolumeClaim) bool {
	if !stringPtrEqual(current.Spec.StorageClassName, desired.Spec.StorageClassName) {
		return true
	}
	if !stringPtrEqual(current.Spec.VolumeAttributesClassName, desired.Spec.VolumeAttributesClassName) {
		return true
	}
	if !equality.Semantic.DeepEqual(normalizeAccessModes(current.Spec.AccessModes), normalizeAccessModes(desired.Spec.AccessModes)) {
		return true
	}
	if desired.Spec.Resources.Requests != nil {
		desiredQty := desired.Spec.Resources.Requests[corev1.ResourceStorage]
		currentQty := current.Spec.Resources.Requests[corev1.ResourceStorage]
		if desiredQty.Cmp(currentQty) != 0 {
			return true
		}
	}
	if desired.Spec.VolumeMode != nil && !equality.Semantic.DeepEqual(current.Spec.VolumeMode, desired.Spec.VolumeMode) {
		return true
	}
	return false
}

func normalizeAccessModes(modes []corev1.PersistentVolumeAccessMode) []corev1.PersistentVolumeAccessMode {
	if len(modes) == 0 {
		return modes
	}
	out := append([]corev1.PersistentVolumeAccessMode(nil), modes...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func stringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func pvcAnnotationsForUpdate(current, desired map[string]string) map[string]string {
	merged := util.MergeAnnotations(current, desired)
	delete(merged, patch.LastAppliedConfig)
	return merged
}

// pvcForUpdate merges the desired PVC spec and metadata onto the live object while
// preserving cluster-assigned spec fields such as volumeName.
func pvcForUpdate(current, desired *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	updated := current.DeepCopy()
	updated.Labels = desired.Labels
	updated.Annotations = pvcAnnotationsForUpdate(current.Annotations, desired.Annotations)
	updated.Spec = *desired.Spec.DeepCopy()
	updated.Spec.VolumeName = current.Spec.VolumeName
	return updated
}

// syncPVCLastAppliedAnnotation stores last-applied from the live PVC spec in etcd, not from intent alone.
func syncPVCLastAppliedAnnotation(c runtimeClient.Client, namespace, name string) error {
	live := &corev1.PersistentVolumeClaim{}
	if err := c.Get(context.TODO(), runtimeClient.ObjectKey{Namespace: namespace, Name: name}, live); err != nil {
		return err
	}
	if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(live); err != nil {
		return err
	}
	return c.Update(context.TODO(), live)
}
