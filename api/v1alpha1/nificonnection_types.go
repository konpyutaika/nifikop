package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiConnectionSpec defines the desired state of NifiConnection
type NifiConnectionSpec struct {
	// the Source component of the connection.
	Source ComponentReference `json:"source"`
	// the Destination component of the connection.
	Destination ComponentReference `json:"destination"`
	// the Configuration of the connection.
	Configuration ConnectionConfiguration `json:"configuration,omitempty"`
	// describes the way the operator will deal with data when a connection will be updated : drop or drain.
	UpdateStrategy ComponentUpdateStrategy `json:"updateStrategy"`
}

type ComponentReference struct {
	// the name of the component.
	Name string `json:"name"`
	// the namespace of the component.
	Namespace string `json:"namespace,omitempty"`
	// the type of the component (e.g. nifidataflow).
	Type ComponentType `json:"type"`
	// the name of the sub component (e.g. queue or port name).
	SubName string `json:"subName,omitempty"`
}

type ConnectionConfiguration struct {
	// the maximum amount of time an object may be in the flow before it will be automatically aged out of the flow.
	FlowFileExpiration string `json:"flowFileExpiration,omitempty"`
	// the maximum data size of objects that can be queued before back pressure is applied.
	BackPressureDataSizeThreshold string `json:"backPressureDataSizeThreshold,omitempty"`
	// the maximum number of objects that can be queued before back pressure is applied.
	BackPressureObjectThreshold *int64 `json:"backPressureObjectThreshold,omitempty"`
	// how to load balance the data in this Connection across the nodes in the cluster.
	LoadBalanceStrategy ConnectionLoadBalanceStrategy `json:"loadBalanceStrategy,omitempty"`
	// the FlowFile Attribute to use for determining which node a FlowFile will go to.
	LoadBalancePartitionAttribute string `json:"loadBalancePartitionAttribute,omitempty"`
	// whether or not data should be compressed when being transferred between nodes in the cluster.
	LoadBalanceCompression ConnectionLoadBalanceCompression `json:"loadBalanceCompression,omitempty"`
	// the comparators used to prioritize the queue.
	Prioritizers []ConnectionPrioritizer `json:"prioritizers,omitempty"`
	// the index of the bend point where to place the connection label.
	LabelIndex *int32 `json:"labelIndex,omitempty"`
	// the bend points on the connection.
	Bends []ConnectionBend `json:"bends,omitempty"`
}

type ConnectionBend struct {
	// The x coordinate.
	X *int64 `json:"posX,omitempty"`
	// The y coordinate.
	Y *int64 `json:"posY,omitempty"`
}

// NifiConnectionStatus defines the observed state of NifiConnection
type NifiConnectionStatus struct {
	// connection ID.
	ConnectionId string `json:"connectionID"`
	// connection current state.
	State ConnectionState `json:"state"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NifiConnection is the Schema for the nificonnections API
type NifiConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiConnectionSpec   `json:"spec,omitempty"`
	Status NifiConnectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NifiConnectionList contains a list of NifiConnection
type NifiConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiConnection{}, &NifiConnectionList{})
}

func (nCon *NifiConnectionSpec) IsValid() bool {
	return nCon.Source.IsValid() && nCon.Destination.IsValid() && nCon.Configuration.IsValid()
}

func (ref *ComponentReference) IsValid() bool {
	return ref.Type == ComponentDataflow && ref.SubName != ""
}

func (conf *ConnectionConfiguration) IsValid() bool {
	if conf.GetLoadBalanceStrategy() == StrategyPartitionByAttribute && len(conf.GetLoadBalancePartitionAttribute()) == 0 {
		return false
	}
	return true
}

func (conf *ConnectionConfiguration) GetFlowFileExpiration() string {
	return conf.FlowFileExpiration
}

func (conf *ConnectionConfiguration) GetBackPressureDataSizeThreshold() string {
	if len(conf.BackPressureDataSizeThreshold) > 0 {
		return conf.BackPressureDataSizeThreshold
	}
	return "1 GB"
}

func (conf *ConnectionConfiguration) GetBackPressureObjectThreshold() int64 {
	if conf.BackPressureObjectThreshold != nil {
		return *conf.BackPressureObjectThreshold
	}
	return 10000
}

func (conf *ConnectionConfiguration) GetLoadBalanceStrategy() ConnectionLoadBalanceStrategy {
	if len(conf.LoadBalanceStrategy) > 0 {
		return conf.LoadBalanceStrategy
	}
	return StrategyDoNotLoadBalance
}

func (conf *ConnectionConfiguration) GetLoadBalancePartitionAttribute() string {
	return conf.LoadBalancePartitionAttribute
}

func (conf *ConnectionConfiguration) GetLoadBalanceCompression() ConnectionLoadBalanceCompression {
	if len(conf.LoadBalanceCompression) > 0 {
		return conf.LoadBalanceCompression
	}
	return CompressionDoNotCompress
}

func (conf *ConnectionConfiguration) GetPrioritizers() []ConnectionPrioritizer {
	return conf.Prioritizers
}

func (conf *ConnectionConfiguration) GetStringPrioritizers() []string {
	var prefix string = "org.apache.nifi.prioritizer."
	prioritizers := []string{}
	for _, prioritizer := range conf.Prioritizers {
		prioritizers = append(prioritizers, prefix+string(prioritizer))
	}
	return prioritizers
}

func (conf *ConnectionConfiguration) GetLabelIndex() int32 {
	if conf.LabelIndex != nil {
		return *conf.LabelIndex
	}
	return 0
}

func (conf *ConnectionConfiguration) GetBends() []ConnectionBend {
	return conf.Bends
}
