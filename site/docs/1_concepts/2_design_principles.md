---
id: 2_design_principles
title: Design Principles
sidebar_label: Design Principles
---

This operator is built following the logic implied by the [operator sdk framework] (https://sdk.operatorframework.io/).
What we want to offer with NiFiKop is that we provide as much automation as possible to manage NiFi at scale and in the most automated way possible.

## Separation of concerns

Kubernetes is designed for automation. Right out of the box, the Kubernetes core has a lot of automation features built in. You can use Kubernetes to automate the deployment and execution of workloads, and you can automate how Kubernetes does it.

The Kubernetes operator model concept allows you to extend cluster behavior without changing the Kubernetes code itself by binding controllers to one or more custom resources. Operators are clients of the Kubernetes API that act as controllers for a custom resource.

There are two things we can think of when we talk about operators in Kubernetes:

- Automate the deployment of my entire stack.
- Automate the actions required by the deployment.

For NiFiKop, we focus primarily on NiFi for the stack concept, what does that mean?

- We do not manage other components that can be integrated with NiFi Cluster like Prometheus, Zookeeper, NiFi registry etc.
- We want to provide as many tools as possible to automate the work on NiFi (cluster deployment, data flow and user management, etc.).

We consider that for NiFiKop to be a low-level operator, focused on NiFi and only NiFi, and if we want to manage a complex stack with e.g. Zookeeper, NiFi Registry, Prometheus etc. we need something else working at a higher level, like Helm charts or another operator controlling NiFiKop and other resources.

## One controller per resource

The operator should reflect as much as possible the behavior of the solution we want to automate. If we take our case with NiFi, what we can say is that:

- You can have one or more NiFi clusters
- On your cluster you can define a NiFi registry client, but it is not mandatory.
- You can also define users and groups and deploy a DataFlow if you want.

This means that your cluster is not defined by what is deployed on it, and what you deploy on it does not depend on it.
To be more explicit, just because I deploy a NiFi cluster doesn't mean the DataFlow deployed on it will stay there, we can move the DataFlow from one cluster to another.

To manage this, we need to create different kinds of resources ([NifiCluster], [NifiDataflow], [NifiParameterContext], [NifiUser], [NifiUserGroup], [NifiRegistryClient], [NifiNodeGroupAutoscaler], [NifiConnection]) with one controller per resource that will manage its own resource.
In this way, we can say:

- I deploy a NiFiCluster
- I define a NiFiDataflow that will be deployed on this cluster, and if I want to change cluster, I just have to change the `ClusterRef` to change cluster


[NifiCluster]: ../5_references/1_nifi_cluster
[NifiDataflow]: ../5_references/5_nifi_dataflow
[NifiParameterContext]: ../5_references/4_nifi_parameter_context
[NifiUser]: ../5_references/2_nifi_user
[NifiUserGroup]: ../5_references/6_nifi_usergroup
[NifiRegistryClient]: ../5_references/3_nifi_registry_client
[NifiNodeGroupAutoscaler]: ../5_references/7_nifi_nodegroup_autoscaler
[NifiConnection]: ../5_references/8_nifi_connection