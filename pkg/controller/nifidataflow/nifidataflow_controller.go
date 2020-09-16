// Copyright 2020 Orange SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package apis

package nifidataflow

import (
	"context"
	"reflect"
	"time"

	"emperror.dev/errors"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/dataflow"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/Orange-OpenSource/nifikop/pkg/k8sutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/controller/common"
	"github.com/go-logr/logr"

	"github.com/Orange-OpenSource/nifikop/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_nifidataflow")

var dataflowFinalizer = "finalizer.nifidataflows.nifi.orange.com"

// Add creates a new NifiCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, namespaces []string) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNifiDataflow{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nifidataflow-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NifiDataflow
	err = c.Watch(&source.Kind{Type: &v1alpha1.NifiDataflow{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	if err != nil {
		if _, ok := err.(*meta.NoKindMatchError); !ok {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileNifiDataflow implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNifiDataflow{}

// ReconcileNifiCluster reconciles a NifiDataflow object
type ReconcileNifiDataflow struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifidataflows,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifidataflows/status,verbs=get;update;patch

// Reconcile reads that state of the cluster for a NifiDataflow object and makes changes based on the state read
// and what is in the NifiDataflow.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNifiDataflow) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NifiDataflow")
	var err error

	// Get a context for the request
	ctx := context.Background()

	// Fetch the NifiDataflow instance
	instance := &v1alpha1.NifiDataflow{}
	if err = r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return common.Reconciled()
		}
		// Error reading the object - requeue the request.
		return common.RequeueWithError(reqLogger, err.Error(), err)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), dataflowFinalizer) {
		reqLogger.Info("Adding Finalizer for NifiDataflow")
		instance.SetFinalizers(append(instance.GetFinalizers(), dataflowFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to update NifiDataflow", err)
	}

	// Get the referenced NifiRegistryClient
	var registryClient *v1alpha1.NifiRegistryClient
	var registryClientNamespace string
	if instance.Spec.RegistryClientRef != nil {
		registryClientNamespace =
			common.GetRegistryClientRefNamespace(instance.Namespace, *instance.Spec.RegistryClientRef)

		if registryClient, err = k8sutil.LookupNifiRegistryClient(r.client,
			instance.Spec.RegistryClientRef.Name, registryClientNamespace); err != nil {

			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				reqLogger.Info("Registry client is already gone, there is nothing we can do")
				if err = r.removeFinalizer(ctx, instance); err != nil {
					return common.RequeueWithError(reqLogger, "failed to remove finalizer", err)
				}
				return common.Reconciled()
			}

			// the cluster does not exist - should have been caught pre-flight
			return common.RequeueWithError(reqLogger, "failed to lookup referenced registry client", err)
		}
	}

	var parameterContext *v1alpha1.NifiParameterContext
	var parameterContextNamespace string
	if instance.Spec.ParameterContextRef != nil {
		parameterContextNamespace =
			common.GetParameterContextRefNamespace(instance.Namespace, *instance.Spec.ParameterContextRef)

		if parameterContext, err = k8sutil.LookupNifiParameterContext(r.client,
			instance.Spec.ParameterContextRef.Name, parameterContextNamespace); err != nil {

			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				reqLogger.Info("Parameter context is already gone, there is nothing we can do")
				if err = r.removeFinalizer(ctx, instance); err != nil {
					return common.RequeueWithError(reqLogger, "failed to remove finalizer", err)
				}
				return common.Reconciled()
			}

			// the cluster does not exist - should have been caught pre-flight
			return common.RequeueWithError(reqLogger, "failed to lookup referenced parameter-contest", err)
		}
	}

	// Check if cluster references are the same
	clusterNamespace := common.GetClusterRefNamespace(instance.Namespace, instance.Spec.ClusterRef)
	if registryClient != nil &&
		(registryClientNamespace != clusterNamespace ||
			registryClient.Spec.ClusterRef.Name != instance.Spec.ClusterRef.Name ||
			(parameterContext != nil &&
				(parameterContextNamespace != clusterNamespace ||
					parameterContext.Spec.ClusterRef.Name != instance.Spec.ClusterRef.Name))) {

		return common.RequeueWithError(
			reqLogger,
			"failed to lookup referenced cluster, due to inconsistency",
			errors.New("inconsistent cluster references"))
	}

	var cluster *v1alpha1.NifiCluster
	if cluster, err = k8sutil.LookupNifiCluster(r.client, instance.Spec.ClusterRef.Name, clusterNamespace); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			reqLogger.Info("Cluster is already gone, there is nothing we can do")
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return common.RequeueWithError(reqLogger, "failed to remove finalizer", err)
			}
			return common.Reconciled()
		}

		// the cluster does not exist - should have been caught pre-flight
		return common.RequeueWithError(reqLogger, "failed to lookup referenced cluster", err)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, reqLogger, instance, cluster)
	}

	if *instance.Spec.RunOnce && instance.Status.State == v1alpha1.DataflowStateRan {
		return common.Reconciled()
	}

	// Check if the dataflow already exist
	existing, err := dataflow.DataflowExist(r.client, instance, cluster)
	if err != nil {
		return common.RequeueWithError(reqLogger, "failure checking for existing dataflow", err)
	}

	// Create dataflow if it doesn't already exist
	if !existing {

		processGroupStatus, err := dataflow.CreateDataflow(r.client, instance, cluster, registryClient)
		if err != nil {
			return common.RequeueWithError(reqLogger, "failure creating dataflow", err)
		}

		// Set dataflow status
		instance.Status = *processGroupStatus
		instance.Status.State = v1alpha1.DataflowStateCreated

		if err := r.client.Status().Update(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to update NifiDataflow status", err)
		}

		existing = true
	}

	// In case where the flow is not sync
	if instance.Status.State == v1alpha1.DataflowStateOutOfSync {
		status, err := dataflow.SyncDataflow(r.client, instance, cluster, registryClient, parameterContext)
		if status != nil {
			instance.Status = *status
			if err := r.client.Status().Update(ctx, instance); err != nil {
				return common.RequeueWithError(reqLogger, "failed to update NifiDataflow status", err)
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
				return common.RequeueWithError(reqLogger, "failed to sync NiFiDataflow", err)
			}
		}

		instance.Status.State = v1alpha1.DataflowStateInSync
		if err := r.client.Status().Update(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to update NifiDataflow status", err)
		}
	}

	// Check if the flow is out of sync
	isOutOfSink, err := dataflow.IsOutOfSyncDataflow(r.client, instance, cluster, registryClient, parameterContext)
	if err != nil {
		return common.RequeueWithError(reqLogger, "failed to check NifiDataflow sync", err)
	}

	if isOutOfSink {
		instance.Status.State = v1alpha1.DataflowStateOutOfSync
		if err := r.client.Status().Update(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to update NifiDataflow status", err)
		}
		return common.Requeue()
	}

	// Schedule the flow
	if instance.Status.State == v1alpha1.DataflowStateCreated ||
		instance.Status.State == v1alpha1.DataflowStateStarting ||
		instance.Status.State == v1alpha1.DataflowStateInSync ||
		(!*instance.Spec.RunOnce && instance.Status.State == v1alpha1.DataflowStateRan) {

		instance.Status.State = v1alpha1.DataflowStateStarting
		if err := r.client.Status().Update(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to update NifiDataflow status", err)
		}

		if err := dataflow.ScheduleDataflow(r.client, instance, cluster); err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NifiFlowControllerServiceScheduling, errorfactory.NifiFlowScheduling:
				return common.RequeueAfter(time.Duration(5) * time.Second)
			default:
				return common.RequeueWithError(reqLogger, "failed to run NifiDataflow", err)
			}
		}

		instance.Status.State = v1alpha1.DataflowStateRan
		if err := r.client.Status().Update(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to update NifiDataflow status", err)
		}
	}

	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, cluster, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to ensure NifiCluster label on dataflow", err)
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to update NifiDataflow", err)
	}

	reqLogger.Info("Ensured Dataflow")

	if *instance.Spec.RunOnce {
		return common.Reconciled()
	}

	return common.RequeueAfter(time.Duration(5) * time.Second)
}

func (r *ReconcileNifiDataflow) ensureClusterLabel(ctx context.Context, cluster *v1alpha1.NifiCluster,
	flow *v1alpha1.NifiDataflow) (*v1alpha1.NifiDataflow, error) {

	labels := common.ApplyClusterRefLabel(cluster, flow.GetLabels())
	if !reflect.DeepEqual(labels, flow.GetLabels()) {
		flow.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, flow)
	}
	return flow, nil
}

func (r *ReconcileNifiDataflow) updateAndFetchLatest(ctx context.Context,
	flow *v1alpha1.NifiDataflow) (*v1alpha1.NifiDataflow, error) {

	typeMeta := flow.TypeMeta
	err := r.client.Update(ctx, flow)
	if err != nil {
		return nil, err
	}
	flow.TypeMeta = typeMeta
	return flow, nil
}

func (r *ReconcileNifiDataflow) checkFinalizers(ctx context.Context, reqLogger logr.Logger,
	flow *v1alpha1.NifiDataflow, cluster *v1alpha1.NifiCluster) (reconcile.Result, error) {

	reqLogger.Info("NiFi dataflow is marked for deletion")
	var err error
	if util.StringSliceContains(flow.GetFinalizers(), dataflowFinalizer) {
		if err = r.finalizeNifiDataflow(reqLogger, flow, cluster); err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NifiConnectionDropping, errorfactory.NifiFlowDraining:
				return common.RequeueAfter(time.Duration(5) * time.Second)
			default:
				return common.RequeueWithError(reqLogger, "failed to finalize NiFiDataflow", err)
			}
		}
		if err = r.removeFinalizer(ctx, flow); err != nil {
			return common.RequeueWithError(reqLogger, "failed to remove finalizer from dataflow", err)
		}
	}
	return common.Reconciled()
}

func (r *ReconcileNifiDataflow) removeFinalizer(ctx context.Context, flow *v1alpha1.NifiDataflow) error {
	flow.SetFinalizers(util.StringSliceRemove(flow.GetFinalizers(), dataflowFinalizer))
	_, err := r.updateAndFetchLatest(ctx, flow)
	return err
}

func (r *ReconcileNifiDataflow) finalizeNifiDataflow(reqLogger logr.Logger, flow *v1alpha1.NifiDataflow,
	cluster *v1alpha1.NifiCluster) error {

	exists, err := dataflow.DataflowExist(r.client, flow, cluster)
	if err != nil {
		return err
	}

	if exists {
		if _, err = dataflow.RemoveDataflow(r.client, flow, cluster); err != nil {
			return err
		}
		reqLogger.Info("Delete dataflow")
	}

	return nil
}
