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
	"strings"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/resources/templates"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	nifiutil "github.com/Orange-OpenSource/nifikop/pkg/util/nifi"
	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) service(id int32, log logr.Logger) runtimeClient.Object {

	usedPorts := generateServicePortForInternalListeners(r.NifiCluster.Spec.ListenersConfig.InternalListeners)

	return &corev1.Service{
		ObjectMeta: templates.ObjectMeta(nifiutil.ComputeNodeName(id, r.NifiCluster.Name),
			//fmt.Sprintf("%s-%d", r.NifiCluster.Name, id),
			util.MergeLabels(
				nifiutil.LabelsForNifi(r.NifiCluster.Name),
				map[string]string{"nodeId": fmt.Sprintf("%d", id)},
			),
			r.NifiCluster),
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeClusterIP,
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector:        util.MergeLabels(nifiutil.LabelsForNifi(r.NifiCluster.Name), map[string]string{"nodeId": fmt.Sprintf("%d", id)}),
			Ports:           usedPorts,
		},
	}
}

func (r *Reconciler) externalServices(log logr.Logger) []runtimeClient.Object {

	var services []runtimeClient.Object
	for _, eService := range r.NifiCluster.Spec.ExternalServices {

		annotations := &eService.ServiceAnnotations
		if err := mergo.Merge(annotations, r.NifiCluster.Spec.Service.Annotations); err != nil {
			log.Error(err, "error occurred during merging service annotations")
		}

		usedPorts := r.generateServicePortForExternalListeners(eService)
		services = append(services, &corev1.Service{
			ObjectMeta: templates.ObjectMetaWithAnnotations(eService.Name, nifiutil.LabelsForNifi(r.NifiCluster.Name),
				*annotations, r.NifiCluster),
			Spec: corev1.ServiceSpec{
				Type:                     eService.Spec.Type,
				SessionAffinity:          corev1.ServiceAffinityClientIP,
				Selector:                 nifiutil.LabelsForNifi(r.NifiCluster.Name),
				Ports:                    usedPorts,
				ClusterIP:                eService.Spec.ClusterIP,
				ExternalIPs:              eService.Spec.ExternalIPs,
				LoadBalancerIP:           eService.Spec.LoadBalancerIP,
				LoadBalancerSourceRanges: eService.Spec.LoadBalancerSourceRanges,
				ExternalName:             eService.Spec.ExternalName,
			},
		})
	}
	return services
}

//
func generateServicePortForInternalListeners(listeners []v1alpha1.InternalListenerConfig) []corev1.ServicePort {
	var usedPorts []corev1.ServicePort

	for _, iListeners := range listeners {
		usedPorts = append(usedPorts, corev1.ServicePort{
			Name:       strings.ReplaceAll(iListeners.Name, "_", ""),
			Port:       iListeners.ContainerPort,
			TargetPort: intstr.FromInt(int(iListeners.ContainerPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	return usedPorts
}

func (r *Reconciler) generateServicePortForExternalListeners(eService v1alpha1.ExternalServiceConfig) []corev1.ServicePort {
	var usedPorts []corev1.ServicePort

	for _, port := range eService.Spec.PortConfigs {
		for _, iListener := range r.NifiCluster.Spec.ListenersConfig.InternalListeners {
			if port.InternalListenerName == iListener.Name {
				usedPorts = append(usedPorts, corev1.ServicePort{
					Name:       strings.ReplaceAll(iListener.Name, "_", ""),
					Port:       port.Port,
					TargetPort: intstr.FromInt(int(iListener.ContainerPort)),
					Protocol:   corev1.ProtocolTCP,
				})
			}
		}
	}

	return usedPorts
}
