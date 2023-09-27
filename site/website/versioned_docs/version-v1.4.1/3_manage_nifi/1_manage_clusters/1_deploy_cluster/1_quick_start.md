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
helm install zookeeper bitnami/zookeeper \
    --namespace=zookeeper \
    --set resources.requests.memory=256Mi \
    --set resources.requests.cpu=250m \
    --set resources.limits.memory=256Mi \
    --set resources.limits.cpu=250m \
    --set global.storageClass=standard \
    --set networkPolicy.enabled=true \
    --set replicaCount=3 \
    --create-namespace
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

### On OpenShift
#### Install Zookeeper

We need to get the uid/gid for the RunAsUser and the fsGroup for the namespace we deploy zookeeper in.

Get the zookeeper allowed uid/gid.

```bash
zookeper_uid=$(kubectl get namespace zookeeper -o=jsonpath='{.metadata.annotations.openshift\.io/sa\.scc\.supplemental-groups}' | sed 's/\/10000$//' | tr -d '[:space:]')
```
Specify the runAsUser and fsGroup Parameter on install of zookeeper.

```bash
helm install zookeeper bitnami/zookeeper \
    --set resources.requests.memory=256Mi \
    --set resources.requests.cpu=250m \
    --set resources.limits.memory=256Mi \
    --set resources.limits.cpu=250m \
    --set global.storageClass=standard \
    --set networkPolicy.enabled=true \
    --set replicaCount=3 \
    --set containerSecurityContext.runAsUser=$zookeper_uid \
    --set podSecurityContext.fsGroup=$zookeper_uid
```

#### Deploy NiFi cluster

And after you can deploy a simple NiFi cluster.

```bash
# Add your zookeeper svc name to the configuration
kubectl create -n nifi -f config/samples/simplenificluster.yaml
### On OpenShift

We need to get the uid/gid for the RunAsUser and the fsGroup for the namespace we deploy our nificluster in.

```bash
uid=$(kubectl get namespace nifi -o=jsonpath='{.metadata.annotations.openshift\.io/sa\.scc\.supplemental-groups}' | sed 's/\/10000$//' | tr -d '[:space:]')
```

Then update the config/samples/openshift file with our uid value.

```bash
sed -i "s/1000690000/$uid/g" config/samples/openshift.yaml
```

And after you can deploy a simple NiFi cluster.

```bash
kubectl create -n nifi -f config/samples/openshift.yaml
```