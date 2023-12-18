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
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/usergroup"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var userGroupFinalizer = fmt.Sprintf("nifiusergroups.%s/finalizer", v1.GroupVersion.Group)

// NifiUserGroupReconciler reconciles a NifiUserGroup object.
type NifiUserGroupReconciler struct {
	client.Client
	Log             zap.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiusergroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiusergroups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifiusergroups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiUserGroup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *NifiUserGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	interval := util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset)
	var err error

	// Fetch the NifiUserGroup instance
	instance := &v1.NifiUserGroup{}
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
			return RequeueWithError(r.Log, "could not apply last state to annotation for user group "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiUserGroup "+instance.Name, err)
		}
		o, _ = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Check if the cluster reference changed.
	original := &v1.NifiUserGroup{}
	current := instance.DeepCopy()
	patchCurrent := client.MergeFromWithOptions(current.DeepCopy(), client.MergeFromWithOptimisticLock{})
	json.Unmarshal(o, original)
	if !v1.ClusterRefsEquals([]v1.ClusterReference{original.Spec.ClusterRef, instance.Spec.ClusterRef}) {
		instance.Spec.ClusterRef = original.Spec.ClusterRef
	}

	// Ensure the cluster ref consistency with all users
	var users []*v1.NifiUser
	for _, userRef := range instance.Spec.UsersRef {
		var user *v1.NifiUser
		userNamespace := GetUserRefNamespace(current.Namespace, userRef)
		if user, err = k8sutil.LookupNifiUser(r.Client, userRef.Name, userNamespace); err != nil {
			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(current.ObjectMeta) {
				r.Log.Error("User group is already gone, there is nothing we can do",
					zap.String("userGroup", instance.Name))
				if err = r.removeFinalizer(ctx, current, patchCurrent); err != nil {
					return RequeueWithError(r.Log, "failed to remove finalizer for user group "+instance.Name, err)
				}
				return Reconciled()
			}

			r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceUserError",
				fmt.Sprintf("Failed to lookup reference user : %s in %s",
					userRef.Name, userNamespace))

			// the cluster does not exist - should have been caught pre-flight
			return RequeueWithError(r.Log, "failed to lookup referenced user "+user.Name+" in group "+instance.Name, err)
		}

		// Check if cluster references are the same
		clusterNamespace := GetClusterRefNamespace(current.Namespace, current.Spec.ClusterRef)
		if user != nil && (userNamespace != clusterNamespace || user.Spec.ClusterRef.Name != current.Spec.ClusterRef.Name) {
			msg := fmt.Sprintf("Failed to ensure consistency in cluster referece : %s in %s, with user : %s in %s",
				instance.Spec.ClusterRef.Name, clusterNamespace, userRef.Name, userRef.Namespace)
			r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError", msg)
			return RequeueWithError(r.Log, msg, errors.New("inconsistent cluster references"))
		}

		users = append(users, user)
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
			r.Log.Error("Cluster is already gone, there is nothing we can do",
				zap.String("userGroup", instance.Name),
				zap.String("clusterName", clusterRef.Name))
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer for user group "+instance.Name, err)
			}
			return Reconciled()
		}

		// If the referenced cluster no more exist, just skip the deletion requirement in cluster ref change case.
		if !v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
				return RequeueWithError(r.Log, "could not apply last state to annotation for user group "+instance.Name, err)
			}
			if err := r.Client.Patch(ctx, current, patchCurrent); err != nil {
				return RequeueWithError(r.Log, "failed to update NifiUserGroup "+instance.Name, err)
			}
			return RequeueAfter(interval)
		}

		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to lookup reference cluster : %s in %s",
				instance.Spec.ClusterRef.Name, clusterRef.Namespace))

		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to lookup referenced cluster for user group "+instance.Name, err)
	}

	// Get the referenced NifiCluster
	var cluster *v1.NifiCluster
	if cluster, err = k8sutil.LookupNifiCluster(r.Client, instance.Spec.ClusterRef.Name, clusterRef.Namespace); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			r.Log.Error("Cluster is gone already, there is nothing we can do",
				zap.String("userGroup", instance.Name),
				zap.String("clusterName", clusterRef.Name))
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer from NifiUserGroup "+instance.Name, err)
			}
			return Reconciled()
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
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, fmt.Sprintf("failed to remove finalizer from NifiUserGroup %s", instance.Name), err)
			}
			return Reconciled()
		}
		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to create HTTP client the for referenced cluster "+clusterRef.Name+" for user group "+instance.Name, err)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, users, clientConfig, patchInstance)
	}

	// Ensure the cluster is ready to receive actions
	if !clusterConnect.IsReady(r.Log) {
		r.Log.Debug("Cluster is not ready yet, will wait until it is.",
			zap.String("userGroup", instance.Name),
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
		if err := usergroup.RemoveUserGroup(instance, users, clientConfig); err != nil {
			msg := fmt.Sprintf("Failed to delete NifiUserGroup %s from cluster %s before moving in %s",
				instance.Name, original.Spec.ClusterRef.Name, original.Spec.ClusterRef.Namespace)
			r.Recorder.Event(instance, corev1.EventTypeWarning, "RemoveError", msg)
			return RequeueWithError(r.Log, msg, err)
		}
		// Update the last view configuration to the current one.
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for user group "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, current, patchCurrent); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiUserGroup "+instance.Name, err)
		}
		return RequeueAfter(interval)
	}

	// Block the userGroup creation in NiFi, if pure single user authentication
	if cluster.IsPureSingleUser() {
		r.Log.Debug("Cluster is in pure single user authentication, can't create user group.",
			zap.String("userGroup", instance.Name),
			zap.String("clusterName", clusterRef.Name))

		return RequeueAfter(interval)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciling",
		fmt.Sprintf("Reconciling user group %s", instance.Name))

	// Check if the NiFi user group already exist
	exist, err := usergroup.ExistUserGroup(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing user group "+instance.Name, err)
	}

	if !exist {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Creating",
			fmt.Sprintf("Creating registry client %s", instance.Name))

		// Create NiFi user group
		status, err := usergroup.CreateUserGroup(instance, users, clientConfig)
		if err != nil {
			return RequeueWithError(r.Log, "failure creating user group "+instance.Name, err)
		}

		instance.Status = *status
		if err := r.updateStatus(ctx, instance, current.Status); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiUserGroup "+instance.Name, err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Created",
			fmt.Sprintf("Created user group %s", instance.Name))
	}

	// Sync UserGroup resource with NiFi side component
	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronizing",
		fmt.Sprintf("Synchronizing user group %s", instance.Name))
	status, err := usergroup.SyncUserGroup(instance, users, clientConfig)
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "SynchronizingFailed",
			fmt.Sprintf("Synchronizing user group %s failed", instance.Name))
		return RequeueWithError(r.Log, "failed to sync NifiUserGroup "+instance.Name, err)
	}

	instance.Status = *status
	if err := r.updateStatus(ctx, instance, current.Status); err != nil {
		return RequeueWithError(r.Log, "failed to update status for NifiUserGroup "+instance.Name, err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
		fmt.Sprintf("Synchronized user group %s", instance.Name))

	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiCluster label on user group "+current.Name, err)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), userGroupFinalizer) {
		r.Log.Info("Adding Finalizer for NifiUserGroup",
			zap.String("userGroup", instance.Name))
		instance.SetFinalizers(append(instance.GetFinalizers(), userGroupFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiUserGroup "+current.Name, err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciled",
		fmt.Sprintf("Reconciling user group %s", instance.Name))

	r.Log.Debug("Ensured User Group",
		zap.String("userGroup", instance.Name))

	return RequeueAfter(interval)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiUserGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logCtr, err := GetLogConstructor(mgr, &v1.NifiUserGroup{})
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.NifiUserGroup{}).
		WithLogConstructor(logCtr).
		Complete(r)
}

func (r *NifiUserGroupReconciler) ensureClusterLabel(ctx context.Context, cluster clientconfig.ClusterConnect,
	userGroup *v1.NifiUserGroup, patcher client.Patch) (*v1.NifiUserGroup, error) {
	labels := ApplyClusterReferenceLabel(cluster, userGroup.GetLabels())
	if !reflect.DeepEqual(labels, userGroup.GetLabels()) {
		userGroup.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, userGroup, patcher)
	}
	return userGroup, nil
}

func (r *NifiUserGroupReconciler) updateAndFetchLatest(ctx context.Context,
	userGroup *v1.NifiUserGroup, patcher client.Patch) (*v1.NifiUserGroup, error) {
	typeMeta := userGroup.TypeMeta
	err := r.Client.Patch(ctx, userGroup, patcher)
	if err != nil {
		return nil, err
	}
	userGroup.TypeMeta = typeMeta
	return userGroup, nil
}

func (r *NifiUserGroupReconciler) checkFinalizers(ctx context.Context, userGroup *v1.NifiUserGroup,
	users []*v1.NifiUser, config *clientconfig.NifiConfig, patcher client.Patch) (reconcile.Result, error) {
	r.Log.Info("NiFi user group is marked for deletion. Removing finalizers.",
		zap.String("userGroup", userGroup.Name))
	var err error
	if util.StringSliceContains(userGroup.GetFinalizers(), userGroupFinalizer) {
		if err = r.finalizeNifiNifiUserGroup(userGroup, users, config); err != nil {
			return RequeueWithError(r.Log, "failed to finalize nifiusergroup "+userGroup.Name, err)
		}
		if err = r.removeFinalizer(ctx, userGroup, patcher); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from user group"+userGroup.Name, err)
		}
	}
	return Reconciled()
}

func (r *NifiUserGroupReconciler) removeFinalizer(ctx context.Context, userGroup *v1.NifiUserGroup, patcher client.Patch) error {
	r.Log.Debug("Removing finalizer for NifiUserGroup",
		zap.String("userGroup", userGroup.Name))
	userGroup.SetFinalizers(util.StringSliceRemove(userGroup.GetFinalizers(), userGroupFinalizer))
	_, err := r.updateAndFetchLatest(ctx, userGroup, patcher)
	return err
}

func (r *NifiUserGroupReconciler) finalizeNifiNifiUserGroup(
	userGroup *v1.NifiUserGroup,
	users []*v1.NifiUser,
	config *clientconfig.NifiConfig) error {
	if err := usergroup.RemoveUserGroup(userGroup, users, config); err != nil {
		return err
	}

	r.Log.Info("Deleted NifiUserGroup",
		zap.String("userGroup", userGroup.Name))

	return nil
}

func (r *NifiUserGroupReconciler) updateStatus(ctx context.Context, userGroup *v1.NifiUserGroup, currentStatus v1.NifiUserGroupStatus) error {
	if !reflect.DeepEqual(userGroup.Status, currentStatus) {
		return r.Client.Status().Update(ctx, userGroup)
	}
	return nil
}
