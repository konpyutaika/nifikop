package v1alpha1

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

// ConvertNifiClusterTo converts a v1alpha1 to v1 (Hub).
func (src *NifiDataflow) ConvertTo(dst conversion.Hub) error {
	ncV1 := dst.(*v1.NifiDataflow)

	if err := ConvertNifiDataflowTo(src, ncV1); err != nil {
		return fmt.Errorf("unable to convert NifiDataflow %s/%s to version: %v, err: %w", src.Namespace, src.Name, dst.GetObjectKind().GroupVersionKind().Version, err)
	}

	return nil
}

// ConvertFrom converts a v1 (Hub) to v1alpha1 (local).
func (dst *NifiDataflow) ConvertFrom(src conversion.Hub) error { //nolint
	ncV1 := src.(*v1.NifiDataflow)
	dst.ObjectMeta = ncV1.ObjectMeta
	if err := ConvertNifiDatflowFrom(dst, ncV1); err != nil {
		return fmt.Errorf("unable to convert NiFiCluster %s/%s from version: %v, err: %w", dst.Namespace, dst.Name, src.GetObjectKind().GroupVersionKind().Version, err)
	}
	return nil
}

// ---- Convert TO ----

// ConvertNifiDataflowTo use to convert v1alpha1.NifiDataflow to v1.NifiDataflow.
func ConvertNifiDataflowTo(src *NifiDataflow, dst *v1.NifiDataflow) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertNifiDataflowSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertNifiDataflowStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertNifiDataflowSpec(src *NifiDataflowSpec, dst *v1.NifiDataflow) error {
	if src == nil {
		return nil
	}

	dst.Spec.ParentProcessGroupID = src.ParentProcessGroupID
	dst.Spec.BucketId = src.BucketId
	dst.Spec.FlowId = src.FlowId
	dst.Spec.FlowVersion = src.FlowVersion
	convertNifiDataflowFlowPosition(src.FlowPosition, dst)
	convertNifiDataflowParameterContextRef(src.ParameterContextRef, dst)
	if src.SyncMode != nil {
		dstSyncMode := v1.DataflowSyncMode(*src.SyncMode)
		dst.Spec.SyncMode = &dstSyncMode
	}
	dst.Spec.SkipInvalidControllerService = src.SkipInvalidControllerService
	dst.Spec.SkipInvalidComponent = src.SkipInvalidComponent
	convertNifiDataflowClusterRef(src.ClusterRef, dst)
	convertNifiDataflowRegistryClientRef(src.RegistryClientRef, dst)
	dst.Spec.UpdateStrategy = v1.ComponentUpdateStrategy(src.UpdateStrategy)

	return nil
}

func convertNifiDataflowFlowPosition(src *FlowPosition, dst *v1.NifiDataflow) {
	if src == nil {
		return
	}
	dst.Spec.FlowPosition = &v1.FlowPosition{}
	if src.X != nil {
		dst.Spec.FlowPosition.X = src.X
	}
	if src.Y != nil {
		dst.Spec.FlowPosition.Y = src.Y
	}
}

func convertNifiDataflowParameterContextRef(src *ParameterContextReference, dst *v1.NifiDataflow) {
	if src == nil {
		return
	}
	dstParameterContextRef := getV1ParameterContextRef(*src)
	dst.Spec.ParameterContextRef = &dstParameterContextRef
}

func convertNifiDataflowClusterRef(src ClusterReference, dst *v1.NifiDataflow) {
	dst.Spec.ClusterRef = getV1ClusterReference(src)
}

func convertNifiDataflowRegistryClientRef(src *RegistryClientReference, dst *v1.NifiDataflow) {
	if src == nil {
		return
	}
	dst.Spec.RegistryClientRef = &v1.RegistryClientReference{
		Name:      src.Name,
		Namespace: src.Namespace,
	}
}

func convertNifiDataflowStatus(src *NifiDataflowStatus, dst *v1.NifiDataflow) error {
	if src == nil {
		return nil
	}
	dst.Status.ProcessGroupID = src.ProcessGroupID
	dst.Status.State = v1.DataflowState(src.State)
	convertNifiDataflowLatestUpdateRequest(src.LatestUpdateRequest, dst)
	convertNifiDataflowLatestDropRequest(src.LatestDropRequest, dst)

	return nil
}

func convertNifiDataflowLatestUpdateRequest(src *UpdateRequest, dst *v1.NifiDataflow) {
	if src == nil {
		return
	}
	dst.Status.LatestUpdateRequest = &v1.UpdateRequest{
		Type:             v1.DataflowUpdateRequestType(src.Type),
		Id:               src.Id,
		Uri:              src.Uri,
		LastUpdated:      src.LastUpdated,
		Complete:         src.Complete,
		FailureReason:    src.FailureReason,
		PercentCompleted: src.PercentCompleted,
		State:            src.State,
	}
}

func convertNifiDataflowLatestDropRequest(src *DropRequest, dst *v1.NifiDataflow) {
	if src == nil {
		return
	}
	dst.Status.LatestDropRequest = &v1.DropRequest{
		ConnectionId:     src.ConnectionId,
		Id:               src.Id,
		Uri:              src.Uri,
		LastUpdated:      src.LastUpdated,
		Finished:         src.Finished,
		FailureReason:    src.FailureReason,
		PercentCompleted: src.PercentCompleted,
		CurrentCount:     src.CurrentCount,
		CurrentSize:      src.CurrentSize,
		Current:          src.Current,
		OriginalCount:    src.OriginalCount,
		OriginalSize:     src.OriginalSize,
		Original:         src.Original,
		DroppedCount:     src.DroppedCount,
		DroppedSize:      src.DroppedSize,
		Dropped:          src.Dropped,
		State:            src.State,
	}
}

// ---- Convert FROM ----

// ConvertNifiDatflowFrom use to convert v1alpha1.NifiCluster from v1.NifiCluster.
func ConvertNifiDatflowFrom(dst *NifiDataflow, src *v1.NifiDataflow) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertFromNifiDataflowSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertFromNifiDataflowStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertFromNifiDataflowSpec(src *v1.NifiDataflowSpec, dst *NifiDataflow) error {
	if src == nil {
		return nil
	}

	dst.Spec.ParentProcessGroupID = src.ParentProcessGroupID
	dst.Spec.BucketId = src.BucketId
	dst.Spec.FlowId = src.FlowId
	dst.Spec.FlowVersion = src.FlowVersion
	convertFromNifiDataflowFlowPosition(src.FlowPosition, dst)
	convertFromNifiDataflowParameterContextRef(src.ParameterContextRef, dst)
	if src.SyncMode != nil {
		dstSyncMode := DataflowSyncMode(*src.SyncMode)
		dst.Spec.SyncMode = &dstSyncMode
	}
	dst.Spec.SkipInvalidControllerService = src.SkipInvalidControllerService
	dst.Spec.SkipInvalidComponent = src.SkipInvalidComponent
	convertFromNifiDataflowClusterRef(src.ClusterRef, dst)
	convertFromNifiDataflowRegistryClientRef(src.RegistryClientRef, dst)
	dst.Spec.UpdateStrategy = ComponentUpdateStrategy(src.UpdateStrategy)

	return nil
}

func convertFromNifiDataflowFlowPosition(src *v1.FlowPosition, dst *NifiDataflow) {
	if src == nil {
		return
	}
	dst.Spec.FlowPosition = &FlowPosition{}
	if src.X != nil {
		dst.Spec.FlowPosition.X = src.X
	}
	if src.Y != nil {
		dst.Spec.FlowPosition.Y = src.Y
	}
}

func convertFromNifiDataflowParameterContextRef(src *v1.ParameterContextReference, dst *NifiDataflow) {
	if src == nil {
		return
	}
	dstParameterContextRef := getParameterContextRef(*src)
	dst.Spec.ParameterContextRef = &dstParameterContextRef
}

func convertFromNifiDataflowClusterRef(src v1.ClusterReference, dst *NifiDataflow) {
	dst.Spec.ClusterRef = getClusterReference(src)
}

func convertFromNifiDataflowRegistryClientRef(src *v1.RegistryClientReference, dst *NifiDataflow) {
	if src == nil {
		return
	}
	dst.Spec.RegistryClientRef = &RegistryClientReference{
		Name:      src.Name,
		Namespace: src.Namespace,
	}
}

func convertFromNifiDataflowStatus(src *v1.NifiDataflowStatus, dst *NifiDataflow) error {
	if src == nil {
		return nil
	}
	dst.Status.ProcessGroupID = src.ProcessGroupID
	dst.Status.State = DataflowState(src.State)
	convertFromNifiDataflowLatestUpdateRequest(src.LatestUpdateRequest, dst)
	convertFromNifiDataflowLatestDropRequest(src.LatestDropRequest, dst)

	return nil
}

func convertFromNifiDataflowLatestUpdateRequest(src *v1.UpdateRequest, dst *NifiDataflow) {
	if src == nil {
		return
	}
	dst.Status.LatestUpdateRequest = &UpdateRequest{
		Type:             DataflowUpdateRequestType(src.Type),
		Id:               src.Id,
		Uri:              src.Uri,
		LastUpdated:      src.LastUpdated,
		Complete:         src.Complete,
		FailureReason:    src.FailureReason,
		PercentCompleted: src.PercentCompleted,
		State:            src.State,
	}
}

func convertFromNifiDataflowLatestDropRequest(src *v1.DropRequest, dst *NifiDataflow) {
	if src == nil {
		return
	}
	dst.Status.LatestDropRequest = &DropRequest{
		ConnectionId:     src.ConnectionId,
		Id:               src.Id,
		Uri:              src.Uri,
		LastUpdated:      src.LastUpdated,
		Finished:         src.Finished,
		FailureReason:    src.FailureReason,
		PercentCompleted: src.PercentCompleted,
		CurrentCount:     src.CurrentCount,
		CurrentSize:      src.CurrentSize,
		Current:          src.Current,
		OriginalCount:    src.OriginalCount,
		OriginalSize:     src.OriginalSize,
		Original:         src.Original,
		DroppedCount:     src.DroppedCount,
		DroppedSize:      src.DroppedSize,
		Dropped:          src.Dropped,
		State:            src.State,
	}
}
