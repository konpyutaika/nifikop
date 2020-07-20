---
id: 1_getting_started
title: Getting Started
sidebar_label: Getting Started
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

The operator installs the 1.11.4 version of Apache NiFi, and can run on Minikube v0.33.1+ and Kubernetes 1.12.0+.

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
helm install nifikop-zk bitnami/zookeeper \
    --set resources.requests.memory=256Mi \
    --set resources.requests.cpu=250m \
    --set resources.limits.memory=256Mi \
    --set resources.limits.cpu=250m \
    --set global.storageClass=local-storage \
    --set networkPolicy.enabled=true \
    --set replicaCount=3 
```

### Install cert-manager

The NiFiKop operator uses `cert-manager` for issuing certificates to users and and nodes, so you'll need to have it setup in case you want to deploy a secured cluster with authentication enabled.

<Tabs
  defaultValue="directly"
  values={[
    { label: 'Directly', value: 'directly', },
    { label: 'helm 3', value: 'helm3', },
    { label: 'helm previous', value: 'helm', },
  ]
}>
<TabItem value="directly">

```bash
# Install the CustomResourceDefinitions and cert-manager itself
kubectl apply -f \
    https://github.com/jetstack/cert-manager/releases/download/v0.15.1/cert-manager.yaml
```
</TabItem>
<TabItem value="helm3">

```bash
# Install CustomResourceDefinitions first
kubectl apply --validate=false -f \
   https://github.com/jetstack/cert-manager/releases/download/v0.15.1/cert-manager.crds.yaml

# Add the jetstack helm repo
helm repo add jetstack https://charts.jetstack.io
helm repo update

# You have to create the namespace before executing following command
helm install cert-manager \
    --namespace cert-manager \
    --version v0.15.1 jetstack/cert-manager
```

</TabItem>
<TabItem value="helm">

```bash
# Install CustomResourceDefinitions first
kubectl apply --validate=false -f \
    https://github.com/jetstack/cert-manager/releases/download/v0.15.1/cert-manager.crds.yaml

# Add the jetstack helm repo
helm repo add jetstack https://charts.jetstack.io
helm repo update

# Using previous versions of helm
helm install --name cert-manager \
    --namespace cert-manager \
    --version v0.15.1 \
    jetstack/cert-manager
```
</TabItem>
</Tabs>

## Installation

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


1. Set `KUBECONFIG` pointing towards your cluster
2. Run `make deploy` (deploys the operator in the current namespace into the cluster)
3. Set your NiFi configurations in a Kubernetes custom resource (sample: `config/samples/simplenificluster.yaml`) and run this command to deploy the NiFi components:

```bash
# Add your zookeeper svc name to the configuration
kubectl create -n nifi -f config/samples/simplenificluster.yaml
```

## Easy way: installing with Helm

Alternatively, if you are using Helm, you can deploy the operator using a Helm chart [Helm chart](https://github.com/Orange-OpenSource/nifikop/tree/master/helm):

> To install the an other version of the operator use `helm install --name=nifikop --namespace=nifi --set operator.image.tag=x.y.z orange-incubator/nifikop`

Deploy the NiFiKop crds : 

<Tabs
  defaultValue="k8s16+"
  values={[
    { label: 'k8s version 1.16+', value: 'k8s16+', },
    { label: 'k8s previous versions', value: 'k8sprev', },
  ]
}>
<TabItem value="k8s16+">

```bash
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/deploy/crds/nifi.orange.com_nificlusters_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/deploy/crds/nifi.orange.com_nifiusers_crd.yaml
```

</TabItem>
<TabItem value="k8sprev">

```bash
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/deploy/crds/v1beta1/nifi.orange.com_nificlusters_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/deploy/crds/v1beta1/nifi.orange.com_nifiusers_crd.yaml
```
</TabItem>
</Tabs>

Now deploy the helm chart :

```bash
helm repo add orange-incubator https://orange-kubernetes-charts-incubator.storage.googleapis.com/
```

<Tabs
  defaultValue="helm3"
  values={[
    { label: 'helm 3', value: 'helm3', },
    { label: 'helm previous', value: 'helm', },
  ]
}>
<TabItem value="helm3">

```bash
# You have to create the namespace before executing following command
helm install nifikop --namespace=nifi orange-incubator/nifikop
```

</TabItem>
<TabItem value="helm">

```bash
helm install --name=nifikop --namespace=nifi orange-incubator/nifikop
```
</TabItem>
</Tabs>

And after you can deploy a simple NiFi cluster.

```bash
# Add your zookeeper svc name to the configuration
kubectl create -n nifi -f config/samples/simplenificluster.yaml
```