---
id: 4_kubernetest_object
title: Kubernetes objects
sidebar_label: Kubernetes objects
---

## Services

Cassandra Pods will be accessible via Kubernetes headless services. CassKop will create a service for each
Cassandra DC define in the Topology section.

Service will be used by application to connect to the Cassandra Cluster.
Service will also be used for Cassandra to find others SEEDS nodes in the cluster.

## Statefulset

- **Statefulsets** is a powerful entity in Kubernetes to manage Pods, associated with some essential conventions :
    - Pod name: pods are created sequentially, starting with the name of the statefulset and ending with zero : 
    `<statefulset-name>-<ordinal-index>`. 
    - Network address: the statefulset uses a headless service to control the domain name of its pods. As each pod is
      created, it gets a matching DNS subdomain
    `<pod-name>.<service-name>.<namespace>`.

