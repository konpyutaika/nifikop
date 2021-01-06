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

package resources

import (
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler holds CR for Nifi
type Reconciler struct {
	client.Client
	DirectClient client.Reader
	NifiCluster  *v1alpha1.NifiCluster
}

// ComponentReconciler describes the Reconcile method
type ComponentReconciler interface {
	Reconcile(log logr.Logger) error
}

// ResourceWithLogs function with log parameter
type ResourceWithLogs func(log logr.Logger) runtime.Object

// ResourceWithNodeConfigAndVolume function with nodeConfig, persistentVolumeClaims and log parameters
type ResourceWithNodeConfigAndVolume func(id int32, nodeConfig *v1alpha1.NodeConfig, pvcs []corev1.PersistentVolumeClaim, log logr.Logger) runtime.Object

// ResourceWithNodeConfigAndString function with nodeConfig, string and log parameters
type ResourceWithNodeConfigAndString func(id int32, nodeConfig *v1alpha1.NodeConfig, t string, su []string, log logr.Logger) runtime.Object

// ResourceWithNodeIdAndStorage function with nodeConfig, storageConfig and log parameters
type ResourceWithNodeIdAndStorage func(id int32, storage v1alpha1.StorageConfig, log logr.Logger) runtime.Object

// ResourceWithNodeIdAndLog function with nodeConfig and log parameters
type ResourceWithNodeIdAndLog func(id int32, log logr.Logger) runtime.Object
