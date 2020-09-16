package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NifiParameterContextSpec defines the desired state of NifiParameterContext
// +k8s:openapi-gen=true
type NifiParameterContextSpec struct {
	// the Description of the Parameter Context.
	Description string `json:"description,omitempty"`
	// a list of non-sensitive Parameters.
	Parameters []Parameter `json:"parameters"`
	// contains the reference to the NifiCluster with the one the user is linked.
	ClusterRef ClusterReference `json:"clusterRef"`
	// a list of secret containing sensitive parameters (the key will name of the parameter).
	SecretRefs []SecretReference `json:"secretRefs,omitempty"`
}

type Parameter struct {
	// the name of the Parameter.
	Name string `json:"name"`
	// the value of the Parameter.
	Value string `json:"value,omitempty"`
	// the description of the Parameter.
	Description string `json:"description,omitempty"`
}

// NifiParameterContextStatus defines the observed state of NifiParameterContext
// +k8s:openapi-gen=true
type NifiParameterContextStatus struct {
	// the nifi parameter context id.
	Id string `json:"id"`
	// the last nifi parameter context revision version catched.
	Version int64 `json:"version"`
	// the latest update request.
	LatestUpdateRequest *ParameterContextUpdateRequest `json:"latestUpdateRequest,omitempty"`
}

type ParameterContextUpdateRequest struct {
	// the id of the update request.
	Id string `json:"id"`
	// the uri for this request.
	Uri string `json:"uri"`
	// the timestamp of when the request was submitted This property is read only.
	SubmissionTime string `json:"submissionTime"`
	// the last time this request was updated.
	LastUpdated string `json:"lastUpdated"`
	// whether or not this request has completed.
	Complete bool `json:"complete"`
	// an explication of why the request failed, or null if this request has not failed.
	FailureReason string `json:"failureReason"`
	// the percentage complete of the request, between 0 and 100.
	PercentCompleted int32 `json:"percentCompleted"`
	// the state of the request.
	State string `json:"state"`
}

// NifiParameterContext is the Schema for the nifi parameter context API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type NifiParameterContext struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiParameterContextSpec   `json:"spec,omitempty"`
	Status NifiParameterContextStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NifiParameterContextList contains a list of NifiParameterContext
type NifiParameterContextList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiParameterContext `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiParameterContext{}, &NifiParameterContextList{})
}
