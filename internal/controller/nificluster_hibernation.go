package controller

import (
	"context"
	"fmt"
	"sort"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/util"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

const (
	eventNifiClusterHibernationPending = "NifiClusterHibernationPending"
	eventNifiClusterHibernating        = "NifiClusterHibernating"
	eventNifiClusterHibernated         = "NifiClusterHibernated"
	eventNifiClusterWakingUp           = "NifiClusterWakingUp"
)

// nodesBlockingHibernation returns the nodes whose upscale or downscale is still in progress. Hibernation
// must not start while there are some: removed nodes are only cleaned up through their running pods, so
// stopping those pods would leave the downscale stuck and the cluster never ready again.
// Once hibernation has started, it goes on whatever the node status.
func nodesBlockingHibernation(cluster *v1.NifiCluster) []string {
	if cluster.Status.State.IsHibernation() {
		return nil
	}
	inSpec := make(map[string]bool, len(cluster.Spec.Nodes))
	for _, node := range cluster.Spec.Nodes {
		inSpec[fmt.Sprint(node.Id)] = true
	}
	blocking := make([]string, 0)
	for nodeId, nodeState := range cluster.Status.NodesState {
		state := nodeState.GracefulActionState.State
		if !inSpec[nodeId] || state.IsRunningState() || state.IsRequiredState() {
			blocking = append(blocking, nodeId)
		}
	}
	sort.Strings(blocking)
	return blocking
}

// reconcileHibernation stops every NiFi pod of the cluster. Unlike removing nodes from spec.nodes, it never
// deletes PVCs, node config secrets or node services, and it keeps the node status so that the nodes restart
// on their existing volumes when hibernation is turned off. No call is made to the NiFi API.
func (r *NifiClusterReconciler) reconcileHibernation(ctx context.Context, instance *v1.NifiCluster,
	current *v1.NifiCluster, patcher client.Patch) (reconcile.Result, error) {
	intervalRunning := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_RUNNING_REQUEUE_INTERVAL"], r.RequeueOffset)
	intervalHibernated := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_NO_NODE_INTERVAL"], r.RequeueOffset)

	// A cluster created with hibernate: true must still get its finalizers, so PVCs are handled on deletion.
	instance, err := r.ensureFinalizers(ctx, instance, patcher)
	if err != nil {
		return RequeueWithError(r.Log, "failed to ensure finalizers on nificluster instance "+current.Name, err)
	}

	if !instance.Status.State.IsHibernation() {
		r.Log.Info("NifiCluster starting hibernation", zap.String("clusterName", instance.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, eventNifiClusterHibernating,
			"NifiCluster is hibernating, stopping all nodes")
	}

	podList := &corev1.PodList{}
	if err := r.Client.List(ctx, podList,
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(nifiutil.LabelsForNifi(instance.Name))); err != nil {
		return RequeueWithError(r.Log, "failed to list pods of nificluster "+instance.Name, err)
	}

	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.DeletionTimestamp != nil {
			continue
		}
		r.Log.Info("Deleting pod for hibernation",
			zap.String("clusterName", instance.Name),
			zap.String("podName", pod.Name))
		if err := r.Client.Delete(ctx, pod); err != nil && !apierrors.IsNotFound(err) {
			return RequeueWithError(r.Log, "failed to delete pod "+pod.Name+" for hibernation", err)
		}
	}

	// Mark every node as not ready, so the dependent resources wait for the cluster after wake-up.
	notReadyNodes := make([]string, 0)
	for nodeId, nodeState := range instance.Status.NodesState {
		if nodeState.PodIsReady {
			notReadyNodes = append(notReadyNodes, nodeId)
		}
	}
	sort.Strings(notReadyNodes)
	if len(notReadyNodes) > 0 {
		if err := k8sutil.UpdateNodeStatus(r.Client, notReadyNodes, instance, current.Status, false, r.Log); err != nil {
			return RequeueWithError(r.Log, "failed to update node readiness for hibernation", err)
		}
	}

	if len(podList.Items) > 0 {
		if err := k8sutil.UpdateCRStatus(r.Client, instance, current.Status, v1.NifiClusterHibernating, r.Log); err != nil {
			return RequeueWithError(r.Log, err.Error(), err)
		}
		return RequeueAfter(intervalRunning)
	}

	if instance.Status.State != v1.NifiClusterHibernated {
		r.Log.Info("NifiCluster hibernated", zap.String("clusterName", instance.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, eventNifiClusterHibernated,
			"NifiCluster is hibernated, all nodes are stopped")
	}
	if err := k8sutil.UpdateCRStatus(r.Client, instance, current.Status, v1.NifiClusterHibernated, r.Log); err != nil {
		return RequeueWithError(r.Log, err.Error(), err)
	}
	return RequeueAfter(intervalHibernated)
}

// wakeUpFromHibernation moves a hibernated cluster back to the initialized state. The usual reconcile then
// recreates the pods, which reattach to their PVCs, and sets the cluster to running once the nodes are ready.
// A cluster created with hibernate: true has no node status yet: it goes to the initializing state instead,
// so the reconcile marks its nodes as init cluster nodes before creating them.
func (r *NifiClusterReconciler) wakeUpFromHibernation(instance, current *v1.NifiCluster) error {
	r.Log.Info("NifiCluster waking up from hibernation", zap.String("clusterName", instance.Name))
	r.Recorder.Event(instance, corev1.EventTypeNormal, eventNifiClusterWakingUp,
		"NifiCluster is waking up from hibernation")
	state := v1.NifiClusterInitialized
	if len(instance.Status.NodesState) == 0 {
		state = v1.NifiClusterInitializing
	}
	return k8sutil.UpdateCRStatus(r.Client, instance, current.Status, state, r.Log)
}
