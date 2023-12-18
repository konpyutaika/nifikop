package nificlient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"emperror.dev/errors"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"

	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

const (
	PRIMARY_NODE        = "Primary Node"
	CLUSTER_COORDINATOR = "Cluster Coordinator"
	// ConnectNodeAction states that the NiFi node is connecting to the NiFi Cluster.
	CONNECTING_STATUS = "CONNECTING"
	// ConnectStatus states that the NiFi node is connected to the NiFi Cluster.
	CONNECTED_STATUS = "CONNECTED"
	// DisconnectNodeAction states that the NiFi node is disconnecting from NiFi Cluster.
	DISCONNECTING_STATUS = "DISCONNECTING"
	// DisconnectStatus states that the NiFi node is disconnected from NiFi Cluster.
	DISCONNECTED_STATUS = "DISCONNECTED"
	// OffloadNodeAction states that the NiFi node is offloading data to NiFi Cluster.
	OFFLOADING_STATUS = "OFFLOADING"
	// OffloadStatus states that the NiFi node offloaded data to NiFi Cluster.
	OFFLOADED_STATUS = "OFFLOADED"
	// RemoveNodeAction states that the NiFi node is removing from NiFi Cluster.
	REMOVING_STATUS = "REMOVING"
	// RemoveStatus states that the NiFi node is removed from NiFi Cluster.
	REMOVED_STATUS = "REMOVED"
)

// NiFiClient is the exported interface for NiFi operations.
type NifiClient interface {
	// Access func
	CreateAccessTokenUsingBasicAuth(username, password string, nodeId int32) (*string, error)

	// System func
	DescribeCluster() (*nigoapi.ClusterEntity, error)
	DescribeClusterFromNodeId(nodeId int32) (*nigoapi.ClusterEntity, error)
	DisconnectClusterNode(nId int32) (*nigoapi.NodeEntity, error)
	ConnectClusterNode(nId int32) (*nigoapi.NodeEntity, error)
	OffloadClusterNode(nId int32) (*nigoapi.NodeEntity, error)
	RemoveClusterNode(nId int32) error
	GetClusterNode(nId int32) (*nigoapi.NodeEntity, error)
	RemoveClusterNodeFromClusterNodeId(nId string) error

	// Registry client func
	GetRegistryClient(id string) (*nigoapi.FlowRegistryClientEntity, error)
	CreateRegistryClient(entity nigoapi.FlowRegistryClientEntity) (*nigoapi.FlowRegistryClientEntity, error)
	UpdateRegistryClient(entity nigoapi.FlowRegistryClientEntity) (*nigoapi.FlowRegistryClientEntity, error)
	RemoveRegistryClient(entity nigoapi.FlowRegistryClientEntity) error

	// Flow client func
	GetFlow(id string) (*nigoapi.ProcessGroupFlowEntity, error)
	UpdateFlowControllerServices(entity nigoapi.ActivateControllerServicesEntity) (*nigoapi.ActivateControllerServicesEntity, error)
	UpdateFlowProcessGroup(entity nigoapi.ScheduleComponentsEntity) (*nigoapi.ScheduleComponentsEntity, error)
	GetFlowControllerServices(id string) (*nigoapi.ControllerServicesEntity, error)

	// Drop request func
	GetDropRequest(connectionId, id string) (*nigoapi.DropRequestEntity, error)
	CreateDropRequest(connectionId string) (*nigoapi.DropRequestEntity, error)

	// Process Group func
	GetProcessGroup(id string) (*nigoapi.ProcessGroupEntity, error)
	CreateProcessGroup(entity nigoapi.ProcessGroupEntity, pgParentId string) (*nigoapi.ProcessGroupEntity, error)
	UpdateProcessGroup(entity nigoapi.ProcessGroupEntity) (*nigoapi.ProcessGroupEntity, error)
	RemoveProcessGroup(entity nigoapi.ProcessGroupEntity) error
	CreateConnection(entity nigoapi.ConnectionEntity) (*nigoapi.ConnectionEntity, error)

	// Version func
	CreateVersionUpdateRequest(pgId string, entity nigoapi.VersionControlInformationEntity) (*nigoapi.VersionedFlowUpdateRequestEntity, error)
	GetVersionUpdateRequest(id string) (*nigoapi.VersionedFlowUpdateRequestEntity, error)
	CreateVersionRevertRequest(pgId string, entity nigoapi.VersionControlInformationEntity) (*nigoapi.VersionedFlowUpdateRequestEntity, error)
	GetVersionRevertRequest(id string) (*nigoapi.VersionedFlowUpdateRequestEntity, error)

	// Snippet func
	CreateSnippet(entity nigoapi.SnippetEntity) (*nigoapi.SnippetEntity, error)
	UpdateSnippet(entity nigoapi.SnippetEntity) (*nigoapi.SnippetEntity, error)

	// Processor func
	UpdateProcessor(entity nigoapi.ProcessorEntity) (*nigoapi.ProcessorEntity, error)
	UpdateProcessorRunStatus(id string, entity nigoapi.ProcessorRunStatusEntity) (*nigoapi.ProcessorEntity, error)
	GetProcessor(id string) (*nigoapi.ProcessorEntity, error)

	// Input port func
	UpdateInputPortRunStatus(id string, entity nigoapi.PortRunStatusEntity) (*nigoapi.ProcessorEntity, error)
	GetInputPort(id string) (*nigoapi.PortEntity, error)

	// Output port func
	UpdateOutputPortRunStatus(id string, entity nigoapi.PortRunStatusEntity) (*nigoapi.ProcessorEntity, error)
	GetOutputPort(id string) (*nigoapi.PortEntity, error)

	// Parameter context func
	GetParameterContexts() ([]nigoapi.ParameterContextEntity, error)
	GetParameterContext(id string) (*nigoapi.ParameterContextEntity, error)
	CreateParameterContext(entity nigoapi.ParameterContextEntity) (*nigoapi.ParameterContextEntity, error)
	RemoveParameterContext(entity nigoapi.ParameterContextEntity) error
	CreateParameterContextUpdateRequest(contextId string, entity nigoapi.ParameterContextEntity) (*nigoapi.ParameterContextUpdateRequestEntity, error)
	GetParameterContextUpdateRequest(contextId, id string) (*nigoapi.ParameterContextUpdateRequestEntity, error)

	// User groups func
	GetUserGroups() ([]nigoapi.UserGroupEntity, error)
	GetUserGroup(id string) (*nigoapi.UserGroupEntity, error)
	CreateUserGroup(entity nigoapi.UserGroupEntity) (*nigoapi.UserGroupEntity, error)
	UpdateUserGroup(entity nigoapi.UserGroupEntity) (*nigoapi.UserGroupEntity, error)
	RemoveUserGroup(entity nigoapi.UserGroupEntity) error

	// User func
	GetUsers() ([]nigoapi.UserEntity, error)
	GetUser(id string) (*nigoapi.UserEntity, error)
	CreateUser(entity nigoapi.UserEntity) (*nigoapi.UserEntity, error)
	UpdateUser(entity nigoapi.UserEntity) (*nigoapi.UserEntity, error)
	RemoveUser(entity nigoapi.UserEntity) error

	// Policies func
	GetAccessPolicy(action, resource string) (*nigoapi.AccessPolicyEntity, error)
	CreateAccessPolicy(entity nigoapi.AccessPolicyEntity) (*nigoapi.AccessPolicyEntity, error)
	UpdateAccessPolicy(entity nigoapi.AccessPolicyEntity) (*nigoapi.AccessPolicyEntity, error)
	RemoveAccessPolicy(entity nigoapi.AccessPolicyEntity) error

	// Reportingtask func
	GetReportingTask(id string) (*nigoapi.ReportingTaskEntity, error)
	CreateReportingTask(entity nigoapi.ReportingTaskEntity) (*nigoapi.ReportingTaskEntity, error)
	UpdateReportingTask(entity nigoapi.ReportingTaskEntity) (*nigoapi.ReportingTaskEntity, error)
	UpdateRunStatusReportingTask(id string, entity nigoapi.ReportingTaskRunStatusEntity) (*nigoapi.ReportingTaskEntity, error)
	RemoveReportingTask(entity nigoapi.ReportingTaskEntity) error

	// ControllerConfig func
	GetControllerConfig() (*nigoapi.ControllerConfigurationEntity, error)
	UpdateControllerConfig(entity nigoapi.ControllerConfigurationEntity) (*nigoapi.ControllerConfigurationEntity, error)

	// Connections func
	GetConnection(id string) (*nigoapi.ConnectionEntity, error)
	UpdateConnection(entity nigoapi.ConnectionEntity) (*nigoapi.ConnectionEntity, error)
	DeleteConnection(entity nigoapi.ConnectionEntity) error

	Build() error
}

type nifiClient struct {
	NifiClient
	log        *zap.Logger
	opts       *clientconfig.NifiConfig
	client     *nigoapi.APIClient
	nodeClient map[int32]*nigoapi.APIClient
	timeout    time.Duration
	nodes      []nigoapi.NodeDto

	// client funcs for mocking
	newClient func(*nigoapi.Configuration) *nigoapi.APIClient
}

func New(opts *clientconfig.NifiConfig, logger *zap.Logger) NifiClient {
	nClient := &nifiClient{
		log:     logger,
		opts:    opts,
		timeout: time.Duration(opts.OperationTimeout) * time.Second,
	}

	nClient.newClient = nigoapi.NewAPIClient
	return nClient
}

func (n *nifiClient) Build() error {
	config := n.getNifiGoApiConfig()
	n.client = n.newClient(config)

	n.nodeClient = make(map[int32]*nigoapi.APIClient)
	for nodeId := range n.opts.NodesURI {
		nodeConfig := n.getNiNodeGoApiConfig(nodeId)
		n.nodeClient[nodeId] = n.newClient(nodeConfig)
	}

	if !n.opts.SkipDescribeCluster {
		clusterEntity, err := n.DescribeCluster()
		if err != nil || clusterEntity == nil || clusterEntity.Cluster == nil {
			err = errorfactory.New(errorfactory.NodesUnreachable{}, err, fmt.Sprintf("could not connect to nifi nodes: %s", n.opts.NifiURI))
			return err
		}

		n.nodes = clusterEntity.Cluster.Nodes
	}

	return nil
}

// NewFromConfig is a convenient wrapper around New() and ClusterConfig().
func NewFromConfig(opts *clientconfig.NifiConfig, logger *zap.Logger) (NifiClient, error) {
	var client NifiClient
	var err error

	if opts == nil {
		return nil, errorfactory.New(errorfactory.NilClientConfig{}, errors.New("The NiFi client config is nil"), "The NiFi client config is nil")
	}
	client = New(opts, logger)
	err = client.Build()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (n *nifiClient) getNifiGoApiConfig() (config *nigoapi.Configuration) {
	config = nigoapi.NewConfiguration()

	protocol := "http"
	var transport *http.Transport = nil
	if n.opts.UseSSL {
		transport = &http.Transport{}
		config.Scheme = "HTTPS"
		transport.TLSClientConfig = n.opts.TLSConfig
		protocol = "https"
	}

	if len(n.opts.ProxyUrl) > 0 {
		proxyUrl, err := url.Parse(n.opts.ProxyUrl)
		if err == nil {
			if transport == nil {
				transport = &http.Transport{}
			}
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
	}

	config.HTTPClient = &http.Client{}
	if transport != nil {
		config.HTTPClient = &http.Client{Transport: transport}
	}

	config.BasePath = fmt.Sprintf("%s://%s/nifi-api", protocol, n.opts.NifiURI)
	config.Host = n.opts.NifiURI
	if len(n.opts.NifiURI) == 0 {
		for nodeId := range n.opts.NodesURI {
			config.BasePath = fmt.Sprintf("%s://%s/nifi-api", protocol, n.opts.NodesURI[nodeId].RequestHost)
			config.Host = n.opts.NodesURI[nodeId].RequestHost
		}
	}
	return
}

func (n *nifiClient) getNiNodeGoApiConfig(nodeId int32) (config *nigoapi.Configuration) {
	config = nigoapi.NewConfiguration()
	config.HTTPClient = &http.Client{}
	protocol := "http"

	var transport *http.Transport = nil
	if n.opts.UseSSL {
		transport = &http.Transport{}
		config.Scheme = "HTTPS"
		transport.TLSClientConfig = n.opts.TLSConfig
		protocol = "https"
	}

	if n.opts.ProxyUrl != "" {
		proxyUrl, err := url.Parse(n.opts.ProxyUrl)
		if err == nil {
			if transport == nil {
				transport = &http.Transport{}
			}
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
	}
	config.HTTPClient = &http.Client{}
	if transport != nil {
		config.HTTPClient = &http.Client{Transport: transport}
	}

	config.BasePath = fmt.Sprintf("%s://%s/nifi-api", protocol, n.opts.NodesURI[nodeId].RequestHost)
	config.Host = n.opts.NodesURI[nodeId].RequestHost
	if len(n.opts.NifiURI) != 0 {
		config.Host = n.opts.NifiURI
	}

	return
}

func (n *nifiClient) privilegeCoordinatorClient() (*nigoapi.APIClient, context.Context) {
	if clientId := n.coordinatorNodeId(); clientId != nil {
		return n.nodeClient[*clientId], n.opts.NodesContext[*clientId]
	}

	if clientId := n.privilegeNodeClient(); clientId != nil {
		return n.nodeClient[*clientId], n.opts.NodesContext[*clientId]
	}

	return n.client, nil
}

func (n *nifiClient) privilegeCoordinatorExceptNodeIdClient(nId int32) (*nigoapi.APIClient, context.Context) {
	nodeDto := n.nodeDtoByNodeId(nId)
	if nodeDto == nil || isCoordinator(nodeDto) {
		if clientId := n.firstConnectedNodeId(nId); clientId != nil {
			return n.nodeClient[*clientId], n.opts.NodesContext[*clientId]
		}
	}

	return n.privilegeCoordinatorClient()
}

// TODO : change logic by binding in status the nodeId with the Nifi Cluster Node id ?
func (n *nifiClient) firstConnectedNodeId(excludeId int32) *int32 {
	// Convert nodeId to a Cluster Node for the one to exclude
	excludedNodeDto := n.nodeDtoByNodeId(excludeId)
	// For each NiFi Cluster Node
	for id := range n.nodes {
		nodeDto := n.nodes[id]
		// Check that it's not the one exclueded and it is Connected
		if excludedNodeDto == nil || (nodeDto.NodeId != excludedNodeDto.NodeId && isConnected(excludedNodeDto)) {
			// Check that a Node exist in the NifiCluster definition, and that we have a client associated
			if nId := n.nodeIdByNodeDto(&nodeDto); nId != nil {
				return nId
			}
		}
	}
	return nil
}

func (n *nifiClient) coordinatorNodeId() *int32 {
	for id := range n.nodes {
		nodeDto := n.nodes[id]
		// We return the Node Id associated to the Cluster Node coordinator, if it is connected
		if isCoordinator(&nodeDto) && isConnected(&nodeDto) {
			return n.nodeIdByNodeDto(&nodeDto)
		}
	}
	return nil
}

func (n *nifiClient) privilegeNodeClient() *int32 {
	for id := range n.nodeClient {
		return &id
	}
	return nil
}

func isCoordinator(node *nigoapi.NodeDto) bool {
	// For each role looking that it contains the Coordinator one.
	for _, role := range node.Roles {
		if role == CLUSTER_COORDINATOR {
			return true
		}
	}
	return false
}

func isConnected(node *nigoapi.NodeDto) bool {
	return node.Status == CONNECTED_STATUS
}

func (n *nifiClient) nodeDtoByNodeId(nId int32) *nigoapi.NodeDto {
	for id := range n.nodes {
		nodeDto := n.nodes[id]
		// Check if the Cluster Node uri match with the one associated to the NifiCluster nodeId searched
		if fmt.Sprintf("%s:%d", nodeDto.Address, nodeDto.ApiPort) == fmt.Sprintf(n.opts.NodeURITemplate, nId) {
			return &nodeDto
		}
	}
	return nil
}

func (n *nifiClient) nodeIdByNodeDto(nodeDto *nigoapi.NodeDto) *int32 {
	// Extract the uri associated to the Cluster Node
	searchedUri := fmt.Sprintf("%s:%d", nodeDto.Address, nodeDto.ApiPort)
	// For each uri generated from NifiCluster resources node defined
	for id, uri := range n.opts.NodesURI {
		// Check if we find a match
		if uri.HostListener == searchedUri {
			findId := id
			return &findId
		}
	}

	return nil
}

func (n *nifiClient) setNodeFromNodes(nodeDto *nigoapi.NodeDto) {
	for id := range n.nodes {
		if n.nodes[id].NodeId == nodeDto.NodeId {
			n.nodes[id] = *nodeDto
			break
		}
	}
}
