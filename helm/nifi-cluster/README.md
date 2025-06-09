# nifi-cluster

![Version: 1.14.1](https://img.shields.io/badge/Version-1.14.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.28.0](https://img.shields.io/badge/AppVersion-1.28.0-informational?style=flat-square)

A Helm chart for deploying NiFi clusters in Kubernetes

**Homepage:** <https://github.com/konpyutaika/nifikop>

## Source Code

* <https://github.com/konpyutaika/nifikop>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | zookeeper | 12.4.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| cluster.additionalSharedEnvs | list | `[]` | list of additional environment variables to attach to all init containers and the nifi container https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/2_read_only_config#readonlyconfig |
| cluster.bootstrapProperties | object | `{"nifiJvmMemory":"512m","overrideConfigs":"java.arg.4=-Djava.net.preferIPv4Stack=true\njava.arg.log4shell=-Dlog4j2.formatMsgNoLookups=true\n"}` | You can override individual properties in conf/bootstrap.conf https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#bootstrap_properties |
| cluster.clientType | string | tls | defines if the operator will use basic or tls authentication to query the NiFi cluster. Operator will default to tls if left unset |
| cluster.controllerUserIdentity | string | `nil` | ControllerUserIdentity specifies what to call the static admin user's identity. **Warning: once defined don't change this value either the operator will no longer be able to manage the cluster** |
| cluster.disruptionBudget | object | `{}` | see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster#disruptionbudget |
| cluster.externalServices | list | `[{"metadata":{"annotations":{},"labels":{}},"name":"nifi-cluster-ip","spec":{"portConfigs":[{"internalListenerName":"http","port":8080}],"type":"ClusterIP"}}]` | Additional k8s services to create and target internal listener ports. Ingress will use these to route traffic to the cluster |
| cluster.fullnameOverride | string | `""` |  |
| cluster.image.repository | string | `"apache/nifi"` |  |
| cluster.image.tag | string | `""` | Only set this if you want to override the chart AppVersion |
| cluster.initContainerImage.repository | string | `"bash"` |  |
| cluster.initContainerImage.tag | string | `"5.2.2"` |  |
| cluster.initContainers | list | `[]` | list of init containers to run prior to the deployment |
| cluster.ldapConfiguration | object | `{}` | see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster#ldapconfiguration |
| cluster.listenersConfig | object | `{"internalListeners":[{"containerPort":8080,"name":"http","type":"http"},{"containerPort":6007,"name":"cluster","type":"cluster"},{"containerPort":10000,"name":"s2s","type":"s2s"},{"containerPort":9090,"name":"prometheus","type":"prometheus"}],"sslSecrets":null}` | https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/6_listeners_config |
| cluster.listenersConfig.internalListeners | list | `[{"containerPort":8080,"name":"http","type":"http"},{"containerPort":6007,"name":"cluster","type":"cluster"},{"containerPort":10000,"name":"s2s","type":"s2s"},{"containerPort":9090,"name":"prometheus","type":"prometheus"}]` | List of internal ports exposed for the nifi container. The `type` of port has specific meaning, see:    https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/6_listeners_config#internallistener |
| cluster.listenersConfig.sslSecrets | string | `nil` | Provides the SSL configuration for the cluster, can be fully user provided, user provided CA or auto generated    with cert-manager. See: https://konpyutaika.github.io/nifikop/docs/3_manage_nifi/1_manage_clusters/1_deploy_cluster/4_ssl_configuration |
| cluster.logbackConfig.configPath | string | `"config/logback.xml"` |  |
| cluster.logbackConfig.replaceConfigMap | object | `{}` | A ConfigMap ref to override the default logback configuration see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/2_read_only_config#logbackconfig |
| cluster.logbackConfig.replaceSecretConfig | object | `{}` | A Secret ref to override the default logback configuration see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/2_read_only_config#logbackconfig |
| cluster.managedAdminUsers | list | `[]` | see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster#managedusers |
| cluster.managedReaderUsers | list | `[]` | see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster#managedusers |
| cluster.manager | string | zookeeper | the type of cluster manager: zookeeper or kubernetes. Operator will put zookeeper by default |
| cluster.managerServiceAccount | object | `{"annotations":{},"labels":{},"name":null}` | the kubernetes manager serviceAccount details |
| cluster.managerServiceAccount.annotations | object | `{}` | Annotations to apply to the serviceAccount |
| cluster.managerServiceAccount.labels | object | `{}` | Labels to apply to the serviceAccount |
| cluster.maximumTimerDrivenThreadCount | int | `10` | MaximumTimerDrivenThreadCount defines the maximum number of threads for timer driven processors available to the system. |
| cluster.nameOverride | string | `"nifi-cluster"` | the full name of the cluster. This is used to set a portion of the name of various nifikop resources |
| cluster.nifiControllerTemplate | string | `nil` |  |
| cluster.nifiProperties | object | `{"needClientAuth":false,"overrideConfigMap":{},"overrideConfigs":"nifi.web.proxy.context.path=/nifi-cluster\n","overrideSecretConfig":{},"webProxyHosts":[],"webProxyNodePorts":{"enabled":false,"hosts":[]}}` | You can override the individual properties via the overrideConfigs attribute. These will be provided to all pods via secrets. https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#system_properties |
| cluster.nifiProperties.needClientAuth | bool | `false` | Nifi security client auth |
| cluster.nifiProperties.overrideConfigMap | object | `{}` | A ConfigMap ref to override the default nifi properties see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/2_read_only_config#nifiproperties |
| cluster.nifiProperties.overrideSecretConfig | object | `{}` | A Secret ref to override the default nifi properties see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/2_read_only_config#nifiproperties |
| cluster.nifiProperties.webProxyHosts | list | `[]` | List of allowed HTTP Host header values to consider when NiFi is running securely and will be receiving requests to a different host[:port] than it is bound to. Operator will generate comma separated string from list https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#web-properties |
| cluster.nifiProperties.webProxyNodePorts | object | `{"enabled":false,"hosts":[]}` | In case `cluster.externalServices` contains a service of type `NodePort` and NiFi UI/API needs to be accessed over it, this option will add host:nodePort to the `webProxyHosts` list inside `NiFiCluster`. Note: When adding webProxyHosts as host:port, NiFi will also create entry for host as valid host header. |
| cluster.nodeConfigGroups | object | `{}` | Defines configurations for nodes which can be used in list of nodes in cluster.    See: https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster/3_node_config |
| cluster.nodeUserIdentityTemplate | string | "node-%d-<cluster-name>" | the template to use to create nodes. see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster#nificlusterspec |
| cluster.nodes | list | `[{"id":1,"nodeConfigGroup":"default-group"}]` | Defines the list of nodes in the cluster with their id's and config to apply.    See https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster#nificlusterspec |
| cluster.oneNifiNodePerNode | bool | `false` | whether or not to only deploy one nifi pod per node in this cluster |
| cluster.pod.annotations | object | `{}` | Annotations to apply to every pod |
| cluster.pod.hostAliases | list | `[]` | host aliases to assign to each pod. See: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#hostalias-v1-core |
| cluster.pod.labels | object | `{}` | Labels to apply to every pod |
| cluster.pod.livenessProbe | object | `nil` | The pod liveness probe override: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-command |
| cluster.pod.readinessProbe | object | `nil` | The pod readiness probe override: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-readiness-probes |
| cluster.propagateLabels | bool | `true` |  |
| cluster.retryDurationMinutes | int | `10` | The number of minutes the operator should wait for the cluster to be successfully deployed before retrying |
| cluster.service.annotations | object | `{}` | Annotations to apply to each nifi service |
| cluster.service.headlessEnabled | bool | `true` | Whether or not to create a headless service |
| cluster.service.labels | object | `{}` | Labels to apply to each nifi service |
| cluster.sidecarConfigs | list | `[]` | list of additional sidecar containers to run alongside the nifi pods. See: https://pkg.go.dev/k8s.io/api/core/v1#Container |
| cluster.singleUserConfiguration | object | `{"authorizerEnabled":false,"enabled":false,"secretKeys":{"password":"password","username":"username"},"secretRef":{"name":"single-user-credentials","namespace":"nifi"}}` | see https://konpyutaika.github.io/nifikop/docs/5_references/1_nifi_cluster#singleuserconfiguration |
| cluster.topologySpreadConstraints | list | `[]` | specifies any TopologySpreadConstraint objects to be applied to all nodes. See https://pkg.go.dev/k8s.io/api/core/v1#TopologySpreadConstraint |
| cluster.type | string | internal | type of the cluster: internal or external. Operator will put internal by default |
| cluster.zkAddress | string | `"nifi-cluster-zookeeper:2181"` | the hostname and port of the zookeeper service |
| cluster.zkPath | string | `"/cluster"` | the path in zookeeper to store this cluster's state |
| cluster.zookeeperProperties | object | `{"overrideConfigs":"initLimit=15\nautopurge.purgeInterval=24\nsyncLimit=5\ntickTime=2000\ndataDir=./state/zookeeper\nautopurge.snapRetainCount=30\n"}` | This is only for embedded zookeeper configuration. This is ignored if `zookeeper.enabled` is true. |
| dataflows | list | `[{"bucketId":"","enabled":false,"flowId":"","flowPosition":{"posX":0,"posY":0},"flowVersion":1,"name":"My Special Dataflow","parameterContextRef":{"name":"default","namespace":"nifi"},"registryClientRef":{"name":"default","namespace":"nifi"},"skipInvalidComponent":true,"skipInvalidControllerService":true,"syncMode":"always","updateStrategy":"drain"}]` | Versioned dataflow configurations. This is used to configure versioned dataflows to be deployed to this nifi cluster. Any number may be configured. Note that a _registryClient_ and a _parameterContext_ must be enabled & present in order for a dataflow to be deployed to a cluster. See https://konpyutaika.github.io/nifikop/docs/5_references/5_nifi_dataflow |
| dataflows[0].bucketId | string | `""` | Bucket id can be found in the bucket.yml created when version controlling process groups |
| dataflows[0].flowId | string | `""` | Flow id can be found in the bucket.yml created when version controlling process groups |
| dataflows[0].flowPosition.posX | int | `0` | x coordinate of flow on canvas |
| dataflows[0].flowPosition.posY | int | `0` | y coordinate of flow on canvas |
| dataflows[0].flowVersion | int | `1` | Version of the flow to take from registry |
| dataflows[0].name | string | `"My Special Dataflow"` | Name of the flow |
| dataflows[0].parameterContextRef | object | `{"name":"default","namespace":"nifi"}` | Reference to the ParameterContext object which will be added to this flow |
| dataflows[0].registryClientRef | object | `{"name":"default","namespace":"nifi"}` | reference to the nifi registry client to connect and get versioned flow |
| dataflows[0].syncMode | string | `"always"` | This is one of {never, always, once} |
| extraManifests | list | `[]` | A list of extra Kubernetes manifest with Helm template support, to apply |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `"nginx"` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts | list | `[]` |  |
| ingress.tls | list | `[]` |  |
| logging.enabled | bool | `false` | Whether or not log aggregation via the banzai cloud logging operator is enabled. |
| logging.flow | object | `{"filters":[{"parser":{"parse":{"expression":"/^(?<time>\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2},\\d{3}) (?<level>[^\\s]+) \\[(?<thread>.*)\\] (?<message>.*)$/im","keep_time_key":true,"time_format":"%iso8601","time_key":"time","time_type":"string","type":"regexp"}}}],"match":[{"select":{"labels":{"app":"nifi"}}}],"name":"nifi-cluster-flow"}` | https://banzaicloud.com/docs/one-eye/logging-operator/configuration/flow/ |
| logging.flow.filters | list | `[{"parser":{"parse":{"expression":"/^(?<time>\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2},\\d{3}) (?<level>[^\\s]+) \\[(?<thread>.*)\\] (?<message>.*)$/im","keep_time_key":true,"time_format":"%iso8601","time_key":"time","time_type":"string","type":"regexp"}}}]` | The filters and match configs should be configured just like in the CRDs (linked above) |
| logging.outputs | object | `{"globalOutputRefs":["loki-cluster-output"]}` | Only global outputs that have been created separately to this helm chart supported for now may consider changing this to a Flow per cluster in future |
| monitoring | object | `{"enabled":false,"internalListenersType":"prometheus","path":"/metrics"}` | Monitoring is enabled by the Prometheus operator. This can be deployed stand-alone or as a part of the Rancher Monitoring application. Do not enable this unless you've installed rancher-logging or the Prometheus operator directly. https://rancher.com/docs/rancher/v2.6/en/monitoring-alerting/ Enabling monitoring creates a `ServiceMonitor` custom resource and routes logs to the output of your choice |
| nodeGroupAutoscalers | list | `[{"downscaleStrategy":"lifo","enabled":false,"horizontalAutoscaler":{"maxReplicas":2,"minReplicas":1},"name":"default-group-autoscaler","nodeConfig":{},"nodeConfigGroupId":"default-group","nodeLabelsSelector":{"matchLabels":{"default-scale-group":"true"}},"readOnlyConfig":{},"replicas":null,"upscaleStrategy":"simple"}]` | Nifi NodeGroup Autoscaler configurations. Use this to autoscale any NodeGroup specified in `cluster.nodeConfigGroups`. To autoscale  See https://konpyutaika.github.io/nifikop/docs/5_references/7_nifi_nodegroup_autoscaler |
| parameterContexts | list | `[{"enabled":false,"inheritedParameterContexts":[],"name":"default","parameters":[{"description":"my foo bar property","name":"foo-prop","sensitive":false,"value":"bar-value"},{"description":"my foo bar property","name":"foo-prop-2","sensitive":true,"value":"bar-value-2"}],"secretRefs":[]}]` | Parameter context configurations. This is required if you wish to deploy versioned flows via the dataflow config. However,  it is not required to provide secret refs. You must provide at least one parameter or nifikop will choke on updating dataflows. The `.name` field must be safe in Kubernetes and match the pattern [A-Za-z0-9-] See https://konpyutaika.github.io/nifikop/docs/5_references/4_nifi_parameter_context |
| parameterContexts[0].inheritedParameterContexts | list | `[]` | List of references of Parameter Contexts from which this one inherits parameters |
| parameterContexts[0].secretRefs | list | `[]` | Use the given secret and put its values as sensitive values in this parameter context. The key will be the name of parameter in NiFi. |
| registryClients | list | `[{"description":"Default NiFi Registry client","enabled":false,"endpoint":"http://nifi-registry","name":"default"},{"description":"Alternate NiFi Registry client","enabled":false,"endpoint":"http://nifi-registry","name":"alternate"}]` | registry client configurations. You'd use this to version control process groups & store the configuration in a registry bucket This is required if you wish to deploy versioned flows via the dataflow config The .name field must be safe in Kubernetes and match the pattern [A-Za-z0-9-] See https://konpyutaika.github.io/nifikop/docs/5_references/3_nifi_registry_client |
| userGroups | list | `[]` | Configure user groups. Each will result in the creation of a `NiFiUserGroup` CRD in k8s, which the operator takes and applies to each nifi configuration. See all properties here: https://konpyutaika.github.io/nifikop/docs/5_references/6_nifi_usergroup |
| users | list | `[]` | Configure users. Each will result in the creation of a `NiFiUser` CRD in k8s, which the operator takes and applies to each nifi configuration. See https://konpyutaika.github.io/nifikop/docs/5_references/2_nifi_user. The object's `name` is used for k8s resource `metadata.name` and so should be alphanumeric and hyphenated & <= 64 bytes |
| zookeeper | object | `{"enabled":false,"persistence":{"size":"10Gi","storageClass":"standard"},"replicaCount":1,"resources":{"limits":{"cpu":2,"memory":"500Mi"},"requests":{"cpu":"0.5m","memory":"250Mi"}}}` | zookeeper chart overrides. Please see all the options for the zookeeper chart here: https://github.com/bitnami/charts/tree/main/bitnami/zookeeper |
| zookeeper.enabled | bool | `false` | Whether or not to deploy an independent zookeeper. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
