package nifi

import (
	corev1 "k8s.io/api/core/v1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	nifiutils "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

func (r *Reconciler) allNodeService() runtimeClient.Object {
	usedPorts := generateServicePortForInternalListeners(r.NifiCluster.Spec.ListenersConfig.InternalListeners)

	return &corev1.Service{
		ObjectMeta: templates.ObjectMetaWithAnnotations(
			r.NifiCluster.GetNodeServiceName(),
			nifiutils.LabelsForNifi(r.NifiCluster.Name),
			r.NifiCluster.Spec.Service.Annotations,
			r.NifiCluster),
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeClusterIP,
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector:        nifiutils.LabelsForNifi(r.NifiCluster.Name),
			Ports:           usedPorts,
		},
	}
}
