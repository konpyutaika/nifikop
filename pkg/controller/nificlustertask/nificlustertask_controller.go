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

package nificlustertask

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"emperror.dev/errors"
	v1alpha1 "github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/scale"
	"github.com/Orange-OpenSource/nifikop/pkg/controller/common"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/Orange-OpenSource/nifikop/pkg/k8sutil"
	nifiutil "github.com/Orange-OpenSource/nifikop/pkg/util/nifi"
	"github.com/go-logr/logr"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("controller_nificlustertask")

// Add creates a new NifiCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, namespaces []string) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNifiClusterTask{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	builder := ctrl.NewControllerManagedBy(mgr).For(&v1alpha1.NifiCluster{}).Named("nificlustertask-controller")

	// TODO : review event filter
	err := builder.WithEventFilter(
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				object, err := meta.Accessor(e.ObjectNew)
				if err != nil {
					return false
				}
				if _, ok := object.(*v1alpha1.NifiCluster); ok {
					old := e.ObjectOld.(*v1alpha1.NifiCluster)
					new := e.ObjectNew.(*v1alpha1.NifiCluster)
					for _, nodeState := range new.Status.NodesState {
						if nodeState.GracefulActionState.State.IsRequiredState() || nodeState.GracefulActionState.State.IsRunningState() {
							return true
						}
					}
					//if reflect.DeepEqual(old.Status.NodesState, new.Status.NodesState) {
					//	return true
					//}
					if !reflect.DeepEqual(old.Status.NodesState, new.Status.NodesState) ||
						old.GetDeletionTimestamp() != new.GetDeletionTimestamp() ||
						old.GetGeneration() != new.GetGeneration() {
						return true
					}
					return false
				}
				return true
			},
		}).Complete(r)

	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileNifiClusterTask implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNifiClusterTask{}

// ReconcileNifiCluster reconciles a NifiCluster object
type ReconcileNifiClusterTask struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a NifiCluster object and makes changes based on the state read
// and what is in the NifiCluster.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNifiClusterTask) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NifiCluster")

	ctx := context.Background()

	// Fetch the NifiCluster instance
	instance := &v1alpha1.NifiCluster{}
	err := r.Client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return common.Reconciled()
		}
		// Error reading the object - requeue the request.
		return common.RequeueWithError(reqLogger, err.Error(), err)
	}

	log.V(1).Info("Reconciling")

	var nodesWithRunningNCTask []string

	for nodeId, nodeStatus := range instance.Status.NodesState {
		if nodeStatus.GracefulActionState.State.IsRunningState() {
			nodesWithRunningNCTask = append(nodesWithRunningNCTask, nodeId)
		}
	}

	if len(nodesWithRunningNCTask) > 0 {
		err = r.handlePodRunningTask(instance, nodesWithRunningNCTask, log)
	}

	if err != nil {
		switch errors.Cause(err).(type) {
		case errorfactory.NifiClusterNotReady, errorfactory.ResourceNotReady:
			return reconcile.Result{
				RequeueAfter: time.Duration(15) * time.Second,
			}, nil
		case errorfactory.NifiClusterTaskRunning:
			return reconcile.Result{
				RequeueAfter: time.Duration(20) * time.Second,
			}, nil
		case errorfactory.NifiClusterTaskTimeout, errorfactory.NifiClusterTaskFailure:
			return reconcile.Result{
				RequeueAfter: time.Duration(20) * time.Second,
			}, nil
		default:
			return common.RequeueWithError(log, err.Error(), err)
		}
	}

	var nodesWithDownscaleRequired []string
	var nodesWithUpscaleRequired []string

	for nodeId, nodeStatus := range instance.Status.NodesState {
		if nodeStatus.GracefulActionState.State == v1alpha1.GracefulUpscaleRequired {
			nodesWithUpscaleRequired = append(nodesWithUpscaleRequired, nodeId)
		} else if nodeStatus.GracefulActionState.State == v1alpha1.GracefulDownscaleRequired {
			nodesWithDownscaleRequired = append(nodesWithDownscaleRequired, nodeId)
		}
	}

	if len(nodesWithUpscaleRequired) > 0 {
		err = r.handlePodAddCCTask(instance, nodesWithUpscaleRequired, log)
	} else if len(nodesWithDownscaleRequired) > 0 {
		err = r.handlePodDeleteNCTask(instance, nodesWithDownscaleRequired, log)
	}

	if err != nil {
		switch errors.Cause(err).(type) {
		case errorfactory.NifiClusterNotReady:
			return reconcile.Result{
				RequeueAfter: time.Duration(15) * time.Second,
			}, nil
		case errorfactory.NifiClusterTaskRunning:
			return reconcile.Result{
				RequeueAfter: time.Duration(20) * time.Second,
			}, nil
		case errorfactory.NifiClusterTaskTimeout, errorfactory.NifiClusterTaskFailure:
			return reconcile.Result{
				RequeueAfter: time.Duration(20) * time.Second,
			}, nil
		default:
			return common.RequeueWithError(log, err.Error(), err)
		}
	}

	var nodesWithDownscaleSucceeded []string

	for nodeId, nodeStatus := range instance.Status.NodesState {
		if nodeStatus.GracefulActionState.State == v1alpha1.GracefulDownscaleSucceeded {
			nodesWithDownscaleSucceeded = append(nodesWithDownscaleRequired, nodeId)
		}
	}

	if len(nodesWithDownscaleSucceeded) > 0 {
		err = r.handleNodeRemoveStatus(instance, nodesWithDownscaleSucceeded, log)
	}

	return common.Reconciled()
}

func (r *ReconcileNifiClusterTask) handlePodAddCCTask(nifiCluster *v1alpha1.NifiCluster, nodeIds []string, log logr.Logger) error {
	for _, nodeId := range nodeIds {
		actionStep, taskStartTime, scaleErr := scale.UpScaleCluster(nodeId, nifiCluster.Namespace, nifiCluster.Name)
		if scaleErr != nil {
			log.Info("Nifi cluster communication error during upscaling node(s)", "nodeId(s)", nodeId)
			return errorfactory.New(errorfactory.NifiClusterNotReady{}, scaleErr, fmt.Sprintf("Node id(s): %s", nodeId))
		}
		statusErr := k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster,
			v1alpha1.GracefulActionState{ActionStep: actionStep, State: v1alpha1.GracefulUpscaleRunning,
				TaskStarted: taskStartTime}, log)
		if statusErr != nil {
			return errors.WrapIfWithDetails(statusErr, "could not update status for node", "id(s)", nodeId)
		}
	}
	return nil
}

func (r *ReconcileNifiClusterTask) handlePodDeleteNCTask(nifiCluster *v1alpha1.NifiCluster, nodeIds []string, log logr.Logger) error {
	for _, nodeId := range nodeIds {
		if nifiCluster.Status.NodesState[nodeId].GracefulActionState.ActionStep == v1alpha1.ConnectNodeAction {
			err := r.checkNCActionStep(nodeId, nifiCluster, v1alpha1.ConnectStatus, nil, log)
			if err != nil {
				return err
			}
		}

		actionStep, taskStartTime, err := scale.DisconnectClusterNode(r.Client, nifiCluster, nodeId)
		if err != nil {
			log.Info(fmt.Sprintf("nifi cluster communication error during downscaling node(s) id(s): %s", nodeId))
			return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, fmt.Sprintf("node(s) id(s): %s", nodeId))
		}
		err = k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster,
			v1alpha1.GracefulActionState{ActionStep: actionStep, State: v1alpha1.GracefulDownscaleRunning,
				TaskStarted: taskStartTime}, log)
		if err != nil {
			return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", nodeId)
		}

	}
	return nil
}

// TODO: Review logic to simplify it through generic method
func (r *ReconcileNifiClusterTask) handlePodRunningTask(nifiCluster *v1alpha1.NifiCluster, nodeIds []string, log logr.Logger) error {

	for _, nodeId := range nodeIds {
		// Check if node finished to connect
		if nifiCluster.Status.NodesState[nodeId].GracefulActionState.ActionStep == v1alpha1.ConnectNodeAction {
			err := r.checkNCActionStep(nodeId, nifiCluster, v1alpha1.ConnectStatus, nil, log)
			if err != nil {
				return err
			}
		}

		// Check if node finished to disconnect
		if nifiCluster.Status.NodesState[nodeId].GracefulActionState.ActionStep == v1alpha1.DisconnectNodeAction {
			err := r.checkNCActionStep(nodeId, nifiCluster, v1alpha1.DisconnectStatus, nil, log)
			if err != nil {
				return err
			}
		}

		// If node is disconnected, performing offload
		if nifiCluster.Status.NodesState[nodeId].GracefulActionState.ActionStep == v1alpha1.DisconnectStatus {
			actionStep, taskStartTime, err := scale.OffloadClusterNode(r.Client, nifiCluster, nodeId)
			if err != nil {
				log.Info(fmt.Sprintf("nifi cluster communication error during removing node id: %s", nodeId))
				return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, fmt.Sprintf("node id: %s", nodeId))
			}
			err = k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster,
				v1alpha1.GracefulActionState{ActionStep: actionStep, State: v1alpha1.GracefulDownscaleRunning,
					TaskStarted: taskStartTime}, log)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", nodeId)
			}
		}

		// Check if node finished to offload data
		if nifiCluster.Status.NodesState[nodeId].GracefulActionState.ActionStep == v1alpha1.OffloadNodeAction {
			err := r.checkNCActionStep(nodeId, nifiCluster, v1alpha1.OffloadStatus, nil, log)
			if err != nil {
				return err
			}
		}

		// TODO : Investigate workaround used because of error :
		// 2020-02-18 08:30:07,850 INFO [Process Cluster Protocol Request-4] o.a.n.c.c.node.NodeClusterCoordinator Status of nifi-12-node.nifi-headless.nifi-demo.svc.cluster.local:8080
		// changed from null to NodeConnectionStatus[nodeId=nifi-12-node.nifi-headless.nifi-demo.svc.cluster.local:8080, state=DISCONNECTED, Disconnect Code=Node was Shutdown,
		// Disconnect Reason=Node was Shutdown, updateId=33]
		// If pod finished deletion
		// TODO : work here to manage node Status and state (If disconnected && Removing)
		if nifiCluster.Status.NodesState[nodeId].GracefulActionState.ActionStep == v1alpha1.RemovePodStatus {
			actionStep, taskStartTime, err := scale.RemoveClusterNode(r.Client, nifiCluster, nodeId)
			if err != nil {
				log.Info(fmt.Sprintf("nifi cluster communication error during removing node id: %s", nodeId))
				return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, fmt.Sprintf("node id: %s", nodeId))
			}
			err = k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster,
				v1alpha1.GracefulActionState{ActionStep: actionStep, State: v1alpha1.GracefulDownscaleRunning,
					TaskStarted: taskStartTime}, log)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", nodeId)
			}
		}

		if nifiCluster.Status.NodesState[nodeId].GracefulActionState.ActionStep == v1alpha1.RemoveNodeAction {
			succeedState := v1alpha1.GracefulDownscaleSucceeded
			err := r.checkNCActionStep(nodeId,
				nifiCluster, v1alpha1.RemoveStatus, &succeedState, log)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ReconcileNifiClusterTask) checkNCActionStep(nodeId string, nifiCluster *v1alpha1.NifiCluster, actionStep v1alpha1.ActionStep, state *v1alpha1.State, log logr.Logger) error {
	nodeState := nifiCluster.Status.NodesState[nodeId]

	// Check Nifi cluster action status
	finished, err := scale.CheckIfNCActionStepFinished(nodeState.GracefulActionState.ActionStep, r.Client, nifiCluster, nodeId)
	if err != nil {
		log.Info(fmt.Sprintf("Nifi cluster communication error checking running task: %s", nodeState.GracefulActionState.ActionStep))
		return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, "nifi cluster communication error")
	}

	if finished {
		succeedState := nodeState.GracefulActionState.State
		if state != nil {
			succeedState = *state
		}
		err = k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster,
			v1alpha1.GracefulActionState{State: succeedState,
				TaskStarted: nodeState.GracefulActionState.TaskStarted,
				ActionStep:  actionStep,
			}, log)
		if err != nil {
			return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", nodeId)
		}
	}

	if nodeState.GracefulActionState.State.IsRunningState() {
		parsedTime, err := nifiutil.ParseTimeStampToUnixTime(nodeState.GracefulActionState.TaskStarted)
		if err != nil {
			return errors.WrapIf(err, "could not parse timestamp")
		}

		now, err := nifiutil.ParseTimeStampToUnixTime(time.Now().Format(nifiutil.TimeStampLayout))
		if err != nil {
			return errors.WrapIf(err, "could not parse timestamp")
		}

		if now.Sub(parsedTime).Minutes() > nifiCluster.Spec.NifiClusterTaskSpec.GetDurationMinutes() {
			requiredNCState, err := r.getCorrectRequiredNCState(nodeState.GracefulActionState.State)
			if err != nil {
				return err
			}

			log.Info(fmt.Sprintf("Rollback nifi cluster task: %s", nodeState.GracefulActionState.ActionStep))

			actionStep, taskStartTime, err := scale.ConnectClusterNode(r.Client, nifiCluster, nodeId)

			timedOutNodeNCState := v1alpha1.GracefulActionState{
				State:        requiredNCState,
				ActionStep:   actionStep,
				ErrorMessage: "Timed out waiting for the task to complete",
				TaskStarted:  taskStartTime,
			}

			if err != nil {
				return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, "nifi cluster communication error")
			}

			if err != nil {
				return err
			}
			err = k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster, timedOutNodeNCState, log)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", nodeId)
			}

		}
	}
	// cc task still in progress
	log.Info("Nifi cluster task is still running", "actionStep", actionStep)
	return errorfactory.New(errorfactory.NifiClusterTaskRunning{}, errors.New("Nifi cluster task is still running"), fmt.Sprintf("nc action step: %s", actionStep))
}

// getCorrectRequiredCCState returns the correct Required CC state based on that we upscale or downscale
func (r *ReconcileNifiClusterTask) getCorrectRequiredNCState(ncState v1alpha1.State) (v1alpha1.State, error) {
	if ncState.IsDownscale() {
		return v1alpha1.GracefulDownscaleRequired, nil
	} else if ncState.IsUpscale() {
		return v1alpha1.GracefulUpscaleRequired, nil
	}

	return ncState, errors.NewWithDetails("could not determine if task state is upscale or downscale", "ncState", ncState)
}

func (r *ReconcileNifiClusterTask) handleNodeRemoveStatus(nifiCluster *v1alpha1.NifiCluster, nodeIds []string, log logr.Logger) error {
	for _, nodeId := range nodeIds {
		err := k8sutil.DeleteStatus(r.Client, nodeId, nifiCluster, log)
		if err != nil {
			return errors.WrapIfWithDetails(err, "could not delete status for node", "id", nodeId)
		}
	}
	return nil
}
