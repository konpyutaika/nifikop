---
id: 2_read_only_config
title: Read only configurations
sidebar_label: Read only configurations
---

ReadOnlyConfig object specifies the read-only type Nifi config cluster wide, all theses will be merged with node specified readOnly configurations, so it can be overwritten per node.

```yaml
  # readOnlyConfig specifies the read-only type Nifi config cluster wide, all theses
  # will be merged with node specified readOnly configurations, so it can be overwritten per node.
  readOnlyConfig:
    # NifiProperties configuration that will be applied to the node.
    nifiProperties:
      # Additionnals nifi.properties configuration that will override the one produced based
      # on template and configurations.
      overrideConfigs: |
        nifi.ui.banner.text=NiFiKop by Orange
      # A comma separated list of allowed HTTP Host header values to consider when NiFi
      # is running securely and will be receiving requests to a different host[:port] than it is bound to.
      # https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#web-properties
#      webProxyHost:
      # Nifi security client auth
      needClientAuth: false
      # Indicates which of the configured authorizers in the authorizers.xml file to use
      # https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#authorizer-configuration
#      authorizer:
    # ZookeeperProperties configuration that will be applied to the node.
    zookeeperProperties:
      # Additionnals zookeeper.properties configuration that will override the one produced based
      # on template and configurations.
      overrideConfigs: |
        initLimit=15
        autopurge.purgeInterval=24
        syncLimit=5
        tickTime=2000
        dataDir=./state/zookeeper
        autopurge.snapRetainCount=30
    # BootstrapProperties configuration that will be applied to the node.
    bootstrapProperties:
      # JVM memory settings
      nifiJvmMemory: "512m"
      # Additionnals bootstrap.properties configuration that will override the one produced based
      # on template and configurations.
      # https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#bootstrap_properties
      overrideConfigs: |
        # java.arg.4=-Djava.net.preferIPv4Stack=true

```

## ReadOnlyConfig

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|nifiProperties|[NifiProperties](#nifiproperties)|nifi.properties configuration that will be applied to the node.|No|nil|
|zookeeperProperties|[ZookeeperProperties](#zookeeperproperties)|zookeeper.properties configuration that will be applied to the node.|No|nil|
|bootstrapProperties|[BootstrapProperties](#bootstrapproperties)|bootstrap.conf configuration that will be applied to the node.|No|nil|

## NifiProperties

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|overrideConfigs|string|Additionnals nifi.properties configuration that will override the one produced based on template and configurations.|No|""|
|webProxyHosts|\[ \]string| A list of allowed HTTP Host header values to consider when NiFi is running securely and will be receiving requests to a different host[:port] than it is bound to. [web-properties](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#web-properties)|No|""|
|needClientAuth|boolean|Nifi security client auth.|No|false|
|authorizer|string|Indicates which of the configured authorizers in the authorizers.xml file to use [authorizer-configuration](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#authorizer-configuration)|No|"managed-authorizer"|


## ZookeeperProperties

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|overrideConfigs|string|Additionnals zookeeper.properties configuration that will override the one produced based on template and configurations.|No|""|

## BootstrapProperties

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|overrideConfigs|string|Additionnals bootstrap.conf configuration that will override the one produced based on template and configurations.|No|""|
|NifiJvmMemory|string|JVM memory settings.|No|"512m"|