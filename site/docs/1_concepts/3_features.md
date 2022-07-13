---
id: 3_features
title: Features
sidebar_label: Features
---

To highligt some of the features we needed and were not possible with the operators available, please keep reading 

## Fine Grained Node Config Support

We needed to be able to react to events in a fine-grained way for each Node - and not in the limited way StatefulSet does (which, for example, removes the most recently created Nodes). Some of the available solutions try to overcome these deficits by placing scripts inside the container to generate configs at runtime (a good example is our [Cassandra Operator](https://github.com/Orange-OpenSource/casskop)), whereas the Orange NiFi operator's configurations are deterministically placed in specific Configmaps.

## Graceful NiFi Cluster Scaling

Apache NiFi is a good candidate to create an operator, because everything is made to orchestrate it through REST Api calls. With this comes automation of actions such as scaling, following all required steps : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#decommission-nodes.

## Graceful Rolling Upgrade

Operator supports graceful rolling upgrade. It means the operator will check if the cluster is healthy.

## Dynamic Configuration Support

NiFi operates with two type of configs:

- Read-only
- PerNode

Read only config requires node restart to update all the others may be updated dynamically.
Operator CRD distinguishes these fields, and proceed with the right action. It can be a rolling upgrade, or
a dynamic reconfiguration.

## Dataflow lifecycle management via CRD

In a cloud native approach, we are looking for important management features, which we have applied to NiFi Dataflow : 

- **Automated deployment :** Based on the NiFi registry, you can describe your `NiFiDataflow` resource that will be deployed and run on the targeted NiFi cluster.
- **Portability :** On kubernetes everything is a yaml file, so with NiFiKop we give you the ability to describe your clusters but also the `registry clients`, `parameter contexts` and `dataflows` of your NiFi application, so that you can redeploy the same thing in a different namespace or cluster.
- **State management :** With NiFiKop resources, you can describe what you want, and the operator deals with the NiFi Rest API to make sure the resource stays in sync (even if someone manually makes changes directly on NiFi cluster).
- **Configurations :** Based on the `Parameter Contexts`, NiFiKop allows you to associate to your `Dataflow` (= your applications) with a different configuration depending on the environment !

## Users and access policies management

Without the management of users and access policies associated, it was not possible to have a fully automated NiFi cluster setup due to : 

- **Node scaling :** when a new node joins the cluster it needs to have some roles like `proxy user request`, `view data` etc., by managing users and access policies we can easily create a user for this node with the right accesses.
- **Operator admin rights :** For the operator to manage efficiently the cluster it needs a lot of rights as `deploying process groups`, `empty the queues` etc., these rights are not available by default when you set a user as [InitialAdmin](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#initial-admin-identity). Once again by giving the ability to define users and access policies we go through this.
- **User's access :** as seen just below we need to define the operator as `InitialAdmin`, in this situation there is no more users that can access to the web UI to manually give access to other users. That's why we extend the `InitialAdmin` concept into the operator, giving the ability to define a list of users as admins.

In addition to these requirements to have a fully automated and managed cluster, we introduced some useful features : 

- **User management :** using `NifiUser` resource, you are able to create (or bind an existing) user in NiFi cluster and apply some access policies that will be managed and continuously synced by the operator.
- **Group management :** using `NifiUserGroup` resource, you can create groups in NiFi cluster and apply access policies and a list of `NifiUser` that will be managed and continuously synced by the operator.
- **Default group :** As the definition of `NifiUser` and `NifiUserGroup` resources could be heavy for some simple use cases, we also decided to define two default groups that you can feed with a list of users that will be created and managed by the operator (no kubernetes resources to create) : 
    - **Admins :** a group giving access to everything on the NiFi Cluster,
    - **Readers :** a group giving access as viewer on the NiFi Cluster.

By introducing this feature we are giving you the ability to fully automate your deployment, from the NiFi Cluster to your managed NiFi Dataflow.

## Automatic horizontal NiFi cluster scaling via CRD

NiFiKop supports automatically horizontally scaling `NifiCluster` node groups with a `NifiNodeGroupAutoscaler` custom resource. 

- **Kubernetes native :** The `NifiNodeGroupAutoscaler` controller implements the [Kubernetes scale subresource](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#scale-subresource) and creates a Kubernetes [`HorizontalPodAutoscaler`](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) to automatically scale the pods that NiFiKop creates for `NifiCluster` deployments.
- **Metrics-driven autoscaling :** The `HorizontalPodAutoscaler` can be driven by pod usage metrics (e.g. RAM, CPU) or through NiFi application metrics scraped by Prometheus.
- **Flexible NifiCluster node group autoscaling :** The `NifiNodeGroupAutoscaler` scales specific node groups in your `NifiCluster` and you may have as many autoscalers as you like per `NifiCluster` deployment. For example, a `NifiNodeGroupAutoscaler` may manage high-memory or high-cpu sets of nodes for volume burst scenarios or it may manage every node in your cluster.

Through this set of features, you may elect to have NiFiKop configure automatic horizontal autoscaling for any subset of nodes in your `NifiCluster` deployment.