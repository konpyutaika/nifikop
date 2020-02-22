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
	basePath		= "nifi-api"
	endpointCluster	= "controller/cluster"
	endpointNode	= "controller/cluster/nodes/%s"
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

//
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

