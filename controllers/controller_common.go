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

package controllers

import (
	"fmt"
	"github.com/Orange-OpenSource/nifikop/pkg/util/clientconfig"
	"time"

	"emperror.dev/errors"
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// clusterRefLabel is the label key used for referencing NifiUsers/NifiDataflow
// to a NifiCluster

var ClusterRefLabel = "nifiCluster"

// requeueWithError is a convenience wrapper around logging an error message
// separate from the stacktrace and then passing the error through to the controller
// manager
func RequeueWithError(logger logr.Logger, msg string, err error) (reconcile.Result, error) {
	// Info log the error message and then let the reconciler dump the stacktrace
	logger.Info(msg)
	return reconcile.Result{}, err
}

func Requeue() (reconcile.Result, error) {
	return reconcile.Result{Requeue: true}, nil
}

func RequeueAfter(time time.Duration) (reconcile.Result, error) {
	return reconcile.Result{Requeue: true, RequeueAfter: time}, nil
}

// reconciled returns an empty result with nil error to signal a successful reconcile
// to the controller manager
func Reconciled() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

// clusterLabelString returns the label value for a cluster reference
func ClusterLabelString(cluster *v1alpha1.NifiCluster) string {
	return fmt.Sprintf("%s.%s", cluster.Name, cluster.Namespace)
}

// checkNodeConnectionError is a convenience wrapper for returning from common
// node connection errors
func CheckNodeConnectionError(logger logr.Logger, err error) (ctrl.Result, error) {
	switch errors.Cause(err).(type) {
	case errorfactory.NodesUnreachable:
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Duration(15) * time.Second,
		}, nil
	case errorfactory.NodesNotReady:
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Duration(15) * time.Second,
		}, nil
	case errorfactory.ResourceNotReady:
		logger.Info("Needed resource for node connection not found, may not be ready")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Duration(5) * time.Second,
		}, nil
	default:
		return RequeueWithError(logger, err.Error(), err)
	}
}

// applyClusterRefLabel ensures a map of labels contains a reference to a parent nifi cluster
func ApplyClusterRefLabel(cluster *v1alpha1.NifiCluster, labels map[string]string) map[string]string {
	labelValue := ClusterLabelString(cluster)
	if labels == nil {
		labels = make(map[string]string, 0)
	}
	if label, ok := labels[ClusterRefLabel]; ok {
		if label != labelValue {
			labels[ClusterRefLabel] = labelValue
		}
	} else {
		labels[ClusterRefLabel] = labelValue
	}
	return labels
}

// applyClusterRefLabel ensures a map of labels contains a reference to a parent nifi cluster
func ApplyClusterReferenceLabel(cluster clientconfig.ClusterConnect, labels map[string]string) map[string]string {
	labelValue := cluster.ClusterLabelString()
	if labels == nil {
		labels = make(map[string]string, 0)
	}
	if label, ok := labels[ClusterRefLabel]; ok {
		if label != labelValue {
			labels[ClusterRefLabel] = labelValue
		}
	} else {
		labels[ClusterRefLabel] = labelValue
	}
	return labels
}

// getClusterRefNamespace returns the expected namespace for a Nifi cluster
// referenced by a user/dataflow CR. It takes the namespace of the CR as the first
// argument and the reference itself as the second.
func GetClusterRefNamespace(ns string, ref v1alpha1.ClusterReference) string {
	clusterNamespace := ref.Namespace
	if clusterNamespace == "" {
		return ns
	}
	return clusterNamespace
}

// GetRegistryClientRefNamespace returns the expected namespace for a Nifi registry client
// referenced by a dataflow CR. It takes the namespace of the CR as the first
// argument and the reference itself as the second.
func GetRegistryClientRefNamespace(ns string, ref v1alpha1.RegistryClientReference) string {
	registryClientNamespace := ref.Namespace
	if registryClientNamespace == "" {
		return ns
	}
	return registryClientNamespace
}

// GetParameterContextRefNamespace returns the expected namespace for a Nifi parameter context
// referenced by a dataflow CR. It takes the namespace of the CR as the first
// argument and the reference itself as the second.
func GetParameterContextRefNamespace(ns string, ref v1alpha1.ParameterContextReference) string {
	parameterContextNamespace := ref.Namespace
	if parameterContextNamespace == "" {
		return ns
	}
	return parameterContextNamespace
}

// GetSecretRefNamespace returns the expected namespace for a Nifi secret
// referenced by a parameter context CR. It takes the namespace of the CR as the first
// argument and the reference itself as the second.
func GetSecretRefNamespace(ns string, ref v1alpha1.SecretReference) string {
	secretNamespace := ref.Namespace
	if secretNamespace == "" {
		return ns
	}
	return secretNamespace
}

// GetUserRefNamespace returns the expected namespace for a Nifi user
// referenced by a parameter context CR. It takes the namespace of the CR as the first
// argument and the reference itself as the second.
func GetUserRefNamespace(ns string, ref v1alpha1.UserReference) string {
	userNamespace := ref.Namespace
	if userNamespace == "" {
		return ns
	}
	return userNamespace
}
