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

package v1alpha1

import (
	"strings"

	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ClusterListenerType    = "cluster"
	HttpListenerType       = "http"
	HttpsListenerType      = "https"
	S2sListenerType        = "s2s"
	prometheusListenerType = "prometheus"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiClusterSpec defines the desired state of NifiCluster
type NifiClusterSpec struct {
	// clientType defines if the operator will use basic or tls authentication to query the NiFi cluster.
	// +kubebuilder:validation:Enum={"tls","basic"}
	ClientType ClientConfigType `json:"clientType,omitempty"`
	// type defines if the cluster is internal (i.e manager by the operator) or external.
	// +kubebuilder:validation:Enum={"external","internal"}
	Type ClusterType `json:"type,omitempty"`
	// nodeURITemplate used to dynamically compute node uri (used if external type)
	NodeURITemplate string `json:"nodeURITemplate,omitempty"`
	// nifiURI used access through a LB uri (used if external type)
	NifiURI string `json:"nifiURI,omitempty"`
	// rootProcessGroupId contains the uuid of the root process group for this cluster (used if external type)
	RootProcessGroupId string `json:"rootProcessGroupId,omitempty"`
	// secretRef reference the secret containing the informations required to authentiticate to the cluster (used if external type)
	SecretRef SecretReference `json:"secretRef,omitempty"`
	// proxyUrl defines the proxy required to query the NiFi cluster (used if external type)
	ProxyUrl string `json:"proxyUrl,omitempty"`
	// Service defines the policy for services owned by NiFiKop operator.
	Service ServicePolicy `json:"service,omitempty"`
	// Pod defines the policy for  pods owned by NiFiKop operator.
	Pod PodPolicy `json:"pod,omitempty"`
	// zKAddress specifies the ZooKeeper connection string
	// in the form hostname:port where host and port are those of a Zookeeper server.
	// TODO: rework for nice zookeeper connect string =
	ZKAddress string `json:"zkAddress,omitempty"`
	// zKPath specifies the Zookeeper chroot path as part
	// of its Zookeeper connection string which puts its data under same path in the global ZooKeeper namespace.
	ZKPath string `json:"zkPath,omitempty"`
	// initContainerImage can override the default image used into the init container to check if
	// ZoooKeeper server is reachable.
	InitContainerImage string `json:"initContainerImage,omitempty"`
	// initContainers defines additional initContainers configurations
	InitContainers []corev1.Container `json:"initContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`
	// clusterImage can specify the whole NiFi cluster image in one place
	ClusterImage string `json:"clusterImage,omitempty"`
	// oneNifiNodePerNode if set to true every nifi node is started on a new node, if there is not enough node to do that
	// it will stay in pending state. If set to false the operator also tries to schedule the nifi node to a unique node
	// but if the node number is insufficient the nifi node will be scheduled to a node where a nifi node is already running.
	OneNifiNodePerNode bool `json:"oneNifiNodePerNode,omitempty"`
	// propage
	PropagateLabels bool `json:"propagateLabels,omitempty"`
	// managedAdminUsers contains the list of users that will be added to the managed admin group (with all rights)
	ManagedAdminUsers []ManagedUser `json:"managedAdminUsers,omitempty"`
	// managedReaderUsers contains the list of users that will be added to the managed reader group (with all view rights)
	ManagedReaderUsers []ManagedUser `json:"managedReaderUsers,omitempty"`
	// readOnlyConfig specifies the read-only type Nifi config cluster wide, all theses
	// will be merged with node specified readOnly configurations, so it can be overwritten per node.
	ReadOnlyConfig ReadOnlyConfig `json:"readOnlyConfig,omitempty"`
	// nodeConfigGroups specifies multiple node configs with unique name
	NodeConfigGroups map[string]NodeConfig `json:"nodeConfigGroups,omitempty"`
	// all node requires an image, unique id, and storageConfigs settings
	Nodes []Node `json:"nodes"`
	// Defines the configuration for PodDisruptionBudget
	DisruptionBudget DisruptionBudget `json:"disruptionBudget,omitempty"`
	// LdapConfiguration specifies the configuration if you want to use LDAP
	LdapConfiguration LdapConfiguration `json:"ldapConfiguration,omitempty"`
	// NifiClusterTaskSpec specifies the configuration of the nifi cluster Tasks
	NifiClusterTaskSpec NifiClusterTaskSpec `json:"nifiClusterTaskSpec,omitempty"`
	// TODO : add vault
	//VaultConfig         	VaultConfig         `json:"vaultConfig,omitempty"`
	// listenerConfig specifies nifi's listener specifig configs
	ListenersConfig *ListenersConfig `json:"listenersConfig,omitempty"`
	// SidecarsConfig defines additional sidecar configurations
	SidecarConfigs []corev1.Container `json:"sidecarConfigs,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`
	// ExternalService specifies settings required to access nifi externally
	ExternalServices []ExternalServiceConfig `json:"externalServices,omitempty"`
}

// DisruptionBudget defines the configuration for PodDisruptionBudget
type DisruptionBudget struct {
	// If set to true, will create a podDisruptionBudget
	// +optional
	Create bool `json:"create,omitempty"`
	// The budget to set for the PDB, can either be static number or a percentage
	// +kubebuilder:validation:Pattern:="^[0-9]+$|^[0-9]{1,2}%$|^100%$"
	Budget string `json:"budget,omitempty"`
}

type ServicePolicy struct {
	// HeadlessEnabled specifies if the cluster should use headlessService for Nifi or individual services
	// using service per nodes may come an handy case of service mesh.
	HeadlessEnabled bool `json:"headlessEnabled"`
	// Annotations specifies the annotations to attach to services the operator creates
	Annotations map[string]string `json:"annotations,omitempty"`
}

type PodPolicy struct {
	// Annotations specifies the annotations to attach to pods the operator creates
	Annotations map[string]string `json:"annotations,omitempty"`
}

// rollingUpgradeConfig specifies the rolling upgrade config for the cluster
//RollingUpgradeConfig 	RollingUpgradeConfig 	`json:"rollingUpgradeConfig"`

// RollingUpgradeStatus defines status of rolling upgrade
type RollingUpgradeStatus struct {
	//
	LastSuccess string `json:"lastSuccess"`
	//
	ErrorCount int `json:"errorCount"`
}

// RollingUpgradeConfig defines the desired config of the RollingUpgrade
/*type RollingUpgradeConfig struct {
	// failureThreshold states that how many errors can the cluster tolerate during rolling upgrade
	FailureThreshold	int	`json:"failureThreshold"`
}*/

// Node defines the nifi node basic configuration
type Node struct {
	// Unique Node id
	Id int32 `json:"id"`
	// nodeConfigGroup can be used to ease the node configuration, if set only the id is required
	NodeConfigGroup string `json:"nodeConfigGroup,omitempty"`
	// readOnlyConfig can be used to pass Nifi node config https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html
	// which has type read-only these config changes will trigger rolling upgrade
	ReadOnlyConfig *ReadOnlyConfig `json:"readOnlyConfig,omitempty"`
	// node configuration
	NodeConfig *NodeConfig `json:"nodeConfig,omitempty"`
}

type ReadOnlyConfig struct {
	// MaximumTimerDrivenThreadCount define the maximum number of threads for timer driven processors available to the system.
	MaximumTimerDrivenThreadCount *int32 `json:"maximumTimerDrivenThreadCount,omitempty"`
	// AdditionalSharedEnvs define a set of additional env variables that will shared between all init containers and
	// containers in the pod.
	AdditionalSharedEnvs []corev1.EnvVar `json:"additionalSharedEnvs,omitempty"`
	// NifiProperties configuration that will be applied to the node.
	NifiProperties NifiProperties `json:"nifiProperties,omitempty"`
	// ZookeeperProperties configuration that will be applied to the node.
	ZookeeperProperties ZookeeperProperties `json:"zookeeperProperties,omitempty"`
	// BootstrapProperties configuration that will be applied to the node.
	BootstrapProperties BootstrapProperties `json:"bootstrapProperties,omitempty"`
	// Logback configuration that will be applied to the node.
	LogbackConfig LogbackConfig `json:"logbackConfig,omitempty"`
	// BootstrapNotificationServices configuration that will be applied to the node.
	BootstrapNotificationServicesReplaceConfig BootstrapNotificationServicesConfig `json:"bootstrapNotificationServicesConfig,omitempty"`
}

// NifiProperties configuration that will be applied to the node.
type NifiProperties struct {
	// Additionnals nifi.properties configuration that will override the one produced based on template and
	// configuration
	OverrideConfigMap *ConfigmapReference `json:"overrideConfigMap,omitempty"`
	// Additionnals nifi.properties configuration that will override the one produced based
	// on template, configurations and overrideConfigMap.
	OverrideConfigs string `json:"overrideConfigs,omitempty"`
	// Additionnals nifi.properties configuration that will override the one produced based
	// on template, configurations, overrideConfigMap and overrideConfigs.
	OverrideSecretConfig *SecretConfigReference `json:"overrideSecretConfig,omitempty"`
	// A comma separated list of allowed HTTP Host header values to consider when NiFi
	// is running securely and will be receiving requests to a different host[:port] than it is bound to.
	// https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#web-properties
	WebProxyHosts []string `json:"webProxyHosts,omitempty"`
	// Nifi security client auth
	NeedClientAuth bool `json:"needClientAuth,omitempty"`
	// Indicates which of the configured authorizers in the authorizers.xml file to use
	// https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#authorizer-configuration
	Authorizer string `json:"authorizer,omitempty"`
}

// ZookeeperProperties configuration that will be applied to the node.
type ZookeeperProperties struct {
	// Additionnals zookeeper.properties configuration that will override the one produced based on template and
	// configuration
	OverrideConfigMap *ConfigmapReference `json:"overrideConfigMap,omitempty"`
	// Additionnals zookeeper.properties configuration that will override the one produced based
	// on template and configurations.
	OverrideConfigs string `json:"overrideConfigs,omitempty"`
	// Additionnals zookeeper.properties configuration that will override the one produced based
	// on template, configurations, overrideConfigMap and overrideConfigs.
	OverrideSecretConfig *SecretConfigReference `json:"overrideSecretConfig,omitempty"`
}

// BootstrapProperties configuration that will be applied to the node.
type BootstrapProperties struct {
	// JVM memory settings
	NifiJvmMemory string `json:"nifiJvmMemory,omitempty"`
	// Additionnals bootstrap.properties configuration that will override the one produced based on template and
	// configuration
	OverrideConfigMap *ConfigmapReference `json:"overrideConfigMap,omitempty"`
	// Additionnals bootstrap.properties configuration that will override the one produced based
	// on template and configurations.
	OverrideConfigs string `json:"overrideConfigs,omitempty"`
	// Additionnals bootstrap.properties configuration that will override the one produced based
	// on template, configurations, overrideConfigMap and overrideConfigs.
	OverrideSecretConfig *SecretConfigReference `json:"overrideSecretConfig,omitempty"`
}

// Logback configuration that will be applied to the node.
type LogbackConfig struct {
	// logback.xml configuration that will replace the one produced based on template
	ReplaceConfigMap *ConfigmapReference `json:"replaceConfigMap,omitempty"`
	// logback.xml configuration that will replace the one produced based on template and overrideConfigMap
	ReplaceSecretConfig *SecretConfigReference `json:"replaceSecretConfig,omitempty"`
}

type BootstrapNotificationServicesConfig struct {
	// bootstrap_notifications_services.xml configuration that will replace the one produced based on template
	ReplaceConfigMap *ConfigmapReference `json:"replaceConfigMap,omitempty"`
	// bootstrap_notifications_services.xml configuration that will replace the one produced based on template and overrideConfigMap
	ReplaceSecretConfig *SecretConfigReference `json:"replaceSecretConfig,omitempty"`
}

// NodeConfig defines the node configuration
type NodeConfig struct {
	// provenanceStorage allow to specify the maximum amount of data provenance information to store at a time
	// https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#write-ahead-provenance-repository-properties
	ProvenanceStorage string `json:"provenanceStorage,omitempty"`
	//RunAsUser define the id of the user to run in the Nifi image
	// +kubebuilder:validation:Minimum=1
	RunAsUser *int64 `json:"runAsUser,omitempty"`
	// FSGroup define the id of the group for each volumes in Nifi image
	// +kubebuilder:validation:Minimum=1
	FSGroup *int64 `json:"fsGroup,omitempty"`
	// Set this to true if the instance is a node in a cluster.
	// https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#basic-cluster-setup
	IsNode *bool `json:"isNode,omitempty"`
	//  Docker image used by the operator to create the node associated
	//  https://hub.docker.com/r/apache/nifi/
	Image string `json:"image,omitempty"`
	// imagePullPolicy define the pull policy for NiFi cluster docker image
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// nodeAffinity can be specified, operator populates this value if new pvc added later to node
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`
	// storageConfigs specifies the node related configs
	StorageConfigs []StorageConfig `json:"storageConfigs,omitempty"`
	// serviceAccountName specifies the serviceAccount used for this specific node
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// resourceRequirements works exactly like Container resources, the user can specify the limit and the requests
	// through this property
	// https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	ResourcesRequirements *corev1.ResourceRequirements `json:"resourcesRequirements,omitempty"`
	// imagePullSecrets specifies the secret to use when using private registry
	// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#localobjectreference-v1-core
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// nodeSelector can be specified, which set the pod to fit on a node
	// https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// tolerations can be specified, which set the pod's tolerations
	// https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/#concepts
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Additionnal annotation to attach to the pod associated
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/#syntax-and-character-set
	NodeAnnotations map[string]string `json:"nifiAnnotations,omitempty"`
}

// StorageConfig defines the node storage configuration
type StorageConfig struct {
	// Name of the storage config, used to name PV to reuse into sidecars for example.
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	Name string `json:"name"`
	// Path where the volume will be mount into the main nifi container inside the pod.
	MountPath string `json:"mountPath"`
	// Kubernetes PVC spec
	PVCSpec *corev1.PersistentVolumeClaimSpec `json:"pvcSpec"`
}

//ListenersConfig defines the Nifi listener types
type ListenersConfig struct {
	// internalListeners specifies settings required to access nifi internally
	InternalListeners []InternalListenerConfig `json:"internalListeners"`
	// sslSecrets contains information about ssl related kubernetes secrets if one of the
	// listener setting type set to ssl these fields must be populated to
	SSLSecrets *SSLSecrets `json:"sslSecrets,omitempty"`
	// clusterDomain allow to override the default cluster domain which is "cluster.local"
	ClusterDomain string `json:"clusterDomain,omitempty"`
	// useExternalDNS allow to manage externalDNS usage by limiting the DNS names associated
	// to each nodes and load balancer : <cluster-name>-node-<node Id>.<cluster-name>.<service name>.<cluster domain>
	UseExternalDNS bool `json:"useExternalDNS,omitempty"`
}

// SSLSecrets defines the Nifi SSL secrets
type SSLSecrets struct {
	// tlsSecretName should contain all ssl certs required by nifi including: caCert, caKey, clientCert, clientKey
	// serverCert, serverKey, peerCert, peerKey
	TLSSecretName string `json:"tlsSecretName"`
	// create tells the installed cert manager to create the required certs keys
	Create bool `json:"create,omitempty"`
	// clusterScoped defines if the Issuer created is cluster or namespace scoped
	ClusterScoped bool `json:"clusterScoped,omitempty"`
	// issuerRef allow to use an existing issuer to act as CA :
	// https://cert-manager.io/docs/concepts/issuer/
	IssuerRef *cmmeta.ObjectReference `json:"issuerRef,omitempty"`
	// TODO : add vault
	// +kubebuilder:validation:Enum={"cert-manager","vault"}
	PKIBackend PKIBackend `json:"pkiBackend,omitempty"`
	//,"vault"
}

// TODO : Add vault
// VaultConfig defines the configuration for a vault PKI backend
/*type VaultConfig struct {
	//
	AuthRole  string `json:"authRole"`
	//
	PKIPath   string `json:"pkiPath"`
	//
	IssuePath string `json:"issuePath"`
	//
	UserStore string `json:"userStore"`
}*/

// InternalListenerConfig defines the internal listener config for Nifi
type InternalListenerConfig struct {
	// +kubebuilder:validation:Enum={"cluster", "http", "https", "s2s", "prometheus"}
	// (Optional field) Type allow to specify if we are in a specific nifi listener
	// it's allowing to define some required information such as Cluster Port,
	// Http Port, Https Port or S2S port
	Type string `json:"type,omitempty"`
	// An identifier for the port which will be configured.
	Name string `json:"name"`
	// The container port.
	ContainerPort int32 `json:"containerPort"`
}

type ExternalServiceConfig struct {
	// Name must be unique within a namespace. Is required when creating resources, although
	// some resources may allow a client to request the generation of an appropriate name
	// automatically. Name is primarily intended for creation idempotence and configuration
	// definition.
	// Cannot be updated.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#names
	// +optional
	Name string `json:"name"`
	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`
	// Spec defines the behavior of a service.
	Spec ExternalServiceSpec `json:"spec"`
}

type ExternalServiceSpec struct {
	// Contains the list port for the service and the associated listener
	PortConfigs []PortConfig `json:"portConfigs"`
	// clusterIP is the IP address of the service and is usually assigned
	// randomly by the master. If an address is specified manually and is not in
	// use by others, it will be allocated to the service; otherwise, creation
	// of the service will fail. This field can not be changed through updates.
	// Valid values are "None", empty string (""), or a valid IP address. "None"
	// can be specified for headless services when proxying is not required.
	// Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if
	// type is ExternalName.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies
	// +optional
	ClusterIP string `json:"clusterIP,omitempty" protobuf:"bytes,3,opt,name=clusterIP"`
	// type determines how the Service is exposed. Defaults to ClusterIP. Valid
	// options are ExternalName, ClusterIP, NodePort, and LoadBalancer.
	// "ExternalName" maps to the specified externalName.
	// "ClusterIP" allocates a cluster-internal IP address for load-balancing to
	// endpoints. Endpoints are determined by the selector or if that is not
	// specified, by manual construction of an Endpoints object. If clusterIP is
	// "None", no virtual IP is allocated and the endpoints are published as a
	// set of endpoints rather than a stable IP.
	// "NodePort" builds on ClusterIP and allocates a port on every node which
	// routes to the clusterIP.
	// "LoadBalancer" builds on NodePort and creates an
	// external load-balancer (if supported in the current cloud) which routes
	// to the clusterIP.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
	// +optional
	Type corev1.ServiceType `json:"type,omitempty" protobuf:"bytes,4,opt,name=type,casttype=ServiceType"`
	// externalIPs is a list of IP addresses for which nodes in the cluster
	// will also accept traffic for this service.  These IPs are not managed by
	// Kubernetes.  The user is responsible for ensuring that traffic arrives
	// at a node with this IP.  A common example is external load-balancers
	// that are not part of the Kubernetes system.
	// +optional
	ExternalIPs []string `json:"externalIPs,omitempty" protobuf:"bytes,5,rep,name=externalIPs"`
	// Only applies to Service Type: LoadBalancer
	// LoadBalancer will get created with the IP specified in this field.
	// This feature depends on whether the underlying cloud-provider supports specifying
	// the loadBalancerIP when a load balancer is created.
	// This field will be ignored if the cloud-provider does not support the feature.
	// +optional
	LoadBalancerIP string `json:"loadBalancerIP,omitempty" protobuf:"bytes,8,opt,name=loadBalancerIP"`
	// If specified and supported by the platform, this will restrict traffic through the cloud-provider
	// load-balancer will be restricted to the specified client IPs. This field will be ignored if the
	// cloud-provider does not support the feature."
	// More info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/
	// +optional
	LoadBalancerSourceRanges []string `json:"loadBalancerSourceRanges,omitempty" protobuf:"bytes,9,opt,name=loadBalancerSourceRanges"`
	// externalName is the external reference that kubedns or equivalent will
	// return as a CNAME record for this service. No proxying will be involved.
	// Must be a valid RFC-1123 hostname (https://tools.ietf.org/html/rfc1123)
	// and requires Type to be ExternalName.
	// +optional
	ExternalName string `json:"externalName,omitempty" protobuf:"bytes,10,opt,name=externalName"`
}

type PortConfig struct {
	// The port that will be exposed by this service.
	Port int32 `json:"port" protobuf:"varint,3,opt,name=port"`
	// The name of the listener which will be used as target container.
	InternalListenerName string `json:"internalListenerName"`
}

// LdapConfiguration specifies the configuration if you want to use LDAP
type LdapConfiguration struct {
	// If set to true, we will enable ldap usage into nifi.properties configuration.
	Enabled bool `json:"enabled,omitempty"`
	// Space-separated list of URLs of the LDAP servers (i.e. ldap://<hostname>:<port>).
	Url string `json:"url,omitempty"`
	// Base DN for searching for users (i.e. CN=Users,DC=example,DC=com).
	SearchBase string `json:"searchBase,omitempty"`
	// Filter for searching for users against the 'User Search Base'.
	// (i.e. sAMAccountName={0}). The user specified name is inserted into '{0}'.
	SearchFilter string `json:"searchFilter,omitempty"`
}

// NifiClusterTaskSpec specifies the configuration of the nifi cluster Tasks
type NifiClusterTaskSpec struct {
	// RetryDurationMinutes describes the amount of time the Operator waits for the task
	RetryDurationMinutes int `json:"retryDurationMinutes"`
}

// NifiClusterStatus defines the observed state of NifiCluster
type NifiClusterStatus struct {
	// Store the state of each nifi node
	NodesState map[string]NodeState `json:"nodesState,omitempty"`
	// ClusterState holds info about the cluster state
	State ClusterState `json:"state"`
	// RollingUpgradeStatus defines status of rolling upgrade
	RollingUpgrade RollingUpgradeStatus `json:"rollingUpgradeStatus,omitempty"`
	// RootProcessGroupId contains the uuid of the root process group for this cluster
	RootProcessGroupId string `json:"rootProcessGroupId,omitempty"`
	// PrometheusReportingTask contains the status of the prometheus reporting task managed by the operator
	PrometheusReportingTask PrometheusReportingTaskStatus `json:"prometheusReportingTask,omitempty"`
}

type PrometheusReportingTaskStatus struct {
	// The nifi reporting task's id
	Id string `json:"id"`
	// The last nifi reporting task revision version catched
	Version int64 `json:"version"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NifiCluster is the Schema for the nificlusters API
type NifiCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiClusterSpec   `json:"spec,omitempty"`
	Status NifiClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NifiClusterList contains a list of NifiCluster
type NifiClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiCluster{}, &NifiClusterList{})
}

type ManagedUser struct {
	// identity field is use to define the user identity on NiFi cluster side,
	// it use full when the user's name doesn't suite with Kubernetes resource name.
	Identity string `json:"identity,omitempty"`
	// name field is use to name the NifiUser resource, if not identity is provided it will be used to name
	// the user on NiFi cluster side.
	Name string `json:"name"`
}

func (u *ManagedUser) GetIdentity() string {
	if u.Identity == "" {
		return u.Name
	}
	return u.Identity
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

func (nSpec *NifiClusterSpec) GetInitContainerImage() string {

	if nSpec.InitContainerImage == "" {
		return "busybox"
	}
	return nSpec.InitContainerImage
}

func (lConfig *ListenersConfig) GetClusterDomain() string {
	if len(lConfig.ClusterDomain) == 0 {
		return "cluster.local"
	}

	return lConfig.ClusterDomain
}

func (nReadOnlyConfig *ReadOnlyConfig) GetMaximumTimerDrivenThreadCount() int32 {
	if nReadOnlyConfig.MaximumTimerDrivenThreadCount == nil {
		return 10
	}
	return *nReadOnlyConfig.MaximumTimerDrivenThreadCount
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

//GetImagePullPolicy returns the image pull policy to pull containers images
func (nConfig *NodeConfig) GetImagePullPolicy() corev1.PullPolicy {
	return nConfig.ImagePullPolicy
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

func (nConfig *NodeConfig) GetFSGroup() *int64 {
	var defaultGroupID int64 = 1000
	if nConfig.FSGroup != nil {
		return nConfig.FSGroup
	}

	return func(i int64) *int64 { return &i }(defaultGroupID)
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

//
func (nSpec *NifiClusterSpec) GetMetricPort() *int {

	for _, iListener := range nSpec.ListenersConfig.InternalListeners {
		if iListener.Type == prometheusListenerType {
			val := int(iListener.ContainerPort)
			return &val
		}
	}

	return nil
}

func (cluster *NifiCluster) RootProcessGroupId() string {
	return cluster.Status.RootProcessGroupId
}

func (c *NifiCluster) GetClientType() ClientConfigType {
	if c.Spec.ClientType == "" {
		return ClientConfigTLS
	}
	return c.Spec.ClientType
}

func (c *NifiCluster) GetType() ClusterType {
	if c.Spec.Type == "" {
		return InternalCluster
	}
	return ExternalCluster
}

func (c *NifiCluster) IsSet() bool {
	return (c.GetType() == InternalCluster && len(c.Name) != 0) ||
		(c.GetType() != ExternalCluster && len(c.Spec.NodeURITemplate) != 0 && len(c.Spec.RootProcessGroupId) != 0)
}

func (c *NifiCluster) IsInternal() bool {
	return c.GetType() == InternalCluster
}

func (c NifiCluster) IsExternal() bool {
	return c.GetType() != InternalCluster
}

func (cluster NifiCluster) IsReady() bool {
	for _, nodeState := range cluster.Status.NodesState {
		if nodeState.ConfigurationState != ConfigInSync || nodeState.GracefulActionState.State != GracefulUpscaleSucceeded ||
			!nodeState.PodIsReady {
			return false
		}
	}
	return cluster.Status.State.IsReady()
}

func (cluster *NifiCluster) Id() string {
	return cluster.Name
}
