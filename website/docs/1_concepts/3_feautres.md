---
id: 3_features
title: Features
sidebar_label: Features
---

To highligt some of the features we needed and were not possible with the operators available, please keep reading 

## Fine Grained Node Config Support

We needed to be able to react to events in a fine-grained way for each Node - and not in the limited way StatefulSet does (which, for example, removes the most recently created Nodes). Some of the available solutions try to overcome these deficits by placing scripts inside the container to generate configs at runtime (a good example is our [Cassandra Operator](https://github.com/erdrix/casskop)), whereas the Orange NiFi operator's configurations are deterministically placed in specific Configmaps.

### Graceful NiFi Cluster Scaling

Apache NiFi is a good candidate to create an operator, because everything is made to orchestrate it through REST Api calls. With this comes automation of actions such as scaling, following all required steps : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#decommission-nodes.

### Graceful Rolling Upgrade

Operator supports graceful rolling upgrade, It means the operator will check if the cluster is healthy.

### Dynamic Configuration Support

NiFi operates with two type of configs:

- Read-only
- PerNode

Read only config requires node restart to update all the others may be updated dynamically.
Operator CRD distinguishes these fields, and proceed with the right action. It can be a rolling upgrade, or
a dynamic reconfiguration.