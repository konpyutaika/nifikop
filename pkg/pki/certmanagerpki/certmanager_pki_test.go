package certmanagerpki

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"

	common2 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	certutil "github.com/konpyutaika/nifikop/pkg/util/cert"
	pkicommon "github.com/konpyutaika/nifikop/pkg/util/pki"
)

var log = common.CustomLogger().Named("testing")

func newNodeServerSecret(nodeId int32) *corev1.Secret {
	secret := &corev1.Secret{}
	secret.Name = fmt.Sprintf(pkicommon.NodeServerCertTemplate, "test", nodeId)
	secret.Namespace = "test-namespace"
	cert, key, _, _ := certutil.GenerateTestCert()
	secret.Data = map[string][]byte{
		corev1.TLSCertKey:       cert,
		corev1.TLSPrivateKeyKey: key,
		common2.CoreCACertKey:   cert,
	}
	return secret
}

func newControllerSecret() *corev1.Secret {
	secret := &corev1.Secret{}
	secret.Name = fmt.Sprintf("%s-controller", "test")
	secret.Namespace = "test-namespace"
	cert, key, _, _ := certutil.GenerateTestCert()
	secret.Data = map[string][]byte{
		corev1.TLSCertKey:       cert,
		corev1.TLSPrivateKeyKey: key,
		common2.CoreCACertKey:   cert,
	}
	return secret
}

func newCASecret() *corev1.Secret {
	secret := &corev1.Secret{}
	secret.Name = fmt.Sprintf(pkicommon.NodeCACertTemplate, "test")
	secret.Namespace = "cert-manager"
	cert, key, _, _ := certutil.GenerateTestCert()
	secret.Data = map[string][]byte{
		corev1.TLSCertKey:       cert,
		corev1.TLSPrivateKeyKey: key,
		common2.CoreCACertKey:   cert,
	}
	return secret
}

func newPreCreatedSecret() *corev1.Secret {
	secret := &corev1.Secret{}
	secret.Name = "test-controller"
	secret.Namespace = "test-namespace"
	cert, key, _, _ := certutil.GenerateTestCert()
	secret.Data = map[string][]byte{
		common2.CAPrivateKeyKey: key,
		common2.CACertKey:       cert,
	}
	return secret
}

func TestFinalizePKI(t *testing.T) {
	manager := newMock(newMockCluster())

	if err := manager.FinalizePKI(context.Background(), *log); err != nil {
		t.Error("Expected no error on finalize, got:", err)
	}
}

func TestReconcilePKI(t *testing.T) {
	cluster := newMockCluster()
	manager := newMock(cluster)
	ctx := context.Background()

	for _, node := range cluster.Spec.Nodes {
		manager.client.Create(ctx, newNodeServerSecret(node.Id))
		if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
			if reflect.TypeOf(err) != reflect.TypeOf(errorfactory.ResourceNotReady{}) {
				t.Error("Expected not ready error, got:", reflect.TypeOf(err))
			}
		}
	}

	manager.client.Create(ctx, newControllerSecret())
	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		if reflect.TypeOf(err) != reflect.TypeOf(errorfactory.ResourceNotReady{}) {
			t.Error("Expected not ready error, got:", reflect.TypeOf(err))
		}
	}

	manager.client.Create(ctx, newCASecret())
	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Error("Expected successful reconcile, got:", err)
	}

	cluster.Spec.ListenersConfig.SSLSecrets.Create = false
	manager = newMock(cluster)
	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err == nil {
		t.Error("Expected error got nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errorfactory.ResourceNotReady{}) {
		t.Error("Expected not ready error, got:", reflect.TypeOf(err))
	}
	manager.client.Create(ctx, newPreCreatedSecret())
	if err := manager.ReconcilePKI(ctx, *log, scheme.Scheme, []string{}); err != nil {
		t.Error("Expected successful reconcile, got:", err)
	}
}
