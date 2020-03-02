---
id: 7_implementation_architecture
title: Implementation architecture
sidebar_label: Implementation architecture
---

## 1 Statefulset for each racks

CassKop will create a dedicated statefulset and service for each couple `dc-rack` defined in the
`topology`section. This is done to ensure we'll always have the same amounts of cassandra nodes in each rack for a
specified DC.


![architecture](http://www.plantuml.com/plantuml/proxy?src=https://raw.github.com/Orange-OpenSource/casskop/master/documentation/uml/architecture.puml)

## Sequences

CassKop will works in sequence for each DC-Rack it has created which are different statefulsets kubernetes objects.
Each time we request a change on the cassandracluster CRD which implies rollingUpdate of the statefulset, CassKop will
perform the update on the first dc-rack.

> CassKop will then wait for the operation to complete before starting the upgrade on the next dc-rack!!

If you play with `spec.topology.dc[].rack[].rollingPartition` with value greater than 0, then the rolling update of the rack
won't end and CassKop won't update the next one. In order to allow a statefulset to upgrade completely the rollingPartition must be set to 0 (default).


## Naming convention of created objects

When declaring a new `CassandraCluster`, we need to specify its Name, and all its configuration.

Here is an excerpt of a CassandraCluster CRD definition:

```yaml
apiVersion: "db.orange.com/v1alpha1"
kind: "CassandraCluster"
metadata:
  name: cassandra-demo
  labels:
    cluster: optional-label 
spec:
  ...
  nodesPerRacks: 3
  
  topology:
    dc:
      - name: dc1
        labels:
          location.myorg.com/site : mts
        rack:
          - name: rack1
            labels: 
              location.myorg.com/street : street1
          - name: rack2
            labels: 
              location.myorg.com/street : street2
      - name: dc2
        nodesPerRacks: 4
        labels: 
          location.myorg.com/site : mts
        rack:
          - name: rack1
            labels: 
              location.myorg.com/street : street3
```

A complete example can be found [here](https://raw.github.com/Orange-OpenSource/casskop/samples/cassandracluster-pic.yaml)

Kubernetes objects created by CassKop are named according to : 
- `<CassandraClusterName>-<DCName>-<RackName>` 

> **IMPORTANT:** All elements must be in lowerCase according to Kubernetes DNS naming constraints

### List of resources created as part of the Cassandra cluster

- `<cluster-name>`
    - PodDisruptionBudget: this is checked by Kubernetes and by CassKop and allows only 1 pod disrupted
      on the whole cluster. CassKop won't update statefulset in case there is a disruption.
- `<cluster-name>-<dc-name>`
    - [Headless service](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) at dc level
      used as client applications entry point to contact all nodes in a Cassandra DC.
- `<cluster-name>-<dc-name>-<rack-name>`
    - Statefulset which is in charge of managing Cassandra Pods for dc-name and rack-name
    - Service headless used for Seeds discovery
- `<cluster-name>-<dc-name>-<rack-name>-<idx>`
    - Pods Names in the Statefulset for dc-name and rack-name with ordinal index.
- `data-<cluster-name>-<dc-name>-<rack-name>-<idx>`
    - PersistentVolumeClaim representing the data for the associated Cassandra pod.
- `<cluster-name>-<dc-name>-<rack-name>-exporter-jmx`
    - Service Name for the exporter JMX for dc-name and rack-name
- `<cluster-name>`


With the previous example:
- The CassandraCluster name is `cassandra-demo`
- the first DC is named `dc1`
    - the first rack is named `rack1`
    - the second rack is named `rack2`
- the second DC is named `dc2`
    - the first rack is named `rack1`

Example for DC `dc1-rack1` :
- the statefulsets is named : `cassandra-demo-dc1-rack1`,`cassandra-demo-dc1-rack2`,`cassandra-demo-dc2-rack1`
    - the statefulset Pods name will add the ordinal number suffix :
      `cassandra-demo-dc1-rack1-0`,..,`cassandra-demo-dc1-rack1-n` for each dc-racks
- The services will be names : `cassandra-demo-dc1` and `cassandra-demo-dc2` 
- the associated service for Prometheus metrics export will be named :
  `cassandra-demo-dc1-exporter-jmx`,`cassandra-demo-dc2-exporter-jmx`  
- the PVC (Persistent Volume Claim) of each pods will be named **data-{podName}** ex: `data-cassandra-demo-dc1-rack1-0`
  for each dc-racks
- the [PodDisruptionBudget (PDB)](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) will be named :
  `cassandra-demo` and will target all pods of the cluster
  
> **Note:**: usually the PDB is only used when dealing with pod eviction (draining a kubernetes node). But CassKop
> also checks the PDB to know if it is allowed to make some actions on the cluster (restart Pod, apply changes..). If
> the PDB won't allow CassKop to make the change, it will wait until the PDB rule is satisfied. (We won't be able
> to make any change on the cluster, but it cassandra will continue to work underneeth).
