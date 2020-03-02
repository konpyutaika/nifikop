---
id: 1_overview
title: Overview
sidebar_label: Overview
---

CassKop introduces a new Custom Resource Definition, `CassandraCluster` which allows you to describe the Cassandra cluster you want to deploy.
This works fine within a single Kubernetes cluster, and can allow for instance to deploy Cassandra in a mono-region multi-AZ topology.

But for having more resilience with our Cassandra cluster, we want to be able to spread it on several regions. For doing this with Kubernetes, we need that our Cassandra to spread spread on top of different Kubernetes clusters, deployed independently on different regions.

We introduce [MultiCassKop](https://github.com/Orange-OpenSource/casskop/multi-casskop) a new operator that fits above CassKop. MultiCassKop is a new controler that will be in charge of creating `CassandraClusters` CRD objects in several different Kubernetes clusters and in a manner that all Cassandra nodes will be part of the same ring.

MultiCassKop uses a new custom resource definition, `MultiCasskop` which allows to specify:

- a base for the CassandraCluster object
- an override of this base object for each kubernetes cluster we want to deploy on.

Recap:
Multi-CassKop goal is to bring the ability to deploy a Cassandra cluster within different regions, each of them running
an independant Kubernetes cluster.
Multi-Casskop insure that the Cassandra nodes deployed by each local CassKop will be part of the same Cassandra ring by
managing a coherent creation of CassandraCluster objects from it's own MultiCasskop custom ressource.