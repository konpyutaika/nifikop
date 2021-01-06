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
	"crypto/tls"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/pki/certmanagerpki"
	"github.com/Orange-OpenSource/nifikop/pkg/util/pki"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MockBackend is used for mocking during testing
var MockBackend = v1alpha1.PKIBackend("mock")

// GetPKIManager returns a PKI/User manager interface for a given cluster
func GetPKIManager(client client.Client, cluster *v1alpha1.NifiCluster) pki.Manager {
	switch cluster.Spec.ListenersConfig.SSLSecrets.PKIBackend {

	// Use cert-manager for pki backend
	case v1alpha1.PKIBackendCertManager:
		return certmanagerpki.New(client, cluster)

	// TODO : Add vault
	// Use vault for pki backend
	/*case v1alpha1.PKIBackendVault:
	return vaultpki.New(client, cluster)*/

	// Return mock backend for testing - cannot be triggered by CR due to enum in api schema
	case MockBackend:
		return newMockPKIManager(client, cluster)

	// Default use cert-manager - state explicitly for clarity and to make compiler happy
	default:
		return certmanagerpki.New(client, cluster)

	}
}

// Mock types and functions

type mockPKIManager struct {
	pki.Manager
	client  client.Client
	cluster *v1alpha1.NifiCluster
}

func newMockPKIManager(client client.Client, cluster *v1alpha1.NifiCluster) pki.Manager {
	return &mockPKIManager{client: client, cluster: cluster}
}

func (m *mockPKIManager) ReconcilePKI(ctx context.Context, logger logr.Logger, scheme *runtime.Scheme, externalHostnames []string) error {
	return nil
}

func (m *mockPKIManager) FinalizePKI(ctx context.Context, logger logr.Logger) error {
	return nil
}

func (m *mockPKIManager) ReconcileUserCertificate(ctx context.Context, user *v1alpha1.NifiUser, scheme *runtime.Scheme) (*pki.UserCertificate, error) {
	return &pki.UserCertificate{}, nil
}

func (m *mockPKIManager) FinalizeUserCertificate(ctx context.Context, user *v1alpha1.NifiUser) error {
	return nil
}

func (m *mockPKIManager) GetControllerTLSConfig() (*tls.Config, error) {
	return &tls.Config{}, nil
}
