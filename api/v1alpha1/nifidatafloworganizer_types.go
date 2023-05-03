package v1alpha1

import (
	"strings"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NifiDataflowOrganizerSpec defines the desired state of NifiDataflowOrganizer
type NifiDataflowOrganizerSpec struct {
	// contains the reference to the NifiCluster with the one the user is linked
	ClusterRef v1.ClusterReference `json:"clusterRef"`
	// the groups of dataflow to organize
	Groups []OrganizerGroup `json:"groups"`
}

type OrganizerGroup struct {
	// the UUID of the parent process group where you want to create your group, if not set deploy at root level.
	// +optional
	ParentProcessGroupID string `json:"parentProcessGroupID,omitempty"`
	// the name of the group
	Name string `json:"name"`
	// the color of the group
	// +kubebuilder:default="#FFF7D7"
	// +kubebuilder:validation:Pattern:="^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$"
	// +optional
	Color string `json:"color,omitempty"`
	// the font size of the group
	// +kubebuilder:default="18px"
	// +kubebuilder:validation:Pattern:="^([0-9]+px)$"
	// +optional
	FontSize string `json:"fontSize,omitempty"`
}

// NifiDataflowOrganizerStatus defines the observed state of NifiDataflowOrganizer
type NifiDataflowOrganizerStatus struct {
	// the status of the groups.
	GroupStatus []OrganizerGroupStatus `json:"groupStatus"`
}

type OrganizerGroupStatus struct {
	// the status of the title label.
	TitleStatus OrganizerGroupTitleStatus `json:"titleStatus"`
	// the status of the content label.
	ContentStatus OrganizerGroupContentStatus `json:"contentStatus"`
}

type OrganizerGroupTitleStatus struct {
	// the id of the title label.
	Id string `json:"id"`
}

type OrganizerGroupContentStatus struct {
	// the id of the content label.
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

func (d *OrganizerGroup) GetParentProcessGroupID(rootProcessGroupId string) string {
	if d.ParentProcessGroupID == "" {
		return rootProcessGroupId
	}
	return d.ParentProcessGroupID
}

func (d *OrganizerGroup) GetTitleWidth() float64 {
	return float64(len(d.Name)) * util.ConvertStringToFloat64(strings.Split(d.FontSize, "px")[0]) * 0.65
}

func (d *OrganizerGroup) GetTitleHeight() float64 {
	return util.ConvertStringToFloat64(strings.Split(d.FontSize, "px")[0]) * 1.75
}
