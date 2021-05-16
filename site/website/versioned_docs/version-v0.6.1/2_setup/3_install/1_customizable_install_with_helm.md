---
id: 1_customizable_install_with_helm
title: Customizable install with Helm
sidebar_label: Customizable install with Helm
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Prerequisites

- Perform any necessary [plateform-specific setup](../2_platform_setup/1_gke.md)
- [Install a Helm client](https://github.com/helm/helm#install) with a version higher than 3

## Introduction

This Helm chart install NiFiKop the Orange's Nifi Kubernetes operator to create/configure/manage NiFi 
clusters in a Kubernetes Namespace.

It will use Custom Ressources Definition CRDs:
 
- `nificlusters.nifi.orange.com`, 
- `nifiusers.nifi.orange.com`,  
- `nifiusergroups.nifi.orange.com`, 
- `nifiregistryclients.nifi.orange.com`, 
- `nifiparametercontexts.nifi.orange.com`, 
- `nifidataflows.nifi.orange.com`, 

### Configuration

The following tables lists the configurable parameters of the NiFi Operator Helm chart and their default values.

| Parameter                        | Description                                      | Default                                   |
|----------------------------------|--------------------------------------------------|-------------------------------------------|
| `image.repository`               | Image                                            | `orangeopensource/nifikop`                |
| `image.tag`                      | Image tag                                        | `v0.6.1-release`                          |
| `image.pullPolicy`               | Image pull policy                                | `Always`                                  |
| `image.imagePullSecrets.enabled` | Enable tue use of secret for docker image        | `false`                                   |
| `image.imagePullSecrets.name`    | Name of the secret to connect to docker registry | -                                         |
| `certManager.enabled`            | Enable cert-manager integration                  | `true`                                    |
| `rbacEnable`                     | If true, create & use RBAC resources             | `true`                                    |
| `resources`                      | Pod resource requests & limits                   | `{}`                                      |
| `metricService`                  | deploy service for metrics                       | `false`                                   |
| `debug.enabled`                  | activate DEBUG log level                         | `false`                                   |
| `certManager.clusterScoped`      | If true setup cluster scoped resources           | `false`                            |
| `namespaces`                     | List of namespaces where Operator watches for custom resources. Make sure the operator ServiceAccount is granted `get` permissions on this `Node` resource when using limited RBACs.| `""` i.e. all namespaces |
| `nodeSelector`                   | Node selector configuration for operator pod     | `{}`                                      |
| `affinity`                       | Node affinity configuration for operator pod     | `{}`                                      |
| `tolerations`                    | Toleration configuration for operator pod        | `{}`                                      |
| `serviceAccount.create`          | Whether the SA creation is delegated to the chart or not       | `true`                                      |
| `serviceAccount.name`            | Name of the SA used for NiFiKop deployment       | release name                                     |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install nifikop \
      orange-incubator/nifikop \
      -f values.yaml
```

### Installing the Chart

:::important Skip CRDs
In the case where you don't want to deploy the crds using helm (`--skip-crds`) you need to deploy manually the crds beforehand:

```bash
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nificlusters.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nifiusers.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nifiusergroups.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nifidataflows.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nifiparametercontexts.yaml
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/nifikop/master/config/crd/bases/nifi.orange.com_nifiregistryclients.yaml
```

::: 

<Tabs
  defaultValue="dryrun"
  values={[
    { label: 'dry run', value: 'dryrun', },
    { label: 'release name', value: 'rn', },
    { label: 'set parameters', value: 'set-params', },
  ]
}>
<TabItem value="dryrun">

```bash
helm install nifikop orange-incubator/nifikop \
    --dry-run \
    --debug.enabled \
    --set debug.enabled=true \
    --set namespaces={"nifikop"}
```
</TabItem>
<TabItem value="rn">

```bash
helm install <release name> orange-incubator/nifikop 
```
</TabItem>

<TabItem value="set-params">

```bash
helm install nifikop orange-incubator/nifikop --set namespaces={"nifikop"}
```
</TabItem>
</Tabs>

> the `--replace` flag allow you to reuses a charts release name

### Listing deployed charts

```
helm list
```

### Get Status for the helm deployment

```
helm status nifikop
```

## Uninstaling the Charts

If you want to delete the operator from your Kubernetes cluster, the operator deployment 
should be deleted.

```
helm del nifikop
```

The command removes all the Kubernetes components associated with the chart and deletes the helm release.

:::tip
The CRD created by the chart are not removed by default and should be manually cleaned up (if required)
:::

Manually delete the CRD:

```
kubectl delete crd nificlusters.nifi.orange.com
kubectl delete crd nifiusers.nifi.orange.com
kubectl delete crd nifiusergroups.nifi.orange.com
kubectl delete crd nifiregistryclients.nifi.orange.com
kubectl delete crd nifiparametercontexts.nifi.orange.com
kubectl delete crd nifidataflows.nifi.orange.com
```

:::warning
If you delete the CRD then
It will delete **ALL** Clusters that has been created using this CRD!!!
Please never delete a CRD without very good care
:::

Helm always keeps records of what releases happened. Need to see the deleted releases ? 

```bash
helm list --deleted
```

Need to see all of the releases (deleted and currently deployed, as well as releases that
failed) ?

```bash 
helm list --all
```

Because Helm keeps records of deleted releases, a release name cannot be re-used. (If you really need to re-use a
release name, you can use the `--replace` flag, but it will simply re-use the existing release and replace its
resources.)

Note that because releases are preserved in this way, you can rollback a deleted resource, and have it re-activate.

To purge a release

```bash
helm delete --purge nifikop
```

## Troubleshooting

### Install of the CRD

By default, the chart will install the CRDs, but this installation is global for the whole
cluster, and you may want to not modify the already deployed CRDs.

In this case there is a parameter to say to not install the CRDs :

```
$ helm install --name nifikop ./helm/nifikop --set namespaces={"nifikop"} --skip-crds
```