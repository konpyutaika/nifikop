/*
Copyright 2020.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiNodeGroupAutoscalerSpec defines the desired state of NifiNodeGroupAutoscaler.
type NifiNodeGroupAutoscalerSpec struct {
	// contains the reference to the NifiCluster with the one the dataflow is linked.
	ClusterRef v1.ClusterReference `json:"clusterRef"`
	// reference to the nodeConfigGroup that will be set for nodes that are managed and autoscaled
	// This Id is used to compute the names of some Kubernetes resources, so it must be a safe value.
	// +kubebuilder:validation:Pattern:="[a-z0-9]([-a-z0-9]*[a-z0-9])?"
	// +kubebuilder:validation:MaxLength:=63
	NodeConfigGroupId string `json:"nodeConfigGroupId"`
	// A label selector used to identify & manage Node objects in the referenced NifiCluster. Any node matching this
	// selector will be managed by this autoscaler. Even if that node was previously statically defined.
	NodeLabelsSelector *metav1.LabelSelector `json:"nodeLabelsSelector"`
	// the node readOnlyConfig for each node in the node group
	// +optional
	ReadOnlyConfig *v1.ReadOnlyConfig `json:"readOnlyConfig,omitempty"`
	// the nodeConfig to use for each node in the node group. This will be merged with and is preferred to the configured
	// nodeConfigGroupId
	// +optional
	NodeConfig *v1.NodeConfig `json:"nodeConfig,omitempty"`
	// current number of replicas expected for the node config group
	// +kubebuilder:default:=0
	// +optional
	Replicas int32 `json:"replicas"`
	// The strategy to use when scaling up the nifi cluster
	// +kubebuilder:validation:Enum={"graceful","simple"}
	UpscaleStrategy ClusterScalingStrategy `json:"upscaleStrategy,omitempty"`
	// The strategy to use when scaling down the nifi cluster
	// +kubebuilder:validation:Enum={"lifo","nonprimary","leastbusy"}
	DownscaleStrategy ClusterScalingStrategy `json:"downscaleStrategy,omitempty"`
}

// NifiNodeGroupAutoscalerStatus defines the observed state of NifiNodeGroupAutoscaler.
type NifiNodeGroupAutoscalerStatus struct {
	// The state of this autoscaler
	State NodeGroupAutoscalerState `json:"state"`
	// the current number of replicas in this cluster
	Replicas ClusterReplicas `json:"replicas"`
	// label selectors for cluster child pods. HPA uses this to identify pod replicas
	Selector ClusterReplicaSelector `json:"selector"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector

// NifiNodeGroupAutoscaler is the Schema for the nifinodegroupautoscalers API.
type NifiNodeGroupAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiNodeGroupAutoscalerSpec   `json:"spec,omitempty"`
	Status NifiNodeGroupAutoscalerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NifiNodeGroupAutoscalerList contains a list of NifiNodeGroupAutoscaler.
type NifiNodeGroupAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiNodeGroupAutoscaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiNodeGroupAutoscaler{}, &NifiNodeGroupAutoscalerList{})
}

func (aSpec *NifiNodeGroupAutoscalerSpec) NifiNodeGroupSelectorAsMap() (map[string]string, error) {
	labels, err := metav1.LabelSelectorAsMap(aSpec.NodeLabelsSelector)
	if err != nil {
		return nil, err
	}
	return labels, nil
}
