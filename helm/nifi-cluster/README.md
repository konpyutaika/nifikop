# nifi-cluster

![Version: 1.0.0](https://img.shields.io/badge/Version-1.0.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.19.0](https://img.shields.io/badge/AppVersion-1.19.0-informational?style=flat-square)

A Helm chart for deploying NiFi clusters in Kubernetes

**Homepage:** <https://github.com/konpyutaika/nifikop>

## Source Code

* <https://github.com/konpyutaika/nifikop>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | zookeeper | 10.2.5 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| cluster.additionalSharedEnvs | list | `[]` | list of additional environment variables to attach to all init containers and the nifi container https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/2_read_only_config#readonlyconfig |
| cluster.bootstrapProperties | object | `{"nifiJvmMemory":"512m","overrideConfigs":"java.arg.4=-Djava.net.preferIPv4Stack=true\njava.arg.log4shell=-Dlog4j2.formatMsgNoLookups=true\n"}` | You can override individual properties in config/bootstrap.properties https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#bootstrap_properties |
| cluster.disruptionBudget | object | `{}` | see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/1_nifi_cluster#disruptionbudget |
| cluster.externalServices[0].metadata.annotations | object | `{}` |  |
| cluster.externalServices[0].metadata.labels | object | `{}` |  |
| cluster.externalServices[0].name | string | `"nifi-cluster-ip"` |  |
| cluster.externalServices[0].spec.portConfigs[0].internalListenerName | string | `"http"` |  |
| cluster.externalServices[0].spec.portConfigs[0].port | int | `8080` |  |
| cluster.externalServices[0].spec.type | string | `"ClusterIP"` |  |
| cluster.fullnameOverride | string | `""` |  |
| cluster.image.repository | string | `"apache/nifi"` |  |
| cluster.image.tag | string | `""` | Only set this if you want to override the chart AppVersion |
| cluster.initContainerImage.repository | string | `"busybox"` |  |
| cluster.initContainerImage.tag | string | `"latest"` |  |
| cluster.initContainers | list | `[]` | list of init containers to run prior to the deployment |
| cluster.ldapConfiguration | object | `{}` | see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/1_nifi_cluster#ldapconfiguration |
| cluster.listenersConfig | object | `{"internalListeners":[{"containerPort":8080,"name":"http","type":"http"},{"containerPort":6007,"name":"cluster","type":"cluster"},{"containerPort":10000,"name":"s2s","type":"s2s"},{"containerPort":9090,"name":"prometheus","type":"prometheus"}],"sslSecrets":null}` | https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/6_listeners_config |
| cluster.logbackConfig.configPath | string | `"config/logback.xml"` |  |
| cluster.logbackConfig.replaceConfigMap | object | `{}` | A ConfigMap ref to override the default logback configuration see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/2_read_only_config#logbackconfig |
| cluster.logbackConfig.replaceSecretConfig | object | `{}` | A Secret ref to override the default logback configuration see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/2_read_only_config#logbackconfig |
| cluster.managedAdminUsers | list | `[]` | see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/1_nifi_cluster#managedusers |
| cluster.managedReaderUsers | list | `[]` | see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/1_nifi_cluster#managedusers |
| cluster.maximumEventDrivenThreadCount | int | `10` | MaximumEventDrivenThreadCount defines the maximum number of threads for timer driven processors available to the system. This is a feature enabled by the following PR and should not be used unless you're running nifkop with this PR applied: https://github.com/Orange-OpenSource/nifikop/pull/184 |
| cluster.maximumTimerDrivenThreadCount | int | `10` | MaximumTimerDrivenThreadCount defines the maximum number of threads for timer driven processors available to the system. |
| cluster.nameOverride | string | `"nifi-cluster"` | the full name of the cluster. This is used to set a portion of the name of various nifikop resources |
| cluster.nifiProperties | object | `{"needClientAuth":false,"overrideConfigs":"nifi.web.proxy.context.path=/nifi-cluster\n","webProxyHosts":""}` | You can override the individual properties via the overrideConfigs attribute. These will be provided to all pods via secrets. https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#system_properties |
| cluster.nifiProperties.needClientAuth | bool | `false` | Nifi security client auth |
| cluster.nifiProperties.webProxyHosts | string | `""` | A comma separated list of allowed HTTP Host header values to consider when NiFi is running securely and will be receiving requests to a different host[:port] than it is bound to. https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#web-properties |
| cluster.nodeConfigGroups | object | `{}` | see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/3_node_config |
| cluster.nodes | list | `[{"id":1,"nodeConfigGroup":"default-group"}]` | see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/1_nifi_cluster#nificlusterspec |
| cluster.oneNifiNodePerNode | bool | `false` | whether or not to only deploy one nifi pod per node in this cluster |
| cluster.pod.annotations | object | `{}` | Annotations to apply to every pod |
| cluster.pod.hostAlises | list | `[]` | host aliases to assign to each pod |
| cluster.pod.labels | object | `{}` | Labels to apply to every pod |
| cluster.propagateLabels | bool | `true` |  |
| cluster.retryDurationMinutes | int | `10` | The number of minutes the operator should wait for the cluster to be successfully deployed before retrying |
| cluster.service | object | `{"annotations":{},"headlessEnabled":true,"labels":{}}` | the template to use to create nodes. see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/1_nifi_cluster#nificlusterspec nodeUserIdentityTemplate: n-%d |
| cluster.service.annotations | object | `{}` | Annotations to apply to each nifi service |
| cluster.service.headlessEnabled | bool | `true` | Whether or not to create a headless service |
| cluster.service.labels | object | `{}` | Labels to apply to each nifi service |
| cluster.zkAddress | string | `"nifi-cluster-zookeeper:2181"` | the hostname and port of the zookeeper service |
| cluster.zkPath | string | `"/cluster"` | the path in zookeeper to store this cluster's state |
| cluster.zookeeperProperties | object | `{"overrideConfigs":"initLimit=15\nautopurge.purgeInterval=24\nsyncLimit=5\ntickTime=2000\ndataDir=./state/zookeeper\nautopurge.snapRetainCount=30\n"}` | This is only for embedded zookeeper configuration. This is ignored if an zookeeper.enabled is true. |
| dataflows | list | `[{"bucketId":"","enabled":false,"flowId":"","flowVersion":1,"name":"My Special Dataflow","parameterContextRef":{"name":"default","namespace":"nifi"},"registryClientRef":{"name":"default","namespace":"nifi"},"skipInvalidComponent":true,"skipInvalidControllerService":true,"syncMode":"always","updateStrategy":"drain"}]` | Versioned dataflow configurations. This is used to configure versioned dataflows to be deployed to this nifi cluster. Any number may be configured. Note that a _registryClient_ and a _parameterContext_ must be enabled & present in order for a dataflow to be deployed to a cluster. See https://konpyutaika.github.io/nifikop/docs/5_references/5_nifi_dataflow |
| extraManifests | list | `[]` | A list of extra templated Kubernetes yamls to apply |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `"nginx"` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts | list | `[]` |  |
| ingress.tls | list | `[]` |  |
| logging.enabled | bool | `false` | Whether or not log aggregation via the banzai cloud logging operator is enabled. |
| logging.flow | object | `{"filters":[{"parser":{"parse":{"expression":"/^(?<time>\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2},\\d{3}) (?<level>[^\\s]+) \\[(?<thread>.*)\\] (?<message>.*)$/im","keep_time_key":true,"time_format":"%iso8601","time_key":"time","time_type":"string","type":"regexp"}}}],"match":[{"select":{"labels":{"app":"nifi"}}}],"name":"nifi-cluster-flow"}` | https://banzaicloud.com/docs/one-eye/logging-operator/configuration/flow/ |
| logging.flow.filters | list | `[{"parser":{"parse":{"expression":"/^(?<time>\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2},\\d{3}) (?<level>[^\\s]+) \\[(?<thread>.*)\\] (?<message>.*)$/im","keep_time_key":true,"time_format":"%iso8601","time_key":"time","time_type":"string","type":"regexp"}}}]` | The filters and match configs should be configured just like in the CRDs (linked above) |
| logging.outputs | object | `{"globalOutputRefs":["loki-cluster-output"]}` | Only global outputs that have been created separately to this helm chart supported for now may consider changing this to a Flow per cluster in future |
| monitoring | object | `{"enabled":false}` | Monitoring is enabled by the Prometheus operator. This can be deployed stand-alone or as a part of the Rancher Monitoring application. Do not enable this unless you've installed rancher-logging or the Promtheus operator directly. https://rancher.com/docs/rancher/v2.6/en/monitoring-alerting/ Enabling logging creates a `ServiceResource` custom resource and routes logs to the output of your choice |
| nodeGroupAutoscalers | list | `[{"downscaleStrategy":"lifo","enabled":false,"horizontalAutoscaler":{"maxReplicas":2,"minReplicas":1},"name":"default-group-autoscaler","nodeConfig":{},"nodeConfigGroupId":"default-group","nodeLabelsSelector":{"matchLabels":{"default-scale-group":"true"}},"readOnlyConfig":{},"upscaleStrategy":"simple"}]` | Nifi NodeGroup Autoscaler configurations. Use this to autoscale any NodeGroup specified in `cluster.nodeConfigGroups`. To autoscale  See https://konpyutaika.github.io/nifikop/docs/5_references/7_nifi_nodegroup_autoscaler |
| parameterContexts | list | `[{"enabled":false,"name":"default","parameters":[{"description":"my foo bar property","name":"foo-prop","value":"bar-value"}],"secretRefs":[]}]` | Parameter context configurations. This is required if you wish to deploy versioned flows via the dataflow config. However,  it is not required to provide secret refs. You must provide at least one parameter or nifikop will choke on updating dataflows. The .name field must be safe in Kubernetes and match the pattern [A-Za-z0-9-] See https://konpyutaika.github.io/nifikop/docs/5_references/4_nifi_parameter_context |
| registryClients | list | `[{"description":"Default NiFi Registry client","enabled":false,"endpoint":"http://nifi-registry","name":"default"},{"description":"Alternate NiFi Registry client","enabled":false,"endpoint":"http://nifi-registry","name":"alternate"}]` | registry client configurations. You'd use this to version control process groups & store the configuration in a registry bucket This is required if you wish to deploy versioned flows via the dataflow config The .name field must be safe in Kubernetes and match the pattern [A-Za-z0-9-] See https://konpyutaika.github.io/nifikop/docs/5_references/3_nifi_registry_client |
| userGroups | list | `[]` | Configure user groups. Each will result in the creation of a `NiFiUserGroup` CRD in k8s, which the operator takes and applies to each nifi configuration See https://konpyutaika.github.io/nifikop/docs/5_references/6_nifi_usergroup |
| users | list | `[]` | Configure users. Each will result in the creation of a `NiFiUser` CRD in k8s, which the operator takes and applies to each nifi configuration See https://konpyutaika.github.io/nifikop/docs/5_references/2_nifi_user the user's name is used for k8s resource metadata.name and so should be alphanumeric and hypenated |
| zookeeper | object | `{"enabled":false,"persistence":{"size":"10Gi","storageClass":"ceph-filesystem"},"replicaCount":1,"resources":{"limits":{"cpu":2,"memory":"500Mi"},"requests":{"cpu":"0.5m","memory":"250Mi"}}}` | zookeeper chart overrides |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
