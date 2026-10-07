package operatorpki

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
)

// Every secret this backend touches lives at a predictable name, so something else can
// already be there. The rule, applied wherever a secret is read for signing, written or
// deleted: a secret another object controls is never used, modified or removed. Writing
// to it would fight that object's controller, and deleting it would destroy something
// that is not ours.

// claim makes owner the controller of secret, ready to be written. It refuses a secret
// that something else already controls.
func claim(owner metav1.Object, secret *corev1.Secret, scheme *runtime.Scheme) error {
	err := controllerutil.SetControllerReference(owner, secret, scheme)
	if err == nil {
		return nil
	}
	if k8sutil.IsAlreadyOwnedError(err) {
		return errorfactory.New(errorfactory.InternalError{}, err,
			fmt.Sprintf("secret %s is controlled by another object; refusing to use or modify it", secret.Name))
	}
	return errorfactory.New(errorfactory.InternalError{}, err,
		fmt.Sprintf("could not set controller reference on secret %s", secret.Name))
}

// refuseForeign is claim without the side effect: it reports whether owner could take
// the secret, without changing it.
func refuseForeign(owner metav1.Object, secret *corev1.Secret, scheme *runtime.Scheme) error {
	return claim(owner, secret.DeepCopy(), scheme)
}

// nifiUserKind is the controller a user certificate secret is expected to have.
var nifiUserKind = v1.GroupVersion.WithKind("NifiUser")

// controlledBy reports whether a secret's controller is the named object of the given
// kind, compared as controller-runtime compares owners: by group, kind and name. It is
// used where the owner may already be gone, so its UID cannot be checked.
func controlledBy(secret *corev1.Secret, gvk schema.GroupVersionKind, name string) bool {
	ref := metav1.GetControllerOf(secret)
	if ref == nil {
		return false
	}
	gv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return false
	}
	return gv.Group == gvk.Group && ref.Kind == gvk.Kind && ref.Name == name
}
