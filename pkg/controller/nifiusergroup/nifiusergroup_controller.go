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

package nifiusergroup

import (
	"context"
	"reflect"
	"time"

	"emperror.dev/errors"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/usergroup"
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

var log = logf.Log.WithName("controller_nifiusergroup")

var registryClientFinalizer = "finalizer.nifiusergroups.nifi.orange.com"

// Add creates a new NifiUserGroup Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, namespaces []string) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNifiUserGroup{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nifiusergroup-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NifiUserGRoup
	err = c.Watch(&source.Kind{Type: &v1alpha1.NifiUserGroup{}}, &handler.EnqueueRequestForObject{})
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

// blank assignment to verify that ReconcileNifiUserGroup implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNifiUserGroup{}

// ReconcileNifiUserGroup reconciles a NifiUserGroup object
type ReconcileNifiUserGroup struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifiusergroups,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifiusergroups/status,verbs=get;update;patch

// Reconcile reads that state of the user group for a NifiUserGroup object and makes changes based on the state read
// and what is in the NifiUserGroup.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNifiUserGroup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NifiUserGroup")
	var err error

	// Get a context for the request
	ctx := context.Background()

	// Fetch the NifiUserGroup instance
	instance := &v1alpha1.NifiUserGroup{}
	if err = r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return common.Reconciled()
		}
		// Error reading the object - requeue the request.
		return common.RequeueWithError(reqLogger, err.Error(), err)
	}

	var users []*v1alpha1.NifiUser

	for _, userRef := range instance.Spec.UsersRef {
		var user *v1alpha1.NifiUser
		userNamespace := common.GetUserRefNamespace(instance.Namespace, userRef)

		if user, err = k8sutil.LookupNifiUser(r.client, userRef.Name, userNamespace); err != nil {

			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				reqLogger.Info("User is already gone, there is nothing we can do")
				if err = r.removeFinalizer(ctx, instance); err != nil {
					return common.RequeueWithError(reqLogger, "failed to remove finalizer", err)
				}
				return common.Reconciled()
			}

			// the cluster does not exist - should have been caught pre-flight
			return common.RequeueWithError(reqLogger, "failed to lookup referenced user", err)
		}
		// Check if cluster references are the same
		clusterNamespace := common.GetClusterRefNamespace(instance.Namespace, instance.Spec.ClusterRef)
		if user != nil && (userNamespace != clusterNamespace || user.Spec.ClusterRef.Name != instance.Spec.ClusterRef.Name) {
			return common.RequeueWithError(
				reqLogger,
				"failed to lookup referenced cluster, due to inconsistency",
				errors.New("inconsistent cluster references"))
		}

		users = append(users, user)
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
		return r.checkFinalizers(ctx, reqLogger, instance, users, cluster)
	}

	// Check if the NiFi user group already exist
	exist, err := usergroup.ExistUserGroup(r.client, instance, cluster)
	if err != nil {
		return common.RequeueWithError(reqLogger, "failure checking for existing user group", err)
	}

	if !exist {
		// Create NiFi registry client
		status, err := usergroup.CreateUserGroup(r.client, instance, users, cluster)
		if err != nil {
			return common.RequeueWithError(reqLogger, "failure creating user group", err)
		}

		instance.Status = *status
		if err := r.client.Status().Update(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to update NifiUserGroup status", err)
		}
	}

	// Sync UserGroup resource with NiFi side component
	status, err := usergroup.SyncUserGroup(r.client, instance, users, cluster)
	if err != nil {
		return common.RequeueWithError(reqLogger, "failed to sync NifiUserGroup", err)
	}

	instance.Status = *status
	if err := r.client.Status().Update(ctx, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to update NifiUserGroup status", err)
	}

	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, cluster, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to ensure NifiCluster label on user group", err)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), registryClientFinalizer) {
		reqLogger.Info("Adding Finalizer for NifiUserGroup")
		instance.SetFinalizers(append(instance.GetFinalizers(), registryClientFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to update NifiUserGroup", err)
	}

	reqLogger.Info("Ensured User Group")

	return common.RequeueAfter(time.Duration(15) * time.Second)
}

func (r *ReconcileNifiUserGroup) ensureClusterLabel(ctx context.Context, cluster *v1alpha1.NifiCluster,
	userGroup *v1alpha1.NifiUserGroup) ( *v1alpha1.NifiUserGroup, error) {

	labels := common.ApplyClusterRefLabel(cluster, userGroup.GetLabels())
	if !reflect.DeepEqual(labels, userGroup.GetLabels()) {
		userGroup.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, userGroup)
	}
	return userGroup, nil
}

func (r *ReconcileNifiUserGroup) updateAndFetchLatest(ctx context.Context,
	userGroup *v1alpha1.NifiUserGroup) (*v1alpha1.NifiUserGroup, error) {

	typeMeta := userGroup.TypeMeta
	err := r.client.Update(ctx, userGroup)
	if err != nil {
		return nil, err
	}
	userGroup.TypeMeta = typeMeta
	return userGroup, nil
}

func (r *ReconcileNifiUserGroup) checkFinalizers(ctx context.Context, reqLogger logr.Logger,
	userGroup *v1alpha1.NifiUserGroup, users []*v1alpha1.NifiUser, cluster *v1alpha1.NifiCluster) (reconcile.Result, error) {

	reqLogger.Info("NiFi registry client is marked for deletion")
	var err error
	if util.StringSliceContains(userGroup.GetFinalizers(), registryClientFinalizer) {
		if err = r.finalizeNifiNifiUserGroup(reqLogger, userGroup, users, cluster); err != nil {
			return common.RequeueWithError(reqLogger, "failed to finalize nifiusergroup", err)
		}
		if err = r.removeFinalizer(ctx, userGroup); err != nil {
			return common.RequeueWithError(reqLogger, "failed to remove finalizer from kafkatopic", err)
		}
	}
	return common.Reconciled()
}

func (r *ReconcileNifiUserGroup) removeFinalizer(ctx context.Context, userGroup *v1alpha1.NifiUserGroup) error {
	userGroup.SetFinalizers(util.StringSliceRemove(userGroup.GetFinalizers(), registryClientFinalizer))
	_, err := r.updateAndFetchLatest(ctx, userGroup)
	return err
}

func (r *ReconcileNifiUserGroup) finalizeNifiNifiUserGroup(
	reqLogger logr.Logger,
	userGroup *v1alpha1.NifiUserGroup,
	users []*v1alpha1.NifiUser,
	cluster *v1alpha1.NifiCluster) error {

	if err := usergroup.RemoveUserGroup(r.client, userGroup, users, cluster); err != nil {
		return err
	}
	reqLogger.Info("Delete Registry client")

	return nil
}
