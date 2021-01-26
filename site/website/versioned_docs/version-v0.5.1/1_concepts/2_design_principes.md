---
id: 2_design_principes
title: Design Principes
sidebar_label: Design Principes
---

## Pod level management

NiFi is a stateful application. The first piece of the puzzle is the Node, which is a simple server capable of createing/forming a cluster with other Nodes. Every Node has his own **unique** configuration which differs slightly from all others.

All NiFi on Kubernetes setup use [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) to create a NiFi Cluster. Just to quickly recap from the K8s docs:

>StatefulSet manages the deployment and scaling of a set of Pods, and provide guarantees about their ordering and uniqueness. Like a Deployment, a StatefulSet manages Pods that are based on an identical container spec. Unlike a Deployment, a StatefulSet maintains sticky identities for each of its Pods. These pods are created from the same spec, but are not interchangeable: each has a persistent identifier that is maintained across any rescheduling.

How does this looks from the perspective of Apache NiFi ?

With StatefulSet we get :
- unique Node IDs generated during Pod startup
- networking between Nodes with headless services
- unique Persistent Volumes for Nodes

Using StatefulSet we **lose**  the ability to :

- modify the configuration of unique Nodes
- remove a specific Node from a cluster (StatefulSet always removes the most recently created Node)
- use multiple, different Persistent Volumes for each Node

The Orange NiFi Operator uses `simple` Pods, ConfigMaps, and PersistentVolumeClaims, instead of StatefulSet (based on the design used by [Banzai Cloud Kafka Operator](https://github.com/banzaicloud/kafka-operator)). 
Using these resources allows us to build an Operator which is better suited to NiFi.

With the Orange NiFi operator we can:

- modify the configuration of unique Nodes
- remove specific Nodes from clusters
- use multiple Persistent Volumes for each Node

## Dataflow Lifecycle management

The [Dataflow Lifecycle management feature](./3_features.md#dataflow-lifecycle-management-via-crd) introduces 3 new CRDs :

- **NiFiRegistryClient :** Allowing you to declare a [NiFi registry client](https://nifi.apache.org/docs/nifi-registry-docs/html/getting-started.html#connect-nifi-to-the-registry).
- **NiFiParameterContext :** Allowing you to create parameter context, with two kinds of parameters, a simple `map[string]string` for non-sensitive parameters and a `list of secrets` which contains sensitive parameters.
- **NiFiDataflow :** Allowing you to declare a Dataflow based on a `NiFiRegistryClient` and optionally a `ParameterContext`, which will be deployed and managed by the operator on the `targeted NiFi cluster`.

The following diagram shows the interactions between all the components : 

![dataflow lifecycle management schema](/img/1_concepts/2_design_principes/dataflow_lifecycle_management_schema.jpg)

With each CRD comes a new controller, with a reconcile loop : 

- **NiFiRegistryClient's controller :** 

![NiFi registry client's reconcile loop](/img/1_concepts/2_design_principes/registry_client_reconcile_loop.jpeg)

- **NiFiParameterContext's controller :** 

![NiFi parameter context's reconcile loop](/img/1_concepts/2_design_principes/parameter_context_reconcile_loop.jpeg)

- **NiFiDataflow's controller :** 

![NiFi dataflow's reconcile loop](/img/1_concepts/2_design_principes/dataflow_reconcile_loop.jpeg)

