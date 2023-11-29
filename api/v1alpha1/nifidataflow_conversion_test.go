package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

func TestNifiDataflowConversion(t *testing.T) {
	df := createNifiDataflow()

	// convert v1alpha1 to v1
	v1df := &v1.NifiDataflow{}
	df.ConvertTo(v1df)
	assertDataflowsEqual(df, v1df, t)

	// convert v1 to v1alpha1
	newDataflow := &NifiDataflow{}
	newDataflow.ConvertFrom(v1df)
	assertDataflowsEqual(newDataflow, v1df, t)
}

func assertDataflowsEqual(df *NifiDataflow, v1df *v1.NifiDataflow, t *testing.T) {
	if !reflect.DeepEqual(df.ObjectMeta.Annotations, v1df.ObjectMeta.Annotations) ||
		df.ObjectMeta.Name != v1df.ObjectMeta.Name ||
		df.ObjectMeta.Namespace != v1df.ObjectMeta.Namespace ||
		!reflect.DeepEqual(df.ObjectMeta.Labels, v1df.ObjectMeta.Labels) {
		t.Error("object metas are not equal")
	}

	if df.Spec.ParentProcessGroupID != v1df.Spec.ParentProcessGroupID {
		t.Error("parent process group ids are not equal")
	}
	if df.Spec.BucketId != v1df.Spec.BucketId {
		t.Error("bucket ids are not equal")
	}
	if df.Spec.FlowId != v1df.Spec.FlowId {
		t.Error("flow ids are not equal")
	}
	if df.Spec.FlowVersion != v1df.Spec.FlowVersion {
		t.Error("flow versions are not equal")
	}
	if df.Spec.FlowPosition.X != v1df.Spec.FlowPosition.X ||
		df.Spec.FlowPosition.Y != v1df.Spec.FlowPosition.Y {
		t.Error("Flow positions are not equal")
	}
	if df.Spec.ParameterContextRef.Name != v1df.Spec.ParameterContextRef.Name ||
		df.Spec.ParameterContextRef.Namespace != v1df.Spec.ParameterContextRef.Namespace {
		t.Error("parameter context refs are not equal")
	}
	if df.Spec.ClusterRef.Name != v1df.Spec.ClusterRef.Name ||
		df.Spec.ClusterRef.Namespace != v1df.Spec.ClusterRef.Namespace {
		t.Error("cluster refs are not equal")
	}
	if df.Spec.RegistryClientRef.Name != v1df.Spec.RegistryClientRef.Name ||
		df.Spec.RegistryClientRef.Namespace != v1df.Spec.RegistryClientRef.Namespace {
		t.Error("registry client refs are not equal")
	}
	if string(*df.Spec.SyncMode) != string(*v1df.Spec.SyncMode) {
		t.Error("sync modes are not equal")
	}
	if df.Spec.SkipInvalidComponent != v1df.Spec.SkipInvalidComponent {
		t.Error("skip invalid components are not equal")
	}
	if df.Spec.SkipInvalidControllerService != v1df.Spec.SkipInvalidControllerService {
		t.Error("skip invalid controller services are not equal")
	}
	if string(df.Spec.UpdateStrategy) != string(v1df.Spec.UpdateStrategy) {
		t.Error("update strategies are not equal")
	}
	assertStatusesEqual(df.Status, v1df.Status, t)
}

func assertStatusesEqual(dfs NifiDataflowStatus, v1dfs v1.NifiDataflowStatus, t *testing.T) {
	if string(dfs.State) != string(v1dfs.State) {
		t.Error("dataflow states not equal")
	}
	if dfs.ProcessGroupID != v1dfs.ProcessGroupID {
		t.Error("status process group ids are not equal")
	}
	if dfs.LatestUpdateRequest.Complete != v1dfs.LatestUpdateRequest.Complete ||
		dfs.LatestUpdateRequest.FailureReason != v1dfs.LatestUpdateRequest.FailureReason ||
		dfs.LatestUpdateRequest.Id != v1dfs.LatestDropRequest.Id ||
		dfs.LatestUpdateRequest.LastUpdated != v1dfs.LatestUpdateRequest.LastUpdated ||
		dfs.LatestUpdateRequest.PercentCompleted != v1dfs.LatestDropRequest.PercentCompleted ||
		dfs.LatestUpdateRequest.State != v1dfs.LatestUpdateRequest.State ||
		string(dfs.LatestUpdateRequest.Type) != string(v1dfs.LatestUpdateRequest.Type) ||
		dfs.LatestUpdateRequest.Uri != v1dfs.LatestUpdateRequest.Uri {
		t.Error("status latest update requests are not equal")
	}
	if dfs.LatestDropRequest.ConnectionId != v1dfs.LatestDropRequest.ConnectionId ||
		dfs.LatestDropRequest.Current != v1dfs.LatestDropRequest.Current ||
		dfs.LatestDropRequest.CurrentCount != v1dfs.LatestDropRequest.CurrentCount ||
		dfs.LatestDropRequest.CurrentSize != v1dfs.LatestDropRequest.CurrentSize ||
		dfs.LatestDropRequest.Dropped != v1dfs.LatestDropRequest.Dropped ||
		dfs.LatestDropRequest.DroppedCount != v1dfs.LatestDropRequest.DroppedCount ||
		dfs.LatestDropRequest.DroppedSize != v1dfs.LatestDropRequest.DroppedSize ||
		dfs.LatestDropRequest.FailureReason != v1dfs.LatestDropRequest.FailureReason ||
		dfs.LatestDropRequest.Finished != v1dfs.LatestDropRequest.Finished ||
		dfs.LatestDropRequest.Id != v1dfs.LatestDropRequest.Id ||
		dfs.LatestDropRequest.LastUpdated != v1dfs.LatestDropRequest.LastUpdated ||
		dfs.LatestDropRequest.Original != v1dfs.LatestDropRequest.Original ||
		dfs.LatestDropRequest.OriginalCount != v1dfs.LatestDropRequest.OriginalCount ||
		dfs.LatestDropRequest.OriginalSize != v1dfs.LatestDropRequest.OriginalSize ||
		dfs.LatestDropRequest.PercentCompleted != v1dfs.LatestDropRequest.PercentCompleted ||
		dfs.LatestDropRequest.Uri != v1dfs.LatestDropRequest.Uri ||
		dfs.LatestDropRequest.State != v1dfs.LatestDropRequest.State {
		t.Error("status latest drop requests are not equal")
	}
}

func createNifiDataflow() *NifiDataflow {
	var ver int32 = 5
	var pos int64 = 3
	var syncMode DataflowSyncMode = SyncAlways
	return &NifiDataflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dataflow",
			Namespace: "namespace",
		},
		Spec: NifiDataflowSpec{
			ParentProcessGroupID: "parentID",
			BucketId:             "bucketId",
			FlowId:               "flowId",
			FlowVersion:          &ver,
			FlowPosition: &FlowPosition{
				X: &pos,
				Y: &pos,
			},
			ParameterContextRef: &ParameterContextReference{
				Name:      "pcr",
				Namespace: "namespace",
			},
			SyncMode:                     &syncMode,
			SkipInvalidControllerService: true,
			SkipInvalidComponent:         true,
			ClusterRef: ClusterReference{
				Name:      "cluster",
				Namespace: "namespace",
			},
			RegistryClientRef: &RegistryClientReference{
				Name:      "registry",
				Namespace: "namespace",
			},
			UpdateStrategy: DrainStrategy,
		},
		Status: NifiDataflowStatus{
			ProcessGroupID: "processGroupId",
			State:          DataflowStateCreated,
			LatestUpdateRequest: &UpdateRequest{
				Type:             RevertRequestType,
				Id:               "id",
				Uri:              "uri",
				LastUpdated:      "lastUpdated",
				Complete:         true,
				FailureReason:    "reason",
				PercentCompleted: 5,
				State:            "state",
			},
			LatestDropRequest: &DropRequest{
				ConnectionId:     "connId",
				Id:               "id",
				Uri:              "uri",
				LastUpdated:      "lastUpdated",
				Finished:         true,
				FailureReason:    "reason",
				PercentCompleted: 5,
				CurrentCount:     6,
				CurrentSize:      8,
				Current:          "current",
				OriginalCount:    7,
				OriginalSize:     100,
				Original:         "original",
				DroppedCount:     1,
				DroppedSize:      2,
				Dropped:          "dropped",
				State:            "state",
			},
		},
	}
}
