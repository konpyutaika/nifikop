---
id: 1_nifi_user
title: NiFi User
sidebar_label: NiFi user
---

`NifiUser` is the Schema for the nifi users API.

```yaml
apiVersion: nifi.orange.com/v1alpha1
kind: NifiUser
metadata:
  name: example-user
  namespace: nifi
spec:
  clusterRef:
    name: nifi
  secretName: example-user-secret
  includeJKS: true
```

## NifiUser
|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects users must create.|No|nil|
|spec|[NifiUserSpec](#nifiuserspec)|defines the desired state of NifiUser.|No|nil|
|status|[NifiUserStatus](#nifiuserstatus)|defines the observed state of NifiUser.|No|nil|

## NifiUserSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|secretName|string|  name of the secret where all cert resources will be stored. |Yes| - |
|clusterRef|[ClusterReference](#clusterreference)|  contains the reference to the NifiCluster with the one the user is linked. |Yes| - |
|DNSNames|[]string| list of DNSNames that the user will used to request the NifiCluster (allowing to create the right certificates associated). |Yes| - |
|includeJKS|boolean|  whether or not the the operator also include a Java keystore format (JKS) with you secret. |Yes| - |


## NifiUserStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|state|[UserState](#userstate)|Store the state of each nifi node.|Yes| - |


## UserState

|Name|Value|Description|
|-----|----|------------|
|UserStateCreated|created|describes the status of a NifiUser as created|

## ClusterReference

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string|  name of the NifiCluster. |Yes| - |
|namespace|string|  the NifiCluster namespace location. |Yes| - |