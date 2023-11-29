package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiParameterContextSpec defines the desired state of NifiParameterContext.
type NifiParameterContextSpec struct {
	// the Description of the Parameter Context.
	Description string `json:"description,omitempty"`
	// a list of non-sensitive Parameters.
	Parameters []Parameter `json:"parameters"`
	// contains the reference to the NifiCluster with the one the parameter context is linked.
	ClusterRef ClusterReference `json:"clusterRef,omitempty"`
	// a list of secret containing sensitive parameters (the key will name of the parameter).
	SecretRefs []SecretReference `json:"secretRefs,omitempty"`
	// a list of references of Parameter Contexts from which this one inherits parameters
	InheritedParameterContexts []ParameterContextReference `json:"inheritedParameterContexts,omitempty"`
	// whether or not the operator should take over an existing parameter context if its name is the same.
	DisableTakeOver *bool `json:"disableTakeOver,omitempty"`
}

type Parameter struct {
	// the name of the Parameter.
	Name string `json:"name"`
	// the value of the Parameter.
	Value *string `json:"value,omitempty"`
	// the description of the Parameter.
	Description string `json:"description,omitempty"`
	// Whether the parameter is sensitive or not.
	Sensitive bool `json:"sensitive,omitempty"`
}

// NifiParameterContextStatus defines the observed state of NifiParameterContext.
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

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NifiParameterContext is the Schema for the nifiparametercontexts API.
type NifiParameterContext struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiParameterContextSpec   `json:"spec,omitempty"`
	Status NifiParameterContextStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NifiParameterContextList contains a list of NifiParameterContext.
type NifiParameterContextList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiParameterContext `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiParameterContext{}, &NifiParameterContextList{})
}

func (d *NifiParameterContextSpec) IsTakeOverEnabled() bool {
	if d.DisableTakeOver == nil {
		return true
	}
	return !*d.DisableTakeOver
}
