// Copyright 2020 Orange SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package apis

package pki

import (
	"context"
	"reflect"
	"testing"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log logr.Logger

type mockClient struct {
	client.Client
}

func newMockCluster() *v1alpha1.NifiCluster {
	cluster := &v1alpha1.NifiCluster{}
	cluster.Name = "test"
	cluster.Namespace = "test"
	cluster.Spec = v1alpha1.NifiClusterSpec{}
	cluster.Spec.ListenersConfig = &v1alpha1.ListenersConfig{}
	cluster.Spec.ListenersConfig.InternalListeners = []v1alpha1.InternalListenerConfig{
		{ContainerPort: 80},
	}
	cluster.Spec.ListenersConfig.SSLSecrets = &v1alpha1.SSLSecrets{
		PKIBackend: MockBackend,
	}
	return cluster
}

func TestGetPKIManager(t *testing.T) {
	cluster := newMockCluster()
	mock := GetPKIManager(&mockClient{}, cluster)
	if reflect.TypeOf(mock) != reflect.TypeOf(&mockPKIManager{}) {
		t.Error("Expected mock client got:", reflect.TypeOf(mock))
	}
	ctx := context.Background()

	// Test mock functions
	var err error
	if err = mock.ReconcilePKI(ctx, log, scheme.Scheme, []string{}); err != nil {
		t.Error("Expected nil error got:", err)
	}

	if err = mock.FinalizePKI(ctx, log); err != nil {
		t.Error("Expected nil error got:", err)
	}

	if _, err = mock.ReconcileUserCertificate(ctx, &v1alpha1.NifiUser{}, scheme.Scheme); err != nil {
		t.Error("Expected nil error got:", err)
	}

	if err = mock.FinalizeUserCertificate(ctx, &v1alpha1.NifiUser{}); err != nil {
		t.Error("Expected nil error got:", err)
	}

	if _, err = mock.GetControllerTLSConfig(); err != nil {
		t.Error("Expected nil error got:", err)
	}

	// Test other getters
	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1alpha1.PKIBackendCertManager
	certmanager := GetPKIManager(&mockClient{}, cluster)
	pkiType := reflect.TypeOf(certmanager).String()
	expected := "*certmanagerpki.certManager"
	if pkiType != expected {
		t.Error("Expected:", expected, "got:", pkiType)
	}

	// Default should be cert-manager also
	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1alpha1.PKIBackend("")
	certmanager = GetPKIManager(&mockClient{}, cluster)
	pkiType = reflect.TypeOf(certmanager).String()
	expected = "*certmanagerpki.certManager"
	if pkiType != expected {
		t.Error("Expected:", expected, "got:", pkiType)
	}

	/* TODO : Add Vault
	cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend = v1alpha1.PKIBackendVault
	certmanager = GetPKIManager(&mockClient{}, cluster)
	pkiType = reflect.TypeOf(certmanager).String()
	expected = "*vaultpki.vaultPKI"
	if pkiType != expected {
		t.Error("Expected:", expected, "got:", pkiType)
	}*/

}
