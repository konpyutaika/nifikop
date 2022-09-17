package certmanagerpki

import (
	"context"
	"fmt"
	"reflect"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// reconcile ensures the given kubernetes object
func reconcile(ctx context.Context, log zap.Logger, client client.Client, object runtime.Object, cluster *v1alpha1.NifiCluster) (err error) {
	switch object.(type) {
	case *certv1.Issuer:
		issuer, _ := object.(*certv1.Issuer)
		return reconcileIssuer(ctx, log, client, issuer, cluster)
	case *certv1.ClusterIssuer:
		issuer, _ := object.(*certv1.ClusterIssuer)
		return reconcileClusterIssuer(ctx, log, client, issuer, cluster)
	case *certv1.Certificate:
		cert, _ := object.(*certv1.Certificate)
		return reconcileCertificate(ctx, log, client, cert, cluster)
	case *corev1.Secret:
		secret, _ := object.(*corev1.Secret)
		return reconcileSecret(ctx, log, client, secret, cluster)
	case *v1alpha1.NifiUser:
		user, _ := object.(*v1alpha1.NifiUser)
		return reconcileUser(ctx, log, client, user, cluster)
	default:
		panic(fmt.Sprintf("Invalid object type: %v", reflect.TypeOf(object)))
	}
}

// reconcileClusterIssuer ensures a cert-manager ClusterIssuer
func reconcileClusterIssuer(ctx context.Context, log zap.Logger, client client.Client, issuer *certv1.ClusterIssuer, cluster *v1alpha1.NifiCluster) error {
	obj := &certv1.ClusterIssuer{}
	var err error
	if err = client.Get(ctx, types.NamespacedName{Name: issuer.Name, Namespace: issuer.Namespace}, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return client.Create(ctx, issuer)
	}
	return nil
}

// reconcileIssuer ensures a cert-manager Issuer
func reconcileIssuer(ctx context.Context, log zap.Logger, client client.Client, issuer *certv1.Issuer, cluster *v1alpha1.NifiCluster) error {
	obj := &certv1.Issuer{}
	var err error
	if err = client.Get(ctx, types.NamespacedName{Name: issuer.Name, Namespace: issuer.Namespace}, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return client.Create(ctx, issuer)
	}
	return nil
}

// reconcileCertificate ensures a cert-manager certificate
func reconcileCertificate(ctx context.Context, log zap.Logger, client client.Client, cert *certv1.Certificate, cluster *v1alpha1.NifiCluster) error {
	obj := &certv1.Certificate{}
	var err error
	if err = client.Get(ctx, types.NamespacedName{Name: cert.Name, Namespace: cert.Namespace}, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return client.Create(ctx, cert)
	}
	return nil
}

// reconcileSecret ensures a Kubernetes secret
func reconcileSecret(ctx context.Context, log zap.Logger, client client.Client, secret *corev1.Secret, cluster *v1alpha1.NifiCluster) error {
	obj := &corev1.Secret{}
	var err error
	if err = client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return client.Create(ctx, secret)
	}
	return nil
}

// reconcileUser ensures a v1alpha1.NifiUser
func reconcileUser(ctx context.Context, log zap.Logger, client client.Client, user *v1alpha1.NifiUser, cluster *v1alpha1.NifiCluster) error {
	obj := &v1alpha1.NifiUser{}
	var err error
	if err = client.Get(ctx, types.NamespacedName{Name: user.Name, Namespace: user.Namespace}, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return client.Create(ctx, user)
	}
	return nil
}
