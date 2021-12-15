# Nifi with Keycloak using Nifikops

This is a full working example of setting up Keycloak (OIDC) with nifikops. 



## Prerequisites 

The following services/CRDs are required to be setup prior to this guide. This guide is also for EKS, so you can swap out EKS specifics for whatever cluster distro you are using. I will mark the EKS specifics when they arise.

* cert-manager (`namespace: cert-manager`)
* nginx-ingress (`namespace: nginx`)
* external-dns (`namespace: external-dns`)
* keycloak (`namespace: keycloak`)
* zookeeper (`namespace: nifi`)

We will use the variable `MY_DOMAIN` in place of my domain.

## Setup

In the `step-1` directory below, we have several manifests that need to be applied first.

* `lets-encrypt-issuer.yaml`: This configures our production cluster issuer that will add certs on all our ingress objects.
* `self-signed-issuer.yaml`, `self-signed-cert.yaml` and `nifi-issuer.yaml` are all required to give up https internal addresses (like `cluster.local`) to our cluster. This is required for OIDC integration.
* `namespace.yaml`: Where we intend to launch everything
* `operator.yaml`: The values.yaml for helm Since we are only in the nifi namespace for this example, we choose to only list it there. All the CRDs for the next step are applied here. THIS NEEDS TO BE APPLIED SEPARATELY FROM KUSTOMIZE (check out using flux for managing helm releases.)

In the `step-2` we can now apply our cluster and ingress.

* `ingress.yaml`: This communicates to the outside world our nifi cluster. If we were to have several, we would need different URLS (AFAIK). Notice two things, we use the annotation: `nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"` and the `service` called `nifi-cluster`. This is defined in the `cluster.yaml`.
* `cluster.yaml`: The main cluster. There are a couple notes. 
    - My zookeeper release has the service at `zookeeper:2181`.
    - Keycloak is hosted at `sso.MY_DOMAIN.com`
    - The client_id is not `abcdefghijklmnop123456789` but an actual id.
    - I have setup a realm in Keycloak called `MY_REALM` with a client `nifi` and a callback URL `https://nifi.MY_DOMAIN.com:443/nifi-api/access/oidc/callback`

