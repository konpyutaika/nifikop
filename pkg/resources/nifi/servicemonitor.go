package nifi

import (
	"errors"
	"fmt"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	"github.com/konpyutaika/nifikop/pkg/util"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) serviceMonitor(log logr.Logger) (runtimeClient.Object, error) {
	// ensure there is a prometheus port configured or else fail.
	var found bool
	var prometheusListener v1alpha1.InternalListenerConfig
	for _, listener := range r.NifiCluster.Spec.ListenersConfig.InternalListeners {
		if listener.Type == v1alpha1.PrometheusListenerType {
			found = true
			prometheusListener = listener
			break
		}
	}

	if !found {
		return nil, errors.New("Failed to find a prometheus InternalListener configured. You must configure a prometheus port to enable a ServiceMonitor.")
	}

	matchingLabels := []map[string]string{
		nifiutil.LabelsForNifi(r.NifiCluster.Name),
	}

	selector := metav1.LabelSelector{
		MatchLabels: util.MergeLabels(matchingLabels...),
	}

	return &promv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceMonitor",
			APIVersion: "v1",
		},
		ObjectMeta: templates.ObjectMetaWithAnnotations(
			fmt.Sprintf("%s-service-monitor", r.NifiCluster.Name),
			util.MergeLabels(nifiutil.LabelsForNifi(r.NifiCluster.Name), r.NifiCluster.Labels),
			r.NifiCluster.Spec.Service.Annotations,
			r.NifiCluster,
		),
		Spec: promv1.ServiceMonitorSpec{
			Endpoints: []promv1.Endpoint{
				{
					Port: prometheusListener.Name,
					Path: "/metrics",
				},
			},
			NamespaceSelector: promv1.NamespaceSelector{
				MatchNames: []string{
					r.NifiCluster.Namespace,
				},
			},
			Selector: selector,
		},
	}, nil
}
