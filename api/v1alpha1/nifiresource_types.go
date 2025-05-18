/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"encoding/json"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiResourceSpec defines the desired state of NifiResource
type NifiResourceSpec struct {
	// contains the reference to the NifiCluster with the one the parameter context is linked.
	ClusterRef v1.ClusterReference `json:"clusterRef,omitempty"`
	// the type of the resource (e.g. process-group).
	Type ResourceType `json:"type"`
	// the UUID of the parent process group where you want to deploy your resource, if not set deploy at root level (is not used for all types of resource).
	ParentProcessGroupID string `json:"parentProcessGroupID,omitempty"`
	// the reference to the parent process group where you want to deploy your resource, if not set deploy at root level (is not used for all types of resource).
	ParentProcessGroupRef *v1.ResourceReference `json:"parentProcessGroupRef,omitempty"`
	// the name of the resource (if not set, the name of the CR will be used).
	DisplayName string `json:"displayName,omitempty"`
	// the configuration of the resource (e.g. the process group configuration).
	Configuration runtime.RawExtension `json:"configuration,omitempty"`
}

// NifiResourceStatus defines the observed state of NifiResource
type NifiResourceStatus struct {
	// The nifi resource's id
	Id string `json:"id"`
	// The last nifi resource revision version catched
	Version int64 `json:"version"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NifiResource is the Schema for the nifiresources API
type NifiResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiResourceSpec   `json:"spec,omitempty"`
	Status NifiResourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NifiResourceList contains a list of NifiResource
type NifiResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiResource{}, &NifiResourceList{})
}

func (r *NifiResource) GetDisplayName() string {
	if r.Spec.DisplayName != "" {
		return r.Spec.DisplayName
	}
	return r.Name
}

func (r *NifiResourceSpec) GetParentProcessGroupID(rootProcessGroupId string) string {
	if r.ParentProcessGroupID == "" {
		return rootProcessGroupId
	}
	return r.ParentProcessGroupID
}

func (r *NifiResourceSpec) GetConfiguration() (map[string]interface{}, error) {
	var configuration map[string]interface{}
	if err := json.Unmarshal(r.Configuration.Raw, &configuration); err != nil {
		return nil, err
	}
	return configuration, nil
}

func (r *NifiResourceSpec) IsProcessGroup() bool {
	return r.Type == ResourceProcessGroup
}
