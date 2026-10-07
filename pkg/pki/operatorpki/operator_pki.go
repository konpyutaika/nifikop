package operatorpki

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	pkicommon "github.com/konpyutaika/nifikop/pkg/util/pki"
)

// ReconcilePKI ensures the cluster CA exists and that a NifiUser is present for the
// controller and for every node. The NifiUser controller then calls back into
// ReconcileUserCertificate for each one, which is where certificates are actually
// issued. No cert-manager resource is created or read anywhere on this path.
func (o *operatorPKI) ReconcilePKI(ctx context.Context, logger zap.Logger, scheme *runtime.Scheme,
	externalHostnames []string,
) error {
	logger.Info("Reconciling operator-managed PKI",
		zap.String("clusterName", o.cluster.Name))

	if err := o.supported(); err != nil {
		return err
	}

	if _, err := o.ensureCA(ctx, scheme); err != nil {
		return err
	}

	objects := []*v1.NifiUser{pkicommon.ControllerUserForCluster(o.cluster)}
	objects = append(objects, pkicommon.NodeUsersForCluster(o.cluster, externalHostnames)...)

	for _, user := range objects {
		if err := o.reconcileUser(ctx, logger, user); err != nil {
			return err
		}
	}

	return nil
}

// FinalizePKI removes the material this backend created: the CA, or its copy of a
// user-supplied CA, and the controller and node certificates.
//
// Every one of those secrets lives at a predictable name, so a secret is deleted only
// while it still carries the owner this backend gave it. Anything else found at one of
// those names - the user's own CA secret included - belongs to someone else. The
// cert-manager backend gets the same protection implicitly, by deleting a secret only
// when its same-named Certificate exists.
func (o *operatorPKI) FinalizePKI(ctx context.Context, logger zap.Logger) error {
	logger.Info("Removing operator-managed certificates and secrets",
		zap.String("clusterName", o.cluster.Name))

	if o.cluster.GetSSLSecrets() == nil {
		return nil
	}

	type target struct {
		name  string
		owned func(*corev1.Secret) bool
	}
	targets := []target{
		{o.caSecretName(), func(secret *corev1.Secret) bool { return metav1.IsControlledBy(secret, o.cluster) }},
		{o.cluster.GetNifiControllerUserIdentity(), userOwned(o.cluster.GetNifiControllerUserIdentity())},
	}
	for _, node := range o.cluster.Spec.Nodes {
		targets = append(targets, target{
			fmt.Sprintf(pkicommon.NodeServerCertTemplate, o.cluster.Name, node.Id),
			userOwned(pkicommon.GetNodeUserName(o.cluster, node.Id)),
		})
	}

	for _, t := range targets {
		secret := &corev1.Secret{}
		if err := o.client.Get(ctx, types.NamespacedName{Name: t.name, Namespace: o.cluster.Namespace}, secret); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return err
		}
		if !t.owned(secret) {
			continue
		}
		// Delete only the object whose ownership was just checked. If it was replaced or
		// rewritten in between, the API server refuses and the next pass checks it again.
		uid, resourceVersion := secret.UID, secret.ResourceVersion
		if err := o.client.Delete(ctx, secret,
			client.Preconditions{UID: &uid, ResourceVersion: &resourceVersion}); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// userOwned matches a certificate secret controlled by the named NifiUser. The user is
// compared by name rather than UID: on teardown it has usually been deleted already.
func userOwned(name string) func(*corev1.Secret) bool {
	return func(secret *corev1.Secret) bool {
		return controlledBy(secret, nifiUserKind, name)
	}
}

// reconcileUser ensures a NifiUser exists, matching the cert-manager backend's behaviour.
func (o *operatorPKI) reconcileUser(ctx context.Context, logger zap.Logger, user *v1.NifiUser) error {
	current := &v1.NifiUser{}
	err := o.client.Get(ctx, types.NamespacedName{Name: user.Name, Namespace: user.Namespace}, current)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return errorfactory.New(errorfactory.APIFailure{}, err, "could not look up nifi user")
		}
		if err := o.client.Create(ctx, user); err != nil {
			return errorfactory.New(errorfactory.APIFailure{}, err, "could not create nifi user")
		}
		return nil
	}
	return k8sutil.Reconcile(logger, o.client, user, o.cluster, &o.cluster.Status)
}

// supported refuses configuration this backend cannot honour, rather than quietly doing
// something else. issuerRef names a cert-manager Issuer: silently replacing it with a
// CA this backend generates would change the cluster's trust root without anyone
// choosing that.
func (o *operatorPKI) supported() error {
	sslSecrets := o.cluster.GetSSLSecrets()
	if sslSecrets == nil {
		return errorfactory.New(errorfactory.InternalError{},
			errors.New("cluster has no sslSecrets"), "the operator PKI backend needs sslSecrets")
	}
	if ref := sslSecrets.IssuerRef; ref != nil {
		return errorfactory.New(errorfactory.InternalError{},
			fmt.Errorf("sslSecrets.issuerRef %s %q requires cert-manager, which this cluster is not using",
				ref.Kind, ref.Name),
			"sslSecrets.issuerRef is not supported by the operator PKI backend")
	}
	return nil
}
