---
id: 1_scaling_mechanism
title: Scaling mechanism
sidebar_label: Scaling mechanism
---

This tasks shows you how to perform a gracefull cluster scale up and scale down.

## Before you begin

- Setup NiFiKop by following the instructions in the [Installation guide](../../../2_deploy_nifikop/1_quick_start).
- Deploy the [Simple NiFi](../1_deploy_cluster/1_quick_start) sample cluster.
- Review the [Node](../../../5_references/1_nifi_cluster/4_node) references doc.

## About this task

The [Simple NiFi](../1_deploy_cluster/1_quick_start) example consists of a three nodes NiFi cluster.
A node decommission must follow a strict procedure, described in the [NiFi documentation](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#decommission-nodes) :

1. Disconnect the node
2. Once disconnect completes, offload the node.
3. Once offload completes, delete the node.
4. Once the delete request has finished, stop/remove the NiFi service on the host.


For the moment, we have implemented it as follows in the operator :

1. Disconnect the node
2. Once disconnect completes, offload the node.
3. Once offload completes, delete the pod.
4. Once the pod deletion completes, delete the node.
5. Once the delete request has finished, remove the node from the NifiCluster status.

In addition, we have a regular check that ensure that all nodes have been removed.

In this task, you will first perform a scale up, in adding an new node. Then, you will remove another node that the one created, and observe the decommission's steps.

## Scale up : Add a new node

For this task, we will simply add a node with the same configuration than the other ones, if you want to know more about how to add a node with an other configuration let's have a look to the [Node configuration](../1_deploy_cluster/2_nodes_configuration) documentation page.

1. Add and run a dataflow as the example :

![Scaling dataflow](/img/3_tasks/1_nifi_cluster/2_cluster_scaling/scaling_dataflow.png)

2. Add a new node to the list of `NifiCluster.Spec.Nodes` field, by following the [Node object definition](../../../5_references/1_nifi_cluster/4_node) documentation:

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiCluster
metadata:
  name: simplenifi
spec:
  service:
    headlessEnabled: true
  zkAddress: "zookeepercluster-client.zookeeper:2181"
  zkPath: "/simplenifi"
  clusterImage: "apache/nifi:1.12.1"
  oneNifiNodePerNode: false
  nodeConfigGroups:
    default_group:
      isNode: true
      storageConfigs:
        - mountPath: "/opt/nifi/nifi-current/logs"
          name: logs
          metadata:
            labels:
              my-label: my-value
            annotations:
              my-annotation: my-value
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            storageClassName: "standard"
            resources:
              requests:
                storage: 10Gi
      serviceAccountName: "default"
      resourcesRequirements:
        limits:
          cpu: "2"
          memory: 3Gi
        requests:
          cpu: "1"
          memory: 1Gi
  nodes:
    - id: 0
      nodeConfigGroup: "default_group"
    - id: 1
      nodeConfigGroup: "default_group"
    - id: 2
      nodeConfigGroup: "default_group"
# >>>> START: The new node
    - id: 25
      nodeConfigGroup: "default_group"
# <<<< END
  propagateLabels: true
  nifiClusterTaskSpec:
    retryDurationMinutes: 10
  listenersConfig:
    internalListeners:
      - type: "http"
        name: "http"
        containerPort: 8080
      - type: "cluster"
        name: "cluster"
        containerPort: 6007
      - type: "s2s"
        name: "s2s"
        containerPort: 10000
```

:::important
**Note :** The `Node.Id` field must be unique in the `NifiCluster.Spec.Nodes` list.
:::

3. Apply the new `NifiCluster` configuration :

```sh 
kubectl -n nifi apply -f config/samples/simplenificluster.yaml
```

4. You should now have the following resources into kubernetes :

```console 
kubectl get pods,configmap,pvc -l nodeId=25
NAME                          READY   STATUS    RESTARTS   AGE
pod/simplenifi-25-nodem5jh4   1/1     Running   0          11m

NAME                             DATA   AGE
configmap/simplenifi-config-25   7      11m

NAME                                               STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/simplenifi-25-storagehwn24   Bound    pvc-7da86076-728e-11ea-846d-42010a8400f2   10Gi       RWO            standard       11m
```

And if you go on the NiFi UI, in the cluster administration page :

![Scale up, cluster list](/img/3_tasks/1_nifi_cluster/2_cluster_scaling/scaleup_cluster_list.png)

5. You now have data on the new node :

![Scale up, cluster distribution](/img/3_tasks/1_nifi_cluster/2_cluster_scaling/scaleup_distribution.png)

## Scaledown : Gracefully remove node

For this task, we will simply remove a node and look at that the decommissions steps.

1. Remove the node from the list of `NifiCluster.Spec.Nodes` field :

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiCluster
metadata:
  name: simplenifi
spec:
  headlessServiceEnabled: true
  zkAddresse: "zookeepercluster-client.zookeeper:2181"
  zkPath: "/simplenifi"
  clusterImage: "apache/nifi:1.11.3"
  oneNifiNodePerNode: false
  nodeConfigGroups:
    default_group:
      isNode: true
      storageConfigs:
        - mountPath: "/opt/nifi/nifi-current/logs"
          name: logs
          metadata:
            labels:
              my-label: my-value
            annotations:
              my-annotation: my-value
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            storageClassName: "standard"
            resources:
              requests:
                storage: 10Gi
      serviceAccountName: "default"
      resourcesRequirements:
        limits:
          cpu: "2"
          memory: 3Gi
        requests:
          cpu: "1"
          memory: 1Gi
  nodes:
    - id: 0
      nodeConfigGroup: "default_group"
    - id: 1
      nodeConfigGroup: "default_group"
# >>>> START: node removed
#    - id: 2
#      nodeConfigGroup: "default_group"
# <<<< END
    - id: 25
      nodeConfigGroup: "default_group"
  propagateLabels: true
  nifiClusterTaskSpec:
    retryDurationMinutes: 10
  listenersConfig:
    internalListeners:
      - type: "http"
        name: "http"
        containerPort: 8080
      - type: "cluster"
        name: "cluster"
        containerPort: 6007
      - type: "s2s"
        name: "s2s"
        containerPort: 10000
```

2. Apply the new `NifiCluster` configuration :

```sh 
kubectl -n nifi apply -f config/samples/simplenificluster.yaml
```

3. You can follow the node's action step status in the `NifiCluster.Status` description :

```console 
kubectl describe nificluster simplenifi

...
Status:
  Nodes State:
    ...
    2:
      Configuration State:  ConfigInSync
      Graceful Action State:
        Action State:   GracefulDownscaleRequired
        Error Message:
    ...
...
```

:::tip
The list of decommisions step and their corresponding value for the `Nifi Cluster.Status.Node State.Graceful ActionState.ActionStep` field is described into the [Node State page](../../../5_references/1_nifi_cluster/5_node_state#actionstep)
:::

4. Once the scaledown successfully performed, you should have the data offloaded on the other nodes, and the node state removed from the `NifiCluster.Status.NodesState` list :

:::warning
Keep in mind that the [`NifiCluster.Spec.nifiClusterTaskSpec.retryDurationMinutes`](../../../5_references/1_nifi_cluster/1_nifi_cluster.md#nificlustertaskspec) should be long enough to perform the whole procedure, or you will have some rollback and retry loop.
:::