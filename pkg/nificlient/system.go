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
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
)

func (n *nifiClient) DescribeCluster() (*nigoapi.ClusterEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	clusterEntry, rsp, body, err := client.ControllerApi.GetCluster(context)
	if err := errorGetOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &clusterEntry, nil
}

func (n *nifiClient) DescribeClusterFromNodeId(nodeId int32) (*nigoapi.ClusterEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.nodeClient[nodeId]
	context := n.opts.NodesContext[nodeId]
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	clusterEntry, rsp, body, err := client.ControllerApi.GetCluster(context)
	if err := errorGetOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &clusterEntry, nil
}

func (n *nifiClient) GetClusterNode(nId int32) (*nigoapi.NodeEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorExceptNodeIdClient(nId)
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Find the Cluster node associated to the NifiCluster nodeId
	targetedNode := n.nodeDtoByNodeId(nId)
	if targetedNode == nil {
		log.Error(ErrNifiClusterNodeNotFound, "Error during preparing the request")
		return nil, ErrNifiClusterNodeNotFound
	}

	// Request on Nifi Rest API to get the node information
	nodeEntity, rsp, body, err := client.ControllerApi.GetNode(context, targetedNode.NodeId)

	if err := errorGetOperation(rsp, body, err); err != nil {
		return nil, err
	}

	return &nodeEntity, nil
}

func (n *nifiClient) DisconnectClusterNode(nId int32) (*nigoapi.NodeEntity, error) {
	// Request to update the node status to DISCONNECTING
	nodeEntity, err := n.setClusterNodeStatus(nId, DISCONNECTING_STATUS, DISCONNECTED_STATUS)

	return setClusterNodeStatusReturn(nodeEntity, err, "Disconnect cluster gracefully failed since Nifi node returned non 200")
}

func (n *nifiClient) ConnectClusterNode(nId int32) (*nigoapi.NodeEntity, error) {
	// Request to update the node status to CONNECTING
	nodeEntity, err := n.setClusterNodeStatus(nId, CONNECTING_STATUS, CONNECTED_STATUS)

	return setClusterNodeStatusReturn(nodeEntity, err, "Connect node gracefully failed since Nifi node returned non 200")
}

func (n *nifiClient) OffloadClusterNode(nId int32) (*nigoapi.NodeEntity, error) {
	// Request to update the node status to OFFLOADING
	nodeEntity, err := n.setClusterNodeStatus(nId, OFFLOADING_STATUS, OFFLOADED_STATUS)

	return setClusterNodeStatusReturn(nodeEntity, err, "Offload node gracefully failed since Nifi node returned non 200")
}

func (n *nifiClient) RemoveClusterNode(nId int32) error {
	// Find the Cluster node associated to the NifiCluster nodeId
	targetedNode := n.nodeDtoByNodeId(nId)
	if targetedNode == nil {
		log.Error(ErrNifiClusterNodeNotFound, "Error during preparing the request")
		return ErrNifiClusterNodeNotFound
	}

	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorExceptNodeIdClient(nId)
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the node
	_, rsp, body, err := client.ControllerApi.DeleteNode(context, targetedNode.NodeId)

	return errorDeleteOperation(rsp, body, err)
}

func (n *nifiClient) RemoveClusterNodeFromClusterNodeId(nId string) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the node
	_, rsp, body, err := client.ControllerApi.DeleteNode(context, nId)

	return errorDeleteOperation(rsp, body, err)
}

func (n *nifiClient) setClusterNodeStatus(nId int32, status, expectedActionStatus string) (*nigoapi.NodeEntity, error) {
	// Find the Cluster node associated to the NifiCluster nodeId
	targetedNode := n.nodeDtoByNodeId(nId)
	if targetedNode == nil {
		log.Error(ErrNifiClusterNodeNotFound, "Error during preparing the request")
		return nil, ErrNifiClusterNodeNotFound
	}

	// Check if the targeted node is still in expected status
	// TODO : ensure it may not leads to inconsistent situations
	if targetedNode.Status == expectedActionStatus ||
		targetedNode.Status == status {

		node := nigoapi.NodeEntity{Node: targetedNode}
		return &node, nil
	}

	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorExceptNodeIdClient(nId)
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Update node status to the target one
	targetedNode.Status = string(status)

	// Request on Nifi Rest API to update the node status
	nodeEntity, rsp, body, err := client.ControllerApi.UpdateNode(context, targetedNode.NodeId, nigoapi.NodeEntity{Node: targetedNode})
	if err := errorUpdateOperation(rsp, body, err); err != nil {
		return nil, err
	}

	n.setNodeFromNodes(nodeEntity.Node)
	return &nodeEntity, nil
}

func setClusterNodeStatusReturn(nodeEntity *nigoapi.NodeEntity, err error, messageError string) (*nigoapi.NodeEntity, error) {
	if err != nil && err != ErrNifiClusterNotReturned200 {
		log.Error(err, messageError+"error since Nifi node returned non 200")
		return nil, err
	}

	if err == ErrNifiClusterNotReturned200 {
		log.Error(err, "Could not communicate with nifi node")
		return nil, err
	}

	return nodeEntity, nil
}
