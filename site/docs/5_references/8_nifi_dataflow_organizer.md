---
id: 8_nifi_dataflow_organizer
title: NiFi Dataflow Organizer
sidebar_label: NiFi Dataflow Organizer
---

`NifiDataflowOrganizer` is the Schema through which you configure automatic positioning and grouping of `NifiDataflow`.

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiDataflowOrganizer
metadata:
  name: nifikop
  namespace: instances
spec:
  maxWidth: 1000
  initialPosition:
    posX: 0
    posY: 0
  clusterRef:
    name: nc
    namespace: nifikop
  groups:
    "group 1":
      color: '#9c0300'
      fontSize: 18px
      maxColumnSize: 1
      dataflowRef:
      - name: df1_group1
    "group 2":
      color: '#03009c'
      fontSize: 18px
      maxColumnSize: 2
      dataflowRef:
      - name: df1_group2
      - name: df2_group2
    "group 3":
      color: '#9c6000'
      fontSize: 18px
      maxColumnSize: 3
      dataflowRef:
      - name: df1_group3
      - name: df2_group3
      - name: df3_group3
      - name: df4_group3
    "group 4":
      color: '#229c00'
      fontSize: 18px
      maxColumnSize: 4
      dataflowRef:
      dataflowRef:
      - name: df1_group4
      - name: df2_group4
      - name: df3_group4
```

## NifiDataflowOrganizer

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects dataflows must create.|No|nil|
|spec|[NifiDataflowOrganizerSpec](#nifidatafloworganizerspec)|defines the desired state of NifiDataflowOrganizer.|No|nil|
|status|[NifiDataflowOrganizerStatus](#nifidatafloworganizerstatus)|defines the observed state of NifiDataflowOrganizer.|No|nil|

## NifiDataflowOrganizerSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|clusterRef|[ClusterReference](./2_nifi_user#clusterreference)| contains the reference to the NifiCluster with the one the user is linked. |Yes| - |
|maxWidth|int|the maximum width before moving to the next line. |No| 1000 |
|initalPosition|[OrganizerGroupPosition](#organizergroupposition)|the initial position of all the groups. |No| {"posX": 0, "posY": 0} |
|groups|map\[string\][OrganizerGroup](#organizergroup)|the groups of dataflow to organize. |Yes| - |

## NifiDataflowOrganizerStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|groupStatus|map\[string\][OrganizerGroupStatus](#organizergroupstatus)| the status of the groups. |Yes| - |

## OrganizerGroupStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|titleStatus|[OrganizerGroupTitleStatus](#organizergrouptitlestatus)| the status of the title label. |Yes| - |
|contentStatus|[OrganizerGroupContentStatus](#organizergroupcontentstatus)| the status of the content label. |Yes| - |

## OrganizerGroupTitleStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|id|string| the id of the title label. |Yes| - |

## OrganizerGroupContentStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|id|string| the id of the content label. |Yes| - |

## OrganizerGroupPosition

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|posX|string|the x coordinate. |No| 0 |
|posY|string|the y coordinate. |No| 0 |

## OrganizerGroup

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|parentProcessGroupID|string|the UUID of the parent process group where you want to deploy your groups, if not set deploy at root level. |No| - |
|color|string|the color of the group. |No| #FFF7D7 |
|fontSize|string|the font size of the group. |No| 18px |
|maxColumnSize|int|the maximum number of dataflow on the same line. |No| 5 |
