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
	"encoding/json"
	"fmt"
	"reflect"

	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/parametercontext"
	errorfactory "github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var parameterContextFinalizer = fmt.Sprintf("nifiparametercontexts.%s/finalizer", v1.GroupVersion.Group)

// NifiParameterContextReconciler reconciles a NifiParameterContext object.
type NifiParameterContextReconciler struct {
	client.Client
	Log             zap.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiparametercontexts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiparametercontexts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiparametercontexts/finalizers,verbs=update

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
	interval := util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset)
	var err error

	// Fetch the NifiParameterContext instance
	instance := &v1.NifiParameterContext{}
	if err = r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return Reconciled()
		}
		// Error reading the object - requeue the request.
		return RequeueWithError(r.Log, err.Error(), err)
	}

	patchInstance := client.MergeFromWithOptions(instance.DeepCopy(), client.MergeFromWithOptimisticLock{})
	// Get the last configuration viewed by the operator.
	o, _ := patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	// Create it if not exist.
	if o == nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for parameter context "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiParameterContext "+instance.Name, err)
		}
		o, _ = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Check if the cluster reference changed.
	original := &v1.NifiParameterContext{}
	current := instance.DeepCopy()
	patchCurrent := client.MergeFromWithOptions(current.DeepCopy(), client.MergeFromWithOptimisticLock{})
	json.Unmarshal(o, original)
	if !v1.ClusterRefsEquals([]v1.ClusterReference{original.Spec.ClusterRef, instance.Spec.ClusterRef}) {
		instance.Spec.ClusterRef = original.Spec.ClusterRef
	}

	// Get the referenced secrets
	var parameterSecrets []*corev1.Secret
	for _, parameterSecret := range instance.Spec.SecretRefs {
		secretNamespace := GetSecretRefNamespace(instance.Namespace, parameterSecret)
		var secret *corev1.Secret
		if secret, err = k8sutil.LookupSecret(r.Client, parameterSecret.Name, secretNamespace); err != nil {
			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				r.Log.Error("Secret for parameter context is already gone, there is nothing we can do",
					zap.String("secretName", parameterSecret.Name),
					zap.String("secretNamespace", parameterSecret.Namespace),
					zap.String("parameterContext", instance.Name))
				if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
					return RequeueWithError(r.Log, "failed to remove finalizer for parameter context "+instance.Name, err)
				}
				return Reconciled()
			}

			// the cluster does not exist - should have been caught pre-flight
			return RequeueWithError(r.Log, "failed to lookup referenced secret for parameter context "+instance.Name, err)
		}
		parameterSecrets = append(parameterSecrets, secret)
	}

	// Get the referenced NiFiParameterContext referenced
	var parameterContextRefs []*v1.NifiParameterContext
	for _, parameterContextRef := range instance.Spec.InheritedParameterContexts {
		parameterContextNamespace := GetParameterContextRefNamespace(instance.Namespace, parameterContextRef)
		var parameterContext *v1.NifiParameterContext
		if parameterContext, err = k8sutil.LookupNifiParameterContext(r.Client, parameterContextRef.Name, parameterContextNamespace); err != nil {
			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				r.Log.Info("Secret is already gone, there is nothing we can do")
				if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
					return RequeueWithError(r.Log, "failed to remove finalizer", err)
				}
				return Reconciled()
			}

			// the cluster does not exist - should have been caught pre-flight
			return RequeueWithError(r.Log, "failed to lookup referenced parameter context", err)
		}
		parameterContextRefs = append(parameterContextRefs, parameterContext)
	}

	// Prepare cluster connection configurations
	var clientConfig *clientconfig.NifiConfig
	var clusterConnect clientconfig.ClusterConnect

	// Get the client config manager associated to the cluster ref.
	clusterRef := instance.Spec.ClusterRef
	clusterRef.Namespace = GetClusterRefNamespace(instance.Namespace, instance.Spec.ClusterRef)
	configManager := config.GetClientConfigManager(r.Client, clusterRef)

	// Generate the connect object
	if clusterConnect, err = configManager.BuildConnect(); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			r.Log.Error("Cluster is already gone, there is nothing we can do",
				zap.String("clusterName", clusterRef.Name),
				zap.String("parameterContext", instance.Name))
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer for parameter context "+instance.Name, err)
			}
			return Reconciled()
		}
		// If the referenced cluster no more exist, just skip the deletion requirement in cluster ref change case.
		if !v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
				return RequeueWithError(r.Log, "could not apply last state to annotation for parameter context "+instance.Name, err)
			}
			if err := r.Client.Patch(ctx, current, patchCurrent); err != nil {
				return RequeueWithError(r.Log, "failed to update NifiParameterContext "+instance.Name, err)
			}
			return RequeueAfter(interval)
		}

		msg := fmt.Sprintf("Failed to lookup reference cluster for parameter context %s : %s in %s",
			instance.Name, instance.Spec.ClusterRef.Name, clusterRef.Namespace)
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError", msg)

		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, msg, err)
	}

	// Generate the client configuration.
	clientConfig, err = configManager.BuildConfig()
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to create HTTP client for the referenced cluster for parameter context %s : %s in %s",
				instance.Name, instance.Spec.ClusterRef.Name, clusterRef.Namespace))
		// the cluster is gone, so just remove the finalizer
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, fmt.Sprintf("failed to remove finalizer from NifiParameterContext %s", instance.Name), err)
			}
			return Reconciled()
		}
		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to create HTTP client the for referenced cluster", err)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, parameterSecrets, parameterContextRefs, clientConfig, patchInstance)
	}

	// Ensure the cluster is ready to receive actions
	if !clusterConnect.IsReady(r.Log) {
		r.Log.Debug("Cluster is not ready yet, will wait until it is.",
			zap.String("clusterName", clusterRef.Name),
			zap.String("parameterContext", instance.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, "ReferenceClusterNotReady",
			fmt.Sprintf("The referenced cluster is not ready yet : %s in %s",
				instance.Spec.ClusterRef.Name, clusterConnect.Id()))

		// the cluster does not exist - should have been caught pre-flight
		return RequeueAfter(interval)
	}

	// Ìn case of the cluster reference changed.
	if !v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
		// Delete the resource on the previous cluster.
		if err := parametercontext.RemoveParameterContext(instance, parameterSecrets, parameterContextRefs, clientConfig); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "RemoveError",
				fmt.Sprintf("Failed to delete NifiParameterContext %s from cluster %s before moving in %s",
					instance.Name, original.Spec.ClusterRef.Name, original.Spec.ClusterRef.Name))
			return RequeueWithError(r.Log, "Failed to delete NifiParameterContext before moving "+instance.Name, err)
		}
		// Update the last view configuration to the current one.
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for parameter context "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, current, patchCurrent); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiParameterContext "+instance.Name, err)
		}
		return RequeueAfter(interval)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciling",
		fmt.Sprintf("Reconciling parameter context %s", instance.Name))

	// Check if the NiFi parameter context already exist
	exist, err := parametercontext.ExistParameterContext(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing parameter context with name "+instance.Name, err)
	}

	if !exist {
		// Create NiFi parameter context
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Creating",
			fmt.Sprintf("Creating parameter context %s", instance.Name))

		var status *v1.NifiParameterContextStatus

		status, err = parametercontext.FindParameterContextByName(instance, clientConfig)
		if err != nil {
			return RequeueWithError(r.Log, "failure finding parameter context "+instance.Name, err)
		}

		if status != nil && !instance.Spec.IsTakeOverEnabled() {
			// TakeOver disabled
			return RequeueWithError(r.Log, fmt.Sprintf("parameter context name %s already used and takeOver disabled", instance.Name), err)
		}
		if status == nil {
			// Create NiFi parameter context
			status, err = parametercontext.CreateParameterContext(instance, parameterSecrets, parameterContextRefs, clientConfig)
			if err != nil {
				return RequeueWithError(r.Log, "failure creating parameter context "+instance.Name, err)
			}
		}

		instance.Status = *status
		if err := r.updateStatus(ctx, instance, current.Status); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiParameterContext "+instance.Name, err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Created",
			fmt.Sprintf("Created parameter context %s", instance.Name))
	}

	// Sync ParameterContext resource with NiFi side component
	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronizing",
		fmt.Sprintf("Synchronizing parameter context %s", instance.Name))
	status, err := parametercontext.SyncParameterContext(instance, parameterSecrets, parameterContextRefs, clientConfig)
	if status != nil {
		instance.Status = *status
		if err := r.updateStatus(ctx, instance, current.Status); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiParameterContext "+instance.Name, err)
		}
	}
	if err != nil {
		switch errors.Cause(err).(type) {
		case errorfactory.NifiParameterContextUpdateRequestRunning:
			return RequeueAfter(interval)
		case errorfactory.NifiParameterContextUpdateRequestNotFound:
			r.Log.Warn("The update request for parameter context is already gone, there is nothing we can do",
				zap.String("updateRequest", instance.Status.LatestUpdateRequest.Id),
				zap.String("parameterContext", instance.Name))
			return RequeueAfter(interval)
		default:
			r.Recorder.Event(instance, corev1.EventTypeNormal, "SynchronizingFailed",
				fmt.Sprintf("Synchronizing parameter context %s failed", instance.Name))
			return RequeueWithError(r.Log, "failed to sync NifiParameterContext "+instance.Name, err)
		}
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
		fmt.Sprintf("Synchronized parameter context %s", instance.Name))

	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiCluster label on parameter context "+current.Name, err)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), parameterContextFinalizer) {
		r.Log.Debug("Adding Finalizer for NifiParameterContext",
			zap.String("parameterContext", instance.Name))
		instance.SetFinalizers(append(instance.GetFinalizers(), parameterContextFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiParameterContext "+current.Name, err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciled",
		fmt.Sprintf("Reconciling parameter context %s", instance.Name))

	r.Log.Debug("Ensured Parameter Context",
		zap.String("parameterContext", instance.Name))

	return RequeueAfter(interval)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiParameterContextReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logCtr, err := GetLogConstructor(mgr, &v1.NifiParameterContext{})
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.NifiParameterContext{}).
		WithLogConstructor(logCtr).
		Complete(r)
}

func (r *NifiParameterContextReconciler) ensureClusterLabel(ctx context.Context, cluster clientconfig.ClusterConnect,
	parameterContext *v1.NifiParameterContext, patcher client.Patch) (*v1.NifiParameterContext, error) {
	labels := ApplyClusterReferenceLabel(cluster, parameterContext.GetLabels())
	if !reflect.DeepEqual(labels, parameterContext.GetLabels()) {
		parameterContext.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, parameterContext, patcher)
	}
	return parameterContext, nil
}

func (r *NifiParameterContextReconciler) updateAndFetchLatest(ctx context.Context,
	parameterContext *v1.NifiParameterContext, patcher client.Patch) (*v1.NifiParameterContext, error) {
	typeMeta := parameterContext.TypeMeta
	err := r.Client.Patch(ctx, parameterContext, patcher)
	if err != nil {
		return nil, err
	}
	parameterContext.TypeMeta = typeMeta
	return parameterContext, nil
}

func (r *NifiParameterContextReconciler) checkFinalizers(
	ctx context.Context,
	parameterContext *v1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	parameterContextRefs []*v1.NifiParameterContext,
	config *clientconfig.NifiConfig,
	patcher client.Patch) (reconcile.Result, error) {
	r.Log.Info("NiFi parameter context is marked for deletion. Removing finalizers.",
		zap.String("parameterContext", parameterContext.Name))
	var err error
	if util.StringSliceContains(parameterContext.GetFinalizers(), parameterContextFinalizer) {
		if err = r.finalizeNifiParameterContext(parameterContext, parameterSecrets, parameterContextRefs, config); err != nil {
			return RequeueWithError(r.Log, "failed to finalize parameter context "+parameterContext.Name, err)
		}
		if err = r.removeFinalizer(ctx, parameterContext, patcher); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from parameter context "+parameterContext.Name, err)
		}
	}
	return Reconciled()
}

func (r *NifiParameterContextReconciler) removeFinalizer(ctx context.Context, paramCtxt *v1.NifiParameterContext, patcher client.Patch) error {
	r.Log.Debug("Removing finalizer for NifiParameterContext",
		zap.String("paramaterContext", paramCtxt.Name))
	paramCtxt.SetFinalizers(util.StringSliceRemove(paramCtxt.GetFinalizers(), parameterContextFinalizer))
	_, err := r.updateAndFetchLatest(ctx, paramCtxt, patcher)
	return err
}

func (r *NifiParameterContextReconciler) finalizeNifiParameterContext(
	parameterContext *v1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	parameterContextRefs []*v1.NifiParameterContext,
	config *clientconfig.NifiConfig) error {
	if err := parametercontext.RemoveParameterContext(parameterContext, parameterSecrets, parameterContextRefs, config); err != nil {
		return err
	}
	r.Log.Info("Deleted NifiParameter Context",
		zap.String("parameterContext", parameterContext.Name))

	return nil
}

func (r *NifiParameterContextReconciler) updateStatus(ctx context.Context, parameterContext *v1.NifiParameterContext, currentStatus v1.NifiParameterContextStatus) error {
	if !reflect.DeepEqual(parameterContext.Status, currentStatus) {
		return r.Client.Status().Update(ctx, parameterContext)
	}
	return nil
}
