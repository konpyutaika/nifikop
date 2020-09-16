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

package nificluster

import (
	"context"
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	v1alpha1 "github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
	common "github.com/Orange-OpenSource/nifikop/pkg/controller/common"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/Orange-OpenSource/nifikop/pkg/k8sutil"
	"github.com/Orange-OpenSource/nifikop/pkg/pki"
	"github.com/Orange-OpenSource/nifikop/pkg/resources"
	"github.com/Orange-OpenSource/nifikop/pkg/resources/nifi"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_nificluster")

var clusterFinalizer = "finalizer.nificlusters.nifi.orange.com"
var clusterUsersFinalizer = "users.nificlusters.nifi.orange.com"

// Add creates a new NifiCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, namespaces []string) error {
	return add(mgr, newReconciler(mgr, namespaces))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, namespaces []string) reconcile.Reconciler {
	return &ReconcileNifiCluster{client: mgr.GetClient(), scheme: mgr.GetScheme(), DirectClient: mgr.GetAPIReader(), Namespaces: namespaces}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nificluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NifiCluster
	err = c.Watch(&source.Kind{Type: &v1alpha1.NifiCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner NifiCluster
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.NifiCluster{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource ConfigMap and requeue the owner NifiCluster
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.NifiCluster{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource PersistentVolumeClaim and requeue the owner NifiCluster
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.NifiCluster{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileNifiCluster implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNifiCluster{}

// ReconcileNifiCluster reconciles a NifiCluster object
type ReconcileNifiCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client       client.Client
	DirectClient client.Reader
	scheme       *runtime.Scheme
	Namespaces   []string
}

// Reconcile reads that state of the cluster for a NifiCluster object and makes changes based on the state read
// and what is in the NifiCluster.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNifiCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NifiCluster")

	ctx := context.Background()

	// Fetch the NifiCluster instance
	instance := &v1alpha1.NifiCluster{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
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

	// Check if marked for deletion and run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, reqLogger, instance)
	}

	//
	if len(instance.Status.State) == 0 || instance.Status.State == v1alpha1.NifiClusterInitializing {
		if err := k8sutil.UpdateCRStatus(r.client, instance, v1alpha1.NifiClusterInitializing, reqLogger); err != nil {
			return common.RequeueWithError(log, err.Error(), err)
		}
		for nId := range instance.Spec.Nodes {
			if err := k8sutil.UpdateNodeStatus(r.client, []string{fmt.Sprint(instance.Spec.Nodes[nId].Id)}, instance, v1alpha1.IsInitClusterNode, log); err != nil {
				return common.RequeueWithError(log, err.Error(), err)
			}
		}
		if err := k8sutil.UpdateCRStatus(r.client, instance, v1alpha1.NifiClusterInitialized, reqLogger); err != nil {
			return common.RequeueWithError(log, err.Error(), err)
		}
	}

	if instance.Status.State != v1alpha1.NifiClusterRollingUpgrading {
		if err := k8sutil.UpdateCRStatus(r.client, instance, v1alpha1.NifiClusterReconciling, reqLogger); err != nil {
			return common.RequeueWithError(log, err.Error(), err)
		}
	}

	reconcilers := []resources.ComponentReconciler{
		nifi.New(r.client, r.DirectClient, r.scheme, instance),
	}

	for _, rec := range reconcilers {
		err = rec.Reconcile(reqLogger)
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.NodesUnreachable:
				reqLogger.Info("Nodes unreachable, may still be starting up")
				return reconcile.Result{
					RequeueAfter: time.Duration(15) * time.Second,
				}, nil
			case errorfactory.NodesNotReady:
				reqLogger.Info("Nodes not ready, may still be starting up")
				return reconcile.Result{
					RequeueAfter: time.Duration(15) * time.Second,
				}, nil
			case errorfactory.ResourceNotReady:
				reqLogger.Info("A new resource was not found or may not be ready")
				reqLogger.Info(err.Error())
				return reconcile.Result{
					RequeueAfter: time.Duration(7) * time.Second,
				}, nil
			case errorfactory.ReconcileRollingUpgrade:
				reqLogger.Info("Rolling Upgrade in Progress")
				return reconcile.Result{
					RequeueAfter: time.Duration(15) * time.Second,
				}, nil
			case errorfactory.NifiClusterNotReady:
				return reconcile.Result{
					RequeueAfter: time.Duration(15) * time.Second,
				}, nil
			case errorfactory.NifiClusterTaskRunning:
				return reconcile.Result{
					RequeueAfter: time.Duration(20) * time.Second,
				}, nil
			default:
				return common.RequeueWithError(reqLogger, err.Error(), err)
			}
		}
	}

	reqLogger.Info("ensuring finalizers on nificluster")
	if instance, err = r.ensureFinalizers(ctx, instance); err != nil {
		return common.RequeueWithError(log, "failed to ensure finalizers on nificluster instance", err)
	}

	//Update rolling upgrade last successful state
	if instance.Status.State == v1alpha1.NifiClusterRollingUpgrading {
		if err := k8sutil.UpdateRollingUpgradeState(r.client, instance, time.Now(), reqLogger); err != nil {
			return common.RequeueWithError(reqLogger, err.Error(), err)
		}
	}

	if err := k8sutil.UpdateCRStatus(r.client, instance, v1alpha1.NifiClusterRunning, reqLogger); err != nil {
		return common.RequeueWithError(log, err.Error(), err)
	}

	return common.Reconciled()
}

func (r *ReconcileNifiCluster) checkFinalizers(ctx context.Context, log logr.Logger, cluster *v1alpha1.NifiCluster) (reconcile.Result, error) {
	log.Info("NifiCluster is marked for deletion, checking for children")

	// If the main finalizer is gone then we've already finished up
	if !util.StringSliceContains(cluster.GetFinalizers(), clusterFinalizer) {
		return common.Reconciled()
	}

	var err error

	var namespaces []string
	if r.Namespaces == nil {
		// Fetch a list of all namespaces for DeleteAllOf requests
		namespaces = make([]string, 0)
		var namespaceList corev1.NamespaceList
		if err := r.client.List(ctx, &namespaceList); err != nil {
			return common.RequeueWithError(log, "failed to get namespace list", err)
		}
		for _, ns := range namespaceList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	} else {
		// use configured namespaces
		namespaces = r.Namespaces
	}

	if cluster.Spec.ListenersConfig.SSLSecrets != nil {
		// If we haven't deleted all nifiusers yet, iterate namespaces and delete all nifiusers
		// with the matching label.
		if util.StringSliceContains(cluster.GetFinalizers(), clusterUsersFinalizer) {
			log.Info(fmt.Sprintf("Sending delete nifiusers request to all namespaces for cluster %s/%s", cluster.Namespace, cluster.Name))
			for _, ns := range namespaces {
				if err := r.client.DeleteAllOf(
					ctx,
					&v1alpha1.NifiUser{},
					client.InNamespace(ns),
					client.MatchingLabels{common.ClusterRefLabel: common.ClusterLabelString(cluster)},
				); err != nil {
					if client.IgnoreNotFound(err) != nil {
						return common.RequeueWithError(log, "failed to send delete request for children nifiusers", err)
					}
					log.Info(fmt.Sprintf("No matching nifiusers in namespace: %s", ns))
				}
			}
			if cluster, err = r.removeFinalizer(ctx, cluster, clusterUsersFinalizer); err != nil {
				return common.RequeueWithError(log, "failed to remove users finalizer from nificluster", err)
			}
		}

		// Do any necessary PKI cleanup - a PKI backend should make sure any
		// user finalizations are done before it does its final cleanup
		log.Info("Tearing down any PKI resources for the nificluster")
		if err = pki.GetPKIManager(r.client, cluster).FinalizePKI(ctx, log); err != nil {
			switch err.(type) {
			case errorfactory.ResourceNotReady:
				log.Info("The PKI is not ready to be torn down")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: time.Duration(5) * time.Second,
				}, nil
			default:
				return common.RequeueWithError(log, "failed to finalize PKI", err)
			}
		}

	}

	log.Info("Finalizing deletion of nificluster instance")
	if _, err = r.removeFinalizer(ctx, cluster, clusterFinalizer); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// We may have been a requeue from earlier with all conditions met - but with
			// the state of the finalizer not yet reflected in the response we got.
			return common.Reconciled()
		}
		return common.RequeueWithError(log, "failed to remove main finalizer", err)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileNifiCluster) removeFinalizer(ctx context.Context, cluster *v1alpha1.NifiCluster, finalizer string) (updated *v1alpha1.NifiCluster, err error) {
	cluster.SetFinalizers(util.StringSliceRemove(cluster.GetFinalizers(), finalizer))
	return r.updateAndFetchLatest(ctx, cluster)
}

func (r *ReconcileNifiCluster) updateAndFetchLatest(ctx context.Context, cluster *v1alpha1.NifiCluster) (*v1alpha1.NifiCluster, error) {
	typeMeta := cluster.TypeMeta
	err := r.client.Update(ctx, cluster)
	if err != nil {
		return nil, err
	}
	cluster.TypeMeta = typeMeta
	return cluster, nil
}

func (r *ReconcileNifiCluster) ensureFinalizers(ctx context.Context, cluster *v1alpha1.NifiCluster) (updated *v1alpha1.NifiCluster, err error) {
	finalizers := []string{clusterFinalizer}
	if cluster.Spec.ListenersConfig.SSLSecrets != nil {
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
