package v1alpha1

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

// ConvertTo converts a v1alpha1 to v1 (Hub).
func (src *NifiParameterContext) ConvertTo(dst conversion.Hub) error {
	ncV1 := dst.(*v1.NifiParameterContext)

	if err := ConvertNifiParameterContextTo(src, ncV1); err != nil {
		return fmt.Errorf("unable to convert NifiParameterContext %s/%s to version: %v, err: %w", src.Namespace, src.Name, dst.GetObjectKind().GroupVersionKind().Version, err)
	}

	return nil
}

// ConvertFrom converts a v1 (Hub) to v1alpha1 (local).
func (dst *NifiParameterContext) ConvertFrom(src conversion.Hub) error { //nolint
	ncV1 := src.(*v1.NifiParameterContext)
	dst.ObjectMeta = ncV1.ObjectMeta
	if err := ConvertNifiParameterContextFrom(dst, ncV1); err != nil {
		return fmt.Errorf("unable to convert NifiParameterContext %s/%s from version: %v, err: %w", dst.Namespace, dst.Name, src.GetObjectKind().GroupVersionKind().Version, err)
	}
	return nil
}

// ---- Convert TO ----

// ConvertNifiParameterContextTo use to convert v1alpha1.NifiParameterContext to v1.NifiParameterContext.
func ConvertNifiParameterContextTo(src *NifiParameterContext, dst *v1.NifiParameterContext) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertNifiParameterContextSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertNifiParameterContextStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertNifiParameterContextSpec(src *NifiParameterContextSpec, dst *v1.NifiParameterContext) error {
	if src == nil {
		return nil
	}

	dst.Spec.Description = src.Description
	convertNifiParameterContextParameters(src.Parameters, dst)
	convertNifiParameterContextClusterRef(src.ClusterRef, dst)
	convertNifiParameterContextSecretRefs(src.SecretRefs, dst)
	convertNifiParameterContextInheritedParameterContexts(src.InheritedParameterContexts, dst)
	if src.DisableTakeOver != nil {
		dst.Spec.DisableTakeOver = src.DisableTakeOver
	}

	return nil
}

func convertNifiParameterContextParameters(src []Parameter, dst *v1.NifiParameterContext) {
	dst.Spec.Parameters = []v1.Parameter{}
	for _, srcParameter := range src {
		dstParameter := v1.Parameter{
			Name:        srcParameter.Name,
			Description: srcParameter.Description,
			Sensitive:   srcParameter.Sensitive,
		}
		if srcParameter.Value != nil {
			dstParameter.Value = srcParameter.Value
		}
		dst.Spec.Parameters = append(dst.Spec.Parameters, dstParameter)
	}
}

func convertNifiParameterContextClusterRef(src ClusterReference, dst *v1.NifiParameterContext) {
	dst.Spec.ClusterRef = getV1ClusterReference(src)
}

func convertNifiParameterContextSecretRefs(src []SecretReference, dst *v1.NifiParameterContext) {
	dst.Spec.SecretRefs = []v1.SecretReference{}
	for _, srcSecretRef := range src {
		dst.Spec.SecretRefs = append(dst.Spec.SecretRefs, getV1SecretRef(srcSecretRef))
	}
}

func convertNifiParameterContextInheritedParameterContexts(src []ParameterContextReference, dst *v1.NifiParameterContext) {
	dst.Spec.InheritedParameterContexts = []v1.ParameterContextReference{}
	for _, srcParameterContextReference := range src {
		dst.Spec.InheritedParameterContexts = append(dst.Spec.InheritedParameterContexts, getV1ParameterContextRef(srcParameterContextReference))
	}
}

func convertNifiParameterContextStatus(src *NifiParameterContextStatus, dst *v1.NifiParameterContext) error {
	if src == nil {
		return nil
	}
	dst.Status.Id = src.Id
	dst.Status.Version = src.Version
	convertNifiParameterContextLatestUpdateRequest(src.LatestUpdateRequest, dst)
	return nil
}

func convertNifiParameterContextLatestUpdateRequest(src *ParameterContextUpdateRequest, dst *v1.NifiParameterContext) {
	if src == nil {
		return
	}

	dst.Status.LatestUpdateRequest = &v1.ParameterContextUpdateRequest{
		Id:               src.Id,
		Uri:              src.Uri,
		SubmissionTime:   src.SubmissionTime,
		LastUpdated:      src.LastUpdated,
		Complete:         src.Complete,
		FailureReason:    src.FailureReason,
		PercentCompleted: src.PercentCompleted,
		State:            src.State,
	}
}

// ---- Convert FROM ----

// ConvertNifiParameterContextFrom use to convert v1alpha1.NifiParameterContext from v1.NifiParameterContext.
func ConvertNifiParameterContextFrom(dst *NifiParameterContext, src *v1.NifiParameterContext) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertFromNifiParameterContextSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertFromNifiParameterContextStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertFromNifiParameterContextSpec(src *v1.NifiParameterContextSpec, dst *NifiParameterContext) error {
	if src == nil {
		return nil
	}

	dst.Spec.Description = src.Description
	convertFromNifiParameterContextParameters(src.Parameters, dst)
	convertFromNifiParameterContextClusterRef(src.ClusterRef, dst)
	convertFromNifiParameterContextSecretRefs(src.SecretRefs, dst)
	convertFromNifiParameterContextInheritedParameterContexts(src.InheritedParameterContexts, dst)
	if src.DisableTakeOver != nil {
		dst.Spec.DisableTakeOver = src.DisableTakeOver
	}

	return nil
}

func convertFromNifiParameterContextParameters(src []v1.Parameter, dst *NifiParameterContext) {
	dst.Spec.Parameters = []Parameter{}
	for _, srcParameter := range src {
		dstParameter := Parameter{
			Name:        srcParameter.Name,
			Description: srcParameter.Description,
			Sensitive:   srcParameter.Sensitive,
		}
		if srcParameter.Value != nil {
			dstParameter.Value = srcParameter.Value
		}
		dst.Spec.Parameters = append(dst.Spec.Parameters, dstParameter)
	}
}

func convertFromNifiParameterContextClusterRef(src v1.ClusterReference, dst *NifiParameterContext) {
	dst.Spec.ClusterRef = getClusterReference(src)
}

func convertFromNifiParameterContextSecretRefs(src []v1.SecretReference, dst *NifiParameterContext) {
	dst.Spec.SecretRefs = []SecretReference{}
	for _, srcSecretRef := range src {
		dst.Spec.SecretRefs = append(dst.Spec.SecretRefs, getSecretRef(srcSecretRef))
	}
}

func convertFromNifiParameterContextInheritedParameterContexts(src []v1.ParameterContextReference, dst *NifiParameterContext) {
	dst.Spec.InheritedParameterContexts = []ParameterContextReference{}
	for _, srcParameterContextReference := range src {
		dst.Spec.InheritedParameterContexts = append(dst.Spec.InheritedParameterContexts, getParameterContextRef(srcParameterContextReference))
	}
}

func convertFromNifiParameterContextStatus(src *v1.NifiParameterContextStatus, dst *NifiParameterContext) error {
	if src == nil {
		return nil
	}
	dst.Status.Id = src.Id
	dst.Status.Version = src.Version
	convertFromNifiParameterContextLatestUpdateRequest(src.LatestUpdateRequest, dst)
	return nil
}

func convertFromNifiParameterContextLatestUpdateRequest(src *v1.ParameterContextUpdateRequest, dst *NifiParameterContext) {
	if src == nil {
		return
	}

	dst.Status.LatestUpdateRequest = &ParameterContextUpdateRequest{
		Id:               src.Id,
		Uri:              src.Uri,
		SubmissionTime:   src.SubmissionTime,
		LastUpdated:      src.LastUpdated,
		Complete:         src.Complete,
		FailureReason:    src.FailureReason,
		PercentCompleted: src.PercentCompleted,
		State:            src.State,
	}
}
