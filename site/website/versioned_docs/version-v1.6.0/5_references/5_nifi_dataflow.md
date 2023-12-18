---
id: 5_nifi_dataflow
title: NiFi Dataflow
sidebar_label: NiFi Dataflow
---

`NifiDataflow` is the Schema for the NiFi dataflow API.

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiDataflow
metadata:
  name: dataflow-lifecycle
spec:
  parentProcessGroupID: "16cfd2ec-0174-1000-0000-00004b9b35cc"
  bucketId: "01ced6cc-0378-4893-9403-f6c70d080d4f"
  flowId: "9b2fb465-fb45-49e7-94fe-45b16b642ac9"
  flowVersion: 2
  flowPosition:
    posX: 0
    posY: 0
  syncMode: always
  skipInvalidControllerService: true
  skipInvalidComponent: true
  clusterRef:
    name: nc
    namespace: nifikop
  registryClientRef:
    name: squidflow
    namespace: nifikop
  parameterContextRef:
    name: dataflow-lifecycle
    namespace: nifikop
  updateStrategy: drain
```

## NifiDataflow

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects dataflows must create.|No|nil|
|spec|[NifiDataflowSpec](#nifidataflowspec)|defines the desired state of NifiDataflow.|No|nil|
|status|[NifiDataflowStatus](#nifidataflowstatus)|defines the observed state of NifiDataflow.|No|nil|


## NifiDataflowsSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|parentProcessGroupID|string|the UUID of the parent process group where you want to deploy your dataflow, if not set deploy at root level. |No| - |
|bucketId|string|the UUID of the Bucket containing the flow. |Yes| - |
|flowId|string|the UUID of the flow to run. |Yes| - |
|flowVersion|*int32|the version of the flow to run. |Yes| - |
|flowPosition|[FlowPosition](#flowposition)|the position of your dataflow in the canvas. |No| - |
|syncMode|Enum={"never","always","once"}|if the flow will be synchronized once, continuously or never. |No| always |
|skipInvalidControllerService|bool|whether the flow is considered as ran if some controller services are still invalid or not. |Yes| false |
|skipInvalidComponent|bool|whether the flow is considered as ran if some components are still invalid or not. |Yes| false |
|updateStrategy|[ComponentUpdateStrategy](#componentupdatestrategy)|describes the way the operator will deal with data when a dataflow will be updated : Drop or Drain |Yes| drain |
|clusterRef|[ClusterReference](./2_nifi_user#clusterreference)| contains the reference to the NifiCluster with the one the user is linked. |Yes| - |
|parameterContextRef|[ParameterContextReference](./4_nifi_parameter_context#parametercontextreference)| contains the reference to the ParameterContext with the one the dataflow is linked. |No| - |
|registryClientRef|[RegistryClientReference](./3_nifi_registry_client#registryclientreference)| contains the reference to the NifiRegistry with the one the dataflow is linked. |Yes| - |

## NifiDataflowStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|processGroupID|string| process Group ID. |Yes| - |
|state|[DataflowState](#dataflowstate)| the dataflow current state. |Yes| - |
|latestUpdateRequest|[UpdateRequest](#updaterequest)|the latest update request sent. |Yes| - |
|latestDropRequest|[DropRequest](#droprequest)|the latest queue drop request sent. |Yes| - |

## ComponentUpdateStrategy

|Name|Value|Description|
|-----|----|------------|
|DrainStrategy|drain|leads to shutting down only input components (Input processors, remote input process group) and dropping all flowfiles from the flow.|
|DropStrategy|drop|leads to shutting down all components and dropping all flowfiles from the flow.|

## DataflowState

|Name|Value|Description|
|-----|----|------------|
|DataflowStateCreated|Created|describes the status of a NifiDataflow as created.|
|DataflowStateStarting|Starting|describes the status of a NifiDataflow as starting.|
|DataflowStateRan|Ran|describes the status of a NifiDataflow as running.|
|DataflowStateOutOfSync|OutOfSync|describes the status of a NifiDataflow as out of sync.|
|DataflowStateInSync|InSync|describes the status of a NifiDataflow as in sync.|

## UpdateRequest

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|type|[DataflowUpdateRequestType](#dataflowupdaterequesttype)|defines the type of versioned flow update request. |Yes| - |
|id|string|the id of the update request. |Yes| - |
|uri|string|the uri for this request. |Yes| - |
|lastUpdated|string|the last time this request was updated. |Yes| - |
|complete|bool| whether or not this request has completed. |Yes| false |
|failureReason|string| an explication of why the request failed, or null if this request has not failed. |Yes| - |
|percentCompleted|int32|  the percentage complete of the request, between 0 and 100. |Yes| 0 |
|state|string| the state of the request. |Yes| - |
|notFound|bool| whether or not this request was found. |Yes| false |
|notFoundRetryCount|int32| the number of consecutive retries made in case of a NotFound error (limit: 3). |Yes| 0 |

## DropRequest

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|connectionId|string|the connection id. |Yes| - |
|id|string|the id for this drop request. |Yes| - |
|uri|string|the uri for this request. |Yes| - |
|lastUpdated|string|the last time this request was updated. |Yes| - |
|finished|bool|whether the request has finished. |Yes| false |
|failureReason|string|an explication of why the request failed, or null if this request has not failed. |Yes| - |
|percentCompleted|int32|the percentage complete of the request, between 0 and 100. |Yes| 0 |
|currentCount|int32|the number of flow files currently queued. |Yes| 0 |
|currentSize|int64| the size of flow files currently queued in bytes. |Yes| 0 |
|current|string|the count and size of flow files currently queued. |Yes| - |
|originalCount|int32|the number of flow files to be dropped as a result of this request. |Yes| 0 |
|originalSize|int64| the size of flow files to be dropped as a result of this request in bytes. |Yes| 0 |
|original|string|the count and size of flow files to be dropped as a result of this request. |Yes| - |
|droppedCount|int32|the number of flow files that have been dropped thus far. |Yes| 0 |
|droppedSize|int64| the size of flow files currently queued in bytes. |Yes| 0 |
|Dropped|string|the count and size of flow files that have been dropped thus far. |Yes| - |
|state|string|the state of the request. |Yes| - |
|notFound|bool|whether or not this request was found. |Yes| false |
|notFoundRetryCount|int32| the number of consecutive retries made in case of a NotFound error (limit: 3). |Yes| 0 |
	
## DataflowUpdateRequestType

|Name|Value|Description|
|-----|----|------------|
|RevertRequestType|Revert|defines a revert changes request.|
|UpdateRequestType|Update|defines an update version request.|

## FlowPosition

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|posX|int64|the x coordinate. |No| - |
|posY|int64|the y coordinate. |No| - |