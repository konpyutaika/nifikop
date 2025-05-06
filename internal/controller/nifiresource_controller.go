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
	"reflect"

	"fmt"

	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"

	"github.com/konpyutaika/nifikop/pkg/clientwrappers/inputport"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/outputport"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/processgroup"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

var resourceFinalizer = fmt.Sprintf("nifiresources.%s/finalizer", v1.GroupVersion.Group)

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
	_ = log.FromContext(ctx)
	interval := util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset)
	var err error

	// Fetch the Nifiresource instance
	var instance = &v1alpha1.NifiResource{}
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
			return RequeueWithError(r.Log, "could not apply last state to annotation for NifiResource"+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiResource "+instance.Name, err)
		}
		o, _ = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	r.Log.Info("NifiResource starting reconciliation", zap.String("resourceName", instance.Name))

	// Check if the cluster reference changed.
	original := &v1alpha1.NifiResource{}
	current := instance.DeepCopy()
	patchCurrent := client.MergeFrom(current.DeepCopy())
	json.Unmarshal(o, original)
	if !v1.ClusterRefsEquals([]v1.ClusterReference{original.Spec.ClusterRef, instance.Spec.ClusterRef}) {
		instance.Spec.ClusterRef = original.Spec.ClusterRef
	}

	// Get Extra Configuration
	resourceConfig, err := instance.Spec.GetConfiguration()
	if err != nil {
		return RequeueWithError(r.Log, "failed to reteive configuration for resource "+instance.Name, err)
	}

	// Get Parameter Context
	var parameterContext *v1.NifiParameterContext
	var parameterContextNamespace string
	if val, err := util.DecodeMapToStruct[v1.ParameterContextReference](resourceConfig["parameterContextRef"]); err == nil && val != nil {
		parameterContextNamespace =
			GetParameterContextRefNamespace(current.Namespace, *val)
		if parameterContext, err = k8sutil.LookupNifiParameterContext(r.Client,
			val.Name, parameterContextNamespace); err != nil {
			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				r.Log.Info("Resource context is already gone, there is nothing we can do",
					zap.String("resource", instance.Name))
				if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
					return RequeueWithError(r.Log, "failed to remove finalizer for resource "+instance.Name, err)
				}
				return Reconciled()
			}

			msg := fmt.Sprintf("Failed to lookup reference parameter-context for resource %s: %s in %s",
				instance.Name, instance.Spec.ClusterRef.Name, parameterContextNamespace)
			r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceParameterContextError", msg)

			// the cluster does not exist - should have been caught pre-flight
			return RequeueWithError(r.Log, msg, err)
		}
	}

	var parentProcessGroup *v1alpha1.NifiResource
	var parentProcessGroupNamespace string
	if instance.Spec.ParentProcessGroupReference != nil {
		parentProcessGroupNamespace =
			GetResourceRefNamespace(current.Namespace, *current.Spec.ParentProcessGroupReference)

		if parentProcessGroup, err = k8sutil.LookupNifiResource(r.Client,
			current.Spec.ParentProcessGroupReference.Name, parentProcessGroupNamespace); err != nil {
			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				r.Log.Info("Dataflow is already gone, there is nothing we can do",
					zap.String("dataflow", instance.Name))
				if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
					return RequeueWithError(r.Log, "failed to remove finalizer for dataflow "+instance.Name, err)
				}
				return Reconciled()
			}

			msg := fmt.Sprintf("Failed to lookup reference parent process group for dataflow %s: %s in %s",
				instance.Name, current.Spec.ParentProcessGroupReference.Name, parentProcessGroupNamespace)
			r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceParentProcessGroupError", msg)

			return RequeueWithError(r.Log, msg, err)
		}

		if parentProcessGroup.Spec.Type != v1.ResourceProcessGroup {
			msg := fmt.Sprintf("Parent process group reference %s/%s Type is not ProcessGroup", current.Spec.ParentProcessGroupReference.Name, parentProcessGroupNamespace)
			r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceParentProcessGroupError", msg)

			return RequeueWithError(r.Log, msg, err)
		}

		// Stop self reference
		if parentProcessGroup.Name == instance.Name && parentProcessGroupNamespace == instance.Namespace {
			msg := fmt.Sprintf("Parent process group reference %s/%s is a self reference", current.Spec.ParentProcessGroupReference.Name, parentProcessGroupNamespace)
			r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceParentProcessGroupError", msg)

			return RequeueWithError(r.Log, msg, err)
		}

	}

	// Check if cluster references are the same
	var clusterRefs []v1.ClusterReference

	if parameterContext != nil {
		parameterContextClusterRef := parameterContext.Spec.ClusterRef
		parameterContextClusterRef.Namespace = parameterContextNamespace
		clusterRefs = append(clusterRefs, parameterContextClusterRef)
	}

	currentClusterRef := current.Spec.ClusterRef
	currentClusterRef.Namespace = GetClusterRefNamespace(current.Namespace, current.Spec.ClusterRef)
	clusterRefs = append(clusterRefs, currentClusterRef)

	if !v1.ClusterRefsEquals(clusterRefs) {
		msg := fmt.Sprintf("Failed to lookup reference cluster for resource %s: %s in %s",
			instance.Name, instance.Spec.ClusterRef.Name, currentClusterRef.Namespace)
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError", msg)

		return RequeueWithError(r.Log, msg, errors.New("inconsistent cluster references"))
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

	requeueMaintOp, err := r.MaintenanceOperation(ctx, instance, parentProcessGroup, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failed to perform maintenance operation on "+instance.Name, err)
	} else if requeueMaintOp {
		return RequeueAfter(interval / 3)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, parentProcessGroup, clientConfig, patchInstance)
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
		if _, err := r.RemoveNifiResource(instance, parentProcessGroup, clientConfig); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "RemoveError",
				fmt.Sprintf("Failed to delete NifiResource %s from cluster %s before moving in %s",
					instance.Name, original.Spec.ClusterRef.Name, original.Spec.ClusterRef.Name))
			return RequeueWithError(r.Log, "Failed to delete NifiResource before moving", err)
		}
		// Update the last view configuration to the current one.
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for resource "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, current, patchCurrent); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiResource "+instance.Name, err)
		}
		return RequeueAfter(interval)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciling",
		"Reconciling resource "+instance.Name)

	// Check if the NifiResource already exists
	exist, err := r.CheckNifiResourceExists(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing resource "+instance.Name, err)
	}

	// Create resource
	if !exist {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Creating",
			fmt.Sprintf("Creating resource %s", instance.Name))
		status, err := r.CreateNifiResource(instance, parameterContext, parentProcessGroup, clientConfig)
		if err != nil {
			return RequeueWithError(r.Log, "failure creating resource "+instance.Name, err)
		}

		instance.Status = *status
		instance.Status.State = v1alpha1.ResourceStateCreated

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

		exist = true
	}

	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiCluster label on resource "+current.Name, err)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), resourceFinalizer) {
		r.Log.Debug("Adding Finalizer for NifiResource",
			zap.String("resource", instance.Name))
		instance.SetFinalizers(append(instance.GetFinalizers(), resourceFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiDataflow "+current.Name, err)
	}

	if instance.Status.State == v1alpha1.ResourceStateOutOfSync {
		// Sync NifiResource resource with NiFi side component
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronizing",
			fmt.Sprintf("Synchronizing resource %s", instance.Name))
		status, err := r.SyncNifiResource(instance, parameterContext, parentProcessGroup, clientConfig)
		if err != nil {
			r.Recorder.Event(instance, corev1.EventTypeNormal, "SynchronizingFailed",
				fmt.Sprintf("Synchronizing resource %s failed", instance.Name))
			return RequeueWithError(r.Log, "failed to sync NifiResource "+instance.Name, err)
		}

		instance.Status = *status
		if err := r.updateStatus(ctx, instance, current.Status); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiResource "+instance.Name, err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
			fmt.Sprintf("Synchronized resource %s", instance.Name))
	}

	// Check if the resource is out of sync
	isOutOfSync, err := r.IsOutOfSyncNifiResource(instance, parentProcessGroup, parameterContext, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failed to check sync for NifiResource "+instance.Name, err)
	}

	if isOutOfSync {
		instance.Status.State = v1alpha1.ResourceStateOutOfSync
		if err := r.updateStatus(ctx, instance, current.Status); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiResource "+instance.Name, err)
		}
		return Requeue()
	}

	// Custom specific actions
	requeueCustomPostSync, err := r.CustomPostSyncActions(instance, clientConfig, ctx, current.Status)
	if err != nil {
		return RequeueWithError(r.Log, "failed to run custom actions for NifiResource "+current.Name, err)
	}

	if requeueCustomPostSync {
		return RequeueAfter(interval)
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiResource "+current.Name, err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciled",
		fmt.Sprintf("Reconciling resource %s", instance.Name))

	r.Log.Info("Successfully reconciled NifiResource", zap.String("resourceName", instance.Name))

	return RequeueAfter(interval)
}

func (r *NifiResourceReconciler) CheckNifiResourceExists(resource *v1alpha1.NifiResource, clientConfig *clientconfig.NifiConfig) (bool, error) {
	switch resource.Spec.Type {
	case v1.ResourceProcessGroup:
		return processgroup.ProcessGroupExist(resource, clientConfig)
	default:
		return false, errors.New("Invalid Type for NifiResource")
	}
}

func (r *NifiResourceReconciler) CreateNifiResource(resource *v1alpha1.NifiResource, parameterContext *v1.NifiParameterContext, parentProcessGroup *v1alpha1.NifiResource, clientConfig *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	switch resource.Spec.Type {
	case v1.ResourceProcessGroup:
		return processgroup.CreateProcessGroup(resource, parameterContext, parentProcessGroup, clientConfig)
	default:
		return nil, errors.New("Invalid Type for NifiResource")
	}
}

func (r *NifiResourceReconciler) RemoveNifiResource(resource *v1alpha1.NifiResource, parentProcessGroup *v1alpha1.NifiResource, clientConfig *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	switch resource.Spec.Type {
	case v1.ResourceProcessGroup:
		return processgroup.RemoveProcessGroup(resource, parentProcessGroup, clientConfig)
	default:
		return nil, errors.New("Invalid Type for NifiResource")
	}
}

func (r *NifiResourceReconciler) SyncNifiResource(resource *v1alpha1.NifiResource, parameterContext *v1.NifiParameterContext, parentProcessGroup *v1alpha1.NifiResource, clientConfig *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	switch resource.Spec.Type {
	case v1.ResourceProcessGroup:
		return processgroup.SyncProcessGroup(resource, parameterContext, parentProcessGroup, clientConfig)
	default:
		return nil, errors.New("Invalid Type for NifiResource")
	}
}

func (r *NifiResourceReconciler) IsOutOfSyncNifiResource(resource *v1alpha1.NifiResource, parentProcessGroup *v1alpha1.NifiResource, parameterContext *v1.NifiParameterContext, clientConfig *clientconfig.NifiConfig) (bool, error) {
	switch resource.Spec.Type {
	case v1.ResourceProcessGroup:
		return processgroup.IsOutOfSyncResource(resource, parentProcessGroup, clientConfig, parameterContext)
	default:
		return false, errors.New("Invalid Type for NifiResource")
	}
}

func (r *NifiResourceReconciler) CustomPostSyncActions(resource *v1alpha1.NifiResource, clientConfig *clientconfig.NifiConfig, ctx context.Context, currentStatus v1alpha1.NifiResourceStatus) (bool, error) {
	switch resource.Spec.Type {
	case v1.ResourceProcessGroup:
		if resource.Status.State == v1alpha1.ResourceStateCreated ||
			resource.Status.State == v1alpha1.ResourceStateStarting ||
			resource.Status.State == v1alpha1.ResourceStateInSync ||
			resource.Status.State == v1alpha1.ResourceStateRan {
			// Check if the flow is unscheduled
			isUnscheduled, err := processgroup.IsProcessGroupUnscheduled(resource, clientConfig)
			if err != nil {
				return true, err
			}

			if isUnscheduled {
				r.Log.Debug("Starting process group",
					zap.String("clusterName", resource.Spec.ClusterRef.Name),
					zap.String("processGroup", resource.Name))

				r.Recorder.Event(resource, corev1.EventTypeNormal, "Starting",
					fmt.Sprintf("Starting processGroup %s", resource.Name))

				if err := processgroup.ScheduleProcessGroup(resource, clientConfig); err != nil {
					switch errors.Cause(err).(type) {
					case errorfactory.NifiFlowControllerServiceScheduling, errorfactory.NifiFlowScheduling:
						return true, nil
					default:
						r.Recorder.Event(resource, corev1.EventTypeWarning, "StartingFailed",
							fmt.Sprintf("Starting process group %s failed.", resource.Name))
						return true, err
					}
				}

				if resource.Status.State != v1alpha1.ResourceStateRan {
					resource.Status.State = v1alpha1.ResourceStateRan
					if err := r.updateStatus(ctx, resource, currentStatus); err != nil {
						return true, nil
					}
					r.Log.Info("Successfully ran process group",
						zap.String("clusterName", resource.Spec.ClusterRef.Name),
						zap.String("processGroup", resource.Name))
					r.Recorder.Event(resource, corev1.EventTypeNormal, "Ran",
						fmt.Sprintf("Ran process group %s", resource.Name))
				}
			} else {
				r.Log.Debug("Process group already running, nothing to do",
					zap.String("clusterName", resource.Spec.ClusterRef.Name),
					zap.String("dataflow", resource.Name))
			}
			return false, nil
		} else {
			return false, nil
		}
	default:
		return false, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {

	logCtr, err := GetLogConstructor(mgr, &v1alpha1.NifiResource{})
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiResource{}).
		WithLogConstructor(logCtr).
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

func (r *NifiResourceReconciler) checkFinalizers(ctx context.Context, resource *v1alpha1.NifiResource, parentProcessGroup *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig, patcher client.Patch) (reconcile.Result, error) {
	r.Log.Info("NiFi resource is marked for deletion",
		zap.String("resource", resource.Name))
	var err error
	if util.StringSliceContains(resource.GetFinalizers(), resourceFinalizer) {
		if err = r.finalizeNifiResource(resource, parentProcessGroup, config); err != nil {
			switch errors.Cause(err).(type) {
			// TODO Correct Errors
			case errorfactory.NifiConnectionDropping, errorfactory.NifiFlowDraining:
				return RequeueAfter(util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset))
			default:
				return RequeueWithError(r.Log, "failed to finalize NiFiResource "+resource.Name, err)
			}
		}
		if err = r.removeFinalizer(ctx, resource, patcher); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from resource "+resource.Name, err)
		}
	}

	return Reconciled()
}

func (r *NifiResourceReconciler) removeFinalizer(ctx context.Context, resource *v1alpha1.NifiResource, patcher client.Patch) error {
	r.Log.Info("Removing finalizer for NifiResource",
		zap.String("resource", resource.Name))
	resource.SetFinalizers(util.StringSliceRemove(resource.GetFinalizers(), resourceFinalizer))
	_, err := r.updateAndFetchLatest(ctx, resource, patcher)
	return err
}

func (r *NifiResourceReconciler) finalizeNifiResource(resource *v1alpha1.NifiResource, parentProcessGroup *v1alpha1.NifiResource, config *clientconfig.NifiConfig) error {
	exists, err := r.CheckNifiResourceExists(resource, config)
	if err != nil {
		return err
	}

	if exists {
		r.Recorder.Event(resource, corev1.EventTypeNormal, "Removing",
			fmt.Sprintf("Removing resource %s",
				resource.Name))

		if _, err = r.RemoveNifiResource(resource, parentProcessGroup, config); err != nil {
			return err
		}
		r.Recorder.Event(resource, corev1.EventTypeNormal, "Removed",
			fmt.Sprintf("Removed resource %s",
				resource.Name))

		r.Log.Info("NifiResource deleted",
			zap.String("resource", resource.Name))
	}

	return nil
}

func (r *NifiResourceReconciler) updateStatus(ctx context.Context, resource *v1alpha1.NifiResource, currentStatus v1alpha1.NifiResourceStatus) error {
	if !reflect.DeepEqual(resource.Status, currentStatus) {
		return r.Client.Status().Update(ctx, resource)
	}
	return nil
}

func (r *NifiResourceReconciler) MaintenanceOperation(ctx context.Context, resource *v1alpha1.NifiResource, parentProcessGroup *v1alpha1.NifiResource, clientConfig *clientconfig.NifiConfig) (bool, error) {

	switch resource.Spec.Type {
	case v1.ResourceProcessGroup:
		// Maintenance operation(s) via label
		// Check if maintenance operation is needed
		var maintenanceOpNeeded bool = false
		for labelKey := range resource.Labels {
			if labelKey == nifiutil.StopInputPortLabel || labelKey == nifiutil.StopOutputPortLabel ||
				labelKey == nifiutil.ForceStartLabel || labelKey == nifiutil.ForceStopLabel {
				maintenanceOpNeeded = true
			}
		}

		// Maintenance operation is needed
		if maintenanceOpNeeded {
			r.Recorder.Event(resource, corev1.EventTypeNormal, "MaintenanceOperationInProgress",
				fmt.Sprintf("Syncing process group %s", resource.Name))

			processGroupInformation, err := processgroup.GetProcessGroupInformation(resource, clientConfig)
			if err != nil {
				r.Log.Info("failed to get NifiResource information")
				return true, err
			} else {
				if labelValue, ok := resource.Labels[nifiutil.ForceStopLabel]; ok {
					// Stop process group operation
					if labelValue == "true" {
						err = processgroup.UnscheduleProcessGroup(resource, parentProcessGroup, clientConfig)
						if err != nil {
							r.Log.Info("failed to stop processgroup " + resource.Name)
							return true, err
						}
					}
					return true, nil
				} else if labelValue, ok := resource.Labels[nifiutil.ForceStartLabel]; ok {
					// Start process group operation
					if labelValue == "true" {
						err = processgroup.ScheduleProcessGroup(resource, clientConfig)
						if err != nil {
							r.Log.Info("failed to start process group " + resource.Name)
							return true, err
						}
					}
					return true, nil
				} else {
					if labelValue, ok := resource.Labels[nifiutil.StopInputPortLabel]; ok {
						// Stop input port operation
						for _, port := range processGroupInformation.ProcessGroupFlow.Flow.InputPorts {
							if port.Component.Name == labelValue {
								_, err := inputport.StopPort(port, clientConfig)
								if err != nil {
									r.Log.Info("failed to stop input port " + labelValue)
									return true, err
								}
							}
						}
						return true, nil
					}
					if labelValue, ok := resource.Labels[nifiutil.StopOutputPortLabel]; ok {
						// Stop output port operation
						for _, port := range processGroupInformation.ProcessGroupFlow.Flow.OutputPorts {
							if port.Component.Name == labelValue {
								_, err := outputport.StopPort(port, clientConfig)
								if err != nil {
									r.Log.Info("failed to stop output port " + labelValue)
									return true, err
								}
							}
						}
						return true, nil
					}
				}
			}
		} else {
			return false, nil
		}
	default:
		return false, nil
	}
	return false, nil
}
