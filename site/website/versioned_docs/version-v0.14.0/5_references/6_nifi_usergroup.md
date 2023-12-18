---
id: 6_nifi_usergroup
title: NiFi UserGroup
sidebar_label: NiFi UserGroup
---

`NifiUserGroup` is the Schema for the nifi user groups API.

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiUserGroup
metadata:
  name: group-test
spec:
  clusterRef:
    name: nc
    namespace: nifikop
  usersRef:
    - name: nc-0-node.nc-headless.nifikop.svc.cluster.local
    - name: nc-controller.nifikop.mgt.cluster.local
  accessPolicies:
    - type: global
      action: read
      resource: /counters
```

## NifiUserGroup
|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects usergroups must create.|No|nil|
|spec|[NifiUserGroupSpec](#nifiusergroupspec)|defines the desired state of NifiUserGroup.|No|nil|
|status|[NifiUserGroupStatus](#nifiusergroupstatus)|defines the observed state of NifiUserGroup.|No|nil|

## NifiUserGroupSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|clusterRef|[ClusterReference](./2_nifi_user.md#clusterreference)|  contains the reference to the NifiCluster with the one the user is linked. |Yes| - |
|usersRef|\[&nbsp;\][UserReference](#userref)| contains the list of reference to NifiUsers that are part to the group. |No| [] |
|accessPolicies|\[&nbsp;\][AccessPolicy](./2_nifi_user.md#accesspolicy)| defines the list of access policies that will be granted to the group. |No| [] |

## NifiUserGroupStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|id|string| the nifi usergroup's node id.|Yes| - |
|version|string| the last nifi usergroup's node revision version catched.|Yes| - |

## UserReference

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string| name of the NifiUser. |Yes| - |
|namespace|string| the NifiUser namespace location. |Yes| - |

