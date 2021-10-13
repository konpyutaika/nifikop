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

package certmanagerpki

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	pkicommon "github.com/Orange-OpenSource/nifikop/pkg/util/pki"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// GetControllerTLSConfig creates a TLS config from the user secret created for
// cruise control and manager operations
func (c *certManager) GetControllerTLSConfig() (config *tls.Config, err error) {
	config, err = GetControllerTLSConfigFromSecret(c.client, v1alpha1.SecretReference{
		Namespace: c.cluster.Namespace,
		Name:      fmt.Sprintf(pkicommon.NodeControllerTemplate, c.cluster.Name),
	})
	return
}

func GetControllerTLSConfigFromSecret(client client.Client, ref v1alpha1.SecretReference) (config *tls.Config, err error) {
	config = &tls.Config{}
	tlsKeys := &corev1.Secret{}
	err = client.Get(context.TODO(),
		types.NamespacedName{
			Namespace: ref.Namespace,
			Name:      ref.Name,
		},
		tlsKeys,
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = errorfactory.New(errorfactory.ResourceNotReady{}, err, "controller secret not found")
		}
		return
	}
	clientCert := tlsKeys.Data[corev1.TLSCertKey]
	clientKey := tlsKeys.Data[corev1.TLSPrivateKeyKey]
	caCert := tlsKeys.Data[v1alpha1.CoreCACertKey]

	if len(caCert) == 0 {
		certs := strings.SplitAfter(string(clientCert), "-----END CERTIFICATE-----")
		clientCert = []byte(certs[0])
		caCert = []byte(certs[len(certs)-1])
		if len(certs) == 3 {
			caCert = []byte(certs[len(certs)-2])
		}
	}

	x509ClientCert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		err = errorfactory.New(errorfactory.InternalError{}, err, "could not decode controller certificate")
		return
	}

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(caCert)

	config.Certificates = []tls.Certificate{x509ClientCert}
	config.RootCAs = rootCAs

	return
}
