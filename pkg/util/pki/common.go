// Copyright © 2019 Banzai Cloud
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
// limitations under the License.

package pki

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/erdrix/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/erdrix/nifikop/pkg/resources/templates"
	certutil "github.com/erdrix/nifikop/pkg/util/cert"
	"github.com/erdrix/nifikop/pkg/util/nifi"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// NodeSelfSignerTemplate is the template used for self-signer resources
	NodeSelfSignerTemplate = "%s-self-signer"
	// NodeCACertTemplate is the template used for CA certificate resources

	NodeCACertTemplate = "%s-ca-certificate"
	// NodeServerCertTemplate is the template used for node certificate resources
	NodeServerCertTemplate = "%s-%d-server-certificate"
	// NodeIssuerTemplate is the template used for node issuer resources
	NodeIssuerTemplate = "%s-issuer"
	// NodeControllerTemplate is the template used for operator certificate resources
	NodeControllerTemplate = "%s-controller"
	// NodeControllerFQDNTemplate is combined with the above and cluster namespace
	// to create a 'fake' full-name for the controller user
	NodeControllerFQDNTemplate = "%s.%s.mgt.cluster.local"
	// CAIntermediateTemplate is the template used for intermediate CA resources
	CAIntermediateTemplate = "%s-intermediate.%s.cluster.local"
	// CAFQDNTemplate is the template used for the FQDN of a CA
	CAFQDNTemplate = "%s-ca.%s.cluster.local"
)

// Manager is the main interface for objects performing PKI operations
type Manager interface {
	// ReconcilePKI ensures a PKI for a nifi cluster - should be idempotent.
	// This method should at least setup any issuer needed for user certificates
	// as well as node secrets
	ReconcilePKI(ctx context.Context, logger logr.Logger, scheme *runtime.Scheme, externalHostnames []string) error

	// FinalizePKI performs any cleanup steps necessary for a PKI backend
	FinalizePKI(ctx context.Context, logger logr.Logger) error

	// ReconcileUserCertificate ensures and returns a user certificate - should be idempotent
	ReconcileUserCertificate(ctx context.Context, user *v1alpha1.NifiUser, scheme *runtime.Scheme) (*UserCertificate, error)

	// FinalizeUserCertificate removes/revokes a user certificate
	FinalizeUserCertificate(ctx context.Context, user *v1alpha1.NifiUser) error

	// GetControllerTLSConfig retrieves a TLS configuration for a controller nifi client
	GetControllerTLSConfig() (*tls.Config, error)
}

// UserCertificate is a struct representing the key components of a user TLS certificate
// for use across operations from other packages and internally.
type UserCertificate struct {
	CA          []byte
	Certificate []byte
	Key         []byte

	// TODO : Add Vault
	// Serial is used by vault backend for certificate revocations
	// Serial string

	// TODO : Add Vault
	// jks and password are used by vault backend for passing jks info between itself
	// the cert-manager backend passes it through the k8s secret
	//JKS      []byte
	//Password []byte
}

// DN returns the Distinguished Name of a TLS certificate
func (u *UserCertificate) DN() string {
	// cert has already been validated so we can assume no error
	cert, _ := certutil.DecodeCertificate(u.Certificate)
	return cert.Subject.String()
}

// GetInternalDNSNames returns all potential DNS names for a nifi cluster - including nodes
func GetInternalDNSNames(cluster *v1alpha1.NifiCluster, nodeId int32) (dnsNames []string) {
	dnsNames = make([]string, 0)
	dnsNames = append(dnsNames, ClusterDNSNames(cluster, nodeId)...)
	return
}

// GetCommonName returns the full FQDN for the internal NiFi listener
func GetCommonName(cluster *v1alpha1.NifiCluster) string {
	if cluster.Spec.HeadlessServiceEnabled {
		return fmt.Sprintf("%s.%s.svc.cluster.local", fmt.Sprintf(nifi.HeadlessServiceTemplate, cluster.Name), cluster.Namespace)
	}
	return fmt.Sprintf("%s.%s.svc.cluster.local", fmt.Sprintf(nifi.AllNodeServiceTemplate, cluster.Name), cluster.Namespace)
}

func GetNodeUserName(cluster *v1alpha1.NifiCluster, nodeId int32) string{
	if cluster.Spec.HeadlessServiceEnabled {
		return fmt.Sprintf("%s.%s.%s", fmt.Sprintf(templates.NodeNameTemplate, cluster.Name, nodeId), fmt.Sprintf(nifi.HeadlessServiceTemplate, cluster.Name), cluster.Namespace)
	}
	return fmt.Sprintf("%s.%s.%s", fmt.Sprintf(templates.NodeNameTemplate, cluster.Name, nodeId), fmt.Sprintf(nifi.AllNodeServiceTemplate, cluster.Name), cluster.Namespace)
}

// clusterDNSNames returns all the possible DNS Names for a NiFi Cluster
func ClusterDNSNames(cluster *v1alpha1.NifiCluster, nodeId int32) (names []string) {
	names = make([]string, 0)
	if cluster.Spec.HeadlessServiceEnabled {
		// FQDN
		names = append(names, GetCommonName(cluster))
		names = append(names, fmt.Sprintf("%s.%s", fmt.Sprintf(templates.NodeNameTemplate, cluster.Name, nodeId), GetCommonName(cluster)))

		// SVC notation
		names = append(names,
			fmt.Sprintf("%s.%s.svc", fmt.Sprintf(nifi.HeadlessServiceTemplate, cluster.Name), cluster.Namespace),
			fmt.Sprintf("%s.%s.%s.svc", fmt.Sprintf(templates.NodeNameTemplate, cluster.Name, nodeId), fmt.Sprintf(nifi.HeadlessServiceTemplate, cluster.Name), cluster.Namespace),
		)

		// Namespace notation
		names = append(names,
			fmt.Sprintf("%s.%s", fmt.Sprintf(nifi.HeadlessServiceTemplate, cluster.Name), cluster.Namespace),
			fmt.Sprintf("%s.%s.%s", fmt.Sprintf(templates.NodeNameTemplate, cluster.Name, nodeId), fmt.Sprintf(nifi.HeadlessServiceTemplate, cluster.Name), cluster.Namespace),
		)

		// service name only
		names = append(names,
			fmt.Sprintf("%s", fmt.Sprintf(nifi.HeadlessServiceTemplate, cluster.Name)),
		fmt.Sprintf("%s.%s", fmt.Sprintf(templates.NodeNameTemplate, cluster.Name, nodeId), fmt.Sprintf(nifi.HeadlessServiceTemplate, cluster.Name)))

	} else {
		// FQDN
		names = append(names, GetCommonName(cluster))
		names = append(names, fmt.Sprintf("%s.%s", fmt.Sprintf(templates.NodeNameTemplate, cluster.Name, nodeId), GetCommonName(cluster)))

		// SVC notation
		names = append(names,
			fmt.Sprintf("%s.%s.svc", fmt.Sprintf(nifi.AllNodeServiceTemplate, cluster.Name), cluster.Namespace),
			fmt.Sprintf("%s.%s.%s.svc", fmt.Sprintf(templates.NodeNameTemplate, cluster.Name, nodeId), fmt.Sprintf(nifi.AllNodeServiceTemplate, cluster.Name), cluster.Namespace),
		)

		// Namespace notation
		names = append(names,
			fmt.Sprintf("%s.%s", fmt.Sprintf(nifi.AllNodeServiceTemplate, cluster.Name), cluster.Namespace),
			fmt.Sprintf("%s.%s.%s", fmt.Sprintf(templates.NodeNameTemplate, cluster.Name, nodeId),fmt.Sprintf(nifi.AllNodeServiceTemplate, cluster.Name), cluster.Namespace),
		)

		// service name only
		names = append(names,
			fmt.Sprintf("%s", fmt.Sprintf(nifi.AllNodeServiceTemplate, cluster.Name)),
			fmt.Sprintf("%s.%s", fmt.Sprintf(templates.NodeNameTemplate, cluster.Name, nodeId),fmt.Sprintf(nifi.AllNodeServiceTemplate, cluster.Name)))
	}
	// pod name only
	names = append(names,
		fmt.Sprintf(fmt.Sprintf(templates.NodeNameTemplate, cluster.Name, nodeId)))
	return
}

// LabelsForNifiPKI returns kubernetes labels for a PKI object
func LabelsForNifiPKI(name string) map[string]string {
	return map[string]string{"app": "nifi", "nifi_issuer": fmt.Sprintf(NodeIssuerTemplate, name)}
}

// NodeUsersForCluster returns a NifiUser CR for the node certificates in a NifiCluster
func NodeUsersForCluster(cluster *v1alpha1.NifiCluster, additionalHostnames []string) []*v1alpha1.NifiUser {
	additionalHostnames = append(additionalHostnames)

	var nodeUsers []*v1alpha1.NifiUser

	for _, node := range cluster.Spec.Nodes {
		nodeUsers = append(nodeUsers, nodeUserForClusterNode(cluster, node.Id, additionalHostnames))
	}

	return nodeUsers
}

// NodeUserForClusterNode returns a NifiUser CR for the node certificates in a NifiCluster
func nodeUserForClusterNode(cluster *v1alpha1.NifiCluster, nodeId int32, additionalHostnames []string) *v1alpha1.NifiUser {
	additionalHostnames = append(additionalHostnames)
	return &v1alpha1.NifiUser{
		ObjectMeta: templates.ObjectMeta(GetNodeUserName(cluster, nodeId), LabelsForNifiPKI(cluster.Name), cluster),
		Spec: v1alpha1.NifiUserSpec{
			SecretName: fmt.Sprintf(NodeServerCertTemplate, cluster.Name, nodeId),
			DNSNames:   append(GetInternalDNSNames(cluster, nodeId), additionalHostnames...),
			IncludeJKS: true,
			ClusterRef: v1alpha1.ClusterReference{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		},
	}
}


// ControllerUserForCluster returns a NifiUser CR for the controller/cc certificates in a NifiCluster
func ControllerUserForCluster(cluster *v1alpha1.NifiCluster) *v1alpha1.NifiUser {
	return &v1alpha1.NifiUser{
		ObjectMeta: templates.ObjectMeta(
			fmt.Sprintf(NodeControllerFQDNTemplate, fmt.Sprintf(NodeControllerTemplate, cluster.Name), cluster.Namespace),
			LabelsForNifiPKI(cluster.Name), cluster,
		),
		Spec: v1alpha1.NifiUserSpec{
			SecretName: fmt.Sprintf(NodeControllerTemplate, cluster.Name),
			IncludeJKS: true,
			ClusterRef: v1alpha1.ClusterReference{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		},
	}
}
