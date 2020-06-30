package nificlient

import (
	"fmt"
	"net/http"

	"emperror.dev/errors"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/apis/nifi/v1alpha1"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
)

func (n *nifiClient) DescribeCluster() (*nigoapi.ClusterEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	clusterEntry, rsp, err := client.ControllerApi.GetCluster(nil)
	if rsp != nil && rsp.StatusCode == 404 {
		log.Error(errors.New("404 response from nifi node: "+rsp.Status), "Error during talking to nifi node")
		return nil, ErrNifiClusterReturned404
	}

	if rsp != nil && rsp.StatusCode != 200 {
		log.Error(errors.New("Non 200 response from nifi node: "+rsp.Status), "Error during talking to nifi node")
		return nil, errors.New("Non 200 response from nifi node: " + rsp.Status)
	}

	if err != nil || rsp == nil  {
		log.Error(err, "Error during talking to nifi node")
		return nil, err
	}


	return &clusterEntry, nil
}

func (n *nifiClient) GetClusterNode(nId int32)(*nigoapi.NodeEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorExceptNodeIdClient(nId)
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
	nodeEntity, rsp, err := client.ControllerApi.GetNode(nil, targetedNode.NodeId)

	if rsp != nil && rsp.StatusCode != 200 {
		log.Error(errors.New("Non 200 response from nifi node: "+rsp.Status), "Error during talking to nifi node")
		return nil, ErrNifiClusterNotReturned200
	}

	if err != nil || rsp == nil {
		log.Error(err, "Error during talking to nifi node")
		return nil, err
	}

	return &nodeEntity, nil
}

func (n *nifiClient) DisconnectClusterNode(nId int32) (*nigoapi.NodeEntity, error) {
	// Request to update the node status to DISCONNECTING
	nodeEntity, err := n.setClusterNodeStatus(nId, v1alpha1.DisconnectNodeAction, v1alpha1.DisconnectStatus)

	return setClusterNodeStatusReturn(nodeEntity, err, "Disconnect cluster gracefully failed since Nifi node returned non 200")
}

func (n *nifiClient) ConnectClusterNode(nId int32) (*nigoapi.NodeEntity, error) {
	// Request to update the node status to CONNECTING
	nodeEntity, err := n.setClusterNodeStatus(nId, v1alpha1.ConnectNodeAction, v1alpha1.ConnectStatus)

	return setClusterNodeStatusReturn(nodeEntity, err, "Connect node gracefully failed since Nifi node returned non 200")
}

func (n *nifiClient) OffloadClusterNode(nId int32) (*nigoapi.NodeEntity, error) {
	// Request to update the node status to OFFLOADING
	nodeEntity, err := n.setClusterNodeStatus(nId, v1alpha1.OffloadNodeAction, v1alpha1.OffloadStatus)

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
	client := n.privilegeCoordinatorExceptNodeIdClient(nId)
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the node
	_, rsp, err := client.ControllerApi.DeleteNode(nil, targetedNode.NodeId)
	return removeClusterNodeStatusReturn(rsp, err)
}

func (n *nifiClient) RemoveClusterNodeFromClusterNodeId(nId string) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorClient()
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the node
	_, rsp, err := client.ControllerApi.DeleteNode(nil, nId)
	return removeClusterNodeStatusReturn(rsp, err)
}


func (n *nifiClient) setClusterNodeStatus(nId int32, status, expectedActionStatus v1alpha1.ActionStep) (*nigoapi.NodeEntity, error) {
	// Find the Cluster node associated to the NifiCluster nodeId
	targetedNode := n.nodeDtoByNodeId(nId)
	if targetedNode == nil {
		log.Error(ErrNifiClusterNodeNotFound, "Error during preparing the request")
		return nil, ErrNifiClusterNodeNotFound
	}

	// Check if the targeted node is still in expected status
	// TODO : ensure it may not leads to inconsistent situations
	if targetedNode.Status == string(expectedActionStatus) ||
		targetedNode.Status == string(status) {

		node := nigoapi.NodeEntity{Node: targetedNode}
		return &node, nil
	}

	// Get nigoapi client, favoring the one associated to the coordinator node.
	client := n.privilegeCoordinatorExceptNodeIdClient(nId)
	if client == nil {
		log.Error(ErrNoNodeClientsAvailable, "Error during creating node client")
		return nil, ErrNoNodeClientsAvailable
	}

	// Update node status to the target one
	targetedNode.Status = string(status)

	// Request on Nifi Rest API to update the node status
	nodeEntity, rsp, err := client.ControllerApi.UpdateNode(nil, targetedNode.NodeId, nigoapi.NodeEntity{Node: targetedNode})

	if rsp != nil && rsp.StatusCode != 200 && rsp.StatusCode != 202 {
		log.Error(err, fmt.Sprintf("%s cluster gracefully failed since Nifi node returned non 200", string(status)))
		return nil, ErrNifiClusterNotReturned200
	}

	if err != nil || rsp == nil {
		log.Error(err, "Could not communicate with nifi node")
		return nil , err
	}

	n.setNodeFromNodes(nodeEntity.Node)
	return &nodeEntity, nil
}

func setClusterNodeStatusReturn(nodeEntity *nigoapi.NodeEntity, err error, messageError string) (*nigoapi.NodeEntity, error) {
	if err != nil && err != ErrNifiClusterNotReturned200 {
		log.Error(err, messageError)
		return nil , err
	}

	if err == ErrNifiClusterNotReturned200 {
		log.Error(err, "Could not communicate with nifi node")
		return nil, err
	}

	return nodeEntity, nil
}

func removeClusterNodeStatusReturn( rsp *http.Response, err error) error {

	if rsp != nil && rsp.StatusCode == 404 {
		log.Error(errors.New("404 response from nifi node: "+rsp.Status), "No node to remove found")
		return nil
	}

	if rsp != nil && rsp.StatusCode != 200 {
		log.Error(errors.New("Non 200 response from nifi node: "+rsp.Status), "Error during talking to nifi node")
		return ErrNifiClusterNotReturned200
	}

	if err != nil || rsp == nil {
		log.Error(err, "Error during talking to nifi node")
		return err
	}

	return nil
}