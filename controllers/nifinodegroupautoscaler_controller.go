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
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/autoscale"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var autoscalerFinalizer = "nifinodegroupautoscalers.nifi.konpyutaika.com/finalizer"

// NifiNodeGroupAutoscalerReconciler reconciles a NifiNodeGroupAutoscaler object
type NifiNodeGroupAutoscalerReconciler struct {
	runtimeClient.Client
	APIReader       runtimeClient.Reader
	Scheme          *runtime.Scheme
	Log             logr.Logger
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifinodegroupautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifinodegroupautoscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifinodegroupautoscalers/finalizers,verbs=update
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificlusters,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

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
	_ = r.Log.WithValues("nifinodegroupautoscaler", req.NamespacedName)

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

	// Check if marked for deletion and run finalizers
	if k8sutil.IsMarkedForDeletion(nodeGroupAutoscaler.ObjectMeta) {
		return r.checkFinalizers(ctx, nodeGroupAutoscaler)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(nodeGroupAutoscaler.GetFinalizers(), autoscalerFinalizer) {
		r.Log.Info(fmt.Sprintf("Adding Finalizer for NifiNodeGroupAutoscaler node group %s", nodeGroupAutoscaler.Spec.NodeConfigGroupId))
		nodeGroupAutoscaler.SetFinalizers(append(nodeGroupAutoscaler.GetFinalizers(), autoscalerFinalizer))
	}

	// Get the last configuration viewed by the operator.
	o, err := patch.DefaultAnnotator.GetOriginalConfiguration(nodeGroupAutoscaler)
	// Create it if not exist.
	if o == nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(nodeGroupAutoscaler); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation", err)
		}
		if err := r.Client.Update(ctx, nodeGroupAutoscaler); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiNodeGroupAutoscaler", err)
		}
		o, err = patch.DefaultAnnotator.GetOriginalConfiguration(nodeGroupAutoscaler)
	}

	// Check if the cluster reference changed.
	original := &v1alpha1.NifiNodeGroupAutoscaler{}
	json.Unmarshal(o, original)
	if !v1alpha1.ClusterRefsEquals([]v1alpha1.ClusterReference{original.Spec.ClusterRef, nodeGroupAutoscaler.Spec.ClusterRef}) {
		nodeGroupAutoscaler.Spec.ClusterRef = original.Spec.ClusterRef
	}

	// lookup NifiCluster reference
	// we do not want cached objects here. We want an accurate state of what the cluster is right now, so bypass the client cache by using the APIReader directly.
	cluster := &v1alpha1.NifiCluster{}
	err = r.APIReader.Get(ctx,
		types.NamespacedName{
			Name:      nodeGroupAutoscaler.Spec.ClusterRef.Name,
			Namespace: nodeGroupAutoscaler.Spec.ClusterRef.Namespace,
		},
		cluster)
	if err != nil {
		return RequeueWithError(r.Log, fmt.Sprintf("failed to look up cluster reference %v+", nodeGroupAutoscaler.Spec.ClusterRef), err)
	}

	// Handle HorizontalAutoScaler
	hpa, err := r.horizontalPodAutoscaler(r.Log, cluster, nodeGroupAutoscaler)
	if err != nil {
		return RequeueWithError(r.Log, "failed to generate horizontal pod autoscaler spec", err)
	}
	err = k8sutil.Reconcile(r.Log, r.Client, hpa, nil)
	if err != nil {
		return RequeueWithError(r.Log, "failed to reconcile horizontal pod autoscaler", err)
	}

	// Determine how many replicas there currently are and how many are desired for the appropriate node group
	numDesiredReplicas := nodeGroupAutoscaler.Spec.Replicas
	currentReplicas, maxId, err := r.getManagedNodes(nodeGroupAutoscaler, cluster.Spec.Nodes)
	if err != nil {
		return RequeueWithError(r.Log, "Failed to apply autoscaler node selector to cluster nodes", err)
	}
	numCurrentReplicas := int32(len(currentReplicas))

	// if the current number of nodes being managed by this autoscaler is different than the replica setting,
	// then set the autoscaler status to out of sync to indicate we're changing the NifiCluster node config
	if numDesiredReplicas != numCurrentReplicas {
		r.Log.Info(fmt.Sprintf("Replicas changed from %d to %d", numCurrentReplicas, numDesiredReplicas))
		if err = r.updateAutoscalerReplicaState(ctx, nodeGroupAutoscaler, v1alpha1.AutoscalerStateOutOfSync); err != nil {
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

			startingNodeId := maxId + 1
			if err = r.scaleUp(nodeGroupAutoscaler, cluster, numNodesToAdd, startingNodeId); err != nil {
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
		if err = r.updateAutoscalerReplicaState(ctx, nodeGroupAutoscaler, v1alpha1.AutoscalerStateInSync); err != nil {
			return RequeueWithError(r.Log, fmt.Sprintf("Failed to udpate node group autoscaler state for node group %s", nodeGroupAutoscaler.Spec.NodeConfigGroupId), err)
		}
	} else {
		r.Log.V(5).Info("Cluster replicas config and current number of replicas are the same", "replicas", nodeGroupAutoscaler.Spec.Replicas)
	}
	//TODO ensure finalizers on NifiNodeGroupAutoscaler

	// update replica and replica status
	if err = r.updateAutoscalerReplicaStatus(ctx, cluster, nodeGroupAutoscaler); err != nil {
		return RequeueWithError(r.Log, fmt.Sprintf("Failed to update node group autoscaler replica status for node group %s", nodeGroupAutoscaler.Spec.NodeConfigGroupId), err)
	}

	return reconcile.Result{
		RequeueAfter: util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset),
	}, nil
}

// scaleUp updates the provided cluster.Spec.Nodes list with the appropriate numNodesToAdd according to the autoscaler.Spec.UpscaleStrategy
func (r *NifiNodeGroupAutoscalerReconciler) scaleUp(autoscaler *v1alpha1.NifiNodeGroupAutoscaler, cluster *v1alpha1.NifiCluster, numNodesToAdd int32, startingNodeId int32) error {
	managedNodes, _, err := r.getManagedNodes(autoscaler, cluster.Spec.Nodes)
	if err != nil {
		return errors.WrapIff(err, "Failed to fetch managed nodes for node group %s", autoscaler.Spec.NodeConfigGroupId)
	}

	switch autoscaler.Spec.UpscaleStrategy {
	// Right now Simple is the only option and the default
	case v1alpha1.SimpleClusterUpscaleStrategy:
		fallthrough
	default:
		r.Log.Info(fmt.Sprintf("Using Simple upscale strategy for cluster %s node group %s", cluster.Name, autoscaler.Spec.NodeConfigGroupId))
		simple := &autoscale.SimpleHorizontalUpscaleStrategy{
			NifiNodeGroupAutoscaler: autoscaler,
			MaxNodeId:               startingNodeId,
		}
		nodesToAdd, err := simple.ScaleUp(managedNodes, numNodesToAdd)
		if err != nil {
			return errors.WrapIf(err, "Failed to scale up using the Simple strategy.")
		}
		cluster.Spec.Nodes = append(cluster.Spec.Nodes, nodesToAdd...)
	}
	r.Recorder.Eventf(autoscaler, corev1.EventTypeNormal, "Upscaling",
		"Adding %d more nodes to cluster %s spec.nodes configuration for node group %s", numNodesToAdd, cluster.Name, autoscaler.Spec.NodeConfigGroupId)

	return nil
}

// scaleUp updates the provided cluster.Spec.Nodes list with the appropriate numNodesToRemove according to the autoscaler.Spec.DownscaleStrategy
func (r *NifiNodeGroupAutoscalerReconciler) scaleDown(autoscaler *v1alpha1.NifiNodeGroupAutoscaler, cluster *v1alpha1.NifiCluster, numNodesToRemove int32) error {
	managedNodes, _, err := r.getManagedNodes(autoscaler, cluster.Spec.Nodes)
	if err != nil {
		return errors.WrapIff(err, "Failed to fetch managed nodes for node group %s", autoscaler.Spec.NodeConfigGroupId)
	}

	switch autoscaler.Spec.DownscaleStrategy {

	// Right now LIFO is the only option and the default
	case v1alpha1.LIFOClusterDownscaleStrategy:
		fallthrough
	default:
		r.Log.Info(fmt.Sprintf("Using LIFO downscale strategy for cluster %s node group %s", cluster.Name, autoscaler.Spec.NodeConfigGroupId))
		// remove the last n nodes from the node list
		lifo := &autoscale.LIFOHorizontalDownscaleStrategy{}
		nodesToRemove, err := lifo.ScaleDown(managedNodes, numNodesToRemove)
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

// updateAutoscalerReplicaState updates the state of the autoscaler
func (r *NifiNodeGroupAutoscalerReconciler) updateAutoscalerReplicaState(ctx context.Context, autoscaler *v1alpha1.NifiNodeGroupAutoscaler, state v1alpha1.NodeGroupAutoscalerState) error {
	autoscaler.Status.State = state
	switch state {
	case v1alpha1.AutoscalerStateInSync:
		r.Recorder.Event(autoscaler, corev1.EventTypeNormal, "Synchronized", "Successfully synchronized node group autoscaler.")
	case v1alpha1.AutoscalerStateOutOfSync:
		r.Recorder.Event(autoscaler, corev1.EventTypeNormal, "Synchronizing", "The number of replicas for this node group has changed. Synchronizing.")
	}
	return r.Client.Status().Update(ctx, autoscaler)
}

// updateAutoscalerReplicaStatus updates autoscaler replica status to inform the k8s scale subresource
func (r *NifiNodeGroupAutoscalerReconciler) updateAutoscalerReplicaStatus(ctx context.Context, nifiCluster *v1alpha1.NifiCluster, autoscaler *v1alpha1.NifiNodeGroupAutoscaler) error {
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

	return r.Client.Status().Update(ctx, autoscaler)
}

// getCurrentReplicaPods searches for any pods created in this node scaler's node group
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

// getManagedNodes filters a set of nodes by an autoscaler's configured node selector
func (r *NifiNodeGroupAutoscalerReconciler) getManagedNodes(autoscaler *v1alpha1.NifiNodeGroupAutoscaler, nodes []v1alpha1.Node) (managedNodes []v1alpha1.Node, maxId int32, err error) {
	selector, err := metav1.LabelSelectorAsSelector(autoscaler.Spec.NodeLabelsSelector)
	if err != nil {
		return nil, maxId, err
	}

	max := 0
	for _, node := range nodes {
		if selector.Matches(labels.Set(node.Labels)) {
			managedNodes = append(managedNodes, node)
		}
		max = util.Max(max, int(node.Id))
	}
	return managedNodes, int32(max), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiNodeGroupAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiNodeGroupAutoscaler{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Complete(r)
}

func (r *NifiNodeGroupAutoscalerReconciler) checkFinalizers(ctx context.Context, autoscaler *v1alpha1.NifiNodeGroupAutoscaler) (reconcile.Result, error) {
	r.Log.Info("NifiNodeGroupAutoscaler is marked for deletion")

	var err error
	if util.StringSliceContains(autoscaler.GetFinalizers(), autoscalerFinalizer) {
		if err = r.finalizeNifiNodeGroupAutoscaler(autoscaler); err != nil {
			return RequeueAfter(util.GetRequeueInterval(r.RequeueInterval/3, r.RequeueOffset))
		}
		if err = r.removeFinalizer(ctx, autoscaler); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from autoscaler", err)
		}
	}

	return Reconciled()
}

func (r *NifiNodeGroupAutoscalerReconciler) removeFinalizer(ctx context.Context, autoscaler *v1alpha1.NifiNodeGroupAutoscaler) error {
	autoscaler.SetFinalizers(util.StringSliceRemove(autoscaler.GetFinalizers(), autoscalerFinalizer))
	_, err := r.updateAndFetchLatest(ctx, autoscaler)
	return err
}

func (r *NifiNodeGroupAutoscalerReconciler) finalizeNifiNodeGroupAutoscaler(autoscaler *v1alpha1.NifiNodeGroupAutoscaler) error {

	//TODO: remove nodes from NifiCluster that this auto scaler is controlling?

	return nil
}

func (r *NifiNodeGroupAutoscalerReconciler) updateAndFetchLatest(ctx context.Context,
	autoscaler *v1alpha1.NifiNodeGroupAutoscaler) (*v1alpha1.NifiNodeGroupAutoscaler, error) {

	typeMeta := autoscaler.TypeMeta
	err := r.Client.Update(ctx, autoscaler)
	if err != nil {
		return nil, err
	}
	autoscaler.TypeMeta = typeMeta
	return autoscaler, nil
}

// Create a HorizontalPodAutoscaler CR
func (r *NifiNodeGroupAutoscalerReconciler) horizontalPodAutoscaler(log logr.Logger, nifiCluster *v1alpha1.NifiCluster, autoscaler *v1alpha1.NifiNodeGroupAutoscaler) (runtimeClient.Object, error) {
	resourceName := fmt.Sprintf("%s-%s-hpa", nifiCluster.Name, autoscaler.Spec.NodeConfigGroupId)
	return &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v2beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        resourceName,
			Namespace:   nifiCluster.Namespace,
			Labels:      util.MergeLabels(nifiutil.LabelsForNifi(nifiCluster.Name), nifiCluster.Labels),
			Annotations: util.MergeAnnotations(nifiCluster.Spec.Service.Annotations, autoscaler.Annotations),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         autoscaler.APIVersion,
					Kind:               autoscaler.Kind,
					Name:               autoscaler.Name,
					UID:                autoscaler.UID,
					Controller:         util.BoolPointer(true),
					BlockOwnerDeletion: util.BoolPointer(true),
				},
			},
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "NifiNodeGroupAutoscaler",
				APIVersion: "nifi.konpyutaika.com/v1alpha1",
				Name:       fmt.Sprintf("%s-%s", nifiCluster.Name, autoscaler.Spec.NodeConfigGroupId),
			},
			MinReplicas: &autoscaler.Spec.HorizontalAutoscaler.MinReplicas,
			MaxReplicas: autoscaler.Spec.HorizontalAutoscaler.MaxReplicas,
			Metrics:     autoscaler.Spec.HorizontalAutoscaler.Metrics,
			Behavior:    autoscaler.Spec.HorizontalAutoscaler.Behavior,
		},
	}, nil
}