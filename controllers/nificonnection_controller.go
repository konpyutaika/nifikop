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

	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/erdrix/nigoapi/pkg/nifi"
	"github.com/go-logr/logr"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/dataflow"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
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

	// Verification connection possible
	// TO DO

	// LookUp component
	sourceComponent := &v1alpha1.ComponentInformation{}
	if sourceComponent, err = r.GetDataflowComponentInformation(instance.Spec.Source, true); err != nil {
		r.Log.Info("Error") // TO DO
	}

	destinationComponent := &v1alpha1.ComponentInformation{}
	if destinationComponent, err = r.GetDataflowComponentInformation(instance.Spec.Destination, false); err != nil {
		r.Log.Info("Error") // TO DO
	}
	r.Log.Info("Id: " + sourceComponent.Id)
	r.Log.Info("Id: " + destinationComponent.Id)

	return RequeueAfter(interval)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NifiConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NifiConnection{}).
		Complete(r)
}

func (r *NifiConnectionReconciler) GetDataflowComponentInformation(c v1alpha1.ComponentReference, isSource bool) (*v1alpha1.ComponentInformation, error) {
	instance, err := k8sutil.LookupNifiDataflow(r.Client, c.Name, c.Namespace)
	if err != nil {
		return nil, err
	} else {
		// Prepare cluster connection configurations
		var clientConfig *clientconfig.NifiConfig
		var clusterConnect clientconfig.ClusterConnect

		// Get the client config manager associated to the cluster ref.
		clusterRef := instance.Spec.ClusterRef
		clusterRef.Namespace = GetClusterRefNamespace(instance.Namespace, instance.Spec.ClusterRef)
		configManager := config.GetClientConfigManager(r.Client, clusterRef)

		// Generate the connect object
		if clusterConnect, err = configManager.BuildConnect(); err != nil {
			return nil, err
		}

		// Generate the client configuration.
		clientConfig, err = configManager.BuildConfig()
		if err != nil {
			return nil, err
		}

		// Ensure the cluster is ready to receive actions
		if !clusterConnect.IsReady(r.Log) {
			return nil, errors.New(fmt.Sprintf("Cluster %s in %s not ready for dataflow %s in %s", instance.Spec.ClusterRef.Name, clusterRef.Namespace, instance.Name, instance.Namespace))
		}

		dataflowInformation, err := dataflow.GetDataflowInformation(instance, clientConfig)
		if err != nil {
			return nil, err
		}

		// Error if the dataflow does not exist
		if dataflowInformation == nil {
			return nil, errors.New(fmt.Sprintf("Dataflow %s in %s does not exist in the cluster", instance.Name, instance.Namespace))
		}

		var ports = []nifi.PortEntity{}
		if isSource {
			ports = dataflowInformation.ProcessGroupFlow.Flow.OutputPorts
		} else {
			ports = dataflowInformation.ProcessGroupFlow.Flow.InputPorts
		}

		if len(ports) == 0 {
			return nil, errors.New(fmt.Sprintf("No port available for Dataflow %s in %s", instance.Name, instance.Namespace))
		}

		targetPort := &nifi.PortEntity{}
		foudTarget := false
		for _, port := range ports {
			if port.Component.Name == c.SubName {
				targetPort = &port
				foudTarget = true
			}
		}

		if !foudTarget {
			return nil, errors.New(fmt.Sprintf("Port %s not found : %s in %s", c.SubName, instance.Name, instance.Namespace))
		}

		information := &v1alpha1.ComponentInformation{
			Id:      targetPort.Id,
			Type:    targetPort.Component.Type_,
			GroupId: targetPort.Component.ParentGroupId,
		}
		return information, nil
	}
}
