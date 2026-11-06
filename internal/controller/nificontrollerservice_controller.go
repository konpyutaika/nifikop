package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

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
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/controllerservice"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var controllerServiceFinalizer = fmt.Sprintf("nificontrollerservices.%s/finalizer", v1alpha1.GroupVersion.Group)

// controllerServiceTransitionRequeue is the delay used while NiFi is enabling or disabling a
// controller service. Transitions take a few seconds at most, so waiting for the regular
// interval would needlessly delay the next lifecycle step.
const controllerServiceTransitionRequeue = 5 * time.Second

// NifiControllerServiceReconciler reconciles a NifiControllerService object.
type NifiControllerServiceReconciler struct {
	client.Client
	Log             zap.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificontrollerservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificontrollerservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificontrollerservices/finalizers,verbs=update

// Reconcile drives the NiFi controller service towards the state described by the
// NifiControllerService resource: creation, configuration sync, run state and deletion.
func (r *NifiControllerServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	interval := util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset)
	var err error

	// Fetch the NifiControllerService instance
	instance := &v1alpha1.NifiControllerService{}
	if err = r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return Reconciled()
		}
		// Error reading the object - requeue the request.
		return RequeueWithError(r.Log, err.Error(), err)
	}

	patchInstance := client.MergeFrom(instance.DeepCopy())
	// Get the last configuration viewed by the operator.
	o, _ := patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	// Create it if not exist.
	if o == nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for controller service "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiControllerService "+instance.Name, err)
		}
		o, _ = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Check if the cluster reference changed.
	original := &v1alpha1.NifiControllerService{}
	current := instance.DeepCopy()
	patchCurrent := client.MergeFrom(current.DeepCopy())
	if err := json.Unmarshal(o, original); err != nil {
		return RequeueWithError(r.Log, "could not decode last applied configuration of controller service "+instance.Name, err)
	}
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
		// The cluster is gone: there is nothing left to clean up in NiFi.
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			r.Log.Error("Cluster is already gone, there is nothing we can do",
				zap.String("controllerService", instance.Name),
				zap.String("clusterName", clusterRef.Name))
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer for controller service "+instance.Name, err)
			}
			return Reconciled()
		}
		// If the referenced cluster no more exist, just skip the deletion requirement in cluster ref change case.
		if !v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
				return RequeueWithError(r.Log, "could not apply last state to annotation to controller service "+instance.Name, err)
			}
			if err := r.Client.Patch(ctx, current, patchCurrent); err != nil {
				return RequeueWithError(r.Log, "failed to update NifiControllerService "+instance.Name, err)
			}
			return RequeueAfter(interval)
		}

		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to lookup reference cluster: %s in %s",
				instance.Spec.ClusterRef.Name, clusterRef.Namespace))
		return RequeueWithError(r.Log, "failed to lookup referenced cluster for controller service "+instance.Name, err)
	}

	// Generate the client configuration.
	clientConfig, err = configManager.BuildConfig()
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReferenceClusterError",
			fmt.Sprintf("Failed to create HTTP client for the referenced cluster: %s in %s",
				instance.Spec.ClusterRef.Name, clusterRef.Namespace))
		// the cluster is gone, so just remove the finalizer
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			if err = r.removeFinalizer(ctx, instance, patchInstance); err != nil {
				return RequeueWithError(r.Log, "failed to remove finalizer from NifiControllerService "+instance.Name, err)
			}
			return Reconciled()
		}
		return RequeueWithError(r.Log, "failed to create HTTP client the for referenced cluster "+clusterRef.Name+" for controller service "+instance.Name, err)
	}

	// Check if marked for deletion and if so run finalizers
	if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
		return r.checkFinalizers(ctx, instance, clientConfig, patchInstance, interval)
	}

	// Ensure the cluster is ready to receive actions
	if !clusterConnect.IsReady(r.Log) {
		r.Log.Debug("Cluster is not ready yet, will wait until it is.",
			zap.String("controllerService", instance.Name),
			zap.String("clusterName", clusterRef.Name))
		r.Recorder.Event(instance, corev1.EventTypeNormal, "ReferenceClusterNotReady",
			fmt.Sprintf("The referenced cluster is not ready yet: %s in %s",
				instance.Spec.ClusterRef.Name, clusterConnect.Id()))
		return RequeueAfter(interval)
	}

	// In case of the cluster reference changed.
	if !v1.ClusterRefsEquals([]v1.ClusterReference{instance.Spec.ClusterRef, current.Spec.ClusterRef}) {
		// Delete the resource on the previous cluster.
		if err := controllerservice.RemoveControllerService(instance, clientConfig); err != nil {
			if result, handled := r.handleRemoveError(instance, err, interval); handled {
				return result, nil
			}
			r.Recorder.Event(instance, corev1.EventTypeWarning, "RemoveError",
				fmt.Sprintf("Failed to delete NifiControllerService %s from cluster %s before moving in %s",
					instance.Name, original.Spec.ClusterRef.Name, current.Spec.ClusterRef.Name))
			return RequeueWithError(r.Log, "Failed to delete NifiControllerService before moving", err)
		}
		// Update the last view configuration to the current one.
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(current); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for controller service "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, current, patchCurrent); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiControllerService "+instance.Name, err)
		}
		return RequeueAfter(interval)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciling",
		"Reconciling controller service "+instance.Name)

	// Check if the NiFi controller service already exist
	exist, err := controllerservice.ExistControllerService(instance, clientConfig)
	if err != nil {
		return RequeueWithError(r.Log, "failure checking for existing controller service "+instance.Name, err)
	}

	if !exist {
		// Create NiFi controller service
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Creating",
			fmt.Sprintf("Creating controller service %s", instance.Name))
		status, err := controllerservice.CreateControllerService(instance, clientConfig)
		if err != nil {
			return RequeueWithError(r.Log, "failure creating controller service "+instance.Name, err)
		}

		instance.Status = *status
		if err := r.updateStatus(ctx, instance, current.Status); err != nil {
			return RequeueWithError(r.Log, "failed to update status for NifiControllerService "+instance.Name, err)
		}

		r.Recorder.Event(instance, corev1.EventTypeNormal, "Created",
			fmt.Sprintf("Created controller service %s", instance.Name))
		r.Log.Info("Created controller service",
			zap.String("controllerService", instance.Name))

		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(instance); err != nil {
			return RequeueWithError(r.Log, "could not apply last state to annotation for controller service "+instance.Name, err)
		}
		if err := r.Client.Patch(ctx, instance, patchInstance); err != nil {
			return RequeueWithError(r.Log, "failed to update NifiControllerService "+instance.Name, err)
		}
	}

	// Sync the controller service configuration and run state with the NiFi side component.
	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronizing",
		fmt.Sprintf("Synchronizing controller service %s", instance.Name))
	status, err := controllerservice.SyncControllerService(instance, clientConfig)
	if err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "SynchronizingFailed",
			fmt.Sprintf("Synchronizing controller service %s failed", instance.Name))
		return RequeueWithError(r.Log, "failed to sync NifiControllerService "+instance.Name, err)
	}

	instance.Status = *status
	if err := r.updateStatus(ctx, instance, current.Status); err != nil {
		return RequeueWithError(r.Log, "failed to update status for NifiControllerService "+instance.Name, err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Synchronized",
		fmt.Sprintf("Synchronized controller service %s", instance.Name))
	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, clusterConnect, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to ensure NifiCluster label on controller service "+current.Name, err)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), controllerServiceFinalizer) {
		r.Log.Debug("Adding Finalizer for NifiControllerService",
			zap.String("controllerService", instance.Name))
		instance.SetFinalizers(append(instance.GetFinalizers(), controllerServiceFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance, patchInstance); err != nil {
		return RequeueWithError(r.Log, "failed to update NifiControllerService "+current.Name, err)
	}

	// The service is only usable once ENABLED. While NiFi applies a run state change, or when the
	// wrapper has just requested one, come back quickly to complete the next lifecycle step.
	if instance.Status.State != v1alpha1.ControllerServiceStateEnabled {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "RunStateTransition",
			fmt.Sprintf("Controller service %s is %s, waiting for ENABLED", instance.Name, instance.Status.State))
		return RequeueAfter(controllerServiceTransitionRequeue)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciled",
		fmt.Sprintf("Reconciled controller service %s", instance.Name))

	r.Log.Debug("Ensured Controller Service",
		zap.String("controllerService", instance.Name))

	return RequeueAfter(interval)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiControllerServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logCtr, err := GetLogConstructor(mgr, &v1alpha1.NifiControllerService{})
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiControllerService{}).
		WithLogConstructor(logCtr).
		Complete(r)
}

// handleRemoveError maps the wrapper's deletion outcomes that are expected and transient to a
// requeue with an event, instead of an error. It reports whether the error was handled.
func (r *NifiControllerServiceReconciler) handleRemoveError(controllerService *v1alpha1.NifiControllerService,
	err error, interval time.Duration) (reconcile.Result, bool) {
	switch {
	case errors.Is(err, controllerservice.ErrReferenced):
		r.Recorder.Event(controllerService, corev1.EventTypeWarning, "StillReferenced",
			fmt.Sprintf("Controller service %s cannot be removed yet: %s", controllerService.Name, err.Error()))
		result, _ := RequeueAfter(interval)
		return result, true
	case errors.Is(err, controllerservice.ErrTransitioning):
		r.Recorder.Event(controllerService, corev1.EventTypeNormal, "RunStateTransition",
			fmt.Sprintf("Controller service %s is being disabled before removal", controllerService.Name))
		result, _ := RequeueAfter(controllerServiceTransitionRequeue)
		return result, true
	}
	return reconcile.Result{}, false
}

func (r *NifiControllerServiceReconciler) ensureClusterLabel(ctx context.Context, cluster clientconfig.ClusterConnect,
	controllerService *v1alpha1.NifiControllerService, patcher client.Patch) (*v1alpha1.NifiControllerService, error) {
	labels := ApplyClusterReferenceLabel(cluster, controllerService.GetLabels())
	if !reflect.DeepEqual(labels, controllerService.GetLabels()) {
		controllerService.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, controllerService, patcher)
	}
	return controllerService, nil
}

func (r *NifiControllerServiceReconciler) updateAndFetchLatest(ctx context.Context,
	controllerService *v1alpha1.NifiControllerService, patcher client.Patch) (*v1alpha1.NifiControllerService, error) {
	typeMeta := controllerService.TypeMeta
	err := r.Client.Patch(ctx, controllerService, patcher)
	if err != nil {
		return nil, err
	}
	controllerService.TypeMeta = typeMeta
	return controllerService, nil
}

func (r *NifiControllerServiceReconciler) checkFinalizers(ctx context.Context,
	controllerService *v1alpha1.NifiControllerService, config *clientconfig.NifiConfig, patcher client.Patch,
	interval time.Duration) (reconcile.Result, error) {
	r.Log.Info("NiFi controller service is marked for deletion. Removing finalizers.",
		zap.String("controllerService", controllerService.Name))
	var err error
	if util.StringSliceContains(controllerService.GetFinalizers(), controllerServiceFinalizer) {
		if err = r.finalizeNifiControllerService(controllerService, config); err != nil {
			if result, handled := r.handleRemoveError(controllerService, err, interval); handled {
				return result, nil
			}
			return RequeueWithError(r.Log, "failed to finalize NifiControllerService", err)
		}
		if err = r.removeFinalizer(ctx, controllerService, patcher); err != nil {
			return RequeueWithError(r.Log, "failed to remove finalizer from NifiControllerService", err)
		}
	}
	return Reconciled()
}

func (r *NifiControllerServiceReconciler) removeFinalizer(ctx context.Context, controllerService *v1alpha1.NifiControllerService, patcher client.Patch) error {
	r.Log.Debug("Removing finalizer for NifiControllerService",
		zap.String("controllerService", controllerService.Name))
	controllerService.SetFinalizers(util.StringSliceRemove(controllerService.GetFinalizers(), controllerServiceFinalizer))
	_, err := r.updateAndFetchLatest(ctx, controllerService, patcher)
	return err
}

func (r *NifiControllerServiceReconciler) finalizeNifiControllerService(controllerService *v1alpha1.NifiControllerService,
	config *clientconfig.NifiConfig) error {
	if err := controllerservice.RemoveControllerService(controllerService, config); err != nil {
		return err
	}

	r.Log.Info("Deleted controller service",
		zap.String("controllerService", controllerService.Name))

	return nil
}

func (r *NifiControllerServiceReconciler) updateStatus(ctx context.Context, controllerService *v1alpha1.NifiControllerService, currentStatus v1alpha1.NifiControllerServiceStatus) error {
	if !reflect.DeepEqual(controllerService.Status, currentStatus) {
		return r.Client.Status().Update(ctx, controllerService)
	}
	return nil
}
