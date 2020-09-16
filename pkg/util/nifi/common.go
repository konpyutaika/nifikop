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

package nifi

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
)

const (
	PrefixNodeNameTemplate = "%s-"
	SuffixNodeNameTemplate = "-node"
	RootNodeNameTemplate   = "%d"
	NodeNameTemplate       = PrefixNodeNameTemplate + RootNodeNameTemplate + SuffixNodeNameTemplate
	// AllNodeServiceTemplate template for Nifi all nodes service
	AllNodeServiceTemplate = "%s-all-node"
	// HeadlessServiceTemplate template for Nifi headless service
	HeadlessServiceTemplate = "%s-headless"

	// TimeStampLayout defines the date format used.
	TimeStampLayout = "Mon, 2 Jan 2006 15:04:05 GMT"
)

// ParseTimeStampToUnixTime parses the given CC timeStamp to time format
func ParseTimeStampToUnixTime(timestamp string) (time.Time, error) {

	t, err := time.Parse(TimeStampLayout, timestamp)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func ParseStringToInt32(nodeId string) (int32, error) {
	intNodeId, err := strconv.ParseInt(nodeId, 10, 32)
	int32NodeId := int32(intNodeId)

	return int32NodeId, err
}

func GenerateNiFiAddressFromCluster(cluster *v1alpha1.NifiCluster) string {

	return ComputeNiFiAddress(
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.ListenersConfig.InternalListeners,
	)
}

func ComputeNiFiAddress(
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool,
	internalListeners []v1alpha1.InternalListenerConfig) string {

	return fmt.Sprintf("%s:%d",
		ComputeNiFiHostname(
			clusterName,
			namespace,
			headlessServiceEnabled,
			clusterDomain,
			useExternalDNS),
		InternalListenerForComm(internalListeners).ContainerPort)
}

func GenerateNodeAddressFromCluster(nodeId int32, cluster *v1alpha1.NifiCluster) string {

	return ComputeNodeAddress(
		nodeId,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.ListenersConfig.InternalListeners,
	)
}

func ComputeNodeAddress(
	nodeId int32,
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool,
	internalListeners []v1alpha1.InternalListenerConfig) string {

	return fmt.Sprintf("%s:%d",
		ComputeNodeHostname(
			nodeId,
			clusterName,
			namespace,
			headlessServiceEnabled,
			clusterDomain,
			useExternalDNS),
		InternalListenerForComm(internalListeners).ContainerPort)
}

func ComputeNodeHostnameFromCluster(nodeId int32, cluster *v1alpha1.NifiCluster) string {
	return ComputeNodeHostname(
		nodeId,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
	)
}

func ComputeNodeHostname(
	nodeId int32,
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool) string {

	return fmt.Sprintf("%s.%s",
		ComputeNodeName(nodeId, clusterName),
		ComputeNiFiHostname(clusterName, namespace, headlessServiceEnabled, clusterDomain, useExternalDNS))
}

func ComputeNiFiHostname(
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool) string {

	return fmt.Sprintf("%s.%s",
		ComputeServiceNameFull(clusterName, namespace, headlessServiceEnabled, useExternalDNS),
		clusterDomain)
}

// ComputeNodeServiceNameFull return the node service name in svc format
func ComputeNodeServiceNameFull(nodeId int32, clusterName, namespace string, headlessServiceEnabled, useExternalDNS bool) string {
	return fmt.Sprintf("%s.%s",
		ComputeNodeName(nodeId, clusterName),
		ComputeServiceNameFull(clusterName, namespace, headlessServiceEnabled, useExternalDNS))
}

// ComputeNodeServiceNameNs return the node service name in namespace format
func ComputeNodeServiceNameNs(nodeId int32, clusterName, namespace string, headlessServiceEnabled, useExternalDNS bool) string {
	return fmt.Sprintf("%s.%s",
		ComputeNodeName(nodeId, clusterName),
		ComputeServiceNameWithNamespace(clusterName, namespace, headlessServiceEnabled, useExternalDNS))
}

func ComputeServiceNameFull(clusterName, namespace string, headlessServiceEnabled, useExternalDNS bool) string {
	nsService := ComputeServiceNameWithNamespace(clusterName, namespace, headlessServiceEnabled, useExternalDNS)
	if useExternalDNS {
		return nsService
	}

	return fmt.Sprintf("%s.svc", nsService)
}

// ComputeServiceNameWithNamespace return the service name with namespace with the right format
func ComputeServiceNameWithNamespace(clusterName, namespace string, headlessServiceEnabled, useExternalDNS bool) string {
	serviceName := ComputeServiceName(clusterName, headlessServiceEnabled)
	if useExternalDNS {
		return serviceName
	}
	return fmt.Sprintf("%s.%s", serviceName, namespace)
}

func ComputeNodeServiceName(nodeId int32, clusterName string, headlessServiceEnabled bool) string {
	return fmt.Sprintf("%s.%s", ComputeNodeName(nodeId, clusterName),
		ComputeServiceName(clusterName, headlessServiceEnabled))
}

// ComputeServiceName return the service name with the right format
func ComputeServiceName(clusterName string, headlessServiceEnabled bool) string {
	if headlessServiceEnabled {
		return fmt.Sprintf(HeadlessServiceTemplate, clusterName)
	}

	return fmt.Sprintf(AllNodeServiceTemplate, clusterName)
}

// ComputeNodeName return the node name with right format.
func ComputeNodeName(nodeId int32, clusterName string) string {
	return fmt.Sprintf(NodeNameTemplate, clusterName, nodeId)
}

func InternalListenerForComm(internalListeners []v1alpha1.InternalListenerConfig) v1alpha1.InternalListenerConfig {
	return internalListeners[determineInternalListenerForComm(internalListeners)]
}

func determineInternalListenerForComm(internalListeners []v1alpha1.InternalListenerConfig) int {
	var httpsServerPortId int
	var httpServerPortId int
	for id, iListener := range internalListeners {
		if iListener.Type == v1alpha1.HttpsListenerType {
			httpsServerPortId = id
		} else if iListener.Type == v1alpha1.HttpListenerType {
			httpServerPortId = id
		}
	}
	if &httpsServerPortId != nil {
		return httpsServerPortId
	}
	return httpServerPortId
}
