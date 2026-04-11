---
id: 8_nifi_connection
title: NiFi Connection
sidebar_label: NiFi Connection
---

`NifiConnection` is the Schema for the NiFi connection API.

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiConnection
metadata:
  name: connection
  namespace: instances
spec:
  source:
    name: input
    namespace: instances
    subName: output_1
    type: dataflow
  destination:
    name: output
    namespace: instances
    subName: input_1
    type: dataflow
  configuration:
    flowFileExpiration: 1 hour
    backPressureDataSizeThreshold: 100 GB
    backPressureObjectThreshold: 10000
    loadBalanceStrategy: PARTITION_BY_ATTRIBUTE
    loadBalancePartitionAttribute: partition_attribute
    loadBalanceCompression: DO_NOT_COMPRESS
    prioritizers: 
      - NewestFlowFileFirstPrioritizer
      - FirstInFirstOutPrioritizer
    labelIndex: 0
    bends:
      - posX: 550
        posY: 550
      - posX: 550
        posY: 440
      - posX: 550
        posY: 88
  updateStrategy: drain
```

## NifiDataflow

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|-------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects dataflows must create.|No|nil|
|spec|[NifiConnectionSpec](#nificonnectionspec)|defines the desired state of NifiDataflow.|No|nil|
|status|[NifiConnectionStatus](#nificonnectionstatus)|defines the observed state of NifiDataflow.|No|nil|

## NifiConnectionSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|-------|
|source|[ComponentReference](#componentreference)|the Source component of the connection. |Yes| - |
|destination|[ComponentReference](#componentreference)|the Destination component of the connection. |Yes| - |
|configuration|[ConnectionConfiguration](#connectionconfiguration)|the version of the flow to run. |Yes| - |
|updateStrategy|[ComponentUpdateStrategy](#componentupdatestrategy)|describes the way the operator will deal with data when a connection will be deleted: Drop or Drain |Yes| drain |

## NifiConnectionStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|-------|
|connectionID|string| connection ID. |Yes| - |
|state|[ConnectionState](#connectionstate)| the connection current state. |Yes| - |

## ComponentUpdateStrategy

|Name|Value|Description|
|----|-----|-----------|
|DrainStrategy|drain|leads to block stopping of input/output component until they are empty.|
|DropStrategy|drop|leads to dropping all flowfiles from the connection.|

## ConnectionState

|Name|Value|Description|
|----|-----|-----------|
|ConnectionStateCreated|Created|describes the status of a NifiConnection as created.|
|ConnectionStateOutOfSync|OutOfSync|describes the status of a NifiConnection as out of sync.|
|ConnectionStateInSync|InSync|describes the status of a NifiConnection as in sync.|

## ComponentReference

|Name|Value|Description|Required|Default|
|----|-----|-----------|--------|-------|
|name|string|the name of the component.|Yes| - |
|namespace|string|the namespace of the component.|Yes| - |
|type|[ComponentType](#componenttype)|the type of the component (e.g. nifidataflow).|Yes| - |
|subName|string|the name of the sub component (e.g. queue or port name).|No| - |

## ComponentType

|Name|Value|Description|
|----|-----|-----------|
|ComponentDataflow|dataflow|indicates that the component is a NifiDataflow.|
|ComponentInputPort|input-port|indicates that the component is a NifiInputPort. **(not implemented)**|
|ComponentOutputPort|output-port|indicates that the component is a NifiOutputPort. **(not implemented)**|
|ComponentProcessor|processor|indicates that the component is a NifiProcessor. **(not implemented)**|
|ComponentFunnel|funnel|indicates that the component is a NifiFunnel. **(not implemented)**|
|ComponentProcessGroup|process-group|indicates that the component is a NifiProcessGroup. **(not implemented)**|

## ConnectionConfiguration

|Name|Value|Description|Required|Default|
|----|-----|-----------|--------|-------|
|flowFileExpiration|string|the maximum amount of time an object may be in the flow before it will be automatically aged out of the flow.|No| - |
|backPressureDataSizeThreshold|string|the maximum data size of objects that can be queued before back pressure is applied.|No| 1 GB |
|backPressureObjectThreshold|*int64|the maximum number of objects that can be queued before back pressure is applied.|No| 10000 |
|loadBalanceStrategy|[ConnectionLoadBalanceStrategy](#connectionloadbalancestrategy)|how to load balance the data in this Connection across the nodes in the cluster.|No| DO_NOT_LOAD_BALANCE |
|loadBalancePartitionAttribute|string|the FlowFile Attribute to use for determining which node a FlowFile will go to.|No| - |
|loadBalanceCompression|[ConnectionLoadBalanceCompression](#connectionloadbalancecompression)|whether or not data should be compressed when being transferred between nodes in the cluster.|No| DO_NOT_COMPRESS |
|prioritizers|\[&nbsp;\][ConnectionPrioritizer](#connectionprioritizer)|the comparators used to prioritize the queue.|No| - |
|labelIndex|*int32|the index of the bend point where to place the connection label.|No| - |
|bends|\[&nbsp;\][ConnectionBend](#connectionbend)|the bend points on the connection.|No| - |

## ConnectionLoadBalanceStrategy

|Name|Value|Description|
|----|-----|-----------|
|StrategyDoNotLoadBalance|DO_NOT_LOAD_BALANCE|do not load balance FlowFiles between nodes in the cluster.|
|StrategyPartitionByAttribute|PARTITION_BY_ATTRIBUTE|determine which node to send a given FlowFile to based on the value of a user-specified FlowFile Attribute. All FlowFiles that have the same value for said Attribute will be sent to the same node in the cluster.|
|StrategyRoundRobin|ROUND_ROBIN|flowFiles will be distributed to nodes in the cluster in a Round-Robin fashion. However, if a node in the cluster is not able to receive data as fast as other nodes, that node may be skipped in one or more iterations in order to maximize throughput of data distribution across the cluster.|
|StrategySingle|SINGLE|all FlowFiles will be sent to the same node. Which node they are sent to is not defined.|

## ConnectionLoadBalanceCompression

|Name|Value|Description|
|----|-----|-----------|
|CompressionDoNotCompress|DO_NOT_COMPRESS|flowFiles will not be compressed.|
|CompressionCompressAttributesOnly|COMPRESS_ATTRIBUTES_ONLY|flowFiles' attributes will be compressed, but the flowFiles' contents will not be.|
|CompressionCompressAttributesAndContent|COMPRESS_ATTRIBUTES_AND_CONTENT|flowFiles' attributes and content will be compressed.|

## ConnectionPrioritizer

|Name|Value|Description|
|----|-----|-----------|
|PrioritizerFirstInFirstOutPrioritizer|FirstInFirstOutPrioritizer|given two FlowFiles, the one that reached the connection first will be processed first.|
|PrioritizerNewestFlowFileFirstPrioritizer|NewestFlowFileFirstPrioritizer|given two FlowFiles, the one that is newest in the dataflow will be processed first.|
|PrioritizerOldestFlowFileFirstPrioritizer|OldestFlowFileFirstPrioritizer|given two FlowFiles, the one that is oldest in the dataflow will be processed first. 'This is the default scheme that is used if no prioritizers are selected'.|
|PrioritizerPriorityAttributePrioritizer|PriorityAttributePrioritizer|given two FlowFiles, an attribute called “priority” will be extracted. The one that has the lowest priority value will be processed first.|

## ConnectionBend

|Name|Value|Description|Required|Default|
|----|-----|-----------|--------|-------|
|posX|*int64|the x coordinate.|No| - |
|posY|*int64|the y coordinate.|No| - |
