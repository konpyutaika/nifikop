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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/util"
)

// NifiConnectionReconciler reconciles a NifiConnection object
type NifiConnectionReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	RequeueInterval int
	RequeueOffset   int
}

//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificonnections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificonnections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nifi.konpyutaika.com,resources=nificonnections/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NifiConnection object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *NifiConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("nificonnection", req.NamespacedName)
	interval := util.GetRequeueInterval(r.RequeueInterval, r.RequeueOffset)
	var err error

	// Fetch the NifiConnection instance
	instance := &v1alpha1.NifiConnection{}
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
			return RequeueWithError(r.Log, "failed to update NifiConnection", err)
		}
		o, err = patch.DefaultAnnotator.GetOriginalConfiguration(instance)
	}

	// Validate component
	instance.Spec.Source.Namespace = GetComponentRefNamespace(instance.Namespace, instance.Spec.Source)
	if !instance.Spec.Source.IsValid() {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "SourceInvalid",
			fmt.Sprintf("Failed to validate the source component : %s in %s",
				instance.Spec.Source.Name, instance.Spec.Source.Namespace))
		return RequeueWithError(r.Log, "failed to validate source component", err)
	}

	instance.Spec.Destination.Namespace = GetComponentRefNamespace(instance.Namespace, instance.Spec.Destination)
	if !instance.Spec.Destination.IsValid() {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "DestinationInvalid",
			fmt.Sprintf("Failed to validate the destination component : %s in %s",
				instance.Spec.Destination.Name, instance.Spec.Destination.Namespace))
		return RequeueWithError(r.Log, "failed to validate destination component", err)
	}

	// LookUp component
	//var sourceComponent := r.GetDataflowComponentInformation(r.client, instance.Spec.Source)

	return RequeueAfter(interval)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiConnection{}).
		Complete(r)
}

func (r *NifiConnectionReconciler) GetDataflowComponentInformation(client client.Client, c *v1alpha1.ComponentReference) (*string, error) {
	dataflow, err := k8sutil.LookupNifiDataflow(r.Client, c.Name, c.Namespace)
	if err != nil {
		return nil, err
	} else {
		return &dataflow.Name, nil
	}
}
