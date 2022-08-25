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
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/metrics"
	"github.com/konpyutaika/nifikop/pkg/pki"
	"github.com/konpyutaika/nifikop/pkg/resources"
	"github.com/konpyutaika/nifikop/pkg/resources/nifi"
	"github.com/konpyutaika/nifikop/pkg/util"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
)

var clusterFinalizer = "nificlusters.nifi.konpyutaika.com/finalizer"
var clusterUsersFinalizer = "nificlusters.nifi.konpyutaika.com/users"

// NifiClusterReconciler reconciles a NifiCluster object
type NifiClusterReconciler struct {
	InstrumentedReconciler
	client.Client
	zap.Logger
	*metrics.MetricRegistry
	DirectClient     client.Reader
	Scheme           *runtime.Scheme
	Namespaces       []string
	Recorder         record.EventRecorder
	RequeueIntervals map[string]int
	RequeueOffset    int
}

// Metrics implements InstrumentedReconciler interface
func (r *NifiClusterReconciler) Metrics() *metrics.MetricRegistry {
	return r.MetricRegistry
}

// Log implements InstrumentedReconciler interface
func (r *NifiClusterReconciler) Log() *zap.Logger {
	return &r.Logger
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
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificlusters/finalizers,verbs=update

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
	startTime := time.Now()
	res, err := r.doReconcile(ctx, req)
	r.Metrics().ReconcileDurationHistogram().Observe(time.Since(startTime).Seconds())
	return res, err
}

func (r *NifiClusterReconciler) doReconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	intervalNotReady := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_NOT_READY_REQUEUE_INTERVAL"], r.RequeueOffset)
	intervalRunning := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_RUNNING_REQUEUE_INTERVAL"], r.RequeueOffset)

	// Fetch the NifiCluster instance
	instance := &v1alpha1.NifiCluster{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return Reconciled(r)
		}
		// Error reading the object - requeue the request.
		return RequeueWithError(r, err.Error(), err)
	}
	current := instance.DeepCopy()

	// Check if marked for deletion and run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance)
	}

	if instance.IsExternal() {
		return RequeueAfter(r, intervalRunning)
	}

	if len(instance.Status.State) == 0 || instance.Status.State == v1alpha1.NifiClusterInitializing {
		if err := k8sutil.UpdateCRStatus(r.Client, instance, v1alpha1.NifiClusterInitializing, r.Logger); err != nil {
			return RequeueWithError(r, err.Error(), err)
		}
		for nId := range instance.Spec.Nodes {
			if err := k8sutil.UpdateNodeStatus(r.Client, []string{fmt.Sprint(instance.Spec.Nodes[nId].Id)}, instance, v1alpha1.IsInitClusterNode, r.Logger); err != nil {
				return RequeueWithError(r, err.Error(), err)
			}
		}
		if err := k8sutil.UpdateCRStatus(r.Client, instance, v1alpha1.NifiClusterInitialized, r.Logger); err != nil {
			return RequeueWithError(r, err.Error(), err)
		}
	}

	if instance.Status.State != v1alpha1.NifiClusterRollingUpgrading {
		r.Logger.Info("NifiCluster starting reconciliation", zap.String("clusterName", instance.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, string(v1alpha1.NifiClusterReconciling),
			"NifiCluster starting reconciliation")
	}

	reconcilers := []resources.ComponentReconciler{
		nifi.New(r.Client, r.DirectClient, r.Scheme, instance),
	}

	for _, rec := range reconcilers {
		err = rec.Reconcile(r.Logger)
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NodesUnreachable:
				r.Logger.Info("Nodes unreachable, may still be starting up", zap.String("reason", err.Error()))
				return RequeueAfter(r, intervalNotReady)
			case errorfactory.NodesNotReady:
				r.Logger.Info("Nodes not ready, may still be starting up", zap.String("reason", err.Error()))
				return RequeueAfter(r, intervalNotReady)
			case errorfactory.ResourceNotReady:
				r.Logger.Info("A new resource was not found or may not be ready", zap.String("reason", err.Error()))
				return RequeueAfter(r, intervalNotReady/2)
			case errorfactory.ReconcileRollingUpgrade:
				r.Logger.Info("Rolling Upgrade in Progress", zap.String("reason", err.Error()))
				return RequeueAfter(r, intervalRunning)
			case errorfactory.NifiClusterNotReady:
				return RequeueAfter(r, intervalNotReady)
			case errorfactory.NifiClusterTaskRunning:
				return RequeueAfter(r, intervalRunning)
			default:
				return RequeueWithError(r, err.Error(), err)
			}
		}
	}

	r.Logger.Debug("ensuring finalizers on nificluster", zap.String("clusterName", instance.Name))
	if instance, err = r.ensureFinalizers(ctx, instance); err != nil {
		return RequeueWithError(r, "failed to ensure finalizers on nificluster instance "+current.Name, err)
	}

	//Update rolling upgrade last successful state
	if instance.Status.State == v1alpha1.NifiClusterRollingUpgrading {
		if err := k8sutil.UpdateRollingUpgradeState(r.Client, instance, time.Now(), r.Logger); err != nil {
			return RequeueWithError(r, err.Error(), err)
		}
	}

	if !instance.IsReady() {
		r.Logger.Info("Successfully reconciled NifiCluster", zap.String("clusterName", instance.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, string(v1alpha1.NifiClusterRunning),
			"Successfully reconciled NifiCluster")
		if err := k8sutil.UpdateCRStatus(r.Client, instance, v1alpha1.NifiClusterRunning, r.Logger); err != nil {
			return RequeueWithError(r, err.Error(), err)
		}
	}

	return RequeueAfter(r, intervalRunning)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logCtr, err := GetLogConstructor(mgr, &v1alpha1.NifiCluster{})
	if err != nil {
		return err
	}
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiCluster{}).
		WithLogConstructor(logCtr).
		Owns(&policyv1beta1.PodDisruptionBudget{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{})

	if util.IsK8sPrior1_21() {
		builder.Owns(&policyv1beta1.PodDisruptionBudget{})
	} else {
		builder.Owns(&policyv1.PodDisruptionBudget{})
	}

	return builder.Complete(r)
}

func (r *NifiClusterReconciler) checkFinalizers(ctx context.Context,
	cluster *v1alpha1.NifiCluster) (reconcile.Result, error) {

	r.Logger.Info("NifiCluster is marked for deletion, checking for children", zap.String("clusterName", cluster.Name))

	// If the main finalizer is gone then we've already finished up
	if !util.StringSliceContains(cluster.GetFinalizers(), clusterFinalizer) {
		return Reconciled(r)
	}

	var err error

	var namespaces []string
	if r.Namespaces == nil || len(r.Namespaces) == 0 {
		// Fetch a list of all namespaces for DeleteAllOf requests
		namespaces = make([]string, 0)
		var namespaceList corev1.NamespaceList
		if err := r.Client.List(ctx, &namespaceList); err != nil {
			return RequeueWithError(r, "failed to get namespace list from k8s api", err)
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
			r.Logger.Info("Sending delete nifiusers request to all namespaces for cluster",
				zap.String("namespace", cluster.Namespace),
				zap.String("clusterName", cluster.Name))
			for _, ns := range namespaces {
				if err := r.Client.DeleteAllOf(
					ctx,
					&v1alpha1.NifiUser{},
					client.InNamespace(ns),
					client.MatchingLabels{ClusterRefLabel: ClusterLabelString(cluster)},
				); err != nil {
					if client.IgnoreNotFound(err) != nil {
						return RequeueWithError(r, "failed to send delete request for children nifiusers in namespace "+ns, err)
					}
					r.Logger.Info("No matching nifiusers in namespace", zap.String("namespace", ns))
				}
			}
			if cluster, err = r.removeFinalizer(ctx, cluster, clusterUsersFinalizer); err != nil {
				return RequeueWithError(r, "failed to remove users finalizer from nificluster "+cluster.Name, err)
			}
		}

		// Do any necessary PKI cleanup - a PKI backend should make sure any
		// user finalizations are done before it does its final cleanup
		interval := util.GetRequeueInterval(r.RequeueIntervals["CLUSTER_TASK_NOT_READY_REQUEUE_INTERVAL"]/3, r.RequeueOffset)
		r.Logger.Info("Tearing down any PKI resources for the nificluster",
			zap.String("clusterName", cluster.Name))
		if err = pki.GetPKIManager(r.Client, cluster).FinalizePKI(ctx, r.Logger); err != nil {
			switch err.(type) {
			case errorfactory.ResourceNotReady:
				r.Logger.Warn("The PKI is not ready to be torn down", zap.Error(err))
				return RequeueAfter(r, interval)
			default:
				return RequeueWithError(r, "failed to finalize PKI", err)
			}
		}
	}

	r.Logger.Info("Finalizing deletion of nificluster instance", zap.String("clusterName", cluster.Name))
	if _, err = r.removeFinalizer(ctx, cluster, clusterFinalizer); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// We may have been a requeue from earlier with all conditions met - but with
			// the state of the finalizer not yet reflected in the response we got.
			return Reconciled(r)
		}
		return RequeueWithError(r, "failed to remove main finalizer from NifiCluser "+cluster.Name, err)
	}

	return Reconciled(r)
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
