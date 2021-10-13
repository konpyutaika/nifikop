// Copyright 2020 Orange SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package apis

package nifi

import (
	"fmt"
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/resources/templates"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	nifiutil "github.com/Orange-OpenSource/nifikop/pkg/util/nifi"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) pvc(id int32, storage v1alpha1.StorageConfig, log logr.Logger) runtime.Object {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: templates.ObjectMetaWithGeneratedNameAndAnnotations(
			fmt.Sprintf(templates.NodeStorageTemplate, r.NifiCluster.Name, id),
			util.MergeLabels(
				nifiutil.LabelsForNifi(r.NifiCluster.Name),
				map[string]string{"nodeId": fmt.Sprintf("%d", id)},
			),
			map[string]string{"mountPath": storage.MountPath, "storageName": storage.Name}, r.NifiCluster),
		Spec: *storage.PVCSpec,
	}
}
