package scale

import (
	"fmt"
	"time"

	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/apis/nifi/v1alpha1"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/controller/common"
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/nificlient"
	nifiutil "gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/util/nifi"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("scale-methods")

// TODO : rework upscale to check that the node is connected before ending operation.
// UpScaleCluster upscales Nifi cluster
func UpScaleCluster(nodeId, namespace, clusterName string) (v1alpha1.ActionStep, string, error) {
	actionStep := v1alpha1.ConnectNodeAction
	currentTime := time.Now()
	startTimeStamp := currentTime.Format(nifiutil.TimeStampLayout)
	return actionStep, startTimeStamp, nil
}

// DisconnectClusterNode, perform a node disconnection
func DisconnectClusterNode(client client.Client, cluster *v1alpha1.NifiCluster, nodeId string) (v1alpha1.ActionStep, string, error) {
	var err error

	// Extract nifi node Id, from nifi node address.
	int32NodeId, _ := nifiutil.ParseStringToInt32(nodeId)
	if err != nil {
		return "", "", err
	}

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return "", "", err
	}

	_, err = nClient.DisconnectClusterNode(int32NodeId)
	if err != nil && err != nificlient.ErrNifiClusterNotReturned200 {
		log.Error(err, "could not communicate with nifi node")
		return "", "", err
	}
	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error(err, "Disconnect cluster gracefully failed since Nifi node returned non 200")
		return "", "", err
	}

	log.Info("Disconnect in nifi node")
	startTimeStamp := time.Now().Format(nifiutil.TimeStampLayout)
	actionStep :=  v1alpha1.DisconnectNodeAction
	return actionStep, startTimeStamp, nil
}

// OffloadCluster, perform offload data from a node.
func OffloadClusterNode(client client.Client, cluster *v1alpha1.NifiCluster, nodeId string) (v1alpha1.ActionStep, string, error) {
	var err error

	// Extract nifi node Id, from nifi node address.
	int32NodeId, _ := nifiutil.ParseStringToInt32(nodeId)
	if err != nil {
		return "", "", err
	}

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return "", "", err
	}

	_, err = nClient.OffloadClusterNode(int32NodeId)
	if err != nil && err != nificlient.ErrNifiClusterNotReturned200 {
		log.Error(err, "could not communicate with nifi node")
		return "", "", err
	}

	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error(err, "Offload node gracefully failed since Nifi node returned non 200")
		return "", "", err
	}

	log.Info("Offload in nifi node")
	startTimeStamp := time.Now().Format(nifiutil.TimeStampLayout)
	actionStep :=  v1alpha1.OffloadNodeAction
	return actionStep, startTimeStamp, nil
}

// ConnectClusterNode, perform node connection.
func ConnectClusterNode(client client.Client, cluster *v1alpha1.NifiCluster, nodeId string) (v1alpha1.ActionStep, string, error) {
	var err error

	// Extract nifi node Id, from nifi node address.
	int32NodeId, _ := nifiutil.ParseStringToInt32(nodeId)
	if err != nil {
		return "", "", err
	}

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return "", "", err
	}

	_, err = nClient.ConnectClusterNode(int32NodeId)
	if err != nil && err != nificlient.ErrNifiClusterNotReturned200 {
		log.Error(err, "could not communicate with nifi node")
		return "", "", err
	}

	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error(err, "Connect node gracefully failed since Nifi node returned non 200")
		return "", "", err
	}

	log.Info("Connect in nifi node")
	startTimeStamp := time.Now().Format(nifiutil.TimeStampLayout)
	actionStep :=  v1alpha1.OffloadNodeAction
	return actionStep, startTimeStamp, nil
}

// RemoveClusterNode, perform node connection.
func RemoveClusterNode(client client.Client, cluster *v1alpha1.NifiCluster, nodeId string) (v1alpha1.ActionStep, string, error) {
	var err error

	// Extract NiFi node Id, from NiFi node address.
	int32NodeId, _ := nifiutil.ParseStringToInt32(nodeId)
	if err != nil {
		return "", "", err
	}

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return "", "", err
	}

	err = nClient.RemoveClusterNode(int32NodeId)
	if err == nificlient.ErrNifiClusterNodeNotFound {
		currentTime := time.Now()
		return  v1alpha1.RemoveNodeAction, currentTime.Format(nifiutil.TimeStampLayout), nil
	}
	if err != nil && err != nificlient.ErrNifiClusterNotReturned200 {
		log.Error(err, "could not communicate with nifi node")
		return "", "", err
	}

	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error(err, "Disconnect node gracefully failed since Nifi node returned non 200 or 404")
		return "", "", err
	}

	log.Info("Remove nifi node")
	startTimeStamp := time.Now().Format(nifiutil.TimeStampLayout)
	actionStep :=  v1alpha1.RemoveNodeAction
	return actionStep, startTimeStamp, nil
}

// TODO : rework to check if state is consistent (If waiting removing but disconnected ...
// CheckIfCCTaskFinished checks whether the given CC Task ID finished or not
// headlessServiceEnabled bool, availableNodes []v1alpha1.Node, serverPort int32, nodeId, namespace, clusterName string
func CheckIfNCActionStepFinished(actionStep v1alpha1.ActionStep, client client.Client, cluster *v1alpha1.NifiCluster, nodeId string) (bool, error) {
	var err error

	// Extract nifi node Id, from nifi node address.
	int32NodeId, err := nifiutil.ParseStringToInt32(nodeId)
	if err != nil {
		return false, err
	}

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return false, err
	}

	nodeEntity, err :=  nClient.GetClusterNode(int32NodeId)
	if ( err == nificlient.ErrNifiClusterNodeNotFound || err == nificlient.ErrNifiClusterReturned404) && actionStep == v1alpha1.RemoveNodeAction {
		return true, nil
	}

	if err != nil {
		return false, nil
	}

	nodeStatus := nodeEntity.Node.Status
	switch actionStep {

	case v1alpha1.DisconnectNodeAction:
		if nodeStatus == string(v1alpha1.DisconnectStatus) {
			return true, nil
		}
	case v1alpha1.OffloadNodeAction:
		if nodeStatus == string(v1alpha1.OffloadStatus) {
			return true, nil
		}
	case v1alpha1.ConnectNodeAction:
		if nodeStatus == string(v1alpha1.ConnectStatus) {
			return true, nil
		}
	case v1alpha1.RemoveNodeAction:
		if nodeStatus == string(v1alpha1.DisconnectStatus) {
			return true, nil
		}
	}
	return false, nil
}

func EnsureRemovedNodes(client client.Client, cluster *v1alpha1.NifiCluster) error {
	var err error

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return  err
	}

	clusterEntity, err := nClient.DescribeCluster()
	if err != nil {
		return  err
	}
	// GenerateNodeAddress
	stateAdresses := make(map[string]int32)

	for _, nodeId := range generateNodeStateIdSlice(cluster.Status.NodesState) {
		stateAdresses[nificlient.GenerateNodeAddress(cluster, nodeId)] =  nodeId
	}
	for _, nodeDto := range clusterEntity.Cluster.Nodes {

		if _, ok := stateAdresses[fmt.Sprintf("%s:%d", nodeDto.Address, nodeDto.ApiPort)]; !ok {

			err = nClient.RemoveClusterNodeFromClusterNodeId(nodeDto.NodeId)
			if err == nificlient.ErrNifiClusterNodeNotFound {
				return nil
			}
			if err != nil && err != nificlient.ErrNifiClusterNotReturned200 {
				log.Error(err, "could not communicate with nifi node")
				return err
			}

			if err == nificlient.ErrNifiClusterNotReturned200 {
				log.Error(err, "Disconnect node gracefully failed since Nifi node returned non 200 or 404")
				return err
			}
		}
	}

	return nil
}

func generateNodeStateIdSlice(nodesState map[string]v1alpha1.NodeState) []int32 {
	var nodeIdsSlice []int32

	for nodeId, _ := range nodesState {
		int32NodeId, _ := nifiutil.ParseStringToInt32(nodeId)
		nodeIdsSlice = append(nodeIdsSlice, int32NodeId)
	}
	return nodeIdsSlice
}