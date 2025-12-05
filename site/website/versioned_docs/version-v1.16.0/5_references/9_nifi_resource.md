---
id: 9_nifi_resource
title: NiFi Resource
sidebar_label: NiFi Resource
---

`NifiResource` is the Schema for multiple the NiFi APIs to create different resources.

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiResource
metadata:
  name: process-group-parent
  namespace: nifikop
spec:
  clusterRef:
    name: nc
    namespace: nifikop
  type: process-group
  displayName: Process Group (Parent)
  configuration:
    comments: "I'm the parent"
    logFileSuffix: .parent
    position:
      x: -150
      y: 0
---
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiResource
metadata:
  name: process-group-child
  namespace: nifikop
spec:
  clusterRef:
    name: nc
    namespace: nifikop
  type: process-group
  displayName: Process Group (child)
  parentProcessGroupRef:
    name: process-group-parent
  configuration:
    comments: "I'm the child"
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
|clusterRef|[ClusterReference](./2_nifi_user#clusterreference)|  contains the reference to the NifiCluster with the one the user is linked. |Yes| - |
|type|[ComponentType](#resourcetype)|the type of the resource (e.g. process-group).|Yes| - |
|parentProcessGroupID|string|the UUID of the parent process group where you want to deploy your resource, if not set deploy at root level (is not used for all types of resource). |No| - |
|parentProcessGroupRef|[ResourceReference](#resourcereference)|the reference to the parent process group where you want to deploy your resource, if not set deploy at root level (is not used for all types of resource). |No| - |
|displayName|string|the name of the resource (if not set, the name of the CR will be used). |No| - |
|Configuration|[RawExtension](https://pkg.go.dev/k8s.io/apimachinery/pkg/runtime#RawExtension)|the configuration of the resource (e.g. the process group configuration). |No| - |

## NifiResourceStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|id|string| nifi resource's id. |Yes| - |
|version|int64| the last nifi resource revision version catched. |Yes| - |

## ComponentType

|Name|Value|Description|
|----|-----|-----------|
|ResourcetInputPort|input-port|indicates that the resource created is an `input port`. **(not implemented)**|
|ResourceOutputPort|output-port|indicates that the resource created is an `output port`. **(not implemented)**|
|ResourceProcessor|processor|indicates that the resource created is a `processor`. **(not implemented)**|
|ResourceFunnel|funnel|indicates that the resource created is a `funnel`. **(not implemented)**|
|ResourceProcessGroup|process-group|indicates that the resource created is a `process group`.|
|ResourceControllerService|controller-service|indicates that the resource created is a `controller service`. **(not implemented)**|

## ResourceReference

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string|  name of the NifiResource. |Yes| - |
|namespace|string|  the NifiResource namespace location. |Yes| - |