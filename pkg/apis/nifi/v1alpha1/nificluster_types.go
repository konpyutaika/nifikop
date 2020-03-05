package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	ClusterListenerType = "cluster"
	HttpListenerType 	= "http"
	HttpsListenerType 	= "https"
	S2sListenerType 	= "s2s"
	MetricsPort 		= 9020
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiClusterSpec defines the desired state of NifiCluster
type NifiClusterSpec struct {
	// headlessServiceEnabled specifies if the cluster should use headlessService for Nifi or individual services
	// using service per nodes may come an handy case of service mesh.
	HeadlessServiceEnabled	bool					`json:"headlessServiceEnabled"`
	// listenerConfig specifies nifi's listener specifig configs
	ListenersConfig			ListenersConfig			`json:"listenersConfig"`
	// zKAddresse specifies the ZooKeeper connection string
	// in the form hostname:port where host and port are those of a Zookeeper server.
	// TODO: rework for nice zookeeper connect string =
	ZKAddresse				string					`json:"zkAddresse"`
	// zKPath specifies the Zookeeper chroot path as part
	// of its Zookeeper connection string which puts its data under same path in the global ZooKeeper namespace.
	ZKPath 					string 					`json:"zkPath,omitempty"`
	// clusterImage can specify the whole nificluster image in one place
	ClusterImage			string					`json:"clusterImage,omitempty"`
	// readOnlyConfig specifies the read-only type Nifi config cluster wide, all theses
	// will be merged with node specified readOnly configurations, so it can be overwritten per node.
	ReadOnlyConfig			ReadOnlyConfig			`json:"readOnlyConfig,omitempty"`
	// nodeConfigGroups specifies multiple node configs with unique name
	NodeConfigGroups   		map[string]NodeConfig	`json:"nodeConfigGroups,omitempty"`
	// all node requires an image, unique id, and storageConfigs settings
	Nodes 					[]Node 					`json:"nodes"`
	// oneNifiNodePerNode if set to true every nifi node is started on a new node, if there is not enough node to do that
	// it will stay in pending state. If set to false the operator also tries to schedule the nifi node to a unique node
	// but if the node number is insufficient the nifi node will be scheduled to a node where a nifi node is already running.
	OneNifiNodePerNode 		bool 					`json:"oneNifiNodePerNode"`
	// propage
	PropagateLabels 		bool 					`json:"propagateLabels,omitempty"`
	// LdapConfiguration specifies the configuration if you want to use LDAP
	LdapConfiguration		LdapConfiguration		`json:"ldapConfiguration,omitempty"`
	// NifiClusterTaskSpec specifies the configuration of the nifi cluster Tasks
	NifiClusterTaskSpec 	NifiClusterTaskSpec		`json:"nifiClusterTaskSpec,omitempty"`
}

// rollingUpgradeConfig specifies the rolling upgrade config for the cluster
//RollingUpgradeConfig 	RollingUpgradeConfig 	`json:"rollingUpgradeConfig"`

// NifiClusterStatus defines the observed state of NifiCluster
type NifiClusterStatus struct {
	// Store the state of each nifi node
	NodesState 		map[string]NodeState	`json:"nodesState,omitempty"`
	// ClusterState holds info about the cluster state
	State 			ClusterState 			`json:"state"`
	// RollingUpgradeStatus defines status of rolling upgrade
	RollingUpgrade 	RollingUpgradeStatus 	`json:"rollingUpgradeStatus,omitempty"`
}

// RollingUpgradeStatus defines status of rolling upgrade
type RollingUpgradeStatus struct {
	//
	LastSuccess	string	`json:"lastSuccess"`
	//
	ErrorCount	int		`json:"errorCount"`
}

// RollingUpgradeConfig defines the desired config of the RollingUpgrade
/*type RollingUpgradeConfig struct {
	// failureThreshold states that how many errors can the cluster tolerate during rolling upgrade
	FailureThreshold	int	`json:"failureThreshold"`
}*/

// Node defines the nifi node basic configuration
type Node struct {
	// Unique Node id
	Id				int32 			`json:"id"`
	// nodeConfigGroup can be used to ease the node configuration, if set no only the id is required
	NodeConfigGroup	string			`json:"nodeConfigGroup,omitempty"`
	// readOnlyConfig can be used to pass Nifi node config https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html
	// which has type read-only these config changes will trigger rolling upgrade
	ReadOnlyConfig	*ReadOnlyConfig	`json:"readOnlyConfig,omitempty"`
	// node configuration
	NodeConfig 		*NodeConfig 	`json:"nodeConfig,omitempty"`
}

type ReadOnlyConfig struct {
	// NifiProperties configuration that will be applied to the node.
	NifiProperties 		NifiProperties		`json:"nifiProperties,omitempty"`
	// ZookeeperProperties configuration that will be applied to the node.
	ZookeeperProperties	ZookeeperProperties	`json:"zookeeperProperties,omitempty"`
	// BootstrapProperties configuration that will be applied to the node.
	BootstrapProperties	BootstrapProperties	`json:"bootstrapProperties,omitempty"`
}

// NifiProperties configuration that will be applied to the node.
type NifiProperties struct {
	// Additionnals nifi.properties configuration that will override the one produced based
	// on template and configurations.
	OverrideConfigs 	string	`json:"overrideConfigs,omitempty"`
	// Site to Site properties Secure mode : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#site_to_site_properties
	SiteToSiteSecure	bool	`json:"siteToSiteSecure,omitempty"`
	// Cluster nodes secure mode : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#cluster_common_properties
	ClusterSecure		bool	`json:"clusterSecure,omitempty"`
	// A comma separated list of allowed HTTP Host header values to consider when NiFi
	// is running securely and will be receiving requests to a different host[:port] than it is bound to.
	// https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#web-properties
	WebProxyHost		string	`json:"webProxyHost,omitempty"`
	// Nifi security client auth
	NeedClientAuth		bool	`json:"needClientAuth,omitempty"`
	// Indicates which of the configured authorizers in the authorizers.xml file to use
	// https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#authorizer-configuration
	Authorizer			string	`json:"authorizer,omitempty"`
}

// ZookeeperProperties configuration that will be applied to the node.
type ZookeeperProperties struct {
	// Additionnals zookeeper.properties configuration that will override the one produced based
	// on template and configurations.
	OverrideConfigs string `json:"overrideConfigs,omitempty"`
}

// BootstrapProperties configuration that will be applied to the node.
type BootstrapProperties struct {
	// JVM memory settings
	NifiJvmMemory	string	`json:"nifiJvmMemory,omitempty"`
	// Additionnals bootstrap.properties configuration that will override the one produced based
	// on template and configurations.
	OverrideConfigs string `json:"overrideConfigs,omitempty"`
}

// NodeConfig defines the node configuration
type NodeConfig struct {
	// provenanceStorage allow to specify the maximum amount of data provenance information to store at a time
	// https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#write-ahead-provenance-repository-properties
	ProvenanceStorage		string						  `json:"provenanceStorage,omitempty"`
	//RunAsUser define the id of the user to run in the Nifi image
	// +kubebuilder:validation:Minimum=1
	RunAsUser 				*int64 						  `json:"runAsUser,omitempty"`
	// Set this to true if the instance is a node in a cluster.
	// https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#basic-cluster-setup
	IsNode					*bool						  `json:"isNode,omitempty"`
	//  Docker image used by the operator to create the node associated
	//  https://hub.docker.com/r/apache/nifi/
	Image              		string                        `json:"image,omitempty"`
	// nodeAffinity can be specified, operator populates this value if new pvc added later to node
	NodeAffinity       		*corev1.NodeAffinity          `json:"nodeAffinity,omitempty"`
	// storageConfigs specifies the node related configs
	StorageConfigs     		[]StorageConfig               `json:"storageConfigs,omitempty"`
	// serviceAccountName specifies the serviceAccount used for this specific node
	ServiceAccountName 		string                        `json:"serviceAccountName,omitempty"`
	// resourceRequirements works exactly like Container resources, the user can specify the limit and the requests
	// through this property
	// https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	ResourcesRequirements	*corev1.ResourceRequirements  `json:"resourcesRequirements,omitempty"`
	// imagePullSecrets specifies the secret to use when using private registry
	// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#localobjectreference-v1-core
	ImagePullSecrets   		[]corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// nodeSelector can be specified, which set the pod to fit on a node
	// https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector
	NodeSelector       		map[string]string             `json:"nodeSelector,omitempty"`
	// tolerations can be specified, which set the pod's tolerations
	// https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/#concepts
	Tolerations       		[]corev1.Toleration           `json:"tolerations,omitempty"`
	// Additionnal annotation to attach to the pod associated
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/#syntax-and-character-set
	NodeAnnotations  		map[string]string             `json:"nifiAnnotations,omitempty"`
}

// StorageConfig defines the node storage configuration
type StorageConfig struct {
	// Name of the storage config, used to name PV to reuse into sidecars for example.
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	Name 				string                            	`json:"name"`
	// Path where the volume will be mount into the main nifi container inside the pod.
	MountPath 			string                            	`json:"mountPath"`
	// Kubernetes PVC spec
	PVCSpec				*corev1.PersistentVolumeClaimSpec	`json:"pvcSpec"`
}

//ListenersConfig defines the Nifi listener types
type ListenersConfig struct {
	// externalListeners specifies settings required to access nifi externally
	// TODO: enable externalListener configuration
	//ExternalListeners []ExternalListenerConfig `json:"externalListeners,omitempty"`
	// internalListeners specifies settings required to access nifi internally
	InternalListeners	[]InternalListenerConfig	`json:"internalListeners"`

	// sslSecrets contains information about ssl related kubernetes secrets if one of the
	// listener setting type set to ssl these fields must be populated to
	// TODO: enable this feature.
	//SSLSecrets        *SSLSecrets              `json:"sslSecrets,omitempty"`
}

// SSLSecrets defines the Nifi SSL secrets
// TODO: implement SSL management
type SSLSecrets struct {
	// tlsSecretName should contain all ssl certs required by nifi including: caCert, caKey, clientCert, clientKey
	// serverCert, serverKey, peerCert, peerKey
	TLSSecretName   string		`json:"tlsSecretName"`
	// jksPasswordName should contain a password field which contains the jks password
	JKSPasswordName string 		`json:"jksPasswordName"`
	// create tells the installed cert manager to create the required certs keys
	Create          bool   		`json:"create,omitempty"`

	// +kubebuilder:validation:Enum={"cert-manager"}
	PKIBackend		PKIBackend	`json:"pkiBackend,omitempty"`
}

// ExternalListenerConfig defines the external listener config for Nifi
// TODO: enable configuration of ingress or something like this.
type ExternalListenerConfig struct {
	// TODO: remove type field # specific to Kafka
	//
	Type                 string	`json:"type"`
	//
	Name                 string `json:"name"`
	//
	ExternalStartingPort int32  `json:"externalStartingPort"`
	//
	ContainerPort        int32  `json:"containerPort"`
	//
	HostnameOverride     string `json:"hostnameOverride,omitempty"`
}

// InternalListenerConfig defines the internal listener config for Nifi
type InternalListenerConfig struct {
	// +kubebuilder:validation:Enum={"cluster", "http", "https", "s2s"}
	// (Optional field) Type allow to specify if we are in a specific nifi listener
	// it's allowing to define some required information such as Cluster Port,
	// Http Port, Https Port or S2S port
	Type  			string	`json:"type,omitempty"`
	// An identifier for the port which will be configured.
	Name			string `json:"name"`
	// The container port.
	ContainerPort	int32  `json:"containerPort"`
}

// LdapConfiguration specifies the configuration if you want to use LDAP
type LdapConfiguration struct {
	// If set to true, we will enable ldap usage into nifi.properties configuration.
	Enabled			bool 	`json:"enabled,omitempty"`
	// Space-separated list of URLs of the LDAP servers (i.e. ldap://<hostname>:<port>).
	Url				string	`json:"url,omitempty"`
	// Base DN for searching for users (i.e. CN=Users,DC=example,DC=com).
	SearchBase		string 	`json:"searchBase,omitempty"`
	// Filter for searching for users against the 'User Search Base'.
	// (i.e. sAMAccountName={0}). The user specified name is inserted into '{0}'.
	SearchFilter	string 	`json:"searchFilter,omitempty"`
}

// NifiClusterTaskSpec specifies the configuration of the nifi cluster Tasks
type NifiClusterTaskSpec struct {
	// RetryDurationMinutes describes the amount of time the Operator waits for the task
	RetryDurationMinutes int `json:"retryDurationMinutes"`
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

//
func (nConfig *NodeConfig) GetIsNode() bool {
	if nConfig.IsNode != nil {
		return *nConfig.IsNode
	}
	return true
}

func (nConfig *NodeConfig) GetProvenanceStorage() string {
	if nConfig.ProvenanceStorage != "" {
		return nConfig.ProvenanceStorage
	}
	return "8 GB"
}

// GetNifiJvmMemory returns the default "2g" NifiJvmMemory if not specified otherwise
func (bProperties *BootstrapProperties) GetNifiJvmMemory() string {
	if bProperties.NifiJvmMemory != "" {
		return bProperties.NifiJvmMemory
	}
	return "512m"
}

//
func (nProperties NifiProperties) GetAuthorizer() string {
	if nProperties.Authorizer != "" {
		return nProperties.Authorizer
	}
	return "managed-authorizer"
}







