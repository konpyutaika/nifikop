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
	"github.com/Orange-OpenSource/nifikop/pkg/resources/templates"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	nifiutils "github.com/Orange-OpenSource/nifikop/pkg/util/nifi"
	corev1 "k8s.io/api/core/v1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) headlessService() runtimeClient.Object {

	// InternalListeners ports
	usedPorts := generateServicePortForInternalListeners(r.NifiCluster.Spec.ListenersConfig.InternalListeners)

	return &corev1.Service{
		ObjectMeta: templates.ObjectMetaWithAnnotations(
			fmt.Sprintf(nifiutils.HeadlessServiceTemplate, r.NifiCluster.Name),
			util.MergeLabels(nifiutils.LabelsForNifi(r.NifiCluster.Name), r.NifiCluster.Labels),
			r.NifiCluster.Spec.Service.Annotations,
			r.NifiCluster,
		),
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeClusterIP,
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector:        nifiutils.LabelsForNifi(r.NifiCluster.Name),
			Ports:           usedPorts,
			ClusterIP:       corev1.ClusterIPNone,
		},
	}
}
