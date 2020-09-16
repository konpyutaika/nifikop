package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NifiRegistryClientSpec defines the desired state of NifiRegistryClient
// +k8s:openapi-gen=true
type NifiRegistryClientSpec struct {
	// The URI of the NiFi registry that should be used for pulling the flow.
	Uri string `json:"uri"`
	// The Description of the Registry client.
	Description string `json:"description,omitempty"`
	// Contains the reference to the NifiCluster with the one the registry client is linked.
	ClusterRef ClusterReference `json:"clusterRef"`
}

// NifiRegistryClientStatus defines the observed state of NifiRegistryClient
// +k8s:openapi-gen=true
type NifiRegistryClientStatus struct {
	// The nifi registry client's id
	Id string `json:"id"`
	// The last nifi registry client revision version catched
	Version int64 `json:"version"`
}

// Nifi Registry Client is the Schema for the nifi registry client API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type NifiRegistryClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiRegistryClientSpec   `json:"spec,omitempty"`
	Status NifiRegistryClientStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NifiRegistryClientList contains a list of NifiRegistryClient
type NifiRegistryClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiRegistryClient `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiRegistryClient{}, &NifiRegistryClientList{})
}
