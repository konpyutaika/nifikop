package v1alpha1

import (
	"math"
	"sort"
	"strings"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/util"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NifiDataflowOrganizerSpec defines the desired state of NifiDataflowOrganizer
type NifiDataflowOrganizerSpec struct {
	// contains the reference to the NifiCluster with the one the user is linked
	ClusterRef v1.ClusterReference `json:"clusterRef"`
	// the groups of dataflow to organize
	Groups map[string]OrganizerGroup `json:"groups"`
	// the maximum width before moving to the next line
	// +kubebuilder:default=1000
	// +kubebuilder:validation:Minimum=0
	// +optional
	MaxWidth int `json:"maxWidth,omitempty"`
	// the initial position of all the groups
	// +optional
	InitialPosition OrganizerGroupPosition `json:"initalPosition,omitempty"`
}

type OrganizerGroup struct {
	// the UUID of the parent process group where you want to create your group, if not set deploy at root level.
	// +optional
	ParentProcessGroupID string `json:"parentProcessGroupID,omitempty"`
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
	// contains the reference to the NifiDataflows associated to the group.
	DataflowRef []v1.DataflowReference `json:"dataflowRef,omitempty"`
	// the maximum number of dataflow on the same line
	// +kubebuilder:default=5
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxColumnSize int `json:"maxColumnSize,omitempty"`
}

type OrganizerGroupPosition struct {
	// The x coordinate.
	// +kubebuilder:default=0
	// +optional
	X int64 `json:"posX,omitempty"`
	// The y coordinate.
	// +kubebuilder:default=0
	// +optional
	Y int64 `json:"posY,omitempty"`
}

// NifiDataflowOrganizerStatus defines the observed state of NifiDataflowOrganizer
type NifiDataflowOrganizerStatus struct {
	// the status of the groups.
	GroupStatus map[string]OrganizerGroupStatus `json:"groupStatus"`
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

func (d *OrganizerGroup) GetTitleWidth(name string) float64 {
	return float64(len(name)) * util.ConvertStringToFloat64(strings.Split(d.FontSize, "px")[0]) * 0.65
}

func (d *OrganizerGroup) GetTitleHeight(name string) float64 {
	return util.ConvertStringToFloat64(strings.Split(d.FontSize, "px")[0]) * 1.75
}

func (d *OrganizerGroup) GetContentWidth() float64 {
	return float64(nifiutil.ProcessGroupWidth+nifiutil.ProcessGroupPadding)*d.GetNumberOfColumns() + float64(nifiutil.ProcessGroupPadding)
}

func (d *OrganizerGroup) GetContentHeight() float64 {
	return float64(nifiutil.ProcessGroupHeight+nifiutil.ProcessGroupPadding)*d.GetNumberOfLines() + float64(nifiutil.ProcessGroupPadding)
}

func (d *OrganizerGroup) GetNumberOfLines() float64 {
	return math.Ceil(math.Max(float64(len(d.DataflowRef)), 1) / d.GetNumberOfColumns())
}

func (d *OrganizerGroup) GetNumberOfColumns() float64 {
	if len(d.DataflowRef) < d.MaxColumnSize {
		return math.Max(float64(len(d.DataflowRef)), 1)
	}
	return float64(d.MaxColumnSize)
}

func (d *NifiDataflowOrganizerSpec) GetGroupNames() []string {
	keys := make([]string, 0)
	for k := range d.Groups {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}
