package v1alpha1

// NodeGroupAutoscalerState holds info autoscaler state
type NodeGroupAutoscalerState string

// ClusterReplicas holds info about the current number of replicas in the cluster
type ClusterReplicas int32

// ClusterReplicaSelector holds info about the pod selector for cluster replicas
type ClusterReplicaSelector string

// ClusterScalingStrategy holds info about how a cluster should be scaled
type ClusterScalingStrategy string

const (
	// AutoscalerStateOutOfSync describes the status of a NifiNodeGroupAutoscaler as out of sync
	AutoscalerStateOutOfSync NodeGroupAutoscalerState = "OutOfSync"
	// AutoscalerStateInSync describes the status of a NifiNodeGroupAutoscaler as in sync
	AutoscalerStateInSync NodeGroupAutoscalerState = "InSync"

	// upscale strategy representing 'Scale > Disconnect the nodes > Offload data > Reconnect the node' strategy
	GracefulClusterUpscaleStrategy ClusterScalingStrategy = "graceful"
	// simply add a node to the cluster and nothing else
	SimpleClusterUpscaleStrategy ClusterScalingStrategy = "simple"
	// downscale strategy to remove the last node added
	LIFOClusterDownscaleStrategy ClusterScalingStrategy = "lifo"
	// downscale strategy avoiding primary/coordinator nodes
	NonPrimaryClusterDownscaleStrategy ClusterScalingStrategy = "nonprimary"
	// downscale strategy targeting nodes which are least busy in terms of # flowfiles in queues
	LeastBusyClusterDownscaleStrategy ClusterScalingStrategy = "leastbusy"
)
