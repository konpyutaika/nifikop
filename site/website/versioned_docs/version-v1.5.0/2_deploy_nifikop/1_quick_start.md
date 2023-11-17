---
id: 1_quick_start
title: Quick start
sidebar_label: Quick start
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

For information about versions compatibilty of the operator features with Kubernetes and Apache NiFi, let's have look of the [version compatibility page](../4_compatibility_versions)

## Getting Started

### Cluster Setup

For local testing we recommend following one of the following setup guides:

- [Docker Desktop (Mac)](https://docs.docker.com/desktop/kubernetes)
- [Minikube](https://minikube.sigs.k8s.io/docs/start)
  :::note
  Start Minikube with at least 4gb RAM with `minikube start --memory=4000`
  :::
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- For testing on GKE you can [create a cluster with the command line or the Cloud Console UI](https://cloud.google.com/kubernetes-engine/docs/how-to/creating-a-zonal-cluster).
- For testing on EKS you can [install eksctl](https://eksctl.io/introduction/) and run `eksctl create cluster` to create an EKS cluster/VPC/subnets/etc. This process should take 10-15 minutes.

### Install kubectl

If you do not already have the CLI tool `kubectl` installed, please follow [these instructions to install](https://kubernetes.io/docs/tasks/tools/).

### Configure kubectl

Configure `kubectl` to connect to your cluster by using `kubectl config use-context my-cluster-name`.

- For GKE
  - Configure `gcloud` with `gcloud auth login`.
  - On the Google Cloud Console, the cluster page will have a `Connect` button, which will give a command to run locally that looks like
  ```console
  gcloud container clusters get-credentials CLUSTER_NAME --zone ZONE_NAME --project PROJECT_NAME.
  ```
  - Use `kubectl config get-contexts` to show the contexts available.
  - Run `kubectl config use-context ${gke context}` to access the cluster from `kubectl`.
- For EKS 
  - [Configure your AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) to connect to your project. 
  - Install [eksctl](https://eksctl.io/introduction/) 
  - Run `eksctl utils write-kubeconfig --cluster=${CLUSTER NAME}` to make the context available to `kubectl` 
  - Use `kubectl config get-context`s to show the contexts available. 
  - Run `kubectl config use-context ${eks context}` to access the cluster with `kubectl`.
  
## Install cert-manager

The NiFiKop operator uses `cert-manager` for issuing certificates to users and and nodes, so you'll need to have it setup in case you want to deploy a secured cluster with authentication enabled. The minimum supported cert-manager version is v1.0.

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
    https://github.com/jetstack/cert-manager/releases/download/v1.7.2/cert-manager.yaml
```

</TabItem>
<TabItem value="helm3">

```bash
# Install CustomResourceDefinitions first
kubectl apply --validate=false -f \
   https://github.com/jetstack/cert-manager/releases/download/v1.7.2/cert-manager.crds.yaml

# Add the jetstack helm repo
helm repo add jetstack https://charts.jetstack.io
helm repo update

# You have to create the namespace before executing following command
helm install cert-manager \
    --namespace cert-manager \
    --version v1.7.2 jetstack/cert-manager
```

</TabItem>
</Tabs>

## Deploy NiFiKop

You can deploy the operator using a Helm chart [Helm chart](https://github.com/konpyutaika/nifikop/tree/master/helm):

> To install an other version of the operator use `helm install --name=nifikop --namespace=nifi --set operator.image.tag=x.y.z konpyutaika-incubator/nifikop`

In the case where you don't want to deploy the crds using helm (`--skip-crds`), you have to deploy manually the crds :

```bash
kubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nificlusters.yaml
kubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiusers.yaml
kubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiusergroups.yaml
kubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifidataflows.yaml
kubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiparametercontexts.yaml
kubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiregistryclients.yaml
```

:::important Conversion webhook
In case you keep the conversion webhook enabled (to handle the conversion of resources from `v1alpha1` to `v1`)
you will need to add the following settings to your yaml definition of CRDs:

```yaml
...
annotations:
    cert-manager.io/inject-ca-from: ${namespace}/${certificate_name}
...
spec:
  ...
  conversion:
    strategy: Webhook
    webhook:
      clientConfig:
        service:
          namespace: ${namespace}
          name: ${webhook_service_name}
          path: /convert
      conversionReviewVersions:
        - v1
        - v1alpha1
  ...
```

Where :
- `namespace`: is the namespace in which you will deploy your helm chart.
- `certificate_name`: is `${helm release name}-webhook-cert`
- `webhook_service_name`: is `${helm release name}-webhook-cert`
:::

Now deploy the helm chart :

```bash
# You have to create the namespace before executing following command
helm install nifikop \
    oci://ghcr.io/konpyutaika/helm-charts/nifikop \
    --namespace=nifi \
    --version 1.5.0 \
    --set image.tag=v1.5.0-release \
    --set resources.requests.memory=256Mi \
    --set resources.requests.cpu=250m \
    --set resources.limits.memory=256Mi \
    --set resources.limits.cpu=250m \
    --set namespaces={"nifi"}
```


:::note
Add the following parameter if you are using this instance to only deploy unsecured clusters : `--set certManager.enabled=false`
:::

### On OpenShift
The restricted SCC according to the DOC from OpenShift: "Denies access to all host features and requires pods to be run with a UID, and SELinux context that are allocated to the namespace." So in order to deploy NiFiKop on OpenShift we need to get the openshift.io/sa.scc.uid-range annotation of the namespace that we will deploy NiFiKop into. 

Get the uid for the nifi namespace:
```bash
uid=$(kubectl get namespace nifi -o=jsonpath='{.metadata.annotations.openshift\.io/sa\.scc\.supplemental-groups}' | sed 's/\/10000$//' | tr -d '[:space:]')
```
Set RunAsUser on install with helm:
```bash
helm install nifikop \
    nifikop \
    --namespace=nifi \
    --version 1.1.1 \
    --set image.tag=v1.1.1-release \
    --set resources.requests.memory=256Mi \
    --set resources.requests.cpu=250m \
    --set resources.limits.memory=256Mi \
    --set resources.limits.cpu=250m \
    --set namespaces={"nifi"} \
    --set runAsUser=$uid
```
