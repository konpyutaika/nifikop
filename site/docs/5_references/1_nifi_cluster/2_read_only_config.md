---
id: 2_read_only_config
title: Read only configurations
sidebar_label: Read only configurations
---

ReadOnlyConfig object specifies the read-only type Nifi config cluster wide, all theses will be merged with node specified readOnly configurations, so it can be overwritten per node.

```yaml
readOnlyConfig:
  # MaximumTimerDrivenThreadCount define the maximum number of threads for timer driven processors available to the system.
  maximumTimerDrivenThreadCount: 30
  # MaximumEventDrivenThreadCount define the maximum number of threads for event driven processors available to the system.
  maximumEventDrivenThreadCount: 10
  # Logback configuration that will be applied to the node
  logbackConfig:
    # logback.xml configuration that will replace the one produced based on template
    replaceConfigMap:
      # The key of the value,in data content, that we want use.
      data: logback.xml
      # Name of the configmap that we want to refer.
      name: raw
      # Namespace where is located the secret that we want to refer.
      namespace: nifikop
    # logback.xml configuration that will replace the one produced based on template and overrideConfigMap
    replaceSecretConfig:
      # The key of the value,in data content, that we want use.
      data: logback.xml
      # Name of the configmap that we want to refer.
      name: raw
      # Namespace where is located the secret that we want to refer.
      namespace: nifikop
  # Authorizer configuration that will be applied to the node
  authorizerConfig:
    # An authorizers.xml configuration template that will replace the default template seen in authorizers.go
    replaceTemplateConfigMap:
      # The key of the value, in data content, that we want use.
      data: authorizers.xml
      # Name of the configmap that we want to refer.
      name: raw
      # Namespace where is located the secret that we want to refer.
      namespace: nifikop
    # An authorizers.xml configuration template that will replace the default template seen in authorizers.go and the replaceTemplateConfigMap
    replaceTemplateSecretConfig:
      # The key of the value,in data content, that we want use.
      data: authorizers.xml
      # Name of the configmap that we want to refer.
      name: raw
      # Namespace where is located the secret that we want to refer.
      namespace: nifikop
  # NifiProperties configuration that will be applied to the node.
  nifiProperties:
    # Additionnals nifi.properties configuration that will override the one produced based on template and
    # configuration
    overrideConfigMap:
      # The key of the value,in data content, that we want use.
      data: nifi.properties
      # Name of the configmap that we want to refer.
      name: raw
      # Namespace where is located the secret that we want to refer.
      namespace: nifikop.
    # Additionnals nifi.properties configuration that will override the one produced based
    #	on template, configurations, overrideConfigMap and overrideConfigs.
    overrideSecretConfig:
      # The key of the value,in data content, that we want use.
      data: nifi.properties
      # Name of the configmap that we want to refer.
      name: raw
      # Namespace where is located the secret that we want to refer.
      namespace: nifikop
    # Additionnals nifi.properties configuration that will override the one produced based
    #	on template, configurations and overrideConfigMap
    overrideConfigs: |
      nifi.ui.banner.text=NiFiKop
    # A comma separated list of allowed HTTP Host header values to consider when NiFi
    # is running securely and will be receiving requests to a different host[:port] than it is bound to.
    # https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#web-properties
    #      webProxyHosts:
    # Nifi security client auth
    needClientAuth: false
    # Indicates which of the configured authorizers in the authorizers.xml file to use
    # https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#authorizer-configuration
  #      authorizer:
  # ZookeeperProperties configuration that will be applied to the node.
  zookeeperProperties:
    #      # Additionnals zookeeeper.properties configuration that will override the one produced based on template and
    #      # configuration
    #      overrideConfigMap:
    #        # The key of the value,in data content, that we want use.
    #        data: zookeeeper.properties
    #        # Name of the configmap that we want to refer.
    #        name: raw
    #        # Namespace where is located the secret that we want to refer.
    #        namespace: nifikop.
    #      # Additionnals zookeeeper.properties configuration that will override the one produced based
    #      #	on template, configurations, overrideConfigMap and overrideConfigs.
    #      overrideSecretConfig:
    #        # The key of the value,in data content, that we want use.
    #        data: zookeeeper.properties
    #        # Name of the configmap that we want to refer.
    #        name: raw
    #        # Namespace where is located the secret that we want to refer.
    #        namespace: nifikop
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
    #      # Additionnals bootstrap.properties configuration that will override the one produced based on template and
    #      # configuration
    #      overrideConfigMap:
    #        # The key of the value,in data content, that we want use.
    #        data: bootstrap.properties
    #        # Name of the configmap that we want to refer.
    #        name: raw
    #        # Namespace where is located the secret that we want to refer.
    #        namespace: nifikop.
    #      # Additionnals bootstrap.properties configuration that will override the one produced based
    #      #	on template, configurations, overrideConfigMap and overrideConfigs.
    #      overrideSecretConfig:
    #        # The key of the value,in data content, that we want use.
    #        data: bootstrap.properties
    #        # Name of the configmap that we want to refer.
    #        name: raw
    #        # Namespace where is located the secret that we want to refer.
    #        namespace: nifikop
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
|maximumTimerDrivenThreadCount|int32|define the maximum number of threads for timer driven processors available to the system.|No|10|
|maximumEventDrivenThreadCount|int32|define the maximum number of threads for event driven processors available to the system.|No|1|
|additionalSharedEnvs|\[&nbsp;\][corev1.EnvVar](https://pkg.go.dev/k8s.io/api/core/v1#EnvVar)|define a set of additional env variables that will shared between all init containers and ontainers in the pod..|No|\[&nbsp;\]|
|nifiProperties|[NifiProperties](#nifiproperties)|nifi.properties configuration that will be applied to the node.|No|nil|
|zookeeperProperties|[ZookeeperProperties](#zookeeperproperties)|zookeeper.properties configuration that will be applied to the node.|No|nil|
|bootstrapProperties|[BootstrapProperties](#bootstrapproperties)|bootstrap.conf configuration that will be applied to the node.|No|nil|
|logbackConfig|[LogbackConfig](#logbackconfig)|logback.xml configuration that will be applied to the node.|No|nil|
|authorizerConfig|[AuthorizerConfig](#authorizerconfig)|authorizers.xml configuration template that will be applied to the node.|No|nil|
|bootstrapNotificationServicesConfig|[BootstrapNotificationServices](#bootstrapnotificationservices)|bootstrap_notification_services.xml configuration that will be applied to the node.|No|nil|



## NifiProperties

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|overrideConfigMap|[ConfigmapReference](#configmapreference)|Additionnals nifi.properties configuration that will override the one produced based on template and configuration.|No|nil|
|overrideConfigs|string|Additionnals nifi.properties configuration that will override the one produced based on template, configurations and overrideConfigMap.|No|""|
|overrideSecretConfig|[SecretConfigReference](#secretconfigreference)|Additionnals nifi.properties configuration that will override the one produced based on template, configurations, overrideConfigMap and overrideConfigs.|No|nil|
|webProxyHosts|\[&nbsp;\]string| A list of allowed HTTP Host header values to consider when NiFi is running securely and will be receiving requests to a different host[:port] than it is bound to. [web-properties](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#web-properties)|No|""|
|needClientAuth|boolean|Nifi security client auth.|No|false|
|authorizer|string|Indicates which of the configured authorizers in the authorizers.xml file to use [authorizer-configuration](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#authorizer-configuration)|No|"managed-authorizer"|


## ZookeeperProperties

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|overrideConfigMap|[ConfigmapReference](#configmapreference)|Additionnals zookeeper.properties configuration that will override the one produced based on template and configuration.|No|nil|
|overrideConfigs|string|Additionnals zookeeper.properties configuration that will override the one produced based on template, configurations and overrideConfigMap.|No|""|
|overrideSecretConfig|[SecretConfigReference](#secretconfigreference)|Additionnals zookeeper.properties configuration that will override the one produced based on template, configurations, overrideConfigMap and overrideConfigs.|No|nil|

## BootstrapProperties

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|overrideConfigMap|[ConfigmapReference](#configmapreference)|Additionnals bootstrap.properties configuration that will override the one produced based on template and configuration.|No|nil|
|overrideConfigs|string|Additionnals bootstrap.properties configuration that will override the one produced based on template, configurations and overrideConfigMap.|No|""|
|overrideSecretConfig|[SecretConfigReference](#secretconfigreference)|Additionnals bootstrap.properties configuration that will override the one produced based on template, configurations, overrideConfigMap and overrideConfigs.|No|nil|
|NifiJvmMemory|string|JVM memory settings.|No|"512m"|

## LogbackConfig

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|replaceConfigMap|[ConfigmapReference](#configmapreference)|logback.xml configuration that will replace the one produced based on template.|No|nil|
|replaceSecretConfig|[SecretConfigReference](#secretconfigreference)|logback.xml configuration that will replace the one produced based on template and overrideConfigMap.|No|nil|

## AuthorizerConfig

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|replaceTemplateConfigMap|[ConfigmapReference](#configmapreference)|authorizers.xml configuration template that will replace the default template.|No|nil|
|replaceTemplateSecretConfig|[SecretConfigReference](#secretconfigreference)|authorizers.xml configuration that will replace the default template and the replaceTemplateConfigMap.|No|nil|

## BootstrapNotificationServicesConfig

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|replaceConfigMap|[ConfigmapReference](#configmapreference)|bootstrap_notifications_services.xml configuration that will replace the one produced based on template.|No|nil|
|replaceSecretConfig|[SecretConfigReference](#secretconfigreference)|bootstrap_notifications_services.xml configuration that will replace the one produced based on template and overrideConfigMap.|No|nil|

## ConfigmapReference

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string|Name of the configmap that we want to refer.|Yes|""|
|namespace|string|Namespace where is located the configmap that we want to refer.|No|""|
|data|string|The key of the value,in data content, that we want use.|Yes|""|

## SecretConfigReference

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string|Name of the secret that we want to refer.|Yes|""|
|namespace|string|Namespace where is located the secret that we want to refer.|No|""|
|data|string|The key of the value,in data content, that we want use.|Yes|""|