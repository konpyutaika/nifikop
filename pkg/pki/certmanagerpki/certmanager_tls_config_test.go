package certmanagerpki

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	certutil "github.com/konpyutaika/nifikop/pkg/util/cert"
)

func newMockControllerSecret(valid bool) *corev1.Secret {
	secret := &corev1.Secret{}
	secret.Name = "test-controller"
	secret.Namespace = "test-namespace"
	cert, key, _, _ := certutil.GenerateTestCert()
	if valid {
		secret.Data = map[string][]byte{
			corev1.TLSCertKey:       cert,
			corev1.TLSPrivateKeyKey: key,
			v1.CoreCACertKey:        cert,
		}
	}
	return secret
}

func TestGetControllerTLSConfig(t *testing.T) {
	manager := newMock(newMockCluster())

	// Test good controller secret
	manager.client.Create(context.TODO(), newMockControllerSecret(true))
	if _, err := manager.GetControllerTLSConfig(); err != nil {
		t.Error("Expected no error, got:", err)
	}

	manager = newMock(newMockCluster())

	// Test non-existent controller secret
	if _, err := manager.GetControllerTLSConfig(); err == nil {
		t.Error("Expected error got nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errorfactory.ResourceNotReady{}) {
		t.Error("Expected not ready error, got:", reflect.TypeOf(err))
	}

	// Test invalid controller secret
	manager.client.Create(context.TODO(), newMockControllerSecret(false))
	if _, err := manager.GetControllerTLSConfig(); err == nil {
		t.Error("Expected error got nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errorfactory.InternalError{}) {
		t.Error("Expected internal error, got:", reflect.TypeOf(err))
	}
}
