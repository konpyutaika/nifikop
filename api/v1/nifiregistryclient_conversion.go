package v1

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v2alpha1 "github.com/konpyutaika/nifikop/api/v2alpha1"
)

// ConvertTo converts a v1 NifiRegistryClient to a v2alpha1 (Hub).
func (src *NifiRegistryClient) ConvertTo(dst conversion.Hub) error {
	ncV2 := dst.(*v2alpha1.NifiRegistryClient)

	if err := convertNifiRegistryClientToV2alpha1(src, ncV2); err != nil {
		return fmt.Errorf("unable to convert NifiRegistryClient %s/%s to v2alpha1: %w", src.Namespace, src.Name, err)
	}
	return nil
}

// ConvertFrom converts a v2alpha1 (Hub) NifiRegistryClient to v1.
func (dst *NifiRegistryClient) ConvertFrom(src conversion.Hub) error {
	ncV2 := src.(*v2alpha1.NifiRegistryClient)

	if err := convertNifiRegistryClientFromV2alpha1(ncV2, dst); err != nil {
		return fmt.Errorf("unable to convert NifiRegistryClient %s/%s from v2alpha1: %w", dst.Namespace, dst.Name, err)
	}
	return nil
}

// ---- Convert TO v2alpha1 ----

func convertNifiRegistryClientToV2alpha1(src *NifiRegistryClient, dst *v2alpha1.NifiRegistryClient) error {
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Description = src.Spec.Description
	dst.Spec.ClusterRef = v2alpha1.ClusterReference{
		Name:      src.Spec.ClusterRef.Name,
		Namespace: src.Spec.ClusterRef.Namespace,
	}

	// v1 only supported the NiFi Registry type; map its Uri into RegistryClientConfig.
	dst.Spec.Type = v2alpha1.RegistryClientType
	dst.Spec.RegistryClientConfig = &v2alpha1.RegistryClientConfig{
		Uri: src.Spec.Uri,
	}

	// Status
	dst.Status.Id = src.Status.Id
	dst.Status.Version = src.Status.Version

	return nil
}

// ---- Convert FROM v2alpha1 ----

func convertNifiRegistryClientFromV2alpha1(src *v2alpha1.NifiRegistryClient, dst *NifiRegistryClient) error {
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Description = src.Spec.Description
	dst.Spec.ClusterRef = ClusterReference{
		Name:      src.Spec.ClusterRef.Name,
		Namespace: src.Spec.ClusterRef.Namespace,
	}

	// Only the registry type has a direct Uri equivalent in v1.
	if src.Spec.RegistryClientConfig != nil {
		dst.Spec.Uri = src.Spec.RegistryClientConfig.Uri
	}

	// Status
	dst.Status.Id = src.Status.Id
	dst.Status.Version = src.Status.Version

	return nil
}
