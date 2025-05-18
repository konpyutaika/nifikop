/*
Copyright 2024.

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

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/processgroup"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var resourceFinalizer string = fmt.Sprintf("nifiresources.%s/finalizer", v1alpha1.GroupVersion.Group)

// NifiResourceReconciler reconciles a NifiResource object
type NifiResourceReconciler struct {
	client.Client
	Log             zap.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiresources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiResource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *NifiResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	interval := util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset)
	var err error

	// Fetch the NifiResource instance
	instance := &v1alpha1.NifiResource{}
	if err = r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return Reconciled()
		}
		// Error reading the object - requeue the request.
		return RequeueWithError(r.Log, err.Error(), err)
	}

	patchInstance := client.MergeFrom(instance.DeepCopy())
	// Get the last configuration viewed by the operator.
	o, _ := patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	// Create it if not exist.
	if o == nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for resource "+instance.Name, err)
		}

		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiResource "+instance.Name, err)
		}
		o, _ = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Check if the cluster reference changed.
	original := &v1alpha1.NifiResource{}
	current := instance.DeepCopy()
	patchCurrent := client.MergeFrom(current.DeepCopy())
	json.Unmarshal(o, original)
	if !v1.ClusterRefsEquals([]v1.ClusterReference{original.Spec.ClusterRef, instance.Spec.ClusterRef}) {
		instance.Spec.ClusterRef = original.Spec.ClusterRef
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
				zap.String("resource", instance.Name),
				zap.String("clusterName", clusterRef.Name))
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer for resource "+instance.Name, err)
			}
			return Reconciled()
		}
		// If the referenced cluster no more exist, just skip the deletion requirement in cluster ref change case.
		if !v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
				return RequeueWithError(r.Log, "could not apply last state to annotation to resource "+instance.Name, err)
			}
			if err := r.Client.Patch(ctx, current, patchCurrent); err != nil {
				return RequeueWithError(r.Log, "failed to update NifiResource "+instance.Name, err)
			}
			return RequeueAfter(interval)
		}

		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to lookup reference cluster: %s in %s",
				instance.Spec.ClusterRef.Name, clusterRef.Namespace))
		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to lookup referenced cluster for resource "+instance.Name, err)
	}

	// Generate the client configuration.
	clientConfig, err = configManager.BuildConfig()
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to create HTTP client for the referenced cluster: %s in %s",
				instance.Spec.ClusterRef.Name, clusterRef.Namespace))
		// the cluster is gone, so just remove the finalizer
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer from NifiResource "+instance.Name, err)
			}
			return Reconciled()
		}
		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to create HTTP client the for referenced cluster "+clusterRef.Name+" for resource "+instance.Name, err)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, clientConfig, patchInstance)
	}

	// Ensure the cluster is ready to receive actions
	if !clusterConnect.IsReady(r.Log) {
		r.Log.Debug("Cluster is not ready yet, will wait until it is.",
			zap.String("resource", instance.Name),
			zap.String("clusterName", clusterRef.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, "ReferenceClusterNotReady",
			fmt.Sprintf("The referenced cluster is not ready yet: %s in %s",
				instance.Spec.ClusterRef.Name, clusterConnect.Id()))
		// the cluster does not exist - should have been caught pre-flight
		return RequeueAfter(interval)
	}

	// Ìn case of the cluster reference changed.
	if !v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
		// Delete the resource on the previous cluster.
		if err := r.removeResource(instance, clientConfig); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "RemoveError",
				fmt.Sprintf("Failed to delete NifiResource %s from cluster %s before moving in %s",
					instance.Name, original.Spec.ClusterRef.Name, original.Spec.ClusterRef.Name))
			return RequeueWithError(r.Log, "Failed to delete NifiResource before moving", err)
		}
		// Update the last view configuration to the current one.
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for registry client "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, current, patchCurrent); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiRegistryClient "+instance.Name, err)
		}
		return RequeueAfter(interval)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciling",
		"Reconciling resource "+instance.Name)

	// Check if the NiFi registry client already exist
	exist, err := r.existResource(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing resource "+instance.Name, err)
	}

	if !exist {
		// Create NiFi resource
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Creating",
			fmt.Sprintf("Creating resource %s", instance.Name))
		status, err := r.createResource(instance, clientConfig)
		if err != nil {
			return RequeueWithError(r.Log, "failure creating resource "+instance.Name, err)
		}

		instance.Status = *status
		if err := r.updateStatus(ctx, instance, current.Status); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiResource "+instance.Name, err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Created",
			fmt.Sprintf("Created resource %s", instance.Name))
		r.Log.Info("Created resource",
			zap.String("resource", instance.Name))

		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for resource "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiResource "+instance.Name, err)
		}
	}

	// Sync NifiResource resource with NiFi side component
	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronizing",
		fmt.Sprintf("Synchronizing registry client %s", instance.Name))
	status, err := r.syncResource(instance, clientConfig)
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "SynchronizingFailed",
			fmt.Sprintf("Synchronizing registry client %s failed", instance.Name))
		return RequeueWithError(r.Log, "failed to sync NifiRegistryClient "+instance.Name, err)
	}

	instance.Status = *status
	if err := r.updateStatus(ctx, instance, current.Status); err != nil {
		return RequeueWithError(r.Log, "failed to update status for NifiResource "+instance.Name, err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
		fmt.Sprintf("Synchronized resource %s", instance.Name))
	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiResource label on resource "+current.Name, err)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), resourceFinalizer) {
		r.Log.Debug("Adding Finalizer for NifiResource",
			zap.String("resource", instance.Name))
		instance.SetFinalizers(append(instance.GetFinalizers(), resourceFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiResource "+current.Name, err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciled",
		fmt.Sprintf("Reconciling resource %s", instance.Name))

	r.Log.Debug("Ensured Resource",
		zap.String("resource", instance.Name))

	return RequeueAfter(interval)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiResource{}).
		Complete(r)
}

func (r *NifiResourceReconciler) ensureClusterLabel(ctx context.Context, cluster clientconfig.ClusterConnect,
	resource *v1alpha1.NifiResource, patcher client.Patch) (*v1alpha1.NifiResource, error) {
	labels := ApplyClusterReferenceLabel(cluster, resource.GetLabels())
	if !reflect.DeepEqual(labels, resource.GetLabels()) {
		resource.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, resource, patcher)
	}
	return resource, nil
}

func (r *NifiResourceReconciler) updateAndFetchLatest(ctx context.Context,
	resource *v1alpha1.NifiResource, patcher client.Patch) (*v1alpha1.NifiResource, error) {
	typeMeta := resource.TypeMeta
	err := r.Client.Patch(ctx, resource, patcher)
	if err != nil {
		return nil, err
	}
	resource.TypeMeta = typeMeta
	return resource, nil
}

func (r *NifiResourceReconciler) checkFinalizers(ctx context.Context,
	resource *v1alpha1.NifiResource, config *clientconfig.NifiConfig, patcher client.Patch) (reconcile.Result, error) {
	r.Log.Info("NiFi resource is marked for deletion. Removing finalizers.",
		zap.String("resource", resource.Name))
	var err error
	if util.StringSliceContains(resource.GetFinalizers(), resourceFinalizer) {
		if err = r.finalizeNifiResource(resource, config); err != nil {
			return RequeueWithError(r.Log, "failed to finalize NifiResource", err)
		}
		if err = r.removeFinalizer(ctx, resource, patcher); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from NifiResource", err)
		}
	}
	return Reconciled()
}

func (r *NifiResourceReconciler) removeFinalizer(ctx context.Context, resource *v1alpha1.NifiResource, patcher client.Patch) error {
	r.Log.Debug("Removing finalizer for NifiResource",
		zap.String("resource", resource.Name))
	resource.SetFinalizers(util.StringSliceRemove(resource.GetFinalizers(), resourceFinalizer))
	_, err := r.updateAndFetchLatest(ctx, resource, patcher)
	return err
}

func (r *NifiResourceReconciler) finalizeNifiResource(resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) error {
	if err := r.removeResource(resource, config); err != nil {
		return err
	}

	r.Log.Info("Deleted Resource",
		zap.String("resource", resource.Name))

	return nil
}

func (r *NifiResourceReconciler) updateStatus(ctx context.Context, resource *v1alpha1.NifiResource, currentStatus v1alpha1.NifiResourceStatus) error {
	if !reflect.DeepEqual(resource.Status, currentStatus) {
		return r.Client.Status().Update(ctx, resource)
	}
	return nil
}

func (r *NifiResourceReconciler) existResource(resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) (bool, error) {
	var err error
	exist := false

	if resource.Spec.IsProcessGroup() {
		if exist, err = processgroup.ExistProcessGroup(resource, config); err != nil {
			return exist, err
		}
	}

	return exist, nil
}

func (r *NifiResourceReconciler) createResource(resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	var err error
	status := &v1alpha1.NifiResourceStatus{}

	if resource.Spec.IsProcessGroup() {
		if status, err = processgroup.CreateProcessGroup(resource, config); err != nil {
			return status, err
		}
	}

	return status, nil
}

func (r *NifiResourceReconciler) syncResource(resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	var err error
	status := &v1alpha1.NifiResourceStatus{}

	if resource.Spec.IsProcessGroup() {
		if status, err = processgroup.SyncProcessGroup(resource, config); err != nil {
			return status, err
		}
	}

	return status, nil
}

func (r *NifiResourceReconciler) removeResource(resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) error {
	var err error

	if resource.Spec.IsProcessGroup() {
		if err = processgroup.RemoveProcessGroup(resource, config); err != nil {
			return err
		}
	}

	return nil
}
