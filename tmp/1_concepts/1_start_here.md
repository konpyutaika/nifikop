---
id: 1_start_here
title: Start here
sidebar_label: Start here
---

The Konpyūtāika NiFi operator is a Kubernetes operator to automate provisioning, management, autoscaling and operations of [Apache NiFi](https://nifi.apache.org/) clusters deployed to K8s.

## Overview

Apache NiFi is an open-source solution that support powerful and scalable directed graphs of data routing, transformation, and system mediation logic. 
Some of the high-level capabilities and objectives of Apache NiFi include, and some of the main features of the **NiFiKop** are:

- **Fine grained** node configuration support
- Graceful rolling upgrade
- graceful NiFi cluster **scaling**
- NiFi cluster **auto-scaling** 
- Encrypted communication using SSL
- the provisioning of secure NiFi clusters
- Advanced Dataflow and user management via CRD

Some of the roadmap features :

- Automatic reaction and self healing based on alerts (plugin system, with meaningful default alert plugins)
- graceful NiFi cluster **rebalancing**

## Motivation

There are already some approaches to operating NiFi on Kubernetes, however, we did not find them appropriate for use in a highly dynamic environment, nor capable of meeting our needs.

- [Helm chart](https://github.com/cetic/helm-nifi)
- [Cloudera Nifi Operator](https://blog.cloudera.com/cloudera-flow-management-goes-cloud-native-with-apache-nifi-on-red-hat-openshift-kubernetes-platform/)

Finally, our motivation is to build an open source solution and a community which drives the innovation and features of this operator.