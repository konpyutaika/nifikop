package v1alpha1

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

// ConvertTo converts a v1alpha1 to v1 (Hub).
func (src *NifiUserGroup) ConvertTo(dst conversion.Hub) error {
	ncV1 := dst.(*v1.NifiUserGroup)

	if err := ConvertNifiUserGroupTo(src, ncV1); err != nil {
		return fmt.Errorf("unable to convert NifiUserGroup %s/%s to version: %v, err: %w", src.Namespace, src.Name, dst.GetObjectKind().GroupVersionKind().Version, err)
	}

	return nil
}

// ConvertFrom converts a v1 (Hub) to v1alpha1 (local).
func (dst *NifiUserGroup) ConvertFrom(src conversion.Hub) error { //nolint
	ncV1 := src.(*v1.NifiUserGroup)
	dst.ObjectMeta = ncV1.ObjectMeta
	if err := ConvertNifiUserGroupFrom(dst, ncV1); err != nil {
		return fmt.Errorf("unable to convert NifiUserGroup %s/%s from version: %v, err: %w", dst.Namespace, dst.Name, src.GetObjectKind().GroupVersionKind().Version, err)
	}
	return nil
}

// ---- Convert TO ----

// ConvertNifiUserGroupTo use to convert v1alpha1.NifiUserGroup to v1.NifiUserGroup.
func ConvertNifiUserGroupTo(src *NifiUserGroup, dst *v1.NifiUserGroup) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertNifiUserGroupSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertNifiUserGroupStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertNifiUserGroupSpec(src *NifiUserGroupSpec, dst *v1.NifiUserGroup) error {
	if src == nil {
		return nil
	}
	convertNifiUserGroupClusterRef(src.ClusterRef, dst)
	convertNifiUserGroupUsersRef(src.UsersRef, dst)
	convertNifiUserGroupAccessPolicies(src.AccessPolicies, dst)
	return nil
}

func convertNifiUserGroupClusterRef(src ClusterReference, dst *v1.NifiUserGroup) {
	dst.Spec.ClusterRef = getV1ClusterReference(src)
}

func convertNifiUserGroupUsersRef(src []UserReference, dst *v1.NifiUserGroup) {
	dst.Spec.UsersRef = []v1.UserReference{}
	for _, srcUserRef := range src {
		dst.Spec.UsersRef = append(dst.Spec.UsersRef, v1.UserReference{
			Name:      srcUserRef.Name,
			Namespace: srcUserRef.Namespace,
		})
	}
}

func convertNifiUserGroupAccessPolicies(src []AccessPolicy, dst *v1.NifiUserGroup) {
	dst.Spec.AccessPolicies = []v1.AccessPolicy{}
	for _, srcAccessPolicy := range src {
		dst.Spec.AccessPolicies = append(dst.Spec.AccessPolicies, getV1AccessPolicy(srcAccessPolicy))
	}
}

func convertNifiUserGroupStatus(src *NifiUserGroupStatus, dst *v1.NifiUserGroup) error {
	if src == nil {
		return nil
	}
	dst.Status.Id = src.Id
	dst.Status.Version = src.Version
	return nil
}

// ---- Convert FROM ----

// ConvertNifiUserGroupFrom use to convert v1alpha1.NifiUserGroup From v1.NifiUserGroup.
func ConvertNifiUserGroupFrom(dst *NifiUserGroup, src *v1.NifiUserGroup) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertFromNifiUserGroupSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertFromNifiUserGroupStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertFromNifiUserGroupSpec(src *v1.NifiUserGroupSpec, dst *NifiUserGroup) error {
	if src == nil {
		return nil
	}
	convertFromNifiUserGroupClusterRef(src.ClusterRef, dst)
	convertFromNifiUserGroupUsersRef(src.UsersRef, dst)
	convertFromNifiUserGroupAccessPolicies(src.AccessPolicies, dst)
	return nil
}

func convertFromNifiUserGroupClusterRef(src v1.ClusterReference, dst *NifiUserGroup) {
	dst.Spec.ClusterRef = getClusterReference(src)
}

func convertFromNifiUserGroupUsersRef(src []v1.UserReference, dst *NifiUserGroup) {
	dst.Spec.UsersRef = []UserReference{}
	for _, srcUserRef := range src {
		dst.Spec.UsersRef = append(dst.Spec.UsersRef, UserReference{
			Name:      srcUserRef.Name,
			Namespace: srcUserRef.Namespace,
		})
	}
}

func convertFromNifiUserGroupAccessPolicies(src []v1.AccessPolicy, dst *NifiUserGroup) {
	dst.Spec.AccessPolicies = []AccessPolicy{}
	for _, srcAccessPolicy := range src {
		dst.Spec.AccessPolicies = append(dst.Spec.AccessPolicies, getAccessPolicy(srcAccessPolicy))
	}
}

func convertFromNifiUserGroupStatus(src *v1.NifiUserGroupStatus, dst *NifiUserGroup) error {
	if src == nil {
		return nil
	}
	dst.Status.Id = src.Id
	dst.Status.Version = src.Version
	return nil
}
