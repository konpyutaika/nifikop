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

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
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

// > RequestNiFI
// >> Node
func ComputeNodeName(nodeId int32, clusterName string) string {
	return fmt.Sprintf(NodeNameTemplate, clusterName, nodeId)
}

func ComputeRequestNiFiNodeService(nodeId int32, clusterName string, headlessServiceEnabled bool) string {
	if headlessServiceEnabled {
		return fmt.Sprintf("%s.%s",
			ComputeNodeName(nodeId, clusterName),
			fmt.Sprintf(HeadlessServiceTemplate, clusterName))
	}

	return ComputeNodeName(nodeId, clusterName)
}

func ComputeRequestNiFiNodeNamespace(nodeId int32, clusterName, namespace string, headlessServiceEnabled, useExternalDNS bool) string {
	if useExternalDNS {
		return ComputeRequestNiFiNodeService(nodeId, clusterName, headlessServiceEnabled)
	}
	return fmt.Sprintf("%s.%s",
		ComputeRequestNiFiNodeService(nodeId, clusterName, headlessServiceEnabled), namespace)
}

func ComputeRequestNiFiNodeNamespaceFull(
	nodeId int32,
	clusterName, namespace string,
	headlessServiceEnabled, useExternalDNS bool) string {

	if useExternalDNS {
		return ComputeRequestNiFiNodeNamespace(nodeId, clusterName, namespace, headlessServiceEnabled, useExternalDNS)
	}
	return fmt.Sprintf("%s.svc",
		ComputeRequestNiFiNodeNamespace(nodeId, clusterName, namespace, headlessServiceEnabled, useExternalDNS))
}

func ComputeRequestNiFiNodeHostname(
	nodeId int32,
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool) string {

	return fmt.Sprintf("%s.%s",
		ComputeRequestNiFiNodeNamespaceFull(nodeId, clusterName, namespace, headlessServiceEnabled, useExternalDNS),
		clusterDomain)
}

func ComputeRequestNiFiNodeAddress(
	nodeId int32,
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool,
	internalListeners []v1alpha1.InternalListenerConfig) string {

	return fmt.Sprintf("%s:%d",
		ComputeRequestNiFiNodeHostname(nodeId, clusterName, namespace, headlessServiceEnabled, clusterDomain, useExternalDNS),
		InternalListenerForComm(internalListeners).ContainerPort)
}

func GenerateRequestNiFiNodeAddressFromCluster(nodeId int32, cluster *v1alpha1.NifiCluster) string {
	return ComputeRequestNiFiNodeAddress(
		nodeId,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.ListenersConfig.InternalListeners,
	)
}

func GenerateRequestNiFiNodeHostnameFromCluster(nodeId int32, cluster *v1alpha1.NifiCluster) string {
	return ComputeRequestNiFiNodeHostname(
		nodeId,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
	)
}

// >> All node
func ComputeRequestNiFiAllNodeService(clusterName string, headlessServiceEnabled bool) string {
	if headlessServiceEnabled {
		return fmt.Sprintf(HeadlessServiceTemplate, clusterName)
	}

	return fmt.Sprintf(AllNodeServiceTemplate, clusterName)
}

func ComputeRequestNiFiAllNodeNamespace(clusterName, namespace string, headlessServiceEnabled, useExternalDNS bool) string {
	if useExternalDNS {
		return ComputeRequestNiFiAllNodeService(clusterName, headlessServiceEnabled)
	}
	return fmt.Sprintf("%s.%s",
		ComputeRequestNiFiAllNodeService(clusterName, headlessServiceEnabled), namespace)
}

func ComputeRequestNiFiAllNodeNamespaceFull(
	clusterName, namespace string,
	headlessServiceEnabled, useExternalDNS bool) string {

	if useExternalDNS {
		return ComputeRequestNiFiAllNodeNamespace(clusterName, namespace, headlessServiceEnabled, useExternalDNS)
	}
	return fmt.Sprintf("%s.svc",
		ComputeRequestNiFiAllNodeNamespace(clusterName, namespace, headlessServiceEnabled, useExternalDNS))
}

func ComputeRequestNiFiAllNodeHostname(
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool) string {

	return fmt.Sprintf("%s.%s",
		ComputeRequestNiFiAllNodeNamespaceFull(clusterName, namespace, headlessServiceEnabled, useExternalDNS),
		clusterDomain)
}

func ComputeRequestNiFiAllNodeAddress(
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool,
	internalListeners []v1alpha1.InternalListenerConfig) string {

	return fmt.Sprintf("%s:%d",
		ComputeRequestNiFiAllNodeHostname(clusterName, namespace, headlessServiceEnabled, clusterDomain, useExternalDNS),
		InternalListenerForComm(internalListeners).ContainerPort)
}

func GenerateRequestNiFiAllNodeAddressFromCluster(cluster *v1alpha1.NifiCluster) string {
	return ComputeRequestNiFiAllNodeAddress(
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.ListenersConfig.InternalListeners,
	)
}

func GenerateRequestNiFiAllNodeHostnameFromCluster(cluster *v1alpha1.NifiCluster) string {
	return ComputeRequestNiFiAllNodeHostname(
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
	)
}

// > HostListener

func ComputeHostListenerNodeHostname(
	nodeId int32,
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool) string {

	return fmt.Sprintf("%s.%s", ComputeNodeName(nodeId, clusterName),
		ComputeRequestNiFiAllNodeHostname(clusterName, namespace, headlessServiceEnabled, clusterDomain, useExternalDNS))
}

func ComputeHostListenerNodeAddress(
	nodeId int32,
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool,
	internalListeners []v1alpha1.InternalListenerConfig) string {

	return fmt.Sprintf("%s:%d",
		ComputeHostListenerNodeHostname(nodeId, clusterName, namespace, headlessServiceEnabled, clusterDomain, useExternalDNS),
		InternalListenerForComm(internalListeners).ContainerPort)
}

func GenerateHostListenerNodeAddressFromCluster(nodeId int32, cluster *v1alpha1.NifiCluster) string {
	return ComputeHostListenerNodeAddress(
		nodeId,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.ListenersConfig.InternalListeners,
	)
}

func GenerateHostListenerNodeHostnameFromCluster(nodeId int32, cluster *v1alpha1.NifiCluster) string {
	return ComputeHostListenerNodeHostname(
		nodeId,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
	)
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

// LabelsForNifi returns the labels for selecting the resources
// belonging to the given Nifi CR name.
func LabelsForNifi(name string) map[string]string {
	return map[string]string{"app": "nifi", "nifi_cr": name}
}
