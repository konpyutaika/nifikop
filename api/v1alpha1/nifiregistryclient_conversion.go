package v1alpha1

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

// ConvertTo converts a v1alpha1 to v1 (Hub).
func (src *NifiRegistryClient) ConvertTo(dst conversion.Hub) error {
	ncV1 := dst.(*v1.NifiRegistryClient)

	if err := ConvertNifiRegistryClientTo(src, ncV1); err != nil {
		return fmt.Errorf("unable to convert NifiRegistryClient %s/%s to version: %v, err: %w", src.Namespace, src.Name, dst.GetObjectKind().GroupVersionKind().Version, err)
	}

	return nil
}

// ConvertFrom converts a v1 (Hub) to v1alpha1 (local).
func (dst *NifiRegistryClient) ConvertFrom(src conversion.Hub) error { //nolint
	ncV1 := src.(*v1.NifiRegistryClient)
	dst.ObjectMeta = ncV1.ObjectMeta
	if err := ConvertNifiRegistryClientFrom(dst, ncV1); err != nil {
		return fmt.Errorf("unable to convert NifiRegistryClient %s/%s from version: %v, err: %w", dst.Namespace, dst.Name, src.GetObjectKind().GroupVersionKind().Version, err)
	}
	return nil
}

// ---- Convert TO ----

// ConvertNifiRegistryClientTo use to convert v1alpha1.NifiRegistryClient to v1.NifiRegistryClient.
func ConvertNifiRegistryClientTo(src *NifiRegistryClient, dst *v1.NifiRegistryClient) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertNifiRegistryClientSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertNifiRegistryClientStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertNifiRegistryClientSpec(src *NifiRegistryClientSpec, dst *v1.NifiRegistryClient) error {
	if src == nil {
		return nil
	}
	dst.Spec.Uri = src.Uri
	dst.Spec.Description = src.Description
	convertNifiRegistryClientSpecClusterRef(src.ClusterRef, dst)
	return nil
}

func convertNifiRegistryClientSpecClusterRef(src ClusterReference, dst *v1.NifiRegistryClient) {
	dst.Spec.ClusterRef = getV1ClusterReference(src)
}

func convertNifiRegistryClientStatus(src *NifiRegistryClientStatus, dst *v1.NifiRegistryClient) error {
	if src == nil {
		return nil
	}
	dst.Status.Id = src.Id
	dst.Status.Version = src.Version
	return nil
}

// ---- Convert FROM ----

// ConvertNifiRegistryClientFrom use to convert v1alpha1.NifiRegistryClient from v1.NifiRegistryClient.
func ConvertNifiRegistryClientFrom(dst *NifiRegistryClient, src *v1.NifiRegistryClient) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertFromNifiRegistryClientSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertFromNifiRegistryClientStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertFromNifiRegistryClientSpec(src *v1.NifiRegistryClientSpec, dst *NifiRegistryClient) error {
	if src == nil {
		return nil
	}
	dst.Spec.Uri = src.Uri
	dst.Spec.Description = src.Description
	convertFromNifiRegistryClientSpecClusterRef(src.ClusterRef, dst)
	return nil
}

func convertFromNifiRegistryClientSpecClusterRef(src v1.ClusterReference, dst *NifiRegistryClient) {
	dst.Spec.ClusterRef = getClusterReference(src)
}

func convertFromNifiRegistryClientStatus(src *v1.NifiRegistryClientStatus, dst *NifiRegistryClient) error {
	if src == nil {
		return nil
	}
	dst.Status.Id = src.Id
	dst.Status.Version = src.Version
	return nil
}
