---
id: 3_cassandra_pods_operations
title: Cassandra pods operations
sidebar_label: Cassandra pods operations
---

Some Pods Operations can be triggered automatically by CassKop if :

- `CassandraCluster.spec.autoPilot` is true, that will trigger `cleanup`, `rebuild` and `upgadesstable` operation in
  response to cluster events automatically.
- the `decommission operation` is special and will be triggered automatically each time we need to ScaleDown a Pod.
- the `removenode operation` is also special and may be set manually when needed.

It is also possible to trigger operations "manually", setting some labels on the Pods.

## OperationCleanup

A Cleanup may be automatically triggered by CassKop when it ends Scaling the cluster.
CassKop will set some specific labels on the targeted pods.
We can also set these labels manually, or using the privided plugin (`kubectl casskop cleanup start`)
If we want to see labels for each of the pods of the cluster :

```console
$ kubectl label pod $(kubectl get pods -l app=cassandracluster -o jsonpath='{range .items[*]}{.metadata.name}{" "}') --list
Listing labels for Pod./cassandra-demo-dc1-rack1-0:
 cluster=k8s.pic
 controller-revision-hash=cassandra-demo-dc1-rack1-56c9bbb958
 dc-rack=dc1-rack1
 statefulset.kubernetes.io/pod-name=cassandra-demo-dc1-rack1-0
 app=cassandracluster
 cassandracluster=cassandra-demo
 cassandraclusters.db.orange.com.dc=dc1
 cassandraclusters.db.orange.com.rack=rack1
...
```

Now, to trigger a `cleanup` on pod `cassandra-demo-dc1-rack2-0`

```console
kubectl label pod cassandra-demo-dc1-rack2-0 operation-name=cleanup --overwrite
kubectl label pod cassandra-demo-dc1-rack2-0 operation-status=ToDo --overwrite
```

Automatically, CassKop will detect the change, start the action, and update the status :

```yaml
    dc1-rack2:
      cassandraLastAction:
        Name: UpdateSeedList
        endTime: 2018-09-27T15:26:51Z
        startTime: 2018-09-27T15:23:54Z
        status: Done
      phase: Running
      podLastOperation:
        Name: cleanup
        endTime: 2018-09-27T16:00:52Z
        operatorName: operator-cassandr-f6d2968d4504448180ace041d3818d10-799dbb4zqss8
        podsOK:
        - cassandra-demo-dc1-rack2-0
        - cassandra-demo-dc1-rack2-0
        startTime: 2018-09-27T16:00:32Z
        status: Done
```

The section `podLastOperation` appears and we can see that it has correctly executed the cleanup operation on the 2
nodes

## OperationRebuild

This operation operates on multiple nodes in the cluster. Use this operation when CassKop add a new datacenter to an
existing cluster.

```
$ kubectl casskop rebuild start --from=dc1
```

In the background this command is equivalent to set labels on each pods like :
```
kubectl label pod cassandra-demo-dc2-rack1-0 operation-name=rebuild --overwrite
kubectl label pod cassandra-demo-dc2-rack1-0 operation-status=ToDo --overwrite
kubectl label pod cassandra-demo-dc2-rack1-0 operation-argument=dc1 --overwrite
```

## OperationDecommission

see [UpdateScaleDown](#updatescaledown)

## RollingRestart

This operation can be triggered with the plugin using simple commands as :

```console
$ k casskop restart --crd cassandra-e2e --rack dc1.rack1 dc2.rack1

Namespace cassandra-e2e
Trigger restart of dc1.rack1
Trigger restart of dc2.rack1

$ k casskop restart --crd cassandra-e2e --dc dc1

Namespace cassandra-e2e
Trigger restart of dc1.rack1
Trigger restart of dc1.rack2

$ k casskop restart --crd cassandra-e2e --full

Namespace cassandra-e2e
Trigger restart of dc1.rack1
Trigger restart of dc1.rack2
Trigger restart of dc2.rack1
```

After one of this command, CassKop will do a rolling restart of each rack one at a time avoiding any disruption.