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

package nifiparametercontext

import (
	"context"
	"reflect"
	"time"

	"emperror.dev/errors"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/parametercontext"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/Orange-OpenSource/nifikop/pkg/k8sutil"
	corev1 "k8s.io/api/core/v1"
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

var log = logf.Log.WithName("controller_nifiparametercontext")

var parameterContextFinalizer = "finalizer.nifiparametercontexts.nifi.orange.com"

// Add creates a new NifiParameterContext Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, namespaces []string) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNifiParameterContext{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nifiparametercontext-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NifiParameterContext
	err = c.Watch(&source.Kind{Type: &v1alpha1.NifiParameterContext{}}, &handler.EnqueueRequestForObject{})
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

// blank assignment to verify that ReconcileNifiParameterContext implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNifiParameterContext{}

// ReconcileNifiParameterContext reconciles a NifiParameterContext object
type ReconcileNifiParameterContext struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifiparametercontexts,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=nifi.orange.com,resources=nifiparametercontexts/status,verbs=get;update;patch

// Reconcile reads that state of the parameter context for a NifiParameterContext object and makes changes based on the state read
// and what is in the NifiParameterContext.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNifiParameterContext) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NifiParameterContext")
	var err error

	// Get a context for the request
	ctx := context.Background()

	// Fetch the NifiParameterContext instance
	instance := &v1alpha1.NifiParameterContext{}
	if err = r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return common.Reconciled()
		}
		// Error reading the object - requeue the request.
		return common.RequeueWithError(reqLogger, err.Error(), err)
	}

	// Get the referenced secrets
	var parameterSecrets []*corev1.Secret
	for _, parameterSecret := range instance.Spec.SecretRefs {
		secretNamespace := common.GetSecretRefNamespace(instance.Namespace, parameterSecret)
		var secret *corev1.Secret
		if secret, err = k8sutil.LookupSecret(r.client, parameterSecret.Name, secretNamespace); err != nil {
			// This shouldn't trigger anymore, but leaving it here as a safetybelt
			if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
				reqLogger.Info("Secret is already gone, there is nothing we can do")
				if err = r.removeFinalizer(ctx, instance); err != nil {
					return common.RequeueWithError(reqLogger, "failed to remove finalizer", err)
				}
				return common.Reconciled()
			}

			// the cluster does not exist - should have been caught pre-flight
			return common.RequeueWithError(reqLogger, "failed to lookup referenced secret", err)
		}
		parameterSecrets = append(parameterSecrets, secret)
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
		return r.checkFinalizers(ctx, reqLogger, instance, parameterSecrets, cluster)
	}

	// Check if the NiFi registry client already exist
	exist, err := parametercontext.ExistParameterContext(r.client, instance, cluster)
	if err != nil {
		return common.RequeueWithError(reqLogger, "failure checking for existing parameter context", err)
	}

	if !exist {
		// Create NiFi parameter context
		status, err := parametercontext.CreateParameterContext(r.client, instance, parameterSecrets, cluster)
		if err != nil {
			return common.RequeueWithError(reqLogger, "failure creating parameter context", err)
		}

		instance.Status = *status
		if err := r.client.Status().Update(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to update NifiParameterContext status", err)
		}
	}

	// Sync ParameterContext resource with NiFi side component
	status, err := parametercontext.SyncParameterContext(r.client, instance, parameterSecrets, cluster)
	if status != nil {
		instance.Status = *status
		if err := r.client.Status().Update(ctx, instance); err != nil {
			return common.RequeueWithError(reqLogger, "failed to update NifiParameterContext status", err)
		}
	}
	if err != nil {
		switch errors.Cause(err).(type) {
		case errorfactory.NifiParameterContextUpdateRequestRunning:
			return common.RequeueAfter(time.Duration(5) * time.Second)
		default:
			return common.RequeueWithError(reqLogger, "failed to sync NifiParameterContext", err)
		}
	}

	// Ensure NifiCluster label
	if instance, err = r.ensureClusterLabel(ctx, cluster, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to ensure NifiCluster label on parameter context", err)
	}

	// Ensure finalizer for cleanup on deletion
	if !util.StringSliceContains(instance.GetFinalizers(), parameterContextFinalizer) {
		reqLogger.Info("Adding Finalizer for NifiParameterContext")
		instance.SetFinalizers(append(instance.GetFinalizers(), parameterContextFinalizer))
	}

	// Push any changes
	if instance, err = r.updateAndFetchLatest(ctx, instance); err != nil {
		return common.RequeueWithError(reqLogger, "failed to update NifiParameterContext", err)
	}

	reqLogger.Info("Ensured Parameter Context")

	return common.RequeueAfter(time.Duration(15) * time.Second)
}

func (r *ReconcileNifiParameterContext) ensureClusterLabel(ctx context.Context, cluster *v1alpha1.NifiCluster,
	parameterContext *v1alpha1.NifiParameterContext) (*v1alpha1.NifiParameterContext, error) {

	labels := common.ApplyClusterRefLabel(cluster, parameterContext.GetLabels())
	if !reflect.DeepEqual(labels, parameterContext.GetLabels()) {
		parameterContext.SetLabels(labels)
		return r.updateAndFetchLatest(ctx, parameterContext)
	}
	return parameterContext, nil
}

func (r *ReconcileNifiParameterContext) updateAndFetchLatest(ctx context.Context,
	parameterContext *v1alpha1.NifiParameterContext) (*v1alpha1.NifiParameterContext, error) {

	typeMeta := parameterContext.TypeMeta
	err := r.client.Update(ctx, parameterContext)
	if err != nil {
		return nil, err
	}
	parameterContext.TypeMeta = typeMeta
	return parameterContext, nil
}

func (r *ReconcileNifiParameterContext) checkFinalizers(
	ctx context.Context,
	reqLogger logr.Logger,
	parameterContext *v1alpha1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	cluster *v1alpha1.NifiCluster) (reconcile.Result, error) {

	reqLogger.Info("NiFi parameter context is marked for deletion")
	var err error
	if util.StringSliceContains(parameterContext.GetFinalizers(), parameterContextFinalizer) {
		if err = r.finalizeNifiParameterContext(reqLogger, parameterContext, parameterSecrets, cluster); err != nil {
			return common.RequeueWithError(reqLogger, "failed to finalize parameter context", err)
		}
		if err = r.removeFinalizer(ctx, parameterContext); err != nil {
			return common.RequeueWithError(reqLogger, "failed to remove finalizer from parameter context", err)
		}
	}
	return common.Reconciled()
}

func (r *ReconcileNifiParameterContext) removeFinalizer(ctx context.Context, flow *v1alpha1.NifiParameterContext) error {
	flow.SetFinalizers(util.StringSliceRemove(flow.GetFinalizers(), parameterContextFinalizer))
	_, err := r.updateAndFetchLatest(ctx, flow)
	return err
}

func (r *ReconcileNifiParameterContext) finalizeNifiParameterContext(
	reqLogger logr.Logger,
	parameterContext *v1alpha1.NifiParameterContext,
	parameterSecrets []*corev1.Secret,
	cluster *v1alpha1.NifiCluster) error {

	if err := parametercontext.RemoveParameterContext(r.client, parameterContext, parameterSecrets, cluster); err != nil {
		return err
	}
	reqLogger.Info("Delete Registry client")

	return nil
}
