---
id: 1_cassandra_cluster_config
title: Cassandra cluster configuration
sidebar_label: Cassandra cluster configuration
---

The full schema of the `CassandraCluster` resource is described in the [Cassandra Cluster CRD Definition](#cassandra-cluster-crd-definition-version-020).
All labels that are applied to the desired `CassandraCluster` resource will also be applied to the Kubernetes resources
making up the Cassandra cluster. This provides a convenient mechanism for those resources to be labelled in whatever way
the user requires.