---
id: 1_getting_started
title: Getting Started
sidebar_label: Getting Started
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

The operator installs the 1.11.2 version of Apache NiFi, and can run on Minikube v0.33.1+ and Kubernetes 1.12.0+.

:::info
The operator supports NiFi 1.11.0+
:::

As a pre-requisite it needs a Kubernetes cluster. Also, NiFi requires Zookeeper so you need to first have a Zookeeper cluster if you don't already have one.

> We believe in the `separation of concerns` principle, thus the NiFi operator does not install nor manage Zookeeper.

### Install Zookeeper

To install Zookeeper we recommend using the [Pravega's Zookeeper Operator](https://github.com/pravega/zookeeper-operator).
You can deploy Zookeeper by using the Helm chart.

```bash
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com/
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
helm install zookeeper-operator \
    --namespace=zookeeper \
    --set image.repository=registry.gitlab.si.francetelecom.fr/dfyarchicloud/dfyarchicloud-registry/pravega/zookeeper-operator \
    --set image.tag=0.2.5 \
    banzaicloud-stable/zookeeper-operator
```

</TabItem>
<TabItem value="helm">

```bash
helm install --name zookeeper-operator \
    --namespace=zookeeper \
    --set image.repository=registry.gitlab.si.francetelecom.fr/dfyarchicloud/dfyarchicloud-registry/pravega/zookeeper-operator \
    --set image.tag=0.2.5 \
    banzaicloud-stable/zookeeper-operator
```
</TabItem>
</Tabs>

And after you can deploy a simple cluster, for example with three nodes.

```bash
kubectl create --namespace zookeeper -f - <<EOF
apiVersion: zookeeper.pravega.io/v1beta1
kind: ZookeeperCluster
metadata:
  name: zookeepercluster
  namespace: zookeeper
spec:
  image: 
    repository: registry.gitlab.si.francetelecom.fr/dfyarchicloud/dfyarchicloud-registry/pravega/zookeeper
  replicas: 3
  persistence:
    spec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 20Gi
      storageClassName: local-storage
EOF
```

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

Alternatively, if you are using Helm, you can deploy the operator using a Helm chart [Helm chart](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/tree/master/helm):

> To install the an other version of the operator use `helm install --name=nifikop --namespace=nifi --set operator.image.tag=x.y.z orange-incubator/nifikop`

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
kubectl apply -f https://raw.githubusercontent.com/erdrix/nifikop/master/deploy/crds/nifi.orange.com_nificlusters_crd.yaml
helm install nifikop \
    --namespace=nifi \
    orange-incubator/nifikop \
    --set image.tag=v0.0.1 \
    --set image.repository=registry.gitlab.si.francetelecom.fr/dfyarchicloud/dfyarchicloud-registry/nifikop
```

</TabItem>
<TabItem value="helm">

```bash
helm install --name=nifikop \
    --namespace=nifi \
    orange-incubator/nifikop \
    --set image.tag=v0.0.1-release \
    --set image.repository=registry.gitlab.si.francetelecom.fr/dfyarchicloud/dfyarchicloud-registry/nifikop
```
</TabItem>
</Tabs>

And after you can deploy a simple NiFi cluster.

```bash
# Add your zookeeper svc name to the configuration
kubectl create -n nifi -f config/samples/orange/simplenificluster.yaml
```