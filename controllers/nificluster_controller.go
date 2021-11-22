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
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/Orange-OpenSource/nifikop/pkg/k8sutil"
	"github.com/Orange-OpenSource/nifikop/pkg/pki"
	"github.com/Orange-OpenSource/nifikop/pkg/resources"
	"github.com/Orange-OpenSource/nifikop/pkg/resources/nifi"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
)

var clusterFinalizer = "nificlusters.nifi.orange.com/finalizer"
var clusterUsersFinalizer = "nificlusters.nifi.orange.com/users"

// NifiClusterReconciler reconciles a NifiCluster object
type NifiClusterReconciler struct {
	client.Client
	DirectClient     client.Reader
	Log              logr.Logger
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
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nificlusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nificlusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nificlusters/finalizers,verbs=update

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
	_ = r.Log.WithValues("nificluster", req.NamespacedName)

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

	// Check if marked for deletion and run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance)
	}

	if instance.IsExternal() {
		return reconcile.Result{
			RequeueAfter: time.Duration(15) * time.Second,
		}, nil
	}
	//
	if len(instance.Status.State) == 0 || instance.Status.State == v1alpha1.NifiClusterInitializing {
		if err := k8sutil.UpdateCRStatus(r.Client, instance, v1alpha1.NifiClusterInitializing, r.Log); err != nil {
			return RequeueWithError(r.Log, err.Error(), err)
		}
		for nId := range instance.Spec.Nodes {
			if err := k8sutil.UpdateNodeStatus(r.Client, []string{fmt.Sprint(instance.Spec.Nodes[nId].Id)}, instance, v1alpha1.IsInitClusterNode, r.Log); err != nil {
				return RequeueWithError(r.Log, err.Error(), err)
			}
		}
		if err := k8sutil.UpdateCRStatus(r.Client, instance, v1alpha1.NifiClusterInitialized, r.Log); err != nil {
			return RequeueWithError(r.Log, err.Error(), err)
		}
	}

	if instance.Status.State != v1alpha1.NifiClusterRollingUpgrading {
		if err := k8sutil.UpdateCRStatus(r.Client, instance, v1alpha1.NifiClusterReconciling, r.Log); err != nil {
			return RequeueWithError(r.Log, err.Error(), err)
		}
	}

	reconcilers := []resources.ComponentReconciler{
		nifi.New(r.Client, r.DirectClient, r.Scheme, instance),
	}

	intervalNotReady := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_NOT_READY_REQUEUE_INTERVAL"], r.RequeueOffset)
	intervalRunning := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_RUNNING_REQUEUE_INTERVAL"], r.RequeueOffset)
	for _, rec := range reconcilers {
		err = rec.Reconcile(r.Log)
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NodesUnreachable:
				r.Log.Info("Nodes unreachable, may still be starting up")
				return reconcile.Result{
					RequeueAfter: intervalNotReady,
				}, nil
			case errorfactory.NodesNotReady:
				r.Log.Info("Nodes not ready, may still be starting up")
				return reconcile.Result{
					RequeueAfter: intervalNotReady,
				}, nil
			case errorfactory.ResourceNotReady:
				r.Log.Info("A new resource was not found or may not be ready")
				r.Log.Info(err.Error())
				return reconcile.Result{
					RequeueAfter: intervalNotReady / 2,
				}, nil
			case errorfactory.ReconcileRollingUpgrade:
				r.Log.Info("Rolling Upgrade in Progress")
				return reconcile.Result{
					RequeueAfter: intervalRunning,
				}, nil
			case errorfactory.NifiClusterNotReady:
				return reconcile.Result{
					RequeueAfter: intervalNotReady,
				}, nil
			case errorfactory.NifiClusterTaskRunning:
				return reconcile.Result{
					RequeueAfter: intervalRunning,
				}, nil
			default:
				return RequeueWithError(r.Log, err.Error(), err)
			}
		}
	}

	r.Log.Info("ensuring finalizers on nificluster")
	if instance, err = r.ensureFinalizers(ctx, instance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure finalizers on nificluster instance", err)
	}

	//Update rolling upgrade last successful state
	if instance.Status.State == v1alpha1.NifiClusterRollingUpgrading {
		if err := k8sutil.UpdateRollingUpgradeState(r.Client, instance, time.Now(), r.Log); err != nil {
			return RequeueWithError(r.Log, err.Error(), err)
		}
	}

	if err := k8sutil.UpdateCRStatus(r.Client, instance, v1alpha1.NifiClusterRunning, r.Log); err != nil {
		return RequeueWithError(r.Log, err.Error(), err)
	}

	return Reconciled()
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiCluster{}).
		Owns(&policyv1beta1.PodDisruptionBudget{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}

func (r *NifiClusterReconciler) checkFinalizers(ctx context.Context,
	cluster *v1alpha1.NifiCluster) (reconcile.Result, error) {

	r.Log.Info("NifiCluster is marked for deletion, checking for children")

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
			return RequeueWithError(r.Log, "failed to get namespace list", err)
		}
		for _, ns := range namespaceList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	} else {
		// use configured namespaces
		namespaces = r.Namespaces
	}

	if cluster.IsInternal() && cluster.Spec.ListenersConfig.SSLSecrets != nil {
		// If we haven't deleted all nifiusers yet, iterate namespaces and delete all nifiusers
		// with the matching label.
		if util.StringSliceContains(cluster.GetFinalizers(), clusterUsersFinalizer) {
			r.Log.Info(fmt.Sprintf("Sending delete nifiusers request to all namespaces for cluster %s/%s", cluster.Namespace, cluster.Name))
			for _, ns := range namespaces {
				if err := r.Client.DeleteAllOf(
					ctx,
					&v1alpha1.NifiUser{},
					client.InNamespace(ns),
					client.MatchingLabels{ClusterRefLabel: ClusterLabelString(cluster)},
				); err != nil {
					if client.IgnoreNotFound(err) != nil {
						return RequeueWithError(r.Log, "failed to send delete request for children nifiusers", err)
					}
					r.Log.Info(fmt.Sprintf("No matching nifiusers in namespace: %s", ns))
				}
			}
			if cluster, err = r.removeFinalizer(ctx, cluster, clusterUsersFinalizer); err != nil {
				return RequeueWithError(r.Log, "failed to remove users finalizer from nificluster", err)
			}
		}

		// Do any necessary PKI cleanup - a PKI backend should make sure any
		// user finalizations are done before it does its final cleanup
		interval := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_NOT_READY_REQUEUE_INTERVAL"]/3, r.RequeueOffset)
		r.Log.Info("Tearing down any PKI resources for the nificluster")
		if err = pki.GetPKIManager(r.Client, cluster).FinalizePKI(ctx, r.Log); err != nil {
			switch err.(type) {
			case errorfactory.ResourceNotReady:
				r.Log.Info("The PKI is not ready to be torn down")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: interval,
				}, nil
			default:
				return RequeueWithError(r.Log, "failed to finalize PKI", err)
			}
		}

	}

	r.Log.Info("Finalizing deletion of nificluster instance")
	if _, err = r.removeFinalizer(ctx, cluster, clusterFinalizer); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// We may have been a requeue from earlier with all conditions met - but with
			// the state of the finalizer not yet reflected in the response we got.
			return Reconciled()
		}
		return RequeueWithError(r.Log, "failed to remove main finalizer", err)
	}

	return reconcile.Result{}, nil
}

func (r *NifiClusterReconciler) removeFinalizer(ctx context.Context, cluster *v1alpha1.NifiCluster,
	finalizer string) (updated *v1alpha1.NifiCluster, err error) {

	cluster.SetFinalizers(util.StringSliceRemove(cluster.GetFinalizers(), finalizer))
	return r.updateAndFetchLatest(ctx, cluster)
}

func (r *NifiClusterReconciler) updateAndFetchLatest(ctx context.Context,
	cluster *v1alpha1.NifiCluster) (*v1alpha1.NifiCluster, error) {

	typeMeta := cluster.TypeMeta
	err := r.Client.Update(ctx, cluster)
	if err != nil {
		return nil, err
	}
	cluster.TypeMeta = typeMeta
	return cluster, nil
}

func (r *NifiClusterReconciler) ensureFinalizers(ctx context.Context,
	cluster *v1alpha1.NifiCluster) (updated *v1alpha1.NifiCluster, err error) {

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
	return r.updateAndFetchLatest(ctx, cluster)
}
