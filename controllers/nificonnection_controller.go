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

	"emperror.dev/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/connection"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/dataflow"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
	"github.com/konpyutaika/nigoapi/pkg/nifi"
)

var connectionFinalizer string = fmt.Sprintf("nificonnections.%s/stop-input", v1alpha1.GroupVersion.Group)
var lastAppliedClusterAnnotation string = fmt.Sprintf("%s/last-applied-nificluster", v1alpha1.GroupVersion.Group)

// NifiConnectionReconciler reconciles a NifiConnection object
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
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiConnection object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
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

	// Get the last configuration viewed by the operator.
	o, err := patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	// Create it if not exist.
	if o == nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for connection "+instance.Name, err)
		}

		if err := r.Client.Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
		}
		o, err = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Get the last NiFiCluster viewed by the operator.
	cr, err := k8sutil.GetAnnotation(lastAppliedClusterAnnotation, instance)
	// Create it if not exist.
	if cr == nil {
		jsonResource, err := json.Marshal(v1alpha1.ClusterReference{})
		if err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for connection "+instance.Name, err)
		}

		if err := k8sutil.SetAnnotation(lastAppliedClusterAnnotation, instance, jsonResource); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for connection "+instance.Name, err)
		}

		if err := r.Client.Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
		}
		cr, err = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Check if the source or the destination changed
	original := &v1alpha1.NifiConnection{}
	originalClusterRef := &v1alpha1.ClusterReference{}
	current := instance.DeepCopy()
	json.Unmarshal(o, original)
	json.Unmarshal(cr, originalClusterRef)

	// Validate component
	if !instance.Spec.Configuration.IsValid() {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ConfigurationInvalid",
			fmt.Sprintf("Failed to validate the connection configuration"))
		return RequeueWithError(r.Log, "failed to validate the configuration of connection "+instance.Name, err)
	}

	instance.Spec.Source.Namespace = GetComponentRefNamespace(instance.Namespace, instance.Spec.Source)
	if !instance.Spec.Source.IsValid() {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "SourceInvalid",
			fmt.Sprintf("Failed to validate the source component : %s in %s of type %s",
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type))
		return RequeueWithError(r.Log, "failed to validate source component "+instance.Spec.Source.Name, err)
	}

	instance.Spec.Destination.Namespace = GetComponentRefNamespace(instance.Namespace, instance.Spec.Destination)
	if !instance.Spec.Destination.IsValid() {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "DestinationInvalid",
			fmt.Sprintf("Failed to validate the destination component : %s in %s of type %s",
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
		return RequeueWithError(r.Log, "failed to validate destination component "+instance.Spec.Destination.Name, err)
	}

	// LookUp component
	// Source lookup
	sourceComponent := &v1alpha1.ComponentInformation{}
	if instance.Spec.Source.Type == v1alpha1.ComponentDataflow {
		sourceComponent, err = r.GetDataflowComponentInformation(instance.Spec.Source, true)
	}

	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "SourceNotFound",
			fmt.Sprintf("Failed to retrieve source component information : %s in %s of type %s",
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type))
		return RequeueWithError(r.Log, "failed to retrieve source component "+instance.Spec.Source.Name, err)
	}

	// Destination lookup
	destinationComponent := &v1alpha1.ComponentInformation{}
	if instance.Spec.Source.Type == v1alpha1.ComponentDataflow {
		destinationComponent, err = r.GetDataflowComponentInformation(instance.Spec.Destination, false)
	}

	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "DestinationNotFound",
			fmt.Sprintf("Failed to retrieve destination component information : %s in %s of type %s",
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
		return RequeueWithError(r.Log, "failed to retrieve destination component "+instance.Spec.Destination.Name, err)
	}

	// Verification connection feasible
	var clusterRefs []v1alpha1.ClusterReference
	clusterRefs = append(clusterRefs, sourceComponent.ClusterRef, destinationComponent.ClusterRef)
	if !v1alpha1.ClusterRefsEquals(clusterRefs) {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to determine the cluster of the connection between %s in %s of type %s and %s in %s of type %s",
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
		return RequeueWithError(r.Log, "failed to determine the cluster of the connection "+instance.Name, err)
	}

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

	// Get the client config manager associated to the cluster ref.
	currentClusterRef := sourceComponent.ClusterRef
	clusterRef := *originalClusterRef
	// Set the clusterRef to the current one if the original one is empty (= new resource)
	if clusterRef.Name == "" && clusterRef.Namespace == "" {
		clusterRef = currentClusterRef
	}
	configManager := config.GetClientConfigManager(r.Client, clusterRef)

	// Generate the connect object
	if clusterConnect, err = configManager.BuildConnect(); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		// if k8sutil.IsMarkedForDeletion(current.ObjectMeta) {
		// 	r.Log.Info("Cluster is already gone, there is nothing we can do")
		// 	if err = r.removeFinalizer(ctx, current); err != nil {
		// 		return RequeueWithError(r.Log, "failed to remove finalizer", err)
		// 	}
		// 	return Reconciled()
		// }

		// // If the referenced cluster no more exist, just skip the deletion requirement in cluster ref change case.
		// if !v1alpha1.ClusterRefsEquals([]v1alpha1.ClusterReference{current.Spec.ClusterRef, current.Spec.ClusterRef}) {
		// 	if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
		// 		return RequeueWithError(r.Log, "could not apply last state to annotation", err)
		// 	}
		// 	if err := r.Client.Update(ctx, current); err != nil {
		// 		return RequeueWithError(r.Log, "failed to update NifiDataflow", err)
		// 	}
		// 	return RequeueAfter(time.Duration(15) * time.Second)
		// }
		// r.Recorder.Event(current, corev1.EventTypeWarning, "ReferenceClusterError",
		// 	fmt.Sprintf("Failed to lookup reference cluster : %s in %s",
		// 		current.Spec.ClusterRef.Name, currentClusterRef.Namespace))

		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to lookup referenced cluster", err)
	}

	// Generate the client configuration.
	clientConfig, err = configManager.BuildConfig()
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to create HTTP client for the referenced cluster : %s in %s",
				clusterRef.Name, clusterRef.Namespace))
		// the cluster is gone, so just remove the finalizer
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return RequeueWithError(r.Log, fmt.Sprintf("failed to remove finalizer from NifiConnection %s", instance.Name), err)
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
		r.Log.Debug("Cluster is not ready yet, will wait until it is.",
			zap.String("clusterName", clusterRef.Name),
			zap.String("connection", instance.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, "ReferenceClusterNotReady",
			fmt.Sprintf("The referenced cluster is not ready yet for connection %s : %s in %s",
				instance.Name, clusterRef.Name, clusterConnect.Id()))

		// the cluster does not exist - should have been caught pre-flight
		return RequeueAfter(interval)
	}

	// Ìn case of the cluster reference changed.
	if !v1alpha1.ClusterRefsEquals([]v1alpha1.ClusterReference{clusterRef, currentClusterRef}) {
		// // Delete the resource on the previous cluster.
		// if _, err := dataflow.RemoveDataflow(instance, clientConfig); err != nil {
		// 	r.Recorder.Event(instance, corev1.EventTypeWarning, "RemoveError",
		// 		fmt.Sprintf("Failed to delete NifiDataflow %s from cluster %s before moving in %s",
		// 			instance.Name, original.Spec.ClusterRef.Name, original.Spec.ClusterRef.Name))
		// 	return RequeueWithError(r.Log, "Failed to delete NifiDataflow before moving", err)
		// }
		// // Update the last view configuration to the current one.
		// if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
		// 	return RequeueWithError(r.Log, "could not apply last state to annotation", err)
		// }
		// if err := r.Client.Update(ctx, current); err != nil {
		// 	return RequeueWithError(r.Log, "failed to update NifiDatafllow", err)
		// }
		return RequeueAfter(interval)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciling",
		fmt.Sprintf("Reconciling connection %s between %s in %s of type %s and %s in %s of type %s",
			instance.Name,
			instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
			instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))

	// Check if the connection already exist
	existing, err := connection.ConnectionExist(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing connection named "+instance.Name, err)
	}

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
			return RequeueWithError(r.Log, "could not apply last state to annotation for dataflow "+instance.Name, err)
		}
		// Update last-applied annotation
		if err := r.Client.Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
		}

		// Set connection status
		if instance.Status.State == v1alpha1.ConnectionStateConfigOutOfSync {
			connectionStatus.State = v1alpha1.ConnectionStateConfigOutOfSync
		} else {
			connectionStatus.State = v1alpha1.ConnectionStateCreated
		}
		instance.Status = *connectionStatus
		if err := r.Client.Status().Update(ctx, instance); err != nil {
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
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiConnection "+current.Name, err)
	}

	// // Resync connection source
	// if instance.Status.State == v1alpha1.ConnectionStateSourceOutOfSync {
	// 	if original.Spec.Source.Type == v1alpha1.ComponentDataflow {
	// 		if err := r.StopDataflowComponent(ctx, original.Spec.Source, true); err != nil {
	// 			return RequeueWithError(r.Log, "failed to update label of NifiDataflow "+original.Spec.Source.Name, err)
	// 		}
	// 	}

	// 	if original.Spec.Destination.Type == v1alpha1.ComponentDataflow {
	// 		destinationInstance, err := k8sutil.LookupNifiDataflow(r.Client, original.Spec.Destination.Name, original.Spec.Destination.Namespace)
	// 		if err != nil {
	// 			return RequeueWithError(r.Log, "failed to retrieve information of NifiDataflow "+original.Spec.Destination.Name, err)
	// 		}

	// 		if destinationInstance.Spec.UpdateStrategy == v1alpha1.DropStrategy && instance.Spec.UpdateStrategy == v1alpha1.DropStrategy {
	// 			if err := r.StopDataflowComponent(ctx, original.Spec.Destination, false); err != nil {
	// 				return RequeueWithError(r.Log, "failed to update label of NifiDataflow "+original.Spec.Destination.Name, err)
	// 			}
	// 		}

	// 		connectionEntity, err := connection.GetConnectionInformation(instance, clientConfig)
	// 		if err != nil {
	// 			return RequeueWithError(r.Log, "failed to retrieve information of NifiConnection "+instance.Name, err)
	// 		}
	// 		if !connectionEntity.Component.Source.Running && connectionEntity.Component.Destination.Running &&
	// 			connectionEntity.Status.AggregateSnapshot.FlowFilesQueued == 0 {
	// 			if err := r.StopDataflowComponent(ctx, original.Spec.Destination, false); err != nil {
	// 				return RequeueWithError(r.Log, "failed to update label of NifiDataflow "+original.Spec.Destination.Name, err)
	// 			}
	// 		} else if !connectionEntity.Component.Source.Running && !connectionEntity.Component.Destination.Running &&
	// 			connectionEntity.Status.AggregateSnapshot.FlowFilesQueued == 0 {
	// 			if err := connection.DeleteConnection(instance, clientConfig); err != nil {
	// 				return RequeueWithError(r.Log, "failed to delete connection "+instance.Name, err)
	// 			}

	// 			if err := r.UnStopDataflowComponent(ctx, original.Spec.Source, true); err != nil {
	// 				return RequeueWithError(r.Log, "failed to update label of NifiDataflow "+original.Spec.Source.Name, err)
	// 			}

	// 			if err := r.UnStopDataflowComponent(ctx, original.Spec.Destination, false); err != nil {
	// 				return RequeueWithError(r.Log, "failed to update label of NifiDataflow "+original.Spec.Destination.Name, err)
	// 			}

	// 			// Update the last view configuration to the current one (only for the source).
	// 			instance.Spec = original.Spec
	// 			instance.Spec.Source = current.Spec.Source
	// 			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
	// 				return RequeueWithError(r.Log, "could not apply last state to annotation for dataflow "+instance.Name, err)
	// 			}
	// 			instance.Spec = current.Spec

	// 			// Update last-applied annotation with only the new source
	// 			if err := r.Client.Update(ctx, instance); err != nil {
	// 				return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
	// 			}

	// 			// Update status
	// 			instance.Status.State = v1alpha1.ConnectionStateInSync
	// 			if err := r.Client.Status().Update(ctx, instance); err != nil {
	// 				return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
	// 			}

	// 			r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
	// 				fmt.Sprintf("Synchronized connection %s between %s in %s of type %s and %s in %s of type %s",
	// 					instance.Name,
	// 					instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
	// 					instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
	// 		}
	// 	}
	// }

	// // Resync connection destination
	// if instance.Status.State == v1alpha1.ConnectionStateDestinationOutOfSync {
	// 	if original.Spec.Destination.Type == v1alpha1.ComponentDataflow {
	// 		status, err := connection.SyncConnectionDestination(instance, destinationComponent, clientConfig)
	// 		if status != nil {
	// 			instance.Status = *status
	// 			if err := r.Client.Status().Update(ctx, instance); err != nil {
	// 				return RequeueWithError(r.Log, "failed to update status for NifiConnection "+instance.Name, err)
	// 			}
	// 		}
	// 		if err != nil {
	// 			return RequeueWithError(r.Log, "failed to sync NifiConnection "+instance.Name, err)
	// 		}
	// 	}

	// 	// Update the last view configuration to the current one (only for the destination).
	// 	instance.Spec = original.Spec
	// 	instance.Spec.Destination = current.Spec.Destination
	// 	if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
	// 		return RequeueWithError(r.Log, "could not apply last state to annotation for dataflow "+instance.Name, err)
	// 	}
	// 	instance.Spec = current.Spec

	// 	// Update last-applied annotation with only the new destination
	// 	if err := r.Client.Update(ctx, instance); err != nil {
	// 		return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
	// 	}

	// 	// Update status
	// 	instance.Status.State = v1alpha1.ConnectionStateInSync
	// 	if err := r.Client.Status().Update(ctx, instance); err != nil {
	// 		return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
	// 	}

	// 	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
	// 		fmt.Sprintf("Synchronized connection %s between %s in %s of type %s and %s in %s of type %s",
	// 			instance.Name,
	// 			instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
	// 			instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
	// }

	// Resync connection configuration
	if instance.Status.State == v1alpha1.ConnectionStateConfigOutOfSync {
		status, err := connection.SyncConnectionConfig(instance, sourceComponent, destinationComponent, clientConfig)
		if status != nil {
			instance.Status = *status
			if err := r.Client.Status().Update(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to update status for NifiConnection "+instance.Name, err)
			}
		}
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NifiConnectionSyncing:
				return reconcile.Result{
					RequeueAfter: interval / 3,
				}, nil
			case errorfactory.NifiConnectionDeleting:
				err = r.DeleteConnection(ctx, clientConfig, original, instance)
				if err != nil {
					r.Recorder.Event(instance, corev1.EventTypeWarning, "SynchronizingFailed",
						fmt.Sprintf("Deleting connection %s between %s in %s of type %s and %s in %s of type %s",
							original.Name,
							original.Spec.Source.Name, original.Spec.Source.Namespace, original.Spec.Source.Type,
							original.Spec.Destination.Name, original.Spec.Destination.Namespace, original.Spec.Destination.Type))
					return RequeueWithError(r.Log, "failed to delete NifiConnection "+instance.Name, err)
				} else {
					return reconcile.Result{
						RequeueAfter: interval / 3,
					}, nil
				}
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
		if err := r.Client.Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection "+instance.Name, err)
		}

		instance.Status.State = v1alpha1.ConnectionStateInSync
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiConnection "+instance.Name, err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
			fmt.Sprintf("Synchronized connection %s between %s in %s of type %s and %s in %s of type %s",
				instance.Name,
				instance.Spec.Source.Name, instance.Spec.Source.Namespace, instance.Spec.Source.Type,
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace, instance.Spec.Destination.Type))
	}

	// Check if the connection is out of sync
	// // Check if the destination of the connection is out of sync
	// if original.Spec.Source.Name != instance.Spec.Source.Name ||
	// 	original.Spec.Source.SubName != instance.Spec.Source.SubName {
	// 	instance.Status.State = v1alpha1.ConnectionStateSourceOutOfSync
	// 	if err := r.Client.Status().Update(ctx, instance); err != nil {
	// 		return RequeueWithError(r.Log, "failed to update status for NifiConnection "+instance.Name, err)
	// 	}
	// 	return RequeueAfter(interval / 3)
	// }

	// // Check if the destination of the connection is out of sync
	// if original.Spec.Destination.Name != instance.Spec.Destination.Name ||
	// 	original.Spec.Destination.SubName != instance.Spec.Destination.SubName {
	// 	instance.Status.State = v1alpha1.ConnectionStateDestinationOutOfSync
	// 	if err := r.Client.Status().Update(ctx, instance); err != nil {
	// 		return RequeueWithError(r.Log, "failed to update status for NifiConnection "+instance.Name, err)
	// 	}
	// 	return RequeueAfter(interval / 3)
	// }

	// Check if the configuration of the connection is out of sync
	isOutOfSink, err := connection.IsOutOfSyncConnection(instance, sourceComponent, destinationComponent, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failed to check sync for NifiConnection "+instance.Name, err)
	}

	if isOutOfSink {
		instance.Status.State = v1alpha1.ConnectionStateConfigOutOfSync
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiConnection "+instance.Name, err)
		}
		return RequeueAfter(interval / 3)
	}

	// // Ensure NifiConnection label
	// if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance); err != nil {
	// 	return RequeueWithError(r.Log, "failed to ensure NifiConnection label on connection", err)
	// }

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
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

func (r *NifiConnectionReconciler) updateAndFetchLatest(ctx context.Context,
	connection *v1alpha1.NifiConnection) (*v1alpha1.NifiConnection, error) {

	typeMeta := connection.TypeMeta
	err := r.Client.Update(ctx, connection)
	if err != nil {
		return nil, err
	}
	connection.TypeMeta = typeMeta
	return connection, nil
}

func (r *NifiConnectionReconciler) checkFinalizers(
	ctx context.Context,
	connection *v1alpha1.NifiConnection,
	config *clientconfig.NifiConfig) (reconcile.Result, error) {
	r.Log.Info(fmt.Sprintf("NiFi connection %s is marked for deletion", connection.Name))
	var err error
	if util.StringSliceContains(connection.GetFinalizers(), connectionFinalizer) {
		if err = r.finalizeNifiConnection(connection, config); err != nil {
			return RequeueWithError(r.Log, "failed to finalize connection", err)
		}
		if err = r.removeFinalizer(ctx, connection); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from connection", err)
		}
	}
	return Reconciled()
}

func (r *NifiConnectionReconciler) removeFinalizer(ctx context.Context, connection *v1alpha1.NifiConnection) error {
	r.Log.Info("Removing finalizer for NifiConnection",
		zap.String("connection", connection.Name))
	connection.SetFinalizers(util.StringSliceRemove(connection.GetFinalizers(), connectionFinalizer))
	_, err := r.updateAndFetchLatest(ctx, connection)
	return err
}

func (r *NifiConnectionReconciler) finalizeNifiConnection(
	connection *v1alpha1.NifiConnection,
	config *clientconfig.NifiConfig) error {

	// if err := parametercontext.RemoveParameterContext(connection, config); err != nil {
	// 	return err
	// }
	// r.Log.Info("Delete NifiConnection Context")

	return nil
}

func (r *NifiConnectionReconciler) DeleteConnection(ctx context.Context, clientConfig *clientconfig.NifiConfig, original *v1alpha1.NifiConnection, instance *v1alpha1.NifiConnection) error {
	if original.Spec.Source.Type == v1alpha1.ComponentDataflow {
		if err := r.StopDataflowComponent(ctx, original.Spec.Source, true); err != nil {
			return err
		}
	}

	if original.Spec.Destination.Type == v1alpha1.ComponentDataflow {
		destinationInstance, err := k8sutil.LookupNifiDataflow(r.Client, original.Spec.Destination.Name, original.Spec.Destination.Namespace)
		if err != nil {
			return err
		}

		if destinationInstance.Spec.UpdateStrategy == v1alpha1.DropStrategy && instance.Spec.UpdateStrategy == v1alpha1.DropStrategy {
			if err := r.StopDataflowComponent(ctx, original.Spec.Destination, false); err != nil {
				return err
			}
		}

		connectionEntity, err := connection.GetConnectionInformation(instance, clientConfig)
		if err != nil {
			return err
		}
		if !connectionEntity.Component.Source.Running && connectionEntity.Component.Destination.Running &&
			connectionEntity.Status.AggregateSnapshot.FlowFilesQueued == 0 {
			if err := r.StopDataflowComponent(ctx, original.Spec.Destination, false); err != nil {
				return err
			}
		} else if !connectionEntity.Component.Source.Running && !connectionEntity.Component.Destination.Running &&
			connectionEntity.Status.AggregateSnapshot.FlowFilesQueued == 0 {
			if err := connection.DeleteConnection(instance, clientConfig); err != nil {
				return err
			}

			if err := r.UnStopDataflowComponent(ctx, original.Spec.Source, true); err != nil {
				return err
			}

			if err := r.UnStopDataflowComponent(ctx, original.Spec.Destination, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *NifiConnectionReconciler) GetDataflowComponentInformation(c v1alpha1.ComponentReference, isSource bool) (*v1alpha1.ComponentInformation, error) {
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

		var ports = []nifi.PortEntity{}
		if isSource {
			ports = dataflowInformation.ProcessGroupFlow.Flow.OutputPorts
		} else {
			ports = dataflowInformation.ProcessGroupFlow.Flow.InputPorts
		}

		if len(ports) == 0 {
			return nil, errors.New(fmt.Sprintf("No port available for Dataflow %s in %s", instance.Name, instance.Namespace))
		}

		targetPort := nifi.PortEntity{}
		foundTarget := false
		for _, port := range ports {
			if port.Component.Name == c.SubName {
				targetPort = port
				foundTarget = true
			}
		}

		if !foundTarget {
			return nil, errors.New(fmt.Sprintf("Port %s not found : %s in %s", c.SubName, instance.Name, instance.Namespace))
		}

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

func (r *NifiConnectionReconciler) StopDataflowComponent(ctx context.Context, c v1alpha1.ComponentReference, isSource bool) error {
	instance, err := k8sutil.LookupNifiDataflow(r.Client, c.Name, c.Namespace)
	if err != nil {
		return err
	} else {
		labels := instance.GetLabels()

		if !isSource {
			if label, ok := labels[nifiutil.StopInputPortPrefix]; ok {
				if label != c.SubName {
					return errors.New(fmt.Sprintf("Label %s is already set on the NifiDataflow %s", nifiutil.StopInputPortPrefix, instance.Name))
				}
			} else {
				labels[nifiutil.StopInputPortPrefix] = c.SubName
				instance.SetLabels(labels)
				return r.Client.Update(ctx, instance)
			}
		} else {
			if label, ok := labels[nifiutil.StopOutputPortPrefix]; ok {
				if label != c.SubName {
					return errors.New(fmt.Sprintf("Label %s is already set on the NifiDataflow %s", nifiutil.StopOutputPortPrefix, instance.Name))
				}
			} else {
				labels[nifiutil.StopOutputPortPrefix] = c.SubName
				instance.SetLabels(labels)
				return r.Client.Update(ctx, instance)
			}
		}
	}
	return nil
}

func (r *NifiConnectionReconciler) UnStopDataflowComponent(ctx context.Context, c v1alpha1.ComponentReference, isSource bool) error {
	instance, err := k8sutil.LookupNifiDataflow(r.Client, c.Name, c.Namespace)
	if err != nil {
		return err
	} else {
		labels := instance.GetLabels()

		if !isSource {
			delete(labels, nifiutil.StopInputPortPrefix)
		} else {
			delete(labels, nifiutil.StopOutputPortPrefix)
		}

		instance.SetLabels(labels)
		return r.Client.Update(ctx, instance)
	}
}
