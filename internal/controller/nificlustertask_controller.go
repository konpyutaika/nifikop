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
	"fmt"
	"maps"
	"reflect"
	"slices"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/scale"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

// NifiClusterTaskReconciler reconciles.
type NifiClusterTaskReconciler struct {
	client.Client
	Log              zap.Logger
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
	intervalNotReady := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_NOT_READY_REQUEUE_INTERVAL"], r.RequeueOffset)
	intervalRunning := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_RUNNING_REQUEUE_INTERVAL"], r.RequeueOffset)
	intervalTimeout := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_TIMEOUT_REQUEUE_INTERVAL"], r.RequeueOffset)
	// Fetch the NifiCluster instance
	instance := &v1.NifiCluster{}
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
	// A waiting node must not prevent another node from starting or completing a drain.
	// Use a private copy for each task because a failed status update can mutate its input.
	var delay time.Duration
	var taskErrors []error
	for _, id := range slices.Sorted(maps.Keys(instance.Status.NodesState)) {
		action := instance.Status.NodesState[id].GracefulActionState
		if !action.State.IsRequiredState() && !action.State.IsRunningState() && action.State != v1.GracefulDownscaleSucceeded {
			continue
		}
		cluster := instance.DeepCopy()
		current := instance.Status
		switch action.State {
		case v1.GracefulUpscaleRunning, v1.GracefulDownscaleRunning:
			if started, parseErr := nifiutil.ParseTimeStampToUnixTime(action.TaskStarted); parseErr == nil &&
				time.Since(started).Minutes() > cluster.Spec.NifiClusterTaskSpec.GetDurationMinutes() {
				r.Recorder.Eventf(cluster, corev1.EventTypeWarning, "GracefulTaskOverdue",
					"Node %s has been in %s (%s) since %s", id, action.State, action.ActionStep, action.TaskStarted)
			}
			err = r.handlePodRunningTask(cluster, current, id)
		case v1.GracefulUpscaleRequired:
			err = r.handlePodAddCCTask(cluster, current, id)
		case v1.GracefulDownscaleRequired:
			err = r.handlePodDeleteNCTask(cluster, current, id)
		case v1.GracefulDownscaleSucceeded:
			err = k8sutil.DeleteStatus(r.Client, id, cluster, current, r.Log)
		}
		// A conflict can invalidate the desired membership for subsequent tasks.
		if apierrors.IsConflict(err) {
			return RequeueWithError(r.Log, "NiFi cluster changed during task reconciliation", err)
		}
		next := intervalRunning
		switch errors.Cause(err).(type) {
		case nil:
			instance = cluster
		case errorfactory.NifiClusterNotReady, errorfactory.ResourceNotReady:
			next = intervalNotReady
		case errorfactory.NifiClusterTaskRunning:
		case errorfactory.NifiClusterTaskTimeout, errorfactory.NifiClusterTaskFailure:
			next = intervalTimeout
		default:
			taskErrors = append(taskErrors, errors.WrapIff(err, "reconciling node %s", id))
		}
		if delay == 0 || next < delay {
			delay = next
		}
	}
	if err := errors.Combine(taskErrors...); err != nil {
		return RequeueWithError(r.Log, "Failed to reconcile NiFi cluster tasks", err)
	}
	if delay > 0 {
		return RequeueAfter(delay)
	}
	return Reconciled()
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiClusterTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logCtr, err := GetLogConstructor(mgr, &v1.NifiCluster{})
	if err != nil {
		return err
	}
	builder := ctrl.NewControllerManagedBy(mgr).
		Named("nificlustertask").
		For(&v1.NifiCluster{}).
		WithLogConstructor(logCtr)

	err = builder.WithEventFilter(
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				object, err := meta.Accessor(e.ObjectNew)
				if err != nil {
					return false
				}
				if _, ok := object.(*v1.NifiCluster); ok {
					old := e.ObjectOld.(*v1.NifiCluster)
					new := e.ObjectNew.(*v1.NifiCluster)
					for _, nodeState := range new.Status.NodesState {
						if nodeState.GracefulActionState.State.IsRequiredState() || nodeState.GracefulActionState.State.IsRunningState() {
							return true
						}
					}
					// if reflect.DeepEqual(old.Status.NodesState, new.Status.NodesState) {
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

func (r *NifiClusterTaskReconciler) handlePodAddCCTask(nifiCluster *v1.NifiCluster, currentStatus v1.NifiClusterStatus, nodeId string) error {
	actionStep, taskStartTime, scaleErr := scale.UpScaleCluster(nodeId, nifiCluster.Namespace, nifiCluster.Name)
	if scaleErr != nil {
		r.Log.Error("Nifi cluster communication error during upscaling node(s)",
			zap.String("clusterName", nifiCluster.Name),
			zap.String("nodeId(s)", nodeId))
		return errorfactory.New(errorfactory.NifiClusterNotReady{}, scaleErr, fmt.Sprintf("Node id(s): %s", nodeId))
	}
	statusErr := k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster, currentStatus,
		v1.GracefulActionState{ActionStep: actionStep, State: v1.GracefulUpscaleRunning,
			TaskStarted: taskStartTime}, r.Log)
	if statusErr != nil {
		return errors.WrapIfWithDetails(statusErr, "could not update status for node", "clusterName", nifiCluster.Name, "id(s)", nodeId)
	}
	return nil
}

func (r *NifiClusterTaskReconciler) handlePodDeleteNCTask(nifiCluster *v1.NifiCluster, currentStatus v1.NifiClusterStatus, nodeId string) error {
	action := nifiCluster.Status.NodesState[nodeId].GracefulActionState
	if action.ActionStep == v1.ConnectNodeAction {
		started, parseErr := nifiutil.ParseTimeStampToUnixTime(action.TaskStarted)
		if parseErr != nil || time.Since(started).Minutes() >= nifiCluster.Spec.NifiClusterTaskSpec.GetDurationMinutes() {
			if err := k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster, currentStatus,
				v1.GracefulActionState{State: v1.GracefulDownscaleRequired}, r.Log); err != nil {
				return err
			}
			r.Recorder.Eventf(nifiCluster, corev1.EventTypeWarning, "RollbackReconnectExpired",
				"Retrying graceful disconnect for node %s after rollback reconnect started at %q", nodeId, action.TaskStarted)
			return nil
		}
	}
	clientConfig, err := config.GetClientConfigManager(r.Client, v1.ClusterReference{
		Name: nifiCluster.Name, Namespace: nifiCluster.Namespace,
	}).BuildConfig()
	if err != nil {
		return err
	}
	if action.ActionStep == v1.ConnectNodeAction {
		err := r.checkNCActionStep(clientConfig, nodeId, nifiCluster, currentStatus, v1.ConnectStatus)
		if err != nil {
			return err
		}
	}

	actionStep, taskStartTime, err := scale.DisconnectClusterNode(clientConfig, nodeId)
	if err == nificlient.ErrNifiClusterNodeNotFound && action.ActionStep == v1.CancelPendingPodAction && !clientConfig.SkipDescribeCluster {
		// The fresh membership listing excluded this node. The pod reconciler must
		// still verify it is unbound at deletion time and preserve its claims.
		return k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster, currentStatus,
			v1.GracefulActionState{State: v1.GracefulDownscaleRunning, ActionStep: v1.CancelPendingPodAction}, r.Log)
	}
	if err != nil {
		r.Log.Error("nifi cluster communication error during downscaling node(s) id(s)",
			zap.String("clusterName", nifiCluster.Name),
			zap.String("nodeId", nodeId))
		return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, fmt.Sprintf("node(s) id(s): %s", nodeId))
	}
	err = k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster, currentStatus,
		v1.GracefulActionState{ActionStep: actionStep, State: v1.GracefulDownscaleRunning,
			TaskStarted: taskStartTime}, r.Log)
	if err != nil {
		return errors.WrapIfWithDetails(err, "could not update status for node(s)", "clusterName", nifiCluster.Name, "id(s)", nodeId)
	}
	return nil
}

func (r *NifiClusterTaskReconciler) handlePodRunningTask(nifiCluster *v1.NifiCluster, currentStatus v1.NifiClusterStatus, nodeId string) error {
	action := nifiCluster.Status.NodesState[nodeId].GracefulActionState
	switch action.ActionStep {
	case v1.CancelPendingPodAction, v1.OffloadStatus, v1.RemovePodAction:
		// The cluster reconciler handles Kubernetes resource removal.
		return nil
	case v1.ConnectStatus:
		// Resume an upscale recorded by a previous operator or interrupted after CONNECTED.
		if action.State == v1.GracefulUpscaleRunning {
			return k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster, currentStatus,
				v1.GracefulActionState{State: v1.GracefulUpscaleSucceeded}, r.Log)
		}
		return nil
	}

	clientConfig, err := config.GetClientConfigManager(r.Client, v1.ClusterReference{
		Name: nifiCluster.Name, Namespace: nifiCluster.Namespace,
	}).BuildConfig()
	if err != nil {
		return err
	}

	// Persist one transition per reconcile. Another node can advance while this one waits.
	switch action.ActionStep {
	case v1.ConnectNodeAction:
		return r.checkNCActionStep(clientConfig, nodeId, nifiCluster, currentStatus, v1.ConnectStatus)
	case v1.DisconnectNodeAction:
		return r.checkNCActionStep(clientConfig, nodeId, nifiCluster, currentStatus, v1.DisconnectStatus)
	case v1.OffloadNodeAction:
		return r.checkNCActionStep(clientConfig, nodeId, nifiCluster, currentStatus, v1.OffloadStatus)
	case v1.RemoveNodeAction:
		return r.checkNCActionStep(clientConfig, nodeId, nifiCluster, currentStatus, v1.RemoveStatus)
	case v1.DisconnectStatus:
		action.ActionStep, action.TaskStarted, err = scale.OffloadClusterNode(clientConfig, nodeId)
	case v1.RemovePodStatus:
		action.ActionStep, action.TaskStarted, err = scale.RemoveClusterNode(clientConfig, nodeId)
	default:
		return nil
	}
	if err != nil {
		r.Log.Error("NiFi node action failed", zap.String("nodeId", nodeId), zap.Error(err))
		return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, "could not advance NiFi node task")
	}
	return k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster, currentStatus, action, r.Log)
}

func (r *NifiClusterTaskReconciler) checkNCActionStep(clientConfig *clientconfig.NifiConfig, nodeId string, nifiCluster *v1.NifiCluster, currentStatus v1.NifiClusterStatus, actionStep v1.ActionStep) error {
	nodeState := nifiCluster.Status.NodesState[nodeId]

	// Check Nifi cluster action status
	finished, err := scale.CheckIfNCActionStepFinished(nodeState.GracefulActionState.ActionStep, clientConfig, nodeId)
	if err != nil {
		r.Log.Sugar().Errorw("Nifi cluster communication error checking running task", "error", err,
			"clusterName", nifiCluster.Name,
			"actionStep", nodeState.GracefulActionState.ActionStep)
		return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, "nifi cluster communication error")
	}

	if finished {
		next := nodeState.GracefulActionState
		next.ActionStep = actionStep
		if next.State == v1.GracefulUpscaleRunning && actionStep == v1.ConnectStatus {
			next = v1.GracefulActionState{State: v1.GracefulUpscaleSucceeded}
		}
		if actionStep == v1.RemoveStatus {
			next.State = v1.GracefulDownscaleSucceeded
		}
		err = k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster, currentStatus, next, r.Log)
		if err != nil {
			return errors.WrapIfWithDetails(err, "could not update status for node(s)", "clusterName", nifiCluster.Name, "id(s)", nodeId)
		}
		return nil
	}

	if nodeState.GracefulActionState.ActionStep == v1.RemoveNodeAction {
		_, _, err := scale.RemoveClusterNode(clientConfig, nodeId)
		return err
	}

	if nodeState.GracefulActionState.State.IsRunningState() {
		parsedTime, err := nifiutil.ParseTimeStampToUnixTime(nodeState.GracefulActionState.TaskStarted)
		if err != nil {
			return errors.WrapIf(err, "could not parse timestamp "+nodeState.GracefulActionState.TaskStarted)
		}

		if time.Since(parsedTime).Minutes() > nifiCluster.Spec.NifiClusterTaskSpec.GetDurationMinutes() {
			requiredNCState, err := r.getCorrectRequiredNCState(nodeState.GracefulActionState.State)
			if err != nil {
				return err
			}

			r.Log.Sugar().Infow("Rollback nifi cluster task",
				"clusterName", nifiCluster.Name,
				"actionStep", nodeState.GracefulActionState.ActionStep)

			actionStep, taskStartTime, err := scale.ConnectClusterNode(clientConfig, nodeId)

			timedOutNodeNCState := v1.GracefulActionState{
				State:        requiredNCState,
				ActionStep:   actionStep,
				ErrorMessage: "Timed out waiting for the task to complete",
				TaskStarted:  taskStartTime,
			}

			if err != nil {
				return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, "nifi cluster communication error for cluster "+nifiCluster.Name)
			}

			return k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, nifiCluster, currentStatus, timedOutNodeNCState, r.Log)
		}
	}
	// cc task still in progress
	r.Log.Info("Nifi cluster task is still running", zap.String("actionStep", string(actionStep)))
	return errorfactory.New(errorfactory.NifiClusterTaskRunning{}, errors.New("Nifi cluster task is still running for cluster "+nifiCluster.Name), fmt.Sprintf("nc action step: %s", actionStep))
}

// getCorrectRequiredCCState returns the correct Required CC state based on that we upscale or downscale.
func (r *NifiClusterTaskReconciler) getCorrectRequiredNCState(ncState v1.State) (v1.State, error) {
	if ncState.IsDownscale() {
		return v1.GracefulDownscaleRequired, nil
	} else if ncState.IsUpscale() {
		return v1.GracefulUpscaleRequired, nil
	}

	return ncState, errors.NewWithDetails("could not determine if task state is upscale or downscale", "ncState", ncState)
}
