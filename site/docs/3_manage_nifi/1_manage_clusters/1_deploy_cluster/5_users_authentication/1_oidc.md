---
id: 1_oidc
title: OpenId Connect
sidebar_label: OpenId Connect
---

To enable authentication via OpenId Connect refering to [NiFi Administration guide](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html) required some configuration into `nifi.properties`.

In addition and to ensure multiple identity provider support, we recommended to add the following configuration to your `nifi.properties` : 

```sh
nifi.security.identity.mapping.pattern.dn=CN=([^,]*)(?:, (?:O|OU)=.*)?
nifi.security.identity.mapping.value.dn=$1
nifi.security.identity.mapping.transform.dn=NONE
```

To perform this with `NiFiKop` you just have to configure the `Spec.NifiProperties.OverrideConfigs` field with your OIDC configuration, for example :

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
      # Additionnals nifi.properties configuration that will override the one produced based
      # on template and configurations.
      overrideConfigs: |
        nifi.security.user.oidc.discovery.url=<oidc server discovery url>
        nifi.security.user.oidc.client.id=<oidc client's id>
        nifi.security.user.oidc.client.secret=<oidc client's secret>
        nifi.security.identity.mapping.pattern.dn=CN=([^,]*)(?:, (?:O|OU)=.*)?
        nifi.security.identity.mapping.value.dn=$1
        nifi.security.identity.mapping.transform.dn=NONE
      ...
   ...
...
```