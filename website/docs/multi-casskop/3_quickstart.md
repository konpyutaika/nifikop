---
id: 3_quickstart
title: Quickstart
sidebar_label: Quickstart
---

## Create the MultiCassKop CRD

You can deploy a MultiCassKop CRD instance.

You can find an example in the `multi-casskop/samples/multi-casskop.yaml` file:

this is the structure of the file:

```yaml
apiVersion: db.orange.com/v1alpha1
kind: MultiCasskop
metadata:
  name: multi-casskop-e2e
spec:
  # Add fields here
  deleteCassandraCluster: true
  base:
    <The base of the CassandraCluster object you want Multi-CassKop to create>
    ...
    status: #<!!-- At this time the seedlist must be provided Manually. we know in advance the domain name
            # it's the <pod-name>.<your-external-dns-domain>
      seedlist:
        - cassandra-e2e-dc1-rack1-0.my.zone.dns.net
        - cassandra-e2e-dc1-rack1-1.my.zone.dns.net
        - cassandra-e2e-dc2-rack4-0.my.zone.dns.net
        - cassandra-e2e-dc2-rack4-1.my.zone.dns.net      
  override:
    k8s-cluster1: #<!!-- here is the name which must correspond to the helm argument `k8s.local`
      spec: #<-- Here we defined overrides part for the CassandraCluster for the k8s-cluster1
        pod:
          annotations:
            cni.projectcalico.org/ipv4pools: '["routable"]' #<!-- If using external DNS, change with your current zone
        topology:
          dc:
          ...
    k8s-cluster2: #<!!-- here is the name which must correspond to the helm argument `k8s.remote`
      spec:
        pod:
          annotations:
            cni.projectcalico.org/ipv4pools: '["routable"]' #<!-- If using external DNS, change with your current zone
        imagepullpolicy: IfNotPresent
        topology:
          dc:
          ...
```

You can create the Cluster with :

```sh
k apply -f multi-casskop/samples/multi-casskop.yaml
```

Then you can see the MultiCassKop logs:

```log
time="2019-11-28T15:46:19Z" level=info msg="Just Update CassandraCluster, returning for now.." cluster=cassandra-e2e kubernetes=k8s-cluster1 namespace=cassandra-e2e
time="2019-11-28T15:46:19Z" level=info msg="Cluster is not Ready, we requeue [phase= / action= / status=]" cluster=cassandra-e2e kubernetes=k8s-cluster1 namespace=cassandra-e2e
time="2019-11-28T15:46:49Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing / status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster1 namespace=cassandra-e2e
time="2019-11-28T15:47:19Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing / status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster1 namespace=cassandra-e2e
time="2019-11-28T15:47:49Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing /status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster1 namespace=cassandra-e2e
time="2019-11-28T15:47:19Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing / status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster1 namespace=cassandra-e2e
time="2019-11-28T15:47:49Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing / status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster1 namespace=cassandra-e2e
time="2019-11-28T15:48:19Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing / status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster1 namespace=cassandra-e2e
time="2019-11-28T15:48:49Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing / status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster1 namespace=cassandra-e2e
time="2019-11-28T15:49:19Z" level=info msg="Just Update CassandraCluster, returning for now.." cluster=cassandra-e2e kubernetes=k8s-cluster2 namespace=cassandra-e2e
time="2019-11-28T15:49:49Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing / status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster2 namespace=cassandra-e2e
time="2019-11-28T15:50:19Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing / status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster2 namespace=cassandra-e2e
time="2019-11-28T15:50:49Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing / status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster2 namespace=cassandra-e2e
time="2019-11-28T15:51:19Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing / status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster2 namespace=cassandra-e2e
time="2019-11-28T15:51:49Z" level=info msg="Cluster is not Ready, we requeue [phase=Initializing / action=Initializing / status=Ongoing]" cluster=cassandra-e2e kubernetes=k8s-cluster2 namespace=cassandra-e2e
```

This is the sequence of operations:

- MultiCassKop first creates the CassandraCluster in k8s-cluster1. 
- Then local CassKop starts to creates the associated Cassandra Cluster.
  - When CassKop has created its Cassandra cluster, it updates CassandraCluster object's status with the phase=Running meaning that
  all is ok
- Then MultiCassKop start creating the other CassandraCluster in k8s-cluster2
- Then local CassKop started to creates the associated Cassandra Cluster.
  - Thanks to the routable seed-list configured with external dns names, Cassandra pods are started by connecting to
    already existings Cassandra nodes from k8s-cluster1 with the goal to form a uniq Cassandra Ring.

In resulting, We can see that each clusters have the required pods.

If we go in one of the created pods, we can see that nodetool see pods of both clusters:

```console
cassandra@cassandra-e2e-dc1-rack2-0:/$ nodetool status
Datacenter: dc1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address         Load       Tokens       Owns (effective)  Host ID                               Rack
UN  10.100.146.150  93.95 KiB  256          49.8%             cfabcef2-3f1b-492d-b028-0621eb672ec7  rack2
UN  10.100.146.108  108.65 KiB  256          48.3%             d1185b37-af0a-42f9-ac3f-234e541f14f0  rack1
Datacenter: dc2
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address         Load       Tokens       Owns (effective)  Host ID                               Rack
UN  10.100.151.38   69.89 KiB  256          51.4%             ec9003e0-aa53-4150-b4bb-85193d9fa180  rack5
UN  10.100.150.34   107.89 KiB  256          50.5%             a28c3c59-786f-41b6-8eca-ca7d7d14b6df  rack4

cassandra@cassandra-e2e-dc1-rack2-0:/$
```

## Delete the Cassandra cluster2

If you have set the `deleteCassandraCluster` to true, then when deleting the `MultiCassKop` object, it will cascade the
deletion of the `CassandraCluster` object in the targeted k8s clusters. Then each local CassKop will delete their
Cassandra clusters.

you can see in the MultiCassKop logs

```log
time="2019-11-28T15:44:00Z" level=info msg="Delete CassandraCluster" cluster=cassandra-e2e kubernetes=k8s-cluster1 namespace=cassandra-e2e
time="2019-11-28T15:44:00Z" level=info msg="Delete CassandraCluster" cluster=cassandra-e2e kubernetes=k8s-cluster2 namespace=cassandra-e2e
```