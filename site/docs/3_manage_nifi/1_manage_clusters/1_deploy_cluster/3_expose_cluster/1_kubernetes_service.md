---
id: 1_kubernetes_service
title: Kubernetes service
sidebar_label: Kubernetes service
---

The purpose of this section is to explain how expose your NiFi cluster and access it in and out of kubernetes.

## Listener configuration

The first part to manage when you are configuring your cluster is the ports that will be used for the internal need of NiFi, we will call them `internal listeners`
There is 6 types of `internal listeners` : 
- **cluster**: Define the node’s protocol port (used by cluster nodes to discuss together).
- **http/https**: The HTTP or HTTPS port used to expose NiFi cluster UI (**Note**: use only one of them !)
- **s2s**: The remote input socket port for Site-to-Site communication
- **load-balance**: Cluster node load balancing port
- **prometheus**: Port that will be used to expose the promotheus reporting task

To configure these listeners you must use the [Spec.ListernersConfig.InternalListeners](../../../../5_references/1_nifi_cluster/6_listeners_config#internallistener) field : 

```yaml
listenersConfig:
  internalListeners:
    - type: "https"
      name: "https"
      containerPort: 8443
    - type: "cluster"
      name: "cluster"
      containerPort: 6007
    - type: "s2s"
      name: "s2s"
      containerPort: 10000
    - type: "prometheus"
      name: "prometheus"
      containerPort: 9090
    - type: "load-balance"
      name: "load-balance"
      containerPort: 6342
```

Here we defined a listener by specifying : 
- `type`: one of the six described above (e.g `https`)
- `name`: name of the port that will be attached to pod (e.g `https`)
- `containerPort`: port that will be used by the container inside the pod and configured in NiFi configuration for the listener (e.g `8443`)

If you look at the yaml description of the deployed pod, you should have something like this : 

```yaml
    ports:
    - containerPort: 8443
      name: https
      protocol: TCP
    - containerPort: 6007
      name: cluster
      protocol: TCP
    - containerPort: 6342
      name: load-balance
      protocol: TCP
    - containerPort: 10000
      name: s2s
      protocol: TCP
    - containerPort: 9090
      name: prometheus
      protocol: TCP
```

## Headless vs All service mode


## External service configuration

Sidecar