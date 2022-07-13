package util

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

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
