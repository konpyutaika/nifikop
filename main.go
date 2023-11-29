package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	certv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/controllers"
	"github.com/konpyutaika/nifikop/pkg/common"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(v1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var certManagerEnabled bool
	var webhookEnabled bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":9090", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&certManagerEnabled, "cert-manager-enabled", false, "Enable cert-manager integration")
	flag.BoolVar(&webhookEnabled, "webhook-enabled", true, "Enable CRDs conversion webhook.")

	flag.Parse()

	logger := common.CustomLogger()

	ctrl.SetLogger(zapr.NewLogger(logger))

	watchNamespace, err := getWatchNamespace()
	if err != nil {
		logger.Error("unable to get WATCH_NAMESPACE ENV, will watch and manage resources in all namespaces", zap.Error(err))
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
		logger.Info("WATCH_NAMESPACE ENV provided, will watch and manage resources in defined namespaces",
			zap.String("namespaces", watchNamespace))
		namespaceList = strings.Split(watchNamespace, ",")

		for i := range namespaceList {
			namespaceList[i] = strings.TrimSpace(namespaceList[i])
		}
		// configure cluster-scoped with MultiNamespacedCacheBuilder
		options.NewCache = cache.MultiNamespacedCacheBuilder(namespaceList)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		logger.Error("unable to start manager", zap.Error(err))
		os.Exit(1)
	}

	if err := certv1.AddToScheme(mgr.GetScheme()); err != nil {
		logger.Error("", zap.Error(err))
		os.Exit(1)
	}

	multipliers := *common.NewRequeueConfig()
	if err = (&controllers.NifiClusterReconciler{
		Client:           mgr.GetClient(),
		DirectClient:     mgr.GetAPIReader(),
		Log:              *logger.Named("controllers").Named("NifiCluster"),
		Scheme:           mgr.GetScheme(),
		Namespaces:       namespaceList,
		Recorder:         mgr.GetEventRecorderFor("nifi-cluster"),
		RequeueIntervals: multipliers.ClusterTaskRequeueIntervals,
		RequeueOffset:    multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		logger.Error("unable to create controller", zap.String("controller", "NifiCluster"), zap.Error(err))
		os.Exit(1)
	}

	if err = (&controllers.NifiClusterTaskReconciler{
		Client:           mgr.GetClient(),
		Log:              *logger.Named("controllers").Named("NifiClusterTask"),
		Scheme:           mgr.GetScheme(),
		Recorder:         mgr.GetEventRecorderFor("nifi-cluster-task"),
		RequeueIntervals: multipliers.ClusterTaskRequeueIntervals,
		RequeueOffset:    multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		logger.Error("unable to create controller", zap.String("controller", "NifiClusterTask"), zap.Error(err))
		os.Exit(1)
	}

	if err = (&controllers.NifiUserReconciler{
		Client:          mgr.GetClient(),
		Log:             *logger.Named("controllers").Named("NifiUser"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("nifi-user"),
		RequeueInterval: multipliers.UserRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr, certManagerEnabled); err != nil {
		logger.Error("unable to create controller", zap.String("controller", "NifiUser"), zap.Error(err))
		os.Exit(1)
	}

	if err = (&controllers.NifiUserGroupReconciler{
		Client:          mgr.GetClient(),
		Log:             *logger.Named("controllers").Named("NifiUserGroup"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("nifi-user-group"),
		RequeueInterval: multipliers.UserGroupRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		logger.Error("unable to create controller", zap.String("controller", "NifiUserGroup"), zap.Error(err))
		os.Exit(1)
	}

	if err = (&controllers.NifiDataflowReconciler{
		Client:          mgr.GetClient(),
		Log:             *logger.Named("controllers").Named("NifiDataflow"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("nifi-dataflow"),
		RequeueInterval: multipliers.DataFlowRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		logger.Error("unable to create controller", zap.String("controller", "NifiDataflow"), zap.Error(err))
		os.Exit(1)
	}

	if err = (&controllers.NifiParameterContextReconciler{
		Client:          mgr.GetClient(),
		Log:             *logger.Named("controllers").Named("NifiParameterContext"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("nifi-parameter-context"),
		RequeueInterval: multipliers.ParameterContextRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		logger.Error("unable to create controller", zap.String("controller", "NifiParameterContext"), zap.Error(err))
		os.Exit(1)
	}

	if err = (&controllers.NifiRegistryClientReconciler{
		Client:          mgr.GetClient(),
		Log:             *logger.Named("controllers").Named("NifiRegistryClient"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("nifi-registry-client"),
		RequeueInterval: multipliers.RegistryClientRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		logger.Error("unable to create controller", zap.String("controller", "NifiRegistryClient"), zap.Error(err))
		os.Exit(1)
	}

	if err = (&controllers.NifiNodeGroupAutoscalerReconciler{
		Client:          mgr.GetClient(),
		APIReader:       mgr.GetAPIReader(),
		Scheme:          mgr.GetScheme(),
		Log:             *logger.Named("controllers").Named("NifiNodeGroupAutoscaler"),
		Recorder:        mgr.GetEventRecorderFor("nifi-node-group-autoscaler"),
		RequeueInterval: multipliers.NodeGroupAutoscalerRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		logger.Error("unable to create controller", zap.String("controller", "NifiNodeGroupAutoscaler"), zap.Error(err))
		os.Exit(1)
	}

	if err = (&controllers.NifiConnectionReconciler{
		Client:          mgr.GetClient(),
		Log:             *logger.Named("controllers").Named("NifiConnection"),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("nifi-connection"),
		RequeueInterval: multipliers.ConnectionRequeueInterval,
		RequeueOffset:   multipliers.RequeueOffset,
	}).SetupWithManager(mgr); err != nil {
		logger.Error("unable to create controller", zap.String("controller", "NifiConnection"), zap.Error(err))
		os.Exit(1)
	}

	if webhookEnabled {
		if err = (&v1alpha1.NifiUser{}).SetupWebhookWithManager(mgr); err != nil {
			logger.Error("unable to create webhook", zap.String("webhook", "NifiUser"), zap.Error(err))
			os.Exit(1)
		}
		if err = (&v1alpha1.NifiCluster{}).SetupWebhookWithManager(mgr); err != nil {
			logger.Error("unable to create webhook", zap.String("webhook", "NifiCluster"), zap.Error(err))
			os.Exit(1)
		}
		if err = (&v1alpha1.NifiDataflow{}).SetupWebhookWithManager(mgr); err != nil {
			logger.Error("unable to create webhook", zap.String("webhook", "NifiDataflow"), zap.Error(err))
			os.Exit(1)
		}
		if err = (&v1alpha1.NifiParameterContext{}).SetupWebhookWithManager(mgr); err != nil {
			logger.Error("unable to create webhook", zap.String("webhook", "NifiParameterContext"), zap.Error(err))
			os.Exit(1)
		}
		if err = (&v1alpha1.NifiRegistryClient{}).SetupWebhookWithManager(mgr); err != nil {
			logger.Error("unable to create webhook", zap.String("webhook", "NifiRegistryClient"), zap.Error(err))
			os.Exit(1)
		}
		if err = (&v1alpha1.NifiUserGroup{}).SetupWebhookWithManager(mgr); err != nil {
			logger.Error("unable to create webhook", zap.String("webhook", "NifiUserGroup"), zap.Error(err))
			os.Exit(1)
		}
	}

	// +kubebuilder:scaffold:builder
	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		logger.Error("unable to set up health check", zap.Error(err))
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		logger.Error("unable to set up ready check", zap.Error(err))
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error("unable to start manager", zap.Error(err))
		os.Exit(1)
	}
}

// getWatchNamespace returns the Namespace the operator should be watching for changes.
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
