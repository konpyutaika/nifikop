package nifi

import (
	"github.com/erdrix/nifikop/pkg/resources/templates"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TODO: To remove ? Or to redo
func (r *Reconciler) lbService() runtime.Object {

	usedPorts := r.generateServicePortForInternalListeners()

	usedPorts = append(usedPorts, r.generateServicePortForExternalListeners()...)
	usedPorts = append(usedPorts, r.generateDefaultServicePort()...)

	return &corev1.Service{
		ObjectMeta: templates.ObjectMeta(r.NifiCluster.Name, LabelsForNifi(r.NifiCluster.Name), r.NifiCluster),
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeLoadBalancer,
			SessionAffinity: corev1.ServiceAffinityClientIP,
			Selector:        LabelsForNifi(r.NifiCluster.Name),
			Ports:           usedPorts,
		},
	}
}