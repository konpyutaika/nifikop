package pki

import (
	"context"
	"crypto/tls"
	"fmt"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	certutil "github.com/konpyutaika/nifikop/pkg/util/cert"
	"github.com/konpyutaika/nifikop/pkg/util/nifi"
)

const (
	// NodeSelfSignerTemplate is the template used for self-signer resources.
	NodeSelfSignerTemplate = "%s-self-signer"
	// NodeCACertTemplate is the template used for CA certificate resources.

	NodeCACertTemplate = "%s-ca-certificate"
	// NodeServerCertTemplate is the template used for node certificate resources.
	NodeServerCertTemplate = "%s-%d-server-certificate"
	// NodeIssuerTemplate is the template used for node issuer resources.
	NodeIssuerTemplate = "%s-issuer"
	// NodeControllerFQDNTemplate is combined with the above and cluster namespace
	// to create a 'fake' full-name for the controller user.
	NodeControllerFQDNTemplate = "%s.%s.mgt.%s"
	//
	SpiffeIdTemplate = "spiffe://%s/ns/%s/nifiuser/%s"
	// CAIntermediateTemplate is the template used for intermediate CA resources.
	CAIntermediateTemplate = "%s-intermediate.%s"
	// CAFQDNTemplate is the template used for the FQDN of a CA.
	CAFQDNTemplate = "%s-ca.%s.%s"
)

// Manager is the main interface for objects performing PKI operations.
type Manager interface {
	// ReconcilePKI ensures a PKI for a nifi cluster - should be idempotent.
	// This method should at least setup any issuer needed for user certificates
	// as well as node secrets
	ReconcilePKI(ctx context.Context, logger zap.Logger, scheme *runtime.Scheme, externalHostnames []string) error

	// FinalizePKI performs any cleanup steps necessary for a PKI backend
	FinalizePKI(ctx context.Context, logger zap.Logger) error

	// ReconcileUserCertificate ensures and returns a user certificate - should be idempotent
	ReconcileUserCertificate(ctx context.Context, user *v1.NifiUser, scheme *runtime.Scheme) (*UserCertificate, error)

	// FinalizeUserCertificate removes/revokes a user certificate
	FinalizeUserCertificate(ctx context.Context, user *v1.NifiUser) error

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
	// JKS      []byte
	// Password []byte
}

// DN returns the Distinguished Name of a TLS certificate.
func (u *UserCertificate) DN() string {
	// cert has already been validated so we can assume no error
	cert, _ := certutil.DecodeCertificate(u.Certificate)
	return cert.Subject.String()
}

// GetInternalDNSNames returns all potential DNS names for a nifi cluster - including nodes.
func GetInternalDNSNames(cluster *v1.NifiCluster, nodeId int32) (dnsNames []string) {
	dnsNames = make([]string, 0)
	dnsNames = append(dnsNames, ClusterDNSNames(cluster, nodeId)...)
	return
}

// func GetCommonName(cluster *v1.NifiCluster) string {
//	return nifi.ComputeNiFiHostname(cluster.Name, cluster.Namespace, cluster.Spec.Service.HeadlessEnabled,
//		cluster.Spec.ListenersConfig.GetClusterDomain(), cluster.Spec.ListenersConfig.UseExternalDNS)
//}

func GetNodeUserName(cluster *v1.NifiCluster, nodeId int32) string {
	if cluster.Spec.NodeUserIdentityTemplate != nil {
		return fmt.Sprintf(*cluster.Spec.NodeUserIdentityTemplate, nodeId)
	}
	return nifi.ComputeRequestNiFiNodeHostname(nodeId, cluster.Name, cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled, cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS, cluster.Spec.Service.GetServiceTemplate())
}

// ClusterDNSNames returns all the possible DNS Names for a NiFi Cluster.
func ClusterDNSNames(cluster *v1.NifiCluster, nodeId int32) (names []string) {
	names = make([]string, 0)

	// FQDN
	names = append(names,
		nifi.ComputeRequestNiFiAllNodeHostname(cluster.Name, cluster.Namespace,
			cluster.Spec.ListenersConfig.GetClusterDomain(), cluster.Spec.ListenersConfig.UseExternalDNS,
			cluster.Spec.Service.GetServiceTemplate()))

	names = append(names,
		nifi.ComputeRequestNiFiNodeHostname(nodeId, cluster.Name, cluster.Namespace,
			cluster.Spec.Service.HeadlessEnabled, cluster.Spec.ListenersConfig.GetClusterDomain(),
			cluster.Spec.ListenersConfig.UseExternalDNS, cluster.Spec.Service.GetServiceTemplate()))

	if !cluster.Spec.ListenersConfig.UseExternalDNS {
		// SVC notation
		names = append(names,
			nifi.ComputeRequestNiFiAllNodeNamespaceFull(cluster.Name, cluster.Namespace,
				cluster.Spec.ListenersConfig.UseExternalDNS, cluster.Spec.Service.GetServiceTemplate()))
		names = append(names,
			nifi.ComputeRequestNiFiNodeNamespaceFull(nodeId, cluster.Name, cluster.Namespace,
				cluster.Spec.Service.HeadlessEnabled, cluster.Spec.ListenersConfig.UseExternalDNS,
				cluster.Spec.Service.GetServiceTemplate()))

		// Namespace notation
		names = append(names,
			nifi.ComputeRequestNiFiAllNodeNamespace(cluster.Name, cluster.Namespace,
				cluster.Spec.ListenersConfig.UseExternalDNS, cluster.Spec.Service.GetServiceTemplate()))
		names = append(names,
			nifi.ComputeRequestNiFiNodeNamespace(nodeId, cluster.Name, cluster.Namespace,
				cluster.Spec.Service.HeadlessEnabled, cluster.Spec.ListenersConfig.UseExternalDNS,
				cluster.Spec.Service.GetServiceTemplate()))

		// Service name only
		names = append(names,
			nifi.ComputeRequestNiFiAllNodeService(cluster.Name, cluster.Spec.Service.GetServiceTemplate()))
		names = append(names,
			nifi.ComputeRequestNiFiNodeService(nodeId, cluster.Name, cluster.Spec.Service.HeadlessEnabled,
				cluster.Spec.Service.GetServiceTemplate()))

		// Pod name only
		if cluster.Spec.Service.HeadlessEnabled {
			names = append(names,
				nifi.ComputeNodeName(nodeId, cluster.Name))
		} else {
			names = append(names, nifi.ComputeHostListenerNodeHostname(
				nodeId, cluster.Name, cluster.Namespace, cluster.Spec.ListenersConfig.GetClusterDomain(),
				cluster.Spec.ListenersConfig.UseExternalDNS, cluster.Spec.Service.GetServiceTemplate()))
		}
	}
	return
}

// LabelsForNifiPKI returns kubernetes labels for a PKI object.
func LabelsForNifiPKI(name string) map[string]string {
	return map[string]string{"app": "nifi", "nifi_issuer": fmt.Sprintf(NodeIssuerTemplate, name)}
}

// NodeUsersForCluster returns a NifiUser CR for the node certificates in a NifiCluster.
func NodeUsersForCluster(cluster *v1.NifiCluster, additionalHostnames []string) []*v1.NifiUser {
	var nodeUsers []*v1.NifiUser

	for _, node := range cluster.Spec.Nodes {
		nodeUsers = append(nodeUsers, nodeUserForClusterNode(cluster, node.Id, additionalHostnames))
	}

	return nodeUsers
}

// NodeUserForClusterNode returns a NifiUser CR for the node certificates in a NifiCluster.
func nodeUserForClusterNode(cluster *v1.NifiCluster, nodeId int32, additionalHostnames []string) *v1.NifiUser {
	return &v1.NifiUser{
		ObjectMeta: templates.ObjectMeta(GetNodeUserName(cluster, nodeId), LabelsForNifiPKI(cluster.Name), cluster),
		Spec: v1.NifiUserSpec{
			SecretName: fmt.Sprintf(NodeServerCertTemplate, cluster.Name, nodeId),
			DNSNames:   append(GetInternalDNSNames(cluster, nodeId), additionalHostnames...),
			IncludeJKS: true,
			ClusterRef: v1.ClusterReference{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
			AccessPolicies: []v1.AccessPolicy{
				{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ProxyAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.ProxyAccessPolicyResource},
			},
		},
	}
}

// ControllerUserForCluster returns a NifiUser CR for the controller/cc certificates in a NifiCluster.
func ControllerUserForCluster(cluster *v1.NifiCluster) *v1.NifiUser {
	/*nodeControllerName := fmt.Sprintf(NodeControllerFQDNTemplate,
	cluster.GetNifiControllerUserIdentity(),
	cluster.Namespace,
	cluster.Spec.ListenersConfig.GetClusterDomain())*/

	return &v1.NifiUser{
		ObjectMeta: templates.ObjectMeta(
			cluster.GetNifiControllerUserIdentity(),
			LabelsForNifiPKI(cluster.Name), cluster,
		),
		Spec: v1.NifiUserSpec{
			DNSNames:   []string{cluster.GetNifiControllerUserIdentity()},
			SecretName: cluster.GetNifiControllerUserIdentity(),
			IncludeJKS: true,
			ClusterRef: v1.ClusterReference{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
			AccessPolicies: []v1.AccessPolicy{
				{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.FlowAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.FlowAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ControllerAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.ControllerAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.RestrictedComponentsAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.RestrictedComponentsAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.PoliciesAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.PoliciesAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.TenantsAccessPolicyResource},
				{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.TenantsAccessPolicyResource},
			},
		},
	}
}
