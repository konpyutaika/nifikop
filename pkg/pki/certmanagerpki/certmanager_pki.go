package certmanagerpki

import (
	"context"
	"fmt"

	certv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	pkicommon "github.com/konpyutaika/nifikop/pkg/util/pki"
)

func (c *certManager) FinalizePKI(ctx context.Context, logger zap.Logger) error {
	logger.Info("Removing cert-manager certificates and secrets",
		zap.String("clusterName", c.cluster.Name))

	// Safety check that we are actually doing something
	if c.cluster.Spec.ListenersConfig.SSLSecrets == nil {
		return nil
	}

	if c.cluster.Spec.ListenersConfig.SSLSecrets.Create {
		// Names of our certificates and secrets
		objNames := []types.NamespacedName{
			{Name: c.cluster.GetNifiControllerUserIdentity(), Namespace: c.cluster.Namespace},
		}

		for _, node := range c.cluster.Spec.Nodes {
			objNames = append(objNames, types.NamespacedName{Name: fmt.Sprintf(pkicommon.NodeServerCertTemplate, c.cluster.Name, node.Id), Namespace: c.cluster.Namespace})
		}

		if c.cluster.Spec.ListenersConfig.SSLSecrets.IssuerRef == nil {
			objNames = append(
				objNames,
				types.NamespacedName{Name: fmt.Sprintf(pkicommon.NodeCACertTemplate, c.cluster.Name), Namespace: c.cluster.Namespace})
		}
		for _, obj := range objNames {
			// Delete the certificates first so we don't accidentally recreate the
			// secret after it gets deleted
			cert := &certv1.Certificate{}
			if err := c.client.Get(ctx, obj, cert); err != nil {
				if apierrors.IsNotFound(err) {
					continue
				} else {
					return err
				}
			}
			if err := c.client.Delete(ctx, cert); err != nil {
				return err
			}

			// Might as well delete the secret and leave the controller reference earlier
			// as a safety belt
			secret := &corev1.Secret{}
			if err := c.client.Get(ctx, obj, secret); err != nil {
				if apierrors.IsNotFound(err) {
					continue
				} else {
					return err
				}
			}
			if err := c.client.Delete(ctx, secret); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *certManager) ReconcilePKI(ctx context.Context, logger zap.Logger, scheme *runtime.Scheme, externalHostnames []string) (err error) {
	logger.Info("Reconciling cert-manager PKI",
		zap.String("clusterName", c.cluster.Name))

	resources, err := c.
		nifipki(ctx, scheme, externalHostnames)
	if err != nil {
		return err
	}

	for _, o := range resources {
		if err := reconcile(ctx, logger, c.client, o, c.cluster); err != nil {
			return err
		}
	}

	return nil
}

func (c *certManager) nifipki(ctx context.Context, scheme *runtime.Scheme, externalHostnames []string) ([]runtime.Object, error) {
	sslConfig := c.cluster.Spec.ListenersConfig.SSLSecrets
	if sslConfig.Create {
		if sslConfig.IssuerRef == nil {
			return fullPKI(c.cluster, scheme, externalHostnames), nil
		}
		return userProvidedIssuerPKI(c.cluster, externalHostnames), nil
	}
	return userProvidedPKI(ctx, c.client, c.cluster, scheme, externalHostnames)
}

func userProvidedIssuerPKI(cluster *v1.NifiCluster, externalHostnames []string) []runtime.Object {
	// No need to generate self-signed certs and issuers because the issuer is provided by user
	objects := []runtime.Object{
		// Operator user
		pkicommon.ControllerUserForCluster(cluster),
	}
	// Node "users"
	for _, user := range pkicommon.NodeUsersForCluster(cluster, externalHostnames) {
		objects = append(objects, user)
	}

	return objects
}

func fullPKI(cluster *v1.NifiCluster, scheme *runtime.Scheme, externalHostnames []string) []runtime.Object {
	var objects []runtime.Object

	if cluster.Spec.ListenersConfig.SSLSecrets.ClusterScoped {
		objects = append(objects, []runtime.Object{
			// A self-signer for the CA Certificate
			selfSignerForCluster(cluster, scheme),
			// The CA Certificate
			caCertForCluster(cluster, scheme),
			// A cluster issuer backed by the CA certificate - so it can provision secrets
			// for app in other namespaces
			mainIssuerForCluster(cluster, scheme),
		}...,
		)
	} else {
		objects = append(objects, []runtime.Object{
			// A self-signer for the CA Certificate
			selfSignerForNamespace(cluster, scheme),
			// The CA Certificate
			caCertForNamespace(cluster, scheme),
			// A issuer backed by the CA certificate - so it can provision secrets
			// in this namespace
			mainIssuerForNamespace(cluster, scheme),
		}...,
		)
	}

	objects = append(objects, pkicommon.ControllerUserForCluster(cluster))
	// Node "users"
	for _, user := range pkicommon.NodeUsersForCluster(cluster, externalHostnames) {
		objects = append(objects, user)
	}
	return objects
}

func userProvidedPKI(ctx context.Context, client client.Client, cluster *v1.NifiCluster, scheme *runtime.Scheme, externalHostnames []string) ([]runtime.Object, error) {
	// If we aren't creating the secrets we need a cluster issuer made from the provided secret
	caSecret, err := caSecretForProvidedCert(ctx, client, cluster, scheme)
	if err != nil {
		return nil, err
	}

	objects := []runtime.Object{
		caSecret,
	}
	if cluster.Spec.ListenersConfig.SSLSecrets.ClusterScoped {
		objects = append(objects, mainIssuerForCluster(cluster, scheme))
	} else {
		objects = append(objects, mainIssuerForNamespace(cluster, scheme))
	}

	objects = append(objects, pkicommon.ControllerUserForCluster(cluster))

	// Node "users"
	for _, user := range pkicommon.NodeUsersForCluster(cluster, externalHostnames) {
		objects = append(objects, user)
	}

	return objects, nil
}

func caSecretForProvidedCert(ctx context.Context, client client.Client, cluster *v1.NifiCluster, scheme *runtime.Scheme) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.ListenersConfig.SSLSecrets.TLSSecretName}, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = errorfactory.New(errorfactory.ResourceNotReady{}, err, "could not find provided tls secret")
		} else {
			err = errorfactory.New(errorfactory.APIFailure{}, err, "could not lookup provided tls secret")
		}
		return nil, err
	}

	caKey := secret.Data[v1.CAPrivateKeyKey]
	caCert := secret.Data[v1.CACertKey]

	caSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(pkicommon.NodeCACertTemplate, cluster.Name),
			Namespace: cluster.Namespace,
			Labels:    pkicommon.LabelsForNifiPKI(cluster.Name),
		},
		Data: map[string][]byte{
			v1.CoreCACertKey:        caCert,
			corev1.TLSCertKey:       caCert,
			corev1.TLSPrivateKeyKey: caKey,
		},
	}
	return caSecret, nil
}

func selfSignerForCluster(cluster *v1.NifiCluster, scheme *runtime.Scheme) *certv1.ClusterIssuer {
	selfsignerMeta := templates.ObjectMeta(fmt.Sprintf(pkicommon.NodeSelfSignerTemplate, cluster.Name), pkicommon.LabelsForNifiPKI(cluster.Name), cluster)
	selfsignerMeta.Namespace = metav1.NamespaceAll
	selfsigner := &certv1.ClusterIssuer{
		ObjectMeta: selfsignerMeta,
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				SelfSigned: &certv1.SelfSignedIssuer{},
			},
		},
	}
	controllerutil.SetControllerReference(cluster, selfsigner, scheme)
	return selfsigner
}

func caCertForCluster(cluster *v1.NifiCluster, scheme *runtime.Scheme) *certv1.Certificate {
	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(pkicommon.NodeCACertTemplate, cluster.Name),
			Namespace: cluster.Namespace,
			Labels:    pkicommon.LabelsForNifiPKI(cluster.Name),
		},
		Spec: certv1.CertificateSpec{
			SecretName: fmt.Sprintf(pkicommon.NodeCACertTemplate, cluster.Name),
			CommonName: fmt.Sprintf(pkicommon.CAFQDNTemplate,
				cluster.Name, cluster.Namespace, cluster.Spec.ListenersConfig.GetClusterDomain()),
			IsCA: true,
			IssuerRef: certmeta.ObjectReference{
				Name: fmt.Sprintf(pkicommon.NodeSelfSignerTemplate, cluster.Name),
				Kind: certv1.ClusterIssuerKind,
			},
		},
	}
}

func mainIssuerForCluster(cluster *v1.NifiCluster, scheme *runtime.Scheme) *certv1.ClusterIssuer {
	clusterIssuerMeta := templates.ObjectMeta(fmt.Sprintf(pkicommon.NodeIssuerTemplate, cluster.Name), pkicommon.LabelsForNifiPKI(cluster.Name), cluster)
	clusterIssuerMeta.Namespace = metav1.NamespaceAll
	issuer := &certv1.ClusterIssuer{
		ObjectMeta: clusterIssuerMeta,
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				CA: &certv1.CAIssuer{
					SecretName: fmt.Sprintf(pkicommon.NodeCACertTemplate, cluster.Name),
				},
			},
		},
	}
	controllerutil.SetControllerReference(cluster, issuer, scheme)
	return issuer
}

func selfSignerForNamespace(cluster *v1.NifiCluster, scheme *runtime.Scheme) *certv1.Issuer {
	selfsignerMeta := templates.ObjectMeta(fmt.Sprintf(pkicommon.NodeSelfSignerTemplate, cluster.Name), pkicommon.LabelsForNifiPKI(cluster.Name), cluster)
	selfsignerMeta.Namespace = cluster.Namespace
	selfsigner := &certv1.Issuer{
		ObjectMeta: selfsignerMeta,
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				SelfSigned: &certv1.SelfSignedIssuer{},
			},
		},
	}
	controllerutil.SetControllerReference(cluster, selfsigner, scheme)
	return selfsigner
}

func caCertForNamespace(cluster *v1.NifiCluster, scheme *runtime.Scheme) *certv1.Certificate {
	return &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(pkicommon.NodeCACertTemplate, cluster.Name),
			Namespace: cluster.Namespace,
			Labels:    pkicommon.LabelsForNifiPKI(cluster.Name),
		},
		Spec: certv1.CertificateSpec{
			SecretName: fmt.Sprintf(pkicommon.NodeCACertTemplate, cluster.Name),
			CommonName: fmt.Sprintf(pkicommon.CAFQDNTemplate,
				cluster.Name, cluster.Namespace, cluster.Spec.ListenersConfig.GetClusterDomain()),
			IsCA: true,
			IssuerRef: certmeta.ObjectReference{
				Name: fmt.Sprintf(pkicommon.NodeSelfSignerTemplate, cluster.Name),
				Kind: certv1.IssuerKind,
			},
		},
	}
}

func mainIssuerForNamespace(cluster *v1.NifiCluster, scheme *runtime.Scheme) *certv1.Issuer {
	issuerMeta := templates.ObjectMeta(fmt.Sprintf(pkicommon.NodeIssuerTemplate, cluster.Name), pkicommon.LabelsForNifiPKI(cluster.Name), cluster)
	issuerMeta.Namespace = cluster.Namespace
	issuer := &certv1.Issuer{
		ObjectMeta: issuerMeta,
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				CA: &certv1.CAIssuer{
					SecretName: fmt.Sprintf(pkicommon.NodeCACertTemplate, cluster.Name),
				},
			},
		},
	}
	controllerutil.SetControllerReference(cluster, issuer, scheme)
	return issuer
}
