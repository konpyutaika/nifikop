---
id: 2_cluster_operations
title: Cluster operations
sidebar_label: Cluster operations
---

Those operations are applied at the Cassandra cluster level, as opposite to Pod operations that are executed at pod
level and are discussed in the next section.
Cluster Operations must only be triggered by a change made on the `CassandraCluster` object.

Some updates in the `CassandraCluster` CRD object are forbidden and will be gently dismissed by CassKop:

- `spec.dataCapacity`
- `spec.dataStorage`

Some Updates in the `CassandraCluster` CRD object will trigger a rolling update of the whole cluster such as :

- `spec.resources`
- `spec.baseImage`
- `spec.version`
- `spec.configMap`
- `spec.gcStdout`
- `spec.runAsUser`

Some Updates in the `CassandraCluster` CRD object will not trigger change on the cluster but only in future behavior of
CassKop :

- `spec.autoPilot`
- `spec.autoUpdateSeedList`
- `spec.deletePVC`
- `spec.hardAntiAffinity`
- `spec.rollingPartition`
- `spec.maxPodUnavailable`
- `noCheckStsAreEqual`

CassKop manages rolling updates for each statefulset in the cluster. Then each statefulset is making the rolling
updated of it's pod according to the `partition` defined for each statefulset in
the `spec.topology.dc[].rack[].rollingPartition`.

## Initializing

The First Operation required in a Cassandra Cluster is the initialization.

In this Phase, the CassKop will create the `CassandraCluster.Status` section with an entry for each DC/Rack declared
in the `CassandraCluster.spec.topology` section.

We could also have Initializing status if we decided later to add some DC to our topology.

### With no `topology` defined

For demo we will create this CassandraCluster without topology section

```yaml
apiVersion: "db.orange.com/v1alpha1"
kind: "CassandraCluster"
metadata:
  name: cassandra-demo
  labels:
    cluster: k8s.pic
spec:
  nodesPerRacks: 2
  baseImage: cassandra
  version: latest
  rollingPartition: 0
  dataCapacity: "3Gi"
  dataStorageClass: "local-storage"
  hardAntiAffinity: false
  deletePVC: true
  autoPilot: false
  gcStdout: true
  autoUpdateSeedList: true
  resources:
    requests:
      cpu: '2'
      memory: 2Gi
    limits:
      cpu: '2'
      memory: 2Gi
```

> If no `topology` has been specified, then CassKop creates the default topology and status.

The default topology added by CassKop is :

```yaml
...
  topology:
    dc:
    - name: dc1
      rack:
      - name: rack1
```

The number of cassandra nodes `CassandraCluster.spec.nodesPerRacks` defines the number of cassandra nodes CassKop
must create in each of it's racks. In our example, there is only one default rack, so CassKop will only create 2
nodes.

>**IMPORTANT:** with the default topology there will be no Kubernetes NodesAffinity to spread the Cassandra nodes on the
>cluster. In this case, CassKop will only create one Rack and one DC for Cassandra. It is not recommended as you may
>lose data in case of hardware failure

When Initialization has ended you should have a Status similar to :

```yaml
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: Initializing
        endTime: 2018-09-18T15:10:51Z
        status: Done
      phase: Running
      podLastOperation: {}
  lastClusterAction: Initializing
  lastClusterActionStatus: Done
  phase: Running
  seedlist:
  - cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.cassandra-test
  - cassandra-demo-dc1-rack1-1.cassandra-demo-dc1-rack1.cassandra-test
```

- The Status of the `dc1-rack1` is `Initializing=Done`
- The Status of the Cluster is `Initializing=Done`
- The phase is `Running` which means that each Rack has the desired amount of Nodes.

We asked 2 `nodesPerRacks` and we have one default rack, so we ended with 2 Cassandra nodes in our cluster.

The Cassandra `seedlist` has been initialized and stored in the CassandraCluster.status.seedlist`. It has also been
configured in each of the Cassandra Pods.

We can also confirm that Cassandra knows about the DC and Rack name we have deployed :

```console
$ kubectl exec -ti cassandra-demo-dc1-rack1-0 nodetool status
Datacenter: dc1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.18.72.28   65.86 KiB  32           100.0%            fdc1a9e9-c5c3-4169-ae47-e6843efa096d  rack1
UN  172.18.120.12  65.86 KiB  32           100.0%            509ca725-fbf9-422f-a8e0-5e2a55474f70  rack1
```

### With `topology` defined

In this example, I added a topology defining 2 Cassandra DC and 3 racks in total

```yaml
apiVersion: "db.orange.com/v1alpha1"
kind: "CassandraCluster"
metadata:
  name: cassandra-demo
  labels:
    cluster: k8s.pic
spec:
  nodesPerRacks: 2
  baseImage: cassandra
  version: latest
  rollingPartition: 0
  dataCapacity: "3Gi"
  dataStorageClass: "local-storage"
  hardAntiAffinity: false
  deletePVC: true
  autoPilot: true
  autoUpdateSeedList: true
  resources:
    requests:
      cpu: '2'
      memory: 2Gi
    limits:
      cpu: '2'
      memory: 2Gi
  topology:
    dc:
      - name: dc1
        labels:
          failure-domain.beta.kubernetes.io/region: europe-west1
        rack:
          - name: rack1
            labels:
              failure-domain.beta.kubernetes.io/zone: europe-west1-b
          - name: rack2
            labels:
              failure-domain.beta.kubernetes.io/zone: europe-west1-c
      - name: dc2
        nodesPerRacks: 3
        numTokens: 32
        labels:
          failure-domain.beta.kubernetes.io/region: europe-west1
        rack:
          - name: rack1
            labels:
              failure-domain.beta.kubernetes.io/zone: europe-west1-d
```

With this topology section I also references some **Kubernetes nodes labels**
which will be used to spread the Cassandra
nodes on each Racks on different groups of Kubernetes servers.

> We can see here that we can give specific configuration for the number of pods in the dc2 (`nodesPerRacks: 3`)
> We also allow to configure Cassandra pods with different num_tokens confioguration for each dc : `numTokens`.

CassKop will create a statefulset for each Rack, and start creating the
Cassandra Cluster, starting by nodes from the Rack 1.
When CassKop will end operations on Rack1, it will process the next rack and so on.

The status may be similar to :

```yaml
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: Initializing
        status: Ongoing
      phase: Initializing
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: Initializing
        status: Ongoing
      phase: Initializing
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: Initializing
        status: Ongoing
      phase: Initializing
      podLastOperation: {}
  lastClusterAction: Initializing
  lastClusterActionStatus: Ongoing
  phase: Initializing
  seedlist:
  - cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.cassandra-test
  - cassandra-demo-dc1-rack1-1.cassandra-demo-dc1-rack1.cassandra-test
  - cassandra-demo-dc1-rack2-0.cassandra-demo-dc1-rack2.cassandra-test
  - cassandra-demo-dc2-rack1-0.cassandra-demo-dc2-rack1.cassandra-test
  - cassandra-demo-dc2-rack1-1.cassandra-demo-dc2-rack1.cassandra-test
  - cassandra-demo-dc2-rack1-2.cassandra-demo-dc2-rack1.cassandra-test
```

The creation of the cluster is ongoing.
We can see that, regarding the Cluster Topology, CassKop has created the SeedList.

> CassKop compute a seedlist with 3 nodes in each datacenter (if possible). The Cassandra seeds are always the
> first Cassandra nodes of a statefulset (starting with index 0).

When all racks are in status done, then the `CassandraCluster.status.lastClusterActionStatus` is changed to `Done`.

We can see that internally Cassandra also knows the desired topology :

```
k exec -ti cassandra-demo-dc1-rack1-0 nodetool status
Datacenter: dc1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.18.112.6   126.05 KiB  32           38.5%             1512da3c-f6b2-469f-95d1-2d060043777a  rack1
UN  172.18.64.10   137.08 KiB  32           32.0%             8149054f-4bc3-4093-a6ef-80910c018122  rack2
UN  172.18.88.9    154.54 KiB  32           30.2%             dbe44aa6-6763-4bc1-825a-9ea7d21690e3  rack2
UN  172.18.120.15  119.88 KiB  32           33.7%             c87e858d-66a8-4544-9d28-718a1f94955b  rack1
Datacenter: dc2
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.18.72.8    119.65 KiB  32           26.9%             8688abd3-08b6-44e0-8805-05bd3650eea6  rack1
UN  172.18.104.8   153.08 KiB  32           38.8%             62adf02d-8c55-4d95-a459-45b1c9c3aa91  rack1
```

## UpdateConfigMap

You can find in the [cassandra-configuration](../documentation/description.md#cassandra-configuration) section how you can use
the `spec.configMap` parameter.

>**IMPORTANT:** actually CassKop doesn't monitor changes inside the ConfigMap. If you want to change a parameter in a
>file in the current configMap, you must create a new configMap with the updated version, and then ask CassKop to use
>the new configmap name.

If we add/change/remove the `CassandraCluster.spec.configMapName` then CassKop will start a RollingUpdate of each
CassandraNodes in each Racks, starting from the first Rack defined in the `topology`.

```yaml
apiVersion: "db.orange.com/v1alpha1"
kind: "CassandraCluster"
metadata:
  name: cassandra-demo
  labels:
    cluster: k8s.pic
spec:
  nodesPerRacks: 2
  baseImage: cassandra
  version: latest
  rollingPartition: 0
  dataCapacity: "3Gi"
  dataStorageClass: "local-storage"
  hardAntiAffinity: false
  deletePVC: true
  autoPilot: true
  autoUpdateSeedList: true
  configMapName: cassandra-configmap-v1
  ...
```

First we need to create the configmap exemple:

```console
kubectl apply -f samples/cassandra-configmap-v1.yaml
```

Then we apply the changes in the `CassandraCluster`.

We can see the `CassandraCluster.Status` updated by CassKop

```yaml
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: UpdateConfigMap
        startTime: 2018-09-21T12:24:24Z
        status: Ongoing
      phase: Pending
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: Initializing
        endTime: 2018-09-21T10:33:10Z
        status: Done
      phase: Running
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: Initializing
        endTime: 2018-09-21T10:34:47Z
        status: Done
      phase: Running
      podLastOperation: {}
  lastClusterAction: UpdateConfigMap
  lastClusterActionStatus: Ongoing
```

>**Note:** CassKop won't make a rolling update on the next rack until the status of the current rack becomes`Done`.
>The Operation is processing "rack per rack".

```yaml
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: UpdateConfigMap
        endTime: 2018-09-21T12:26:10Z
        startTime: 2018-09-21T12:24:24Z
        status: Done
      phase: Running
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: UpdateConfigMap
        endTime: 2018-09-21T12:27:25Z
        startTime: 2018-09-21T12:26:10Z
        status: Done
      phase: Running
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: UpdateConfigMap
        startTime: 2018-09-21T12:27:27Z
        status: Ongoing
      phase: Pending
      podLastOperation: {}
  lastClusterAction: UpdateConfigMap
  lastClusterActionStatus: Ongoing
```

## UpdateDockerImage

CassKop allows you to change the Cassandra docker image and gracefully redeploy your whole cluster. 

If we change the `CassandraCluster.spec.baseImage` and or `CassandraCluster.spec.version`  CassKop will start to
perform a RollingUpdate on the whole cluster (for each racks sequentially, in order to change the version of the
Cassandra Docker Image on all nodes.

> See section [Cassandra docker image](../documentation/description.md#cassandra-docker-image)

You can change the docker image used to :

- change the version of Cassandra
- change the version of Java
- Change some configuration parameters for cassandra or jvm if you don't overwrite them with a ConfigMap

The status may be similar to:

```yaml
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: UpdateDockerImage
        startTime: 2018-09-18T16:08:59Z
        status: Ongoing
      phase: Pending
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: Initializing
        endTime: 2018-09-18T16:05:51Z
        status: Done
      phase: Running
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: Initializing
        endTime: 2018-09-18T16:07:52Z
        status: Done
      phase: Running
      podLastOperation: {}
  lastClusterAction: UpdateDockerImage
  lastClusterActionStatus: Ongoing
  phase: Pending
  seedlist:
  - cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.cassandra-test
  - cassandra-demo-dc1-rack1-1.cassandra-demo-dc1-rack1.cassandra-test
  - cassandra-demo-dc1-rack2-0.cassandra-demo-dc1-rack2.cassandra-test
  - cassandra-demo-dc2-rack1-0.cassandra-demo-dc2-rack1.cassandra-test
  - cassandra-demo-dc2-rack1-1.cassandra-demo-dc2-rack1.cassandra-test
  - cassandra-demo-dc2-rack1-2.cassandra-demo-dc2-rack1.cassandra-test
```

We can see that CassKop has started to Update the `dc1-rack1` and it has changed the `lastClusterAction` and
`lastClusterStatus` accordingly.

Once it has finished the first rack, then it processes the next one:

```yaml
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: UpdateDockerImage
        endTime: 2018-09-18T16:10:51Z
        startTime: 2018-09-18T16:08:59Z
        status: Done
      phase: Running
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: UpdateDockerImage
        startTime: 2018-09-18T16:10:51Z
        status: Ongoing
      phase: Pending
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: Initializing
        endTime: 2018-09-18T16:07:52Z
        status: Done
      phase: Running
      podLastOperation: {}
  lastClusterAction: UpdateDockerImage
  lastClusterActionStatus: Ongoing
```

And when all racks are Done:

```yaml
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: UpdateDockerImage
        endTime: 2018-09-18T16:10:51Z
        startTime: 2018-09-18T16:08:59Z
        status: Done
      phase: Running
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: UpdateDockerImage
        endTime: 2018-09-18T16:12:42Z
        startTime: 2018-09-18T16:10:51Z
        status: Done
      phase: Running
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: UpdateDockerImage
        endTime: 2018-09-18T16:14:52Z
        startTime: 2018-09-18T16:12:42Z
        status: Done
      phase: Running
      podLastOperation: {}
  lastClusterAction: UpdateDockerImage
  lastClusterActionStatus: Done
  phase: Running
```

This provides a Central view to monitor what is happening on the Cassandra Cluster.

## UpdateResources

CassKop allows you to configure your Cassandra's pods resources (memory and cpu). 

If we change the `CassandraCluster.spec.resources`, then CassKop will start to make a RollingUpdate on the whole
cluster (for each racks sequentially) to change the version of the Cassandra Docker Image on all nodes.

> See section [Resource limits and requets](../documentation/description.md#resource-limits-and-requests)

For example, to increase Memory/CPU requests and/or limits:

```yaml
    requests:
      cpu: '2'
      memory: 3Gi
    limits:
      cpu: '2'
      memory: 3Gi
```

Then CassKop should output the status: 

```yaml
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: UpdateResources
        startTime: 2018-09-21T15:28:43Z
        status: Ongoing
      phase: Pending
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: UpdateResources
        startTime: 2018-09-21T15:28:43Z
        status: ToDo
      phase: Running
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: UpdateResources
        startTime: 2018-09-21T15:28:43Z
        status: ToDo
      phase: Running
      podLastOperation: {}
  lastClusterAction: UpdateResources
  lastClusterActionStatus: Ongoing
```

We can see that it has staged the `UpdateResources` action in all racks (`status=ToDo`) and has started the action in
the first rack (`status=Ongoing`). Once `Done` it will follow with next rack, and so on.

Upon completion, the status may look like :

```yaml
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: UpdateResources
        endTime: 2018-09-21T15:30:31Z
        startTime: 2018-09-21T15:28:43Z
        status: Done
      phase: Running
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: UpdateResources
        endTime: 2018-09-21T15:32:12Z
        startTime: 2018-09-21T15:30:32Z
        status: Done
      phase: Running
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: UpdateResources
        endTime: 2018-09-21T15:34:07Z
        startTime: 2018-09-21T15:32:13Z
        status: Done
      phase: Running
      podLastOperation: {}
  lastClusterAction: UpdateResources
  lastClusterActionStatus: Done
```

## Scaling the cluster

The Scaling of the Cluster is managed through the nodesPerRacks parameters and through the number of Dcs and Racks
defined in the Topology section.

See section [NodesPerRacks](/docs/casskop/2_deployment_configuration/2_cassandra_config#nodesperracks)

> **NOTE:** if the ScaleUp (or the ScaleDown) may change the SeedList and if `spec.autoUpdateSeedList` is set to `true`
> then CassKop will program a new operation : `UpdateSeedList` which will trigger a rollingUpdate to apply the new
> seedlist on all nodes, once the Scaling is done.

### ScaleUp

CassKop allows you to Scale Up your Cassandra cluster.

There is a global parameter `CassandraCluster.spec.nodesPerRacks` which specify the number of Cassandra nodes we want in
a rack.

It is possible to surcharge this for a particular DC in the `CassandraCluster.spec.topology.dc[<idx>].nodesPerRacks`

Example:

```yaml
  topology:
    dc:
      - name: dc1
        rack:
          - name: rack1
          - name: rack2
      - name: dc2
        nodesPerRacks: 3        <--- We increase by one this value
        rack:
          - name: rack1
```

In this case, we ask to ScaleUp nodes of second DC `dc2`

CassKop takes into account the new target, and starts applying modifications in the cluster :

```yaml
...
    dc1-rack1:
      cassandraLastAction:
        Name: UpdateSeedList
        status: Configuring
      phase: Running
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: UpdateSeedList
        status: Configuring
      phase: Running
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: ScaleUp
        startTime: 2018-09-27T15:02:21Z
        status: Configuring
      phase: Pending
  lastClusterAction: ScaleUp
  lastClusterActionStatus: Ongoing
 ...
 ```

We can see that CassKop:

- Has started the `ScaleUp` action in `dc2-rack1`
- Has found that the SeedList must be updated, and because the autoUpdateSeedList=true it has staged
  (`status=Configuring`) the UpdateSeedList operation for `dc1-rack1` and `dc1-rack2`

When CassKop ends the ScaleUp action in the `dc2-rack1` then it will also stage this rack with `UpdateSeedList=Configuring`.
Once all racks are in this state, CassKop will turn each Rack in status `UpdateSeedList=ToDo`, meaning that it can
start the operation. 

Starting from then, CassKop will iterate on each rack one after the other and get status :

- `UpdateSeedList=Ongoing` meaning that it is currently doing a rolling update on the Rack to update the SeedList seting
  also sets the `startTime`.
- `UpdateSeedList=Done` meaning that the operation is done. (then, it sets the `endTime`)

See evolution of status:

```yaml
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: UpdateSeedList
        endTime: 2018-09-27T15:05:00Z
        startTime: 2018-09-27T15:03:13Z
        status: Done
      phase: Running
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: UpdateSeedList
        startTime: 2018-09-27T15:03:13Z
        status: Ongoing
      phase: Pending
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: UpdateSeedList
        status: ToDo
      phase: Running
  lastClusterAction: UpdateSeedList
  lastClusterActionStatus: Finalizing
  phase: Pending
```

Here is the final topology seen from nodetool :

```console
$ k exec -ti cassandra-demo-dc1-rack1-0 nodetool status
Datacenter: dc1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.18.88.6    211.95 KiB  32           27.4%             dbe44aa6-6763-4bc1-825a-9ea7d21690e3  rack2
UN  172.18.112.5   231.49 KiB  32           29.2%             1512da3c-f6b2-469f-95d1-2d060043777a  rack1
UN  172.18.64.10   188.36 KiB  32           27.6%             8149054f-4bc3-4093-a6ef-80910c018122  rack2
UN  172.18.120.14  237.62 KiB  32           29.8%             c87e858d-66a8-4544-9d28-718a1f94955b  rack1
Datacenter: dc2
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.18.72.8    228.54 KiB  32           22.0%             8688abd3-08b6-44e0-8805-05bd3650eea6  rack1
UN  172.18.104.8   212.34 KiB  32           32.2%             62adf02d-8c55-4d95-a459-45b1c9c3aa91  rack1
UN  172.18.88.9    148.34 KiB  32           31.7%             fecdfb5d-3ad4-4204-8ca5-cc7f1c4c19c4  rack1
```

Note that nodetool prints IP of nodes while kubernetes works with names :

```console
$ k get pods -o wide -l app=cassandracluster
NAME                         READY     STATUS    RESTARTS   AGE       IP              NODE      NOMINATED NODE
cassandra-demo-dc1-rack1-0   1/1       Running   0          14m       172.18.112.5    node006   <none>
cassandra-demo-dc1-rack1-1   1/1       Running   0          15m       172.18.120.14   node003   <none>
cassandra-demo-dc1-rack2-0   1/1       Running   0          13m       172.18.88.6     node005   <none>
cassandra-demo-dc1-rack2-1   1/1       Running   0          13m       172.18.64.10    node004   <none>
cassandra-demo-dc2-rack1-0   1/1       Running   0          10m       172.18.72.8     node008   <none>
cassandra-demo-dc2-rack1-1   1/1       Running   0          11m       172.18.104.8    node007   <none>
cassandra-demo-dc2-rack1-2   1/1       Running   0          12m       172.18.88.9     node005   <none>
```

After the ScaleUp has finished, CassKop must execute a cassandra `cleanup` on each nodes of the Cluster.
This can be manually triggered by setting appropriate labels on each Pods.

CassKop can automate this if `spec.autoPilot` is true by setting the labels on each Pods of the cluster with a ToDo
state and then find thoses pods to sequentially execute thoses actions.

See podOperation [Cleanup](##operationcleanup)!!

## UpdateScaleDown

For ScaleDown, CassKop must perform a clean cassandra `decommission` prior to actually scale down the cluster at
Kubernetes level.

Actually, this is done through CassKop asking the decommission through a jolokia call and waiting for it to be
performed (cassandra node status = decommissionned) before updating kubernetes statefulset (removing the pod).

> **IMPORTANT**: If we ask to scale down more than 1 node at a time, then CassKop will iterate on a single scale down
> until it reaches the requested number of nodes.
> Also CassKop will refuse a scaledown to 0 for a DC if there still have some data replicated to it.

To launch a ScaleDown, we simply need to decrease the value of nodesPerRacks.

```yaml
  topology:
    dc:
      - name: dc1
        rack:
          - name: rack1
          - name: rack2
      - name: dc2
        nodesPerRacks: 2        <--- Get back to 2
        rack:
          - name: rack1
```

We can see in the below example that:

- It has started the `ScaleDown` action in `dc2-rack1`
- CassKop has found that the SeedList must be updated, and it has staged (`status=ToDo`) it for `dc1-rack1` and
  `dc1-rack2`

When CassKop completes the ScaleDown in the `dc2-rack1` then it will stage it also with `UpdateSeedList=ToDo` Once all
racks are in this state, CassKop will turn each Rack in status `UpdateSeedList=Ongoing` meaning that it can start the
operation, it also set the `startTime`

Then, CassKop will iterate on each rack one after the other and get status :

- `UpdateSeedList=Finalizing` meaning that it is currently doing a rolling update on the Rack to update the SeedList
- `UpdateSeedList=Done` meaning that the operation is done. Then, it sets the `endTime`.

```yaml
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: UpdateSeedList
        status: ToDo
      phase: Running
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: UpdateSeedList
        status: ToDo
      phase: Running
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: ScaleDown
        startTime: 2018-09-27T15:22:23Z
        status: Ongoing
      phase: Running
      podLastOperation:
        Name: decommission
        pods:
        - cassandra-demo-dc2-rack1-2
        startTime: 2018-09-27T15:22:23Z
        status: Ongoing
  lastClusterAction: ScaleDown
  lastClusterActionStatus: Ongoing
 ```

When `ScaleDown=Done` CassKop will start the UpdateSeedList operation.

```yaml
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: UpdateSeedList
        startTime: 2018-09-27T15:23:54Z
        status: Finalizing
      phase: Pending
      podLastOperation: {}
    dc1-rack2:
      cassandraLastAction:
        Name: UpdateSeedList
        startTime: 2018-09-27T15:23:54Z
        status: Ongoing
      phase: Running
      podLastOperation: {}
    dc2-rack1:
      cassandraLastAction:
        Name: UpdateSeedList
        startTime: 2018-09-27T15:23:54Z
        status: Ongoing
      phase: Running
      podLastOperation:
        Name: decommission
        endTime: 2018-09-27T15:23:51Z
        podsOK:
        - cassandra-demo-dc2-rack1-2
        startTime: 2018-09-27T15:22:23Z
        status: Done
  lastClusterAction: UpdateSeedList
  lastClusterActionStatus: Finalizing
  phase: Pending
```

It shows also that `podLastOperation` `decommission` is `Done`. CassKop will then rollingUpdate all racks one by one
in order to update the Cassandra seedlist.

## UpdateSeedList

The UpdateSeedList is done automatically by CassKop when the parameter
`CassandraCluster.spec.autoUpdateSeedList` is true (default).

See [ScaleUp](#updatescaleup) and [ScaleDown](#updatescaledown).

## CorrectCRDConfig

The CRD `CassandraCluster` is used to define your cluster configuration. Some fields can't be updated in a kubernetes
clusters. Some fields are taken from the CRD to configure thoses objects, and to be sure we don't update them (to
prevent kubernetes objects in errors), we have configure CassKop to simply ignore/revert unauthorized changed to the
CRD.

Example With this CRD deployed :

```yaml
spec:
  nodesPerRacks: 2
  baseImage: cassandra
  version: latest
  imagePullSecret:
    name: advisedev # To authenticate on docker registry
  rollingPartition: 0
  dataCapacity: "3Gi"                  <-- can't be changed
  dataStorageClass: "local-storage"    <-- can't be changed
  hardAntiAffinity: false
  deletePVC: true
  autoPilot: true
  autoUpdateSeedList: true
```

If we try to update the `dataCapacity` or `dataStorageClass` nothing will happen. And we could see thoses messages in
the logs of CassKop :

```console
time="2018-09-27T17:44:13+02:00" level=warning msg="[cassandra-demo]: CassKop has refused the changed on DataCapacity from [3Gi] to NewValue[4Gi]"
time="2018-09-27T17:44:35+02:00" level=warning msg="[cassandra-demo]: CassKop has refused the changed on DataStorageClass from [local-storage] to NewValue[local-storag]"
```

If you performed the modification by updating your local CRD file and apply it with kubectl you must revert to the old
value.

## Delete a DC

- Prior to delete a DC, you must have ScaleDown to 0 all the Racks, if not, CassKop will refuse and correct the CRD.
- Prior to scaleDown to 0 CassKop will ensure that there are no more data replicated to the DC, if not, CassKop
  will refuse and correct the CRD.
Because CassKop wants that we have the same amounts of pods in all racks, we decided that we would't allow to remove
  only a rack. This will be revert too.
  
> You must ScaleDown to 0 priori to Remoove a DC
> You must change replication factor prior to ScaleDown  to 0 a DC
  
## Kubernetes node maintenance operation

In a normal production environment, CassKop will have spread it's Cassandra pods on differents k8s nodes. If the team
in charge of the machines needs to make some operations on a host they can make a drain.

The Kubernetes drain command will ask the scheduler to make an eviction for all pods on the current nodes, and for many
workloads k8s will reschedule them on other machines. In the case of CassKop cassandra pods, they won't be scheduled
on another host, because they uses local-storage and are stick to a specific host thanks to the PersistentVolumeClaim
kubernetes object.

Example: we drain the node008 for a maintenance operation.

```sh
$ kubectl drainnode node008 --ignore-daemonsets --delete-local-data
```

All pods will be evicted, thoses who can will be rescheduled on another hosts.
Our Cassandra pod won't we able to be schedule elsewhere due to the PVS, and we can see this messages in the k8s events :

```console
0s    Warning   FailedScheduling   Pod   0/8 nodes are available: 1 node(s) were unschedulable, 2 node(s) had taints
that the pod didn't tolerate, 5 node(s) had volume node affinity conflict.
```

It explain that 1 node is unshedulable, this is the one we just drain. the 5 other nodes can't be scheduled by our pod
because they have volume node affinity conflict ()our pods have an affinity on node008).

Once the team have finished their maintenance operation they can bring back the host into the kubernetes cluster. From
then, k8s will be able to reshedule back the cassandra pod into the cluster so that it can re-join the ring.

```console
$ kubectl uncordon node008
node/node008 uncordoned
```

Immediately the pending pod is rescheduled and started on the host.
If the time of interruption was not too long there is nothing more to do, the node will join the ring and re-synchronise
with the cluster. If the time was too long, then it may be needed to schedule some PodOperations that you will find in
nexts sections of this document.

### The PodDisruptionBudget (PDB) protection

If a k8s admin ask to drain a node, this may not been allowed by the cassandracluster regarding it's current state and
the configuration of its PDB (usually only 1 nodes allowed to be in disruption).

Example :

```console
$ kubectl drainnode node008 --ignore-daemonsets --delete-local-data
error when evicting pod "cassandra-demo-dc2-rack1-0" (will retry after 5s): Cannot evict pod as it would violate the pod's disruption budget.
```

The node008 will be flagged as SchedulingDisabled, so that it won't take new workload. It will evict all possible pods,
but if there was an ongoing disruption on the current Cassandra cluster, it won't be allowed to evict the cassandra pod.

Example of a PDB :

```yaml
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  creationTimestamp: 2019-02-13T15:17:05Z
  generation: 1
  labels:
    app: cassandracluster
    cassandracluster: cassandra-test
    cluster: k8s.pic
  name: cassandra-test
  namespace: cassandra-test
  ownerReferences:
  - apiVersion: db.orange.com/v1alpha1
    controller: true
    kind: CassandraCluster
    name: cassandra-test
    uid: 45fc4a22-2fa2-11e9-8df0-009c0296dbc4
  resourceVersion: "12093573"
  selfLink: /apis/policy/v1beta1/namespaces/cassandra-test/poddisruptionbudgets/cassandra-test
  uid: 6bc1bf12-2fa2-11e9-aea5-009c0296e48e
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: cassandracluster
      cassandracluster: cassandra-test
      cluster: k8s.pic
status:
  currentHealthy: 13
  desiredHealthy: 13
  disruptedPods: null
  disruptionsAllowed: 0
  expectedPods: 14
  observedGeneration: 1
```

In this example we see that we allowed only 1 Pod unavailable, and on our cluster we wants to have 14 pods and we only
have 13 healthy, that's why the PDB won't allow the eviction of an additional pod.

To be able to continue, we need to wait or to make appropriate actions so that the Cassandra cluster won't have any
unavailable nodes.

## K8S host major failure: replacing a cassandra node

In the case of a major host failure, it may not be possible to bring back the node to life. We can in this case
consider that our cassandra node is lost and we will want to replace it on another host.

In this case we may have 2 solutions that will require some manual actions :

### Remove old node and create new one

1. In this case we will use CassKop client to schedule a cassandra removenode for the failing node.

```console
$ kubectl casskop removenode --pod <node-id>
```

This will trigger the PodOperation removenode by setting the appropriate labels on a cassandra Pod.

2. Once the node is properly removed, we can free the link between the Pod and the failing host by removing the
   associated PodDisruptionBudget
   
```console
$ kubectl delete pvc data-cassandra-test-dc1-rack2-1
```

This will allow Kubernetes to reschedule the Pod on another free host.

3. Once the node is back in the cluster we need to apply a cleanup on all nodes:

```console
$ kubectl casskop cleanup start
```

you can pause the cleanup and check status with:

```console
$ kubectl casskop cleanup pause
$ kubectl casskop cleanup status
```

### Replace node with a new one

In some cases It may be useful to prefer to replace the node. Because we use a statefulset to deploy cassandra pods,
by definition all pods are identical and we couldn't execute specific actions on a specific node at startup.

For that CassKop provide the ability to execute a `pre_run.sh` script that can be change using the CRD ConfigMap.

To see how to use the configmap see [Overriding Configuration using
configMap](/docs/casskop/2_deployment_configuration/2_cassandra_config#configuration-override-using-configmap)

for example If we want to replace the node cassandra-test-dc1-rack2-1, we first need to retrieve it's IP address from
nodetool status for example :

```console
$ nodetool status
Datacenter: dc1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address         Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.31.184.189  35.18 GiB  256          51.2%             9455a9bd-7a61-405e-8c3f-ee1f72f63500  rack1
UN  172.31.180.138  37 GiB     256          51.0%             1ad1b4b7-c719-4683-8109-31aa9722c1ee  rack2
UN  172.31.179.248  37.86 GiB  256          47.4%             69cbf178-2477-4420-ac71-6fad10f93759  rack2
UN  172.31.182.120  41.76 GiB  256          50.2%             a4ffac86-990d-4487-80a0-b2e177d8e06e  rack1
DN  172.31.183.213  31.14 GiB  256          51.9%             e45107ba-fe7b-4904-98cf-1373d1946bb5  rack2
UN  172.31.181.193  33.15 GiB  256          48.4%             35806f73-17fb-4d91-b2e7-8333f393189b  rack1
```

Then we can edit the ConfigMap to edit the pre_run.sh script :

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cassandra-configmap-pre-run
data:
  pre_run.sh: |-
    echo "** this is a pre-scrip for run.sh that can be edit with configmap"
    test "$(hostname)" == 'cassandra-demo-dc1-rack3-0' && echo "-Dcassandra.replace_address_first_boot=172.31.183.213" > /etc/cassandra/jvm.options
    echo "** end of pre_run.sh script, continue with run.sh"
```

So the Operation will be :

1. Edit the configmap with the appropriate CASSANDRA_REPLACE_NODE IP for the targeted pod name
2. delete the pvc data-cassandra-test-dc1-rack2-1
3. the Pod will boot, execute the pre_run.sh script prior to the /run.sh
4. the new pod replace the dead one by re-syncing the content which could take some times depending on the data size.
5. Do not forget to edit again the ConfigMap and to remove the specific line with replace_node instructions.
