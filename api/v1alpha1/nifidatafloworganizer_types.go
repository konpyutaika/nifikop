package v1alpha1

import (
	v1 "github.com/konpyutaika/nifikop/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NifiDataflowOrganizerSpec defines the desired state of NifiDataflowOrganizer
type NifiDataflowOrganizerSpec struct {
	// contains the reference to the NifiCluster with the one the user is linked
	ClusterRef v1.ClusterReference `json:"clusterRef"`
	// the UUID of the parent process group where you want to create your labels, if not set deploy at root level.
	ParentProcessGroupID string `json:"parentProcessGroupID,omitempty"`
	// Color of the label
	// +optional
	// +kubebuilder:validation:Pattern:="^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$"
	Color string `json:"color,omitempty"`
}

// NifiDataflowOrganizerStatus defines the observed state of NifiDataflowOrganizer
type NifiDataflowOrganizerStatus struct {
	// the status of the title label.
	TitleLabelStatus TitleLabelStatus `json:"titleLabelStatus"`
}

type TitleLabelStatus struct {
	// the id of the title label.
	Id string `json:"id"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NifiDataflowOrganizer is the Schema for the nifidatafloworganizers API
type NifiDataflowOrganizer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiDataflowOrganizerSpec   `json:"spec,omitempty"`
	Status NifiDataflowOrganizerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NifiDataflowOrganizerList contains a list of NifiDataflowOrganizer
type NifiDataflowOrganizerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiDataflowOrganizer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiDataflowOrganizer{}, &NifiDataflowOrganizerList{})
}

func (d *NifiDataflowOrganizerSpec) GetParentProcessGroupID(rootProcessGroupId string) string {
	if d.ParentProcessGroupID == "" {
		return rootProcessGroupId
	}
	return d.ParentProcessGroupID
}
