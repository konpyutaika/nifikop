package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

func TestSubtractNodes(t *testing.T) {
	sourceList := []v1.Node{
		{
			Id: 1,
		},
		{
			Id: 2,
		},
		{
			Id: 3,
		},
	}

	nodesToSubtract := []v1.Node{
		{
			Id: 3,
		},
	}

	// subtract 1 node
	if results := SubtractNodes(sourceList, nodesToSubtract); len(results) != 2 {
		t.Error("There should be two nodes remaining")
	}

	// subtract empty list
	if results := SubtractNodes(sourceList, []v1.Node{}); len(results) != 3 {
		t.Error("there should be 3 results")
	}

	// subtract all nodes
	if results := SubtractNodes(sourceList, sourceList); len(results) != 0 {
		t.Error("There should be two nodes remaining")
	}
}

func TestMergeHostAliasesOverride(t *testing.T) {
	globalAliases := []corev1.HostAlias{
		{
			IP: "1.2.3.4",
			Hostnames: []string{
				"global.host",
			},
		},
	}
	overrideAliases := []corev1.HostAlias{
		{
			IP: "1.2.3.4",
			Hostnames: []string{
				"override.host",
			},
		},
	}

	results := MergeHostAliases(globalAliases, overrideAliases)

	if len(results) != 1 {
		t.Errorf("The merge results are not the correct length: %v+", results)
	}

	if results[0].IP != "1.2.3.4" {
		t.Errorf("results are not correct: %v+", results[0])
	}
	if len(results[0].Hostnames) != 1 || results[0].Hostnames[0] != "override.host" {
		t.Errorf("override did not work: %v+", results[0])
	}
}

func TestMergeHostAliasesJoin(t *testing.T) {
	globalAliases := []corev1.HostAlias{
		{
			IP: "1.2.3.4",
			Hostnames: []string{
				"global.host",
			},
		},
	}
	overrideAliases := []corev1.HostAlias{
		{
			IP: "1.2.3.5",
			Hostnames: []string{
				"override.host",
			},
		},
	}

	results := MergeHostAliases(globalAliases, overrideAliases)

	if len(results) != 2 {
		t.Errorf("The merge results are not the correct length: %v+", results)
	}
}

func TestMergeHostAliasesEmpty(t *testing.T) {
	globalAliases := []corev1.HostAlias{}
	overrideAliases := []corev1.HostAlias{}

	results := MergeHostAliases(globalAliases, overrideAliases)

	if len(results) != 0 {
		t.Errorf("The merge results are not the correct length: %v+", results)
	}
}

func TestStringSliceCompare(t *testing.T) {
	listOriginal := []string{"a", "b", "c"}
	listDisordered := []string{"c", "a", "b"}
	listLess := []string{"a", "b"}
	listMore := []string{"a", "b", "c", "d"}
	listDifferent := []string{"1", "2", "3"}

	// same list
	if results := StringSliceCompare(listOriginal, listOriginal); !results {
		t.Error("The list must be considered as equal")
	}

	// same list but disordered
	if results := StringSliceCompare(listOriginal, listDisordered); !results {
		t.Error("The list must be considered as equal")
	}

	// list with less listLess
	if results := StringSliceCompare(listOriginal, listLess); results {
		t.Error("The list must be considered as different because there is less items")
	}

	// list with more items
	if results := StringSliceCompare(listOriginal, listMore); results {
		t.Error("The list must be considered as different because there is more items")
	}

	// list of same size but different
	if results := StringSliceCompare(listOriginal, listDifferent); results {
		t.Error("The list must be considered as different because there is no identical items")
	}
}

func TestStringSliceStrictCompare(t *testing.T) {
	listOriginal := []string{"a", "b", "c"}
	listDisordered := []string{"c", "a", "b"}
	listLess := []string{"a", "b"}
	listMore := []string{"a", "b", "c", "d"}
	listDifferent := []string{"1", "2", "3"}

	// same list
	if results := StringSliceStrictCompare(listOriginal, listOriginal); !results {
		t.Error("The list must be considered as equal")
	}

	// same list but disordered
	if results := StringSliceStrictCompare(listOriginal, listDisordered); results {
		t.Error("The list must be considered as different because the order is different")
	}

	// list with less listLess
	if results := StringSliceStrictCompare(listOriginal, listLess); results {
		t.Error("The list must be considered as different because there is less items")
	}

	// list with more items
	if results := StringSliceStrictCompare(listOriginal, listMore); results {
		t.Error("The list must be considered as different because there is more items")
	}

	// list of same size but different
	if results := StringSliceStrictCompare(listOriginal, listDifferent); results {
		t.Error("The list must be considered as different because there is no identical items")
	}
}

func TestStringSliceContains(t *testing.T) {
	list := []string{"a", "b", "c"}

	// item in the list
	if results := StringSliceContains(list, "a"); !results {
		t.Error("The item is in the list")
	}

	// same list but disordered
	if results := StringSliceContains(list, "1"); results {
		t.Error("The item is not in the list")
	}
}

func TestStringSliceRemove(t *testing.T) {
	list := []string{"a", "b", "c"}
	listCopy := make([]string, len(list))

	copy(listCopy, list)
	// item in the list
	if results := StringSliceRemove(listCopy, "a"); len(results) != len(list)-1 ||
		results[0] != list[1] || results[1] != list[2] {
		t.Error("The list must have an item less")
	}

	// empty the list
	newlist := []string{"a"}
	if results := StringSliceRemove(newlist, "a"); len(results) != 0 {
		t.Error("The list must be empty")
	}

	copy(listCopy, list)
	// item not in the list
	if results := StringSliceRemove(listCopy, "1"); len(results) != len(list) ||
		results[0] != list[0] || results[1] != list[1] || results[2] != list[2] {
		t.Error("The list should be the same")
	}
}

func TestExtractSecretsResourceVersion(t *testing.T) {
	secret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "secret",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
	}
	secret2 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "secret2",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
	}
	secret3 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "secret",
			Namespace:       "ns2",
			ResourceVersion: "1",
		},
	}

	secretResult := []v1.SecretResourceVersion{
		{
			Name:            "secret",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret2",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret",
			Namespace:       "ns2",
			ResourceVersion: "1",
		},
	}
	if result := ExtractSecretsResourceVersion([]*corev1.Secret{secret1, secret2, secret3}); len(result) != len(secretResult) || !reflect.DeepEqual(result, secretResult) {
		t.Errorf("Expected %v, but got %v", secretResult, result)
	}
}

func TestIsSecretResourceVersionUpdated(t *testing.T) {
	secret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "secret",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
	}
	secret2 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "secret2",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
	}
	secret3 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "secret",
			Namespace:       "ns2",
			ResourceVersion: "1",
		},
	}

	if IsSecretResourceVersionUpdated([]*corev1.Secret{secret1, secret2, secret3}, []v1.SecretResourceVersion{
		{
			Name:            "secret",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret2",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret",
			Namespace:       "ns2",
			ResourceVersion: "1",
		},
	}) {
		t.Errorf("Expected false, but got true")
	}

	if !IsSecretResourceVersionUpdated([]*corev1.Secret{secret1, secret2}, []v1.SecretResourceVersion{
		{
			Name:            "secret",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret2",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret",
			Namespace:       "ns2",
			ResourceVersion: "1",
		},
	}) {
		t.Errorf("Expected true, but got false")
	}

	if !IsSecretResourceVersionUpdated([]*corev1.Secret{secret1, secret2}, []v1.SecretResourceVersion{
		{
			Name:            "secret",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret",
			Namespace:       "ns2",
			ResourceVersion: "1",
		},
	}) {
		t.Errorf("Expected true, but got false")
	}

	if !IsSecretResourceVersionUpdated([]*corev1.Secret{secret1, secret2}, []v1.SecretResourceVersion{
		{
			Name:            "secret",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret2",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret",
			Namespace:       "ns2",
			ResourceVersion: "1",
		},
	}) {
		t.Errorf("Expected true, but got false")
	}

	if !IsSecretResourceVersionUpdated([]*corev1.Secret{secret1, secret2, secret3}, []v1.SecretResourceVersion{
		{
			Name:            "secret",
			Namespace:       "ns",
			ResourceVersion: "2",
		},
		{
			Name:            "secret2",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret",
			Namespace:       "ns2",
			ResourceVersion: "1",
		},
	}) {
		t.Errorf("Expected true, but got false")
	}

	if !IsSecretResourceVersionUpdated([]*corev1.Secret{secret1, secret2, secret3}, []v1.SecretResourceVersion{
		{
			Name:            "secret",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret2",
			Namespace:       "ns",
			ResourceVersion: "1",
		},
		{
			Name:            "secret2",
			Namespace:       "ns2",
			ResourceVersion: "1",
		},
	}) {
		t.Errorf("Expected true, but got false")
	}
}

func TestGetNodeConfigPodOverridesNodeLevelReplacesGroup(t *testing.T) {
	groupOverrides := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			PriorityClassName: "group",
			Containers:        []corev1.Container{{Name: "group-sidecar"}},
		},
	}
	nodeOverrides := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "node-sidecar"}},
		},
	}
	spec := v1.NifiClusterSpec{
		NodeConfigGroups: map[string]v1.NodeConfig{
			"default": {PodOverrides: groupOverrides, PriorityClassName: ptrTo("group-pc")},
		},
	}

	// node without its own override inherits the group's
	got, err := GetNodeConfig(v1.Node{Id: 0, NodeConfigGroup: "default"}, spec)
	assert.NoError(t, err)
	assert.Equal(t, groupOverrides, got.PodOverrides)

	// node with its own override replaces the group's wholesale (no deep merge)
	got, err = GetNodeConfig(v1.Node{
		Id:              1,
		NodeConfigGroup: "default",
		NodeConfig:      &v1.NodeConfig{PodOverrides: nodeOverrides},
	}, spec)
	assert.NoError(t, err)
	assert.Equal(t, nodeOverrides, got.PodOverrides)
	assert.Equal(t, "", got.PodOverrides.Spec.PriorityClassName)
	assert.Len(t, got.PodOverrides.Spec.Containers, 1)
	// other fields still merge as before
	assert.Equal(t, "group-pc", *got.PriorityClassName)

	// the group's template is copied, never aliased into the returned config
	got, err = GetNodeConfig(v1.Node{Id: 2, NodeConfigGroup: "default"}, spec)
	assert.NoError(t, err)
	assert.NotSame(t, groupOverrides, got.PodOverrides)
	assert.Equal(t, groupOverrides, got.PodOverrides)
}

func ptrTo[T any](v T) *T { return &v }
