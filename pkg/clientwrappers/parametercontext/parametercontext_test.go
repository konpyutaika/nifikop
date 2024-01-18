package parametercontext

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExtractMostRecentSecretUpdateTime(t *testing.T) {
	oldestTime := metav1.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	mostRecentTime := metav1.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)

	secret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			ManagedFields: []metav1.ManagedFieldsEntry{
				{
					Time: &oldestTime,
				},
				{
					Time: &mostRecentTime,
				},
			},
		},
	}
	secret2 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			ManagedFields: []metav1.ManagedFieldsEntry{
				{
					Time: &mostRecentTime,
				},
				{
					Time: &oldestTime,
				},
			},
		},
	}
	secret3 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			ManagedFields: []metav1.ManagedFieldsEntry{
				{
					Time: &oldestTime,
				},
				{
					Time: &oldestTime,
				},
			},
		},
	}

	if result := extractMostRecentSecretUpdateTime([]*corev1.Secret{secret1, secret2, secret3}); result == nil || !result.Equal(&mostRecentTime) {
		t.Errorf("Expected %v, but got %v", &mostRecentTime, result)
	}

	if result := extractMostRecentSecretUpdateTime([]*corev1.Secret{secret1, secret2}); result == nil || !result.Equal(&mostRecentTime) {
		t.Errorf("Expected %v, but got %v", &mostRecentTime, result)
	}

	if result := extractMostRecentSecretUpdateTime([]*corev1.Secret{secret1, secret3}); result == nil || !result.Equal(&mostRecentTime) {
		t.Errorf("Expected %v, but got %v", &mostRecentTime, result)
	}

	if result := extractMostRecentSecretUpdateTime([]*corev1.Secret{secret2, secret3}); result == nil || !result.Equal(&mostRecentTime) {
		t.Errorf("Expected %v, but got %v", &mostRecentTime, result)
	}

	if result := extractMostRecentSecretUpdateTime([]*corev1.Secret{secret1}); result == nil || !result.Equal(&mostRecentTime) {
		t.Errorf("Expected %v, but got %v", &mostRecentTime, result)
	}

	if result := extractMostRecentSecretUpdateTime([]*corev1.Secret{secret2}); result == nil || !result.Equal(&mostRecentTime) {
		t.Errorf("Expected %v, but got %v", &mostRecentTime, result)
	}

	if result := extractMostRecentSecretUpdateTime([]*corev1.Secret{secret3}); result == nil || !result.Equal(&oldestTime) {
		t.Errorf("Expected %v, but got %v", &oldestTime, result)
	}
}
