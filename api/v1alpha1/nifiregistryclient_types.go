package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiRegistryClientSpec defines the desired state of NifiRegistryClient.
type NifiRegistryClientSpec struct {
	// The URI of the NiFi registry that should be used for pulling the flow.
	Uri string `json:"uri"`
	// The Description of the Registry client.
	Description string `json:"description,omitempty"`
	// contains the reference to the NifiCluster with the one the registry client is linked.
	ClusterRef ClusterReference `json:"clusterRef,omitempty"`
}

// NifiRegistryClientStatus defines the observed state of NifiRegistryClient.
type NifiRegistryClientStatus struct {
	// The nifi registry client's id
	Id string `json:"id"`
	// The last nifi registry client revision version catched
	Version int64 `json:"version"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NifiRegistryClient is the Schema for the nifiregistryclients API.
type NifiRegistryClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiRegistryClientSpec   `json:"spec,omitempty"`
	Status NifiRegistryClientStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NifiRegistryClientList contains a list of NifiRegistryClient.
type NifiRegistryClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiRegistryClient `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiRegistryClient{}, &NifiRegistryClientList{})
}
