---
id: 3_node_config
title: Node configuration
sidebar_label: Node configuration
---

NodeConfig defines the node configuration

```yaml
   default_group:
      # provenanceStorage allow to specify the maximum amount of data provenance information to store at a time
      # https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#write-ahead-provenance-repository-properties
      provenanceStorage: "10 GB"
      #RunAsUser define the id of the user to run in the Nifi image
      # +kubebuilder:validation:Minimum=1
      runAsUser: 1000
      # Set this to true if the instance is a node in a cluster.
      # https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#basic-cluster-setup
      isNode: true
      # Additionnal metadata to merge to the pod associated
      podMetadata:
        annotations:
          node-annotation: "node-annotation-value"
        labels:
          node-label: "node-label-value"
      # Docker image used by the operator to create the node associated
      # https://hub.docker.com/r/apache/nifi/
#      image: "apache/nifi:1.11.2"
      # nodeAffinity can be specified, operator populates this value if new pvc added later to node
      # https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#node-affinity
#      nodeAffinity:
      # imagePullPolicy define the pull policy for NiFi cluster docker image
      imagePullPolicy: IfNotPresent
      # priorityClassName define the name of the priority class to be applied to these nodes
      priorityClassName: "example-priority-class-name"
      # externalVolumeConfigs specifies a list of volume to mount into the main container.
      externalVolumeConfigs:
        - name: example-volume
          mountPath: "/opt/nifi/example"
          secret:
            secretName: "raw-controller"
      # storageConfigs specifies the node related configs
      storageConfigs:
        # Name of the storage config, used to name PV to reuse into sidecars for example.
        - name: provenance-repository
          # Path where the volume will be mount into the main nifi container inside the pod.
          mountPath: "/opt/nifi/provenance_repository"
          # Metadata to attach to the PVC that gets created
          metadata:
            labels:
              my-label: my-value
            annotations:
              my-annotation: my-value
          # Kubernetes PVC spec
          # https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#create-a-persistentvolumeclaim
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            storageClassName: "standard"
            resources:
              requests:
                storage: 10Gi
        - mountPath: "/opt/nifi/nifi-current/logs"
          name: logs
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            storageClassName: "standard"
            resources:
              requests:
                storage: 10Gi
```

## NodeConfig

| Field                 | Type                                                                                         |Description|Required|Default|
|-----------------------|----------------------------------------------------------------------------------------------|-----------|--------|--------|
| provenanceStorage     | string                                                                                       |provenanceStorage allow to specify the maximum amount of data provenance information to store at a time: [write-ahead-provenance-repository-properties](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#write-ahead-provenance-repository-properties)|No|"8 GB"|
| runAsUser             | int64                                                                                        |define the id of the user to run in the Nifi image|No|1000|
| fsGroup               | int64                                                                                        |define the id of the group for each volumes in Nifi image|No|1000|
| isNode                | boolean                                                                                      |Set this to true if the instance is a node in a cluster: [basic-cluster-setup](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#basic-cluster-setup)|No|true|
| image                 | string                                                                                       | Docker image used by the operator to create the node associated. [Nifi docker registry](https://hub.docker.com/r/apache/nifi/)|No|""|
| imagePullPolicy       | [PullPolicy](https://godoc.org/k8s.io/api/core/v1#PullPolicy)                                | define the pull policy for NiFi cluster docker image.)|No|""|
| nodeAffinity          | string                                                                                       | operator populates this value if new pvc added later to node [node-affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#node-affinity)|No|nil|
| storageConfigs        | \[&nbsp;\][StorageConfig](#storageconfig)                                                        |specifies the node related configs.|No|nil|
| externalVolumeConfigs | \[&nbsp;\][ExternalVolumeConfig](#externalvolumeconfig)                                          |specifies a list of volume to mount into the main container.|No|nil|
| serviceAccountName    | string                                                                                       |specifies the serviceAccount used for this specific node.|No|"default"|
| resourcesRequirements | [ResourceRequirements](https://godoc.org/k8s.io/api/core/v1#ResourceRequirements)            | works exactly like Container resources, the user can specify the limit and the requests through this property [manage-compute-resources-container](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/).|No|nil|
| imagePullSecrets      | \[&nbsp;\][LocalObjectReference](https://godoc.org/k8s.io/api/core/v1#TypedLocalObjectReference) |specifies the secret to use when using private registry.|No|nil|
| nodeSelector          | map\[string\]string                                                                          |nodeSelector can be specified, which set the pod to fit on a node [nodeselector](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)|No|nil|
| tolerations           | \[&nbsp;\][Toleration](https://godoc.org/k8s.io/api/core/v1#Toleration)                          |tolerations can be specified, which set the pod's tolerations [taint-and-toleration](https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/#concepts).|No|nil|
| podMetadata           | [Metadata](#metadata)                                                                        |define additionnal metadata to merge to the Pod associated.|No|nil|
| hostAliases      | \[&nbsp;\][HostAlias](https://pkg.go.dev/k8s.io/api/core/v1#HostAlias) | A list of host aliases to include in each pod's /etc/hosts configuration in the scenario where DNS is not available.           | No       | \[&nbsp;\]       |
| priorityClassName     | string                                                                                       | Specify the name of the priority class to apply to pods created with this node config | No | nil|

## StorageConfig

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string|Name of the storage config, used to name PV to reuse into sidecars for example.|Yes| - |
|mountPath|string|Path where the volume will be mount into the main nifi container inside the pod.|Yes| - |
|metadata|[Metadata](#metadata)|Define additional metadata to merge to the PVC associated.|No| - |
|pvcSpec|[PersistentVolumeClaimSpec](https://godoc.org/k8s.io/api/core/v1#PersistentVolumeClaimSpec)|Kubernetes PVC spec. [create-a-persistentvolumeclaim](https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#create-a-persistentvolumeclaim).|Yes| - |

## ExternalVolumeConfig

| Field                                                             |Type| Description |Required|Default|
|-------------------------------------------------------------------|----|-------------|--------|--------|
|| [VolueMount](https://pkg.go.dev/k8s.io/api/core/v1#VolumeMount)   |describes a mounting of a Volume within a container.| Yes         | - |
|| [VolumeSource](https://pkg.go.dev/k8s.io/api/core/v1#VolumeSource) | VolumeSource represents the location and type of the mounted volume. | Yes         | - |

## Metadata

| Field                                                             |Type| Description |Required|Default|
|-------------------------------------------------------------------|----|-------------|--------|--------|
| annotations | map\[string\]string | Additionnal annotation to merge to the resource associated [annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/#syntax-and-character-set). |No|nil|
| labels  | map\[string\]string | Additionnal labels to merge to the resource associated [labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set).               |No|nil|
