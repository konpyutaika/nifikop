package v1alpha1

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

// ConvertTo converts a v1alpha1 to v1 (Hub).
func (src *NifiUser) ConvertTo(dst conversion.Hub) error {
	ncV1 := dst.(*v1.NifiUser)

	if err := ConvertNifiUserTo(src, ncV1); err != nil {
		return fmt.Errorf("unable to convert NifiUser %s/%s to version: %v, err: %w", src.Namespace, src.Name, dst.GetObjectKind().GroupVersionKind().Version, err)
	}

	return nil
}

// ConvertFrom converts a v1 (Hub) to v1alpha1 (local).
func (dst *NifiUser) ConvertFrom(src conversion.Hub) error { //nolint
	ncV1 := src.(*v1.NifiUser)
	dst.ObjectMeta = ncV1.ObjectMeta
	if err := ConvertNifiUserFrom(dst, ncV1); err != nil {
		return fmt.Errorf("unable to convert NifiUser %s/%s from version: %v, err: %w", dst.Namespace, dst.Name, src.GetObjectKind().GroupVersionKind().Version, err)
	}
	return nil
}

// ---- Convert TO ----

// ConvertNifiUserTo use to convert v1alpha1.NifiUser to v1.NifiUser.
func ConvertNifiUserTo(src *NifiUser, dst *v1.NifiUser) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertNifiUserSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertNifiUserStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertNifiUserSpec(src *NifiUserSpec, dst *v1.NifiUser) error {
	if src == nil {
		return nil
	}
	dst.Spec.Identity = src.Identity
	dst.Spec.SecretName = src.SecretName
	convertNifiUserClusterRef(src.ClusterRef, dst)
	dst.Spec.DNSNames = src.DNSNames
	dst.Spec.IncludeJKS = src.IncludeJKS
	if src.CreateCert != nil {
		dst.Spec.CreateCert = src.CreateCert
	}
	convertNifiUserAccessPolicies(src.AccessPolicies, dst)
	return nil
}

func convertNifiUserClusterRef(src ClusterReference, dst *v1.NifiUser) {
	dst.Spec.ClusterRef = getV1ClusterReference(src)
}

func convertNifiUserAccessPolicies(src []AccessPolicy, dst *v1.NifiUser) {
	dst.Spec.AccessPolicies = []v1.AccessPolicy{}
	for _, srcAccessPolicy := range src {
		dst.Spec.AccessPolicies = append(dst.Spec.AccessPolicies, getV1AccessPolicy(srcAccessPolicy))
	}
}

func convertNifiUserStatus(src *NifiUserStatus, dst *v1.NifiUser) error {
	if src == nil {
		return nil
	}
	dst.Status.Id = src.Id
	dst.Status.Version = src.Version
	return nil
}

// ---- Convert FROM ----

// ConvertNifiUserFrom use to convert v1alpha1.NifiUser from v1.NifiUser.
func ConvertNifiUserFrom(dst *NifiUser, src *v1.NifiUser) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertFromNifiUserSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertFromNifiUserStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertFromNifiUserSpec(src *v1.NifiUserSpec, dst *NifiUser) error {
	if src == nil {
		return nil
	}
	dst.Spec.Identity = src.Identity
	dst.Spec.SecretName = src.SecretName
	convertFromNifiUserClusterRef(src.ClusterRef, dst)
	dst.Spec.DNSNames = src.DNSNames
	dst.Spec.IncludeJKS = src.IncludeJKS
	if src.CreateCert != nil {
		dst.Spec.CreateCert = src.CreateCert
	}
	convertFromNifiUserAccessPolicies(src.AccessPolicies, dst)
	return nil
}

func convertFromNifiUserClusterRef(src v1.ClusterReference, dst *NifiUser) {
	dst.Spec.ClusterRef = getClusterReference(src)
}

func convertFromNifiUserAccessPolicies(src []v1.AccessPolicy, dst *NifiUser) {
	dst.Spec.AccessPolicies = []AccessPolicy{}
	for _, srcAccessPolicy := range src {
		dst.Spec.AccessPolicies = append(dst.Spec.AccessPolicies, getAccessPolicy(srcAccessPolicy))
	}
}

func convertFromNifiUserStatus(src *v1.NifiUserStatus, dst *NifiUser) error {
	if src == nil {
		return nil
	}
	dst.Status.Id = src.Id
	dst.Status.Version = src.Version
	return nil
}
