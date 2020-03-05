package nifi

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/erdrix/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/erdrix/nifikop/pkg/resources/templates"
	"github.com/erdrix/nifikop/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) pvc(id int32, storage v1alpha1.StorageConfig, log logr.Logger) runtime.Object {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: templates.ObjectMetaWithGeneratedNameAndAnnotations(
			fmt.Sprintf(templates.NodeStorageTemplate, r.NifiCluster.Name, id),
			util.MergeLabels(
				LabelsForNifi(r.NifiCluster.Name),
				map[string]string{"nodeId": fmt.Sprintf("%d", id)},
			),
			map[string]string{"mountPath": storage.MountPath, "storageName": storage.Name}, r.NifiCluster),
		Spec: *storage.PVCSpec,
	}
}