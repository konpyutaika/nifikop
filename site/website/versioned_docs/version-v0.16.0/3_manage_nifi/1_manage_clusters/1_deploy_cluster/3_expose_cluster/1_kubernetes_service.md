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

:::important
Really important thing, you can add additional `internal listeners` without type, it means that these listeners are not related to internal NiFi behaviour. 
It might be really useful, if you are exposing a NiFi processor through a port (e.g http endpoint to receive HTTP request inside of NiFi) :

```yaml
listenersConfig:
  internalListeners:
    ...
    - name: "http-tracking"
      containerPort: 8081
```
:::

## Headless vs All service mode

To internally expose the NiFi cluster nodes, there is one major constraint which is that each node must be accessible one by one by the controller and the other nodes!

To do this, a single traditional Kubernetes service will not suffice, as it will balance the traffic between all nodes, which is not what we want.

There are two ways to solve this problem:
- **Use a [headless service](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services)**: this is the most appropriate way to expose your nodes internally, using this method you will simply deploy a service that will allow you to access each pod one by one via DNS resolution.
- **Use one service per node**: this way we create one service per node, giving you one cluster IP per node to access a single node directly!

To configure how you want to expose your NiFi node internally, you simply set the `Spec.HeadlessEnabled` field, if true you will be in the first mode, if not in the second.

:::note
In most cases, the `headless mode` should be used. An example where you need the other mode would be integration with Istio service mesh, which does not support headless service integration.
:::


## External service configuration

Once you have considered how to expose your service internally in the k8s cluster, you may want to expose your cluster externally, for example to give access to your cluster to your users, or to expose your prometheus endpoint.
To configure the external exposure of your cluster pods, you should use the [Spec.ExternalServices](../../../../5_references/1_nifi_cluster/7_external_service_config) field.

It takes as configuration:
- `name`: which will be used to name the kubernetes service.
- `spec`:
    - `type`: how the service is exposed (following the definition of [ServiceType](https://godoc.org/k8s.io/api/core/v1#ServiceType))
    - `portConfigs` : a list of port configurations:
        - `port` : the port that will be used by the service to expose the associated `internal listener`.
        - `internalListernerName` : name of the `internal listener` to expose

If we take a concrete example:

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
    - name: "http-tracking"
      containerPort: 8081
externalServices:
  - name: cluster-access
    spec:
      portConfigs:
        - internalListenerName: https
          port: 443
        - internalListenerName: http-tracking
          port: 80
      type: LoadBalancer
```

Here, we expose our `internal listeners`: `https` and `http-tracking` through a Kubernetes service: `cluster-access`, with two different ports: `443` and `80`.
If you look at the services, you should see something like this.

```console
kubectl get services
NAME                TYPE           CLUSTER-IP    EXTERNAL-IP      PORT(S)                                AGE
cluster-access      LoadBalancer   10.88.21.98   35.180.241.132   443:30421/TCP,80:30521/TCP             20d
```

If you add the `external ip` in your `Spec.ReadOnlyConfig.NifiProperties.WebProxyHosts` field, you will be able to access your cluster by: `https://<external-ip>` and your NiFi processor http endpoint by: `http://<external-ip>`.

:::note
For any additional configuration please refer to [this page](../../../../5_references/1_nifi_cluster/7_external_service_config).
:::
