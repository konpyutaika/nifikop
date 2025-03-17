---
id: 7_external_service_config
title: External Service Config
sidebar_label: External Service Config
---

ListenersConfig defines the Nifi listener types:

```yaml
  externalServices:
    - name: "clusterip"
      spec:
        type: ClusterIP
        portConfigs:
          - port: 8080
            internalListenerName: "http"
          - port: 7182
            internalListenerName: "my-custom-listener"
            protocol: TCP
      metadata:
        annotations:
          toto: tata
        labels:
          titi: tutu
```

Load balancer example:

```yaml
externalServices:
    - name: "nlb"
      spec:
        type: LoadBalancer
        loadBalancerClass: "service.k8s.aws/nlb"
        portConfigs:
          - port: 8080
            internalListenerName: "http"
          - port: 7890
            internalListenerName: "my-custom-udp-listener"
            protocol: UDP
      metadata:
        annotations:
          toto: tata
        labels:
          titi: tutu
```

## ExternalServiceConfig

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string| Must be unique within a namespace. Name is primarily intended for creation idempotence and configuration.| Yes | - |
|metadata|[Metadata](#metadata)| Defines additional metadata to merge with the associated service.| No | - |
|spec|[ExternalServiceSpec](#externalservicespec)| defines the behavior of a service.| Yes |  |

## ExternalServiceSpec

Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|portConfigs||\[&nbsp;\][PortConfig](#portconfig)| Contains the list port for the service and the associated listener| Yes | - |
|clusterIP|string| More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies | No | - |
|type|[ServiceType](https://godoc.org/k8s.io/api/core/v1#ServiceType)| type determines how the Service is exposed. Defaults to ClusterIP. Valid options are ExternalName, ClusterIP, NodePort, and LoadBalancer. | No | - |
|externalIPs|\[&nbsp;\]string| externalIPs is a list of IP addresses for which nodes in the cluster will also accept traffic for this service.  These IPs are not managed by Kubernetes | No | - |
|loadBalancerIP|string| Only applies to Service Type: LoadBalancer. LoadBalancer will get created with the IP specified in this field. | No | - |
|loadBalancerSourceRanges|\[&nbsp;\]string| If specified and supported by the platform, this will restrict traffic through the cloud-provider load-balancer will be restricted to the specified client IPs | No | - |
|externalName|string| externalName is the external reference that kubedns or equivalent will return as a CNAME record for this service. No proxying will be involved. | No | - |
|loadBalancerClass|string| loadBalancerClass is the class of the load balancer implementation this Service belongs to. | No | - |
|externalTrafficPolicy|string| See the Kubernetes [traffic policies](https://kubernetes.io/docs/reference/networking/virtual-ips/#traffic-policies) documentation. | No | Depends on the `Service` type. |
|internalTrafficPolicy|string| See the Kubernetes [traffic policies](https://kubernetes.io/docs/reference/networking/virtual-ips/#traffic-policies) documentation. | No | Depends on the `Service` type. |

## PortConfig

Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|port|int32| The port that will be exposed by this service. | Yes | - |
|internalListenerName|string| The name of the listener which will be used as target container. | Yes | - |
|nodePort|int32| The port that will expose this service externally. (Only if the service is of type NodePort) | No | - |
|protocol|[Protocol](https://pkg.go.dev/k8s.io/api/core/v1#Protocol)| The network protocol for this service port. Must be one of the protocol enum values (i.e. TCP, UDP, SCTP).  | No | `TCP` |

## Metadata

Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
| annotations | map\[string\]string | Additional annotations to merge with the associated service [annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/#syntax-and-character-set). | No | `nil` |
| labels  | map\[string\]string | Additional labels to merge with the associated service [labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set). | No | `nil` |
