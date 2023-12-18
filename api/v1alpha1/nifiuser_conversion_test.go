package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

func TestNifiUserConversion(t *testing.T) {
	nu := createNifiUser()
	// convert v1alhpa1 to v1
	v1nu := &v1.NifiUser{}
	nu.ConvertTo(v1nu)
	assertNifiUsersEqual(nu, v1nu, t)

	// convert v1 to v1alpha1
	newUser := &NifiUser{}
	newUser.ConvertFrom(v1nu)
	assertNifiUsersEqual(newUser, v1nu, t)
}

func assertNifiUsersEqual(u1 *NifiUser, u2 *v1.NifiUser, t *testing.T) {
	if u1.ObjectMeta.Name != u2.ObjectMeta.Name ||
		u1.ObjectMeta.Namespace != u2.ObjectMeta.Namespace {
		t.Error("object metas not equal")
	}
	if u1.Spec.CreateCert != u2.Spec.CreateCert {
		t.Error("create certs not equal")
	}
	if u1.Spec.IncludeJKS != u2.Spec.IncludeJKS {
		t.Error("include JKSs not equal")
	}
	if u1.Spec.CreateCert != u2.Spec.CreateCert {
		t.Error("create certs not equal")
	}
	if u1.Spec.SecretName != u2.Spec.SecretName {
		t.Error("secret names not equal")
	}
	if !accessPoliciesEqual(u1.Spec.AccessPolicies, u2.Spec.AccessPolicies) {
		t.Error("access policies not equal")
	}

	if u1.Status.Id != u2.Status.Id {
		t.Error("status ids not equal")
	}
	if u1.Status.Version != u2.Status.Version {
		t.Error("status versions not equal")
	}
}

func accessPoliciesEqual(ap1 []AccessPolicy, ap2 []v1.AccessPolicy) bool {
	if len(ap1) != len(ap2) {
		return false
	}

	for i, ap := range ap1 {
		if string(ap.Action) != string(ap2[i].Action) ||
			string(ap.Resource) != string(ap2[i].Resource) ||
			string(ap.Type) != string(ap2[i].Type) ||
			ap.ComponentId != ap2[i].ComponentId ||
			ap.ComponentType != ap2[i].ComponentType {
			return false
		}
	}
	return true
}

func createNifiUser() *NifiUser {
	var createCert bool = true
	return &NifiUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "user",
			Namespace: "namespace",
		},
		Spec: NifiUserSpec{
			Identity:   "identity",
			SecretName: "secret",
			ClusterRef: ClusterReference{
				Name:      "cluster",
				Namespace: "namespace",
			},
			DNSNames:   []string{"dns1", "dns2"},
			IncludeJKS: true,
			CreateCert: &createCert,
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
		Status: NifiUserStatus{
			Id:      "id",
			Version: 6,
		},
	}
}
