---
id: 1_quick_start
title: Quick start
sidebar_label: Quick start
---


import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

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

As a pre-requisite NiFi requires Zookeeper so you need to first have a Zookeeper cluster if you don't already have one.

> We believe in the `separation of concerns` principle, thus the NiFi operator does not install nor manage Zookeeper.

## Install Zookeeper

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

## Deploy NiFi cluster

And after you can deploy a simple NiFi cluster.

```bash
# Add your zookeeper svc name to the configuration
kubectl create -n nifi -f config/samples/simplenificluster.yaml
```