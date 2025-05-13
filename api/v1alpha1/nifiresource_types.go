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
	"errors"
	"fmt"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	k8sv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NifiResourceSpec defines the desired state of NifiResource
type NifiResourceSpec struct {
	Type v1.ResourceType `json:"type"`
	// the name of the Resource that will be deployed
	Name string `json:"name"`
	// the UUID of the parent process group where you want to deploy your dataflow, if not set deploy at root level.
	ParentProcessGroupID string `json:"parentProcessGroupID,omitempty"`
	// contains the reference to the NifiCluster with the one the dataflow is linked.
	ClusterRef v1.ClusterReference `json:"clusterRef,omitempty"`
	// the comments added to the resource being deployed
	Comments string `json:"comments,omitempty"`
	// other configuration specific to the Type
	Configuration k8sv1.JSON `json:"configuration,omitempty"`
	// contains the reference to the NifiResource which is the Parent Process Group for this dataflow
	ParentProcessGroupReference *v1.ResourceReference `json:"parentProcessGroupRef,omitempty"`
}

func (d NifiResourceSpec) GetConfiguration() (map[string]interface{}, error) {
	var configMap map[string]interface{}
	if err := json.Unmarshal(d.Configuration.Raw, &configMap); err != nil {
		return nil, errors.New("failed to unmarshal Configuration")
	} else {
		return configMap, nil
	}
}

// NifiResourceStatus defines the observed state of NifiResource
type NifiResourceStatus struct {
	// Objects uuid
	UUID string `json:"uuid"`
	// the current state.
	State ResourceState `json:"state"`
	// the latest queue drop request sent.
	LatestDropRequest *v1.DropRequest `json:"latestDropRequest,omitempty"`
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

func (d *NifiResourceSpec) GetParentProcessGroupID(rootProcessGroupId string, parentProcessGroup *NifiResource) string {
	if parentProcessGroup != nil && parentProcessGroup.Spec.Type == v1.ResourceProcessGroup && parentProcessGroup.Status.UUID != "" {
		return parentProcessGroup.Status.UUID
	} else if d.ParentProcessGroupID == "" {
		return rootProcessGroupId
	} else {
		return d.ParentProcessGroupID
	}
}

func ExtractParameterContextReference(resourceSpec *NifiResourceSpec) (*v1.ParameterContextReference, error) {

	var configMap map[string]interface{}
	if err := json.Unmarshal(resourceSpec.Configuration.Raw, &configMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Configuration: %w", err)
	}

	raw, ok := configMap["parameterContextRef"]
	if !ok || raw == nil {
		return nil, errors.New("parameterContextRef not found or nil")
	}

	// Step 1: assert it's a map (optional, for safety)
	paramMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("parameterContextRef is not a map")
	}

	// Step 2: marshal to JSON
	rawBytes, err := json.Marshal(paramMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameterContextRef: %w", err)
	}

	// Step 3: unmarshal to typed struct
	var parameterContextReference v1.ParameterContextReference
	if err := json.Unmarshal(rawBytes, &parameterContextReference); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameterContextRef: %w", err)
	}

	return &parameterContextReference, nil
}

func init() {
	SchemeBuilder.Register(&NifiResource{}, &NifiResourceList{})
}
