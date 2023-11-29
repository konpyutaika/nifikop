package scale

import (
	"fmt"
	"time"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

var log = common.CustomLogger().Named("scale-method")

// TODO : rework upscale to check that the node is connected before ending operation.
// UpScaleCluster upscales Nifi cluster.
func UpScaleCluster(nodeId, namespace, clusterName string) (v1.ActionStep, string, error) {
	actionStep := v1.ConnectNodeAction
	currentTime := time.Now()
	startTimeStamp := currentTime.Format(nifiutil.TimeStampLayout)
	return actionStep, startTimeStamp, nil
}

// DisconnectClusterNode, perform a node disconnection.
func DisconnectClusterNode(config *clientconfig.NifiConfig, nodeId string) (v1.ActionStep, string, error) {
	var err error

	// Extract nifi node Id, from nifi node address.
	int32NodeId, _ := nifiutil.ParseStringToInt32(nodeId)
	if err != nil {
		return "", "", err
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return "", "", err
	}

	_, err = nClient.DisconnectClusterNode(int32NodeId)
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Disconnect node gracefully"); err != nil {
		return "", "", err
	}

	log.Info("Disconnect in nifi node")
	startTimeStamp := time.Now().Format(nifiutil.TimeStampLayout)
	actionStep := v1.DisconnectNodeAction
	return actionStep, startTimeStamp, nil
}

// OffloadCluster, perform offload data from a node.
func OffloadClusterNode(config *clientconfig.NifiConfig, nodeId string) (v1.ActionStep, string, error) {
	var err error

	// Extract nifi node Id, from nifi node address.
	int32NodeId, _ := nifiutil.ParseStringToInt32(nodeId)
	if err != nil {
		return "", "", err
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return "", "", err
	}

	_, err = nClient.OffloadClusterNode(int32NodeId)
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Offload node gracefully"); err != nil {
		return "", "", err
	}

	log.Info("Offload in nifi node")
	startTimeStamp := time.Now().Format(nifiutil.TimeStampLayout)
	actionStep := v1.OffloadNodeAction
	return actionStep, startTimeStamp, nil
}

// ConnectClusterNode, perform node connection.
func ConnectClusterNode(config *clientconfig.NifiConfig, nodeId string) (v1.ActionStep, string, error) {
	var err error

	// Extract nifi node Id, from nifi node address.
	int32NodeId, _ := nifiutil.ParseStringToInt32(nodeId)
	if err != nil {
		return "", "", err
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return "", "", err
	}

	_, err = nClient.ConnectClusterNode(int32NodeId)
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Connect node gracefully"); err != nil {
		return "", "", err
	}

	log.Info("Connect in nifi node")
	startTimeStamp := time.Now().Format(nifiutil.TimeStampLayout)
	actionStep := v1.OffloadNodeAction
	return actionStep, startTimeStamp, nil
}

// RemoveClusterNode, perform node connection.
func RemoveClusterNode(config *clientconfig.NifiConfig, nodeId string) (v1.ActionStep, string, error) {
	var err error

	// Extract NiFi node Id, from NiFi node address.
	int32NodeId, _ := nifiutil.ParseStringToInt32(nodeId)
	if err != nil {
		return "", "", err
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return "", "", err
	}

	err = nClient.RemoveClusterNode(int32NodeId)
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Disconnect node gracefully"); err != nil {
		if err == nificlient.ErrNifiClusterNodeNotFound {
			currentTime := time.Now()
			return v1.RemoveNodeAction, currentTime.Format(nifiutil.TimeStampLayout), nil
		}
		return "", "", err
	}

	log.Info("Remove nifi node")
	startTimeStamp := time.Now().Format(nifiutil.TimeStampLayout)
	actionStep := v1.RemoveNodeAction
	return actionStep, startTimeStamp, nil
}

// TODO : rework to check if state is consistent (If waiting removing but disconnected ...
// CheckIfCCTaskFinished checks whether the given CC Task ID finished or not
// headlessServiceEnabled bool, availableNodes []v1.Node, serverPort int32, nodeId, namespace, clusterName string.
func CheckIfNCActionStepFinished(actionStep v1.ActionStep, config *clientconfig.NifiConfig, nodeId string) (bool, error) {
	var err error

	// Extract nifi node Id, from nifi node address.
	int32NodeId, err := nifiutil.ParseStringToInt32(nodeId)
	if err != nil {
		return false, err
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	nodeEntity, err := nClient.GetClusterNode(int32NodeId)
	if (err == nificlient.ErrNifiClusterNodeNotFound || err == nificlient.ErrNifiClusterReturned404) && actionStep == v1.RemoveNodeAction {
		return true, nil
	}

	if err != nil {
		return false, nil
	}

	nodeStatus := nodeEntity.Node.Status
	switch actionStep {
	case v1.DisconnectNodeAction:
		if nodeStatus == string(v1.DisconnectStatus) {
			return true, nil
		}
	case v1.OffloadNodeAction:
		if nodeStatus == string(v1.OffloadStatus) {
			return true, nil
		}
	case v1.ConnectNodeAction:
		if nodeStatus == string(v1.ConnectStatus) {
			return true, nil
		}
	case v1.RemoveNodeAction:
		if nodeStatus == string(v1.DisconnectStatus) {
			return true, nil
		}
	}
	return false, nil
}

func EnsureRemovedNodes(config *clientconfig.NifiConfig, cluster *v1.NifiCluster) error {
	var err error

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	clusterEntity, err := nClient.DescribeCluster()
	if err != nil {
		return err
	}
	// GenerateNodeAddress
	stateAdresses := make(map[string]int32)

	for _, nodeId := range generateNodeStateIdSlice(cluster.Status.NodesState) {
		stateAdresses[nifiutil.GenerateHostListenerNodeAddressFromCluster(nodeId, cluster)] = nodeId
	}
	for _, nodeDto := range clusterEntity.Cluster.Nodes {
		if _, ok := stateAdresses[fmt.Sprintf("%s:%d", nodeDto.Address, nodeDto.ApiPort)]; !ok {
			err = nClient.RemoveClusterNodeFromClusterNodeId(nodeDto.NodeId)
			if err := clientwrappers.ErrorRemoveOperation(log, err, "Remove node gracefully"); err != nil {
				return err
			}
		}
	}

	return nil
}

func generateNodeStateIdSlice(nodesState map[string]v1.NodeState) []int32 {
	var nodeIdsSlice []int32

	for nodeId := range nodesState {
		int32NodeId, _ := nifiutil.ParseStringToInt32(nodeId)
		nodeIdsSlice = append(nodeIdsSlice, int32NodeId)
	}
	return nodeIdsSlice
}
