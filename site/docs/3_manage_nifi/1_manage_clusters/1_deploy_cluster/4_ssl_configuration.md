---
id: 4_ssl_configuration
title: SSL configuration
sidebar_label: SSL configuration
---

The `NiFi operator` makes securing your NiFi cluster with SSL easy. You may provide your own certificates, or instruct the operator to create them for you from your cluster configuration.

Below this is an example configuration required to secure your cluster with SSL:

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiCluster
...
spec:
  ...
  readOnlyConfig:
    # NifiProperties configuration that will be applied to the node.
    nifiProperties:
      webProxyHosts:
        - nifistandard2.trycatchlearn.fr:8443
        ...
  ...
  listenersConfig:
    internalListeners:
      - type: "https"
        name: "https"
        containerPort: 8443
      - type: "cluster"
        name: "cluster"
        containerPort: 6007
      - type: "s2s"
        name: "s2s"
        containerPort: 10000
    sslSecrets:
      tlsSecretName: "test-nifikop"
      create: true
```

- `readOnlyConfig.nifiProperties.webProxyHosts`: A list of allowed HTTP Host header values to consider when NiFi is running securely and will be receiving requests to a different host[:port] than it is bound to. [web-properties](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#web-properties)

If `listenersConfig.sslSecrets.create` is set to `false` (with the default `cert-manager` PKI backend), the operator will look for the secret at `listenersConfig.sslSecrets.tlsSecretName` and expect these values:

| key | value |
|------|-------|
| caCert | The CA certificate |
| caKey | The CA private key |

The operator uses this CA material to issue node and controller certificates via cert-manager.

## Without cert-manager: operator-managed PKI

By default the operator issues node and user certificates through cert-manager. On clusters without cert-manager, it can issue them itself instead.

The backend is chosen once, when the operator starts:

- `-cert-manager-enabled=true` (Helm `certManager.enabled: true`, the default): certificates are issued by cert-manager, as before.
- `-cert-manager-enabled=false` (Helm `certManager.enabled: false`): the operator issues certificates itself.
- Flag not passed: the operator asks the API server whether it serves cert-manager `Certificate`s and picks the matching backend. If that check fails, the operator assumes cert-manager is installed and logs an error.

Restart the operator if cert-manager is installed or removed afterwards. Moving an existing cluster from one backend to the other is not supported, because the other backend issues new certificates from a different CA. A `NifiCluster` that sets `listenersConfig.sslSecrets.pkiBackend: cert-manager` always uses cert-manager, whatever the operator was started with.

:::warning
Earlier versions issued certificates through cert-manager even when the operator ran with `-cert-manager-enabled=false`; the flag only turned off watching cert-manager `Certificate`s. If you run the operator that way on a cluster that has cert-manager, set `listenersConfig.sslSecrets.pkiBackend: cert-manager` on your clusters, or start the operator with `-cert-manager-enabled=true`, before upgrading. Otherwise their certificates are reissued from a new CA.
:::

With the operator-managed PKI:

- The CA is kept in the `<cluster name>-ca-certificate` secret (`ca.crt`, `tls.crt`, `tls.key`), owned by the `NifiCluster`. It is an RSA 4096 CA valid for 10 years.
- Each `NifiUser`, including the controller and node users, gets a secret with the same keys the cert-manager backend produces (`tls.crt`, `tls.key`, `ca.crt`, `keystore.jks`, `truststore.jks`, `password`), owned by that `NifiUser`. Pods mount them exactly as before.
- Certificates are valid for one year and are reissued two thirds of the way through their lifetime. The keystore password is kept across reissues.
- With `sslSecrets.create: false`, the secret named by `sslSecrets.tlsSecretName` supplies `caCert` and `caKey`, as described above. The operator copies the CA into `<cluster name>-ca-certificate` and never writes to your secret. The copy is deleted with the cluster; your secret is not. Your secret must not itself be named `<cluster name>-ca-certificate`. RSA and ECDSA keys are accepted, and `caCert` may include intermediate certificates. If your secret is deleted, the operator stops issuing certificates.
- `sslSecrets.issuerRef` names a cert-manager issuer, so it is rejected.
- When the cluster is deleted, the operator removes only the secrets it created.

:::tip
Set [`readOnlyConfig.nifiProperties.tlsAutoReload.enabled: true`](../../../5_references/1_nifi_cluster/2_read_only_config.md) so that NiFi picks up reissued certificates without a restart. Otherwise a reissued certificate is only used once the pod restarts, which has to happen before the previous certificate expires.
:::

The operator's own webhook also needs a serving certificate. Without cert-manager, provide one with `webhook.tls.mode: existingSecret` or disable the webhook with `webhook.enabled: false`.

## Using an existing Issuer

As described in the [Reference section](../../../5_references/1_nifi_cluster/6_listeners_config.md#sslsecrets), instead of using a self-signed certificate as CA, you can use an existing one.
In order to do so, you only have to refer it into your `Spec.ListenerConfig.SslSecrets.IssuerRef` field.

:::warning
When using [cert-manager Issuer](https://cert-manager.io/docs/concepts/issuer/), please make sure that the hostname 
(default `clusterName-id-node.namespace.svc.cluster.local`) for the nodes in cluster is 64 bytes or less, otherwise the webhook
of cert-manager will fail. You can try to use shorter name for NiFiCluster or modify `nodeUserIdentityTemplate` to keep 
the name length under 64 bytes.
:::

### Example: Let's Encrypt

Let's say you have an existing DNS server, with [external dns](https://github.com/kubernetes-sigs/external-dns) deployed into your cluster's namespace.
You can easily use [Let's Encrypt](https://letsencrypt.org/) as an authority for your certificate.

To do this, you have to:

1. Create an issuer:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # You must replace this email address with your own.
    # Let's Encrypt will use this to contact you about expiring
    # certificates, and issues related to your account.
    email: <your email address>
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      # Secret resource used to store the account's private key.
      name: example-issuer-account-key
    # Add a single challenge solver, HTTP01 using nginx
    solvers:
      - http01:
          ingress:
            ingressTemplate:
              metadata:
                annotations:
                  "external-dns.alpha.kubernetes.io/ttl": "5"
```

2. Setup External dns and correctly create your issuer into your cluster configuration:

```yaml 
apiVersion: nifi.konpyutaika.com/v1
kind: NifiCluster
...
spec:
  ...
  # TLS is configured through listenersConfig and sslSecrets.
  ...
  listenersConfig:
    clusterDomain: <DNS zone name>
    useExternalDNS: true
    ...
    sslSecrets:
      tlsSecretName: "test-nifikop"
      create: true
      issuerRef:
        name: letsencrypt-staging
        kind: Issuer
```

## Create SSL credentials

You may use `NifiUser` resource to create new user certificates for your applications, allowing them to authenticate and query your Nifi cluster.

To create a new client you will need to generate new certificates sign by the CA. The operator can automate this for you using the `NifiUser` CRD:

```console
cat << EOF | kubectl apply -n nifi -f -
apiVersion:  nifi.konpyutaika.com/v1
kind: NifiUser
metadata:
  name: example-client
  namespace: nifi
spec:
  clusterRef:
    name: nifi
  secretName: example-client-secret
EOF
```

This will create a user and store its credentials in the secret `example-client-secret`. The secret contains these fields:

| key | value |
|-----|-------|
| ca.crt | The CA certificate |
| tls.crt | The user certificate |
| tls.key | The user private key |

You can then mount these secret to your pod. Alternatively, you can write them to your local machine by running:

```console
kubectl get secret example-client-secret -o jsonpath="{['data']['ca\.crt']}" | base64 -d > ca.crt
kubectl get secret example-client-secret -o jsonpath="{['data']['tls\.crt']}" | base64 -d > tls.crt
kubectl get secret example-client-secret -o jsonpath="{['data']['tls\.key']}" | base64 -d > tls.key
```

The operator can also include a Java keystore format (JKS) with your user secret if you'd like. Add `includeJKS`: `true` to the `spec` like shown above, and then the user-secret will gain these additional fields:

| key | value |
|-----|-------|
| tls.jks | The java keystore containing both the user keys and the CA (use this for your keystore AND truststore) |
| pass.txt | The password to decrypt the JKS file (this will be randomly generated) |
