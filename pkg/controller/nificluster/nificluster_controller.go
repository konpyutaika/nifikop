package nificluster

import (
	"context"
	"emperror.dev/errors"
	"github.com/go-logr/logr"
	v1alpha1 "github.com/orangeopensource/nifi-operator/pkg/apis/nifi/v1alpha1"
	"github.com/orangeopensource/nifi-operator/pkg/errorfactory"
	"github.com/orangeopensource/nifi-operator/pkg/k8sutil"
	"github.com/orangeopensource/nifi-operator/pkg/resources"
	"github.com/orangeopensource/nifi-operator/pkg/resources/nifi"
	"github.com/orangeopensource/nifi-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

var log = logf.Log.WithName("controller_nificluster")

var clusterFinalizer =  "finalizer.nificlusters.nifi.orange.com"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new NifiCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNifiCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner NifiCluster
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
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
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a NifiCluster object and makes changes based on the state read
// and what is in the NifiCluster.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
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
			return reconciled()
		}
		// Error reading the object - requeue the request.
		return requeueWithError(reqLogger, err.Error(), err)
	}

	// Check if marked for deletion and run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, reqLogger, instance)
	}

	if instance.Status.State != v1alpha1.NifiClusterRollingUpgrading {
		if err := k8sutil.UpdateCRStatus(r.client, instance, v1alpha1.NifiClusterReconciling, reqLogger); err != nil {
			return requeueWithError(log, err.Error(), err)
		}
	}

	reconcilers := []resources.ComponentReconciler{
//		envoy.New(r.Client, instance),
//		istioingress.New(r.Client, instance),
//		kafkamonitoring.New(r.Client, instance),
//		cruisecontrolmonitoring.New(r.Client, instance),
		nifi.New(r.client, r.scheme, instance),
//		cruisecontrol.New(r.Client, instance),
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
				return requeueWithError(reqLogger, err.Error(), err)
			}
		}
	}

	reqLogger.Info("ensuring finalizers on nificluster")
	if instance, err = r.ensureFinalizers(ctx, instance); err != nil {
		return requeueWithError(log, "failed to ensure finalizers on kafkacluster instance", err)
	}

	//Update rolling upgrade last successful state
	if instance.Status.State == v1alpha1.NifiClusterRollingUpgrading {
		if err := k8sutil.UpdateRollingUpgradeState(r.client, instance, time.Now(), reqLogger); err != nil {
			return requeueWithError(reqLogger, err.Error(), err)
		}
	}

	if err := k8sutil.UpdateCRStatus(r.client, instance, v1alpha1.NifiClusterRunning, reqLogger); err != nil {
		return requeueWithError(log, err.Error(), err)
	}

	return reconciled()
}

func (r *ReconcileNifiCluster) checkFinalizers(ctx context.Context, log logr.Logger, cluster *v1alpha1.NifiCluster) (reconcile.Result, error) {
	log.Info("NifiCluster is marked for deletion, checking for children")

	// If the main finalizer is gone then we've already finished up
	if !util.StringSliceContains(cluster.GetFinalizers(), clusterFinalizer) {
		return reconciled()
	}

	var err error

	// Fetch a list of all namespaces for DeleteAllOf requests
	var namespaces corev1.NamespaceList
	if err := r.client.List(ctx, &namespaces); err != nil {
		return requeueWithError(log, "failed to get namespace list", err)
	}

	/*// If we haven't deleted all kafkatopics yet, iterate namespaces and delete all kafkatopics
	// with the matching label.
	if util.StringSliceContains(cluster.GetFinalizers(), clusterTopicsFinalizer) {
		log.Info(fmt.Sprintf("Sending delete kafkatopics request to all namespaces for cluster %s/%s", cluster.Namespace, cluster.Name))
		for _, ns := range namespaces.Items {
			if err := r.Client.DeleteAllOf(
				ctx,
				&v1alpha1.KafkaTopic{},
				client.InNamespace(ns.Name),
				client.MatchingLabels{clusterRefLabel: clusterLabelString(cluster)},
			); err != nil {
				if client.IgnoreNotFound(err) != nil {
					return requeueWithError(log, "failed to send delete request for children kafkatopics", err)
				}
				log.Info(fmt.Sprintf("No matching kafkatopics in namespace: %s", ns.Name))
			}

		}
		if cluster, err = r.removeFinalizer(ctx, cluster, clusterTopicsFinalizer); err != nil {
			return requeueWithError(log, "failed to remove topics finalizer from kafkacluster", err)
		}
	}*/

	/*// If any of the topics still exist, it means their finalizer is still running.
	// Wait to make sure we have fully cleaned up zookeeper. Also if we delete
	// our kafkausers before all topics are finished cleaning up, we will lose
	// our controller certificate.
	log.Info("Ensuring all topics have finished cleaning up")
	var childTopics v1alpha1.KafkaTopicList
	if err = r.Client.List(
		ctx,
		&childTopics,
		client.InNamespace(metav1.NamespaceAll),
		client.MatchingLabels{clusterRefLabel: clusterLabelString(cluster)},
	); err != nil {
		return requeueWithError(log, "failed to list kafkatopics", err)
	}
	if len(childTopics.Items) > 0 {
		log.Info(fmt.Sprintf("Still waiting for the following topics to be deleted: %v", topicListToStrSlice(childTopics)))
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Duration(3) * time.Second,
		}, nil
	}*/

	if cluster.Spec.ListenersConfig.SSLSecrets != nil {
		/*// If we haven't deleted all kafkausers yet, iterate namespaces and delete all kafkausers
		// with the matching label.
		if util.StringSliceContains(cluster.GetFinalizers(), clusterUsersFinalizer) {
			log.Info(fmt.Sprintf("Sending delete kafkausers request to all namespaces for cluster %s/%s", cluster.Namespace, cluster.Name))
			for _, ns := range namespaces.Items {
				if err := r.Client.DeleteAllOf(
					ctx,
					&v1alpha1.KafkaUser{},
					client.InNamespace(ns.Name),
					client.MatchingLabels{clusterRefLabel: clusterLabelString(cluster)},
				); err != nil {
					if client.IgnoreNotFound(err) != nil {
						return requeueWithError(log, "failed to send delete request for children kafkausers", err)
					}
					log.Info(fmt.Sprintf("No matching kafkausers in namespace: %s", ns.Name))
				}
			}
			if cluster, err = r.removeFinalizer(ctx, cluster, clusterUsersFinalizer); err != nil {
				return requeueWithError(log, "failed to remove users finalizer from kafkacluster", err)
			}
		}*/

		/*// Do any necessary PKI cleanup - a PKI backend should make sure any
		// user finalizations are done before it does its final cleanup
		log.Info("Tearing down any PKI resources for the kafkacluster")
		if err = pki.GetPKIManager(r.client, cluster).FinalizePKI(ctx, log); err != nil {
			switch err.(type) {
			case errorfactory.ResourceNotReady:
				log.Info("The PKI is not ready to be torn down")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: time.Duration(5) * time.Second,
				}, nil
			default:
				return requeueWithError(log, "failed to finalize PKI", err)
			}
		}*/

	}

	log.Info("Finalizing deletion of nificluster instance")
	if _, err = r.removeFinalizer(ctx, cluster, clusterFinalizer); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// We may have been a requeue from earlier with all conditions met - but with
			// the state of the finalizer not yet reflected in the response we got.
			return reconciled()
		}
		return requeueWithError(log, "failed to remove main finalizer", err)
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


// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *v1alpha1.NifiCluster) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

func (r *ReconcileNifiCluster) ensureFinalizers(ctx context.Context, cluster *v1alpha1.NifiCluster) (updated *v1alpha1.NifiCluster, err error) {
	/*finalizers := []string{clusterFinalizer, clusterTopicsFinalizer}
	if cluster.Spec.ListenersConfig.SSLSecrets != nil {
		finalizers = append(finalizers, clusterUsersFinalizer)
	}
	for _, finalizer := range finalizers {
		if util.StringSliceContains(cluster.GetFinalizers(), finalizer) {
			continue
		}
		cluster.SetFinalizers(append(cluster.GetFinalizers(), finalizer))
	}*/
	return r.updateAndFetchLatest(ctx, cluster)
}
