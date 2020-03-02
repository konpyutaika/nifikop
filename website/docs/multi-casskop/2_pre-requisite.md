---
id: 2_pre-requisite
title: Pre-requisite
sidebar_label: Pre-requisite
---

In order to have a working Multi-CassKop operator we need to have at least 2 k8s clusters: k8s-cluster-1 and k8s-cluster-2

- k8s >=v1.15 installed on each site, with kubectl configure to access both of thems
- The pods of each site must be able to reach pods on other sites, this is outside of the scope of Multi-Casskop and can
  be achieve by different solutions such as:
  - in our on-premise cluster, we leverage [Calico](https://www.projectcalico.org/why-bgp/) routable IP pool in order to make this possible
  - this can also be done using mesh service such as istio
  - there may be other solutions as well
- having casskop installed (With its ConfigMap) in each namespace see [CassKop installation](#install-casskop)
- having [External-DNS](https://github.com/kubernetes-sigs/external-dns) with RFC2136 installed in each namespace to
  manage your DNS sub zone. see [Install external dns](#install-external-dns)
- You need to create secrets from targeted k8s clusters in current. see[Bootstrap](#bootstrap-api-access-to-k8s-cluster-2-from-k8s-cluster-1)
- You may need to create network policies for Multi-Casskop inter-site communications to k8s apis, if using so.

> /!\ We have only tested t/.he configuration with Calico routable IP pool & external DNS with RFC2136 configuration.

## Bootstrap API access to k8s-cluster-2 from k8s-cluster-1

Multi-Casskop will be deployed in k8s-cluster-1, change your kubectl context to point to this cluster.

In order to allow our Multi-CassKop controller to have access to k8s-cluster-2 from k8s-cluster-1, we are going to use
[kubemcsa](https://github.com/admiraltyio/multicluster-service-account/releases/tag/v0.6.1) from
[Admiralty](https://admiralty.io/) to be able to export secret from k8s-cluster-2 to k8s-cluster1

```sh
kubemcsa export --context=cluster2 --namespace cassandra-e2e cassandra-operator --as k8s-cluster2 | kubectl apply -f -
```

> This will create in current k8s cluster which must be k8s-cluster-1, the k8s secret associated to the
> **cassandra-operator** service account of namespace **cassandra-e2e** in k8s-cluster2.
> /!\ The Secret will be created with the name **k8s-cluster2** and this name must be used when starting Multi-CassKop and
> in the MultiuCssKop CRD definition see below

This Diagram show how each component is connected:

![Multi-CassKop](../../img/multi-casskop/multi-casskop.jpg)

MultiCassKop starts by iterrating on every contexts passed in parameters then it register the controller. 
The controller needs to be able to interract with MultiCasskop and CassandraCluster CRD objetcs.
In addition the controller needs to watch for MultiCasskop as it will need to react on any changes that occured on
thoses objects for the given namespace.

## Install CassKop

CassKop must be deployed on each targeted Kubernetes clusters.

Add the Helm repository for CassKop

```console
$ helm repo add casskop https://Orange-OpenSource.github.io/casskop/helm
$ helm repo update
```

Connect to each kubernetes you want to deploy your Cassandra clusters to and install CassKop:

```console
$ helm install --name casskop casskop/cassandra-operator
```

## Install External-DNS

[External-DNS](https://github.com/kubernetes-sigs/external-dns) must be installed in each Kubernetes clusters.
Configure your external DNS with a custom values pointing to your zone and deploy it in your namespace 

```console
helm install -f /private/externaldns-values.yaml --name casskop-dns external-dns 
```

## Install Multi-CassKop

Proceed with Multi-CassKop installation only when [Pre-requisites](#pre-requisites) are fulfilled.

Deployment with Helm. Multi-CassKop and CassKop shared the same github/helm repo and semantic version.

```sh
helm install --name multi-casskop casskop/multi-casskop --set k8s.local=k8s-cluster1 --set k8s.remote={k8s-cluster2}
```

> if you get an error complaining that the CRD already exists, then replay it with `--no-hooks`

When starting Multi-CassKop, we need to give some parameters:

- k8s.local is the name of the k8s-cluster we want to refere to when talking to this cluster.
- k8s.remote is a list of other kubernetes we want to connect to.

> Names used there should map with the name used in the MultiCassKop CRD definition)
> the Names in `k8s.remote` must match the names of the secret exported with the [kubemcsa](#bootstrap-api-access-to-k8s-cluster-2-from-k8s-cluster-1) command

When starting, our MultiCassKop controller should log something similar to:

```log
time="2019-11-28T14:51:57Z" level=info msg="Configuring Client 1 for local cluster k8s-cluster1 (first in arg list). using local k8s api access"
time="2019-11-28T14:51:57Z" level=info msg="Configuring Client 2 for distant cluster k8s-cluster2. using imported secret of same name"
time="2019-11-28T14:51:57Z" level=info msg="Creating Controller"
time="2019-11-28T14:51:57Z" level=info msg="Create Client 1 for Cluster k8s-cluster1"
time="2019-11-28T14:51:57Z" level=info msg="Add CRDs to Cluster k8s-cluster1 Scheme"
time="2019-11-28T14:51:57Z" level=info msg="Create Client 2 for Cluster k8s-cluster2"
time="2019-11-28T14:51:58Z" level=info msg="Add CRDs to Cluster k8s-cluster2 Scheme"
time="2019-11-28T14:51:58Z" level=info msg="Configuring Watch for MultiCasskop"
time="2019-11-28T14:51:58Z" level=info msg="Configuring Watch for MultiCasskop"
time="2019-11-28T14:51:58Z" level=info msg="Writing ready file."
time="2019-11-28T14:51:58Z" level=info msg="Starting Manager."
```

We see it successfully created a k8s client for each of our cluster.
Then it do nothing, it is waiting for MultiCassKop objects.