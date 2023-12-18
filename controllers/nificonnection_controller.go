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
	"github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/connection"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/dataflow"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

var connectionFinalizer string = fmt.Sprintf("nificonnections.%s/finalizer", v1alpha1.GroupVersion.Group)

// NifiConnectionReconciler reconciles a NifiConnection object.
type NifiConnectionReconciler struct {
	client.Client
	Log             zap.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificonnections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificonnections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificonnections/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *NifiConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	interval := util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset)
	var err error

	// Fetch the NifiConnection instance
	instance := &v1alpha1.NifiConnection{}
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
			return RequeueWithError(r.Log, "could not apply last state to annotation for connection "+instance.Name, err)
		}

		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
		}
		o, _ = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Get the last NiFiCluster viewed by the operator.
	cr, _ := k8sutil.GetAnnotation(nifiutil.LastAppliedClusterAnnotation, instance)
	// Create it if not exist.
	if cr == nil {
		jsonResource, err := json.Marshal(v1.ClusterReference{})
		if err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for connection "+instance.Name, err)
		}

		if err := k8sutil.SetAnnotation(nifiutil.LastAppliedClusterAnnotation, instance, jsonResource); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for connection "+instance.Name, err)
		}

		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
		}
		cr, _ = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Check if the source or the destination changed
	original := &v1alpha1.NifiConnection{}
	originalClusterRef := &v1.ClusterReference{}
	current := instance.DeepCopy()
	patchCurrent := client.MergeFromWithOptions(current.DeepCopy(), client.MergeFromWithOptimisticLock{})
	json.Unmarshal(o, original)
	json.Unmarshal(cr, originalClusterRef)

	// Validate component
	if !instance.Spec.Configuration.IsValid() {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ConfigurationInvalid",
			fmt.Sprintf("Failed to validate the connection configuration: %s in %s of type %s",
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type))
		return RequeueWithError(r.Log, "failed to validate the configuration of connection "+instance.Name, err)
	}

	// Retrieve the namespace of the source component
	instance.Spec.Source.Namespace = GetComponentRefNamespace(instance.Namespace, instance.Spec.Source)
	// If the source component is invalid, requeue with error
	if !instance.Spec.Source.IsValid() {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "SourceInvalid",
			fmt.Sprintf("Failed to validate the source component: %s in %s of type %s",
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type))
		return RequeueWithError(r.Log, "failed to validate source component "+instance.Spec.Source.Name, err)
	}

	// Retrieve the namespace of the destination component
	instance.Spec.Destination.Namespace = GetComponentRefNamespace(instance.Namespace, instance.Spec.Destination)
	// If the destination component is invalid, requeue with error
	if !instance.Spec.Destination.IsValid() {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "DestinationInvalid",
			fmt.Sprintf("Failed to validate the destination component: %s in %s of type %s",
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
		return RequeueWithError(r.Log, "failed to validate destination component "+instance.Spec.Destination.Name, err)
	}

	// Check if the 2 components are in the same NifiCluster and retrieve it
	currentClusterRef, err := r.RetrieveNifiClusterRef(instance.Spec.Source, instance.Spec.Destination)
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to determine the cluster of the connection between %s in %s of type %s and %s in %s of type %s",
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
		return RequeueWithError(r.Log, "failed to determine the cluster of the connection "+instance.Name, err)
	}

	// Get the client config manager associated to the cluster ref.
	clusterRef := *originalClusterRef
	// Set the clusterRef to the current one if the original one is empty (= new resource)
	if clusterRef.Name == "" && clusterRef.Namespace == "" {
		clusterRef = *currentClusterRef
	}

	// Ìn case of the cluster reference changed.
	if !v1.ClusterRefsEquals([]v1.ClusterReference{clusterRef, *currentClusterRef}) {
		// Prepare cluster connection configurations
		var clientConfig *clientconfig.NifiConfig
		var clusterConnect clientconfig.ClusterConnect

		// Generate the connect object
		configManager := config.GetClientConfigManager(r.Client, clusterRef)
		if clusterConnect, err = configManager.BuildConnect(); err != nil {
			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				r.Log.Info("Cluster is already gone, there is nothing we can do")
				if err = r.removeFinalizer(ctx, current, patchCurrent); err != nil {
					return RequeueWithError(r.Log, "failed to remove finalizer", err)
				}
				return Reconciled()
			}

			// the cluster does not exist - should have been caught pre-flight
			return RequeueWithError(r.Log, "failed to lookup referenced cluster", err)
		}
		// Generate the client configuration.
		clientConfig, err = configManager.BuildConfig()
		if err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
				fmt.Sprintf("Failed to create HTTP client for the referenced cluster: %s in %s",
					clusterRef.Name, clusterRef.Namespace))

			// the cluster does not exist - should have been caught pre-flight
			return RequeueWithError(r.Log, "failed to create HTTP client the for referenced cluster", err)
		}

		// Ensure the cluster is ready to receive actions
		if !clusterConnect.IsReady(r.Log) {
			r.Log.Debug("Cluster is not ready yet, will wait until it is.",
				zap.String("clusterName", clusterRef.Name),
				zap.String("connection", instance.Name))
			r.Recorder.Event(instance, corev1.EventTypeNormal, "ReferenceClusterNotReady",
				fmt.Sprintf("The referenced cluster is not ready yet for connection %s: %s in %s",
					instance.Name, clusterRef.Name, clusterConnect.Id()))
		}

		// Delete the resource on the previous cluster.
		err := r.DeleteConnection(ctx, clientConfig, original, instance)
		if err != nil {
			switch errors.Cause(err).(type) {
			// If the connection is still deleting, requeue
			case errorfactory.NifiConnectionDeleting:
				r.Recorder.Event(instance, corev1.EventTypeWarning, "Deleting",
					fmt.Sprintf("Deleting the connection %s between %s in %s of type %s and %s in %s of type %s",
						original.Name,
						original.Spec.Source.Name, original.Spec.Source.Namespace, original.Spec.Source.Type,
						original.Spec.Destination.Name, original.Spec.Destination.Namespace, original.Spec.Destination.Type))
				return reconcile.Result{
					RequeueAfter: interval / 3,
				}, nil
			// If error during deletion, requeue with error
			default:
				r.Recorder.Event(instance, corev1.EventTypeWarning, "DeleteError",
					fmt.Sprintf("Failed to delete the connection %s between %s in %s of type %s and %s in %s of type %s",
						original.Name,
						original.Spec.Source.Name, original.Spec.Source.Namespace, original.Spec.Source.Type,
						original.Spec.Destination.Name, original.Spec.Destination.Namespace, original.Spec.Destination.Type))
				return RequeueWithError(r.Log, "failed to delete NifiConnection "+instance.Name, err)
			}
		}

		r.Recorder.Event(instance, corev1.EventTypeWarning, "Deleted",
			fmt.Sprintf("The connection %s between %s in %s of type %s and %s in %s of type %s has been deleted",
				original.Name,
				original.Spec.Source.Name, original.Spec.Source.Namespace, original.Spec.Source.Type,
				original.Spec.Destination.Name, original.Spec.Destination.Namespace, original.Spec.Destination.Type))

		// Update the last view configuration to the current one.
		clusterRefJsonResource, err := json.Marshal(v1.ClusterReference{})
		if err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for connection "+instance.Name, err)
		}
		if err := k8sutil.SetAnnotation(nifiutil.LastAppliedClusterAnnotation, instance, clusterRefJsonResource); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for connection "+instance.Name, err)
		}

		// Update last-applied annotation
		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
		}

		return RequeueAfter(interval)
	}

	// LookUp component
	// Source lookup
	sourceComponent := &v1alpha1.ComponentInformation{}
	if instance.Spec.Source.Type == v1alpha1.ComponentDataflow {
		sourceComponent, err = r.GetDataflowComponentInformation(instance.Spec.Source, true)
	}

	// If the source cannot be found, requeue with error
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "SourceNotFound",
			fmt.Sprintf("Failed to retrieve source component information: %s in %s of type %s",
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type))
		return RequeueWithError(r.Log, "failed to retrieve source component "+instance.Spec.Source.Name, err)
	}

	// Destination lookup
	destinationComponent := &v1alpha1.ComponentInformation{}
	if instance.Spec.Source.Type == v1alpha1.ComponentDataflow {
		destinationComponent, err = r.GetDataflowComponentInformation(instance.Spec.Destination, false)
	}

	// If the destination cannot be found, requeue with error
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "DestinationNotFound",
			fmt.Sprintf("Failed to retrieve destination component information: %s in %s of type %s",
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
		return RequeueWithError(r.Log, "failed to retrieve destination component "+instance.Spec.Destination.Name, err)
	}

	// Check if the 2 components are on the same level in the NiFi canvas
	if sourceComponent.ParentGroupId != destinationComponent.ParentGroupId {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ParentGroupIdError",
			fmt.Sprintf("Failed to match parent group id from %s in %s of type %s to %s in %s of type %s",
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
		return RequeueWithError(r.Log, "failed to match parent group id", err)
	}

	// Prepare cluster connection configurations
	var clientConfig *clientconfig.NifiConfig
	var clusterConnect clientconfig.ClusterConnect

	// Generate the connect object
	configManager := config.GetClientConfigManager(r.Client, clusterRef)
	if clusterConnect, err = configManager.BuildConnect(); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			r.Log.Info("Cluster is already gone, there is nothing we can do")
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer", err)
			}
			return Reconciled()
		}

		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to lookup referenced cluster", err)
	}

	// Generate the client configuration.
	clientConfig, err = configManager.BuildConfig()
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to create HTTP client for the referenced cluster: %s in %s",
				clusterRef.Name, clusterRef.Namespace))
		// the cluster is gone, so just remove the finalizer
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, fmt.Sprintf("failed to remove finalizer from NifiConnection %s", instance.Name), err)
			}
			return Reconciled()
		}
		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to create HTTP client the for referenced cluster", err)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, clientConfig, patchInstance)
	}

	// Ensure the cluster is ready to receive actions
	if !clusterConnect.IsReady(r.Log) {
		r.Log.Debug("Cluster is not ready yet, will wait until it is.",
			zap.String("clusterName", clusterRef.Name),
			zap.String("connection", instance.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, "ReferenceClusterNotReady",
			fmt.Sprintf("The referenced cluster is not ready yet for connection %s: %s in %s",
				instance.Name, clusterRef.Name, clusterConnect.Id()))

		// the cluster does not exist - should have been caught pre-flight
		return RequeueAfter(interval)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciling",
		fmt.Sprintf("Reconciling connection %s between %s in %s of type %s and %s in %s of type %s",
			instance.Name,
			instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
			instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))

	// Check if the connection already exists
	existing, err := connection.ConnectionExist(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing connection named "+instance.Name, err)
	}

	// If the connection does not exist, create it
	if !existing {
		connectionStatus, err := connection.CreateConnection(instance, sourceComponent, destinationComponent, clientConfig)
		if err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "CreationFailed",
				fmt.Sprintf("Creation failed connection %s between %s in %s of type %s and %s in %s of type %s",
					instance.Name,
					instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
					instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
			return RequeueWithError(r.Log, "failure creating connection "+instance.Name, err)
		}

		// Update the last view configuration to the current one.
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for connection "+instance.Name, err)
		}

		// Update the last view configuration to the current one.
		clusterRefJsonResource, err := json.Marshal(clusterRef)
		if err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for connection "+instance.Name, err)
		}

		if err := k8sutil.SetAnnotation(nifiutil.LastAppliedClusterAnnotation, instance, clusterRefJsonResource); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for connection "+instance.Name, err)
		}
		// Update last-applied annotation
		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
		}

		// Set connection status
		if instance.Status.State == v1alpha1.ConnectionStateOutOfSync {
			connectionStatus.State = v1alpha1.ConnectionStateOutOfSync
		} else {
			connectionStatus.State = v1alpha1.ConnectionStateCreated
		}
		instance.Status = *connectionStatus
		if err := r.updateStatus(ctx, instance, current.Status); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiConnection "+instance.Name, err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Created",
			fmt.Sprintf("Created connection %s between %s in %s of type %s and %s in %s of type %s",
				instance.Name,
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), connectionFinalizer) {
		r.Log.Info("Adding Finalizer for NifiConnection")
		instance.SetFinalizers(append(instance.GetFinalizers(), connectionFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiConnection "+current.Name, err)
	}

	// If the connection is out of sync, sync it
	if instance.Status.State == v1alpha1.ConnectionStateOutOfSync {
		status, err := connection.SyncConnectionConfig(instance, sourceComponent, destinationComponent, clientConfig)
		if status != nil {
			instance.Status = *status
			if err := r.updateStatus(ctx, instance, current.Status); err != nil {
				return RequeueWithError(r.Log, "failed to update status for NifiConnection "+instance.Name, err)
			}
		}
		if err != nil {
			switch errors.Cause(err).(type) {
			// If the connection is still syncing, requeue
			case errorfactory.NifiConnectionSyncing:
				r.Log.Debug("Connection syncing",
					zap.String("connection", instance.Name))
				return reconcile.Result{
					RequeueAfter: interval / 3,
				}, nil
			// If the connection needs to be deleted, delete it
			case errorfactory.NifiConnectionDeleting:
				err = r.DeleteConnection(ctx, clientConfig, original, instance)
				if err != nil {
					switch errors.Cause(err).(type) {
					// If the connection is still deleting, requeue
					case errorfactory.NifiConnectionDeleting:
						r.Recorder.Event(instance, corev1.EventTypeWarning, "Deleting",
							fmt.Sprintf("Deleting the connection %s between %s in %s of type %s and %s in %s of type %s",
								original.Name,
								original.Spec.Source.Name, original.Spec.Source.Namespace, original.Spec.Source.Type,
								original.Spec.Destination.Name, original.Spec.Destination.Namespace, original.Spec.Destination.Type))
						return reconcile.Result{
							RequeueAfter: interval / 3,
						}, nil
					// If error during deletion, requeue with error
					default:
						r.Recorder.Event(instance, corev1.EventTypeWarning, "DeleteError",
							fmt.Sprintf("Failed to delete the connection %s between %s in %s of type %s and %s in %s of type %s",
								original.Name,
								original.Spec.Source.Name, original.Spec.Source.Namespace, original.Spec.Source.Type,
								original.Spec.Destination.Name, original.Spec.Destination.Namespace, original.Spec.Destination.Type))
						return RequeueWithError(r.Log, "failed to delete NifiConnection "+instance.Name, err)
					}
					// If the connection has been deleted, requeue
				} else {
					r.Recorder.Event(instance, corev1.EventTypeWarning, "Deleted",
						fmt.Sprintf("The connection %s between %s in %s of type %s and %s in %s of type %s has been deleted",
							original.Name,
							original.Spec.Source.Name, original.Spec.Source.Namespace, original.Spec.Source.Type,
							original.Spec.Destination.Name, original.Spec.Destination.Namespace, original.Spec.Destination.Type))

					return reconcile.Result{
						RequeueAfter: interval / 3,
					}, nil
				}
			// If error during syncing, requeue with error
			default:
				r.Recorder.Event(instance, corev1.EventTypeWarning, "SynchronizingFailed",
					fmt.Sprintf("Syncing connection %s between %s in %s of type %s and %s in %s of type %s",
						instance.Name,
						instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
						instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
				return RequeueWithError(r.Log, "failed to sync NifiConnection "+instance.Name, err)
			}
		}

		// Update the last view configuration to the current one.
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for dataflow "+instance.Name, err)
		}
		// Update last-applied annotation
		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
		}

		// Update the state of the connection to indicate that it is synced
		instance.Status.State = v1alpha1.ConnectionStateInSync
		if err := r.updateStatus(ctx, instance, current.Status); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiConnection "+instance.Name, err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
			fmt.Sprintf("Synchronized connection %s between %s in %s of type %s and %s in %s of type %s",
				instance.Name,
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
	}

	// Check if the connection is out of sync
	isOutOfSink, err := connection.IsOutOfSyncConnection(instance, sourceComponent, destinationComponent, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failed to check sync for NifiConnection "+instance.Name, err)
	}

	// If the connection is out of sync, update the state of the connection to indicate it
	if isOutOfSink {
		instance.Status.State = v1alpha1.ConnectionStateOutOfSync
		if err := r.updateStatus(ctx, instance, current.Status); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiConnection "+instance.Name, err)
		}
		return RequeueAfter(interval / 3)
	}

	// Ensure NifiConnection label
	if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiConnection label on connection", err)
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiConnection", err)
	}

	r.Log.Debug("Ensured Connection",
		zap.String("sourceName", instance.Spec.Source.Name),
		zap.String("sourceNamespace", instance.Spec.Source.Namespace),
		zap.String("sourceType", string(instance.Spec.Source.Type)),
		zap.String("destinationName", instance.Spec.Destination.Name),
		zap.String("destinationNamespace", instance.Spec.Destination.Namespace),
		zap.String("destinationType", string(instance.Spec.Destination.Type)),
		zap.String("connection", instance.Name))

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciled",
		fmt.Sprintf("Success fully reconciled connection %s between %s in %s of type %s and %s in %s of type %s",
			instance.Name,
			instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
			instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))

	return RequeueAfter(interval / 3)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiConnection{}).
		Complete(r)
}

// Set the label specifying the cluster used by the NifiConnection.
func (r *NifiConnectionReconciler) ensureClusterLabel(ctx context.Context, cluster clientconfig.ClusterConnect,
	connection *v1alpha1.NifiConnection, patcher client.Patch) (*v1alpha1.NifiConnection, error) {
	labels := ApplyClusterReferenceLabel(cluster, connection.GetLabels())
	if !reflect.DeepEqual(labels, connection.GetLabels()) {
		connection.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, connection, patcher)
	}
	return connection, nil
}

// Update the NifiConnection resource and return the latest version of it.
func (r *NifiConnectionReconciler) updateAndFetchLatest(ctx context.Context,
	connection *v1alpha1.NifiConnection, patcher client.Patch) (*v1alpha1.NifiConnection, error) {
	typeMeta := connection.TypeMeta
	err := r.Client.Patch(ctx, connection, patcher)
	if err != nil {
		return nil, err
	}
	connection.TypeMeta = typeMeta
	return connection, nil
}

// Check if the finalizer is present on the NifiConnection resource.
func (r *NifiConnectionReconciler) checkFinalizers(
	ctx context.Context,
	connection *v1alpha1.NifiConnection,
	config *clientconfig.NifiConfig, patcher client.Patch) (reconcile.Result, error) {
	r.Log.Info(fmt.Sprintf("NiFi connection %s is marked for deletion", connection.Name))
	var err error
	if util.StringSliceContains(connection.GetFinalizers(), connectionFinalizer) {
		if err = r.finalizeNifiConnection(ctx, connection, config); err != nil {
			return RequeueWithError(r.Log, "failed to finalize connection", err)
		}
		if err = r.removeFinalizer(ctx, connection, patcher); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from connection", err)
		}
	}
	return Reconciled()
}

// Remove the finalizer on the NifiConnection resource.
func (r *NifiConnectionReconciler) removeFinalizer(ctx context.Context, connection *v1alpha1.NifiConnection, patcher client.Patch) error {
	r.Log.Info("Removing finalizer for NifiConnection",
		zap.String("connection", connection.Name))
	connection.SetFinalizers(util.StringSliceRemove(connection.GetFinalizers(), connectionFinalizer))
	_, err := r.updateAndFetchLatest(ctx, connection, patcher)
	return err
}

// Delete the connection to finalize the NifiConnection.
func (r *NifiConnectionReconciler) finalizeNifiConnection(
	ctx context.Context,
	instance *v1alpha1.NifiConnection,
	config *clientconfig.NifiConfig) error {
	r.Log.Debug("Finalize the NifiConnection",
		zap.String("connection", instance.Name))

	exists, err := connection.ConnectionExist(instance, config)
	if err != nil {
		return err
	}

	// Check if the connection still exists in NiFi
	if exists {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Removing",
			fmt.Sprintf("Removing connection %s between %s in %s of type %s and %s in %s of type %s",
				instance.Name,
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))

		// Delete the connection
		if err := r.DeleteConnection(ctx, config, instance, instance); err != nil {
			return err
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Removed",
			fmt.Sprintf("Removed connection %s between %s in %s of type %s and %s in %s of type %s",
				instance.Name,
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))

		r.Log.Info("Connection deleted",
			zap.String("connection", instance.Name))
	}

	return nil
}

// Delete the connection.
func (r *NifiConnectionReconciler) DeleteConnection(ctx context.Context, clientConfig *clientconfig.NifiConfig,
	original *v1alpha1.NifiConnection, instance *v1alpha1.NifiConnection) error {
	r.Log.Debug("Delete the connection",
		zap.String("name", instance.Name),
		zap.String("sourceName", original.Spec.Source.Name),
		zap.String("sourceNamespace", original.Spec.Source.Namespace),
		zap.String("sourceType", string(original.Spec.Source.Type)),
		zap.String("destinationName", original.Spec.Destination.Name),
		zap.String("destinationNamespace", original.Spec.Destination.Namespace),
		zap.String("destinationType", string(original.Spec.Destination.Type)))

	// Check if the source component is a NifiDataflow
	if original.Spec.Source.Type == v1alpha1.ComponentDataflow {
		// Retrieve NifiDataflow information
		sourceInstance, err := k8sutil.LookupNifiDataflow(r.Client, original.Spec.Source.Name, original.Spec.Source.Namespace)
		if err != nil {
			return err
		}

		// Check is the NifiDataflow's update strategy is on drain
		if sourceInstance.Spec.UpdateStrategy == v1.DrainStrategy {
			// Check if the dataflow is empty
			isEmpty, err := dataflow.IsDataflowEmpty(sourceInstance, clientConfig)
			if err != nil {
				return err
			}

			// If the dataflow is empty, stop the output-port of the dataflow
			if isEmpty {
				if err := r.StopDataflowComponent(ctx, original.Spec.Source, true); err != nil {
					return err
				}
			}
		}
	}

	// Check if the destination component is a NifiDataflow
	if original.Spec.Destination.Type == v1alpha1.ComponentDataflow {
		// Retrieve NifiDataflow information
		destinationInstance, err := k8sutil.LookupNifiDataflow(r.Client, original.Spec.Destination.Name, original.Spec.Destination.Namespace)
		if err != nil {
			return err
		}

		// If the NifiDataflow's update strategy is on drop and the NifiConnection's too, stop the input-port of the dataflow
		if destinationInstance.Spec.UpdateStrategy == v1.DropStrategy && instance.Spec.UpdateStrategy == v1.DropStrategy {
			if err := r.StopDataflowComponent(ctx, original.Spec.Destination, false); err != nil {
				return err
			}
		}

		// Retrieve the connection information
		connectionEntity, err := connection.GetConnectionInformation(instance, clientConfig)
		if err != nil {
			return err
		}
		if connectionEntity == nil {
			return nil
		}

		// If the source is stopped, the connection is not empty and the connections's update strategy is on drain:
		// force the dataflow to stay started
		if !connectionEntity.Component.Source.Running &&
			connectionEntity.Status.AggregateSnapshot.FlowFilesQueued != 0 &&
			instance.Spec.UpdateStrategy == v1.DrainStrategy {
			if err := r.ForceStartDataflowComponent(ctx, original.Spec.Destination); err != nil {
				return err
			}
			// If the source is stopped, the destination is running and the connection is empty:
			// unforce the dataflow to stay started and stop the input-port of the dataflow
		} else if !connectionEntity.Component.Source.Running && connectionEntity.Component.Destination.Running &&
			connectionEntity.Status.AggregateSnapshot.FlowFilesQueued == 0 {
			if err := r.UnForceStartDataflowComponent(ctx, original.Spec.Destination); err != nil {
				return err
			}
			if err := r.StopDataflowComponent(ctx, original.Spec.Destination, false); err != nil {
				return err
			}
			// If the source is stopped, the destination is stopped, the connection is not empty and the destination's update strategy is on drop:
			// empty the connection
		} else if !connectionEntity.Component.Source.Running && !connectionEntity.Component.Destination.Running &&
			connectionEntity.Status.AggregateSnapshot.FlowFilesQueued != 0 && destinationInstance.Spec.UpdateStrategy == v1.DropStrategy &&
			instance.Spec.UpdateStrategy == v1.DropStrategy {
			if err := connection.DropConnectionFlowFiles(instance, clientConfig); err != nil {
				return err
			}
			// If the source is stopped, the destination is stopped and the connection is empty:
			// delete the connection, unstop the output-port of the source and unstop the input-port of th destination
		} else if !connectionEntity.Component.Source.Running && !connectionEntity.Component.Destination.Running &&
			connectionEntity.Status.AggregateSnapshot.FlowFilesQueued == 0 {
			if err := connection.DeleteConnection(instance, clientConfig); err != nil {
				return err
			}

			// Check if the source component is a NifiDataflow
			if original.Spec.Source.Type == v1alpha1.ComponentDataflow {
				if err := r.UnStopDataflowComponent(ctx, original.Spec.Source, true); err != nil {
					return err
				}
			}

			if err := r.UnStopDataflowComponent(ctx, original.Spec.Destination, false); err != nil {
				return err
			}
			return nil
		}
	}
	return errorfactory.NifiConnectionDeleting{}
}

// Retrieve the clusterRef based on the source and the destination of the connection.
func (r *NifiConnectionReconciler) RetrieveNifiClusterRef(src v1alpha1.ComponentReference, dst v1alpha1.ComponentReference) (*v1.ClusterReference, error) {
	r.Log.Debug("Retrieve the cluster reference from the source and the destination",
		zap.String("sourceName", src.Name),
		zap.String("sourceNamespace", src.Namespace),
		zap.String("sourceType", string(src.Type)),
		zap.String("destinationName", dst.Name),
		zap.String("destinationNamespace", dst.Namespace),
		zap.String("destinationType", string(dst.Type)))

	var srcClusterRef = v1.ClusterReference{}
	// Retrieve the source clusterRef from a NifiDataflow resource
	if src.Type == v1alpha1.ComponentDataflow {
		srcDataflow, err := k8sutil.LookupNifiDataflow(r.Client, src.Name, src.Namespace)
		if err != nil {
			return nil, err
		}

		srcClusterRef = srcDataflow.Spec.ClusterRef
	}

	var dstClusterRef = v1.ClusterReference{}
	// Retrieve the destination clusterRef from a NifiDataflow resource
	if dst.Type == v1alpha1.ComponentDataflow {
		dstDataflow, err := k8sutil.LookupNifiDataflow(r.Client, dst.Name, dst.Namespace)
		if err != nil {
			return nil, err
		}

		dstClusterRef = dstDataflow.Spec.ClusterRef
	}

	// Check that the source and the destination reference the same cluster
	if !v1.ClusterRefsEquals([]v1.ClusterReference{srcClusterRef, dstClusterRef}) {
		return nil, errors.New(fmt.Sprintf("Source cluster %s in %s is different from Destination cluster %s in %s",
			srcClusterRef.Name, srcClusterRef.Namespace,
			dstClusterRef.Name, dstClusterRef.Namespace))
	}

	return &srcClusterRef, nil
}

// Retrieve port information from a NifiDataflow.
func (r *NifiConnectionReconciler) GetDataflowComponentInformation(c v1alpha1.ComponentReference, isSource bool) (*v1alpha1.ComponentInformation, error) {
	var portType string = "input"
	if isSource {
		portType = "output"
	}
	r.Log.Debug("Retrieve the dataflow port information",
		zap.String("dataflowName", c.Name),
		zap.String("dataflowNamespace", c.Namespace),
		zap.String("portName", c.SubName),
		zap.String("portType", portType))

	instance, err := k8sutil.LookupNifiDataflow(r.Client, c.Name, c.Namespace)
	if err != nil {
		return nil, err
	} else {
		// Prepare cluster connection configurations
		var clientConfig *clientconfig.NifiConfig
		var clusterConnect clientconfig.ClusterConnect

		// Get the client config manager associated to the cluster ref.
		clusterRef := instance.Spec.ClusterRef
		clusterRef.Namespace = GetClusterRefNamespace(instance.Namespace, instance.Spec.ClusterRef)
		configManager := config.GetClientConfigManager(r.Client, clusterRef)

		// Generate the connect object
		if clusterConnect, err = configManager.BuildConnect(); err != nil {
			return nil, err
		}

		// Generate the client configuration.
		clientConfig, err = configManager.BuildConfig()
		if err != nil {
			return nil, err
		}

		// Ensure the cluster is ready to receive actions
		if !clusterConnect.IsReady(r.Log) {
			return nil, errors.New(fmt.Sprintf("Cluster %s in %s not ready for dataflow %s in %s", clusterRef.Name, clusterRef.Namespace, instance.Name, instance.Namespace))
		}

		dataflowInformation, err := dataflow.GetDataflowInformation(instance, clientConfig)
		if err != nil {
			return nil, err
		}

		// Error if the dataflow does not exist
		if dataflowInformation == nil {
			return nil, errors.New(fmt.Sprintf("Dataflow %s in %s does not exist in the cluster", instance.Name, instance.Namespace))
		}

		// Retrieve the ports
		var ports = []nifi.PortEntity{}
		if isSource {
			ports = dataflowInformation.ProcessGroupFlow.Flow.OutputPorts
		} else {
			ports = dataflowInformation.ProcessGroupFlow.Flow.InputPorts
		}

		// Error if no port exists in the dataflow
		if len(ports) == 0 {
			return nil, errors.New(fmt.Sprintf("No port available for Dataflow %s in %s", instance.Name, instance.Namespace))
		}

		// Search the targeted port
		targetPort := nifi.PortEntity{}
		foundTarget := false
		for _, port := range ports {
			if port.Component.Name == c.SubName {
				targetPort = port
				foundTarget = true
			}
		}

		// Error if the targeted port is not found
		if !foundTarget {
			return nil, errors.New(fmt.Sprintf("Port %s not found: %s in %s", c.SubName, instance.Name, instance.Namespace))
		}

		// Return all the information on the targeted port of the dataflow
		information := &v1alpha1.ComponentInformation{
			Id:            targetPort.Id,
			Type:          targetPort.Component.Type_,
			GroupId:       targetPort.Component.ParentGroupId,
			ParentGroupId: dataflowInformation.ProcessGroupFlow.ParentGroupId,
			ClusterRef:    clusterRef,
		}
		return information, nil
	}
}

// Set the maintenance label to force the stop of a port.
func (r *NifiConnectionReconciler) StopDataflowComponent(ctx context.Context, c v1alpha1.ComponentReference, isSource bool) error {
	var portType string = "input"
	if isSource {
		portType = "output"
	}
	r.Log.Debug("Set label to stop the port of the dataflow",
		zap.String("dataflowName", c.Name),
		zap.String("dataflowNamespace", c.Namespace),
		zap.String("portName", c.SubName),
		zap.String("portType", portType))

	// Retrieve K8S Dataflow object
	instance, err := k8sutil.LookupNifiDataflow(r.Client, c.Name, c.Namespace)
	instanceOriginal := instance.DeepCopy()
	if err != nil {
		return err
	} else {
		labels := instance.GetLabels()

		// Check that the label is not already set with a different value
		if !isSource {
			if label, ok := labels[nifiutil.StopInputPortLabel]; ok {
				if label != c.SubName {
					return errors.New(fmt.Sprintf("Label %s is already set on the NifiDataflow %s", nifiutil.StopInputPortLabel, instance.Name))
				}
			} else {
				labels[nifiutil.StopInputPortLabel] = c.SubName
				instance.SetLabels(labels)
				return r.Client.Patch(ctx, instance, client.MergeFromWithOptions(instanceOriginal, client.MergeFromWithOptimisticLock{}))
			}
		} else {
			// Set the label
			if label, ok := labels[nifiutil.StopOutputPortLabel]; ok {
				if label != c.SubName {
					return errors.New(fmt.Sprintf("Label %s is already set on the NifiDataflow %s", nifiutil.StopOutputPortLabel, instance.Name))
				}
			} else {
				labels[nifiutil.StopOutputPortLabel] = c.SubName
				instance.SetLabels(labels)
				return r.Client.Patch(ctx, instance, client.MergeFromWithOptions(instanceOriginal, client.MergeFromWithOptimisticLock{}))
			}
		}
	}
	return nil
}

// Unset the maintenance label to force the stop of a port.
func (r *NifiConnectionReconciler) UnStopDataflowComponent(ctx context.Context, c v1alpha1.ComponentReference, isSource bool) error {
	r.Log.Debug("Unset label to stop the port of the dataflow",
		zap.String("dataflowName", c.Name),
		zap.String("dataflowNamespace", c.Namespace))

	// Retrieve K8S Dataflow object
	instance, err := k8sutil.LookupNifiDataflow(r.Client, c.Name, c.Namespace)
	instanceOriginal := instance.DeepCopy()
	if err != nil {
		return err
	} else {
		// Set the label
		labels := instance.GetLabels()

		if !isSource {
			// If the label is set with the correct value, delete it
			if label, ok := labels[nifiutil.StopInputPortLabel]; ok {
				if label == c.SubName {
					delete(labels, nifiutil.StopInputPortLabel)
				}
			}
		} else {
			// If the label is set with the correct value, delete it
			if label, ok := labels[nifiutil.StopOutputPortLabel]; ok {
				if label == c.SubName {
					delete(labels, nifiutil.StopOutputPortLabel)
				}
			}
		}

		instance.SetLabels(labels)
		return r.Client.Patch(ctx, instance, client.MergeFromWithOptions(instanceOriginal, client.MergeFromWithOptimisticLock{}))
	}
}

// Set the maintenance label to force the start of a dataflow.
func (r *NifiConnectionReconciler) ForceStartDataflowComponent(ctx context.Context, c v1alpha1.ComponentReference) error {
	r.Log.Debug("Set label to force the start of the dataflow",
		zap.String("dataflowName", c.Name),
		zap.String("dataflowNamespace", c.Namespace))

	// Retrieve K8S Dataflow object
	instance, err := k8sutil.LookupNifiDataflow(r.Client, c.Name, c.Namespace)
	instanceOriginal := instance.DeepCopy()
	if err != nil {
		return err
	} else {
		labels := instance.GetLabels()
		// Check that the label is not already set with a different value
		if label, ok := labels[nifiutil.ForceStartLabel]; ok {
			if label != "true" {
				return errors.New(fmt.Sprintf("Label %s is already set on the NifiDataflow %s", nifiutil.StopInputPortLabel, instance.Name))
			}
		} else {
			// Set the label
			labels[nifiutil.ForceStartLabel] = "true"
			instance.SetLabels(labels)
			return r.Client.Patch(ctx, instance, client.MergeFromWithOptions(instanceOriginal, client.MergeFromWithOptimisticLock{}))
		}
	}
	return nil
}

// Unset the maintenance label to force the start of a dataflow.
func (r *NifiConnectionReconciler) UnForceStartDataflowComponent(ctx context.Context, c v1alpha1.ComponentReference) error {
	r.Log.Debug("Unset label to force the start of the dataflow",
		zap.String("dataflowName", c.Name),
		zap.String("dataflowNamespace", c.Namespace))

	// Retrieve K8S Dataflow object
	instance, err := k8sutil.LookupNifiDataflow(r.Client, c.Name, c.Namespace)
	instanceOriginal := instance.DeepCopy()
	if err != nil {
		return err
	} else {
		// Unset the label
		labels := instance.GetLabels()

		delete(labels, nifiutil.ForceStartLabel)

		instance.SetLabels(labels)
		return r.Client.Patch(ctx, instance, client.MergeFromWithOptions(instanceOriginal, client.MergeFromWithOptimisticLock{}))
	}
}

func (r *NifiConnectionReconciler) updateStatus(ctx context.Context, connection *v1alpha1.NifiConnection, currentStatus v1alpha1.NifiConnectionStatus) error {
	if !reflect.DeepEqual(connection.Status, currentStatus) {
		return r.Client.Status().Update(ctx, connection)
	}
	return nil
}
