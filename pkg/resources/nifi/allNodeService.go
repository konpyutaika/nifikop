package nifi

import (
	"fmt"
	"github.com/erdrix/nifikop/pkg/resources/templates"
	nifiutils "github.com/erdrix/nifikop/pkg/util/nifi"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) allNodeService() runtime.Object {

	usedPorts := r.generateServicePortForInternalListeners()
	usedPorts = append(usedPorts, r.generateDefaultServicePort()...)

	return &corev1.Service{
		ObjectMeta: templates.ObjectMeta(fmt.Sprintf(nifiutils.AllNodeServiceTemplate, r.NifiCluster.Name), LabelsForNifi(r.NifiCluster.Name), r.NifiCluster),
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeClusterIP,
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector:        LabelsForNifi(r.NifiCluster.Name),
			Ports:           usedPorts,
		},
	}
}