// Copyright 2020 Orange SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package apis

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NifiUserSpec defines the desired state of NifiUser
// +k8s:openapi-gen=true
type NifiUserSpec struct {
	// Name of the secret where all cert resources will be stored
	SecretName string `json:"secretName"`
	// contains the reference to the NifiCluster with the one the user is linked
	ClusterRef ClusterReference `json:"clusterRef"`
	// List of DNSNames that the user will used to request the NifiCluster (allowing to create the right certificates associated)
	DNSNames []string `json:"dnsNames,omitempty"`
	//
	//TopicGrants []UserTopicGrant `json:"topicGrants,omitempty"`

	// Whether or not the the operator also include a Java keystore format (JKS) with you secret
	IncludeJKS bool `json:"includeJKS,omitempty"`
}

// NifiUserStatus defines the observed state of NifiUser
// +k8s:openapi-gen=true
type NifiUserStatus struct {
	State UserState `json:"state"`
	//	ACLs  []string  `json:"acls,omitempty"`
}

// Nifi User is the Schema for the nifi users API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type NifiUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiUserSpec   `json:"spec,omitempty"`
	Status NifiUserStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NifiUserList contains a list of NifiUser
type NifiUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiUser{}, &NifiUserList{})
}
