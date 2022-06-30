package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiConnectionSpec defines the desired state of NifiConnection
type NifiConnectionSpec struct {
	// the Source component of the connection.
	Source ComponentReference `json:"source"`
	// the Destination component of the connection.
	Destination ComponentReference `json:"destination"`
}

type ComponentReference struct {
	// the name of the component.
	Name string `json:"name"`
	// the namespace of the component.
	Namespace string `json:"namespace,omitempty"`
	// the type of the component (e.g. nifidataflow).
	Type ComponentType `json:"type"`
	// the name of the sub component (e.g. queue or port name)
	SubName string `json:"subName,omitempty"`
}

// NifiConnectionStatus defines the observed state of NifiConnection
type NifiConnectionStatus struct {
	// connection ID
	ConnectionId string `json:"connectionID"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NifiConnection is the Schema for the nificonnections API
type NifiConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiConnectionSpec   `json:"spec,omitempty"`
	Status NifiConnectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NifiConnectionList contains a list of NifiConnection
type NifiConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiConnection{}, &NifiConnectionList{})
}

func (nCon *NifiConnectionSpec) IsValid() bool {
	return nCon.Source.IsValid() && nCon.Destination.IsValid()
}

func (compRef *ComponentReference) IsValid() bool {
	return compRef.Type == ComponentDataflow && compRef.SubName != ""
}
