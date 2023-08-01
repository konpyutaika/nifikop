---
id: 0_design_principles
title: Design Principles
sidebar_label: Design Principles
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

Using StatefulSet we **lose** the ability to :

- modify the configuration of unique Nodes
- remove a specific Node from a cluster (StatefulSet always removes the most recently created Node)
- use multiple, different Persistent Volumes for each Node

The NiFi Operator uses `simple` Pods, ConfigMaps, and PersistentVolumeClaims, instead of StatefulSet (based on the design used by [Banzai Cloud Kafka Operator](https://github.com/banzaicloud/kafka-operator)).
Using these resources allows us to build an Operator which is better suited to NiFi.

With the NiFi operator we can:

- modify the configuration of unique Nodes
- remove specific Nodes from clusters
- use multiple Persistent Volumes for each Node