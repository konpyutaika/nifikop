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

package nifiuser

import (
	"context"
	"reflect"
	"time"

	"emperror.dev/errors"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"

	v1alpha1 "github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
	common "github.com/Orange-OpenSource/nifikop/pkg/controller/common"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/Orange-OpenSource/nifikop/pkg/k8sutil"
	"github.com/Orange-OpenSource/nifikop/pkg/pki"
	pkicommon "github.com/Orange-OpenSource/nifikop/pkg/util/pki"
	"github.com/go-logr/logr"

	"github.com/Orange-OpenSource/nifikop/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
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

var log = logf.Log.WithName("controller_nifiuser")

var userFinalizer = "finalizer.nifiusers.nifi.orange.com"

// Add creates a new NifiCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, namespaces []string) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNifiUser{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nifiuser-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NifiUser
	err = c.Watch(&source.Kind{Type: &v1alpha1.NifiUser{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner NifiUser
	err = c.Watch(&source.Kind{Type: &certv1.Certificate{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.NifiUser{},
	})
	if err != nil {
		if _, ok := err.(*meta.NoKindMatchError); !ok {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileNifiUser implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNifiUser{}

// ReconcileNifiCluster reconciles a NifiUser object
type ReconcileNifiUser struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifiusers,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifiusers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=clusterissuers,verbs=get;list;watch;create;update;patch;delete

// Reconcile reads that state of the cluster for a NifiUser object and makes changes based on the state read
// and what is in the NifiUser.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNifiUser) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NifiUser")
	var err error

	// create a context for the request
	ctx := context.Background()

	// Fetch the NifiUser instance
	instance := &v1alpha1.NifiUser{}
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
			reqLogger.Info("Cluster is gone already, there is nothing we can do")
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return common.RequeueWithError(reqLogger, "failed to remove finalizer from NifiUser", err)
			}
			return common.Reconciled()
		}
		return common.RequeueWithError(reqLogger, "failed to lookup referenced cluster", err)
	}
	// Avoid panic if the user wants to create a nifi user but the cluster is in plaintext mode
	// TODO: refactor this and use webhook to validate if the cluster is eligible to create a nifi user
	if cluster.Spec.ListenersConfig.SSLSecrets == nil {
		return common.RequeueWithError(reqLogger, "could not create Nifi user since cluster does not use ssl", errors.New("failed to create Nifi user"))
	}

	pkiManager := pki.GetPKIManager(r.client, cluster)

	// Reconcile no matter what to get a user certificate instance for ACL management
	// TODO (tinyzimmer): This can go wrong if the user made a mistake in their secret path
	// using the vault backend, then tried to delete and fix it. Should probably
	// have the PKIManager export a GetUserCertificate specifically for deletions
	// that will allow the error to fall through if the certificate doesn't exist.
	user, err := pkiManager.ReconcileUserCertificate(ctx, instance, r.scheme)
	if err != nil {
		switch errors.Cause(err).(type) {
		case errorfactory.ResourceNotReady:
			reqLogger.Info("generated secret not found, may not be ready")
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: time.Duration(5) * time.Second,
			}, nil
		case errorfactory.FatalReconcileError:
			// TODO: (tinyzimmer) - Sleep for longer for now to give user time to see the error
			// But really we should catch these kinds of issues in a pre-admission hook in a future PR
			// The user can fix while this is looping and it will pick it up next reconcile attempt
			reqLogger.Error(err, "Fatal error attempting to reconcile the user certificate. If using vault perhaps a permissions issue or improperly configured PKI?")
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: time.Duration(15) * time.Second,
			}, nil
		case errorfactory.VaultAPIFailure:
			// Same as above in terms of things that could be checked pre-flight on the cluster
			reqLogger.Error(err, "Vault API error attempting to reconcile the user certificate. If using vault perhaps a permissions issue or improperly configured PKI?")
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: time.Duration(15) * time.Second,
			}, nil
		default:
			return common.RequeueWithError(reqLogger, "failed to reconcile user secret", err)
		}
	}

	// check if marked for deletion
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		reqLogger.Info("Nifi user is marked for deletion, revoking certificates")
		if err = pkiManager.FinalizeUserCertificate(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to finalize user certificate", err)
		}
		return r.checkFinalizers(ctx, reqLogger, cluster, instance, user)
	}

	// ensure a NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, cluster, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to ensure NifiCluster label on user", err)
	}

	// If topic grants supplied, grab a broker connection and set ACLs
	// TODO : Check Grant into NiFi
	/*if len(instance.Spec.TopicGrants) > 0 {
		broker, close, err := newBrokerConnection(reqLogger, r.Client, cluster)
		if err != nil {
			return checkBrokerConnectionError(reqLogger, err)
		}
		defer close()

		// TODO (tinyzimmer): Should probably take this opportunity to see if we are removing any ACLs
		for _, grant := range instance.Spec.TopicGrants {
			reqLogger.Info(fmt.Sprintf("Ensuring %s ACLs for User: %s -> Topic: %s", grant.AccessType, user.DN(), grant.TopicName))
			// CreateUserACLs returns no error if the ACLs already exist
			if err = broker.CreateUserACLs(grant.AccessType, grant.PatternType, user.DN(), grant.TopicName); err != nil {
				return common.RequeueWithError(reqLogger, "failed to ensure ACLs for NifiUser", err)
			}
		}
	}*/

	// ensure a finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), userFinalizer) {
		r.addFinalizer(reqLogger, instance)
		if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to update NifiUser with finalizer", err)
		}
	}

	// set user status
	instance.Status = v1alpha1.NifiUserStatus{
		State: v1alpha1.UserStateCreated,
	}
	if err := r.client.Status().Update(ctx, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to update NifiUser status", err)
	}

	return common.Reconciled()
}

func (r *ReconcileNifiUser) ensureClusterLabel(ctx context.Context, cluster *v1alpha1.NifiCluster, user *v1alpha1.NifiUser) (*v1alpha1.NifiUser, error) {
	labels := common.ApplyClusterRefLabel(cluster, user.GetLabels())
	if !reflect.DeepEqual(labels, user.GetLabels()) {
		user.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, user)
	}
	return user, nil
}

func (r *ReconcileNifiUser) updateAndFetchLatest(ctx context.Context, user *v1alpha1.NifiUser) (*v1alpha1.NifiUser, error) {
	typeMeta := user.TypeMeta
	err := r.client.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	user.TypeMeta = typeMeta
	return user, nil
}

func (r *ReconcileNifiUser) checkFinalizers(ctx context.Context, reqLogger logr.Logger, cluster *v1alpha1.NifiCluster, instance *v1alpha1.NifiUser, user *pkicommon.UserCertificate) (reconcile.Result, error) {
	// run finalizers
	var err error
	if util.StringSliceContains(instance.GetFinalizers(), userFinalizer) {
		/*if len(instance.Spec.TopicGrants) > 0 {
			if err = r.finalizeNifiUserACLs(reqLogger, cluster, user); err != nil {
				return common.RequeueWithError(reqLogger, "failed to finalize NifiUser", err)
			}
		}*/
		// remove finalizer
		if err = r.removeFinalizer(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to remove finalizer from NifiUser", err)
		}
	}
	return common.Reconciled()
}

func (r *ReconcileNifiUser) removeFinalizer(ctx context.Context, user *v1alpha1.NifiUser) error {
	user.SetFinalizers(util.StringSliceRemove(user.GetFinalizers(), userFinalizer))
	_, err := r.updateAndFetchLatest(ctx, user)
	return err
}

func (r *ReconcileNifiUser) finalizeNifiUserACLs(reqLogger logr.Logger, cluster *v1alpha1.NifiCluster, user *pkicommon.UserCertificate) error {
	if k8sutil.IsMarkedForDeletion(cluster.ObjectMeta) {
		reqLogger.Info("Cluster is being deleted, skipping ACL deletion")
		return nil
	}

	return nil
}

func (r *ReconcileNifiUser) addFinalizer(reqLogger logr.Logger, user *v1alpha1.NifiUser) {
	reqLogger.Info("Adding Finalizer for the NifiUser")
	user.SetFinalizers(append(user.GetFinalizers(), userFinalizer))
	return
}
