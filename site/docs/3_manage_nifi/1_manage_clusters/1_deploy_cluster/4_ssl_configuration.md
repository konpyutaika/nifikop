---
id: 4_ssl_configuration
title: SSL configuration
sidebar_label: SSL configuration
---

The `NiFi operator` makes securing your NiFi cluster with SSL. You may provide your own certificates, or instruct the operator to create them for from your cluster configuration.

Below this is an example configuration required to secure your cluster with SSL :

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

- `readOnlyConfig.nifiProperties.webProxyHosts` : A list of allowed HTTP Host header values to consider when NiFi is running securely and will be receiving requests to a different host[:port] than it is bound to. [web-properties](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#web-properties)

If `listenersConfig.sslSecrets.create` is set to `false`, the operator will look for the secret at `listenersConfig.sslSecrets.tlsSecretName` and expect these values :

| key | value |
|-----|-------|
| caCert | The CA certificate |
| caKey | The CA private key |
| clientCert | A client certificate (this will be used by operator for NiFI operations) |
| clientKey | The private key for clientCert |

## Using an existing Issuer

As described in the [Reference section](../../../5_references/1_nifi_cluster/6_listeners_config.md#sslsecrets), instead of using a self-signed certificate as CA, you can use an existing one.
In order to do so, you only have to refer it into your `Spec.ListenerConfig.SslSecrets.IssuerRef` field.

### Example : Let's encrypt

Let's say you have an existing DNS server, with [external dns](https://github.com/kubernetes-sigs/external-dns) deployed into your cluster's namespace.
You can easily use Let's encrypt as authority for your certificate.

To do this, you have to :

1. Create an issuer :

```yaml
apiVersion: cert-manager.io/v1alpha2
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

2. Setup External dns and correctly create your issuer into your cluster configuration :

```yaml 
apiVersion: nifi.konpyutaika.com/v1
kind: NifiCluster
...
spec:
  ...
  clusterSecure: true
  siteToSiteSecure: true
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

You may use `NifiUser` resource to create new certificates for your applications, allowing them to query your Nifi cluster.

To create a new client you will need to generate new certificates sign by the CA. The operator can automate this for you using the `NifiUser` CRD :

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

This will create a user and store its credentials in the secret `example-client-secret`. The secret contains these fields :

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

The operator can also include a Java keystore format (JKS) with your user secret if you'd like. Add `includeJKS`: `true` to the `spec` like shown above, and then the user-secret will gain these additional fields :

| key | value |
|-----|-------|
| tls.jks | The java keystore containing both the user keys and the CA (use this for your keystore AND truststore) |
| pass.txt | The password to decrypt the JKS (this will be randomly generated) |