package nifi

import (
	"fmt"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/resources/templates"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/util"
	nifiutils "gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/util/nifi"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) headlessService() runtime.Object {

	// InternalListeners ports
	usedPorts :=  r.generateServicePortForInternalListeners()

	// Additionnal ports
	usedPorts = append(usedPorts, r.generateDefaultServicePort()...)

	return &corev1.Service{
		ObjectMeta: templates.ObjectMetaWithAnnotations(
			fmt.Sprintf(nifiutils.HeadlessServiceTemplate, r.NifiCluster.Name),
			util.MergeLabels(LabelsForNifi(r.NifiCluster.Name), r.NifiCluster.Labels),
			r.NifiCluster.Spec.Service.Annotations,
			r.NifiCluster,
		),
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeClusterIP,
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector:        LabelsForNifi(r.NifiCluster.Name),
			Ports:           usedPorts,
			ClusterIP:       corev1.ClusterIPNone,
		},
	}
}

