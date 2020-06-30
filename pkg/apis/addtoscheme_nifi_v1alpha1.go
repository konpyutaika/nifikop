package apis

import (
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/apis/nifi/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
}