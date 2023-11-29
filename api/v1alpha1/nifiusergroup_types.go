package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiUserGroupSpec defines the desired state of NifiUserGroup.
type NifiUserGroupSpec struct {
	// clusterRef contains the reference to the NifiCluster with the one the registry client is linked.
	ClusterRef ClusterReference `json:"clusterRef"`
	// userRef contains the list of reference to NifiUsers that are part to the group.
	UsersRef []UserReference `json:"usersRef,omitempty"`
	// accessPolicies defines the list of access policies that will be granted to the group.
	AccessPolicies []AccessPolicy `json:"accessPolicies,omitempty"`
}

// NifiUserGroupStatus defines the observed state of NifiUserGroup.
type NifiUserGroupStatus struct {
	// The nifi usergroup's node id
	Id string `json:"id"`
	// The last nifi usergroup's node revision version catched
	Version int64 `json:"version"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NifiUserGroup is the Schema for the nifiusergroups API.
type NifiUserGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiUserGroupSpec   `json:"spec,omitempty"`
	Status NifiUserGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NifiUserGroupList contains a list of NifiUserGroup.
type NifiUserGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiUserGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiUserGroup{}, &NifiUserGroupList{})
}

func (n NifiUserGroup) GetIdentity() string {
	return fmt.Sprintf("%s-%s", n.Namespace, n.Name)
}
