---
id: 3_nifi_registry_client
title: NiFi Registry Client
sidebar_label: NiFi Registry Client
---

`NifiRegistryClient` is the Schema for the NiFi registry client API.

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiRegistryClient
metadata:
  name: squidflow
spec:
  clusterRef:
    name: nc
    namespace: nifikop
  description: "Squidflow demo"
  uri: "http://nifi-registry:18080"
```

## NifiRegistryClient
|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects registry clients must create.|No|nil|
|spec|[NifiRegistryClientSpec](#nifiregistryclientspec)|defines the desired state of NifiRegistryClient.|No|nil|
|status|[NifiRegistryClientStatus](#nifiregistryclientstatus)|defines the observed state of NifiRegistryClient.|No|nil|

## NifiRegistryClientsSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|description|string| describes the Registry client. |No| - |
|uri|string| URI of the NiFi registry that should be used for pulling the flow. |Yes| - |
|clusterRef|[ClusterReference](./2_nifi_user#clusterreference)|  contains the reference to the NifiCluster with the one the user is linked. |Yes| - |

## NifiRegistryClientStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|id|string| nifi registry client's id. |Yes| - |
|version|int64| the last nifi registry client revision version catched. |Yes| - |