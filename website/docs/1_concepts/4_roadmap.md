---
id: 4_roadmap
title: Roadmap
sidebar_label: Roadmap
---

## NiFi cluster installation

|                       |           |
| --------------------- | --------- |
| Status                | Done      |
| Priority              | High      |
| Targeted Start date   | Jan 2020  |

## Graceful NiFi Cluster Scaling

|                       |           |
| --------------------- | --------- |
| Status                | Done      |
| Priority              | High      |
| Targeted Start date   | Jan 2020  |

Apache NiFi is a good candidate to create an operator, because everything is made to orchestrate it through REST Api calls. With this comes automation of actions such as scaling, following all required steps : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#decommission-nodes.

## Authentification management

|                       |       |
| --------------------- | ----- |
| Status                | To Do |
| Priority              | High  |
| Targeted Start date   | -     |


## Communication via SSL

|                       |       |
| --------------------- | ----- |
| Status                | To Do |
| Priority              | High  |
| Targeted Start date   | -     |


The operator fully automates NiFi's SSL support.
The operator can provision the required secrets and certificates for you, or you can provide your own.

## Monitoring via Prometheus

|                       |       |
| --------------------- | ----- |
| Status                | To Do |
| Priority              | High  |
| Targeted Start date   | -     |

The NiFi operator exposes NiFi JMX metrics to Prometheus.

## Dataflow management via CRD

|                       |           |
| --------------------- | --------- |
| Status                | To Do     |
| Priority              | Medium    |
| Targeted Start date   | -         |

## Reacting on Alerts

|                       |       |
| --------------------- | ----- |
| Status                | To Do |
| Priority              | Low   |
| Targeted Start date   | -     |

The NiFi Operator acts as a **Prometheus Alert Manager**. It receives alerts defined in Prometheus, and creates actions based on Prometheus alert annotations.

Currently, there are three actions expected :
- upscale cluster (add a new Node)
- downscale cluster (remove a Node)
- add additional disk to a Node

## Seamless Istio mesh support

|                       |       |
| --------------------- | ----- |
| Status                | To Do |
| Priority              | Low   |
| Targeted Start date   | -     |

- Operator allows to use ClusterIP services instead of Headless, which still works better in case of Service meshes.
- To avoid too early nifi initialization, which might lead to unready sidecar container. The operator will use a small script to
mitigate this behaviour. All NiFi image can be used the only one requirement is an available **curl** command.
- To access a NiFi cluster which runs inside the mesh. Operator will supports creating Istio ingress gateways.