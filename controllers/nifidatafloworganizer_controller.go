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

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/datafloworganizer"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

var dataflowOrganizerFinalizer = "nifidatafloworganizers.nifi.konpyutaika.com/finalizer"

// NifiDataflowOrganizerReconciler reconciles a NifiDataflowOrganizer object
type NifiDataflowOrganizerReconciler struct {
	client.Client
	Log             zap.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifidatafloworganizers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifidatafloworganizers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nifidatafloworganizers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiDataflowOrganizer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NifiDataflowOrganizerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	interval := util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset)
	var err error

	// Fetch the NifiDataflowOrganizer instance
	instance := &v1alpha1.NifiDataflowOrganizer{}
	if err = r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return Reconciled()
		}
		// Error reading the object - requeue the request.
		return RequeueWithError(r.Log, err.Error(), err)
	}

	// Get the last configuration viewed by the operator.
	o, _ := patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	// Create it if not exist.
	if o == nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for NifiDataflowOrganizer "+instance.Name, err)
		}
		if err := r.Client.Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDataflowOrganizer "+instance.Name, err)
		}
		o, _ = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Check if the cluster reference changed.
	original := &v1alpha1.NifiDataflowOrganizer{}
	current := instance.DeepCopy()
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
			r.Log.Error("Cluster is already gone, there is nothing we can do",
				zap.String("clusterName", clusterRef.Name),
				zap.String("dataflowOrganizer", instance.Name))
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer for NifiDataflowOrganizer "+instance.Name, err)
			}
			return Reconciled()
		}
		// If the referenced cluster no more exist, just skip the deletion requirement in cluster ref change case.
		if !v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
				return RequeueWithError(r.Log, "could not apply last state to annotation for NifiDataflowOrganizer "+instance.Name, err)
			}
			if err := r.Client.Update(ctx, current); err != nil {
				return RequeueWithError(r.Log, "failed to update NifiDataflowOrganizer "+instance.Name, err)
			}
			return RequeueAfter(interval)
		}

		msg := fmt.Sprintf("Failed to lookup reference cluster for NifiDataflowOrganizer %s : %s in %s",
			instance.Name, instance.Spec.ClusterRef.Name, clusterRef.Namespace)
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError", msg)

		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, msg, err)
	}

	// Generate the client configuration.
	clientConfig, err = configManager.BuildConfig()
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to create HTTP client for the referenced cluster for NifiDataflowOrganizer %s : %s in %s",
				instance.Name, instance.Spec.ClusterRef.Name, clusterRef.Namespace))
		// the cluster is gone, so just remove the finalizer
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			if err = r.removeFinalizer(ctx, instance); err != nil {
				return RequeueWithError(r.Log, fmt.Sprintf("failed to remove finalizer from NifiDataflowOrganizer %s", instance.Name), err)
			}
			return Reconciled()
		}
		// the cluster does not exist - should have been caught pre-flight
		return RequeueWithError(r.Log, "failed to create HTTP client the for referenced cluster", err)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, clientConfig)
	}

	// Ensure the cluster is ready to receive actions
	if !clusterConnect.IsReady(r.Log) {
		r.Log.Debug("Cluster is not ready yet, will wait until it is.",
			zap.String("clusterName", clusterRef.Name),
			zap.String("dataflowOrganizer", instance.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, "ReferenceClusterNotReady",
			fmt.Sprintf("The referenced cluster is not ready yet : %s in %s",
				instance.Spec.ClusterRef.Name, clusterConnect.Id()))

		// the cluster does not exist - should have been caught pre-flight
		return RequeueAfter(interval)
	}

	// Ìn case of the cluster reference changed.
	if !v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
		// Delete the resource on the previous cluster.
		for groupName := range instance.Spec.Groups {
			groupStatus := instance.Status.GroupStatus[groupName]
			if err := datafloworganizer.RemoveDataflowOrganizerGroup(groupStatus, clientConfig); err != nil {
				r.Recorder.Event(instance, corev1.EventTypeWarning, "RemoveError",
					fmt.Sprintf("Failed to delete group %s of NifiDataflowOrganizer %s from cluster %s before moving in %s",
						groupName, instance.Name, original.Spec.ClusterRef.Name, original.Spec.ClusterRef.Name))
				return RequeueWithError(r.Log, "Failed to delete group "+groupName+" NifiDataflowOrganizer before moving "+instance.Name, err)
			}
		}

		// Update the last view configuration to the current one.
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for NifiDataflowOrganizer "+instance.Name, err)
		}
		if err := r.Client.Update(ctx, current); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiDataflowOrganizer "+instance.Name, err)
		}
		return RequeueAfter(interval)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciling",
		fmt.Sprintf("Reconciling NifiDataflowOrganizer %s", instance.Name))

	if instance.Status.GroupStatus == nil {
		instance.Status.GroupStatus = make(map[string]v1alpha1.OrganizerGroupStatus)
	}

	// Check if some groups have been deleted
	for _, groupName := range original.Spec.GetGroupNames() {
		if _, ok := instance.Spec.Groups[groupName]; !ok {
			groupStatus := instance.Status.GroupStatus[groupName]
			if err := datafloworganizer.RemoveDataflowOrganizerGroup(groupStatus, clientConfig); err != nil {
				r.Recorder.Event(instance, corev1.EventTypeWarning, "RemoveGroupError",
					fmt.Sprintf("Failed to delete group %s of NifiDataflowOrganizer %s",
						groupName, instance.Name))
				return RequeueWithError(r.Log, "Failed to delete group "+groupName+" NifiDataflowOrganizer "+instance.Name, err)
			}

			// TODO FIX THIS
			delete(instance.Status.GroupStatus, groupName)
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
				return RequeueWithError(r.Log, "could not apply last state to annotation for NifiDataflowOrganizer "+instance.Name, err)
			}
			if err := r.Client.Status().Update(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to update status for NifiDataflowOrganizer "+instance.Name, err)
			}
			if err := r.Client.Update(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to update NifiDataflowOrganizer "+instance.Name, err)
			}
			r.Recorder.Event(instance, corev1.EventTypeNormal, "RemovedGroup",
				fmt.Sprintf("Removed group %s of NifiDataflowOrganizer %s", groupName, instance.Name))
			return RequeueAfter(interval)
		} else {
			for _, dataflowRefOrginal := range original.Spec.Groups[groupName].DataflowRef {
				notFound := true
				for _, dataflowRefInstance := range instance.Spec.Groups[groupName].DataflowRef {
					if dataflowRefOrginal.Name == dataflowRefInstance.Name &&
						GetDataflowRefNamespace(original.Namespace, dataflowRefOrginal) == GetDataflowRefNamespace(instance.Namespace, dataflowRefInstance) {
						notFound = false
						break
					}
				}

				if notFound {
					dfInstance, err := k8sutil.LookupNifiDataflow(r.Client, dataflowRefOrginal.Name, GetDataflowRefNamespace(original.Namespace, dataflowRefOrginal))
					if err != nil {
						return RequeueWithError(r.Log, "failed to lookup NifiDataflow "+dataflowRefOrginal.Name, err)
					}
					dfInstanceOriginal := dfInstance.DeepCopy()

					dfInstance.Spec.FlowPosition = nil

					if err := r.Client.Patch(ctx, dfInstance, client.MergeFrom(dfInstanceOriginal)); err != nil {
						return RequeueWithError(r.Log, "failed to patch NifiDataflow "+dfInstance.Name, err)
					}
				}
			}
		}
	}

	// TODO IF TITLE WIDER THAT CONTENT MAKE CONTENT AT LEAST AS WIDE AS THE TITLE
	groupPosX, groupPosY := float64(instance.Spec.InitialPosition.X), float64(instance.Spec.InitialPosition.Y)
	// Check if the NiFi dataflow organizer resources already exist
	for _, groupName := range instance.Spec.GetGroupNames() {
		group := instance.Spec.Groups[groupName]
		groupStatus := instance.Status.GroupStatus[groupName]
		exist, err := datafloworganizer.ExistDataflowOrganizerGroup(group, groupStatus, clientConfig)
		if err != nil {
			return RequeueWithError(r.Log, "failure checking for existing NifiDataflowOrganizer with name "+instance.Name, err)
		}

		if !exist {
			// Create NiFi dataflow organizer resources
			r.Recorder.Event(instance, corev1.EventTypeNormal, "Creating",
				fmt.Sprintf("Creating group %s of NifiDataflowOrganizer %s", groupName, instance.Name))

			status, err := datafloworganizer.CreateDataflowOrganizerGroup(groupPosX, groupPosY, groupName, group, groupStatus, clientConfig)
			if err != nil {
				r.Recorder.Event(instance, corev1.EventTypeWarning, "CreationFailed",
					fmt.Sprintf("Creation failed group %s of NifiDataflowOrganizer %s",
						groupName,
						instance.Name))
				return RequeueWithError(r.Log, "failure creating group "+groupName+" of NifiDataflowOrganizer "+instance.Name, err)
			}

			groupStatus = *status
			instance.Status.GroupStatus[groupName] = groupStatus
			if err := r.Client.Status().Update(ctx, instance); err != nil {
				return RequeueWithError(r.Log, "failed to update status for NifiDataflowOrganizer "+instance.Name, err)
			}
			r.Recorder.Event(instance, corev1.EventTypeNormal, "Created",
				fmt.Sprintf("Created group %s of NifiDataflowOrganizer %s", groupName, instance.Name))
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronizing",
			fmt.Sprintf("Synchronizing group %s of NifiDataflowOrganizer %s", groupName, instance.Name))
		status, err := datafloworganizer.SyncDataflowOrganizerGroup(groupPosX, groupPosY, groupName, group, groupStatus, clientConfig)
		if err != nil {
			return RequeueWithError(r.Log, "failed to sync group "+groupName+" of NifiDataflowOrganizer "+instance.Name, err)
		}

		instance.Status.GroupStatus[groupName] = *status
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiDataflowOrganizer "+instance.Name, err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
			fmt.Sprintf("Synchronized group %s of NifiDataflowOrganizer %s", groupName, instance.Name))

		dfOffsetX, dfOffsetY := nifiutil.ProcessGroupPadding+int(groupPosX), nifiutil.ProcessGroupPadding+int(groupPosY)+int(group.GetTitleHeight(groupName))
		for index, dfRef := range group.DataflowRef {
			dfInstance, err := k8sutil.LookupNifiDataflow(r.Client, dfRef.Name, GetDataflowRefNamespace(instance.Namespace, dfRef))
			if err != nil {
				return RequeueWithError(r.Log, "failed to lookup NifiDataflow "+dfRef.Name, err)
			}
			dfInstanceOriginal := dfInstance.DeepCopy()

			if index > 0 && index%group.MaxColumnSize == 0 {
				dfOffsetX = nifiutil.ProcessGroupPadding + int(groupPosX)
				dfOffsetY = dfOffsetY + nifiutil.ProcessGroupPadding + nifiutil.ProcessGroupHeight
			}

			dfInstance.Spec.FlowPosition = &v1.FlowPosition{
				X: util.Int64Pointer(int64((index%group.MaxColumnSize)*nifiutil.ProcessGroupWidth + dfOffsetX)),
				Y: util.Int64Pointer(int64(dfOffsetY)),
			}

			if err := r.Client.Patch(ctx, dfInstance, client.MergeFrom(dfInstanceOriginal)); err != nil {
				return RequeueWithError(r.Log, "failed to patch NifiDataflow "+dfInstance.Name, err)
			}
			dfOffsetX = dfOffsetX + nifiutil.ProcessGroupPadding
		}

		if groupPosX > float64(instance.Spec.MaxWidth) {
			groupPosX, groupPosY = 0.0, groupPosY+group.GetContentHeight()+group.GetTitleHeight(groupName)
		} else {
			groupPosX = groupPosX + group.GetContentWidth()
		}
	}

	// ensure a NifiDataflowOrganizer label
	if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiDataflowOrganizer label on "+current.Name, err)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), dataflowOrganizerFinalizer) {
		r.Log.Info("Adding Finalizer for NifiDataflowOrganizer " + instance.Name)
		instance.SetFinalizers(append(instance.GetFinalizers(), dataflowOrganizerFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiDataflowOrganizer "+current.Name, err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciled",
		fmt.Sprintf("Reconciling NifiDataflowOrganizer %s", instance.Name))

	r.Log.Debug("Ensured NifiDataflowOrganizer",
		zap.String("dataflowOrganizer", instance.Name))

	return RequeueAfter(interval)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiDataflowOrganizerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logCtr, err := GetLogConstructor(mgr, &v1alpha1.NifiDataflowOrganizer{})
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiDataflowOrganizer{}).
		WithLogConstructor(logCtr).
		Complete(r)
}

func (r *NifiDataflowOrganizerReconciler) ensureClusterLabel(ctx context.Context, cluster clientconfig.ClusterConnect,
	dataflowOrganizer *v1alpha1.NifiDataflowOrganizer) (*v1alpha1.NifiDataflowOrganizer, error) {

	labels := ApplyClusterReferenceLabel(cluster, dataflowOrganizer.GetLabels())
	if !reflect.DeepEqual(labels, dataflowOrganizer.GetLabels()) {
		dataflowOrganizer.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, dataflowOrganizer)
	}
	return dataflowOrganizer, nil
}
func (r *NifiDataflowOrganizerReconciler) updateAndFetchLatest(ctx context.Context,
	dataflowOrganizer *v1alpha1.NifiDataflowOrganizer) (*v1alpha1.NifiDataflowOrganizer, error) {

	typeMeta := dataflowOrganizer.TypeMeta
	err := r.Client.Update(ctx, dataflowOrganizer)
	if err != nil {
		return nil, err
	}
	dataflowOrganizer.TypeMeta = typeMeta
	return dataflowOrganizer, nil
}

func (r *NifiDataflowOrganizerReconciler) checkFinalizers(
	ctx context.Context,
	dataflowOrganizer *v1alpha1.NifiDataflowOrganizer,
	config *clientconfig.NifiConfig) (reconcile.Result, error) {
	r.Log.Info("NifiDataflowOrganizer is marked for deletion. Removing finalizers.",
		zap.String("dataflowOrganizer", dataflowOrganizer.Name))
	var err error
	if util.StringSliceContains(dataflowOrganizer.GetFinalizers(), dataflowOrganizerFinalizer) {
		if err = r.finalizeNifiDataflowOrganizer(ctx, dataflowOrganizer, config); err != nil {
			return RequeueWithError(r.Log, "failed to finalize NifiDataflowOrganizer "+dataflowOrganizer.Name, err)
		}
		if err = r.removeFinalizer(ctx, dataflowOrganizer); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from NifiDataflowOrganizer "+dataflowOrganizer.Name, err)
		}
	}
	return Reconciled()
}
func (r *NifiDataflowOrganizerReconciler) removeFinalizer(
	ctx context.Context,
	dataflowOrganizer *v1alpha1.NifiDataflowOrganizer) error {
	r.Log.Debug("Removing finalizer for NifiDataflowOrganizer",
		zap.String("dataflowOrganizer", dataflowOrganizer.Name))
	dataflowOrganizer.SetFinalizers(util.StringSliceRemove(dataflowOrganizer.GetFinalizers(), dataflowOrganizerFinalizer))
	_, err := r.updateAndFetchLatest(ctx, dataflowOrganizer)
	return err
}

func (r *NifiDataflowOrganizerReconciler) finalizeNifiDataflowOrganizer(
	ctx context.Context,
	dataflowOrganizer *v1alpha1.NifiDataflowOrganizer,
	config *clientconfig.NifiConfig) error {
	for groupName := range dataflowOrganizer.Spec.Groups {
		groupStatus := dataflowOrganizer.Status.GroupStatus[groupName]

		for _, dfRef := range dataflowOrganizer.Spec.Groups[groupName].DataflowRef {
			dfInstance, err := k8sutil.LookupNifiDataflow(r.Client, dfRef.Name, GetDataflowRefNamespace(dataflowOrganizer.Namespace, dfRef))
			if err != nil {
				return err
			}
			dfInstanceOriginal := dfInstance.DeepCopy()

			dfInstance.Spec.FlowPosition = nil

			if err := r.Client.Patch(ctx, dfInstance, client.MergeFrom(dfInstanceOriginal)); err != nil {
				return err
			}
		}

		if err := datafloworganizer.RemoveDataflowOrganizerGroup(groupStatus, config); err != nil {
			return err
		}
	}

	r.Log.Info("Deleted NifiDataflowOrganizer",
		zap.String("dataflowOrganizer", dataflowOrganizer.Name))

	return nil
}
