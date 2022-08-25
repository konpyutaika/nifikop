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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/erdrix/nigoapi/pkg/nifi"
	"github.com/go-logr/logr"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/connection"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/dataflow"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var connectionFinalizer string = fmt.Sprintf("nificonnections.%s/stop-input", v1alpha1.GroupVersion.Group)
var lastAppliedClusterAnnotation string = fmt.Sprintf("%s/last-applied-nificluster", v1alpha1.GroupVersion.Group)

// NifiConnectionReconciler reconciles a NifiConnection object
type NifiConnectionReconciler struct {
	client.Client
	Log             logr.Logger
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
	_ = r.Log.WithValues("nificonnection", req.NamespacedName)
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
			return RequeueWithError(r.Log, "could not apply last state to annotation", err)
		}

		if err := r.Client.Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection", err)
		}
		o, err = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Get the last NiFiCluster viewed by the operator.
	cr, err := k8sutil.GetAnnotation(lastAppliedClusterAnnotation, instance)
	// Create it if not exist.
	if cr == nil {
		jsonResource, err := json.Marshal(v1alpha1.ClusterReference{})
		if err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation", err)
		}

		if err := k8sutil.SetAnnotation(lastAppliedClusterAnnotation, instance, jsonResource); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation", err)
		}

		if err := r.Client.Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection", err)
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
	if !current.Spec.Configuration.IsValid() {
		r.Recorder.Event(current, corev1.EventTypeWarning, "ConfigurationInvalid",
			fmt.Sprintf("Failed to validate the connection configuration"))
		return RequeueWithError(r.Log, "failed to validate connection configuration", err)
	}

	current.Spec.Source.Namespace = GetComponentRefNamespace(current.Namespace, current.Spec.Source)
	if !current.Spec.Source.IsValid() {
		r.Recorder.Event(current, corev1.EventTypeWarning, "SourceInvalid",
			fmt.Sprintf("Failed to validate the source component : %s in %s of type %s",
				current.Spec.Source.Name, current.Spec.Source.Namespace, current.Spec.Source.Type))
		return RequeueWithError(r.Log, "failed to validate source component", err)
	}

	current.Spec.Destination.Namespace = GetComponentRefNamespace(current.Namespace, current.Spec.Destination)
	if !current.Spec.Destination.IsValid() {
		r.Recorder.Event(current, corev1.EventTypeWarning, "DestinationInvalid",
			fmt.Sprintf("Failed to validate the destination component : %s in %s of type %s",
				current.Spec.Destination.Name, current.Spec.Destination.Namespace, current.Spec.Destination.Type))
		return RequeueWithError(r.Log, "failed to validate destination component", err)
	}

	// LookUp component
	// Source lookup
	sourceComponent := &v1alpha1.ComponentInformation{}
	if current.Spec.Source.Type == v1alpha1.ComponentDataflow {
		sourceComponent, err = r.GetDataflowComponentInformation(current.Spec.Source, true)
	}

	if err != nil {
		r.Recorder.Event(current, corev1.EventTypeWarning, "SourceNotFound",
			fmt.Sprintf("Failed to retrieve source component information : %s in %s of type %s",
				current.Spec.Source.Name, current.Spec.Source.Namespace, current.Spec.Source.Type))
		return RequeueWithError(r.Log, "failed to retrieve source component", err)
	}

	// Destination lookup
	destinationComponent := &v1alpha1.ComponentInformation{}
	if current.Spec.Source.Type == v1alpha1.ComponentDataflow {
		destinationComponent, err = r.GetDataflowComponentInformation(current.Spec.Destination, false)
	}

	if err != nil {
		r.Recorder.Event(current, corev1.EventTypeWarning, "DestinationNotFound",
			fmt.Sprintf("Failed to retrieve destination component information : %s in %s of type %s",
				current.Spec.Destination.Name, current.Spec.Destination.Namespace, current.Spec.Destination.Type))
		return RequeueWithError(r.Log, "failed to retrieve destination component", err)
	}

	// Verification connection feasible
	var clusterRefs []v1alpha1.ClusterReference
	clusterRefs = append(clusterRefs, sourceComponent.ClusterRef, destinationComponent.ClusterRef)
	if !v1alpha1.ClusterRefsEquals(clusterRefs) {
		r.Recorder.Event(current, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to determine the cluster of the connection between %s in %s of type %s and %s in %s of type %s",
				current.Spec.Source.Name, current.Spec.Source.Namespace, current.Spec.Source.Type,
				current.Spec.Destination.Name, current.Spec.Destination.Namespace, current.Spec.Destination.Type))
		return RequeueWithError(r.Log, "failed to determine the cluster of the connection", err)
	}

	if sourceComponent.ParentGroupId != destinationComponent.ParentGroupId {
		r.Recorder.Event(current, corev1.EventTypeWarning, "ParentGroupIdError",
			fmt.Sprintf("Failed to match parent group id from %s in %s of type %s to %s in %s of type %s",
				current.Spec.Source.Name, current.Spec.Source.Namespace, current.Spec.Source.Type,
				current.Spec.Destination.Name, current.Spec.Destination.Namespace, current.Spec.Destination.Type))
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
		r.Recorder.Event(current, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to create HTTP client for the referenced cluster : %s in %s",
				clusterRef.Name, clusterRef.Namespace))
		// the cluster is gone, so just remove the finalizer
		if k8sutil.IsMarkedForDeletion(current.ObjectMeta) {
			if err = r.removeFinalizer(ctx, current); err != nil {
				return RequeueWithError(r.Log, fmt.Sprintf("failed to remove finalizer from NifiConnection %s", current.Name), err)
			}
			return Reconciled()
		}
		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to create HTTP client the for referenced cluster", err)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(current.ObjectMeta) {
		return r.checkFinalizers(ctx, current, clientConfig)
	}

	// Ensure the cluster is ready to receive actions
	if !clusterConnect.IsReady(r.Log) {
		r.Log.Info("Cluster is not ready yet, will wait until it is.")
		r.Recorder.Event(current, corev1.EventTypeNormal, "ReferenceClusterNotReady",
			fmt.Sprintf("The referenced cluster is not ready yet : %s in %s",
				clusterRef.Name, clusterConnect.Id()))

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

	r.Recorder.Event(current, corev1.EventTypeNormal, "Reconciling",
		fmt.Sprintf("Reconciling connection %s", current.Name))

	// Check if the connection already exist
	existing, err := connection.ConnectionExist(current, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing connection", err)
	}

	if !existing {
		connectionStatus, err := connection.CreateConnection(sourceComponent, destinationComponent, &current.Spec.Configuration, current.Name, clientConfig)
		if err != nil {
			r.Recorder.Event(current, corev1.EventTypeWarning, "CreationFailed",
				fmt.Sprintf("Creation failed connection %s between %s in %s of type %s and %s in %s of type %s",
					current.Name,
					current.Spec.Source.Name, current.Spec.Source.Namespace, current.Spec.Source.Type,
					current.Spec.Destination.Name, current.Spec.Destination.Namespace, current.Spec.Destination.Type))
			return RequeueWithError(r.Log, "failure creating connection", err)
		}

		// Set connection status
		current.Status = *connectionStatus

		if err := r.Client.Status().Update(ctx, current); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection status", err)
		}

		r.Recorder.Event(current, corev1.EventTypeNormal, "Created",
			fmt.Sprintf("Created connection %s between %s in %s of type %s and %s in %s of type %s",
				current.Name,
				current.Spec.Source.Name, current.Spec.Source.Namespace, current.Spec.Source.Type,
				current.Spec.Destination.Name, current.Spec.Destination.Namespace, current.Spec.Destination.Type))
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(current.GetFinalizers(), connectionFinalizer) {
		r.Log.Info("Adding Finalizer for NifiConnection")
		current.SetFinalizers(append(current.GetFinalizers(), connectionFinalizer))
	}

	// Push any changes
	if current, err = r.updateAndFetchLatest(ctx, current); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiConnection", err)
	}

	if instance.Status.State == v1alpha1.ConnectionStateOutOfSync {
		status, err := connection.SyncConnection(instance, sourceComponent, destinationComponent, clientConfig)
		if status != nil {
			instance.Status = *status
			if err := r.Client.Status().Update(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to update NifiConnection status", err)
			}
		}
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NifiConnectionSyncing:
				return reconcile.Result{
					RequeueAfter: interval / 3,
				}, nil
			default:
				r.Recorder.Event(instance, corev1.EventTypeWarning, "SynchronizingFailed",
					fmt.Sprintf("Syncing connection %s between %s in %s of type %s and %s in %s of type %s",
						current.Name,
						current.Spec.Source.Name, current.Spec.Source.Namespace, current.Spec.Source.Type,
						current.Spec.Destination.Name, current.Spec.Destination.Namespace, current.Spec.Destination.Type))
				return RequeueWithError(r.Log, "failed to sync NifiConnection", err)
			}
		}
	}

	// Check if the connection is out of sync
	isOutOfSink, err := connection.IsOutOfSyncConnection(instance, sourceComponent, destinationComponent, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failed to check NifiConnection sync", err)
	}

	if isOutOfSink {
		instance.Status.State = v1alpha1.ConnectionStateOutOfSync
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiConnection status", err)
		}
		return Requeue()
	}

	// // Ensure NifiConnection label
	// if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance); err != nil {
	// 	return RequeueWithError(r.Log, "failed to ensure NifiConnection label on connection", err)
	// }

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiConnection", err)
	}

	// r.Log.Info("Ensured Connection")

	r.Recorder.Event(current, corev1.EventTypeNormal, "Reconciled",
		fmt.Sprintf("Success fully reconciled connection %s", current.Name))

	r.Log.Info("Ensured Connection")

	return RequeueAfter(interval)
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
	r.Log.V(5).Info(fmt.Sprintf("Removing finalizer for NifiConnection %s", connection.Name))
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

		targetPort := &nifi.PortEntity{}
		foundTarget := false
		for _, port := range ports {
			if port.Component.Name == c.SubName {
				targetPort = &port
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