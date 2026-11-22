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
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
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
	patchInstance := runtimeClient.MergeFrom(nodeGroupAutoscaler.DeepCopy())

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

	groupSelector, err := metav1.LabelSelectorAsSelector(nodeGroupAutoscaler.Spec.NodeLabelsSelector)
	if err != nil {
		return RequeueWithError(r.Log, "Invalid node group selector", err)
	}
	var specReplicas int32
	var joining bool
	specIDs := v1.NodesToIdMap(cluster.Spec.Nodes)
	for _, node := range cluster.Spec.Nodes {
		if groupSelector.Matches(labels.Set(node.Labels)) {
			specReplicas++
			state := cluster.Status.NodesState[fmt.Sprint(node.Id)].GracefulActionState.State
			if state == "" || state == v1.GracefulUpscaleRequired || state == v1.GracefulUpscaleRunning {
				joining = true
			}
		}
	}

	clusterLabels := labels.Set(nifiutil.LabelsForNifi(cluster.Name))
	pods := &corev1.PodList{}
	if err := r.APIReader.List(ctx, pods, runtimeClient.InNamespace(cluster.Namespace), runtimeClient.MatchingLabels(clusterLabels)); err != nil {
		return RequeueWithError(r.Log, "Failed to list NiFi cluster pods", err)
	}
	var replicas int32
	var draining bool
	var podNodeIDs []string
	for _, pod := range pods.Items {
		id, ok := pod.Labels["nodeId"]
		if !ok {
			continue
		}
		podNodeIDs = append(podNodeIDs, id)
		if groupSelector.Matches(labels.Set(pod.Labels)) {
			replicas++
			nodeID, err := nifiutil.ParseStringToInt32(id)
			if err != nil {
				return RequeueWithError(r.Log, "Invalid NiFi pod nodeId", err)
			}
			if _, present := specIDs[nodeID]; !present {
				draining = true
			}
		}
	}

	target := nodeGroupAutoscaler.Spec.Replicas
	state := v1alpha1.AutoscalerStateOutOfSync
	if target == specReplicas && target == replicas && !draining {
		state = v1alpha1.AutoscalerStateInSync
	}
	requirements, _ := clusterLabels.AsSelector().Requirements()
	nodeGroupAutoscaler.Status = v1alpha1.NifiNodeGroupAutoscalerStatus{
		State:    state,
		Replicas: v1alpha1.ClusterReplicas(replicas),
		Selector: v1alpha1.ClusterReplicaSelector(groupSelector.Add(requirements...).String()),
	}
	if err := r.updateStatus(ctx, nodeGroupAutoscaler, current.Status); err != nil {
		return RequeueWithError(r.Log, "Failed to update node group autoscaler status", err)
	}
	if state != current.Status.State {
		reason := "Synchronizing"
		if state == v1alpha1.AutoscalerStateInSync {
			reason = "Synchronized"
		}
		r.Recorder.Eventf(nodeGroupAutoscaler, corev1.EventTypeNormal, reason,
			"Requested %d replicas, with %d nodes in spec and %d pods", target, specReplicas, replicas)
	}

	clusterPatch := runtimeClient.MergeFromWithOptions(cluster.DeepCopy(), runtimeClient.MergeFromWithOptimisticLock{})
	switch {
	case target > specReplicas && !draining:
		err = r.scaleUp(nodeGroupAutoscaler, cluster, target-specReplicas, podNodeIDs)
	case target < specReplicas && !joining:
		err = r.scaleDown(nodeGroupAutoscaler, cluster, specReplicas-target)
	default:
		return RequeueAfter(util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset))
	}
	if err != nil {
		return RequeueWithError(r.Log, "Failed to scale NiFi cluster", err)
	}
	if err := r.Client.Patch(ctx, cluster, clusterPatch); err != nil {
		return RequeueWithError(r.Log, "Failed to patch NiFi cluster nodes", err)
	}

	return RequeueAfter(util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset))
}

// scaleUp updates the provided cluster.Spec.Nodes list with the appropriate numNodesToAdd according to the autoscaler.Spec.UpscaleStrategy.
func (r *NifiNodeGroupAutoscalerReconciler) scaleUp(autoscaler *v1alpha1.NifiNodeGroupAutoscaler, cluster *v1.NifiCluster, numNodesToAdd int32, podNodeIDs []string) error {
	switch autoscaler.Spec.UpscaleStrategy {
	// Right now Simple is the only option and the default
	case v1alpha1.SimpleClusterUpscaleStrategy:
		fallthrough
	default:
		r.Log.Info(fmt.Sprintf("Using Simple upscale strategy for cluster %s node group %s", cluster.Name, autoscaler.Spec.NodeConfigGroupId))
		simple := &autoscale.SimpleHorizontalUpscaleStrategy{
			PodNodeIDs:              podNodeIDs,
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
