---
id: 5_node_state
title: Node state
sidebar_label: Node state
---

Holds information about nifi state

## NodeState

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|gracefulActionState|[GracefulActionState](#gracefulactionstate)| holds info about nifi cluster action status.| - | - |
|configurationState|[ConfigurationState](#configurationstate)| holds info about the config.| - | - |
|initClusterNode|[InitClusterNode](#initclusternode)| contains if this nodes was part of the initial cluster.| - | - |
|podIsReady|bool| True if the pod for this node is up and running. Otherwise false.| - | - |
|creationTime|[v1.Time](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Time)| The time at which this node was created and added to the cluster| - | - |


## GracefulActionState 

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|errorMessage|string| holds the information what happened with Nifi Cluster. | - | "" |
|actionStep|[ActionStep](#actionstep)| holds info about the action step ran.| No | nil |
|taskStarted|string| hold the time when the execution started.| No | "" |
|actionState|[State](#state)| holds the information about Action state.| No | nil |

## ConfigurationState

|Name|Value|Description|
|-----|----|------------|
|ConfigInSync|ConfigInSync|states that the generated nodeConfig is in sync with the Node|
|ConfigOutOfSync|ConfigOutOfSync|states that the generated nodeConfig is out of sync with the Node|

## InitClusterNode

|Name|Value|Description|
|-----|----|------------|
|IsInitClusterNode|true|states the node is part of initial cluster setup|
|NotInitClusterNode|false|states the node is not part of initial cluster setup|

## State

### Upscale

|Name|Value|Description|
|-----|----|------------|
|GracefulUpscaleRequired|GracefulUpscaleRequired|states that a node upscale is required.|
|GracefulUpscaleRunning|GracefulUpscaleRunning|states that the node upscale task is still running.|
|GracefulUpscaleSucceeded|GracefulUpscaleSucceeded|states the node is updated gracefully.|

### Downscale

|Name|Value|Description|
|-----|----|------------|
|GracefulDownscaleRequired|GracefulDownscaleRequired|states that a node downscale is required|
|GracefulDownscaleRunning|GracefulDownscaleRunning|states that the node downscale is still running in|
|GracefulUpscaleSucceeded|GracefulUpscaleSucceeded|states the node is updated gracefully|

## ActionStep
|Name|Value|Description|
|-----|----|------------|
|DisconnectNodeAction|DISCONNECTING|states that the NiFi node is disconnecting from NiFi Cluster.|
|DisconnectStatus|DISCONNECTED|states that the NiFi node is disconnected from NiFi Cluster.|
|OffloadNodeAction|OFFLOADING|states that the NiFi node is offloading data to NiFi Cluster.|
|OffloadStatus|OFFLOADED|states that the NiFi node offloaded data to NiFi Cluster.|
|RemovePodAction|POD_REMOVING|states that the NiFi node pod and object related are removing by operator.|
|RemovePodStatus|POD_REMOVED|states that the NiFi node pod and object related have been removed by operator.|
|RemoveNodeAction|REMOVING|states that the NiFi node is removing from NiFi Cluster.|
|RemoveStatus|REMOVED|states that the NiFi node is removed from NiFi Cluster.|
|ConnectNodeAction|CONNECTING|states that the NiFi node is connecting to the NiFi Cluster.|
|ConnectStatus|CONNECTED|states that the NiFi node is connected to the NiFi Cluster.|