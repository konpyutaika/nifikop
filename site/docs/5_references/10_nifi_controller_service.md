---
id: 10_nifi_controller_service
title: NiFi Controller Service
sidebar_label: NiFi Controller Service
---

`NifiControllerService` is the Schema for the NiFi controller service API. It manages controller services defined at controller level (**Controller Settings** in the NiFi UI), the scope that controller-scoped components such as flow registry clients and reporting tasks can reference.

The only type supported today is `org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider`, the HTTP client provider required by the Bitbucket flow registry client.

## Quick example

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiControllerService
metadata:
  name: web-client
spec:
  clusterRef:
    name: nc
    namespace: nifikop
  description: "Shared HTTP client for registry clients"
  type: org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider
  standardWebClientServiceProviderSpec:
    connectTimeout: "10 secs"
    readTimeout: "10 secs"
    writeTimeout: "10 secs"
```

## Lifecycle

The operator creates the service, then enables it. A controller service is only usable by other components once its state is `ENABLED`, and NiFi refuses configuration changes and deletion while a service is enabled. The operator therefore drives the run state itself:

- after creation, and after any configuration change, the service is enabled;
- when the configuration drifts from the resource, the service is disabled first, updated, then enabled again;
- on deletion, the service is disabled before being removed.

NiFi applies state changes asynchronously. While a transition is in progress the status reports `ENABLING` or `DISABLING` and the operator requeues after a few seconds.

Deletion is refused while other components, for example a `NifiRegistryClient`, still reference the service. The resource stays with its finalizer, a `StillReferenced` warning event names the referencing components, and the operator retries at the regular requeue interval. Remove the referencing resources first.

Only the properties described by the resource are managed. Properties set directly in NiFi on the same service, such as an SSL context service or a proxy configuration, are left untouched.

## NifiControllerService

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|Metadata that all persisted resources must have, which includes all objects controller services must create.|No|nil|
|spec|[NifiControllerServiceSpec](#nificontrollerservicespec)|Defines the desired state of NifiControllerService.|No|nil|
|status|[NifiControllerServiceStatus](#nificontrollerservicestatus)|Defines the observed state of NifiControllerService.|No|nil|

## NifiControllerServiceSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|description|string|Describes the controller service. Stored as the component comments in NiFi.|No| - |
|clusterRef|[ClusterReference](./2_nifi_user#clusterreference)|Contains the reference to the NifiCluster with which the controller service is linked.|Yes| - |
|type|Enum=`org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider`|The NiFi class name of the controller service.|No|`org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider`|
|standardWebClientServiceProviderSpec|[StandardWebClientServiceProviderSpec](#standardwebclientserviceproviderspec)|Configuration of a StandardWebClientServiceProvider. Required when `type` is `org.apache.nifi.web.client.provider.service.StandardWebClientServiceProvider`.|No| - |

:::note
A CEL validation rule on the CRD enforces that `standardWebClientServiceProviderSpec` is present when `type` selects the StandardWebClientServiceProvider.
:::

## StandardWebClientServiceProviderSpec

Mirrors the properties of the NiFi `StandardWebClientServiceProvider`. Durations use the NiFi time period format, for example `"10 secs"` or `"1 min"`.

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|connectTimeout|string|Maximum amount of time to wait before failing during initial socket connection.|No|`10 secs`|
|readTimeout|string|Maximum amount of time to wait before failing while reading socket responses.|No|`10 secs`|
|writeTimeout|string|Maximum amount of time to wait before failing while writing socket requests.|No|`10 secs`|

## NifiControllerServiceStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|id|string|The nifi controller service's id.|Yes| - |
|version|int64|The last nifi controller service revision version.|Yes| - |
|state|Enum=`ENABLED`;`ENABLING`;`DISABLED`;`DISABLING`|The last run state of the controller service observed in NiFi.|No| - |

`kubectl get nificontrollerservices` shows the type and the observed state as columns.
