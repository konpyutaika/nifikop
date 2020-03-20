---
id: 1_pic
title: Kubernetes@Pic
sidebar_label: Kubernetes@Pic
---

1. Follow the instructions of the [documentation](https://kubernetes.pages.gitlab.si.francetelecom.fr/deploy-k8s-rke-pic/) to access to Pic Kubernetes clusters.

2. Once logged on one of Pic's cluster, you can follow the [Getting started guide](/nifikop/docs/7_dfy_orange/1_getting_started).

:::important
If you want to use an other docker registry than [dfyarchicloud-registry](https://gitlab.si.francetelecom.fr/dfyarchicloud/dfyarchicloud-registry/container_registry), [Artifactory](https://artifactory.packages.install-os.multis.p.fti.net/webapp/#/packages/docker/?state=eyJxdWVyeSI6e319) or [Proxy Artifactory](https://ext-dockerio.artifactory.si.francetelecom.fr/webapp/#/packages/docker/?state=eyJxdWVyeSI6e319), you have to create a `docker-registry` secret as described [here](https://dfyarchicloud.app.cf.sph.hbx.geo.francetelecom.fr/kubernetes/gitlab/registry-authent-k8s/).
Once the secret is created, if the image is the operator one, you have to deploy it adding the following arguments to the helm command : 

- `--set image.imagePullSecrets.enabled=true`
- `--set image.imagePullSecrets.name=<docker registry secret name>`

If the images are the ones used by the `NifiCluster`, you have to add the following configuration into your `NifiCluster.spec` definition :

```yaml 
spec: 
...
  imagePullSecrets: 
    - <docker registry secret name 1>
    ...
    - <docker registry secret name n>
...
```
:::

3. When you successfully deployed your cluster, there is still a step needed to access to the Nifi UI. You have to create a `Traefik Ingress` which will expose your loadbalancer.

    1. First, you have to get the service name created by the operator for your cluster : 

    ```console 
    kubectl -n nifi get svc -l nifi_cr=<NifiCluster's name>
    NAME                  TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                                         AGE
    simplenifi            LoadBalancer   10.43.138.161   <pending>     8080:31970/TCP,6007:32297/TCP,10000:32617/TCP   14h
    simplenifi-headless   ClusterIP      None            <none>        8080/TCP,6007/TCP,10000/TCP                     14h
    ```

    2. You have to take the `LoadBalancer`. You're now able to deploy your ingress with following manifest :

    ```yaml
    apiVersion: extensions/v1beta1
    kind: Ingress
    metadata:
      name: demo-ingress
      namespace: nifi
    spec:
      rules:
        - host: <unique identifier>.dev.k8s.m1.orangeportails.net
          http:
            paths:
              - backend:
                  serviceName: <LoadBalancer's name>
                  servicePort: <http or https port configured as Internal listener in your NifiCluster>
                path: /
    ```

    :::tip
    You can find some examples in the [config/samples/orange folder](https://github.com/erdrix/nifikop/tree/master/config/samples/orange)
    :::