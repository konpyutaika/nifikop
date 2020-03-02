---
id: 3_cassandra_storage
title: Cassandra storage
sidebar_label: Cassandra storage
---
## Configuration

Cassandra is a stateful application. It needs to store data on disks. CassKop allows you to configure the type of
storage you want to use.

Storage can be configured using the `storage` property in `CassandraCluster.spec`

> **Important:** Once the Cassandra cluster is deployed, the storage cannot be changed.

Persistent storage uses Persistent Volume Claims to provision persistent volumes for storing data.
The `PersistentVolumes` are acquired using a `PersistentVolumeClaim` which is managed by CassKop. The
`PersistentVolumeClaim` can use a `StorageClass` to trigger automatic volume provisioning.

> It is recommended to uses local-storage with quick ssd disk access for low latency. We have only tested the
> `local-storage` storage class within CassKop.

CassandraCluster fragment of persistent storage definition :

```
# ...
  dataCapacity: "300Gi"
  dataStorageClass: "local-storage"
  deletePVC: true
# ...
```

- `dataCapacity` (required): Defines the size of the persistent volume claim, for example, "1000Gi".
- `dataStorageClass`(optional): Define the type of storage to uses (or use
  default one). We recommand to uses local-storage for better performances but
  it can be any storage with high ssd througput.
- `deletePVC`(optional): Boolean value which specifies if the Persistent Volume Claim has to be deleted when the cluster
  is deleted. Default is `false`.

> **WARNING**: Resizing persistent storage for existing CassandraCluster is not currently supported. You must decide the
> necessary storage size before deploying the cluster.

The above example asks that each nodes will have 300Gi of data volumes to persist the Cassandra data's using the
local-storage storage class provider.
The parameter deletePVC is used to control if the data storage must persist when the according statefulset is deleted.

> **WARNING:** If we don't specify dataCapacity, then CassKop will uses the Docker Container ephemeral storage, and
> all data will be lost in case of a cassandra node reboot.


## Persistent volume claim

When the persistent storage is used, it will create PersistentVolumeClaims with the following names:

`data-<cluster-name>-<dc-name>-<rack-name>-<idx>`

Persistent Volume Claim for the volume used for storing data to the cluster `<cluster-name>` for the Cassandra DC
`<dc-name>` and the rack `<rack-name>` for the Pod with ID `<idx>`.

> **IMPORTANT**: Note that with local-storage the PVC object makes a link between the Pod and the Node. While this
> object is existing the Pod will be sticked to the node chosen by the scheduler. In the case you want to move the
> Cassandra node to a new kubernetes node, you will need at some point to manually delete the associate PVC so that the
> scheduler can choose another Node for scheduling. This is cover in the Operation document.
