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
	"encoding/json"
	"fmt"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	usercli "github.com/konpyutaika/nifikop/pkg/clientwrappers/user"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/pki"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
)

var userFinalizer = "nifiusers.nifi.konpyutaika.com/finalizer"

// NifiUserReconciler reconciles a NifiUser object
type NifiUserReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiusers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiusers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiusers/finalizers,verbs=update
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=clusterissuers,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiUser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *NifiUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("nifiuser", req.NamespacedName)
	interval := util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset)
	var err error

	// Fetch the NifiUser instance
	instance := &v1alpha1.NifiUser{}
	if err = r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return Reconciled()
		}
		// Error reading the object - requeue the request.
		return RequeueWithError(r.Log, err.Error(), err)
	}

	// Get the last configuration viewed by the operator.
	o, err := patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	// Create it if not exist.
	if o == nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation", err)
		}
		if err := r.Client.Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiUser", err)
		}
		o, err = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Check if the cluster reference changed.
	original := &v1alpha1.NifiUser{}
	current := instance.DeepCopy()
	json.Unmarshal(o, original)
	if !v1alpha1.ClusterRefsEquals([]v1alpha1.ClusterReference{original.Spec.ClusterRef, instance.Spec.ClusterRef}) {
		instance.Spec.ClusterRef = original.Spec.ClusterRef
	}

	// Prepare cluster connection configurations
	var clientConfig *clientconfig.NifiConfig
	var clusterConnect clientconfig.ClusterConnect

	// Get the client config manager associated to the cluster ref.
	clusterRef := instance.Spec.ClusterRef
	clusterRef.Namespace = GetClusterRefNamespace(instance.Namespace, instance.Spec.ClusterRef)
	configManager := config.GetClientConfigManager(r.Client, clusterRef)

	// Generate the connect object
	if clusterConnect, err = configManager.BuildConnect(); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			r.Log.Info("Cluster is gone already, there is nothing we can do")
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer from NifiUser", err)
			}
			return Reconciled()
		}

		// If the referenced cluster no more exist, just skip the deletion requirement in cluster ref change case.
		if !v1alpha1.ClusterRefsEquals([]v1alpha1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
				return RequeueWithError(r.Log, "could not apply last state to annotation", err)
			}
			if err := r.Client.Update(ctx, current); err != nil {
				return RequeueWithError(r.Log, "failed to update NifiUser", err)
			}
			return RequeueAfter(time.Duration(15) * time.Second)
		}

		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to lookup reference cluster : %s in %s",
				instance.Spec.ClusterRef.Name, clusterRef.Namespace))
		return RequeueWithError(r.Log, "failed to lookup referenced cluster", err)
	}

	// Get the referenced NifiCluster
	var cluster *v1alpha1.NifiCluster
	if cluster, err = k8sutil.LookupNifiCluster(r.Client, instance.Spec.ClusterRef.Name, clusterRef.Namespace); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			r.Log.Info("Cluster is gone already, there is nothing we can do")
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer from NifiUser", err)
			}
			return Reconciled()
		}
	}

	if v1alpha1.ClusterRefsEquals([]v1alpha1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) &&
		instance.Spec.GetCreateCert() && !clusterConnect.IsExternal() {

		// Avoid panic if the user wants to create a nifi user but the cluster is in plaintext mode
		// TODO: refactor this and use webhook to validate if the cluster is eligible to create a nifi user
		if cluster.Spec.ListenersConfig.SSLSecrets == nil {
			return RequeueWithError(r.Log, "could not create Nifi user since cluster does not use ssl", errors.New("failed to create Nifi user"))
		}

		pkiManager := pki.GetPKIManager(r.Client, cluster)

		r.Recorder.Event(instance, corev1.EventTypeNormal, "ReconcilingCertificate",
			fmt.Sprintf("Reconciling certificate for nifi user %s", instance.Name))
		// Reconcile no matter what to get a user certificate instance for ACL management
		// TODO (tinyzimmer): This can go wrong if the user made a mistake in their secret path
		// using the vault backend, then tried to delete and fix it. Should probably
		// have the PKIManager export a GetUserCertificate specifically for deletions
		// that will allow the error to fall through if the certificate doesn't exist.
		_, err := pkiManager.ReconcileUserCertificate(ctx, instance, r.Scheme)
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.ResourceNotReady:
				r.Log.Info("generated secret not found, may not be ready")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: interval / 3,
				}, nil
			case errorfactory.FatalReconcileError:
				// TODO: (tinyzimmer) - Sleep for longer for now to give user time to see the error
				// But really we should catch these kinds of issues in a pre-admission hook in a future PR
				// The user can fix while this is looping and it will pick it up next reconcile attempt
				r.Log.Error(err, "Fatal error attempting to reconcile the user certificate. If using vault perhaps a permissions issue or improperly configured PKI?")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: interval,
				}, nil
			case errorfactory.VaultAPIFailure:
				// Same as above in terms of things that could be checked pre-flight on the cluster
				r.Log.Error(err, "Vault API error attempting to reconcile the user certificate. If using vault perhaps a permissions issue or improperly configured PKI?")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: interval,
				}, nil
			default:
				return RequeueWithError(r.Log, "failed to reconcile user secret", err)
			}
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "ReconciledCertificate",
			fmt.Sprintf("Reconciled certificate for nifi user %s", instance.Name))

		// check if marked for deletion
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			r.Log.Info("Nifi user is marked for deletion, revoking certificates")
			if err = pkiManager.FinalizeUserCertificate(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to finalize user certificate", err)
			}
		}
	}

	// Generate the client configuration.
	clientConfig, err = configManager.BuildConfig()
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to create HTTP client for the referenced cluster : %s in %s",
				instance.Spec.ClusterRef.Name, clusterRef.Namespace))
		// the cluster is gone, so just remove the finalizer
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return RequeueWithError(r.Log, fmt.Sprintf("failed to remove finalizer from NifiUser %s", instance.Name), err)
			}
			return Reconciled()
		}
		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to create HTTP client the for referenced cluster", err)
	}

	// check if marked for deletion
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, clientConfig)
	}

	// Ensure the cluster is ready to receive actions
	if !clusterConnect.IsReady(r.Log) {
		r.Log.Info("Cluster is not ready yet, will wait until it is.")
		r.Recorder.Event(instance, corev1.EventTypeNormal, "ReferenceClusterNotReady",
			fmt.Sprintf("The referenced cluster is not ready yet : %s in %s",
				instance.Spec.ClusterRef.Name, clusterConnect.Id()))
		// the cluster does not exist - should have been caught pre-flight
		return RequeueAfter(interval)
	}

	// Ìn case of the cluster reference changed.
	if !v1alpha1.ClusterRefsEquals([]v1alpha1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
		// Delete the resource on the previous cluster.
		if err := usercli.RemoveUser(instance, clientConfig); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "RemoveError",
				fmt.Sprintf("Failed to delete NifiUser %s from cluster %s before moving in %s",
					instance.Name, original.Spec.ClusterRef.Name, original.Spec.ClusterRef.Name))
			return RequeueWithError(r.Log, "Failed to delete NifiUser before moving", err)
		}
		// Update the last view configuration to the current one.
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation", err)
		}
		if err := r.Client.Update(ctx, current); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiUser", err)
		}
		return RequeueAfter(interval)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciling",
		fmt.Sprintf("Reconciling user %s", instance.Name))

	// Check if the NiFi user already exist
	exist, err := usercli.ExistUser(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing registry client", err)
	}

	if !exist {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Creating",
			fmt.Sprintf("Creating user %s", instance.Name))

		var status *v1alpha1.NifiUserStatus

		status, err = usercli.FindUserByIdentity(instance, clientConfig)
		if err != nil {
			return RequeueWithError(r.Log, "failure finding user", err)
		}

		if status == nil {
			// Create NiFi registry client
			status, err = usercli.CreateUser(instance, clientConfig)
			if err != nil {
				return RequeueWithError(r.Log, "failure creating user", err)
			}
		}

		instance.Status = *status
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiUser status", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Created",
			fmt.Sprintf("Created user %s", instance.Name))
	}

	// Sync user resource with NiFi side component
	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronizing",
		fmt.Sprintf("Synchronizing user %s", instance.Name))
	status, err := usercli.SyncUser(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failed to sync NifiUser", err)
	}

	instance.Status = *status
	if err := r.Client.Status().Update(ctx, instance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiUser status", err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
		fmt.Sprintf("Synchronized user %s", instance.Name))

	// ensure a NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiCluster label on user", err)
	}

	// ensure a finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), userFinalizer) {
		r.addFinalizer(instance)
		if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiUser with finalizer", err)
		}
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiUser", err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciled",
		fmt.Sprintf("Reconciling user %s", instance.Name))

	r.Log.Info("Ensured user")

	return RequeueAfter(interval)

	// set user status
	//instance.Status = v1alpha1.NifiUserStatus{
	//	State: v1alpha1.UserStateCreated,
	//}
	//if err := r.Client.Status().Update(ctx, instance); err != nil {
	//	return RequeueWithError(r.Log, "failed to update NifiUser status", err)
	//}

	//return Reconciled()
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiUserReconciler) SetupWithManager(mgr ctrl.Manager, certManagerEnabled bool) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiUser{}).
		Owns(&corev1.Secret{})

	if certManagerEnabled {
		builder.Owns(&certv1.Certificate{})
	}

	return builder.Complete(r)
}

func (r *NifiUserReconciler) ensureClusterLabel(ctx context.Context, cluster clientconfig.ClusterConnect, user *v1alpha1.NifiUser) (*v1alpha1.NifiUser, error) {
	labels := ApplyClusterReferenceLabel(cluster, user.GetLabels())
	if !reflect.DeepEqual(labels, user.GetLabels()) {
		user.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, user)
	}
	return user, nil
}

func (r *NifiUserReconciler) updateAndFetchLatest(ctx context.Context, user *v1alpha1.NifiUser) (*v1alpha1.NifiUser, error) {
	typeMeta := user.TypeMeta
	err := r.Client.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	user.TypeMeta = typeMeta
	return user, nil
}

func (r *NifiUserReconciler) checkFinalizers(ctx context.Context, user *v1alpha1.NifiUser, config *clientconfig.NifiConfig) (reconcile.Result, error) {
	r.Log.Info(fmt.Sprintf("NiFi user %s is marked for deletion", user.Name))
	var err error
	if util.StringSliceContains(user.GetFinalizers(), userFinalizer) {
		if err = r.finalizeNifiUser(user, config); err != nil {
			return RequeueWithError(r.Log, "failed to finalize nifiuser", err)
		}
		// remove finalizer
		if err = r.removeFinalizer(ctx, user); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from NifiUser", err)
		}
	}
	return Reconciled()
}

func (r *NifiUserReconciler) removeFinalizer(ctx context.Context, user *v1alpha1.NifiUser) error {
	r.Log.V(5).Info(fmt.Sprintf("Removing finalizer for NifiUser %s", user.Name))
	user.SetFinalizers(util.StringSliceRemove(user.GetFinalizers(), userFinalizer))
	_, err := r.updateAndFetchLatest(ctx, user)
	return err
}

func (r *NifiUserReconciler) finalizeNifiUser(user *v1alpha1.NifiUser, config *clientconfig.NifiConfig) error {
	if err := usercli.RemoveUser(user, config); err != nil {
		return err
	}
	r.Log.Info(fmt.Sprintf("Deleted user %s", user.Name))
	return nil
}

func (r *NifiUserReconciler) addFinalizer(user *v1alpha1.NifiUser) {
	r.Log.Info(fmt.Sprintf("Adding Finalizer for the NifiUser %s", user.Name))
	user.SetFinalizers(append(user.GetFinalizers(), userFinalizer))
	return
}
