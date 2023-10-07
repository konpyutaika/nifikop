/*
Copyright 2020.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"reflect"

	v1 "github.com/konpyutaika/nifikop/api/v1"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NifiParameterContextSecretReconciler reconciles
type NifiParameterContextSecretReconciler struct {
	client.Client
	Log             zap.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

const (
	secretRefField = ".spec.secretRef"
)

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiUserGroup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *NifiParameterContextSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the NifiParameter instance
	instance := &v1.NifiParameterContext{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return Reconciled()
		}
		// Error reading the object - requeue the request.
		return RequeueWithError(r.Log, err.Error(), err)
	}

	current := instance.DeepCopy()
	instance.Status.SecretsState = v1.ParameterContextSecretStateOutOfDate
	if err := r.updateStatus(ctx, instance, current.Status); err != nil {
		return RequeueWithError(r.Log, "failed to update status for NifiParameterContext "+instance.Name, err)
	}
	return Reconciled()
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiParameterContextSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logCtr, err := GetLogConstructor(mgr, &v1.NifiParameterContext{})
	if err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1.NifiParameterContext{}, secretRefField, func(rawObj client.Object) []string {
		parameterContext := rawObj.(*v1.NifiParameterContext)
		if len(parameterContext.Spec.SecretRefs) == 0 {
			return nil
		}
		secretRefs := make([]string, len(parameterContext.Spec.SecretRefs))
		for i, secretRef := range parameterContext.Spec.SecretRefs {
			secretRefs[i] = fmt.Sprintf("%s;%s", secretRef.Name, GetSecretRefNamespace(parameterContext.Namespace, secretRef))
		}
		return secretRefs
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("nifiparametercontext").
		WithLogConstructor(logCtr).
		Watches(&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSecret),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}

func (r *NifiParameterContextSecretReconciler) findObjectsForSecret(secret client.Object) []reconcile.Request {
	attachedNifiParameterContext := &v1.NifiParameterContextList{}
	if secret.GetNamespace() != "instances" {
		return []reconcile.Request{}
	}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(secretRefField, fmt.Sprintf("%s;%s", secret.GetName(), secret.GetNamespace())),
	}
	err := r.List(context.TODO(), attachedNifiParameterContext, listOps)
	if err != nil {
		return []reconcile.Request{}
	}
	requests := make([]reconcile.Request, len(attachedNifiParameterContext.Items))
	for i, item := range attachedNifiParameterContext.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}

	return requests
}

func (r *NifiParameterContextSecretReconciler) updateStatus(ctx context.Context, parameterContext *v1.NifiParameterContext, currentStatus v1.NifiParameterContextStatus) error {
	if !reflect.DeepEqual(parameterContext.Status, currentStatus) {
		return r.Client.Status().Update(ctx, parameterContext)
	}
	return nil
}
