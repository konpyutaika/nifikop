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
	"emperror.dev/errors"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/parametercontext"
	errorfactory "github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/Orange-OpenSource/nifikop/pkg/k8sutil"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
)

var parameterContextFinalizer = "nifiparametercontexts.nifi.orange.com/finalizer"

// NifiParameterContextReconciler reconciles a NifiParameterContext object
type NifiParameterContextReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifiparametercontexts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifiparametercontexts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifiparametercontexts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiParameterContext object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *NifiParameterContextReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("nifiparametercontext", req.NamespacedName)

	var err error

	// Fetch the NifiParameterContext instance
	instance := &v1alpha1.NifiParameterContext{}
	if err = r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return Reconciled()
		}
		// Error reading the object - requeue the request.
		return RequeueWithError(r.Log, err.Error(), err)
	}

	// Get the referenced secrets
	var parameterSecrets []*corev1.Secret
	for _, parameterSecret := range instance.Spec.SecretRefs {
		secretNamespace := GetSecretRefNamespace(instance.Namespace, parameterSecret)
		var secret *corev1.Secret
		if secret, err = k8sutil.LookupSecret(r.Client, parameterSecret.Name, secretNamespace); err != nil {
			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				r.Log.Info("Secret is already gone, there is nothing we can do")
				if err = r.removeFinalizer(ctx, instance); err != nil {
					return RequeueWithError(r.Log, "failed to remove finalizer", err)
				}
				return Reconciled()
			}

			// the cluster does not exist - should have been caught pre-flight
			return RequeueWithError(r.Log, "failed to lookup referenced secret", err)
		}
		parameterSecrets = append(parameterSecrets, secret)
	}

	// Get the referenced NifiCluster
	clusterNamespace := GetClusterRefNamespace(instance.Namespace, instance.Spec.ClusterRef)
	var cluster *v1alpha1.NifiCluster
	if cluster, err = k8sutil.LookupNifiCluster(r.Client, instance.Spec.ClusterRef.Name, clusterNamespace); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			r.Log.Info("Cluster is already gone, there is nothing we can do")
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer", err)
			}
			return Reconciled()
		}

		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to lookup referenced cluster", err)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, parameterSecrets, cluster)
	}

	// Check if the NiFi registry client already exist
	exist, err := parametercontext.ExistParameterContext(r.Client, instance, cluster)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing parameter context", err)
	}

	if !exist {
		// Create NiFi parameter context
		status, err := parametercontext.CreateParameterContext(r.Client, instance, parameterSecrets, cluster)
		if err != nil {
			return RequeueWithError(r.Log, "failure creating parameter context", err)
		}

		instance.Status = *status
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiParameterContext status", err)
		}
	}

	// Sync ParameterContext resource with NiFi side component
	status, err := parametercontext.SyncParameterContext(r.Client, instance, parameterSecrets, cluster)
	if status != nil {
		instance.Status = *status
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiParameterContext status", err)
		}
	}
	if err != nil {
		switch errors.Cause(err).(type) {
		case errorfactory.NifiParameterContextUpdateRequestRunning:
			return RequeueAfter(time.Duration(5) * time.Second)
		default:
			return RequeueWithError(r.Log, "failed to sync NifiParameterContext", err)
		}
	}

	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, cluster, instance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiCluster label on parameter context", err)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), parameterContextFinalizer) {
		r.Log.Info("Adding Finalizer for NifiParameterContext")
		instance.SetFinalizers(append(instance.GetFinalizers(), parameterContextFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiParameterContext", err)
	}

	r.Log.Info("Ensured Parameter Context")

	return RequeueAfter(time.Duration(15) * time.Second)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiParameterContextReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiParameterContext{}).
		Complete(r)
}

func (r *NifiParameterContextReconciler) ensureClusterLabel(ctx context.Context, cluster *v1alpha1.NifiCluster,
	parameterContext *v1alpha1.NifiParameterContext) (*v1alpha1.NifiParameterContext, error) {

	labels := ApplyClusterRefLabel(cluster, parameterContext.GetLabels())
	if !reflect.DeepEqual(labels, parameterContext.GetLabels()) {
		parameterContext.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, parameterContext)
	}
	return parameterContext, nil
}

func (r *NifiParameterContextReconciler) updateAndFetchLatest(ctx context.Context,
	parameterContext *v1alpha1.NifiParameterContext) (*v1alpha1.NifiParameterContext, error) {

	typeMeta := parameterContext.TypeMeta
	err := r.Client.Update(ctx, parameterContext)
	if err != nil {
		return nil, err
	}
	parameterContext.TypeMeta = typeMeta
	return parameterContext, nil
}

func (r *NifiParameterContextReconciler) checkFinalizers(
	ctx context.Context,
	parameterContext *v1alpha1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	cluster *v1alpha1.NifiCluster) (reconcile.Result, error) {

	r.Log.Info("NiFi parameter context is marked for deletion")
	var err error
	if util.StringSliceContains(parameterContext.GetFinalizers(), parameterContextFinalizer) {
		if err = r.finalizeNifiParameterContext(parameterContext, parameterSecrets, cluster); err != nil {
			return RequeueWithError(r.Log, "failed to finalize parameter context", err)
		}
		if err = r.removeFinalizer(ctx, parameterContext); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from parameter context", err)
		}
	}
	return Reconciled()
}

func (r *NifiParameterContextReconciler) removeFinalizer(ctx context.Context, flow *v1alpha1.NifiParameterContext) error {
	flow.SetFinalizers(util.StringSliceRemove(flow.GetFinalizers(), parameterContextFinalizer))
	_, err := r.updateAndFetchLatest(ctx, flow)
	return err
}

func (r *NifiParameterContextReconciler) finalizeNifiParameterContext(
	parameterContext *v1alpha1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	cluster *v1alpha1.NifiCluster) error {

	if err := parametercontext.RemoveParameterContext(r.Client, parameterContext, parameterSecrets, cluster); err != nil {
		return err
	}
	r.Log.Info("Delete Registry client")

	return nil
}
