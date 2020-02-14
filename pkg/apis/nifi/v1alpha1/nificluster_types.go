package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiClusterSpec defines the desired state of NifiCluster
type NifiClusterSpec struct {
	// headlessServiceEnabled specifies if the cluster should use headlessService for Nifi or individual services
	// using service per nodes may come an handy case of service mesh.
	HeadlessServiceEnabled	bool	`json:"headlessServiceEnabled"`

	// listenerConfig specifies nifi's listener specifig configs
	ListenersConfig	ListenersConfig	`json:"listenersConfig"`

	// zKAddresse specifies the ZooKeeper connection string
	// in the form hostname:port where host and port are those of a Zookeeper server.
	ZKAddresse	string	`json:"zkAddresse"`

	// zKPath specifies the Zookeeper chroot path as part
	// of its Zookeeper connection string which puts its data under same path in the global ZooKeeper namespace.
	ZKPath string `json:"zkPath,omitempty"`

	// rackAwarness add support for Nifi related metadatas should be placed
	// RackAwareness	*RackAwareness	`json:"rackAwareness,omitempty"`

	// clusterImage can specify the whole nificluster image in one place
	ClusterImage	string	`json:"clusterImage,omitempty"`

	// readOnlyConfig specifies the read-only type Nifi config cluster wide, all theses
	// will be merged withj node specified readOnly configurations, so it can be overwritten per node.
	ReadOnlyConfig	string	`json:"readOnlyConfig,omitempty"`

	// clusterWideConfig specifies the cluster-wide nifi config, all these can be overriden per node.
	ClusterWideConfig	string	`json:"clusterWideConfig,omitempty"`

	// nodeConfigGroups specifies multiple node configs with unique name
	NodeConfigGroups   map[string]NodeConfig `json:"nodeConfigGroups,omitempty"`

	// all node requires an image, unique id, and storageConfigs settings
	Nodes []Node `json:"nodes"`

	// rollingUpgradeConfig specifies the rolling upgrade config for the cluster
	RollingUpgradeConfig RollingUpgradeConfig `json:"rollingUpgradeConfig"`

	// oneNifiNodePerNode if set to true every nifi node is started on a new node, if there is not enough node to do that
	// it will stay in pending state. If set to false the operator also tries to schedule the nifi node to a unique node
	// but if the node number is insufficient the nifi node will be scheduled to a node where a nifi node is already running.
	OneNifiNodePerNode bool `json:"oneNifiNodePerNode"`

	//
	PropagateLabels bool `json:"propagateLabels,omitempty"`

	// NifiClusterTaskSpec specifies the configuration of the nifi cluster Tasks
	NifiClusterTaskSpec NifiClusterTaskSpec	`json:"nifiClusterTaskSpec,omitempty"`
}

// NifiClusterStatus defines the observed state of NifiCluster
type NifiClusterStatus struct {
	//
	NodesState map[string]NodeState `json:"nodesState,omitempty"`

	//
	State ClusterState `json:"state"`

	//
	RollingUpgrade RollingUpgradeStatus `json:"rollingUpgradeStatus,omitempty"`
}

// RollingUpgradeStatus defines status of rolling upgrade
type RollingUpgradeStatus struct {
	//
	LastSuccess string `json:"lastSuccess"`
	//
	ErrorCount  int    `json:"errorCount"`
}

// RollingUpgradeConfig defines the desired config of the RollingUpgrade
type RollingUpgradeConfig struct {
	// failureThreshold states that how many errors can the cluster tolerate during rolling upgrade
	FailureThreshold int `json:"failureThreshold"`
}

// Node defines the nifi node basic configuration
type Node struct {
	// Unique Node id which is used as nifi config nifi.id
	Id  int32 `json:"id"`
	// nodeConfigGroup can be used to ease the node configuration, if set no only the id is required
	NodeConfigGroup string `json:"nodeConfigGroup,omitempty"`
	// readOnlyConfig can be used to pass Nifi node config https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html
	// which has type read-only these config changes will trigger rolling upgrade
	ReadOnlyConfig string `json:"readOnlyConfig,omitempty"`
	// node configuration
	NodeConfig *NodeConfig `json:"nodeConfig,omitempty"`
}

// NodeConfig defines the node configuration
type NodeConfig struct {
	//RunAsUser define the id of the user to run in the Cassandra image
	// +kubebuilder:validation:Minimum=1
	RunAsUser *int64 `json:"runAsUser,omitempty"`

	// Docker image used by the operator to create the node associated
	Image              		string                        `json:"image,omitempty"`
	// nodeAffinity can be specified, operator populates this value if new pvc added later to node
	NodeAffinity       		*corev1.NodeAffinity          `json:"nodeAffinity,omitempty"`
	// config parameter can be used to pass Nifi node config https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html
	// TODO: to remove
	Config             		string                        `json:"config,omitempty"`
	// storageConfigs specifies the node log related configs
	StorageConfigs     		[]StorageConfig               `json:"storageConfigs,omitempty"`
	// serviceAccountName specifies the serviceAccount used for this specific node
	ServiceAccountName 		string                        `json:"serviceAccountName,omitempty"`
	// resourceRequirements works exactly like Container resources, the user can specify the limit and the requests
	// through this property
	ResourcesRequirements	*corev1.ResourceRequirements  `json:"resourceRequirements,omitempty"`
	// imagePullSecrets specifies the secret to use when using private registry
	ImagePullSecrets   		[]corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// nodeSelector can be specified, which set the pod to fit on a node
	NodeSelector       		map[string]string             `json:"nodeSelector,omitempty"`
	// tolerations can be specified, which set the pod's tolerations
	Tolerations       		[]corev1.Toleration           `json:"tolerations,omitempty"`
/*
	NifiHeapOpts      string                        `json:"nifiHeapOpts,omitempty"`
	NifiJVMPerfOÒpts   string                        `json:"nifiJvmPerfOpts,omitempty"`
 */
	//
	NodeAnnotations  		map[string]string             `json:"nifiAnnotations,omitempty"`
}

// TODO: maybe useless.
// RackAwareness defines the required fields to enable nifi's rack aware feature
type RackAwareness struct {
	Labels []string `json:"labels"`
}

// StorageConfig defines the node storage configuration
type StorageConfig struct {
	//
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	Name 			string                            	`json:"name"`
	//
	MountPath 			string                            	`json:"mountPath"`
	//
	PVCSpec   			*corev1.PersistentVolumeClaimSpec	`json:"pvcSpec"`
	//
	IsProvenanceStorage	bool								`json:"isProvenanceStorage,omitempty"`
}

//ListenersConfig defines the Nifi listener types
type ListenersConfig struct {
	// externalListeners specifies settings required to access nifi externally
	ExternalListeners []ExternalListenerConfig `json:"externalListeners,omitempty"`
	// internalListeners specifies settings required to access nifi internally
	InternalListeners []InternalListenerConfig `json:"internalListeners"`

	// sslSecrets contains information about ssl related kubernetes secrets if one of the
	// listener setting type set to ssl these fields must be populated to
	SSLSecrets        *SSLSecrets              `json:"sslSecrets,omitempty"`
}

// SSLSecrets defines the Nifi SSL secrets
type SSLSecrets struct {
	// tlsSecretName should contain all ssl certs required by nifi including: caCert, caKey, clientCert, clientKey
	// serverCert, serverKey, peerCert, peerKey
	TLSSecretName   string `json:"tlsSecretName"`
	// jksPasswordName should contain a password field which contains the jks password
	JKSPasswordName string `json:"jksPasswordName"`
	// create tells the installed cert manager to create the required certs keys
	Create          bool   `json:"create,omitempty"`

	// +kubebuilder:validation:Enum={"cert-manager"}
	PKIBackend PKIBackend `json:"pkiBackend,omitempty"`
}

// ExternalListenerConfig defines the external listener config for Nifi
type ExternalListenerConfig struct {
	// TODO: remove type field # specific to Kafka
	Type                 string `json:"type"`
	Name                 string `json:"name"`
	ExternalStartingPort int32  `json:"externalStartingPort"`
	ContainerPort        int32  `json:"containerPort"`
	HostnameOverride     string `json:"hostnameOverride,omitempty"`
}

// InternalListenerConfig defines the internal listener config for Nifi
// TODO: improve logic about port usage.
type InternalListenerConfig struct {
	// +kubebuilder:validation:Enum={"cluster", "http", "https", "s2s" }
	Type                            string `json:"type,omitempty"`
	Name                            string `json:"name"`
	ContainerPort                   int32  `json:"containerPort"`
}

// NifiClusterTaskSpec specifies the configuration of the nifi cluster Tasks
type NifiClusterTaskSpec struct {
	// RetryDurationMinutes describes the amount of time the Operator waits for the task
	RetryDurationMinutes int `json:"RetryDurationMinutes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NifiCluster is the Schema for the nificlusters API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=nificlusters,scope=Namespaced
type NifiCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiClusterSpec   `json:"spec,omitempty"`
	Status NifiClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NifiClusterList contains a list of NifiCluster
type NifiClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiCluster{}, &NifiClusterList{})
}

// GetZkPath returns the default "/" ZkPath if not specified otherwise
func (nSpec *NifiClusterSpec) GetZkPath() string {
	const prefix = "/"
	if nSpec.ZKPath == "" {
		return prefix
	} else if !strings.HasPrefix(nSpec.ZKPath, prefix) {
		return prefix + nSpec.ZKPath
	} else {
		return nSpec.ZKPath
	}
}

func (nTaskSpec *NifiClusterTaskSpec) GetDurationMinutes() float64 {
	if nTaskSpec.RetryDurationMinutes == 0 {
		return 5
	}
	return float64(nTaskSpec.RetryDurationMinutes)
}


// GetServiceAccount returns the Kubernetes Service Account to use for Nifi Cluster
func (nConfig *NodeConfig) GetServiceAccount() string {
	if nConfig.ServiceAccountName != "" {
		return nConfig.ServiceAccountName
	}
	return "default"
}

//GetTolerations returns the tolerations for the given node
func (nConfig *NodeConfig) GetTolerations() []corev1.Toleration {
	return nConfig.Tolerations
}

// GetNodeSelector returns the node selector for the given node
func (nConfig *NodeConfig) GetNodeSelector() map[string]string {
	return nConfig.NodeSelector
}

//GetImagePullSecrets returns the list of Secrets needed to pull Containers images from private repositories
func (nConfig *NodeConfig) GetImagePullSecrets() []corev1.LocalObjectReference {
	return nConfig.ImagePullSecrets
}

//
func (nConfig *NodeConfig) GetNodeAnnotations() map[string]string {
	return nConfig.NodeAnnotations
}

// GetResources returns the nifi node specific Kubernetes resource
func (nConfig *NodeConfig) GetResources() *corev1.ResourceRequirements {
	if nConfig.ResourcesRequirements != nil {
		return nConfig.ResourcesRequirements
	}
	return &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu":    resource.MustParse("1000m"),
			"memory": resource.MustParse("1Gi"),
		},
		Requests: corev1.ResourceList{
			"cpu":    resource.MustParse("1000m"),
			"memory": resource.MustParse("1Gi"),
		},
	}
}

//
func (nConfig *NodeConfig) GetRunAsUser() *int64 {
	var defaultUserID int64 = 1000
	if nConfig.RunAsUser != nil {
		return nConfig.RunAsUser
	}

	return func(i int64) *int64 { return &i }(defaultUserID)
}







