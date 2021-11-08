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

package main

import (
	"flag"
	"fmt"
	"github.com/Orange-OpenSource/nifikop/pkg/common"
	"github.com/Orange-OpenSource/nifikop/version"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func printVersion() {
	setupLog.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	setupLog.Info(fmt.Sprintf("Go Version: %s", runtime.APIVersionInternal))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var certManagerEnabled bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&certManagerEnabled, "cert-manager-enabled", false, "Enable cert-manager integration")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	watchNamespace, err := getWatchNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get WatchNamespace, "+
			"the manager will watch and manage resources in all Namespaces")
	}

	options := ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "f1c5ece8.example.com",
	}

	// Add support for MultiNamespace set in WATCH_NAMESPACE (e.g ns1,ns2)
	var namespaceList []string
	if watchNamespace != "" {
		setupLog.Info("manager set up with multiple namespaces", "namespaces", watchNamespace)
		namespaceList = strings.Split(watchNamespace, ",")

		for i := range namespaceList {
			namespaceList[i] = strings.TrimSpace(namespaceList[i])
		}
		// configure cluster-scoped with MultiNamespacedCacheBuilder
		options.NewCache = cache.MultiNamespacedCacheBuilder(namespaceList)
	}

	// NewFileReady returns a Ready that uses the presence of a file on disk to
	// communicate whether the operator is ready. The operator's Pod definition
	// "stat /tmp/operator-sdk-ready".
	setupLog.Info("Writing ready file.")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := certv1.AddToScheme(mgr.GetScheme()); err != nil {
		setupLog.Error(err, "")
		os.Exit(1)
	}

	multipliers := *common.NewRequeueConfig()
	if err = (&controllers.NifiClusterReconciler{
		Client:           mgr.GetClient(),
		DirectClient:     mgr.GetAPIReader(),
		Log:              ctrl.Log.WithName("controllers").WithName("NifiCluster"),
		Scheme:           mgr.GetScheme(),
		Namespaces:       namespaceList,
		Recorder:         mgr.GetEventRecorderFor("nifi-cluster"),
		RequeueIntervals: multipliers.ClusterTaskRequeueIntervals,
		RequeueOffset:    multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NifiCluster")
		os.Exit(1)
	}

	if err = (&controllers.NifiClusterTaskReconciler{
		Client:           mgr.GetClient(),
		Log:              ctrl.Log.WithName("controllers").WithName("NifiClusterTask"),
		Scheme:           mgr.GetScheme(),
		Recorder:         mgr.GetEventRecorderFor("nifi-cluster-task"),
		RequeueIntervals: multipliers.ClusterTaskRequeueIntervals,
		RequeueOffset:    multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NifiClusterTask")
		os.Exit(1)
	}

	if err = (&controllers.NifiUserReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("NifiUser"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("nifi-user"),
		RequeueInterval: multipliers.UserRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr, certManagerEnabled); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NifiUser")
		os.Exit(1)
	}

	if err = (&controllers.NifiUserGroupReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("NifiUserGroup"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("nifi-user-group"),
		RequeueInterval: multipliers.UserGroupRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NifiUserGroup")
		os.Exit(1)
	}

	if err = (&controllers.NifiDataflowReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("NifiDataflow"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("nifi-dataflow"),
		RequeueInterval: multipliers.DataFlowRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NifiDataflow")
		os.Exit(1)
	}

	if err = (&controllers.NifiParameterContextReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("NifiParameterContext"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("nifi-parameter-context"),
		RequeueInterval: multipliers.ParameterContextRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NifiParameterContext")
		os.Exit(1)
	}

	if err = (&controllers.NifiRegistryClientReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("NifiRegistryClient"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("nifi-registry-client"),
		RequeueInterval: multipliers.RegistryClientRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NifiRegistryClient")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}
