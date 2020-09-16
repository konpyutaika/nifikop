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

package nifiregistryclient

import (
	"context"
	"reflect"
	"time"

	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/registryclient"
	"github.com/Orange-OpenSource/nifikop/pkg/k8sutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/controller/common"
	"github.com/go-logr/logr"

	"github.com/Orange-OpenSource/nifikop/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_nifiregistryclient")

var registryClientFinalizer = "finalizer.nifiregistryclients.nifi.orange.com"

// Add creates a new NifiRegistryClient Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, namespaces []string) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNifiRegistryClient{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nifiregistryclient-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NifiRegistryClient
	err = c.Watch(&source.Kind{Type: &v1alpha1.NifiRegistryClient{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	if err != nil {
		if _, ok := err.(*meta.NoKindMatchError); !ok {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileNifiRegistryClient implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNifiRegistryClient{}

// ReconcileNifiRegistryClient reconciles a NifiRegistryClient object
type ReconcileNifiRegistryClient struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifiregistryclients,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifiregistryclients/status,verbs=get;update;patch

// Reconcile reads that state of the registry client for a NifiRegistryClient object and makes changes based on the state read
// and what is in the NifiRegistryClient.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNifiRegistryClient) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NifiRegistryClient")
	var err error

	// Get a context for the request
	ctx := context.Background()

	// Fetch the NifiRegistryClient instance
	instance := &v1alpha1.NifiRegistryClient{}
	if err = r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return common.Reconciled()
		}
		// Error reading the object - requeue the request.
		return common.RequeueWithError(reqLogger, err.Error(), err)
	}

	// Get the referenced NifiCluster
	clusterNamespace := common.GetClusterRefNamespace(instance.Namespace, instance.Spec.ClusterRef)
	var cluster *v1alpha1.NifiCluster
	if cluster, err = k8sutil.LookupNifiCluster(r.client, instance.Spec.ClusterRef.Name, clusterNamespace); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			reqLogger.Info("Cluster is already gone, there is nothing we can do")
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return common.RequeueWithError(reqLogger, "failed to remove finalizer", err)
			}
			return common.Reconciled()
		}

		// the cluster does not exist - should have been caught pre-flight
		return common.RequeueWithError(reqLogger, "failed to lookup referenced cluster", err)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, reqLogger, instance, cluster)
	}

	// Check if the NiFi registry client already exist
	exist, err := registryclient.ExistRegistryClient(r.client, instance, cluster)
	if err != nil {
		return common.RequeueWithError(reqLogger, "failure checking for existing registry client", err)
	}

	if !exist {
		// Create NiFi registry client
		status, err := registryclient.CreateRegistryClient(r.client, instance, cluster)
		if err != nil {
			return common.RequeueWithError(reqLogger, "failure creating registry client", err)
		}

		instance.Status = *status
		if err := r.client.Status().Update(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to update NifiRegistryClient status", err)
		}
	}

	// Sync RegistryClient resource with NiFi side component
	status, err := registryclient.SyncRegistryClient(r.client, instance, cluster)
	if err != nil {
		return common.RequeueWithError(reqLogger, "failed to sync NifiRegistryClient", err)
	}

	instance.Status = *status
	if err := r.client.Status().Update(ctx, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to update NifiRegistryClient status", err)
	}

	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, cluster, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to ensure NifiCluster label on registry client", err)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), registryClientFinalizer) {
		reqLogger.Info("Adding Finalizer for NifiRegistryClient")
		instance.SetFinalizers(append(instance.GetFinalizers(), registryClientFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to update NifiRegistryClient", err)
	}

	reqLogger.Info("Ensured Registry Client")

	return common.RequeueAfter(time.Duration(15) * time.Second)
}

func (r *ReconcileNifiRegistryClient) ensureClusterLabel(ctx context.Context, cluster *v1alpha1.NifiCluster,
	registryClient *v1alpha1.NifiRegistryClient) (*v1alpha1.NifiRegistryClient, error) {

	labels := common.ApplyClusterRefLabel(cluster, registryClient.GetLabels())
	if !reflect.DeepEqual(labels, registryClient.GetLabels()) {
		registryClient.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, registryClient)
	}
	return registryClient, nil
}

func (r *ReconcileNifiRegistryClient) updateAndFetchLatest(ctx context.Context,
	registryClient *v1alpha1.NifiRegistryClient) (*v1alpha1.NifiRegistryClient, error) {

	typeMeta := registryClient.TypeMeta
	err := r.client.Update(ctx, registryClient)
	if err != nil {
		return nil, err
	}
	registryClient.TypeMeta = typeMeta
	return registryClient, nil
}

func (r *ReconcileNifiRegistryClient) checkFinalizers(ctx context.Context, reqLogger logr.Logger,
	registryClient *v1alpha1.NifiRegistryClient, cluster *v1alpha1.NifiCluster) (reconcile.Result, error) {

	reqLogger.Info("NiFi registry client is marked for deletion")
	var err error
	if util.StringSliceContains(registryClient.GetFinalizers(), registryClientFinalizer) {
		if err = r.finalizeNifiRegistryClient(reqLogger, registryClient, cluster); err != nil {
			return common.RequeueWithError(reqLogger, "failed to finalize kafkatopic", err)
		}
		if err = r.removeFinalizer(ctx, registryClient); err != nil {
			return common.RequeueWithError(reqLogger, "failed to remove finalizer from kafkatopic", err)
		}
	}
	return common.Reconciled()
}

func (r *ReconcileNifiRegistryClient) removeFinalizer(ctx context.Context, flow *v1alpha1.NifiRegistryClient) error {
	flow.SetFinalizers(util.StringSliceRemove(flow.GetFinalizers(), registryClientFinalizer))
	_, err := r.updateAndFetchLatest(ctx, flow)
	return err
}

func (r *ReconcileNifiRegistryClient) finalizeNifiRegistryClient(reqLogger logr.Logger, registryClient *v1alpha1.NifiRegistryClient,
	cluster *v1alpha1.NifiCluster) error {

	if err := registryclient.RemoveRegistryClient(r.client, registryClient, cluster); err != nil {
		return err
	}
	reqLogger.Info("Delete Registry client")

	return nil
}
