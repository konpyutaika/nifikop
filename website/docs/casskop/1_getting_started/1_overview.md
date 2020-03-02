---
id: 1_overview
title: Overview
sidebar_label: Overview
---

The Cassandra Kubernetes Operator (CassKop) makes it easy to run Apache Cassandra on Kubernetes. Apache Cassandra is
a popular free, open-source distributed wide column store NoSQL database management system.

CassKop will allow to easily create and manage Rack aware Cassandra Clusters.

## What is a Kubernetes operator

Kubernetes Operators are first-class citizens of a Kubernetes cluster and are 
application-specific controllers that extends Kubernetes to create, configure, and manage instances of complex applications.

We have choosen to use a [Custom Resource Definition (CRD)](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/) 
which creates a new Object Kind named **CassandraCluster** in Kubernetes which allow us to :

- Store object and state directly into Kubernetes
- Manage declarative inputs to set up the Cluster
- Declarative Updates - performing workflow actions on an existing CassandraCluster is 
straightforward. Updating a cluster amounts to updating the required declarative attributes 
in the CRD Yaml with new data and re-applying the CRD using kubectl.
CassKop's diff-based reconciliation logic ensures that only the required changes are made to a CassandraCluster.
- CassKop monitors create/update events for the CRD and performs the required actions.
- CassKop runs in a loop and reacts to events as they happen to reconcile a desired 
state with the actual state of the system.

## CassKop is our cassandra operator

CassKop will define a new Kubernetes object named `CassandraCluster` which 
will be used to describe and instantiate a Cassandra Cluster in Kubernetes. Example: 
- [cluster definition](https://github.com/Orange-OpenSource/casskop/tree/master/samples/cassandracluster-pic-test-acceptance-3.yaml)

CassKop is a Kubernetes custom controller which will loop over events on 
`CassandraCluster` objects and reconcile with kubernetes resources needed to create a valid 
Cassandra Cluster deployment.

CassKop is listening only in the Kubernetes namespace it is deployed in, and is able to manage several Cassandra Clusters within this namespace. 

When receiving a CassandraCluster object, CassKop will start creating required Kubernetes
resources, such as Services, Statefulsets, and so on.

Every time the desired CassandraCluster resource is updated by the user, CassKop performs
corresponding updates on the Kubernetes resources, so that the Cassandra cluster reflects the state of the desired cluster resource. Such updates might trigger a rolling update of the pods.

Finally, when the desired resource is deleted, CassKop starts to undeploy the cluster and delete all related Kubernetes resources.