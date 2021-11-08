package common

import (
	"github.com/Orange-OpenSource/nifikop/pkg/nificlient"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	"github.com/Orange-OpenSource/nifikop/pkg/util/clientconfig"
	"github.com/go-logr/logr"
)

//// NewFromCluster is a convenient wrapper around New() and ClusterConfig()
//func NewFromCluster(k8sclient client.Client, cluster *v1alpha1.NifiCluster) (nificlient.NifiClient, error) {
//	var client nificlient.NifiClient
//	var err error
//	var opts *clientconfig.NifiConfig
//
//	if opts, err = tls.New(k8sclient,
//		v1alpha1.ClusterReference{Name: cluster.Name, Namespace: cluster.Namespace}).BuildConfig(); err != nil {
//		return nil, err
//	}
//	client = nificlient.New(opts)
//	err = client.Build()
//	if err != nil {
//		return nil, err
//	}
//
//	return client, nil
//}
//
//// NewNifiFromCluster points to the function for retrieving nifi clients,
//// use as var so it can be overwritten from unit tests
//var NewNifiFromCluster = NewFromCluster
//
//// newNodeConnection is a convenience wrapper for creating a node connection
//// and creating a safer close function
//func NewNodeConnection(log logr.Logger, client client.Client, cluster *v1alpha1.NifiCluster) (node nificlient.NifiClient, err error) {
//
//	// Get a nifi connection
//	log.Info(fmt.Sprintf("Retrieving Nifi client for %s/%s", cluster.Namespace, cluster.Name))
//	node, err = NewNifiFromCluster(client, cluster)
//	if err != nil {
//		return
//	}
//	return
//}

// NewNifiFromCluster points to the function for retrieving nifi clients,
// use as var so it can be overwritten from unit tests
var NewNifiFromConfig = nificlient.NewFromConfig

// newNodeConnection is a convenience wrapper for creating a node connection
// and creating a safer close function
func NewClusterConnection(log logr.Logger, config *clientconfig.NifiConfig) (node nificlient.NifiClient, err error) {

	// Get a nifi connection
	node, err = NewNifiFromConfig(config)
	if err != nil {
		return
	}
	return
}

type RequeueConfig struct {
	UserRequeueInterval             int
	RegistryClientRequeueInterval   int
	ParameterContextRequeueInterval int
	UserGroupRequeueInterval        int
	DataFlowRequeueInterval         int
	ClusterTaskRequeueIntervals     map[string]int
	RequeueOffset                   int
}

func NewRequeueConfig() *RequeueConfig {
	return &RequeueConfig{
		ClusterTaskRequeueIntervals: map[string]int{
			"CLUSTER_TASK_RUNNING_REQUEUE_INTERVAL":   util.MustConvertToInt(util.GetEnvWithDefault("CLUSTER_TASK_RUNNING_REQUEUE_INTERVAL", "20"), "CLUSTER_TASK_RUNNING_REQUEUE_INTERVAL"),
			"CLUSTER_TASK_TIMEOUT_REQUEUE_INTERVAL":   util.MustConvertToInt(util.GetEnvWithDefault("CLUSTER_TASK_TIMEOUT_REQUEUE_INTERVAL", "20"), "CLUSTER_TASK_TIMEOUT_REQUEUE_INTERVAL"),
			"CLUSTER_TASK_NOT_READY_REQUEUE_INTERVAL": util.MustConvertToInt(util.GetEnvWithDefault("CLUSTER_TASK_NOT_READY_REQUEUE_INTERVAL", "15"), "CLUSTER_TASK_NODES_UNREACHABLE_REQUEUE_INTERVAL"),
		},
		UserRequeueInterval:             util.MustConvertToInt(util.GetEnvWithDefault("USERS_REQUEUE_INTERVAL", "15"), "USERS_REQUEUE_INTERVAL"),
		RegistryClientRequeueInterval:   util.MustConvertToInt(util.GetEnvWithDefault("REGISTRY_CLIENT_REQUEUE_INTERVAL", "15"), "REGISTRY_CLIENT_REQUEUE_INTERVAL"),
		ParameterContextRequeueInterval: util.MustConvertToInt(util.GetEnvWithDefault("PARAMETER_CONTEXT_REQUEUE_INTERVAL", "15"), "PARAMETER_CONTEXT_REQUEUE_INTERVAL"),
		UserGroupRequeueInterval:        util.MustConvertToInt(util.GetEnvWithDefault("USER_GROUP_REQUEUE_INTERVAL", "15"), "USER_GROUP_REQUEUE_INTERVAL"),
		DataFlowRequeueInterval:         util.MustConvertToInt(util.GetEnvWithDefault("DATAFLOW_REQUEUE_INTERVAL", "15"), "DATAFLOW_REQUEUE_INTERVAL"),
		RequeueOffset:                   util.MustConvertToInt(util.GetEnvWithDefault("REQUEUE_OFFSET", "0"), "REQUEUE_OFFSET"),
	}
}
