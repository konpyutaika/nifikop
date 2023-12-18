package resources

import (
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

// Reconciler holds CR for Nifi.
type Reconciler struct {
	client.Client
	DirectClient             client.Reader
	NifiCluster              *v1.NifiCluster
	NifiClusterCurrentStatus v1.NifiClusterStatus
}

// ComponentReconciler describes the Reconcile method.
type ComponentReconciler interface {
	Reconcile(log zap.Logger) error
}

// ResourceWithLogs function with log parameter.
type ResourceWithLogs func(log zap.Logger) runtime.Object

// ResourceWithNodeConfigAndVolume function with nodeConfig, persistentVolumeClaims and log parameters.
type ResourceWithNodeConfigAndVolume func(id int32, nodeConfig *v1.NodeConfig, pvcs []corev1.PersistentVolumeClaim, log zap.Logger) runtime.Object

// ResourceWithNodeConfigAndString function with nodeConfig, string and log parameters.
type ResourceWithNodeConfigAndString func(id int32, nodeConfig *v1.NodeConfig, t string, su []string, log zap.Logger) runtime.Object

// ResourceWithNodeIdAndStorage function with nodeConfig, storageConfig and log parameters.
type ResourceWithNodeIdAndStorage func(id int32, storage v1.StorageConfig, log zap.Logger) runtime.Object

// ResourceWithNodeIdAndLog function with nodeConfig and log parameters.
type ResourceWithNodeIdAndLog func(id int32, log zap.Logger) runtime.Object
