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

## State management

To manage its cluster and states, NiFi needs Zookeeper or rights on the Kubernetes cluster to manage `Leases` and `ConfigMaps` resources in the namespace where it is deployed.

In the case of Zookeeper, you must first have a Zookeeper cluster if you don't already have one.
Otherwise, you need to provide the corresponding role to the NiFi cluster's `ServiceAccount`.

> We believe in the `separation of concerns` principle, thus the NiFi operator does not install nor manage Zookeeper.

### Installing Zookeeper

To install Zookeeper we recommend using the [Bitnami's Zookeeper chart](https://github.com/bitnami/charts/tree/master/bitnami/zookeeper).

```bash
helm install zookeeper oci://registry-1.docker.io/bitnamicharts/zookeeper \
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

#### On OpenShift

We need to get the uid/gid for the RunAsUser and the fsGroup for the namespace we deploy zookeeper in.

Get the zookeeper allowed uid/gid.

```bash
zookeper_uid=$(kubectl get namespace zookeeper -o=jsonpath='{.metadata.annotations.openshift\.io/sa\.scc\.supplemental-groups}' | sed 's/\/10000$//' | tr -d '[:space:]')
```
Specify the runAsUser and fsGroup Parameter on install of zookeeper.

```bash
helm install zookeeper oci://registry-1.docker.io/bitnamicharts/zookeeper \
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

### Enabling Kubernetes State Management

When using native Kubernetes State Management from NiFi, you need to make sure that the `ServiceAccount` used by NiFi has the correct rights to manage the needed Kubernetes resources.

```yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: simplenifi
  namespace: nifi
rules:
- apiGroups: ["coordination.k8s.io"]
  resources: ["leases"]
  verbs: ["*"]
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["*"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: simplenifi
  namespace: nifi
subjects:
  - kind: ServiceAccount
    name: default
    namespace: nifi
roleRef:
  kind: Role
  name: simplenifi
  apiGroup: rbac.authorization.k8s.io
```

:::info
In this case, you need to set `clusterManager` in `NiFiCluster`'s specification to `kubernetes`.
You can also use the Helm chart to create your cluster and it will take care of it for you.
:::

## Deploy NiFi cluster

And after you can deploy a simple NiFi cluster.

```bash
# Add your zookeeper svc name to the configuration
kubectl create -n nifi -f config/samples/simplenificluster.yaml
```

### On OpenShift

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