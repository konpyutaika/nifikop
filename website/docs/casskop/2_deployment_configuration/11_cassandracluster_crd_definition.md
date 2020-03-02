---
id: 11_cassandracluster_crd_definition
title: Cassandra cluster CRD definition
sidebar_label: Cassandra cluster CRD definition
---

The CRD Type is how we want to declare a CassandraCluster Object into Kubernetes.

To achieve this, we update the CRD to manage both :

- The new topology section
- The new CassandraRack

![architecture](http://www.plantuml.com/plantuml/proxy?src=https://raw.github.com/Orange-OpenSource/casskop/master/documentation/uml/crd.puml)