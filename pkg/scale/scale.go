// Copyright © 2019 Banzai Cloud
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
// limitations under the License.

package scale

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/openshift/origin/Godeps/_workspace/src/k8s.io/kubernetes/pkg/util/json"
	"github.com/orangeopensource/nifi-operator/pkg/apis/nifi/v1alpha1"
	nifiutil "github.com/orangeopensource/nifi-operator/pkg/util/nifi"
	"io/ioutil"
	"net/http"
	"strconv"

	//	"strconv"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	basePath				= "nifi-api"

	endpointCluster	= "controller/cluster"
	endpointNode	= "controller/cluster/nodes/%s"


	removeNodeAction       		= "remove_node"
	stateAction 				= "state"
	addNodeAction         		= "add_node"
	getTaskListAction        	= "user_tasks"
	nifiClusterStateAction  	= "nifi_cluster_state"
	clusterLoad              	= "load"
	rebalanceAction         	= "rebalance"
	killProposalAction       	= "stop_proposal_execution"
	serviceNameTemplate      	= "%s-cruisecontrol-svc"
	nodeAlive              		= "ALIVE"
)

var errNodeNotConnected = errors.New("The targeted node id disconnected")
var errNifiClusterNotReturned200 = errors.New("non 200 response from nifi cluster")
var errNifiClusterReturned404 = errors.New("404 response from nifi cluster")

var log = logf.Log.WithName("cruise-control-methods")

func generateUrlForNN(headlessServiceEnabled bool, nodeId , serverPort int32, endpoint, namespace string, clusterName string) string {
	var baseUrl string
	baseUrl = nifiutil.ComputeHostname(headlessServiceEnabled, nodeId, clusterName, namespace)
	return "http://" + fmt.Sprintf("%s:%d/%s/%s", baseUrl, serverPort, basePath, endpoint)
}

func putNifiNode(headlessServiceEnabled bool, nodeId, serverPort int32, endpoint, namespace, clusterName, action, nifiNodeId string) (*http.Response, error) {

	requestURl := generateUrlForNN(headlessServiceEnabled, nodeId, serverPort, endpoint, namespace, clusterName)

	var bodyStr = []byte(fmt.Sprintf(`{"node":{"nodeId": "%s", "status": "%s"}}`, nifiNodeId, action))

	req, err := http.NewRequest(http.MethodPut, requestURl, bytes.NewBuffer(bodyStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		log.Error(err, "error during talking to nifi node")
		return nil, err
	}
	if rsp.StatusCode != 200 && rsp.StatusCode != 202 {
		log.Error(errors.New("Non 200 response from nifi node: "+rsp.Status), "error during talking to nifi node")
		return nil, errNifiClusterNotReturned200
	}

	return rsp, nil
}

func getNifiNode(headlessServiceEnabled bool, nodeId, serverPort int32, endpoint, namespace, clusterName string) (*http.Response, error) {

	requestURl := generateUrlForNN(headlessServiceEnabled, nodeId, serverPort, endpoint, namespace, clusterName)
	rsp, err := http.Get(requestURl)
	if err != nil {
		log.Error(err, "error during talking to nifi node")
		return nil, err
	}
	if rsp.StatusCode == 404 {
		log.Error(errors.New("404 response from nifi node: "+rsp.Status), "error during talking to nifi node")
		return rsp, errNifiClusterReturned404
	}

	if rsp.StatusCode != 200 {
		log.Error(errors.New("Non 200 response from nifi node: "+rsp.Status), "error during talking to nifi node")
		return nil, errors.New("Non 200 response from nifi node: " + rsp.Status)
	}

	return rsp, nil
}

func deleteNifiNode(headlessServiceEnabled bool, nodeId, serverPort int32, endpoint, namespace, clusterName string) (*http.Response, error) {

	requestURl := generateUrlForNN(headlessServiceEnabled, nodeId, serverPort, endpoint, namespace, clusterName)
	req, err := http.NewRequest(http.MethodDelete, requestURl, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		log.Error(err, "error during talking to nifi node")
		return nil, err
	}
	if rsp.StatusCode != 200 {
		log.Error(errors.New("Non 200 response from nifi node: "+rsp.Status), "error during talking to nifi node")
		return nil, errors.New("Non 200 response from nifi node: " + rsp.Status)
	}

	return rsp, nil
}

func GetNifiClsuterNodesStatus(headlessServiceEnabled bool, nodeId, serverPort int32, namespace, clusterName string) (map[string]interface{}, error) {

	rsp, err := getNifiNode(headlessServiceEnabled, nodeId, serverPort, endpointCluster, namespace, clusterName)

	if err != nil {
		log.Error(err, "can't work with nifi node because it is not ready")
		return nil, err
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	err = rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func GetNifiClusterNodeStatus(headlessServiceEnabled bool, nodeId, serverPort int32, namespace, clusterName, targetNodeId string) (map[string]interface{}, error) {

	rsp, err := getNifiNode(headlessServiceEnabled, nodeId, serverPort, fmt.Sprintf(endpointNode, targetNodeId), namespace, clusterName)

	if err != nil {
		log.Error(err, "can't work with nifi node because it is not ready")
		return nil, err
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	err = rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func isNifiNodeReady(nodeId, namespace, clusterName string) (bool, error) {

	running := false

	/*options := map[string]string{
		"json": "true",
	}

	rsp, err := getCruiseControl(clusterLoad, namespace, options, clusterName)
	if err != nil {
		log.Error(err, "can't work with cruise-control because it is not ready")
		return running, err
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return running, err
	}

	err = rsp.Body.Close()
	if err != nil {
		return running, err
	}

	var response map[string]interface{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return running, err
	}

	bIdToFloat, _ := strconv.ParseFloat(brokerId, 32)

	for _, broker := range response["brokers"].([]interface{}) {
		if broker.(map[string]interface{})["Broker"].(float64) == bIdToFloat &&
			broker.(map[string]interface{})["BrokerState"].(string) == brokerAlive {
			log.Info("broker is available in cruise-control")
			running = true
			break
		}
	}*/
	// TODO: to remove after implementation
	running = true
	return running, nil
}

// GetBrokerIDWithLeastPartition returns
/*func GetBrokerIDWithLeastPartition(namespace, clusterName string) (string, error) {


	brokerWithLeastPartition := ""

	err := GetCruiseControlStatus(namespace, clusterName)
	if err != nil {
		return brokerWithLeastPartition, err
	}

	options := map[string]string{
		"json": "true",
	}

	rsp, err := getCruiseControl(kafkaClusterStateAction, namespace, options, clusterName)
	if err != nil {
		log.Error(err, "can't work with cruise-control because it is not ready")
		return brokerWithLeastPartition, err
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return brokerWithLeastPartition, err
	}

	err = rsp.Body.Close()
	if err != nil {
		return brokerWithLeastPartition, err
	}

	var response map[string]interface{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return brokerWithLeastPartition, err
	}

	replicaCountByBroker := response["KafkaBrokerState"].(map[string]interface{})["ReplicaCountByBrokerId"].(map[string]interface{})
	replicaCount := float64(99999)

	for brokerID, replica := range replicaCountByBroker {
		if replicaCount > replica.(float64) {
			replicaCount = replica.(float64)
			brokerWithLeastPartition = brokerID
		}
	}
	return brokerWithLeastPartition, nil

}*/


// UpScaleCluster upscales Kafka cluster
func UpScaleCluster(nodeId, namespace, clusterName string) (v1alpha1.ActionStep, string, error) {
	actionStep := v1alpha1.ConnectNodeAction
	currentTime := time.Now()
	startTimeStamp := currentTime.Format("Mon, 2 Jan 2006 15:04:05 GMT")
	return actionStep, startTimeStamp, nil
}

func getNifiNodeIdFromAddress(headlessServiceEnabled bool, nodeId, serverPort int32, namespace, clusterName, searchedAddress string) (string, error) {
	var clusterStatus map[string]interface{}
	var err error

	clusterStatus, err = GetNifiClsuterNodesStatus(headlessServiceEnabled, nodeId, serverPort, namespace, clusterName)
	if err != nil {
		return "", err
	}

	var targetNodeId string

	for _, node := range clusterStatus["cluster"].(map[string]interface{})["nodes"].([]interface{}){
		address := node.(map[string]interface{})["address"].(string)
		if address == searchedAddress {
			targetNodeId = node.(map[string]interface{})["nodeId"].(string)
		}
	}

	return targetNodeId, nil
}

func getAvailableNifiClusterNode(headlessServiceEnabled bool, availableNodes []v1alpha1.Node, serverPort int32, namespace, clusterName string) (v1alpha1.Node, error) {
	var err error
	for _, n := range availableNodes {
		_, err = GetNifiClsuterNodesStatus(headlessServiceEnabled, n.Id, serverPort, namespace, clusterName)
		if err == nil {
			return n, nil
		}
	}
	return v1alpha1.Node{}, err
}

// DownsizeCluster downscales Nifi cluster
func DisconnectClusterNode(headlessServiceEnabled bool, availableNodes []v1alpha1.Node, serverPort int32, nodeId, namespace, clusterName string) (v1alpha1.ActionStep, string, error) {
	var node v1alpha1.Node
	var err error
	var rsp map[string]interface{}

	// Look for available nifi node.
	node, err = getAvailableNifiClusterNode(headlessServiceEnabled, availableNodes, serverPort, namespace, clusterName)

	if &node == nil {
		return "", "", err
	}

	var dResp *http.Response

	// Extract nifi node Id, from nifi node address.
	intNodeId, err := strconv.ParseInt(nodeId,10, 32)
	int32NodeId := int32(intNodeId)
	searchedAddress := nifiutil.ComputeHostname(headlessServiceEnabled, int32NodeId, clusterName, namespace)

	targetNodeId, err := getNifiNodeIdFromAddress(headlessServiceEnabled, node.Id, serverPort, namespace, clusterName, searchedAddress)
	if err != nil {
		return "", "", err
	}

	rsp, err = GetNifiClusterNodeStatus(headlessServiceEnabled, node.Id, serverPort, namespace, clusterName, targetNodeId)

	if err != nil {
		return "", "", err
	}
	if rsp["node"].(map[string]interface{})["status"].(string) != string(v1alpha1.ConnectStatus) {
		return "", "", errNodeNotConnected
	}

	// Disconnect node
	dResp, err = putNifiNode(headlessServiceEnabled, node.Id, serverPort, fmt.Sprintf(endpointNode, targetNodeId), namespace, clusterName, string(v1alpha1.DisconnectNodeAction), targetNodeId)
	if err != nil && err != errNifiClusterNotReturned200 {
		log.Error(err, "Disconnect cluster gracefully failed since Nifi node returned non 200")
		return "", "", err
	}
	if err == errNifiClusterNotReturned200 {
		log.Error(err, "could not communicate with nifi node")
		return "", "", err
	}

	log.Info("Disconnect in nifi node")
	startTimeStamp := dResp.Header.Get("Date")
	actionStep :=  v1alpha1.DisconnectNodeAction
	return actionStep, startTimeStamp, nil
}

func OffloadClusterNode(headlessServiceEnabled bool, availableNodes []v1alpha1.Node, serverPort int32, nodeId, namespace, clusterName string) (v1alpha1.ActionStep, string, error) {
	var node v1alpha1.Node
	var err error

	// Look for available nifi node.
	node, err = getAvailableNifiClusterNode(headlessServiceEnabled, availableNodes, serverPort, namespace, clusterName)

	if &node == nil {
		return "", "", err
	}

	var dResp *http.Response

	// Extract nifi node Id, from nifi node address.
	intNodeId, err := strconv.ParseInt(nodeId,10, 32)
	int32NodeId := int32(intNodeId)
	searchedAddress := nifiutil.ComputeHostname(headlessServiceEnabled, int32NodeId, clusterName, namespace)

	targetNodeId, err := getNifiNodeIdFromAddress(headlessServiceEnabled, node.Id, serverPort, namespace, clusterName, searchedAddress)
	if err != nil {
		return "", "", err
	}

	// Offload node
	dResp, err = putNifiNode(headlessServiceEnabled, node.Id, serverPort, fmt.Sprintf(endpointNode, targetNodeId), namespace, clusterName, string(v1alpha1.OffloadNodeAction), targetNodeId)
	if err != nil && err != errNifiClusterNotReturned200 {
		log.Error(err, "Offload node gracefully failed since Nifi node returned non 200")
		return "", "", err
	}
	if err == errNifiClusterNotReturned200 {
		log.Error(err, "could not communicate with nifi node")
		return "", "", err
	}

	log.Info("Offload in nifi node")
	startTimeStamp := dResp.Header.Get("Date")
	actionStep :=  v1alpha1.OffloadNodeAction
	return actionStep, startTimeStamp, nil
}

func RemoveClusterNode(headlessServiceEnabled bool, availableNodes []v1alpha1.Node, serverPort int32, nodeId, namespace, clusterName string) (v1alpha1.ActionStep, string, error) {
	var node v1alpha1.Node
	var err error

	// Look for available nifi node.
	node, err = getAvailableNifiClusterNode(headlessServiceEnabled, availableNodes, serverPort, namespace, clusterName)

	if &node == nil {
		return "", "", err
	}

	var dResp *http.Response


	// Extract nifi node Id, from nifi node address.
	intNodeId, err := strconv.ParseInt(nodeId,10, 32)
	int32NodeId := int32(intNodeId)
	searchedAddress := nifiutil.ComputeHostname(headlessServiceEnabled, int32NodeId, clusterName, namespace)

	targetNodeId, err := getNifiNodeIdFromAddress(headlessServiceEnabled, node.Id, serverPort, namespace, clusterName, searchedAddress)
	if err != nil {
		return "", "", err
	}

	// Remove node
	dResp, err = deleteNifiNode(headlessServiceEnabled, node.Id, serverPort, fmt.Sprintf(endpointNode, targetNodeId), namespace, clusterName)
	if err != nil && err != errNifiClusterNotReturned200 {
		log.Error(err, "Remove node gracefully failed since Nifi node returned non 200")
		return "", "", err
	}
	if err == errNifiClusterNotReturned200 {
		log.Error(err, "could not communicate with nifi node")
		return "", "", err
	}

	log.Info("Remove in nifi node")
	startTimeStamp := dResp.Header.Get("Date")
	actionStep :=  v1alpha1.RemoveNodeAction
	return actionStep, startTimeStamp, nil
}

// RebalanceCluster rebalances Kafka cluster using CC
func RebalanceCluster(namespace, ccEndpoint, clusterName string) (string, error) {
	/*
	err := GetCruiseControlStatus(namespace, clusterName)
	if err != nil {
		return "", err
	}

	options := map[string]string{
		"dryrun": "false",
		"json":   "true",
	}

	dResp, err := postCruiseControl(rebalanceAction, namespace, options, clusterName)
	if err != nil {
		log.Error(err, "can't rebalance cluster gracefully since post to cruise-control failed")
		return "", err
	}
	log.Info("Initiated rebalance in cruise control")

	uTaskId := dResp.Header.Get("User-Task-Id")*/

	// TODO: to remove after implementation
	uTaskId := "mock"

	return uTaskId, nil
}

// RunPreferedLeaderElectionInCluster runs leader election in  Kafka cluster using CC
func RunPreferedLeaderElectionInCluster(namespace, clusterName string) (string, error) {

	/*err := GetCruiseControlStatus(namespace, clusterName)
	if err != nil {
		return "", err
	}

	options := map[string]string{
		"dryrun": "false",
		"json":   "true",
		"goals":  "PreferredLeaderElectionGoal",
	}

	dResp, err := postCruiseControl(rebalanceAction, namespace, options, clusterName)
	if err != nil {
		log.Error(err, "can't rebalance cluster gracefully since post to cruise-control failed")
		return "", err
	}
	log.Info("Initiated rebalance in cruise control")

	uTaskId := dResp.Header.Get("User-Task-Id")*/

	// TODO: to remove after implementation
	uTaskId := "mock"

	return uTaskId, nil
}
func ConnectClusterNode(headlessServiceEnabled bool, availableNodes []v1alpha1.Node, serverPort int32, nodeId, namespace, clusterName string) (v1alpha1.ActionStep, string, error) {
	var node v1alpha1.Node
	var err error

	// Look for available nifi node.
	node, err = getAvailableNifiClusterNode(headlessServiceEnabled, availableNodes, serverPort, namespace, clusterName)

	if &node == nil {
		return "", "", err
	}

	var dResp *http.Response

	// Extract nifi node Id, from nifi node address.
	intNodeId, err := strconv.ParseInt(nodeId,10, 32)
	int32NodeId := int32(intNodeId)
	searchedAddress := nifiutil.ComputeHostname(headlessServiceEnabled, int32NodeId, clusterName, namespace)

	targetNodeId, err := getNifiNodeIdFromAddress(headlessServiceEnabled, node.Id, serverPort, namespace, clusterName, searchedAddress)
	if err != nil {
		return "", "", err
	}

	// Connect node
	dResp, err = putNifiNode(headlessServiceEnabled, node.Id, serverPort, fmt.Sprintf(endpointNode, targetNodeId), namespace, clusterName, string(v1alpha1.ConnectNodeAction), targetNodeId)
	if err != nil && err != errNifiClusterNotReturned200 {
		log.Error(err, "Connect node gracefully failed since Nifi node returned non 200")
		return "", "", err
	}
	if err == errNifiClusterNotReturned200 {
		log.Error(err, "could not communicate with nifi node")
		return "", "", err
	}

	log.Info("Connect in nifi node")
	startTimeStamp := dResp.Header.Get("Date")
	actionStep :=  v1alpha1.ConnectNodeAction
	return actionStep, startTimeStamp, nil
}

// KillNCTask kills the specified CC task
func KillNCTask(namespace, clusterName string) error {
	/*err := GetCruiseControlStatus(namespace, clusterName)
	if err != nil {
		return err
	}
	options := map[string]string{
		"json": "true",
	}

	_, err = postCruiseControl(killProposalAction, namespace, options, clusterName)
	if err != nil {
		log.Error(err, "can't kill running tasks since post to cruise-control failed")
		return err
	}
	log.Info("Task killed")*/

	return nil
}

// CheckIfCCTaskFinished checks whether the given CC Task ID finished or not
// headlessServiceEnabled bool, availableNodes []v1alpha1.Node, serverPort int32, nodeId, namespace, clusterName string
func CheckIfNCTaskFinished(headlessServiceEnabled bool, availableNodes []v1alpha1.Node, serverPort int32, actionStep v1alpha1.ActionStep,nodeId, namespace, clusterName string) (bool, error) {

	/*gResp, err := getCruiseControl(getTaskListAction, namespace, map[string]string{
		"json":          "true",
		"user_task_ids": uTaskId,
	}, clusterName)
	if err != nil {
		log.Error(err, "can't get task list from cruise-control")
		return false, err
	}

	var taskLists map[string]interface{}

	body, err := ioutil.ReadAll(gResp.Body)
	if err != nil {
		return false, err
	}

	err = gResp.Body.Close()
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(body, &taskLists)
	if err != nil {
		return false, err
	}
	// TODO use struct instead of casting things
	for _, task := range taskLists["userTasks"].([]interface{}) {
		if task.(map[string]interface{})["Status"].(string) != "Completed" {
			log.Info("Cruise control task  still running", "taskID", uTaskId)
			return false, nil
		}
	}
	log.Info("Cruise control task finished", "taskID", uTaskId)*/
	return true, nil
}

// CheckIfCCTaskFinished checks whether the given CC Task ID finished or not
// headlessServiceEnabled bool, availableNodes []v1alpha1.Node, serverPort int32, nodeId, namespace, clusterName string
func CheckIfNCActionStepFinished(headlessServiceEnabled bool, availableNodes []v1alpha1.Node, serverPort int32, actionStep v1alpha1.ActionStep, nodeId, namespace, clusterName string) (bool, error) {
	var node v1alpha1.Node
	var err error

	// Look for available nifi node.
	node, err = getAvailableNifiClusterNode(headlessServiceEnabled, availableNodes, serverPort, namespace, clusterName)

	if &node == nil {
		return false, err
	}

	var dResp map[string]interface{}

	// Extract nifi node Id, from nifi node address.
	intNodeId, err := strconv.ParseInt(nodeId,10, 32)
	int32NodeId := int32(intNodeId)
	searchedAddress := nifiutil.ComputeHostname(headlessServiceEnabled, int32NodeId, clusterName, namespace)

	targetNodeId, err := getNifiNodeIdFromAddress(headlessServiceEnabled, node.Id, serverPort, namespace, clusterName, searchedAddress)
	if err != nil {
		return false, err
	}

	dResp, err = GetNifiClusterNodeStatus(headlessServiceEnabled, node.Id, serverPort, namespace, clusterName, targetNodeId)

	if err == errNifiClusterReturned404 && actionStep == v1alpha1.RemoveNodeAction {
		return true, nil
	}

	if err != nil {
		return false, nil
	}



	currentStatus := dResp["node"].(map[string]interface{})["status"].(string)
	switch actionStep {

		case v1alpha1.DisconnectNodeAction:
			if currentStatus == string(v1alpha1.DisconnectStatus) {
				return true, nil
			}
		case v1alpha1.OffloadNodeAction:
			if currentStatus == string(v1alpha1.OffloadStatus) {
				return true, nil
			}
		case v1alpha1.ConnectNodeAction:
			if currentStatus == string(v1alpha1.ConnectStatus) {
				return true, nil
			}
	}
	return false, nil
}

