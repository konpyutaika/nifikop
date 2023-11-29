package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

func TestNifiUserGroupConversion(t *testing.T) {
	ug := createNifiUserGroup()

	// convert v1alpha1 to v1
	v1ug := &v1.NifiUserGroup{}
	ug.ConvertTo(v1ug)
	assertUserGroupsEqual(ug, v1ug, t)

	// convert v1 to v1alpha1
	newUg := &NifiUserGroup{}
	newUg.ConvertFrom(v1ug)
	assertUserGroupsEqual(newUg, v1ug, t)
}

func assertUserGroupsEqual(ug *NifiUserGroup, v1ug *v1.NifiUserGroup, t *testing.T) {
	if ug.ObjectMeta.Name != v1ug.ObjectMeta.Name ||
		ug.ObjectMeta.Namespace != v1ug.ObjectMeta.Namespace {
		t.Error("object metas not equal")
	}
	if ug.Spec.ClusterRef.Name != v1ug.Spec.ClusterRef.Name ||
		ug.Spec.ClusterRef.Namespace != v1ug.Spec.ClusterRef.Namespace {
		t.Error("cluster refs not equal")
	}
	if !userRefsEqual(ug.Spec.UsersRef, v1ug.Spec.UsersRef) {
		t.Error("user refs not equal")
	}
	if !accessPoliciesEqual(ug.Spec.AccessPolicies, v1ug.Spec.AccessPolicies) {
		t.Error("access policies not equal")
	}

	if ug.Status.Id != v1ug.Status.Id {
		t.Error("status ids not equal")
	}
	if ug.Status.Version != v1ug.Status.Version {
		t.Error("status versions not equal")
	}
}

func userRefsEqual(ur1 []UserReference, ur2 []v1.UserReference) bool {
	if len(ur1) != len(ur2) {
		return false
	}
	for i, ur := range ur1 {
		if ur.Name != ur2[i].Name || ur.Namespace != ur2[i].Namespace {
			return false
		}
	}
	return true
}

func createNifiUserGroup() *NifiUserGroup {
	return &NifiUserGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "userGroup",
			Namespace: "namespace",
		},
		Spec: NifiUserGroupSpec{
			ClusterRef: ClusterReference{
				Name:      "cluster",
				Namespace: "namespace",
			},
			UsersRef: []UserReference{
				{
					Name:      "user",
					Namespace: "namespace",
				},
			},
			AccessPolicies: []AccessPolicy{
				{
					Type:          ComponentAccessPolicyType,
					Action:        ReadAccessPolicyAction,
					Resource:      ComponentsAccessPolicyResource,
					ComponentType: "type",
					ComponentId:   "id",
				},
			},
		},
		Status: NifiUserGroupStatus{
			Id:      "id",
			Version: 6,
		},
	}
}
