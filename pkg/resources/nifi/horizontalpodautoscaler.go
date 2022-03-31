package nifi

import (
	"fmt"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	"github.com/konpyutaika/nifikop/pkg/util"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create a HorizontalPodAutoscaler CR
func (r *Reconciler) horizontalPodAutoscaler(log logr.Logger) (runtimeClient.Object, error) {
	return &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v2",
		},
		ObjectMeta: templates.ObjectMetaWithAnnotations(
			fmt.Sprintf("%s-hpa", r.NifiCluster.Name),
			util.MergeLabels(nifiutil.LabelsForNifi(r.NifiCluster.Name), r.NifiCluster.Labels),
			r.NifiCluster.Spec.Service.Annotations,
			r.NifiCluster,
		),
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "NifiCluster",
				APIVersion: "nifi.konpyutaika.com/v1alpha1",
				Name:       r.NifiCluster.Name,
			},
			MinReplicas: &r.NifiCluster.Spec.AutoScalingConfig.HorizontalAutoscaler.MinReplicas,
			MaxReplicas: r.NifiCluster.Spec.AutoScalingConfig.HorizontalAutoscaler.MaxReplicas,
			Metrics:     r.NifiCluster.Spec.AutoScalingConfig.HorizontalAutoscaler.Metrics,
			Behavior:    r.NifiCluster.Spec.AutoScalingConfig.HorizontalAutoscaler.Behavior,
		},
	}, nil
}
