package resources

import (
	"github.com/go-logr/logr"
	"github.com/orangeopensource/nifi-operator/pkg/apis/nifi/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler holds CR for Nifi
type Reconciler struct {
	client.Client
	NifiCluster *v1alpha1.NifiCluster
}

// ComponentReconciler describes the Reconcile method
type ComponentReconciler interface  {
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