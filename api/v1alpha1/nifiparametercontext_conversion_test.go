package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

func TestNifiParameterContextConversion(t *testing.T) {
	pc := createParameterContext()

	// convert v1alpha1 to v1
	v1pc := &v1.NifiParameterContext{}
	pc.ConvertTo(v1pc)
	assertParameterContextsEqual(pc, v1pc, t)

	// convert v1 to v1alpha1
	newPc := &NifiParameterContext{}
	newPc.ConvertFrom(v1pc)
	assertParameterContextsEqual(newPc, v1pc, t)
}

func assertParameterContextsEqual(pc *NifiParameterContext, v1pc *v1.NifiParameterContext, t *testing.T) {
	if pc.ObjectMeta.Name != v1pc.ObjectMeta.Name ||
		pc.ObjectMeta.Namespace != v1pc.ObjectMeta.Namespace {
		t.Error("Object Metas are not equal")
	}
	if pc.Spec.ClusterRef.Name != v1pc.Spec.ClusterRef.Name ||
		pc.Spec.ClusterRef.Namespace != v1pc.Spec.ClusterRef.Namespace {
		t.Error("cluster refs are not equal")
	}
	if pc.Spec.Description != v1pc.Spec.Description {
		t.Error("descriptions are not equal")
	}
	if pc.Spec.DisableTakeOver != v1pc.Spec.DisableTakeOver {
		t.Error("disable takeover not equal")
	}
	if !secretRefsEqual(pc.Spec.SecretRefs, v1pc.Spec.SecretRefs) {
		t.Error("Secret refs not equal")
	}
	if !inheritedParameterContextsEqual(pc.Spec.InheritedParameterContexts, v1pc.Spec.InheritedParameterContexts) {
		t.Error("Inherited parameter contexts not equal")
	}
	assertPCStatusesEqual(pc.Status, v1pc.Status, t)
}

func assertPCStatusesEqual(s1 NifiParameterContextStatus, s2 v1.NifiParameterContextStatus, t *testing.T) {
	if s1.Id != s2.Id {
		t.Error("status IDs not equal")
	}
	if s1.Version != s2.Version {
		t.Error("status versions not equal")
	}
	if s1.LatestUpdateRequest.Id != s2.LatestUpdateRequest.Id ||
		s1.LatestUpdateRequest.Complete != s2.LatestUpdateRequest.Complete ||
		s1.LatestUpdateRequest.FailureReason != s2.LatestUpdateRequest.FailureReason ||
		s1.LatestUpdateRequest.LastUpdated != s2.LatestUpdateRequest.LastUpdated ||
		s1.LatestUpdateRequest.PercentCompleted != s2.LatestUpdateRequest.PercentCompleted ||
		s1.LatestUpdateRequest.SubmissionTime != s2.LatestUpdateRequest.SubmissionTime ||
		s1.LatestUpdateRequest.State != s2.LatestUpdateRequest.State ||
		s1.LatestUpdateRequest.Uri != s2.LatestUpdateRequest.Uri {
		t.Error("status latest update requests not equal")
	}
}

func secretRefsEqual(sr1 []SecretReference, sr2 []v1.SecretReference) bool {
	if len(sr1) != len(sr2) {
		return false
	}

	for i, pc := range sr1 {
		if pc.Name != sr2[i].Name ||
			pc.Namespace != sr2[i].Namespace {
			return false
		}
	}
	return true
}

func inheritedParameterContextsEqual(ipc1 []ParameterContextReference, ipc2 []v1.ParameterContextReference) bool {
	if len(ipc1) != len(ipc2) {
		return false
	}

	for i, pc := range ipc1 {
		if pc.Name != ipc2[i].Name ||
			pc.Namespace != ipc2[i].Namespace {
			return false
		}
	}
	return true
}

func createParameterContext() *NifiParameterContext {
	var param string = "blah"
	var takeover bool = true
	return &NifiParameterContext{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "parameterContext",
			Namespace: "namespace",
		},
		Spec: NifiParameterContextSpec{
			Description: "description",
			Parameters: []Parameter{
				{
					Name:        "parameter",
					Value:       &param,
					Sensitive:   true,
					Description: "description",
				},
			},
			ClusterRef: ClusterReference{
				Name:      "cluster",
				Namespace: "namespace",
			},
			SecretRefs: []SecretReference{
				{
					Name:      "secret",
					Namespace: "namespace",
				},
			},
			InheritedParameterContexts: []ParameterContextReference{
				{
					Name:      "paramContextRef",
					Namespace: "namespace",
				},
			},
			DisableTakeOver: &takeover,
		},
		Status: NifiParameterContextStatus{
			Id:      "id",
			Version: 6,
			LatestUpdateRequest: &ParameterContextUpdateRequest{
				Id:               "id",
				Uri:              "uri",
				SubmissionTime:   "subTime",
				LastUpdated:      "lastUpdated",
				Complete:         true,
				FailureReason:    "reason",
				PercentCompleted: 5,
				State:            "state",
			},
		},
	}
}
