---
id: 2_nodes_configuration
title: Nodes configuration
sidebar_label: Nodes configuration
---

In the quick start section, you deployed a simple `NifiCluster` resource, which deploys a NiFi cluster. But in many cases, you may need to tune the cluster nodes to match your needs.
In this section, we'll try to cover the various things that can be specified for cluster node configuration.

## ReadOnlyConfig and NodeConfigGroups

To set up your `NiFi` cluster with `NiFiKop`, the first thing to understand is the difference between `readOnlyConfig` and `nodeConfigGroups`.

### Configure multiple node groups

In NiFiKop you can define different types of nodes using the `Spec.NodeConfigGroups` field. This field allows you to define as many node configurations as you want.
Once a `NodeConfigGroup` has been defined, you can define it with your node declaration to say "I want to add a new node with this type of configuration".

The main purpose of a [NodeConfigGroup] is to define the purely technical requirements for the pod that will be deployed (storage configurations, resource requirements, docker image, pod location, etc).

For example, you can have this node group configuration : 

```yaml
  # nodeConfigGroups specifies multiple node configs with unique name
  nodeConfigGroups:
    default_group:
      # provenanceStorage allow to specify the maximum amount of data provenance information to store at a time
      # https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#write-ahead-provenance-repository-properties
      provenanceStorage: "10 GB"
      #RunAsUser define the id of the user to run in the Nifi image
      # +kubebuilder:validation:Minimum=1
      runAsUser: 1000
      serviceAccountName: "default"
      # resourceRequirements works exactly like Container resources, the user can specify the limit and the requests
      # through this property
      # https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
      resourcesRequirements:
        limits:
          cpu: "2"
          memory: 3Gi
        requests:
          cpu: "1"
          memory: 3Gi
    high_mem_group:
      # provenanceStorage allow to specify the maximum amount of data provenance information to store at a time
      # https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#write-ahead-provenance-repository-properties
      provenanceStorage: "10 GB"
      #RunAsUser define the id of the user to run in the Nifi image
      # +kubebuilder:validation:Minimum=1
      runAsUser: 1000
      serviceAccountName: "default"
      # resourceRequirements works exactly like Container resources, the user can specify the limit and the requests
      # through this property
      # https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
      resourcesRequirements:
        limits:
          cpu: "2"
          memory: 30Gi
        requests:
          cpu: "1"
          memory: 30Gi
```

In this example, we have defined two different node configurations:
- `default_group`: Saying that we want **3Gi** of RAM for a node using this configuration.
- `high_mem_group`: Saying that we want **30Gi** of RAM for a node using this configuration.

- Now we can declare the nodes of our cluster using this configuration: 

```yaml
- id: 0
  nodeConfigGroup: "default_group"
- id: 2
  nodeConfigGroup: "high_mem_group"
- id: 3
  nodeConfigGroup: "high_mem_group"
- id: 5
  nodeConfig:
    resourcesRequirements:
      limits:
        cpu: "2"
        memory: 3Gi
      requests:
        cpu: "1"
        memory: 1Gi
```

In this example, you can see that we have defined one node using the node configuration `default_group` (id = 0), 2 nodes using `high_mem_group` (id = 2 and 3) and you also have the possibility to define the node configuration group directly at the node level (not reusable by another node) like for node 5.

#### Storage management

One of the most important configurations for a node in the case of a NiFi cluster is data persistence. As we run on kubernetes, whenever a pod is deleted, all data that is not stored on persistent storage will be lost, which is really bad for a statefull application like NiFi.
To avoid this, you can define how and what you want to persist in NiFi in the [NodeConfigGroup].

##### Data persistence

The first way to define data persistence is to use the [Spec.NodeConfigGroup.StorageConfig](../../../5_references/1_nifi_cluster/3_node_config#storageconfig) field.

This field allows you to define a storage set giving:
- `name`: a unique name to identify the storage config
- `metadata`: labels and annotations to attach to the PVC getting created.
- `pvcSpec` : a Kubernetes PVC spec definition
- `mountPath` : the path where the volume will be mounted into the main nifi container inside the pod (i.e the path were you want the data to be persisted).

:::note
If you don't replace them in the `nifi.properties` file using [NiFi-configurations](#nifi-configurations), here is a list of paths that should be associated with a storage configuration:
- `/opt/nifi/data` : contains 
  - `/opt/nifi/data/flow.xml.gz`: flow configuration files
  - `/opt/nifi/data/archive`: NiFi archive
  - `/opt/nifi/data/templates`: templates directory
  - `/opt/nifi/data/database_repository`: Database persistence
- `/opt/nifi/nifi-current/logs`: NiFi logs files
- `/opt/nifi/flowfile_repository`: flowfiles repository
- `/opt/nifi/content_repository`: NiFi content repository
- `/opt/nifi/provenance_repository`: NiFi provenance repository
- `/opt/nifi/nifi-current/conf`: NiFi configurations
:::

Here is an example we use in production for to persist data:

```yaml
...
storageConfigs:
  - mountPath: /opt/nifi/nifi-current/logs
    name: logs
    reclaimPolicy: Delete
    metadata:
      labels:
        my-label: my-value
      annotations:
        my-annotation: my-value
    pvcSpec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 100Gi
      storageClassName: ssd-wait
  - mountPath: /opt/nifi/data
    name: data
    reclaimPolicy: Delete
    metadata:
      labels:
        my-label: my-value
      annotations:
        my-annotation: my-value
    pvcSpec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 50Gi
      storageClassName: ssd-wait
  - mountPath: /opt/nifi/extensions
    name: extensions-repository
    reclaimPolicy: Delete
    metadata:
      labels:
        my-label: my-value
      annotations:
        my-annotation: my-value
    pvcSpec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 5Gi
      storageClassName: ssd-wait
  - mountPath: /opt/nifi/flowfile_repository
    name: flowfile-repository
    reclaimPolicy: Delete
    metadata:
      labels:
        my-label: my-value
      annotations:
        my-annotation: my-value
    pvcSpec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 100Gi
      storageClassName: ssd-wait
  - mountPath: /opt/nifi/nifi-current/conf
    name: conf
    reclaimPolicy: Delete
    metadata:
      labels:
        my-label: my-value
      annotations:
        my-annotation: my-value
    pvcSpec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 5Gi
      storageClassName: ssd-wait
  - mountPath: /opt/nifi/content_repository
    name: content-repository
    reclaimPolicy: Delete
    metadata:
      labels:
        my-label: my-value
      annotations:
        my-annotation: my-value
    pvcSpec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 500Gi
      storageClassName: ssd-wait
  - mountPath: /opt/nifi/provenance_repository
    name: provenance-repository
    reclaimPolicy: Delete
    metadata:
      labels:
        my-label: my-value
      annotations:
        my-annotation: my-value
    pvcSpec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 500Gi
```

##### External volumes

In some cases, you may want to mount a volume that is not managed by the operator to add a configuration or persist data using an existing volume. To do this, use the [Spec.NodeConfigGroup.StorageConfig](../../../5_references/1_nifi_cluster/3_node_config#externalvolumeconfig) field.

:::info
For a complete overview of node configuration possibilities, please refer to the reference page [NodeConfigGroup]
:::

## NiFi configurations

Once you have correctly defined the pods that will be deployed for your NiFi cluster, you may still have some configuration to do but at the NiFi level this time!
For this, the field to configure is [ReadOnlyConfig] which can be used at the global level `Spec.ReadOnlyConfig` or at the node level like for `NodeConfigGroup`.

There is some configuration that can be passed directly into this field like : 
- **maximumTimerDrivenThreadCount**: define the maximum number of threads for timer driven processors available to the system.
- **maximumEventDrivenThreadCount**: define the maximum number of threads for event driven processors available to the system.

And other configurations (e.g. configuration files) that can be defined using `kubernetes Secret`, `ConfigMap` or directly using the `override` field.

### ConfigMap, Secret, Inline

For most of the configuration files that can be overwritten (see section below), there are 4 ways to define the configuration:
- `default`: if nothing is specified, a default configuration is defined by the operator and will be used as is.
- `kubernetes secret`: you reference a data key in a secret that will contain your configuration (used to define sensitive configurations like client secret, password, etc.)
- `kubernetes configMap`: you reference a data key in a configMap that will contain your configuration.
- `override field`: you define the configuration directly in the `NiFiCluster` object using a string.

When more than one configuration type is defined, the following priority is applied when the same configuration field is defined more than once: `Secret` > `ConfigMap` > `Override` > `Default`, which follow the security priority.

Let's take an example : 

```yaml
 nifiProperties:
      # Additionnals nifi.properties configuration that will override the one produced based on template and
      # configuration
      overrideConfigMap:
        # The key of the value,in data content, that we want use.
        data: nifi.properties
        # Name of the configmap that we want to refer.
        name: raw
        # Namespace where is located the secret that we want to refer.
        namespace: nifikop
      # Additionnals nifi.properties configuration that will override the one produced based
      #	on template, configurations, overrideConfigMap and overrideConfigs.
      overrideSecretConfig:
        # The key of the value,in data content, that we want use.
        data: nifi.properties
        # Name of the configmap that we want to refer.
        name: raw
        # Namespace where is located the secret that we want to refer.
        namespace: nifikop
      # Additionnals nifi.properties configuration that will override the one produced based
      #	on template, configurations and overrideConfigMap
      overrideConfigs: |
        nifi.ui.banner.text=NiFiKop
        nifi.sensitive.props.key=thisIsABadSensitiveKeyPassword
```

In this example if we have the `nifi.sensitive.props.key` key defined in the secret `raw`, it will override the one defined in `overrideConfigs` field.

### Overridable configurations

Here is the list of configuration that you can override for NiFi :
- [nifi.properties](https://github.com/konpyutaika/nifikop/blob/master/pkg/resources/templates/config/nifi_properties.go)
- [zookeeper.properties](https://github.com/konpyutaika/nifikop/blob/master/pkg/resources/templates/config/zookeeper_properties.go)
- [bootstrap.properties](https://github.com/konpyutaika/nifikop/blob/master/pkg/resources/templates/config/bootstrap_properties.go)
- [logback.xml](https://github.com/konpyutaika/nifikop/blob/master/pkg/resources/templates/config/logback.xml.go)
- [authorizers.xml](https://github.com/konpyutaika/nifikop/blob/master/pkg/resources/templates/config/authorizers.go)
- [bootstrap_notification_services.xml](https://github.com/konpyutaika/nifikop/blob/master/pkg/resources/templates/config/bootstrap_notifications_services.go)

:::warning
Keep in mind that some changes to the default configuration may cause the operator's behavior to break, so keep that in mind!
Just because it's allowed doesn't mean it works :)
:::

## Advanced configuration

In some cases, using the default content or provenance configuration for storage may not be sufficient, for example you may need to create multiple directories for your content or provenance repository in order to [set up a high performance installation](https://community.cloudera.com/t5/Community-Articles/HDF-CFM-NIFI-Best-practices-for-setting-up-a-high/ta-p/244999).
As described in the NiFi Administration Guide, you can do this by using the [nifi.content.repository.directory.default*](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#file-system-content-repository-properties) and [nifi.provenance.repository.directory.default*](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#write-ahead-provenance-repository-properties) properties.

Here is an example of how to do this in the `NiFiCluster` configuration:

```yaml
...
  readOnlyConfig:
    nifiProperties:
      overrideConfigs: |
        nifi.content.repository.directory.dir1=../content-additional/dir1
        nifi.content.repository.directory.dir2=../content-additional/dir2
        nifi.content.repository.directory.dir3=../content-additional/dir3
        nifi.provenance.repository.directory.dir1=../provenance-additional/dir1
        nifi.provenance.repository.directory.dir2=../provenance-additional/dir2
...
  nodeConfigGroups:
    default_group:
      ...
      storageConfigs:
      - mountPath: "/opt/nifi/content-additional/dir1"
        name: content-repository-dir1
        metadata:
          labels:
            my-label: my-value
          annotations:
            my-annotation: my-value
        pvcSpec:
          accessModes:
            - ReadWriteOnce
          storageClassName: {{ storageClassName }}
          resources:
            requests:
              storage: 100G
      - mountPath: "/opt/nifi/content-additional/dir2"
        name: content-repository-dir2
        metadata:
          labels:
            my-label: my-value
          annotations:
            my-annotation: my-value
        pvcSpec:
          accessModes:
            - ReadWriteOnce
          storageClassName: {{ storageClassName }}
          resources:
            requests:
              storage: 100G
      - mountPath: "/opt/nifi/content-additional/dir3"
        name: content-repository-dir3
        metadata:
          labels:
            my-label: my-value
          annotations:
            my-annotation: my-value
        pvcSpec:
          accessModes:
            - ReadWriteOnce
          storageClassName: {{ storageClassName }}
          resources:
            requests:
              storage: 100G
      - mountPath: "/opt/nifi/provenance-additional/dir1"
        name: provenance-repository-dir1
        metadata:
          labels:
            my-label: my-value
          annotations:
            my-annotation: my-value
        pvcSpec:
          accessModes:
            - ReadWriteOnce
          storageClassName: {{ storageClassName }}
          resources:
            requests:
              storage: 100G
      - mountPath: "/opt/nifi/provenance-additional/dir2"
        name: provenance-repository-dir2
        metadata:
          labels:
            my-label: my-value
          annotations:
            my-annotation: my-value
        pvcSpec:
          accessModes:
            - ReadWriteOnce
          storageClassName: {{ storageClassName }}
          resources:
            requests:
              storage: 100G
      ...
```


[NodeConfigGroup]: ../../../5_references/1_nifi_cluster/3_node_config
[ReadOnlyConfig]: ../../../5_references/1_nifi_cluster/2_read_only_config