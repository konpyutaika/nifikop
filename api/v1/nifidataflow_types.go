package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiDataflowSpec defines the desired state of NifiDataflow.
type NifiDataflowSpec struct {
	// the UUID of the parent process group where you want to deploy your dataflow, if not set deploy at root level.
	ParentProcessGroupID string `json:"parentProcessGroupID,omitempty"`
	// the UUID of the Bucket containing the flow.
	BucketId string `json:"bucketId"`
	// the UUID of the flow to run.
	FlowId string `json:"flowId"`
	// the version of the flow to run, then the latest version of flow will be used.
	FlowVersion *int32 `json:"flowVersion,omitempty"`
	// the position of your dataflow in the canvas.
	FlowPosition *FlowPosition `json:"flowPosition,omitempty"`
	// contains the reference to the ParameterContext with the one the dataflow is linked.
	ParameterContextRef *ParameterContextReference `json:"parameterContextRef,omitempty"`
	// if the flow will be synchronized once, continuously or never
	SyncMode *DataflowSyncMode `json:"syncMode,omitempty"`
	// whether the flow is considered as ran if some controller services are still invalid or not.
	SkipInvalidControllerService bool `json:"skipInvalidControllerService,omitempty"`
	// whether the flow is considered as ran if some components are still invalid or not.
	SkipInvalidComponent bool `json:"skipInvalidComponent,omitempty"`
	// contains the reference to the NifiCluster with the one the dataflow is linked.
	ClusterRef ClusterReference `json:"clusterRef,omitempty"`
	// contains the reference to the NifiRegistry with the one the dataflow is linked.
	RegistryClientRef *RegistryClientReference `json:"registryClientRef,omitempty"`
	// describes the way the operator will deal with data when a dataflow will be updated : drop or drain
	UpdateStrategy ComponentUpdateStrategy `json:"updateStrategy"`
}

type FlowPosition struct {
	// The x coordinate.
	X *int64 `json:"posX,omitempty"`
	// The y coordinate.
	Y *int64 `json:"posY,omitempty"`
}

type UpdateRequest struct {
	// defines the type of versioned flow update request.
	Type DataflowUpdateRequestType `json:"type"`
	// the id of the update request.
	Id string `json:"id"`
	// the uri for this request.
	Uri string `json:"uri"`
	// the last time this request was updated.
	LastUpdated string `json:"lastUpdated"`
	// whether or not this request has completed.
	Complete bool `json:"complete"`
	// an explication of why the request failed, or null if this request has not failed.
	FailureReason string `json:"failureReason"`
	// the percentage complete of the request, between 0 and 100.
	PercentCompleted int32 `json:"percentCompleted"`
	// the state of the request
	State string `json:"state"`
	// whether or not this request was found.
	NotFound bool `json:"notFound,omitempty"`
	// the number of consecutive retries made in case of a NotFound error (limit: 3).
	NotFoundRetryCount int32 `json:"notFoundRetryCount,omitempty"`
}

type DropRequest struct {
	// the connection id.
	ConnectionId string `json:"connectionId"`
	// the id for this drop request.
	Id string `json:"id"`
	// the uri for this request.
	Uri string `json:"uri"`
	// the last time this request was updated.
	LastUpdated string `json:"lastUpdated"`
	// whether the request has finished.
	Finished bool `json:"finished"`
	// an explication of why the request failed, or null if this request has not failed.
	FailureReason string `json:"failureReason"`
	// the percentage complete of the request, between 0 and 100.
	PercentCompleted int32 `json:"percentCompleted"`
	// the number of flow files currently queued.
	CurrentCount int32 `json:"currentCount"`
	// the size of flow files currently queued in bytes.
	CurrentSize int64 `json:"currentSize"`
	// the count and size of flow files currently queued.
	Current string `json:"current"`
	// the number of flow files to be dropped as a result of this request.
	OriginalCount int32 `json:"originalCount"`
	// the size of flow files to be dropped as a result of this request in bytes.
	OriginalSize int64 `json:"originalSize"`
	// the count and size of flow files to be dropped as a result of this request.
	Original string `json:"original"`
	// the number of flow files that have been dropped thus far.
	DroppedCount int32 `json:"droppedCount"`
	// the size of flow files currently queued in bytes.
	DroppedSize int64 `json:"droppedSize"`
	// the count and size of flow files that have been dropped thus far.
	Dropped string `json:"dropped"`
	// the state of the request
	State string `json:"state"`
	// whether or not this request was found.
	NotFound bool `json:"notFound,omitempty"`
	// the number of consecutive retries made in case of a NotFound error (limit: 3).
	NotFoundRetryCount int32 `json:"notFoundRetryCount,omitempty"`
}

// NifiDataflowStatus defines the observed state of NifiDataflow.
type NifiDataflowStatus struct {
	// process Group ID
	ProcessGroupID string `json:"processGroupID"`
	// the dataflow current state.
	State DataflowState `json:"state"`
	// the latest version update request sent.
	LatestUpdateRequest *UpdateRequest `json:"latestUpdateRequest,omitempty"`
	// the latest queue drop request sent.
	LatestDropRequest *DropRequest `json:"latestDropRequest,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// NifiDataflow is the Schema for the nifidataflows API.
type NifiDataflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiDataflowSpec   `json:"spec,omitempty"`
	Status NifiDataflowStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NifiDataflowList contains a list of NifiDataflow.
type NifiDataflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiDataflow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiDataflow{}, &NifiDataflowList{})
}

func (d *NifiDataflowSpec) GetSyncMode() DataflowSyncMode {
	if d.SyncMode == nil {
		return SyncAlways
	}
	return *d.SyncMode
}

func (d *NifiDataflowSpec) SyncOnce() bool {
	return d.GetSyncMode() == SyncOnce
}

func (d *NifiDataflowSpec) SyncAlways() bool {
	return d.GetSyncMode() == SyncAlways
}

func (d *NifiDataflowSpec) SyncNever() bool {
	return d.GetSyncMode() == SyncNever
}

func (d *NifiDataflowSpec) GetParentProcessGroupID(rootProcessGroupId string) string {
	if d.ParentProcessGroupID == "" {
		return rootProcessGroupId
	}
	return d.ParentProcessGroupID
}

func (p *FlowPosition) GetX() int64 {
	if p.X == nil || *p.X == 0 {
		return 1
	}
	return *p.X
}

func (p *FlowPosition) GetY() int64 {
	if p.Y == nil || *p.Y == 0 {
		return 1
	}
	return *p.Y
}
