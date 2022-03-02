---
id: 4_node
title: Node
sidebar_label: Node
---

Node defines the nifi node basic configuration

```yaml
    - id: 0
      # nodeConfigGroup can be used to ease the node configuration, if set only the id is required
      nodeConfigGroup: "default_group"
      # readOnlyConfig can be used to pass Nifi node config
      # which has type read-only these config changes will trigger rolling upgrade
      readOnlyConfig:
        nifiProperties:
          overrideConfigs: |
            nifi.ui.banner.text=NiFiKop by Orange - Node 0
      # node configuration
#       nodeConfig:
    - id: 2
      # readOnlyConfig can be used to pass Nifi node config
      # which has type read-only these config changes will trigger rolling upgrade
      readOnlyConfig:
        overrideConfigs: |
          nifi.ui.banner.text=NiFiKop by Orange - Node 2
      # node configuration
      nodeConfig:
        resourcesRequirements:
          limits:
            cpu: "2"
            memory: 3Gi
          requests:
            cpu: "1"
            memory: 1Gi
        storageConfigs:
          # Name of the storage config, used to name PV to reuse into sidecars for example.
          - name: provenance-repository
            # Path where the volume will be mount into the main nifi container inside the pod.
            mountPath: "/opt/nifi/provenance_repository"
            # Kubernetes PVC spec
            # https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#create-a-persistentvolumeclaim
            pvcSpec:
              accessModes:
                - ReadWriteOnce
              storageClassName: "standard"
              resources:
                requests:
                  storage: 8Gi
```

## Node

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|id|int32| unique Node id. |Yes| - |
|nodeConfigGroup|string|  can be used to ease the node configuration, if set only the id is required |No| "" |
|readOnlyConfig|[ReadOnlyConfig](./2_read_only_config.md)| readOnlyConfig can be used to pass Nifi node config which has type read-only these config changes will trigger rolling upgrade.| No | nil |
|nodeConfig|[NodeConfig](./3_node_config.md)| node configuration. |No| nil |

