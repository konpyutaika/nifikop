---
id: 0_design_principles
title: Design Principles
sidebar_label: Design Principles
---

:::info
These feature have been scpoed by the community, please find the discussion and technical scoping [here] (https://docs.google.com/document/d/1QNGSNNjWx4CGt5-NvX9ZArQMfyrwjw-B95f54GUNdB0/edit#heading=h.t9xh94v7viuj).
:::

## Design reflexion

If you read the technical scoping above, we explored many options for enabling automatic scaling of NiFi clusters.
After much discussion, it turned out that we wanted to mimic the approach and design behind auto-scaling a deployment with [HPA] (https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/).

If we look at how this works, you define a `Deployment`, which will manage a `ReplicaSet` which will manage `Pods`. And you define your `HPA` which will manage the scale field of the `Deployment`.
For our `NiFiCluster` we considered the same kind of separation of concerns: we define a new resource `NifiNodeGroupAutoScaler` that manages the `NifiCluster` that will manage the `Pods`. And you define your `HPA` which will manage the scale field of the `Deployment`.

This is the basis of the thinking. There was another inspiration for designing the functionality, which is that we wanted to keep the possibility of different types of node groups and manage them separately, so we pushed by thinking about similar existing models, and we thought about how in the Kubernetes Cloud Cluster (EKS, GKE etc.) nodes can be managed.
You can define fixed groups of nodes, you can auto-scale others.

And finally, we wanted to separate the `NifiCluster` itself from the `autoscaling management` and allow mixing the two, allowing you to have a cluster initially with no scaling at all, add scaling from a subset of nodes with a given configuration, and finally disable autoscaling without any impact.

## Implementation

Referring to the official guideline, the recommended approach is to implement [the sub resource scale in the CRD](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#scale-subresource).

This approach requires to define :
- `specReplicasPath` defines the JSONPath inside of a custom resource that corresponds to `scale.spec.replicas`
- `statusReplicasPath` defines the JSONPath inside of a custom resource that corresponds to `scale.status.replicas`
- `labelSelectorPath` defines the JSONPath inside of a custom resource that corresponds to `scale.Status.Selector`

we add a new resource : [NifiNodeGroupAutoScaler](../../../../5_references/7_nifi_nodegroup_autoscaler), with the following fields :  
- `spec.nifiClusterRef` : reference to the NiFi cluster resource that will be autoscaled
- `spec.nodeConfigGroupId` : reference to the nodeConfigGroup that will be used for nodes managed by the auto scaling.
- `spec.readOnlyConfig` : defines a readOnlyConfig to apply to each node in this node group. Any settings here will override those set in the configured `NifiCluster`.
- `spec.nodeConfig` : defines a nodeConfig to apply to each node in this node group. Any settings here will override those set in the configured `nodeConfigGroupId`.
- `spec.replicas` : current number of replicas expected for the node config group
- `spec.upscaleStrategy` : strategy used to upscale (simple)
- `spec.downscaleStrategy` : strategy used to downscale (lifo)

Here is a representation  of dependencies:

![auto scaling schema](/img/auto_scaling.jpg)