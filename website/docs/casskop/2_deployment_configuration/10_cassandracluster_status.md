---
id: 10_cassandracluster_status
title: Cassandra cluster status
sidebar_label: Cassandra cluster status
---

You can request kubernetes Object `cassandracluster` representing the Cassandra cluster to retrieve information about
it's status :

```yaml
$ kubectl describe cassandracluster cassandra
...
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: ScaleUp
        endTime: 2018-07-12T14:10:28Z
        startTime: 2018-07-12T14:09:34Z
        status: Done
      phase: Running
      podLastOperation:
        Name: cleanup
        endTime: 2018-07-12T14:07:35Z
        podsOK:
        - cassandra-demo-dc1-rack1-0
        - cassandra-demo-dc1-rack1-1
        - cassandra-demo-dc1-rack1-2
        startTime: 2018-07-12T14:06:22Z
        status: Done
    dc1-rack2:
      cassandraLastAction:
        Name: ScaleUp
        endTime: 2018-07-12T14:10:58Z
        startTime: 2018-07-12T14:10:28Z
        status: Done
      phase: Running
      podLastOperation:
        Name: cleanup
        endTime: 2018-07-12T14:08:16Z
        podsOK:
        - cassandra-demo-dc1-rack2-0
        - cassandra-demo-dc1-rack2-1
        - cassandra-demo-dc1-rack2-2
        startTime: 2018-07-12T14:08:09Z
        status: Done
  lastClusterAction: ScaleUp
  lastClusterActionStatus: Done        
...
  phase: Running
  seedlist:
  - cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.cassandra-demo.svc.kaas-prod-priv-sph
```

The CassandraCluster prints out it's whole status.

- **seedlist**: it is the Cassandra SEED List used in the Cluster.
- **Phase** : it's the global state for the cassandra cluster which can have different values :
    - **Initialization**, we just launched a new cluster, and waiting for its requested state
    - **Running**, the cluster is running normally
    - **Pending**, the number of Nodes requested has changed, waiting for reconciliation
- **lastClusterAction** Is the Last Action at the Cluster level
- **lastClusterActionStatus** Is the Last Action Status at the Cluster level
- **CassandraRackStatus** represents a map of statuses for each of the Cassandra Racks in the Cluster
  - **{Cassandra DC-Rack Name}**
    - **Cassandra Last Action**: it's an action which is ongoing on the Cassandra cluster :
        - **Name**: name of the Action
            - **UpdateConfigMap** a new ConfigMap has been submitted to the cluster
            - **UpdateDockerImage** a new Docker Image has been submitted to the cluster
            - **UpdateSeedList** a new SeedList must be deployed on the cluster
            - **UpdateResources** CassKop must apply new resources values for it's statefulsets            
            - **RollingRestart** CassKop performs a rollingrestart on the target statefulset
            - **ScaleUp** a scale Up has been requested
            - **ScaleDown** a scale Down has been requested.
            - **UpdateStatefulset** a change has been submitted to the statefulset, but CassKop doesn't know exactly
              which one.              
        - **Status**: status of the Action
            - **Configuring**: Only used for UpdateSeedList, we need to synchronise all statefulset with this operation before starting it
            - **ToDo**: an action is scheduled
            - **Ongoing**: an action is ongoing, see Start Time
            - **Continue**: the action may be continuing (used for ScaleDown)
            - **Done**: the action is Done, see End Time
        - **Start Time**: time of start of the operation
        - **End Time**: time of end of the operation
    - **Pod Last Operation**: it's an operation done at Pod Level
        - **Name**: Name of the Operation
            - **decommissioning**: a nodetool decommissioning must be performed on a pod
            - **cleanup**: a nodetool cleanup must be performed on a pod
            - **rebuild**: a nodetool rebuild must be performed on a pod
            - **upgradesstables**: a nodetool upgradesstables must be performed on a pod            
        - **Status**:
            - **Manual**: an operation is recommended to be scheduled by a human
            - **ToDo**: an operation is scheduled    
            - **Ongoing**: an operation is ongoing, see start time
            - **Done**: an operation is done, see end time
        - **Pods**: list of Pods on which the operation is ongoing
        - **PodsOK**: list of Pods on which the operation is done
        - **PodsKO**: list of Pods on which the operation has not been completed correctly
        - **Start Time**: time of start for an operation
        - **End Time**: time of end for an operation        
  
> When Status=Done for each Racks, then there is no specific action ongoing on the cluster and the
> lastClusterActionStatus will turn also to Done.