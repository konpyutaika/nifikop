package v1alpha1

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v2alpha1 "github.com/konpyutaika/nifikop/api/v2alpha1"
)

// ConvertTo converts a v1alpha1 to v2alpha1 (Hub).
func (src *NifiRegistryClient) ConvertTo(dst conversion.Hub) error {
	ncV2 := dst.(*v2alpha1.NifiRegistryClient)

	if err := ConvertNifiRegistryClientTo(src, ncV2); err != nil {
		return fmt.Errorf("unable to convert NifiRegistryClient %s/%s to version: %v, err: %w", src.Namespace, src.Name, dst.GetObjectKind().GroupVersionKind().Version, err)
	}

	return nil
}

// ConvertFrom converts a v2alpha1 (Hub) to v1alpha1 (local).
func (dst *NifiRegistryClient) ConvertFrom(src conversion.Hub) error { //nolint
	ncV2 := src.(*v2alpha1.NifiRegistryClient)
	dst.ObjectMeta = ncV2.ObjectMeta
	if err := ConvertNifiRegistryClientFrom(dst, ncV2); err != nil {
		return fmt.Errorf("unable to convert NifiRegistryClient %s/%s from version: %v, err: %w", dst.Namespace, dst.Name, src.GetObjectKind().GroupVersionKind().Version, err)
	}
	return nil
}

// ---- Convert TO ----

// ConvertNifiRegistryClientTo converts a v1alpha1.NifiRegistryClient to v2alpha1.NifiRegistryClient.
func ConvertNifiRegistryClientTo(src *NifiRegistryClient, dst *v2alpha1.NifiRegistryClient) error {
	dst.ObjectMeta = src.ObjectMeta

	if err := convertNifiRegistryClientSpec(&src.Spec, dst); err != nil {
		return err
	}

	if err := convertNifiRegistryClientStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

func convertNifiRegistryClientSpec(src *NifiRegistryClientSpec, dst *v2alpha1.NifiRegistryClient) error {
	if src == nil {
		return nil
	}
	dst.Spec.Description = src.Description
	dst.Spec.ClusterRef = v2alpha1.ClusterReference{
		Name:      src.ClusterRef.Name,
		Namespace: src.ClusterRef.Namespace,
	}
	// v1alpha1 only supported the NiFi Registry type.
	dst.Spec.Type = v2alpha1.RegistryClientType
	dst.Spec.RegistryClientConfig = &v2alpha1.RegistryClientConfig{
		Uri: src.Uri,
	}
	return nil
}

func convertNifiRegistryClientStatus(src *NifiRegistryClientStatus, dst *v2alpha1.NifiRegistryClient) error {
	if src == nil {
		return nil
	}
	dst.Status.Id = src.Id
	dst.Status.Version = src.Version
	return nil
}

// ---- Convert FROM ----

// ConvertNifiRegistryClientFrom converts a v2alpha1.NifiRegistryClient to v1alpha1.NifiRegistryClient.
func ConvertNifiRegistryClientFrom(dst *NifiRegistryClient, src *v2alpha1.NifiRegistryClient) error {
	dst.ObjectMeta = src.ObjectMeta

	if err := convertFromNifiRegistryClientSpec(&src.Spec, dst); err != nil {
		return err
	}

	if err := convertFromNifiRegistryClientStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

func convertFromNifiRegistryClientSpec(src *v2alpha1.NifiRegistryClientSpec, dst *NifiRegistryClient) error {
	if src == nil {
		return nil
	}
	dst.Spec.Description = src.Description
	dst.Spec.ClusterRef = ClusterReference{
		Name:      src.ClusterRef.Name,
		Namespace: src.ClusterRef.Namespace,
	}
	// v1alpha1 only has Uri; use RegistryClientConfig if present.
	if src.RegistryClientConfig != nil {
		dst.Spec.Uri = src.RegistryClientConfig.Uri
	}
	return nil
}

func convertFromNifiRegistryClientStatus(src *v2alpha1.NifiRegistryClientStatus, dst *NifiRegistryClient) error {
	if src == nil {
		return nil
	}
	dst.Status.Id = src.Id
	dst.Status.Version = src.Version
	return nil
}
