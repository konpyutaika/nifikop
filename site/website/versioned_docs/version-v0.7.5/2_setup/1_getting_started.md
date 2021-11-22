---
id: 1_getting_started
title: Getting Started
sidebar_label: Getting Started
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

The operator installs the 1.12.1 version of Apache NiFi, can run on Minikube v0.33.1+ and **Kubernetes 1.16.0+**, and require **Helm 3**.

:::info
The operator supports NiFi 1.11.0+
:::

As a pre-requisite it needs a Kubernetes cluster. Also, NiFi requires Zookeeper so you need to first have a Zookeeper cluster if you don't already have one.

> We believe in the `separation of concerns` principle, thus the NiFi operator does not install nor manage Zookeeper.

## Prerequisites

### Install Zookeeper

To install Zookeeper we recommend using the [Bitnami's Zookeeper chart](https://github.com/bitnami/charts/tree/master/bitnami/zookeeper).

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
```

```bash
# You have to create the namespace before executing following command
helm install zookeeper bitnami/zookeeper \
    --set resources.requests.memory=256Mi \
    --set resources.requests.cpu=250m \
    --set resources.limits.memory=256Mi \
    --set resources.limits.cpu=250m \
    --set global.storageClass=standard \
    --set networkPolicy.enabled=true \
    --set replicaCount=3
```

:::warning
Replace the `storageClass` parameter value with your own.
:::

### Install cert-manager

The NiFiKop operator uses `cert-manager` for issuing certificates to users and and nodes, so you'll need to have it setup in case you want to deploy a secured cluster with authentication enabled.

<Tabs
defaultValue="directly"
values={[
{ label: 'Directly', value: 'directly', },
{ label: 'helm 3', value: 'helm3', },
]
}>
<TabItem value="directly">

```bash
# Install the CustomResourceDefinitions and cert-manager itself
kubectl apply -f \
    https://github.com/jetstack/cert-manager/releases/download/v1.2.0/cert-manager.yaml
```

</TabItem>
<TabItem value="helm3">

```bash
# Install CustomResourceDefinitions first
kubectl apply --validate=false -f \
   https://github.com/jetstack/cert-manager/releases/download/v1.2.0/cert-manager.crds.yaml

# Add the jetstack helm repo
helm repo add jetstack https://charts.jetstack.io
helm repo update

# You have to create the namespace before executing following command
helm install cert-manager \
    --namespace cert-manager \
    --version v1.2.0 jetstack/cert-manager
```

</TabItem>
</Tabs>

## Installation

## Installing with Helm

You can deploy the operator using a Helm chart [Helm chart](https://github.com/Orange-OpenSource/nifikop/tree/master/helm):

> To install an other version of the operator use `helm install --name=nifikop --namespace=nifi --set operator.image.tag=x.y.z orange-incubator/nifikop`

In the case where you don't want to deploy the crds using helm (`--skip-crds`), you have to deploy manually the crds :

```bash
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nificlusters.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nifiusers.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nifiusergroups.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nifidataflows.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nifiparametercontexts.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nifiregistryclients.yaml
```

Add the orange incubator repository :

```bash

helm repo add orange-incubator https://orange-kubernetes-charts-incubator.storage.googleapis.com/
```

Now deploy the helm chart :

```bash
# You have to create the namespace before executing following command
helm install nifikop \
    orange-incubator/nifikop \
    --namespace=nifi \
    --version 0.7.5 \
    --set image.tag=v0.7.5-release \
    --set resources.requests.memory=256Mi \
    --set resources.requests.cpu=250m \
    --set resources.limits.memory=256Mi \
    --set resources.limits.cpu=250m \
    --set namespaces={"nifi"}
```

:::note
Add the following parameter if you are using this instance to only deploy unsecured clusters : `--set certManager.enabled=false`
:::

## Create custom storage class

We recommend to use a **custom StorageClass** to leverage the volume binding mode `WaitForFirstConsumer`

```bash
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: exampleStorageclass
parameters:
  type: pd-standard
provisioner: kubernetes.io/gce-pd
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```

:::tip
Remember to set your NiFiCluster CR properly to use the newly created StorageClass.
:::

## Deploy NiFi cluster

And after you can deploy a simple NiFi cluster.

```bash
# Add your zookeeper svc name to the configuration
kubectl create -n nifi -f config/samples/simplenificluster.yaml
```
