package v1alpha1

import (
	v1 "github.com/konpyutaika/nifikop/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NifiControllerServiceType is the fully qualified NiFi class name of a controller service.
// +kubebuilder:validation:Enum={"org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider"}
type NifiControllerServiceType string

const (
	// StandardWebClientServiceProviderType is the HTTP client provider consumed, among others,
	// by the Bitbucket flow registry client.
	StandardWebClientServiceProviderType NifiControllerServiceType = "org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider"
)

// NifiControllerServiceState is the run state of a controller service as reported by NiFi.
type NifiControllerServiceState string

const (
	ControllerServiceStateEnabled   NifiControllerServiceState = "ENABLED"
	ControllerServiceStateEnabling  NifiControllerServiceState = "ENABLING"
	ControllerServiceStateDisabled  NifiControllerServiceState = "DISABLED"
	ControllerServiceStateDisabling NifiControllerServiceState = "DISABLING"
)

// NifiControllerServiceSpec defines the desired state of NifiControllerService.
// +kubebuilder:validation:XValidation:rule="self.type != 'org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider' || has(self.standardWebClientServiceProviderSpec)",message="standardWebClientServiceProviderSpec is required when type is 'org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider'"
type NifiControllerServiceSpec struct {
	// The description of the controller service, stored as the component comments in NiFi.
	// +optional
	Description string `json:"description,omitempty"`
	// Reference to the NifiCluster this controller service is linked to.
	// +optional
	ClusterRef v1.ClusterReference `json:"clusterRef,omitempty"`
	// Type of the controller service. The service is created at controller level, so that
	// controller-scoped components such as flow registry clients can reference it.
	// +kubebuilder:default=org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider
	// +optional
	Type NifiControllerServiceType `json:"type,omitempty"`

	// StandardWebClientServiceProviderSpec holds the configuration of a StandardWebClientServiceProvider.
	// Required when type is "org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider".
	// +optional
	StandardWebClientServiceProviderSpec *StandardWebClientServiceProviderSpec `json:"standardWebClientServiceProviderSpec,omitempty"`
}

// GetType returns the NiFi class name of this controller service.
func (r *NifiControllerServiceSpec) GetType() string {
	return string(r.Type)
}

// StandardWebClientServiceProviderSpec mirrors the properties of the NiFi
// StandardWebClientServiceProvider. Durations use the NiFi time period format, e.g. "10 secs".
type StandardWebClientServiceProviderSpec struct {
	// Maximum amount of time to wait before failing during initial socket connection.
	// +kubebuilder:default="10 secs"
	// +optional
	ConnectTimeout string `json:"connectTimeout,omitempty"`
	// Maximum amount of time to wait before failing while reading socket responses.
	// +kubebuilder:default="10 secs"
	// +optional
	ReadTimeout string `json:"readTimeout,omitempty"`
	// Maximum amount of time to wait before failing while writing socket requests.
	// +kubebuilder:default="10 secs"
	// +optional
	WriteTimeout string `json:"writeTimeout,omitempty"`
}

// NifiControllerServiceStatus defines the observed state of NifiControllerService.
type NifiControllerServiceStatus struct {
	// The nifi controller service's id.
	Id string `json:"id"`
	// The last nifi controller service revision version caught.
	Version int64 `json:"version"`
	// The last run state of the controller service observed in NiFi.
	// +optional
	State NifiControllerServiceState `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

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
