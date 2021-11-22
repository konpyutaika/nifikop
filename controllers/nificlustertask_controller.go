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
	"fmt"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/scale"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/Orange-OpenSource/nifikop/pkg/k8sutil"
	"github.com/Orange-OpenSource/nifikop/pkg/nificlient/config"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	"github.com/Orange-OpenSource/nifikop/pkg/util/clientconfig"
	nifiutil "github.com/Orange-OpenSource/nifikop/pkg/util/nifi"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
)

// NifiClusterTaskReconciler reconciles
type NifiClusterTaskReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	Recorder         record.EventRecorder
	RequeueIntervals map[string]int
	RequeueOffset    int
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiUserGroup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *NifiClusterTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("nificlustertask", req.NamespacedName)

	intervalNotReady := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_NOT_READY_REQUEUE_INTERVAL"], r.RequeueOffset)
	intervalRunning := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_RUNNING_REQUEUE_INTERVAL"], r.RequeueOffset)
	intervalTimeout := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_TIMEOUT_REQUEUE_INTERVAL"], r.RequeueOffset)
	// Fetch the NifiCluster instance
	instance := &v1alpha1.NifiCluster{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return Reconciled()
		}
		// Error reading the object - requeue the request.
		return RequeueWithError(r.Log, err.Error(), err)
	}

	r.Log.V(1).Info("Reconciling")

	var nodesWithRunningNCTask []string

	for nodeId, nodeStatus := range instance.Status.NodesState {
		if nodeStatus.GracefulActionState.State.IsRunningState() {
			nodesWithRunningNCTask = append(nodesWithRunningNCTask, nodeId)
		}
	}

	if len(nodesWithRunningNCTask) > 0 {
		err = r.handlePodRunningTask(instance, nodesWithRunningNCTask, r.Log)
	}

	if err != nil {
		switch errors.Cause(err).(type) {
		case errorfactory.NifiClusterNotReady, errorfactory.ResourceNotReady:
			return reconcile.Result{
				RequeueAfter: intervalNotReady,
			}, nil
		case errorfactory.NifiClusterTaskRunning:
			return reconcile.Result{
				RequeueAfter: intervalRunning,
			}, nil
		case errorfactory.NifiClusterTaskTimeout, errorfactory.NifiClusterTaskFailure:

			return reconcile.Result{
				RequeueAfter: intervalTimeout,
			}, nil
		default:
			return RequeueWithError(r.Log, err.Error(), err)
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
		err = r.handlePodAddCCTask(instance, nodesWithUpscaleRequired)
	} else if len(nodesWithDownscaleRequired) > 0 {
		err = r.handlePodDeleteNCTask(instance, nodesWithDownscaleRequired)
	}

	if err != nil {
		switch errors.Cause(err).(type) {
		case errorfactory.NifiClusterNotReady:
			return reconcile.Result{
				RequeueAfter: intervalNotReady,
			}, nil
		case errorfactory.NifiClusterTaskRunning:
			return reconcile.Result{
				RequeueAfter: intervalRunning,
			}, nil
		case errorfactory.NifiClusterTaskTimeout, errorfactory.NifiClusterTaskFailure:
			return reconcile.Result{
				RequeueAfter: intervalTimeout,
			}, nil
		default:
			return RequeueWithError(r.Log, err.Error(), err)
		}
	}

	var nodesWithDownscaleSucceeded []string

	for nodeId, nodeStatus := range instance.Status.NodesState {
		if nodeStatus.GracefulActionState.State == v1alpha1.GracefulDownscaleSucceeded {
			nodesWithDownscaleSucceeded = append(nodesWithDownscaleRequired, nodeId)
		}
	}

	if len(nodesWithDownscaleSucceeded) > 0 {
		err = r.handleNodeRemoveStatus(instance, nodesWithDownscaleSucceeded)
	}

	return Reconciled()
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiClusterTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiCluster{})

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

func (r *NifiClusterTaskReconciler) handlePodAddCCTask(nifiCluster *v1alpha1.NifiCluster, nodeIds []string) error {
	for _, nodeId := range nodeIds {
		actionStep, taskStartTime, scaleErr := scale.UpScaleCluster(nodeId, nifiCluster.Namespace, nifiCluster.Name)
		if scaleErr != nil {
			r.Log.Info("Nifi cluster communication error during upscaling node(s)", "nodeId(s)", nodeId)
			return errorfactory.New(errorfactory.NifiClusterNotReady{}, scaleErr, fmt.Sprintf("Node id(s): %s", nodeId))
		}
		statusErr := k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster,
			v1alpha1.GracefulActionState{ActionStep: actionStep, State: v1alpha1.GracefulUpscaleRunning,
				TaskStarted: taskStartTime}, r.Log)
		if statusErr != nil {
			return errors.WrapIfWithDetails(statusErr, "could not update status for node", "id(s)", nodeId)
		}
	}
	return nil
}

func (r *NifiClusterTaskReconciler) handlePodDeleteNCTask(nifiCluster *v1alpha1.NifiCluster, nodeIds []string) error {
	// Prepare cluster connection configurations
	var clientConfig *clientconfig.NifiConfig
	var err error

	// Get the client config manager associated to the cluster ref.
	clusterRef := v1alpha1.ClusterReference{
		Name:      nifiCluster.Name,
		Namespace: nifiCluster.Namespace,
	}
	configManager := config.GetClientConfigManager(r.Client, clusterRef)
	if clientConfig, err = configManager.BuildConfig(); err != nil {
		return err
	}

	for _, nodeId := range nodeIds {
		if nifiCluster.Status.NodesState[nodeId].GracefulActionState.ActionStep == v1alpha1.ConnectNodeAction {
			err := r.checkNCActionStep(nodeId, nifiCluster, v1alpha1.ConnectStatus, nil)
			if err != nil {
				return err
			}
		}

		actionStep, taskStartTime, err := scale.DisconnectClusterNode(clientConfig, nodeId)
		if err != nil {
			r.Log.Info(fmt.Sprintf("nifi cluster communication error during downscaling node(s) id(s): %s", nodeId))
			return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, fmt.Sprintf("node(s) id(s): %s", nodeId))
		}
		err = k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster,
			v1alpha1.GracefulActionState{ActionStep: actionStep, State: v1alpha1.GracefulDownscaleRunning,
				TaskStarted: taskStartTime}, r.Log)
		if err != nil {
			return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", nodeId)
		}

	}
	return nil
}

// TODO: Review logic to simplify it through generic method
func (r *NifiClusterTaskReconciler) handlePodRunningTask(nifiCluster *v1alpha1.NifiCluster, nodeIds []string, log logr.Logger) error {
	// Prepare cluster connection configurations
	var clientConfig *clientconfig.NifiConfig
	var err error

	// Get the client config manager associated to the cluster ref.
	clusterRef := v1alpha1.ClusterReference{
		Name:      nifiCluster.Name,
		Namespace: nifiCluster.Namespace,
	}
	configManager := config.GetClientConfigManager(r.Client, clusterRef)
	if clientConfig, err = configManager.BuildConfig(); err != nil {
		return err
	}

	for _, nodeId := range nodeIds {
		// Check if node finished to connect
		if nifiCluster.Status.NodesState[nodeId].GracefulActionState.ActionStep == v1alpha1.ConnectNodeAction {
			err := r.checkNCActionStep(nodeId, nifiCluster, v1alpha1.ConnectStatus, nil)
			if err != nil {
				return err
			}
		}

		// Check if node finished to disconnect
		if nifiCluster.Status.NodesState[nodeId].GracefulActionState.ActionStep == v1alpha1.DisconnectNodeAction {
			err := r.checkNCActionStep(nodeId, nifiCluster, v1alpha1.DisconnectStatus, nil)
			if err != nil {
				return err
			}
		}

		// If node is disconnected, performing offload
		if nifiCluster.Status.NodesState[nodeId].GracefulActionState.ActionStep == v1alpha1.DisconnectStatus {
			actionStep, taskStartTime, err := scale.OffloadClusterNode(clientConfig, nodeId)
			if err != nil {
				r.Log.Info(fmt.Sprintf("nifi cluster communication error during removing node id: %s", nodeId))
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
			err := r.checkNCActionStep(nodeId, nifiCluster, v1alpha1.OffloadStatus, nil)
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
			actionStep, taskStartTime, err := scale.RemoveClusterNode(clientConfig, nodeId)
			if err != nil {
				r.Log.Info(fmt.Sprintf("nifi cluster communication error during removing node id: %s", nodeId))
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
				nifiCluster, v1alpha1.RemoveStatus, &succeedState)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *NifiClusterTaskReconciler) checkNCActionStep(nodeId string, nifiCluster *v1alpha1.NifiCluster, actionStep v1alpha1.ActionStep, state *v1alpha1.State) error {
	// Prepare cluster connection configurations
	var clientConfig *clientconfig.NifiConfig
	var err error

	// Get the client config manager associated to the cluster ref.
	clusterRef := v1alpha1.ClusterReference{
		Name:      nifiCluster.Name,
		Namespace: nifiCluster.Namespace,
	}
	configManager := config.GetClientConfigManager(r.Client, clusterRef)
	if clientConfig, err = configManager.BuildConfig(); err != nil {
		return err
	}

	nodeState := nifiCluster.Status.NodesState[nodeId]

	// Check Nifi cluster action status
	finished, err := scale.CheckIfNCActionStepFinished(nodeState.GracefulActionState.ActionStep, clientConfig, nodeId)
	if err != nil {
		r.Log.Info(fmt.Sprintf("Nifi cluster communication error checking running task: %s", nodeState.GracefulActionState.ActionStep))
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
			}, r.Log)
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

			r.Log.Info(fmt.Sprintf("Rollback nifi cluster task: %s", nodeState.GracefulActionState.ActionStep))

			actionStep, taskStartTime, err := scale.ConnectClusterNode(clientConfig, nodeId)

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
			err = k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster, timedOutNodeNCState, r.Log)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", nodeId)
			}

		}
	}
	// cc task still in progress
	r.Log.Info("Nifi cluster task is still running", "actionStep", actionStep)
	return errorfactory.New(errorfactory.NifiClusterTaskRunning{}, errors.New("Nifi cluster task is still running"), fmt.Sprintf("nc action step: %s", actionStep))
}

// getCorrectRequiredCCState returns the correct Required CC state based on that we upscale or downscale
func (r *NifiClusterTaskReconciler) getCorrectRequiredNCState(ncState v1alpha1.State) (v1alpha1.State, error) {
	if ncState.IsDownscale() {
		return v1alpha1.GracefulDownscaleRequired, nil
	} else if ncState.IsUpscale() {
		return v1alpha1.GracefulUpscaleRequired, nil
	}

	return ncState, errors.NewWithDetails("could not determine if task state is upscale or downscale", "ncState", ncState)
}

func (r *NifiClusterTaskReconciler) handleNodeRemoveStatus(nifiCluster *v1alpha1.NifiCluster, nodeIds []string) error {
	for _, nodeId := range nodeIds {
		err := k8sutil.DeleteStatus(r.Client, nodeId, nifiCluster, r.Log)
		if err != nil {
			return errors.WrapIfWithDetails(err, "could not delete status for node", "id", nodeId)
		}
	}
	return nil
}
