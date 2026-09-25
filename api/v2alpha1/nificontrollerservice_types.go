package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NifiRegistryClientType defines the type of registry client.
// +kubebuilder:validation:Enum={"org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider"}
type NifiControllerServiceType string

const (
	StandardWebClientServiceProviderType NifiControllerServiceType = "org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider"
)

// NifiControllerServiceSpec defines the desired state of NifiControllerService.
// +kubebuilder:validation:XValidation:rule="self.type != 'org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider' || has(self.standardWebClientServiceProviderSpec)",message="standardWebClientServiceProviderSpec is required when type is 'org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider'"
type NifiControllerServiceSpec struct {
	// The description of the controller service.
	// +optional
	Description string `json:"description,omitempty"`
	// Reference to the NifiCluster this controller service is linked to.
	// +optional
	ClusterRef ClusterReference `json:"clusterRef,omitempty"`
	// Type of the controller service.
	// +kubebuilder:default=org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider
	// +optional
	Type NifiControllerServiceType `json:"type,omitempty"`

	// StandardWebClientServiceProviderSpec holds configuration for a StandardWebClientServiceProvider type service.
	// Required when type is "org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider".
	// +optional
	StandardWebClientServiceProviderSpec *StandardWebClientServiceProviderSpec `json:"standardWebClientServiceProviderSpec,omitempty"`
}

// GetType returns the NiFi API type string (full class name) for this controller service.
func (r *NifiControllerServiceSpec) GetType() string {
	switch r.Type {
	case StandardWebClientServiceProviderType:
		return string(r.Type)
	default:
		return string(StandardWebClientServiceProviderType)
	}
}

type StandardWebClientServiceProviderSpec struct {
	// +kubebuilder:default="10 secs"
	ConnectTimeout string `json:"connectTimeout,omitempty"`
	// +kubebuilder:default="10 secs"
	ReadTimeout string `json:"readTimeout,omitempty"`
	// +kubebuilder:default="10 secs"
	WriteTimeout string `json:"writeTimeout,omitempty"`
}

// NifiControllerServiceStatus defines the observed state of NifiControllerService.
type NifiControllerServiceStatus struct {
	// The nifi controller service's id.
	Id string `json:"id"`
	// The last nifi registry client revision version caught.
	Version int64 `json:"version"`
	// The last observed resource versions of the referenced secrets.
	// +optional
	LatestSecretsResourceVersion []SecretResourceVersion `json:"latestSecretsResourceVersion,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// NifiControllerService is the Schema for the nificontrollerservices API.
type NifiControllerService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiControllerServiceSpec   `json:"spec,omitempty"`
	Status NifiControllerServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NifiControllerServiceList contains a list of NifiControllerService.
type NifiControllerServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiControllerService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiControllerService{}, &NifiControllerServiceList{})
}
