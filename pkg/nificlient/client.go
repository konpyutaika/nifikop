// Copyright 2020 Orange SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package apis
package nificlient

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("nifi_client")

const (
	PRIMARY_NODE        = "Primary Node"
	CLUSTER_COORDINATOR = "Cluster Coordinator"
)

// NiFiClient is the exported interface for NiFi operations
type NifiClient interface {
	// System func
	DescribeCluster() (*nigoapi.ClusterEntity, error)
	DisconnectClusterNode(nId int32) (*nigoapi.NodeEntity, error)
	ConnectClusterNode(nId int32) (*nigoapi.NodeEntity, error)
	OffloadClusterNode(nId int32) (*nigoapi.NodeEntity, error)
	RemoveClusterNode(nId int32) error
	GetClusterNode(nId int32) (*nigoapi.NodeEntity, error)
	RemoveClusterNodeFromClusterNodeId(nId string) error

	// Registry client func
	GetRegistryClient(id string) (*nigoapi.RegistryClientEntity, error)
	CreateRegistryClient(entity nigoapi.RegistryClientEntity) (*nigoapi.RegistryClientEntity, error)
	UpdateRegistryClient(entity nigoapi.RegistryClientEntity) (*nigoapi.RegistryClientEntity, error)
	RemoveRegistryClient(entity nigoapi.RegistryClientEntity) error

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

	// Input port func
	UpdateInputPortRunStatus(id string, entity nigoapi.PortRunStatusEntity) (*nigoapi.ProcessorEntity, error)

	// Parameter context func
	GetParameterContext(id string) (*nigoapi.ParameterContextEntity, error)
	CreateParameterContext(entity nigoapi.ParameterContextEntity) (*nigoapi.ParameterContextEntity, error)
	RemoveParameterContext(entity nigoapi.ParameterContextEntity) error
	CreateParameterContextUpdateRequest(contextId string, entity nigoapi.ParameterContextEntity) (*nigoapi.ParameterContextUpdateRequestEntity, error)
	GetParameterContextUpdateRequest(contextId, id string) (*nigoapi.ParameterContextUpdateRequestEntity, error)

	Build() error
}

type nifiClient struct {
	NifiClient
	opts       *NifiConfig
	client     *nigoapi.APIClient
	nodeClient map[int32]*nigoapi.APIClient
	timeout    time.Duration
	nodes      []nigoapi.NodeDto

	// client funcs for mocking
	newClient func(*nigoapi.Configuration) *nigoapi.APIClient
}

func New(opts *NifiConfig) NifiClient {
	nClient := &nifiClient{
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
	for nodeId, _ := range n.opts.NodesURI {
		nodeConfig := n.getNiNodeGoApiConfig(nodeId)
		n.nodeClient[nodeId] = n.newClient(nodeConfig)
	}

	clusterEntity, err := n.DescribeCluster()
	if err != nil || clusterEntity == nil || clusterEntity.Cluster == nil {
		err = errorfactory.New(errorfactory.NodesUnreachable{}, err, fmt.Sprintf("could not connect to nifi nodes: %s", n.opts.NifiURI))
		return err
	}

	n.nodes = clusterEntity.Cluster.Nodes

	return nil
}

// NewFromCluster is a convenient wrapper around New() and ClusterConfig()
func NewFromCluster(k8sclient client.Client, cluster *v1alpha1.NifiCluster) (NifiClient, error) {
	var client NifiClient
	var err error

	opts, err := ClusterConfig(k8sclient, cluster)
	if err != nil {
		return nil, err
	}

	client = New(opts)
	err = client.Build()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (n *nifiClient) getNifiGoApiConfig() (config *nigoapi.Configuration) {
	config = nigoapi.NewConfiguration()

	protocol := "http"
	if n.opts.UseSSL {
		config.Scheme = "HTTPS"
		n.opts.TLSConfig.BuildNameToCertificate()
		transport := &http.Transport{TLSClientConfig: n.opts.TLSConfig}
		config.HTTPClient = &http.Client{Transport: transport}
		protocol = "https"
	}
	config.BasePath = fmt.Sprintf("%s://%s/nifi-api", protocol, n.opts.NifiURI)
	config.Host = n.opts.NifiURI

	return
}

func (n *nifiClient) getNiNodeGoApiConfig(nodeId int32) (config *nigoapi.Configuration) {
	config = nigoapi.NewConfiguration()
	config.HTTPClient = &http.Client{}
	protocol := "http"

	if n.opts.UseSSL {
		config.Scheme = "HTTPS"
		n.opts.TLSConfig.BuildNameToCertificate()
		transport := &http.Transport{TLSClientConfig: n.opts.TLSConfig}
		config.HTTPClient = &http.Client{Transport: transport}
		protocol = "https"
	}
	config.BasePath = fmt.Sprintf("%s://%s/nifi-api", protocol, n.opts.NodesURI[nodeId])
	config.Host = n.opts.NifiURI

	return
}

func (n *nifiClient) privilegeCoordinatorClient() *nigoapi.APIClient {
	if clientId := n.coordinatorNodeId(); clientId != nil {
		return n.nodeClient[*clientId]
	}

	if clientId := n.privilegeNodeClient(); clientId != nil {
		return n.nodeClient[*clientId]
	}

	return n.client
}

func (n *nifiClient) privilegeCoordinatorExceptNodeIdClient(nId int32) *nigoapi.APIClient {
	nodeDto := n.nodeDtoByNodeId(nId)
	if nodeDto == nil || isCoordinator(nodeDto) {
		if clientId := n.firstConnectedNodeId(nId); clientId != nil {
			return n.nodeClient[*clientId]
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
	return node.Status == string(v1alpha1.ConnectStatus)
}

func (n *nifiClient) nodeDtoByNodeId(nId int32) *nigoapi.NodeDto {
	for id := range n.nodes {
		nodeDto := n.nodes[id]
		// Check if the Cluster Node uri match with the one associated to the NifiCluster nodeId searched
		if fmt.Sprintf("%s:%d", nodeDto.Address, nodeDto.ApiPort) == fmt.Sprintf(n.opts.nodeURITemplate, nId) {
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
		if uri == searchedUri {
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
