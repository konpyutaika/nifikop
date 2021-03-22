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
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/dataflow"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/Orange-OpenSource/nifikop/pkg/k8sutil"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
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

var dataflowFinalizer = "nifidataflows.nifi.orange.com/finalizer"

// NifiDataflowReconciler reconciles a NifiDataflow object
type NifiDataflowReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifidataflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifidataflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifidataflows/finalizers,verbs=update

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

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), dataflowFinalizer) {
		r.Log.Info("Adding Finalizer for NifiDataflow")
		instance.SetFinalizers(append(instance.GetFinalizers(), dataflowFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiDataflow", err)
	}

	// Get the referenced NifiRegistryClient
	var registryClient *v1alpha1.NifiRegistryClient
	var registryClientNamespace string
	if instance.Spec.RegistryClientRef != nil {
		registryClientNamespace =
			GetRegistryClientRefNamespace(instance.Namespace, *instance.Spec.RegistryClientRef)

		if registryClient, err = k8sutil.LookupNifiRegistryClient(r.Client,
			instance.Spec.RegistryClientRef.Name, registryClientNamespace); err != nil {

			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				r.Log.Info("Registry client is already gone, there is nothing we can do")
				if err = r.removeFinalizer(ctx, instance); err != nil {
					return RequeueWithError(r.Log, "failed to remove finalizer", err)
				}
				return Reconciled()
			}

			// the cluster does not exist - should have been caught pre-flight
			return RequeueWithError(r.Log, "failed to lookup referenced registry client", err)
		}
	}

	var parameterContext *v1alpha1.NifiParameterContext
	var parameterContextNamespace string
	if instance.Spec.ParameterContextRef != nil {
		parameterContextNamespace =
			GetParameterContextRefNamespace(instance.Namespace, *instance.Spec.ParameterContextRef)

		if parameterContext, err = k8sutil.LookupNifiParameterContext(r.Client,
			instance.Spec.ParameterContextRef.Name, parameterContextNamespace); err != nil {

			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				r.Log.Info("Parameter context is already gone, there is nothing we can do")
				if err = r.removeFinalizer(ctx, instance); err != nil {
					return RequeueWithError(r.Log, "failed to remove finalizer", err)
				}
				return Reconciled()
			}

			// the cluster does not exist - should have been caught pre-flight
			return RequeueWithError(r.Log, "failed to lookup referenced parameter-contest", err)
		}
	}

	// Check if cluster references are the same
	clusterNamespace := GetClusterRefNamespace(instance.Namespace, instance.Spec.ClusterRef)
	if registryClient != nil &&
		(registryClientNamespace != clusterNamespace ||
			registryClient.Spec.ClusterRef.Name != instance.Spec.ClusterRef.Name ||
			(parameterContext != nil &&
				(parameterContextNamespace != clusterNamespace ||
					parameterContext.Spec.ClusterRef.Name != instance.Spec.ClusterRef.Name))) {

		return RequeueWithError(
			r.Log,
			"failed to lookup referenced cluster, due to inconsistency",
			errors.New("inconsistent cluster references"))
	}

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
		return r.checkFinalizers(ctx, instance, cluster)
	}

	if instance.Spec.GetRunOnce() && instance.Status.State == v1alpha1.DataflowStateRan {
		return Reconciled()
	}

	// Check if the dataflow already exist
	existing, err := dataflow.DataflowExist(r.Client, instance, cluster)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing dataflow", err)
	}

	// Create dataflow if it doesn't already exist
	if !existing {

		processGroupStatus, err := dataflow.CreateDataflow(r.Client, instance, cluster, registryClient)
		if err != nil {
			return RequeueWithError(r.Log, "failure creating dataflow", err)
		}

		// Set dataflow status
		instance.Status = *processGroupStatus
		instance.Status.State = v1alpha1.DataflowStateCreated

		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDataflow status", err)
		}

		existing = true
	}

	// In case where the flow is not sync
	if instance.Status.State == v1alpha1.DataflowStateOutOfSync {
		status, err := dataflow.SyncDataflow(r.Client, instance, cluster, registryClient, parameterContext)
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
					RequeueAfter: time.Duration(5) * time.Second,
				}, nil
			default:
				return RequeueWithError(r.Log, "failed to sync NiFiDataflow", err)
			}
		}

		instance.Status.State = v1alpha1.DataflowStateInSync
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDataflow status", err)
		}
	}

	// Check if the flow is out of sync
	isOutOfSink, err := dataflow.IsOutOfSyncDataflow(r.Client, instance, cluster, registryClient, parameterContext)
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
		(!instance.Spec.GetRunOnce() && instance.Status.State == v1alpha1.DataflowStateRan) {

		instance.Status.State = v1alpha1.DataflowStateStarting
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDataflow status", err)
		}

		if err := dataflow.ScheduleDataflow(r.Client, instance, cluster); err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NifiFlowControllerServiceScheduling, errorfactory.NifiFlowScheduling:
				return RequeueAfter(time.Duration(5) * time.Second)
			default:
				return RequeueWithError(r.Log, "failed to run NifiDataflow", err)
			}
		}

		instance.Status.State = v1alpha1.DataflowStateRan
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDataflow status", err)
		}
	}

	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, cluster, instance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiCluster label on dataflow", err)
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiDataflow", err)
	}

	r.Log.Info("Ensured Dataflow")

	if instance.Spec.GetRunOnce() {
		return Reconciled()
	}

	return RequeueAfter(time.Duration(5) * time.Second)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiDataflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiDataflow{}).
		Complete(r)
}

func (r *NifiDataflowReconciler) ensureClusterLabel(ctx context.Context, cluster *v1alpha1.NifiCluster,
	flow *v1alpha1.NifiDataflow) (*v1alpha1.NifiDataflow, error) {

	labels := ApplyClusterRefLabel(cluster, flow.GetLabels())
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
	cluster *v1alpha1.NifiCluster) (reconcile.Result, error) {

	r.Log.Info("NiFi dataflow is marked for deletion")
	var err error
	if util.StringSliceContains(flow.GetFinalizers(), dataflowFinalizer) {
		if err = r.finalizeNifiDataflow(flow, cluster); err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NifiConnectionDropping, errorfactory.NifiFlowDraining:
				return RequeueAfter(time.Duration(5) * time.Second)
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
	flow.SetFinalizers(util.StringSliceRemove(flow.GetFinalizers(), dataflowFinalizer))
	_, err := r.updateAndFetchLatest(ctx, flow)
	return err
}

func (r *NifiDataflowReconciler) finalizeNifiDataflow(flow *v1alpha1.NifiDataflow, cluster *v1alpha1.NifiCluster) error {

	exists, err := dataflow.DataflowExist(r.Client, flow, cluster)
	if err != nil {
		return err
	}

	if exists {
		if _, err = dataflow.RemoveDataflow(r.Client, flow, cluster); err != nil {
			return err
		}
		r.Log.Info("Delete dataflow")
	}

	return nil
}
