package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var nificlusterlog = logf.Log.WithName("nificluster-resource")

func (r *NifiCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-nifi-konpyutaika-com-v1alpha1-nificluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=nifi.konpyutaika.com,resources=nificlusters,verbs=create;update,versions=v1alpha1,name=vnificluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &NifiCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *NifiCluster) ValidateCreate() error {
	nificlusterlog.Info("Validating cluster CR Create", "name", r.Name)

	return r.validateNifiCluster()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *NifiCluster) ValidateUpdate(old runtime.Object) error {
	nificlusterlog.Info("Validating cluster CR Update", "name", r.Name)

	return r.validateNifiCluster()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *NifiCluster) ValidateDelete() error {
	// nothing to do for delete
	return nil
}

func (r *NifiCluster) validateNifiCluster() error {
	var allErrs field.ErrorList
	if err := r.validateNodeAndAutoscalingConfig(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "nifi.konpyutaika.com", Kind: "NifiCluster"},
		r.Name, allErrs)
}

func (r *NifiCluster) validateNodeAndAutoscalingConfig() *field.Error {
	// one of Spec.Nodes or Spec.AutoscalingConfig must be configured
	if (r.Spec.Nodes == nil || len(r.Spec.Nodes) == 0) && !r.Spec.AutoScalingConfig.Enabled {
		return field.Invalid(field.NewPath("spec").Child("Nodes"), r.Spec.Nodes, "You must configure one of Spec.Nodes or Spec.AutoscalingConfig.Enabled in order to create a cluster.")
	}

	// you can't configure both Spec.Nodes and Spec.AutoscalingConfig
	if r.Spec.Nodes != nil && r.Spec.AutoScalingConfig.Enabled {
		return field.Invalid(field.NewPath("spec").Child("Nodes"), r.Spec.Nodes, "You may not configure Spec.Nodes when Spec.AutoscalingConfig.Enabled is true. Deployments must be static or dynamic, not both.")
	}

	return nil
}
