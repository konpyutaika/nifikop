---
id: 2_rickaastley
title: Rickaastley
sidebar_label: Rickaastley
---

NiFi requires Zookeeper so you first need to have a Zookeeper cluster if you don't already have one.
If you want to deploy a secured cluster you also have to install `cert-manager` to manage certificates.

For the `RicKaaStley` cluster, we will follow the secured requirements : 

- Dex instance
- Cert-manager instance
- External DNS and Traefik instance
- Zookeeper cluster

## Expose services

To expose our DEX and NiFi services, we will have to use an ingress, this requires to have :

- A calico routable pool IP, wich is accessible at least from VDI.
- A dns zone.

## External DNS 

The first step is to deploy the `external-dns` instance, for more information and configuration let's check the [documentation](https://hebex-wiki.orangeportails.net/index.php?title=RicKaaStley/Exposition_DNS_des_Services): 
 
```bash
helm repo add rickaastley https://artifactory.si.francetelecom.fr/virt-pfs-rickaastley-helm
```

```bash
helm install nifi-dns-rec-sph rickaastley/external-dns \
  -f config/samples/orange/rickaastley/external-dns/values.yaml
```

### Traefik

Once the `external-dns` is setup, we are able to instantiate our ingress-controller on routable ip pool :

```bash
helm repo add dfyarchicloud https://artifactory.packages.install-os.multis.p.fti.net/virt-sdfy-dfyarchicloud-helm
```

```bash
kubectl apply -f config/samples/orange/rickaastley/ingress/network-policies.yaml
helm install traefik \
       -f config/samples/orange/rickaastley/ingress/values.yaml \
       dfyarchicloud/traefik
```

:::note
To adapt your setup with your own configuration, let's have a look on the [full documentation's page](https://dfyarchicloud.app.cf.sph.hbx.geo.francetelecom.fr/kubernetes/ingress/traefik-metal-lb/#deploiement-de-traefik)
:::

## Zookeeper

To install Zookeeper we recommend using the [Bitnami's Zookeeper chart](https://github.com/bitnami/charts/tree/master/bitnami/zookeeper).

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
```

```bash
helm install zookeeper bitnami/zookeeper \
  --set global.imageRegistry=ext-dockerio.artifactory.si.francetelecom.fr \
  --set resources.requests.memory=256Mi \
  --set resources.requests.cpu=250m \
  --set resources.limits.memory=256Mi \
  --set resources.limits.cpu=250m \
  --set global.storageClass=local-storage \
  --set networkPolicy.enabled=true \
  --set clusterDomain=kaas-rec-sph.local \
  --set replicaCount=3
```

## Cert-manager

The NiFiKop operator uses `cert-manager` for issuing certificates to users and nodes, so you'll need to have it setup in case you want to deploy a secured cluster with enabled authentication.
To suit with the `RicKaaStley` cluster requirements, we manage our own `helm chart` : [cert-manager-rickaastley-compliant](https://gitlab.si.francetelecom.fr/kubernetes/cert-manager-rickaastley-compliant)

```bash
helm repo add dfy-cda-shared https://artifactory.packages.install-os.multis.p.fti.net/dfy-cda-shared-helm
```

```bash
#helm install cert-manager dfy-cda-shared/cert-manager-rickaastley-compliant \
helm install cert-manager https://artifactory.packages.install-os.multis.p.fti.net:443/dfy-cda-shared-helm/cert-manager-rickaastley-compliant-v0.1.0.tgz \
    --set global.clusterScoped=false \
    --set cainjector.enabled=false \
    --set webhook.enabled=false \
    --set image.tag=v0.15.1 \
    --set image.repository=ext-quayio.artifactory.si.francetelecom.fr/jetstack/cert-manager-controller \
    --set resources.requests.memory=256Mi \
    --set resources.requests.cpu=250m \
    --set resources.limits.memory=256Mi \
    --set resources.limits.cpu=250m
```

## Dex

As we still don't have an authentication system ready, let's a static DEX instance for now : 

```bash
helm repo add stable https://artifactory.si.francetelecom.fr/virt_helm_pfs-noh
```

```bash
kubectl create -f config/samples/orange/rickaastley/dex/network-policies.yaml
helm install dex \
    stable/dex \
    --set crd.present=true \
    --set rbac.create=false \
    --set image=ext-quayio.artifactory.packages.install-os.multis.p.fti.net/dexidp/dex \
    --set certs.image=ext-gcrio.artifactory.packages.install-os.multis.p.fti.net/google_containers/kubernetes-dashboard-init-amd64 \
    --set config.issuer=http://dex.nifi.pns.svc.rickaastley.p.fti.net \
    -f config/samples/orange/rickaastley/dex/values.yaml
kubectl create -f config/samples/orange/rickaastley/dex/role.yaml
```

## NiFi Cluster

### NiFiKop

```bash
helm repo add orange-incubator https://orange-kubernetes-charts-incubator.storage.googleapis.com/
```

```console
helm install nifikop \
    helm/nifikop \
    --set image.tag=0.1.0-ft_override-cluster-domain \
    --set image.repository=registry.gitlab.si.francetelecom.fr/dfyarchicloud/dfyarchicloud-registry/nifikop \
    --set resources.requests.memory=256Mi \
    --set resources.requests.cpu=250m \
    --set resources.limits.memory=256Mi \
    --set resources.limits.cpu=250m
```
    
### Network policies 

As all communications are denied by default, we have to declare the required ones by our NiFiCluster: 

```bash
kubectl apply -f config/samples/orange/rickaastley/network-policies/network-policies.yaml
```

### NiFi Cluster

Afterwards, you can deploy your NiFi cluster.

```bash
kubectl apply -f config/samples/orange/rickaastley/secured_nifi_cluster_dex.yaml
kubectl apply -f config/samples/orange/rickaastley/traefik-ingress.yaml
```