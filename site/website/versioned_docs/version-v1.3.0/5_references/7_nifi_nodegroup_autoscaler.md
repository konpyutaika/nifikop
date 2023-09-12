---
id: 7_nifi_nodegroup_autoscaler
title: NiFi NodeGroup Autoscaler
sidebar_label: NiFi NodeGroup Autoscaler
---

`NifiNodeGroupAutoscaler` is the Schema through which you configure automatic scaling of `NifiCluster` deployments.

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiNodeGroupAutoscaler
metadata:
  name: nifinodegroupautoscaler-sample
spec:
  # contains the reference to the NifiCluster with the one the node group autoscaler is linked.
  clusterRef:
    name: nificluster-name
    namespace: nifikop
  # defines the id of the NodeConfig contained in NifiCluster.Spec.NodeConfigGroups
  nodeConfigGroupId: default-node-group
  # The selector used to identify nodes in NifiCluster.Spec.Nodes this autoscaler will manage
  # Use Node.Labels in combination with this selector to clearly define which nodes will be managed by this autoscaler 
  nodeLabelsSelector: 
    matchLabels:
      nifi_cr: nificluster-name
      nifi_node_group: default-node-group
  # the strategy used to decide how to add nodes to a nifi cluster
  upscaleStrategy: simple
  # the strategy used to decide how to remove nodes from an existing cluster
  downscaleStrategy: lifo
```

## NifiNodeGroupAutoscaler
|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects nodegroupautoscalers must create.|No|nil|
|spec|[NifiNodeGroupAutoscalerSpec](#nifinodegroupautoscalerspec)|defines the desired state of NifiNodeGroupAutoscaler.|No|nil|
|status|[NifiNodeGroupAutoscalerStatus](#nifinodegroupautoscalerstatus)|defines the observed state of NifiNodeGroupAutoscaler.|No|nil|

## NifiNodeGroupAutoscalerSpec

|Field| Type                                                                                |Description|Required|Default|
|-----|-------------------------------------------------------------------------------------|-----------|--------|--------|
|clusterRef| [ClusterReference](./2_nifi_user#clusterreference)                                |  contains the reference to the NifiCluster containing the node group this autoscaler should manage. |Yes| - |
|nodeConfigGroupId| string                                                                              | defines the id of the [NodeConfig](./1_nifi_cluster/3_node_config) contained in `NifiCluster.Spec.NodeConfigGroups`. |Yes| - |
|nodeLabelsSelector| [LabelSelector](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#LabelSelector) | defines the set of labels used to identify nodes in a `NifiCluster` node config group. Use `Node.Labels` in combination with this selector to clearly define which nodes will be managed by this autoscaler. Take care to avoid having mutliple autoscalers managing the same nodes. |Yes| - |
|readOnlyConfig| [ReadOnlyConfig](./1_nifi_cluster/2_read_only_config)                             | defines a readOnlyConfig to apply to each node in this node group. Any settings here will override those set in the configured `nodeConfigGroupId`. |Yes| - |
|nodeConfig| [NodeConfig](./1_nifi_cluster/3_node_config)                | defines a nodeConfig to apply to each node in this node group. Any settings here will override those set in the configured `nodeConfigGroupId`. |Yes| - |
|upscaleStrategy| string                                                                              | The strategy NiFiKop will use to scale up the nodes managed by this autoscaler. Must be one of {`simple`}. |Yes| - |
|downscaleStrategy| string                                                                              | The strategy NiFiKop will use to scale down the nodes managed by this autoscaler. Must be one of {`lifo`}. |Yes| - |
|replicas| int                                                                                 | the initial number of replicas to configure the `HorizontalPodAutoscaler` with. After the initial configuration, this `replicas` configuration will be automatically updated by the Kubernetes `HorizontalPodAutoscaler` controller. |No| 0 |

## NifiNodeGroupAutoscalerStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|state|string| the state of the nodegroup autoscaler. This is set by the autoscaler. |No| - |
|replicas|int| the current number of replicas running in the node group this autoscaler is managing. This is set by the autoscaler.|No| - |
|selector|string| the [selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) used by the `HorizontalPodAutoscaler` controller to identify the replicas in this node group. This is set by the autoscaler.|No| - |