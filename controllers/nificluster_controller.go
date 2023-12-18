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
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/pki"
	"github.com/konpyutaika/nifikop/pkg/resources"
	"github.com/konpyutaika/nifikop/pkg/resources/nifi"
	"github.com/konpyutaika/nifikop/pkg/util"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

var clusterFinalizer string = fmt.Sprintf("nificlusters.%s/finalizer", v1.GroupVersion.Group)
var clusterUsersFinalizer string = fmt.Sprintf("nificlusters.%s/users", v1.GroupVersion.Group)

// NifiClusterReconciler reconciles a NifiCluster object.
type NifiClusterReconciler struct {
	client.Client
	DirectClient     client.Reader
	Log              zap.Logger
	Scheme           *runtime.Scheme
	Namespaces       []string
	Recorder         record.EventRecorder
	RequeueIntervals map[string]int
	RequeueOffset    int
}

// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;delete;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="policy",resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificlusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificlusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificlusters/finalizers,verbs=get;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *NifiClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
	current := instance.DeepCopy()
	patchInstance := client.MergeFromWithOptions(instance.DeepCopy(), client.MergeFromWithOptimisticLock{})

	// Check if marked for deletion and run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, patchInstance)
	}

	if instance.IsExternal() {
		return RequeueAfter(time.Duration(15) * time.Second)
	}

	if len(instance.Spec.Nodes) == 0 && len(instance.Status.NodesState) == 0 {
		intervalNoNode := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_NO_NODE_INTERVAL"], r.RequeueOffset)
		r.Recorder.Event(instance, corev1.EventTypeNormal, string(v1.NifiClusterNoNodes),
			"NifiCluster has no node, nothing to do")
		if err := k8sutil.UpdateCRStatus(r.Client, instance, current.Status, v1.NifiClusterNoNodes, r.Log); err != nil {
			return RequeueWithError(r.Log, err.Error(), err)
		}
		return RequeueAfter(intervalNoNode)
	}

	//
	if len(instance.Spec.Nodes) > 0 && (len(instance.Status.State) == 0 || instance.Status.State == v1.NifiClusterInitializing || instance.Status.State == v1.NifiClusterNoNodes) {
		if err := k8sutil.UpdateCRStatus(r.Client, instance, current.Status, v1.NifiClusterInitializing, r.Log); err != nil {
			return RequeueWithError(r.Log, err.Error(), err)
		}
		for nId := range instance.Spec.Nodes {
			if err := k8sutil.UpdateNodeStatus(r.Client, []string{fmt.Sprint(instance.Spec.Nodes[nId].Id)}, instance, current.Status, v1.IsInitClusterNode, r.Log); err != nil {
				return RequeueWithError(r.Log, err.Error(), err)
			}
		}
		if err := k8sutil.UpdateCRStatus(r.Client, instance, current.Status, v1.NifiClusterInitialized, r.Log); err != nil {
			return RequeueWithError(r.Log, err.Error(), err)
		}
	}

	if instance.Status.State != v1.NifiClusterRollingUpgrading {
		r.Log.Info("NifiCluster starting reconciliation", zap.String("clusterName", instance.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, string(v1.NifiClusterReconciling),
			"NifiCluster starting reconciliation")
	}

	reconcilers := []resources.ComponentReconciler{
		nifi.New(r.Client, r.DirectClient, r.Scheme, instance, current.Status),
	}

	intervalNotReady := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_NOT_READY_REQUEUE_INTERVAL"], r.RequeueOffset)
	intervalRunning := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_RUNNING_REQUEUE_INTERVAL"], r.RequeueOffset)
	for _, rec := range reconcilers {
		err = rec.Reconcile(r.Log)
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NodesUnreachable:
				r.Log.Info("Nodes unreachable, may still be starting up", zap.String("reason", err.Error()))
				return RequeueAfter(intervalNotReady)
			case errorfactory.NodesNotReady:
				r.Log.Info("Nodes not ready, may still be starting up", zap.String("reason", err.Error()))
				return RequeueAfter(intervalNotReady)
			case errorfactory.ResourceNotReady:
				r.Log.Info("A new resource was not found or may not be ready", zap.String("reason", err.Error()))
				return RequeueAfter(intervalNotReady)
			case errorfactory.ReconcileRollingUpgrade:
				r.Log.Info("Rolling Upgrade in Progress", zap.String("reason", err.Error()))
				return RequeueAfter(intervalRunning)
			case errorfactory.NifiClusterNotReady:
				return RequeueAfter(intervalNotReady)
			case errorfactory.NifiClusterTaskRunning:
				return RequeueAfter(intervalRunning)
			default:
				return RequeueWithError(r.Log, err.Error(), err)
			}
		}
	}

	r.Log.Debug("ensuring finalizers on nificluster", zap.String("clusterName", instance.Name))
	if instance, err = r.ensureFinalizers(ctx, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure finalizers on nificluster instance "+current.Name, err)
	}

	// Update rolling upgrade last successful state
	if instance.Status.State == v1.NifiClusterRollingUpgrading {
		if err := k8sutil.UpdateRollingUpgradeState(r.Client, instance, current.Status, time.Now(), r.Log); err != nil {
			return RequeueWithError(r.Log, err.Error(), err)
		}
	}

	if !instance.IsReady() {
		r.Log.Info("Successfully reconciled NifiCluster", zap.String("clusterName", instance.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, string(v1.NifiClusterRunning),
			"Successfully reconciled NifiCluster")
		if err := k8sutil.UpdateCRStatus(r.Client, instance, current.Status, v1.NifiClusterRunning, r.Log); err != nil {
			return RequeueWithError(r.Log, err.Error(), err)
		}
	}

	return RequeueAfter(intervalRunning)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logCtr, err := GetLogConstructor(mgr, &v1.NifiCluster{})
	if err != nil {
		return err
	}
	if util.IsK8sPrior1_21() {
		return ctrl.NewControllerManagedBy(mgr).
			For(&v1.NifiCluster{}).
			WithLogConstructor(logCtr).
			Owns(&policyv1beta1.PodDisruptionBudget{}).
			Owns(&corev1.Service{}).
			Owns(&corev1.Pod{}).
			Owns(&corev1.ConfigMap{}).
			Owns(&corev1.PersistentVolumeClaim{}).
			Complete(r)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.NifiCluster{}).
		WithLogConstructor(logCtr).
		Owns(&policyv1.PodDisruptionBudget{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}

func (r *NifiClusterReconciler) checkFinalizers(ctx context.Context,
	cluster *v1.NifiCluster, patcher client.Patch) (reconcile.Result, error) {
	r.Log.Info("NifiCluster is marked for deletion, checking for children", zap.String("clusterName", cluster.Name))

	// If the main finalizer is gone then we've already finished up
	if !util.StringSliceContains(cluster.GetFinalizers(), clusterFinalizer) {
		return Reconciled()
	}

	var err error

	var namespaces []string
	if r.Namespaces == nil || len(r.Namespaces) == 0 {
		// Fetch a list of all namespaces for DeleteAllOf requests
		namespaces = make([]string, 0)
		var namespaceList corev1.NamespaceList
		if err := r.Client.List(ctx, &namespaceList); err != nil {
			return RequeueWithError(r.Log, "failed to get namespace list from k8s api", err)
		}
		for _, ns := range namespaceList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	} else {
		// use configured namespaces
		namespaces = r.Namespaces
	}

	if result, err := r.finalizePVCs(ctx, cluster); err != nil {
		return result, err
	}

	if cluster.IsInternal() && cluster.Spec.ListenersConfig.SSLSecrets != nil {
		// If we haven't deleted all nifiusers yet, iterate namespaces and delete all nifiusers
		// with the matching label.
		if util.StringSliceContains(cluster.GetFinalizers(), clusterUsersFinalizer) {
			r.Log.Info("Sending delete nifiusers request to all namespaces for cluster",
				zap.String("namespace", cluster.Namespace),
				zap.String("clusterName", cluster.Name))
			for _, ns := range namespaces {
				if err := r.Client.DeleteAllOf(
					ctx,
					&v1.NifiUser{},
					client.InNamespace(ns),
					client.MatchingLabels{ClusterRefLabel: ClusterLabelString(cluster)},
				); err != nil {
					if client.IgnoreNotFound(err) != nil {
						return RequeueWithError(r.Log, "failed to send delete request for children nifiusers in namespace "+ns, err)
					}
					r.Log.Info("No matching nifiusers in namespace", zap.String("namespace", ns))
				}
			}
			if cluster, err = r.removeFinalizer(ctx, cluster, clusterUsersFinalizer, patcher); err != nil {
				return RequeueWithError(r.Log, "failed to remove users finalizer from nificluster "+cluster.Name, err)
			}
		}

		// Do any necessary PKI cleanup - a PKI backend should make sure any
		// user finalizations are done before it does its final cleanup
		interval := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_NOT_READY_REQUEUE_INTERVAL"], r.RequeueOffset)
		r.Log.Info("Tearing down any PKI resources for the nificluster",
			zap.String("clusterName", cluster.Name))
		if err = pki.GetPKIManager(r.Client, cluster).FinalizePKI(ctx, r.Log); err != nil {
			switch err.(type) {
			case errorfactory.ResourceNotReady:
				r.Log.Warn("The PKI is not ready to be torn down", zap.Error(err))
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: interval,
				}, nil
			default:
				return RequeueWithError(r.Log, "failed to finalize PKI", err)
			}
		}
	}

	r.Log.Info("Finalizing deletion of nificluster instance", zap.String("clusterName", cluster.Name))
	if _, err = r.removeFinalizer(ctx, cluster, clusterFinalizer, patcher); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// We may have been a requeue from earlier with all conditions met - but with
			// the state of the finalizer not yet reflected in the response we got.
			return Reconciled()
		}
		return RequeueWithError(r.Log, "failed to remove main finalizer from NifiCluser "+cluster.Name, err)
	}

	return Reconciled()
}

func (r *NifiClusterReconciler) finalizePVCs(ctx context.Context, cluster *v1.NifiCluster) (reconcile.Result, error) {
	// remove PVC owner references if they're configured to be retained. This will prevent the PVC from getting garbage collected.
	foundPVCList := &corev1.PersistentVolumeClaimList{}
	matchingLabels := client.MatchingLabels{
		"nifi_cr":                           cluster.Name,
		nifiutil.NifiDataVolumeMountKey:     "true",
		nifiutil.NifiVolumeReclaimPolicyKey: string(corev1.PersistentVolumeReclaimRetain),
	}
	if err := r.Client.List(ctx, foundPVCList, client.ListOption(client.InNamespace(cluster.Namespace)), client.ListOption(matchingLabels)); err != nil {
		return RequeueWithError(r.Log, "failed to get PVC list from k8s api", err)
	}
	for _, pvc := range foundPVCList.Items {
		r.Log.Debug("Removing owner references for PVC as it is configured to be retained.", zap.String("clusterName", cluster.Name), zap.String("pvcName", pvc.Name))
		patch := client.MergeFrom(pvc.DeepCopy())
		// clear owner refs so that the PVC does not get deleted when the NifiCluster is deleted.
		pvc.SetOwnerReferences(nil)
		if err := r.Client.Patch(ctx, &pvc, patch); err != nil {
			return RequeueWithError(r.Log, "failed to delete owner references for PVC "+pvc.Name, err)
		}
	}
	return Reconciled()
}

func (r *NifiClusterReconciler) removeFinalizer(ctx context.Context, cluster *v1.NifiCluster,
	finalizer string, patcher client.Patch) (updated *v1.NifiCluster, err error) {
	cluster.SetFinalizers(util.StringSliceRemove(cluster.GetFinalizers(), finalizer))
	return r.updateAndFetchLatest(ctx, cluster, patcher)
}

func (r *NifiClusterReconciler) updateAndFetchLatest(ctx context.Context,
	cluster *v1.NifiCluster, patcher client.Patch) (*v1.NifiCluster, error) {
	typeMeta := cluster.TypeMeta
	err := r.Client.Patch(ctx, cluster, patcher)
	if err != nil {
		return nil, err
	}
	cluster.TypeMeta = typeMeta
	return cluster, nil
}

func (r *NifiClusterReconciler) ensureFinalizers(ctx context.Context,
	cluster *v1.NifiCluster, patcher client.Patch) (updated *v1.NifiCluster, err error) {
	finalizers := []string{clusterFinalizer}
	if cluster.IsInternal() && cluster.Spec.ListenersConfig.SSLSecrets != nil {
		finalizers = append(finalizers, clusterUsersFinalizer)
	}
	for _, finalizer := range finalizers {
		if util.StringSliceContains(cluster.GetFinalizers(), finalizer) {
			continue
		}
		cluster.SetFinalizers(append(cluster.GetFinalizers(), finalizer))
	}
	return r.updateAndFetchLatest(ctx, cluster, patcher)
}
