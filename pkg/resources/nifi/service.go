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
	"github.com/go-logr/logr"
	"github.com/Orange-OpenSource/nifikop/pkg/resources/templates"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

func (r *Reconciler) service(id int32, log logr.Logger) runtime.Object {

	usedPorts := r.generateServicePortForInternalListeners()

	usedPorts = append(usedPorts, r.generateServicePortForExternalListeners()...)
	usedPorts = append(usedPorts, r.generateDefaultServicePort()...)

	return &corev1.Service{
		ObjectMeta: templates.ObjectMeta(fmt.Sprintf("%s-%d", r.NifiCluster.Name, id),
			util.MergeLabels(
				LabelsForNifi(r.NifiCluster.Name),
				map[string]string{"nodeId": fmt.Sprintf("%d", id)},
			),
			r.NifiCluster),
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeClusterIP,
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector:        util.MergeLabels(LabelsForNifi(r.NifiCluster.Name), map[string]string{"nodeId": fmt.Sprintf("%d", id)}),
			Ports:           usedPorts,
		},
	}
}

//
func (r *Reconciler) generateServicePortForInternalListeners() []corev1.ServicePort {
	var usedPorts []corev1.ServicePort

	for _, iListeners := range r.NifiCluster.Spec.ListenersConfig.InternalListeners {
		usedPorts = append(usedPorts, corev1.ServicePort{
			Name:       strings.ReplaceAll(iListeners.Name, "_", ""),
			Port:       iListeners.ContainerPort,
			TargetPort: intstr.FromInt(int(iListeners.ContainerPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	return usedPorts
}

//
func (r *Reconciler) generateServicePortForExternalListeners() []corev1.ServicePort {
	var usedPorts []corev1.ServicePort

	/*for _, eListener := range r.NifiCluster.Spec.ListenersConfig.ExternalListeners {
		usedPorts = append(usedPorts, corev1.ServicePort{
			Name:       eListener.Name,
			Protocol:   corev1.ProtocolTCP,
			Port:       eListener.ContainerPort,
			TargetPort: intstr.FromInt(int(eListener.ContainerPort)),
		})
	}*/

	return usedPorts
}

//
func (r *Reconciler) generateDefaultServicePort() []corev1.ServicePort {

	usedPorts := []corev1.ServicePort{
		// Prometheus metrics port for monitoring
		/*{
			Name:       "metrics",
			Protocol:   corev1.ProtocolTCP,
			Port:       v1alpha1.MetricsPort,
			TargetPort: intstr.FromInt(v1alpha1.MetricsPort),
		},*/
	}

	return usedPorts
}
