package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiUserSpec defines the desired state of NifiUser.
type NifiUserSpec struct {
	// identity field is used to define the user identity on NiFi cluster side, when the user's name doesn't
	// suit with Kubernetes resource name.
	Identity string `json:"identity,omitempty"`
	// Name of the secret where all cert resources will be stored
	SecretName string `json:"secretName,omitempty"`
	// contains the reference to the NifiCluster with the one the user is linked
	ClusterRef ClusterReference `json:"clusterRef"`
	// List of DNSNames that the user will used to request the NifiCluster (allowing to create the right certificates associated)
	DNSNames []string `json:"dnsNames,omitempty"`
	// Whether or not the the operator also include a Java keystore format (JKS) with you secret
	IncludeJKS bool `json:"includeJKS,omitempty"`
	// Whether or not a certificate will be created for this user.
	CreateCert *bool `json:"createCert,omitempty"`
	// accessPolicies defines the list of access policies that will be granted to the group.
	AccessPolicies []AccessPolicy `json:"accessPolicies,omitempty"`
}

// NifiUserStatus defines the observed state of NifiUser.
type NifiUserStatus struct {
	// The nifi user's node id
	Id string `json:"id"`
	// The last nifi  user's node revision version catched
	Version int64 `json:"version"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// NifiUser is the Schema for the nifiusers API.
type NifiUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiUserSpec   `json:"spec,omitempty"`
	Status NifiUserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NifiUserList contains a list of NifiUser.
type NifiUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiUser{}, &NifiUserList{})
}

func (u *NifiUserSpec) GetCreateCert() bool {
	if u.CreateCert != nil {
		return *u.CreateCert
	}
	return true
}

func (u *NifiUser) GetIdentity() string {
	if u.Spec.Identity == "" {
		return u.Name
	}
	return u.Spec.Identity
}
