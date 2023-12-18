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
	"fmt"
	"reflect"

	"emperror.dev/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/autoscale"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/util"
)

var autoscalerFinalizer string = fmt.Sprintf("nifinodegroupautoscalers.%s/finalizer", v1alpha1.GroupVersion.Group)

// NifiNodeGroupAutoscalerReconciler reconciles a NifiNodeGroupAutoscaler object.
type NifiNodeGroupAutoscalerReconciler struct {
	runtimeClient.Client
	APIReader       runtimeClient.Reader
	Scheme          *runtime.Scheme
	Log             zap.Logger
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifinodegroupautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifinodegroupautoscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifinodegroupautoscalers/finalizers,verbs=update
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificlusters,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiNodeGroupAutoscaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *NifiNodeGroupAutoscalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// @TODO: Manage dead lock when pending node because not enough resources
	// by implementing a brut force deletion on nificluster controller.
	nodeGroupAutoscaler := &v1alpha1.NifiNodeGroupAutoscaler{}
	err := r.Client.Get(ctx, req.NamespacedName, nodeGroupAutoscaler)

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
	current := nodeGroupAutoscaler.DeepCopy()
	patchInstance := runtimeClient.MergeFromWithOptions(nodeGroupAutoscaler.DeepCopy(), runtimeClient.MergeFromWithOptimisticLock{})

	// Check if marked for deletion and run finalizers
	if k8sutil.IsMarkedForDeletion(nodeGroupAutoscaler.ObjectMeta) {
		return r.checkFinalizers(ctx, nodeGroupAutoscaler, patchInstance)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(nodeGroupAutoscaler.GetFinalizers(), autoscalerFinalizer) {
		r.Log.Info(fmt.Sprintf("Adding Finalizer for NifiNodeGroupAutoscaler node group %s", nodeGroupAutoscaler.Spec.NodeConfigGroupId))
		nodeGroupAutoscaler.SetFinalizers(append(nodeGroupAutoscaler.GetFinalizers(), autoscalerFinalizer))
	}

	// lookup NifiCluster reference
	// we do not want cached objects here. We want an accurate state of what the cluster is right now, so bypass the client cache by using the APIReader directly.
	clusterRef := nodeGroupAutoscaler.Spec.ClusterRef
	clusterRef.Namespace = GetClusterRefNamespace(nodeGroupAutoscaler.Namespace, clusterRef)
	cluster := &v1.NifiCluster{}
	err = r.APIReader.Get(ctx,
		types.NamespacedName{
			Name:      clusterRef.Name,
			Namespace: clusterRef.Namespace,
		},
		cluster)
	if err != nil {
		return RequeueWithError(r.Log, fmt.Sprintf("failed to look up cluster reference %v+", clusterRef), err)
	}

	// Determine how many replicas there currently are and how many are desired for the appropriate node group
	numDesiredReplicas := nodeGroupAutoscaler.Spec.Replicas
	currentReplicas, err := r.getManagedNodes(nodeGroupAutoscaler, cluster.Spec.Nodes)
	if err != nil {
		return RequeueWithError(r.Log, "Failed to apply autoscaler node selector to cluster nodes", err)
	}
	numCurrentReplicas := int32(len(currentReplicas))

	// if the current number of nodes being managed by this autoscaler is different than the replica setting,
	// then set the autoscaler status to out of sync to indicate we're changing the NifiCluster node config
	// Additionally, if the autoscaler state is currently out of sync then scale up/down
	if numDesiredReplicas != numCurrentReplicas || nodeGroupAutoscaler.Status.State == v1alpha1.AutoscalerStateOutOfSync {
		r.Log.Info(fmt.Sprintf("Replicas changed from %d to %d", numCurrentReplicas, numDesiredReplicas))
		if err = r.updateAutoscalerReplicaState(ctx, nodeGroupAutoscaler, current.Status, v1alpha1.AutoscalerStateOutOfSync); err != nil {
			return RequeueWithError(r.Log, fmt.Sprintf("Failed to udpate node group autoscaler state for node group %s", nodeGroupAutoscaler.Spec.NodeConfigGroupId), err)
		}

		// json merge patch is a full-replace strategy. This means we must compute the entire NifiCluster.Spec.Nodes list as it should look after scaling.
		// The optimistic lock here ensures that we only patch the latest version of the NifiCluster to avoid stomping on changes any other process makes.
		// Ideally, we could use a strategic merge, but it's not supported for CRDs: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#advanced-features-and-flexibility
		clusterPatch := runtimeClient.MergeFromWithOptions(cluster.DeepCopy(), runtimeClient.MergeFromWithOptimisticLock{})

		if numDesiredReplicas > numCurrentReplicas {
			// need to increase node group
			numNodesToAdd := numDesiredReplicas - numCurrentReplicas
			r.Log.Info(fmt.Sprintf("Adding %d more nodes to cluster %s spec.nodes configuration for node group %s", numNodesToAdd, cluster.Name, nodeGroupAutoscaler.Spec.NodeConfigGroupId))

			if err = r.scaleUp(nodeGroupAutoscaler, cluster, numNodesToAdd); err != nil {
				return RequeueWithError(r.Log, fmt.Sprintf("Failed to scale cluster %s up for node group %s", cluster.Name, nodeGroupAutoscaler.Spec.NodeConfigGroupId), err)
			}
		} else if numDesiredReplicas < numCurrentReplicas {
			// need to decrease node group
			numNodesToRemove := numCurrentReplicas - numDesiredReplicas
			r.Log.Info(fmt.Sprintf("Removing %d nodes from cluster %s spec.nodes configuration for node group %s", numNodesToRemove, cluster.Name, nodeGroupAutoscaler.Spec.NodeConfigGroupId))

			if err = r.scaleDown(nodeGroupAutoscaler, cluster, numNodesToRemove); err != nil {
				return RequeueWithError(r.Log, fmt.Sprintf("Failed to scale cluster %s down for node group %s", cluster.Name, nodeGroupAutoscaler.Spec.NodeConfigGroupId), err)
			}
		}

		// patch nificluster resource with added/removed nodes
		if err = r.Client.Patch(ctx, cluster, clusterPatch); err != nil {
			return RequeueWithError(r.Log, fmt.Sprintf("Failed to patch nifi cluster with changes in nodes. Tried to apply the following patch:\n %v+", clusterPatch), err)
		}

		// update autoscaler state to InSync.
		if err = r.updateAutoscalerReplicaState(ctx, nodeGroupAutoscaler, current.Status, v1alpha1.AutoscalerStateInSync); err != nil {
			return RequeueWithError(r.Log, fmt.Sprintf("Failed to udpate node group autoscaler state for node group %s", nodeGroupAutoscaler.Spec.NodeConfigGroupId), err)
		}
	} else {
		r.Log.Info("Cluster replicas config and current number of replicas are the same", zap.Int32("replicas", nodeGroupAutoscaler.Spec.Replicas))
	}

	// update replica and replica status
	if err = r.updateAutoscalerReplicaStatus(ctx, cluster, current.Status, nodeGroupAutoscaler); err != nil {
		return RequeueWithError(r.Log, fmt.Sprintf("Failed to update node group autoscaler replica status for node group %s", nodeGroupAutoscaler.Spec.NodeConfigGroupId), err)
	}

	return RequeueAfter(util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset))
}

// scaleUp updates the provided cluster.Spec.Nodes list with the appropriate numNodesToAdd according to the autoscaler.Spec.UpscaleStrategy.
func (r *NifiNodeGroupAutoscalerReconciler) scaleUp(autoscaler *v1alpha1.NifiNodeGroupAutoscaler, cluster *v1.NifiCluster, numNodesToAdd int32) error {
	switch autoscaler.Spec.UpscaleStrategy {
	// Right now Simple is the only option and the default
	case v1alpha1.SimpleClusterUpscaleStrategy:
		fallthrough
	default:
		r.Log.Info(fmt.Sprintf("Using Simple upscale strategy for cluster %s node group %s", cluster.Name, autoscaler.Spec.NodeConfigGroupId))
		simple := &autoscale.SimpleHorizontalUpscaleStrategy{
			NifiCluster:             cluster,
			NifiNodeGroupAutoscaler: autoscaler,
		}
		nodesToAdd, err := simple.ScaleUp(numNodesToAdd)
		if err != nil {
			return errors.WrapIf(err, "Failed to scale up using the Simple strategy.")
		}
		cluster.Spec.Nodes = append(cluster.Spec.Nodes, nodesToAdd...)
	}
	r.Recorder.Eventf(autoscaler, corev1.EventTypeNormal, "Upscaling",
		"Adding %d more nodes to cluster %s spec.nodes configuration for node group %s", numNodesToAdd, cluster.Name, autoscaler.Spec.NodeConfigGroupId)

	return nil
}

// scaleUp updates the provided cluster.Spec.Nodes list with the appropriate numNodesToRemove according to the autoscaler.Spec.DownscaleStrategy.
func (r *NifiNodeGroupAutoscalerReconciler) scaleDown(autoscaler *v1alpha1.NifiNodeGroupAutoscaler, cluster *v1.NifiCluster, numNodesToRemove int32) error {
	switch autoscaler.Spec.DownscaleStrategy {
	// Right now LIFO is the only option and the default
	case v1alpha1.LIFOClusterDownscaleStrategy:
		fallthrough
	default:
		r.Log.Info(fmt.Sprintf("Using LIFO downscale strategy for cluster %s node group %s", cluster.Name, autoscaler.Spec.NodeConfigGroupId))
		// remove the last n nodes from the node list
		lifo := &autoscale.LIFOHorizontalDownscaleStrategy{
			NifiCluster:             cluster,
			NifiNodeGroupAutoscaler: autoscaler,
		}
		nodesToRemove, err := lifo.ScaleDown(numNodesToRemove)
		if err != nil {
			return errors.WrapIf(err, "Failed to scale cluster down via LIFO strategy.")
		}
		// remove the computed set of nodes from the cluster
		cluster.Spec.Nodes = util.SubtractNodes(cluster.Spec.Nodes, nodesToRemove)

		r.Recorder.Eventf(autoscaler, corev1.EventTypeNormal, "Downscaling",
			"Using LIFO downscale strategy for cluster %s node group %s", cluster.Name, autoscaler.Spec.NodeConfigGroupId)
	}

	return nil
}

// updateAutoscalerReplicaState updates the state of the autoscaler.
func (r *NifiNodeGroupAutoscalerReconciler) updateAutoscalerReplicaState(ctx context.Context, autoscaler *v1alpha1.NifiNodeGroupAutoscaler,
	currentStatus v1alpha1.NifiNodeGroupAutoscalerStatus, state v1alpha1.NodeGroupAutoscalerState) error {
	autoscaler.Status.State = state
	switch state {
	case v1alpha1.AutoscalerStateInSync:
		r.Recorder.Event(autoscaler, corev1.EventTypeNormal, "Synchronized", "Successfully synchronized node group autoscaler.")
	case v1alpha1.AutoscalerStateOutOfSync:
		r.Recorder.Event(autoscaler, corev1.EventTypeNormal, "Synchronizing", "The number of replicas for this node group has changed. Synchronizing.")
	}
	return r.updateStatus(ctx, autoscaler, currentStatus)
}

// TODO : discuss about replacing by looking for NifiCluster.Spec.Nodes instead
// updateAutoscalerReplicaStatus updates autoscaler replica status to inform the k8s scale subresource.
func (r *NifiNodeGroupAutoscalerReconciler) updateAutoscalerReplicaStatus(ctx context.Context, nifiCluster *v1.NifiCluster,
	currentStatus v1alpha1.NifiNodeGroupAutoscalerStatus, autoscaler *v1alpha1.NifiNodeGroupAutoscaler) error {
	podList, err := r.getCurrentReplicaPods(ctx, autoscaler)
	if err != nil {
		return err
	}

	replicas := v1alpha1.ClusterReplicas(int32(len(podList.Items)))
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: autoscaler.Spec.NodeLabelsSelector.MatchLabels,
	})
	if err != nil {
		return errors.WrapIf(err, "Failed to get label selector to update CR")
	}

	replicaSelector := v1alpha1.ClusterReplicaSelector(selector.String())
	autoscaler.Status.Replicas = replicas
	autoscaler.Status.Selector = replicaSelector

	return r.updateStatus(ctx, autoscaler, currentStatus)
}

// getCurrentReplicaPods searches for any pods created in this node scaler's node group.
func (r *NifiNodeGroupAutoscalerReconciler) getCurrentReplicaPods(ctx context.Context, autoscaler *v1alpha1.NifiNodeGroupAutoscaler) (*corev1.PodList, error) {
	podList := &corev1.PodList{}
	replicaLabels, err := autoscaler.Spec.NifiNodeGroupSelectorAsMap()
	if err != nil {
		return nil, err
	}
	// find replica pods for this autoscaler
	labelsToMatch := []map[string]string{
		replicaLabels,
	}
	matchingLabels := runtimeClient.MatchingLabels(util.MergeLabels(labelsToMatch...))

	err = r.Client.List(ctx, podList,
		runtimeClient.ListOption(runtimeClient.InNamespace(autoscaler.Namespace)), runtimeClient.ListOption(matchingLabels))
	if err != nil {
		return nil, errors.WrapIf(err, fmt.Sprintf("failed to query for replica podList for node group %s", autoscaler.Spec.NodeConfigGroupId))
	}
	return podList, nil
}

// getManagedNodes filters a set of nodes by an autoscaler's configured node selector.
func (r *NifiNodeGroupAutoscalerReconciler) getManagedNodes(autoscaler *v1alpha1.NifiNodeGroupAutoscaler, nodes []v1.Node) (managedNodes []v1.Node, err error) {
	selector, err := metav1.LabelSelectorAsSelector(autoscaler.Spec.NodeLabelsSelector)
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if selector.Matches(labels.Set(node.Labels)) {
			managedNodes = append(managedNodes, node)
		}
	}
	return managedNodes, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiNodeGroupAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logCtr, err := GetLogConstructor(mgr, &v1alpha1.NifiNodeGroupAutoscaler{})
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiNodeGroupAutoscaler{}).
		WithLogConstructor(logCtr).
		Complete(r)
}

func (r *NifiNodeGroupAutoscalerReconciler) checkFinalizers(ctx context.Context, autoscaler *v1alpha1.NifiNodeGroupAutoscaler, patcher runtimeClient.Patch) (reconcile.Result, error) {
	r.Log.Info("NifiNodeGroupAutoscaler is marked for deletion")

	var err error
	if util.StringSliceContains(autoscaler.GetFinalizers(), autoscalerFinalizer) {
		// no further actions necessary prior to removing finalizer.
		if err = r.removeFinalizer(ctx, autoscaler, patcher); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from autoscaler", err)
		}
	}

	return Reconciled()
}

func (r *NifiNodeGroupAutoscalerReconciler) removeFinalizer(ctx context.Context, autoscaler *v1alpha1.NifiNodeGroupAutoscaler, patcher runtimeClient.Patch) error {
	autoscaler.SetFinalizers(util.StringSliceRemove(autoscaler.GetFinalizers(), autoscalerFinalizer))
	_, err := r.updateAndFetchLatest(ctx, autoscaler, patcher)
	return err
}

func (r *NifiNodeGroupAutoscalerReconciler) updateAndFetchLatest(ctx context.Context,
	autoscaler *v1alpha1.NifiNodeGroupAutoscaler, patcher runtimeClient.Patch) (*v1alpha1.NifiNodeGroupAutoscaler, error) {
	typeMeta := autoscaler.TypeMeta
	err := r.Client.Patch(ctx, autoscaler, patcher)
	if err != nil {
		return nil, err
	}
	autoscaler.TypeMeta = typeMeta
	return autoscaler, nil
}

func (r *NifiNodeGroupAutoscalerReconciler) updateStatus(ctx context.Context, autoscaler *v1alpha1.NifiNodeGroupAutoscaler, currentStatus v1alpha1.NifiNodeGroupAutoscalerStatus) error {
	if !reflect.DeepEqual(autoscaler.Status, currentStatus) {
		return r.Client.Status().Update(ctx, autoscaler)
	}
	return nil
}
