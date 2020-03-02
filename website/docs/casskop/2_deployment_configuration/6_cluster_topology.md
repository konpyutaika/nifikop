---
id: 6_cluster_topology
title: Cluster topology - Cassandra rack aware deployments
sidebar_label: Cluster topology - Cassandra rack aware deployments
---

CassKop rack awareness feature helps to spread the Cassandra nodes replicas among different racks in the
kubernetes infrastructure. Enabling rack awareness helps to improve availability of Cassandra nodes and the data they
are hosting, through correct use of Cassandra's replication factor. 
 
> **Note:** Rack might represent an availability zone, data center or an actual physical rack in your data center.

## Quick overview

CassKop awareness can be configured in the `CassandraCluster.spec.topology` section. 

If the `topology` section is missing then CassKop will create a default Cassandra DC `dc1` and a default Rack
`rack1`. 
In this case, CassKop will not use specific kubernetes nodes labels for placement and in consequence the cluster is
not rack aware. 

In the `topology` section we can declare all the Cassandra Datacenters (DCs) and Racks, and for each of them we provide
labels which will need to match the labels assigned to the Kubernetes cluster nodes. 

Each of our rack targets Kubernetes nodes having the combination of labels defined in the `dc` section and in the `rack`
section. The labels are used in Kubernetes when scheduling the Cassandra pods to kubernetes nodes.
This has the effect of spreading the Cassandra nodes across physical zones.

 > **IMPORTANT:** CassKop doesn't rely on specific labels and will adapt to any topology you may define in your
 > datacenter or cloud provider. 

## Kubernetes nodes labels

Cassandra will run on Kubernetes Nodes, which may already have some labels representing their geographic topology.

Example :

![](../docs/assets/kubernetes-operators/topology-custom-example.png)

Example of labels for node001 in our dc:

```yaml
location.myorg.com/bay=1
location.myorg.com/building=building
location.myorg.com/label=SC_K08_-_KUBERNETES
location.myorg.com/room=Salle_1
location.myorg.com/site=SiteName
location.myorg.com/street=Rue_3
```

In the cloud the labels used for topology may better look like :

![](../docs/assets/kubernetes-operators/topology-custom-example.png)

```yaml
beta.kubernetes.io/fluentd-ds-ready=true
failure-domain.beta.kubernetes.io/region=europe-west1
kubernetes.io/hostname=gke-demo-default-pool-0c404f82-0100
beta.kubernetes.io/arch=amd64
failure-domain.beta.kubernetes.io/zone=europe-west1-d
beta.kubernetes.io/os=linux
cloud.google.com/gke-os-distribution=cos
beta.kubernetes.io/instance-type=n1-standard-4
cloud.google.com/gke-nodepool=default-pool
```

The idea is to use the Kubernetes nodes labels which refer to their physical position in the datacenters, to allow or
not Cassandra Pods placement.

Because CassKop manages its Cassandra Node Pods through 
[statefulset](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/), 
each Pod will inherit the placement configuration of its parent
statefulset. In order to place Pods on different racks, we need to associate a new statefulset with specialized nodes
placement constraints for each Cassandra Rack we define.


>Because the AZ feature was not yet available for statefulsets when we began to develop CassKop, we chose to
>implement 1 statefulset for 1 Cassandra Rack. Hopefully, we will be able to benefit of improvements about the
>topology-aware dynamic provisioning feature proposed in futur version of K8S (more informations here :
>https://kubernetes.io/blog/2018/10/11/topology-aware-volume-provisioning-in-kubernetes/
>)

See Example of configuration with topology : [cassandracluster-demo-gke.yaml](samples/cassandracluster-demo-gke.yaml)

## Configuring pod scheduling

When two applications are scheduled to the same Kubernetes node, both applications might use the same resources like
disk I/O and impact performances. It may be recommended to schedule Cassandra Pods in a way that avoids sharing nodes
with other critical workloads. Using the right nodes or dedicated a set of nodes only for cassandra are the best ways to
avoid such problems.

[Placement of Pods](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) in a Statefulsets are done 
using **NodeAffinity** and **PodAntiAffinity** : 

- [nodeAffinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) will be used to select specific
  nodes which have specific targeted labels.
- [podAntiAffinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity) can
  be used to ensure that critical applications are never scheduled on the same node.
  
### Cassandra node placement in the Kubernetes cluster

#### Node affinity

To target a Specific Kubernetes group of Nodes CassKop needs to define a specific **nodeAffinity**
section in the targeted `dc-rack` statefulset to match specific kubernetes nodes labels.

Example. If we want to deploy our statefulset only on nodes which have theses labels :

```yaml
location.myorg.com/site=Valbonne
location.myorg.com/building=HT2
location.myorg.com/room=Salle_1
location.myorg.com/street=Rue_11
```

CassKop need to add this section in the Statefulset definition :

```yaml
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - key: location.myorg.com/building
            operator: In
            values:
            - HT2
          - key: location.myorg.com/room
            operator: In
            values:
            - Salle_1
          - key: location.myorg.com/site
            operator: In
            values:
            - Valbonne
          - key: location.myorg.com/street
            operator: In
            values:
            - Rue_9
```

All Pods that will be created from this statefulset will target only nodes which have theses 4 labels.
If there is no kubernetes nodes available with theses labels, then the statefulset will remain stuck in a pending state,
until we correct either labels on nodes, or deployment constraints definition of the statefulset. 

We can also use specific labels to dedicate some kubernetes nodes to some type of work, for instance dedicated some
nodes for Cassandra. 

You can define custom labels on kubernetes nodes :

```sh
kubectl label node <your-node> <label-name>=<label-value>
```

> This is done automatically by combining the labels you specify in the Cassandra CRD definition in the `Topology`
> section. 

#### Pod anti affinity

If we loose a Kubernetes node, then we may want to limit the impact to loosing only one Cassandra node. Otherwise,
depending on the replication factor, we may have a data loss. 

In our CassandraCluster, the statefulset will target a pool of Kubernetes nodes using it's NodeSelector we just saw
above.
All theses Pods will inherit the specific labels from the Statefulset. 

To implement the limitation "one Cassandra Pod per Kubernetes node", we use the pod definition section
`podAntiAffinity`. This tells kubernetes that it can't deploy to a Kubernetes node, if a pod having the same labels
already exists. 

```yaml
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - podAffinityTerm:
          labelSelector:
            matchLabels:
              app: cassandracluster
              cassandracluster: cassandra
              cluster: k8s.kaas
          topologyKey: kubernetes.io/hostname
        weight: 100
```

This Tells that we **Require** not deploy to a node if a pod already exists with theses existing labels. 

> This is configured by default by CassKop

#### Using dedicated nodes

Cluster administrators can mark selected kubernetes nodes as tainted. Nodes with taints are excluded from regular
scheduling and normal pods will not be scheduled to run on them. Only services which can tolerate the taint set on the
node can be scheduled on it. The only other services running on such nodes will be kubernetes system services such  as
log collectors or software  defined networks

Taints can be used to create dedicated nodes. Running Cassandra on dedicated nodes can have many advantages. There will
be no other applications running on the same nodes which could cause disturbance or consume the resources needed  for
Cassandra. That can lead to improved performance and stability.

Example of tainting a node :

```sh
kubectl taint node <your-node> dedicated=Cassandra:NoSchedule
```

Additionally, add a label to the selected nodes as well

```sh
kubectl label node <your-node> dedicated=Cassandra
```

The `toleration` must be applied on the statefulset at the same level as `affinity`.

```yaml
...
    tolerations:
      - key: "dedicated"
        operator: "Equal"
        value: "Cassandra"
        effect: "NoSchedule"
...
```

> **IMPORTANT:** toleration must be used with node affinity on the same labels
> TODO: actually the toleration is not implemented in CassKop


### Configuring hard antiAffinity in Cassandra cluster

In development environment, we may have other concern than in production and we may allow our cluster to deploy several
nodes on the same kubernetes nodes.

The boolean `hardAntiAffinity` parameter in `CassandraCluster.spec` will define if we want the constraint to be
**Required** or **Preferred**


If `hardAntiAffinity=false` then the podAntiAffinity will be **preferred** instead of **required** and then kubernetes
will try to not put the cassandra node on the kubernetes node **BUT** it will allow to do it if it has no other choices.

## Cassandra notion of dc and racks

As we previously see, the Cassandra rack awareness is defined using several Cassandra datacenters `dc`s and `rack`s.
The `CassandraCluster.spec.topology` section allows us to define the virtual notion of DC & Rack.
For each we will define Kubernetes `labels` that will be used for pod placement.

> The name and numbers of labels used to define a DC & rack is not defined in advance and will be defined in each
> CassandraCluster manifests, depending on labels presents on Kubernetes Nodes and the required topology

## Configure the CassandraCluster CRD for dc & rack

If the section topology is missing: 
- then CassKop will deploy without label constraints to any available kubernetes nodes in the cluster
- there will be only one DC defined named by default **dc1**
- there will be only one Rack defined named by default **rack1**

If the **topology** section is defined, this will enable to create specific Cassandra dcs and racks: this creates a
specific statefulset for each rack and deploys on Kubernetes nodes matching the desired Node Labels (the concatenation
of DC labels + Rack labels for each statefulset)

Example of topology section:

```yaml
...
  nodesPerRacks: 3
...
  topology:
    dc:
      - name: dc1
        labels:
          location.k8s.myorg.com/site : Valbonne
          location.k8s.myorg.com/building : HT2
        rack:
          - name: rack1
            labels:
              location.k8s.myorg.com/room : Salle_1
              location.k8s.myorg.com/street : Rue_9
          - name: rack2
            labels:
              location.k8s.myorg.com/room : Salle_1
              location.k8s.myorg.com/street : Rue_10
      - name: dc2
        nodesPerRacks: 4
        labels:
          location.k8s.myorg.com/site : Valbonne
          location.k8s.myorg.com/building : HT2
        rack:
          - name: rack1
            labels:
              location.k8s.myorg.com/room : Salle_1
              location.k8s.myorg.com/street : Rue_11
```

- This will create 2 Cassandra DC (`dc1` & `dc2`)
    - For DC `dc1` it will create 2 Racks : `rack1` and `rack2`
      - In each of theses Racks there will be 3 Cassandra nodes
      - The `dc1` will have 6 nodes
    - For DC `dc2` it will create 1 Rack : `rack1`
      - the `dc2` overwrite the global parameter `nodesPerRacks=3` with a value of `4`. 
      - The `dc2` will have 4 nodes

> **Important:** We want to have the same numbers of Cassandra nodes in each Racks for a dedicated
> Datacenter. We can still have different values for different datacenters. 


The NodeSelectors labels for each Rack will be the aggregation of labels of the DC and the labels for the Racks :

For instance with this example, NodeSelectors labels for `dc1` / `rack2` will be :

 ```yaml
 location.k8s.myorg.com/site : Valbonne
 location.k8s.myorg.com/building : HT2
 location.k8s.myorg.com/room : Salle_1
 location.k8s.myorg.com/street : Rue_10
 ```

> The `dc` / `rack` topology definition is generic and does not rely on particular labels. It uses the ones
> corresponding to your needs.

> **Note:** The names for the dc and rack must be lowercase and respect Kubernetes DNS naming which follow [RFC 1123
> definition](http://tools.ietf.org/html/rfc1123#section-2) which can be expressed with this regular expression :
> `[a-z0-9]([-a-z0-9]*[a-z0-9])?`


### How CassKop configures dc and rack in Cassandra

CassKop will add 2 specific labels on each created Pod to tell them in witch Cassandra DC and Rack they belong :

Example :

```yaml
cassandraclusters.db.orange.com.dc=dc1
cassandraclusters.db.orange.com.rack=rack1
```

Using the Kubernetes DownwardAPI, CassKop will inject into the Cassandra Image 2 environment variables, from theses
2 labels. Excerpt from the Statefulset template :

```yaml
...
v1.EnvVar{
	Name: "CASSANDRA_DC",
	ValueFrom: &v1.EnvVarSource{
		FieldRef: &v1.ObjectFieldSelector{
			FieldPath: "metadata.labels['cassandraclusters.db.orange.com.dc']",
		},
	},
},
v1.EnvVar{
	Name: "CASSANDRA_RACK",
	ValueFrom: &v1.EnvVarSource{
		FieldRef: &v1.ObjectFieldSelector{
			FieldPath: "metadata.labels['cassandraclusters.db.orange.com.rack']",
		},
	},
},
...
```

In order to allow configuring Cassandra with the DC and Rack information, we use a specific [Cassandra
Image](https://github.com/Orange-OpenSource/cassandra-image/), which has a startup script that will retrieve
theses environment variables, and configure the Cassandra `cassandra-rackdc.properties` file with the values for dc and
rack.

The Cassandra Image makes us of the `GossipingPropertyFileSnitch` Cassandra Snitch, so that both Kubernetes and
Cassandra are aware of the chosen topology.