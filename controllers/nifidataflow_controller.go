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
	"strconv"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/dataflow"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
)

var dataflowFinalizer = "nifidataflows.nifi.konpyutaika.com/finalizer"

// NifiDataflowReconciler reconciles a NifiDataflow object
type NifiDataflowReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifidataflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifidataflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifidataflows/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiDataflow object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *NifiDataflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("nifidataflow", req.NamespacedName)

	var err error
	interval := util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset)
	// Fetch the NifiDataflow instance
	instance := &v1alpha1.NifiDataflow{}
	if err = r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return Reconciled()
		}
		// Error reading the object - requeue the request.
		return RequeueWithError(r.Log, err.Error(), err)
	}

	// Get the last configuration viewed by the operator.
	o, err := patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	// Create it if not exist.
	if o == nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation", err)
		}
		if err := r.Client.Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiRegistryClient", err)
		}
		o, err = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Check if the cluster reference changed.
	original := &v1alpha1.NifiDataflow{}
	current := instance.DeepCopy()
	json.Unmarshal(o, original)
	if !v1alpha1.ClusterRefsEquals([]v1alpha1.ClusterReference{original.Spec.ClusterRef, instance.Spec.ClusterRef}) {
		instance.Spec.ClusterRef = original.Spec.ClusterRef
	}

	// Get the referenced NifiRegistryClient
	var registryClient *v1alpha1.NifiRegistryClient
	var registryClientNamespace string
	if instance.Spec.RegistryClientRef != nil {
		registryClientNamespace =
			GetRegistryClientRefNamespace(current.Namespace, *current.Spec.RegistryClientRef)

		if registryClient, err = k8sutil.LookupNifiRegistryClient(r.Client,
			current.Spec.RegistryClientRef.Name, registryClientNamespace); err != nil {

			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				r.Log.Info("Dataflow is already gone, there is nothing we can do")
				if err = r.removeFinalizer(ctx, instance); err != nil {
					return RequeueWithError(r.Log, "failed to remove finalizer", err)
				}
				return Reconciled()
			}

			r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceRegistryClientError",
				fmt.Sprintf("Failed to lookup reference registry client : %s in %s",
					current.Spec.RegistryClientRef.Name, registryClientNamespace))

			// the cluster does not exist - should have been caught pre-flight
			return RequeueWithError(r.Log, "failed to lookup referenced registry client", err)
		}
	}

	var parameterContext *v1alpha1.NifiParameterContext
	var parameterContextNamespace string
	if current.Spec.ParameterContextRef != nil {
		parameterContextNamespace =
			GetParameterContextRefNamespace(current.Namespace, *current.Spec.ParameterContextRef)

		if parameterContext, err = k8sutil.LookupNifiParameterContext(r.Client,
			current.Spec.ParameterContextRef.Name, parameterContextNamespace); err != nil {

			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				r.Log.Info("Dataflow context is already gone, there is nothing we can do")
				if err = r.removeFinalizer(ctx, instance); err != nil {
					return RequeueWithError(r.Log, "failed to remove finalizer", err)
				}
				return Reconciled()
			}

			r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceParameterContextError",
				fmt.Sprintf("Failed to lookup reference parameter-context : %s in %s",
					instance.Spec.ClusterRef.Name, parameterContextNamespace))

			// the cluster does not exist - should have been caught pre-flight
			return RequeueWithError(r.Log, "failed to lookup referenced parameter-contest", err)
		}
	}

	// Check if cluster references are the same
	var clusterRefs []v1alpha1.ClusterReference

	registryClusterRef := registryClient.Spec.ClusterRef
	registryClusterRef.Namespace = registryClientNamespace
	clusterRefs = append(clusterRefs, registryClusterRef)

	if parameterContext != nil {
		parameterContextClusterRef := parameterContext.Spec.ClusterRef
		parameterContextClusterRef.Namespace = parameterContextNamespace
		clusterRefs = append(clusterRefs, parameterContextClusterRef)
	}

	currentClusterRef := current.Spec.ClusterRef
	currentClusterRef.Namespace = GetClusterRefNamespace(current.Namespace, current.Spec.ClusterRef)
	clusterRefs = append(clusterRefs, currentClusterRef)

	if !v1alpha1.ClusterRefsEquals(clusterRefs) {

		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to lookup reference cluster : %s in %s",
				instance.Spec.ClusterRef.Name, currentClusterRef.Namespace))

		return RequeueWithError(
			r.Log,
			"failed to lookup referenced cluster, due to inconsistency",
			errors.New("inconsistent cluster references"))
	}

	// Prepare cluster connection configurations
	var clientConfig *clientconfig.NifiConfig
	var clusterConnect clientconfig.ClusterConnect

	// Get the client config manager associated to the cluster ref.
	clusterRef := instance.Spec.ClusterRef
	clusterRef.Namespace = currentClusterRef.Namespace
	configManager := config.GetClientConfigManager(r.Client, clusterRef)

	// Generate the connect object
	if clusterConnect, err = configManager.BuildConnect(); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			r.Log.Info("Cluster is already gone, there is nothing we can do")
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer", err)
			}
			return Reconciled()
		}

		// If the referenced cluster no more exist, just skip the deletion requirement in cluster ref change case.
		if !v1alpha1.ClusterRefsEquals([]v1alpha1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
				return RequeueWithError(r.Log, "could not apply last state to annotation", err)
			}
			if err := r.Client.Update(ctx, current); err != nil {
				return RequeueWithError(r.Log, "failed to update NifiDataflow", err)
			}
			return RequeueAfter(time.Duration(15) * time.Second)
		}
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to lookup reference cluster : %s in %s",
				instance.Spec.ClusterRef.Name, currentClusterRef.Namespace))

		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to lookup referenced cluster", err)
	}

	// Generate the client configuration.
	clientConfig, err = configManager.BuildConfig()
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to create HTTP client for the referenced cluster : %s in %s",
				instance.Spec.ClusterRef.Name, currentClusterRef.Namespace))
		// the cluster is gone, so just remove the finalizer
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return RequeueWithError(r.Log, fmt.Sprintf("failed to remove finalizer from NifiDataflow %s", instance.Name), err)
			}
			return Reconciled()
		}
		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to create HTTP client the for referenced cluster", err)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, clientConfig)
	}

	// Ensure the cluster is ready to receive actions
	if !clusterConnect.IsReady(r.Log) {
		r.Log.Info("Cluster is not ready yet, will wait until it is.")
		r.Recorder.Event(instance, corev1.EventTypeNormal, "ReferenceClusterNotReady",
			fmt.Sprintf("The referenced cluster is not ready yet : %s in %s",
				instance.Spec.ClusterRef.Name, clusterConnect.Id()))

		// the cluster does not exist - should have been caught pre-flight
		return RequeueAfter(interval)
	}

	// Ìn case of the cluster reference changed.
	if !v1alpha1.ClusterRefsEquals([]v1alpha1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
		// Delete the resource on the previous cluster.
		if _, err := dataflow.RemoveDataflow(instance, clientConfig); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "RemoveError",
				fmt.Sprintf("Failed to delete NifiDataflow %s from cluster %s before moving in %s",
					instance.Name, original.Spec.ClusterRef.Name, original.Spec.ClusterRef.Name))
			return RequeueWithError(r.Log, "Failed to delete NifiDataflow before moving", err)
		}
		// Update the last view configuration to the current one.
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation", err)
		}
		if err := r.Client.Update(ctx, current); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDatafllow", err)
		}
		return RequeueAfter(interval)
	}

	if (instance.Spec.SyncNever() && len(instance.Status.State) > 0) ||
		(instance.Spec.SyncOnce() && instance.Status.State == v1alpha1.DataflowStateRan) {
		return Reconciled()
	}

	r.Recorder.Event(instance, corev1.EventTypeWarning, "Reconciling",
		fmt.Sprintf("Reconciling failed dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s}",
			instance.Name, instance.Spec.BucketId,
			instance.Spec.FlowId, strconv.FormatInt(int64(*instance.Spec.FlowVersion), 10)))

	// Check if the dataflow already exist
	existing, err := dataflow.DataflowExist(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing dataflow", err)
	}

	// Create dataflow if it doesn't already exist
	if !existing {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Creating",
			fmt.Sprintf("Creating dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s}",
				instance.Name, instance.Spec.BucketId,
				instance.Spec.FlowId, strconv.FormatInt(int64(*instance.Spec.FlowVersion), 10)))

		processGroupStatus, err := dataflow.CreateDataflow(instance, clientConfig, registryClient)
		if err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "CreationFailed",
				fmt.Sprintf("Creation failed dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s}",
					instance.Name, instance.Spec.BucketId,
					instance.Spec.FlowId, strconv.FormatInt(int64(*instance.Spec.FlowVersion), 10)))
			return RequeueWithError(r.Log, "failure creating dataflow", err)
		}

		// Set dataflow status
		instance.Status = *processGroupStatus
		instance.Status.State = v1alpha1.DataflowStateCreated

		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDataflow status", err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Created",
			fmt.Sprintf("Created dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s}",
				instance.Name, instance.Spec.BucketId,
				instance.Spec.FlowId, strconv.FormatInt(int64(*instance.Spec.FlowVersion), 10)))

		existing = true
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), dataflowFinalizer) {
		r.Log.Info("Adding Finalizer for NifiDataflow")
		instance.SetFinalizers(append(instance.GetFinalizers(), dataflowFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiDataflow", err)
	}

	if instance.Spec.SyncNever() {
		return Reconciled()
	}

	// In case where the flow is not sync
	if instance.Status.State == v1alpha1.DataflowStateOutOfSync {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronizing",
			fmt.Sprintf("Syncing dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s}",
				instance.Name, instance.Spec.BucketId,
				instance.Spec.FlowId, strconv.FormatInt(int64(*instance.Spec.FlowVersion), 10)))

		status, err := dataflow.SyncDataflow(instance, clientConfig, registryClient, parameterContext)
		if status != nil {
			instance.Status = *status
			if err := r.Client.Status().Update(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to update NifiDataflow status", err)
			}
		}
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NifiConnectionDropping,
				errorfactory.NifiFlowUpdateRequestRunning,
				errorfactory.NifiFlowDraining,
				errorfactory.NifiFlowControllerServiceScheduling,
				errorfactory.NifiFlowScheduling, errorfactory.NifiFlowSyncing:
				return reconcile.Result{
					RequeueAfter: interval / 3,
				}, nil
			default:
				r.Recorder.Event(instance, corev1.EventTypeWarning, "SynchronizingFailed",
					fmt.Sprintf("Syncing dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s} failed",
						instance.Name, instance.Spec.BucketId,
						instance.Spec.FlowId, strconv.FormatInt(int64(*instance.Spec.FlowVersion), 10)))
				return RequeueWithError(r.Log, "failed to sync NiFiDataflow", err)
			}
		}

		instance.Status.State = v1alpha1.DataflowStateInSync
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDataflow status", err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
			fmt.Sprintf("Synchronized dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s}",
				instance.Name, instance.Spec.BucketId,
				instance.Spec.FlowId, strconv.FormatInt(int64(*instance.Spec.FlowVersion), 10)))
	}

	// Check if the flow is out of sync
	isOutOfSink, err := dataflow.IsOutOfSyncDataflow(instance, clientConfig, registryClient, parameterContext)
	if err != nil {
		return RequeueWithError(r.Log, "failed to check NifiDataflow sync", err)
	}

	if isOutOfSink {
		instance.Status.State = v1alpha1.DataflowStateOutOfSync
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDataflow status", err)
		}
		return Requeue()
	}

	// Schedule the flow
	if instance.Status.State == v1alpha1.DataflowStateCreated ||
		instance.Status.State == v1alpha1.DataflowStateStarting ||
		instance.Status.State == v1alpha1.DataflowStateInSync ||
		(!instance.Spec.SyncOnce() && instance.Status.State == v1alpha1.DataflowStateRan) {

		instance.Status.State = v1alpha1.DataflowStateStarting
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDataflow status", err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Starting",
			fmt.Sprintf("Starting dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s}",
				instance.Name, instance.Spec.BucketId,
				instance.Spec.FlowId, strconv.FormatInt(int64(*instance.Spec.FlowVersion), 10)))

		if err := dataflow.ScheduleDataflow(instance, clientConfig); err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NifiFlowControllerServiceScheduling, errorfactory.NifiFlowScheduling:
				return RequeueAfter(interval / 3)
			default:
				r.Recorder.Event(instance, corev1.EventTypeWarning, "StartingFailed",
					fmt.Sprintf("Starting dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s} failed.",
						instance.Name, instance.Spec.BucketId,
						instance.Spec.FlowId, strconv.FormatInt(int64(*instance.Spec.FlowVersion), 10)))
				return RequeueWithError(r.Log, "failed to run NifiDataflow", err)
			}
		}

		instance.Status.State = v1alpha1.DataflowStateRan
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDataflow status", err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Ran",
			fmt.Sprintf("Ran dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s}",
				instance.Name, instance.Spec.BucketId,
				instance.Spec.FlowId, strconv.FormatInt(int64(*instance.Spec.FlowVersion), 10)))
	}

	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiCluster label on dataflow", err)
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiDataflow", err)
	}

	r.Log.Info("Ensured Dataflow")

	r.Recorder.Event(instance, corev1.EventTypeWarning, "Reconciled",
		fmt.Sprintf("Success fully ensured dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s}",
			instance.Name, instance.Spec.BucketId,
			instance.Spec.FlowId, strconv.FormatInt(int64(*instance.Spec.FlowVersion), 10)))

	if instance.Spec.SyncOnce() {
		return Reconciled()
	}

	return RequeueAfter(interval / 3)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiDataflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiDataflow{}).
		Complete(r)
}

func (r *NifiDataflowReconciler) ensureClusterLabel(ctx context.Context, cluster clientconfig.ClusterConnect,
	flow *v1alpha1.NifiDataflow) (*v1alpha1.NifiDataflow, error) {

	labels := ApplyClusterReferenceLabel(cluster, flow.GetLabels())
	if !reflect.DeepEqual(labels, flow.GetLabels()) {
		flow.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, flow)
	}
	return flow, nil
}

func (r *NifiDataflowReconciler) updateAndFetchLatest(ctx context.Context,
	flow *v1alpha1.NifiDataflow) (*v1alpha1.NifiDataflow, error) {

	typeMeta := flow.TypeMeta
	err := r.Client.Update(ctx, flow)
	if err != nil {
		return nil, err
	}
	flow.TypeMeta = typeMeta
	return flow, nil
}

func (r *NifiDataflowReconciler) checkFinalizers(ctx context.Context, flow *v1alpha1.NifiDataflow,
	config *clientconfig.NifiConfig) (reconcile.Result, error) {
	r.Log.Info(fmt.Sprintf("NiFi dataflow %s is marked for deletion", flow.Name))
	var err error
	if util.StringSliceContains(flow.GetFinalizers(), dataflowFinalizer) {
		if err = r.finalizeNifiDataflow(flow, config); err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NifiConnectionDropping, errorfactory.NifiFlowDraining:
				return RequeueAfter(util.GetRequeueInterval(r.RequeueInterval/3, r.RequeueOffset))
			default:
				return RequeueWithError(r.Log, "failed to finalize NiFiDataflow", err)
			}
		}
		if err = r.removeFinalizer(ctx, flow); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from dataflow", err)
		}
	}

	return Reconciled()
}

func (r *NifiDataflowReconciler) removeFinalizer(ctx context.Context, flow *v1alpha1.NifiDataflow) error {
	r.Log.V(5).Info(fmt.Sprintf("Removing finalizer for NifiDataflow %s", flow.Name))
	flow.SetFinalizers(util.StringSliceRemove(flow.GetFinalizers(), dataflowFinalizer))
	_, err := r.updateAndFetchLatest(ctx, flow)
	return err
}

func (r *NifiDataflowReconciler) finalizeNifiDataflow(flow *v1alpha1.NifiDataflow, config *clientconfig.NifiConfig) error {

	exists, err := dataflow.DataflowExist(flow, config)
	if err != nil {
		return err
	}

	if exists {
		r.Recorder.Event(flow, corev1.EventTypeNormal, "Removing",
			fmt.Sprintf("Removing dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s}",
				flow.Name, flow.Spec.BucketId,
				flow.Spec.FlowId, strconv.FormatInt(int64(*flow.Spec.FlowVersion), 10)))

		if _, err = dataflow.RemoveDataflow(flow, config); err != nil {
			return err
		}
		r.Recorder.Event(flow, corev1.EventTypeNormal, "Removed",
			fmt.Sprintf("Removed dataflow %s based on flow {bucketId : %s, flowId: %s, version: %s}",
				flow.Name, flow.Spec.BucketId,
				flow.Spec.FlowId, strconv.FormatInt(int64(*flow.Spec.FlowVersion), 10)))

		r.Log.Info("Dataflow deleted")
	}

	return nil
}
