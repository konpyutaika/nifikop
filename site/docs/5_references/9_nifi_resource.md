---
id: 8_nifi_resource
title: NiFi Resource
sidebar_label: NiFi Resource
---

`NifiResource` is a generic Schema for the NiFi API.

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiResource
metadata:
  name: resource
  namespace: instances
spec:
  parentProcessGroupID: "16cfd2ec-0174-1000-0000-00004b9b35cc"
  name: Process Group Instance
  type: process-group
  comments: Example Process Group
  # parentProcessGroupRef:
  #   name: parent-dataflow-lifecycle
  #   namespace: nifikop
  configuration:
    defaultBackPressureDataSizeThreshold: 1 GB
    defaultBackPressureObjectThreshold: 10000
		defaultFlowFileExpiration: 30 secs
    executionEngine: INHERITED
    flowfileConcurrency: UNBOUNDED
    flowfileOutboundPolicy: STREAM_WHEN_AVAILABLE
    logFileSuffix: example
    maxConcurrentTasks: 4
    statelessFlowTimeout: 30 secs
    position:
      posX: 0
      posY: 0
    parameterContextRef:
      name: dataflow-lifecycle
      namespace: nifikop
    updateStrategy: drain
  clusterRef:
    name: nc
    namespace: nifikop
```

## NifiResource

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|-------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects dataflows must create.|No|nil|
|spec|[NifiResourceSpec](#nifiresourcespec)|defines the desired state of NifiResource.|No|nil|
|status|[NifiResourceStatus](#nifiresourcestatus)|defines the observed state of NifiResource.|No|nil|

## NifiResourceSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|-------|
|parentProcessGroupID|string|the UUID of the parent process group where you want to deploy your resource, if not set deploy at root level. |No| - |
|name|string|the Name that will be used within NiFi to refer to this resource. |Yes| - |
|type|[ResourceType](#resourcetype)|the Type of resource this manifest refers to. |Yes| - |
|comments|string|the user defined comments to add into the resource within NiFi |Yes| - |
|configuration|[ResourceConfiguration](#resourceconfiguration)|  |Yes| drain |
|clusterRef|[ClusterReference](./2_nifi_user#clusterreference)| contains the reference to the NifiCluster with the one the resource is linked. |Yes| - |
|parentProcessGroupRef|[ResourceReference](./9_nifi_resource#resourcereference)| contains the reference to the Resource with the one the dataflow is linked. |No| - |

## NifiResourceStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|-------|
|UUID|string| the UUID of the resource within NiFi. |Yes| - |
|state|[ResourceState](#resourcestate)| the resource current state. |Yes| - |

## ResourceState

|Name|Value|Description|
|-----|----|------------|
|ResourceStateCreated|Created|describes the status of a NifiResource as created.|
|DataflowStateStarting|Starting|describes the status of a NifiResource as starting.|
|ResourceStateRan|Ran|describes the status of a NifiResource as running.|
|ResourceStateOutOfSync|OutOfSync|describes the status of a NifiResource as out of sync.|
|ResourceStateInSync|InSync|describes the status of a NifiResource as in sync.|

## ResourceType

|Name|Value|Description|
|----|-----|-----------|
|ResourceProcessGroup|process-group|indicates that the reource is a Process Group.|
|ResourceInputPort|input-port|indicates that the resource is a Input Port. **(not implemented)**|
|ResourceOutputPort|output-port|indicates that the resource is a Output Port. **(not implemented)**|
|ResourceProcessor|processor|indicates that the resource is a Processor. **(not implemented)**|
|ResourceFunnel|funnel|indicates that the resource is a Funnel. **(not implemented)**|
|ResourceControllerService|controller-service|indicates that the resource is a Controller Service. **(not implemented)**|

## ResourceConfiguration

### Process Group

|Name|Value|Description|Required|Default|
|----|-----|-----------|--------|-------|
|defaultFlowFileExpiration|string|the maximum amount of time an object may be in the flow before it will be automatically aged out of the flow.|No| - |
|defaultBackPressureDataSizeThreshold|string|the maximum data size of objects that can be queued before back pressure is applied.|No| 1 GB |
|defaultBackPressureObjectThreshold|*int64|the maximum number of objects that can be queued before back pressure is applied.|No| 10000 |
|executionEngine|[ProcessGroupExecutionEngine](#processgroupexecutionengine)|The Execution Engine that should be used to run the components within the group.|No| - |
|flowfileConcurrency|string|The configured FlowFile Concurrency for the Process Group|No| - |
|flowfileOutboundPolicy|string|The FlowFile Outbound Policy for the Process Group|No| - |
|logFileSuffix|string|The log file suffix for this Process Group for dedicated logging.|No| - |
|position|[ResourcePosition](#flowposition)|the position of your resource in the canvas. |No| - |
|statelessFlowTimeout|string| The maximum amount of time that the flow is allows to run using the Stateless engine before it times out and is considered a failure|No| - |
|maxConcurrentTasks|int32|The maximum number of concurrent tasks that should be scheduled for this Process Group when using the Stateless Engine|No| - |
|parameterContextRef|[ParameterContextReference](./4_nifi_parameter_context#parametercontextreference)| contains the reference to the ParameterContext with the one the dataflow is linked. |No| - |
|updateStrategy|[ComponentUpdateStrategy](./5_nifi_dataflow##componentupdatestrategy)|describes the way the operator will deal with data when a resource will be updated: Drop or Drain |No| drain |
|skipInvalidControllerService|bool|whether the flow is considered as ran if some controller services are still invalid or not. |No| false |
|skipInvalidComponent|bool|whether the flow is considered as ran if some components are still invalid or not. |No| false |

## ResourcePosiiton

|Name|Value|Description|Required|Default|
|----|-----|-----------|--------|-------|
|posX|*int64|the x coordinate.|No| - |
|posY|*int64|the y coordinate.|No| - |

## ProcessGroupExecutionEngine

|Name|Value|Description|
|----|-----|-----------|
|STANDARD|STANDARD|the standard execution engine within Nifi |
|STATELESS|STATELESS|no state will be stored.|
|INHERITED|INHERITED|use the execution engine defined in the parent process group.|

## ResourceReference

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string| name of the NifiResource. |Yes| - |
|namespace|string| the NifiResource namespace location. |No| - |