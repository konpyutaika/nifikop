---
id: 6_listeners_config
title: Listeners Config
sidebar_label: Listeners Config
---

ListenersConfig defines the Nifi listener types :

```yaml
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
      - type: "prometheus"
        name: "prometheus"
        containerPort: 9090
    sslSecrets:
      tlsSecretName: "test-nifikop"
      create: true
```

## ListenersConfig

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|internalListeners|\[&nbsp;\][InternalListener](#internallistener)| specifies settings required to access nifi internally.| Yes | - |
|sslSecrets|[SSLSecrets](#sslsecrets)| contains information about ssl related kubernetes secrets if one of the listener setting type set to ssl these fields must be populated to.| Yes | nil |
|clusterDomain|string|  allow to override the default cluster domain which is "cluster.local".| Yes | `cluster.local` |
|useExternalDNS|string|  allow to manage externalDNS usage by limiting the DNS names associated to each nodes and load balancer: `<cluster-name>-node-<node Id>.<cluster-name>.<service name>.<cluster domain>`| Yes | false |

## InternalListener

Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|type|enum{ "cluster", "http", "https", "s2s", "prometheus"}| allow to specify if we are in a specific nifi listener it's allowing to define some required information such as Cluster Port, Http Port, Https Port, S2S or Prometheus port| Yes | - |
|name|string| an identifier for the port which will be configured. | Yes | - |
|containerPort|int32| the containerPort. | Yes | - |


## SSLSecrets

Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|tlsSecretName|string| should contain all ssl certs required by nifi including: caCert, caKey, clientCert, clientKey serverCert, serverKey, peerCert, peerKey. | Yes | - |
|create|boolean| tells the installed cert manager to create the required certs keys. | Yes | - |
|clusterScoped|boolean| defines if the Issuer created is cluster or namespace scoped. | Yes | - |
|issuerRef|[ObjectReference](https://docs.cert-manager.io/en/release-0.9/reference/api-docs/index.html#objectreference-v1alpha1)| IssuerRef allow to use an existing issuer to act as CA: https://cert-manager.io/docs/concepts/issuer/ | No | - |
|pkiBackend|enum{"cert-manager"}| | Yes | - |

