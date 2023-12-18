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
	"reflect"

	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	certv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	usercli "github.com/konpyutaika/nifikop/pkg/clientwrappers/user"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/pki"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var userFinalizer = fmt.Sprintf("nifiusers.%s/finalizer", v1.GroupVersion.Group)

// NifiUserReconciler reconciles a NifiUser object.
type NifiUserReconciler struct {
	client.Client
	Log             zap.Logger
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
	interval := util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset)
	var err error

	// Fetch the NifiUser instance
	instance := &v1.NifiUser{}
	if err = r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return Reconciled()
		}
		// Error reading the object - requeue the request.
		return RequeueWithError(r.Log, err.Error(), err)
	}

	patchInstance := client.MergeFromWithOptions(instance.DeepCopy(), client.MergeFromWithOptimisticLock{})
	// Get the last configuration viewed by the operator.
	o, _ := patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	// Create it if not exist.
	if o == nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for nifi user "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiUser "+instance.Name, err)
		}
		o, _ = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Check if the cluster reference changed.
	original := &v1.NifiUser{}
	current := instance.DeepCopy()
	patchCurrent := client.MergeFromWithOptions(current.DeepCopy(), client.MergeFromWithOptimisticLock{})
	json.Unmarshal(o, original)
	if !v1.ClusterRefsEquals([]v1.ClusterReference{original.Spec.ClusterRef, instance.Spec.ClusterRef}) {
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
			r.Log.Error("Cluster is gone already, there is nothing we can do",
				zap.String("user", instance.Name),
				zap.String("clusterName", clusterRef.Name))
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer from NifiUser "+instance.Name, err)
			}
			return Reconciled()
		}

		// If the referenced cluster no more exist, just skip the deletion requirement in cluster ref change case.
		if !v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
				return RequeueWithError(r.Log, "could not apply last state to annotation for user "+instance.Name, err)
			}
			if err := r.Client.Patch(ctx, current, patchCurrent); err != nil {
				return RequeueWithError(r.Log, "failed to update NifiUser "+instance.Name, err)
			}
			return RequeueAfter(interval)
		}

		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to lookup reference cluster : %s in %s",
				instance.Spec.ClusterRef.Name, clusterRef.Namespace))
		return RequeueWithError(r.Log, "failed to lookup referenced cluster "+clusterRef.Name+" for user "+instance.Name, err)
	}

	// Get the referenced NifiCluster
	var cluster *v1.NifiCluster
	if cluster, err = k8sutil.LookupNifiCluster(r.Client, instance.Spec.ClusterRef.Name, clusterRef.Namespace); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			r.Log.Error("Cluster is gone already, there is nothing we can do",
				zap.String("user", instance.Name),
				zap.String("clusterName", clusterRef.Name))
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer from NifiUser "+instance.Name, err)
			}
			return Reconciled()
		}
	}

	if v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) &&
		instance.Spec.GetCreateCert() && !clusterConnect.IsExternal() {
		// Avoid panic if the user wants to create a nifi user but the cluster is in plaintext mode
		// TODO: refactor this and use webhook to validate if the cluster is eligible to create a nifi user
		if cluster.Spec.ListenersConfig.SSLSecrets == nil {
			return RequeueWithError(r.Log, "could not create Nifi user since cluster does not use ssl. user: "+instance.Name, errors.New("failed to create Nifi user"))
		}

		pkiManager := pki.GetPKIManager(r.Client, cluster)

		// check if marked for deletion. Otherwise, reconcile the certificate
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			r.Log.Info("Nifi user is marked for deletion, revoking certificates and removing finalizers.",
				zap.String("user", instance.Name))
			if err = pkiManager.FinalizeUserCertificate(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to finalize certificate for user "+instance.Name, err)
			}
		} else {
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
					r.Log.Debug("generated secret not found, may not be ready",
						zap.String("user", instance.Name))

					return ctrl.Result{
						Requeue:      true,
						RequeueAfter: interval,
					}, nil
				case errorfactory.FatalReconcileError:
					// TODO: (tinyzimmer) - Sleep for longer for now to give user time to see the error
					// But really we should catch these kinds of issues in a pre-admission hook in a future PR
					// The user can fix while this is looping and it will pick it up next reconcile attempt
					r.Log.Error("Fatal error attempting to reconcile the user certificate. If using vault perhaps a permissions issue or improperly configured PKI?",
						zap.String("user", instance.Name),
						zap.Error(err))
					return ctrl.Result{
						Requeue:      true,
						RequeueAfter: interval,
					}, nil
				case errorfactory.VaultAPIFailure:
					// Same as above in terms of things that could be checked pre-flight on the cluster
					r.Log.Error("Vault API error attempting to reconcile the user certificate. If using vault perhaps a permissions issue or improperly configured PKI?",
						zap.String("user", instance.Name),
						zap.Error(err))
					return ctrl.Result{
						Requeue:      true,
						RequeueAfter: interval,
					}, nil
				default:
					return RequeueWithError(r.Log, "failed to reconcile secret for user "+instance.Name, err)
				}
			}

			r.Recorder.Event(instance, corev1.EventTypeNormal, "ReconciledCertificate",
				fmt.Sprintf("Reconciled certificate for nifi user %s", instance.Name))
		}
	}

	// Block the user creation in NiFi, if pure single user authentication
	if cluster.IsPureSingleUser() {
		r.Log.Debug("Cluster is in pure single user authentication, can't create user.",
			zap.String("user", instance.Name),
			zap.String("clusterName", clusterRef.Name))

		return RequeueAfter(interval)
	}

	// Generate the client configuration.
	clientConfig, err = configManager.BuildConfig()
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to create HTTP client for the referenced cluster : %s in %s",
				instance.Spec.ClusterRef.Name, clusterRef.Namespace))
		// the cluster is gone, so just remove the finalizer
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, fmt.Sprintf("failed to remove finalizer from NifiUser %s", instance.Name), err)
			}
			return Reconciled()
		}
		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to create HTTP client the for referenced cluster "+clusterRef.Name+" for user "+instance.Name, err)
	}

	// check if marked for deletion
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, clientConfig, patchInstance)
	}

	// Ensure the cluster is ready to receive actions
	if !clusterConnect.IsReady(r.Log) {
		r.Log.Debug("Cluster is not ready yet, will wait until it is.",
			zap.String("user", instance.Name),
			zap.String("clusterName", clusterRef.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, "ReferenceClusterNotReady",
			fmt.Sprintf("The referenced cluster is not ready yet: %s in %s",
				instance.Spec.ClusterRef.Name, clusterConnect.Id()))
		// the cluster does not exist - should have been caught pre-flight
		return RequeueAfter(interval)
	}

	// Ìn case of the cluster reference changed.
	if !v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
		// Delete the resource on the previous cluster.
		if err := usercli.RemoveUser(instance, clientConfig); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "RemoveError",
				fmt.Sprintf("Failed to delete NifiUser %s from cluster %s before moving in namespace %s",
					instance.Name, original.Spec.ClusterRef.Name, original.Spec.ClusterRef.Namespace))
			return RequeueWithError(r.Log, "Failed to delete NifiUser before moving into different namespace. user: "+instance.Name, err)
		}
		// Update the last view configuration to the current one.
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for user "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, current, patchCurrent); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiUser "+instance.Name, err)
		}
		return RequeueAfter(interval)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciling",
		fmt.Sprintf("Reconciling user %s", instance.Name))

	// Check if the NiFi user already exist
	exist, err := usercli.ExistUser(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing user "+instance.Name, err)
	}

	if !exist {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Creating",
			fmt.Sprintf("Creating user %s", instance.Name))

		var status *v1.NifiUserStatus

		status, err = usercli.FindUserByIdentity(instance, clientConfig)
		if err != nil {
			return RequeueWithError(r.Log, "failure finding user "+instance.Name, err)
		}

		if status == nil {
			// Create NiFi registry client
			status, err = usercli.CreateUser(instance, clientConfig)
			if err != nil {
				return RequeueWithError(r.Log, "failure creating user "+instance.Name, err)
			}
		}

		instance.Status = *status
		if err := r.updateStatus(ctx, instance, current.Status); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiUser "+instance.Name, err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Created",
			fmt.Sprintf("Created user %s", instance.Name))
	}

	// Sync user resource with NiFi side component
	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronizing",
		fmt.Sprintf("Synchronizing user %s", instance.Name))
	status, err := usercli.SyncUser(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failed to sync NifiUser "+instance.Name, err)
	}

	instance.Status = *status
	if err := r.updateStatus(ctx, instance, current.Status); err != nil {
		return RequeueWithError(r.Log, "failed to update status for NifiUser "+instance.Name, err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
		fmt.Sprintf("Synchronized user %s", instance.Name))

	// ensure a NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiCluster label on user "+current.Name, err)
	}

	// ensure a finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), userFinalizer) {
		r.addFinalizer(instance)
		if instance, err = r.updateAndFetchLatest(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update finalizer for NifiUser "+current.Name, err)
		}
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiUser "+current.Name, err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciled",
		fmt.Sprintf("Reconciling user %s", instance.Name))

	r.Log.Debug("Ensured user",
		zap.String("user", instance.Name))

	return RequeueAfter(interval)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiUserReconciler) SetupWithManager(mgr ctrl.Manager, certManagerEnabled bool) error {
	logCtr, err := GetLogConstructor(mgr, &v1.NifiUser{})
	if err != nil {
		return err
	}
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1.NifiUser{}).
		WithLogConstructor(logCtr).
		Owns(&corev1.Secret{})

	if certManagerEnabled {
		builder.Owns(&certv1.Certificate{})
	}

	return builder.Complete(r)
}

func (r *NifiUserReconciler) ensureClusterLabel(ctx context.Context, cluster clientconfig.ClusterConnect, user *v1.NifiUser, patcher client.Patch) (*v1.NifiUser, error) {
	labels := ApplyClusterReferenceLabel(cluster, user.GetLabels())
	if !reflect.DeepEqual(labels, user.GetLabels()) {
		user.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, user, patcher)
	}
	return user, nil
}

func (r *NifiUserReconciler) updateAndFetchLatest(ctx context.Context, user *v1.NifiUser, patcher client.Patch) (*v1.NifiUser, error) {
	typeMeta := user.TypeMeta
	err := r.Client.Patch(ctx, user, patcher)
	if err != nil {
		return nil, err
	}
	user.TypeMeta = typeMeta
	return user, nil
}

func (r *NifiUserReconciler) checkFinalizers(ctx context.Context, user *v1.NifiUser, config *clientconfig.NifiConfig, patcher client.Patch) (reconcile.Result, error) {
	r.Log.Info("NiFi user is marked for deletion. Removing finalizers.",
		zap.String("user", user.Name))
	var err error
	if util.StringSliceContains(user.GetFinalizers(), userFinalizer) {
		if err = r.finalizeNifiUser(user, config); err != nil {
			return RequeueWithError(r.Log, "failed to finalize nifiuser "+user.Name, err)
		}
		// remove finalizer
		if err = r.removeFinalizer(ctx, user, patcher); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from NifiUser "+user.Name, err)
		}
	}
	return Reconciled()
}

func (r *NifiUserReconciler) removeFinalizer(ctx context.Context, user *v1.NifiUser, patcher client.Patch) error {
	r.Log.Debug("Removing finalizer for NifiUser",
		zap.String("user", user.Name))
	user.SetFinalizers(util.StringSliceRemove(user.GetFinalizers(), userFinalizer))
	_, err := r.updateAndFetchLatest(ctx, user, patcher)
	return err
}

func (r *NifiUserReconciler) finalizeNifiUser(user *v1.NifiUser, config *clientconfig.NifiConfig) error {
	if err := usercli.RemoveUser(user, config); err != nil {
		return err
	}
	r.Log.Info("Deleted user",
		zap.String("user", user.Name))
	return nil
}

func (r *NifiUserReconciler) addFinalizer(user *v1.NifiUser) {
	r.Log.Debug("Adding Finalizer for the NifiUser",
		zap.String("user", user.Name))
	user.SetFinalizers(append(user.GetFinalizers(), userFinalizer))
}

func (r *NifiUserReconciler) updateStatus(ctx context.Context, user *v1.NifiUser, currentStatus v1.NifiUserStatus) error {
	if !reflect.DeepEqual(user.Status, currentStatus) {
		return r.Client.Status().Update(ctx, user)
	}
	return nil
}
