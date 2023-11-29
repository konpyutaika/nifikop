package nifi

import (
	"fmt"
	"strconv"
	"time"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

const (
	PrefixNodeNameTemplate = "%s-"
	SuffixNodeNameTemplate = "-node"
	RootNodeNameTemplate   = "%d"
	NodeNameTemplate       = PrefixNodeNameTemplate + RootNodeNameTemplate + SuffixNodeNameTemplate

	// TimeStampLayout defines the date format used.
	TimeStampLayout = "Mon, 2 Jan 2006 15:04:05 GMT"
)

var (
	NifiDataVolumeMountKey     = fmt.Sprintf("%s/nifi-data", v1.GroupVersion.Group)
	NifiVolumeReclaimPolicyKey = fmt.Sprintf("%s/nifi-volume-reclaim-policy", v1.GroupVersion.Group)
)

var (
	StopInputPortLabel  = fmt.Sprintf("%s/stop-input", v1.GroupVersion.Group)
	StopOutputPortLabel = fmt.Sprintf("%s/stop-output", v1.GroupVersion.Group)
	ForceStopLabel      = fmt.Sprintf("%s/force-stop", v1.GroupVersion.Group)
	ForceStartLabel     = fmt.Sprintf("%s/force-start", v1.GroupVersion.Group)
)

var (
	LastAppliedClusterAnnotation = fmt.Sprintf("%s/last-applied-nificluster", v1.GroupVersion.Group)
)

// ParseTimeStampToUnixTime parses the given CC timeStamp to time format.
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
// >> Node.
func ComputeNodeName(nodeId int32, clusterName string) string {
	return fmt.Sprintf(NodeNameTemplate, clusterName, nodeId)
}

func ComputeRequestNiFiNodeService(nodeId int32, clusterName string,
	headlessServiceEnabled bool, serviceTemplate string) string {
	if headlessServiceEnabled {
		return fmt.Sprintf("%s.%s",
			ComputeNodeName(nodeId, clusterName),
			fmt.Sprintf(serviceTemplate, clusterName))
	}

	return ComputeNodeName(nodeId, clusterName)
}

func ComputeRequestNiFiNodeNamespace(nodeId int32, clusterName, namespace string, headlessServiceEnabled,
	useExternalDNS bool, serviceTemplate string) string {
	if useExternalDNS {
		return ComputeRequestNiFiNodeService(nodeId, clusterName, headlessServiceEnabled, serviceTemplate)
	}
	return fmt.Sprintf("%s.%s",
		ComputeRequestNiFiNodeService(nodeId, clusterName, headlessServiceEnabled, serviceTemplate), namespace)
}

func ComputeRequestNiFiNodeNamespaceFull(
	nodeId int32,
	clusterName, namespace string,
	headlessServiceEnabled, useExternalDNS bool,
	serviceTemplate string) string {
	if useExternalDNS {
		return ComputeRequestNiFiNodeNamespace(nodeId, clusterName, namespace, headlessServiceEnabled,
			useExternalDNS, serviceTemplate)
	}
	return fmt.Sprintf("%s.svc",
		ComputeRequestNiFiNodeNamespace(nodeId, clusterName, namespace, headlessServiceEnabled,
			useExternalDNS, serviceTemplate))
}

func ComputeRequestNiFiNodeHostname(
	nodeId int32,
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool,
	serviceTemplate string) string {
	return fmt.Sprintf("%s.%s",
		ComputeRequestNiFiNodeNamespaceFull(nodeId, clusterName, namespace,
			headlessServiceEnabled, useExternalDNS, serviceTemplate),
		clusterDomain)
}

func ComputeRequestNiFiNodeAddress(
	nodeId int32,
	clusterName, namespace string,
	headlessServiceEnabled bool,
	clusterDomain string,
	useExternalDNS bool,
	internalListeners []v1.InternalListenerConfig,
	serviceTemplate string) string {
	return fmt.Sprintf("%s:%d",
		ComputeRequestNiFiNodeHostname(nodeId, clusterName, namespace,
			headlessServiceEnabled, clusterDomain, useExternalDNS, serviceTemplate),
		InternalListenerForComm(internalListeners).ContainerPort)
}

func GenerateRequestNiFiNodeAddressFromCluster(nodeId int32, cluster *v1.NifiCluster) string {
	return ComputeRequestNiFiNodeAddress(
		nodeId,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.ListenersConfig.InternalListeners,
		cluster.Spec.Service.GetServiceTemplate(),
	)
}

func GenerateRequestNiFiNodeHostnameFromCluster(nodeId int32, cluster *v1.NifiCluster) string {
	return ComputeRequestNiFiNodeHostname(
		nodeId,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.Service.HeadlessEnabled,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.Service.GetServiceTemplate(),
	)
}

// >> All node.
func ComputeRequestNiFiAllNodeService(
	clusterName string, serviceTemplate string) string {
	return fmt.Sprintf(serviceTemplate, clusterName)
}

func ComputeRequestNiFiAllNodeNamespace(
	clusterName, namespace string, useExternalDNS bool, serviceTemplate string) string {
	if useExternalDNS {
		return ComputeRequestNiFiAllNodeService(clusterName, serviceTemplate)
	}
	return fmt.Sprintf("%s.%s",
		ComputeRequestNiFiAllNodeService(clusterName, serviceTemplate), namespace)
}

func ComputeRequestNiFiAllNodeNamespaceFull(
	clusterName, namespace string, useExternalDNS bool, serviceTemplate string) string {
	if useExternalDNS {
		return ComputeRequestNiFiAllNodeNamespace(clusterName, namespace, useExternalDNS, serviceTemplate)
	}
	return fmt.Sprintf("%s.svc",
		ComputeRequestNiFiAllNodeNamespace(clusterName, namespace, useExternalDNS, serviceTemplate))
}

func ComputeRequestNiFiAllNodeHostname(
	clusterName, namespace string,
	clusterDomain string,
	useExternalDNS bool,
	serviceTemplate string) string {
	return fmt.Sprintf("%s.%s",
		ComputeRequestNiFiAllNodeNamespaceFull(clusterName, namespace, useExternalDNS, serviceTemplate),
		clusterDomain)
}

func ComputeRequestNiFiAllNodeAddress(
	clusterName, namespace string,
	clusterDomain string,
	useExternalDNS bool,
	internalListeners []v1.InternalListenerConfig,
	serviceTemplate string) string {
	return fmt.Sprintf("%s:%d",
		ComputeRequestNiFiAllNodeHostname(clusterName, namespace, clusterDomain, useExternalDNS, serviceTemplate),
		InternalListenerForComm(internalListeners).ContainerPort)
}

func GenerateRequestNiFiAllNodeAddressFromCluster(cluster *v1.NifiCluster) string {
	return ComputeRequestNiFiAllNodeAddress(
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.ListenersConfig.InternalListeners,
		cluster.Spec.Service.GetServiceTemplate(),
	)
}

func GenerateRequestNiFiAllNodeHostnameFromCluster(cluster *v1.NifiCluster) string {
	return ComputeRequestNiFiAllNodeHostname(
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.Service.GetServiceTemplate(),
	)
}

// > HostListener

func ComputeHostListenerNodeHostname(
	nodeId int32,
	clusterName, namespace string,
	clusterDomain string,
	useExternalDNS bool,
	serviceTemplate string) string {
	return fmt.Sprintf("%s.%s", ComputeNodeName(nodeId, clusterName),
		ComputeRequestNiFiAllNodeHostname(clusterName, namespace, clusterDomain, useExternalDNS, serviceTemplate))
}

func ComputeHostListenerNodeAddress(
	nodeId int32,
	clusterName, namespace string,
	clusterDomain string,
	useExternalDNS bool,
	internalListeners []v1.InternalListenerConfig,
	serviceTemplate string) string {
	return fmt.Sprintf("%s:%d",
		ComputeHostListenerNodeHostname(nodeId, clusterName, namespace, clusterDomain, useExternalDNS, serviceTemplate),
		InternalListenerForComm(internalListeners).ContainerPort)
}

func GenerateHostListenerNodeAddressFromCluster(nodeId int32, cluster *v1.NifiCluster) string {
	return ComputeHostListenerNodeAddress(
		nodeId,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.ListenersConfig.InternalListeners,
		cluster.Spec.Service.GetServiceTemplate(),
	)
}

func GenerateHostListenerNodeHostnameFromCluster(nodeId int32, cluster *v1.NifiCluster) string {
	return ComputeHostListenerNodeHostname(
		nodeId,
		cluster.Name,
		cluster.Namespace,
		cluster.Spec.ListenersConfig.GetClusterDomain(),
		cluster.Spec.ListenersConfig.UseExternalDNS,
		cluster.Spec.Service.GetServiceTemplate(),
	)
}

func InternalListenerForComm(internalListeners []v1.InternalListenerConfig) v1.InternalListenerConfig {
	return internalListeners[determineInternalListenerForComm(internalListeners)]
}

func determineInternalListenerForComm(internalListeners []v1.InternalListenerConfig) int {
	var serverPortId int
	for id, iListener := range internalListeners {
		if iListener.Type == v1.HttpsListenerType {
			serverPortId = id
			break
		} else if iListener.Type == v1.HttpListenerType {
			serverPortId = id
		}
	}
	return serverPortId
}

// LabelsForNifi returns the labels for selecting the resources
// belonging to the given Nifi CR name.
func LabelsForNifi(name string) map[string]string {
	return map[string]string{"app": "nifi", "nifi_cr": name}
}
