# zookeeper

![Version: 7.6.2](https://img.shields.io/badge/Version-7.6.2-informational?style=flat-square) ![AppVersion: 3.7.0](https://img.shields.io/badge/AppVersion-3.7.0-informational?style=flat-square)

A centralized service for maintaining configuration information, naming, providing distributed synchronization, and providing group services for distributed applications.

**Homepage:** <https://github.com/bitnami/charts/tree/master/bitnami/zookeeper>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Bitnami | containers@bitnami.com |  |

## Source Code

* <https://github.com/bitnami/bitnami-docker-zookeeper>
* <https://zookeeper.apache.org/>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | common | 1.x.x |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| allowAnonymousLogin | bool | `true` |  |
| auth.clientPassword | string | `""` |  |
| auth.clientUser | string | `""` |  |
| auth.enabled | bool | `false` |  |
| auth.existingSecret | string | `""` |  |
| auth.serverPasswords | string | `""` |  |
| auth.serverUsers | string | `""` |  |
| autopurge.purgeInterval | int | `0` |  |
| autopurge.snapRetainCount | int | `3` |  |
| clusterDomain | string | `"cluster.local"` |  |
| commonAnnotations | object | `{}` |  |
| commonLabels | object | `{}` |  |
| config | string | `""` |  |
| containerPort | int | `2181` |  |
| containerSecurityContext.enabled | bool | `true` |  |
| containerSecurityContext.runAsNonRoot | bool | `true` |  |
| containerSecurityContext.runAsUser | int | `1001` |  |
| customLivenessProbe | object | `{}` |  |
| customReadinessProbe | object | `{}` |  |
| dataLogDir | string | `""` |  |
| diagnosticMode.args[0] | string | `"infinity"` |  |
| diagnosticMode.command[0] | string | `"sleep"` |  |
| diagnosticMode.enabled | bool | `false` |  |
| electionContainerPort | int | `3888` |  |
| extraDeploy | list | `[]` |  |
| extraVolumeMounts | list | `[]` |  |
| extraVolumes | list | `[]` |  |
| followerContainerPort | int | `2888` |  |
| fourlwCommandsWhitelist | string | `"srvr, mntr, ruok"` |  |
| fullnameOverride | string | `""` |  |
| global.imagePullSecrets | list | `[]` |  |
| global.imageRegistry | string | `""` |  |
| global.storageClass | string | `""` |  |
| heapSize | int | `1024` |  |
| hostAliases | list | `[]` |  |
| image.debug | bool | `false` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.pullSecrets | list | `[]` |  |
| image.registry | string | `"docker.io"` |  |
| image.repository | string | `"bitnami/zookeeper"` |  |
| image.tag | string | `"3.7.0-debian-10-r264"` |  |
| initContainers | list | `[]` |  |
| initLimit | int | `10` |  |
| jvmFlags | string | `""` |  |
| kubeVersion | string | `""` |  |
| listenOnAllIPs | bool | `false` |  |
| livenessProbe.enabled | bool | `true` |  |
| livenessProbe.failureThreshold | int | `6` |  |
| livenessProbe.initialDelaySeconds | int | `30` |  |
| livenessProbe.periodSeconds | int | `10` |  |
| livenessProbe.probeCommandTimeout | int | `2` |  |
| livenessProbe.successThreshold | int | `1` |  |
| livenessProbe.timeoutSeconds | int | `5` |  |
| logLevel | string | `"ERROR"` |  |
| maxClientCnxns | int | `60` |  |
| maxSessionTimeout | int | `40000` |  |
| metrics.containerPort | int | `9141` |  |
| metrics.enabled | bool | `false` |  |
| metrics.prometheusRule.enabled | bool | `false` |  |
| metrics.prometheusRule.namespace | string | `""` |  |
| metrics.prometheusRule.rules | list | `[]` |  |
| metrics.prometheusRule.selector | object | `{}` |  |
| metrics.service.annotations."prometheus.io/path" | string | `"/metrics"` |  |
| metrics.service.annotations."prometheus.io/port" | string | `"{{ .Values.metrics.service.port }}"` |  |
| metrics.service.annotations."prometheus.io/scrape" | string | `"true"` |  |
| metrics.service.port | int | `9141` |  |
| metrics.service.type | string | `"ClusterIP"` |  |
| metrics.serviceMonitor.additionalLabels | object | `{}` |  |
| metrics.serviceMonitor.enabled | bool | `false` |  |
| metrics.serviceMonitor.interval | string | `""` |  |
| metrics.serviceMonitor.metricRelabelings | list | `[]` |  |
| metrics.serviceMonitor.namespace | string | `""` |  |
| metrics.serviceMonitor.relabelings | list | `[]` |  |
| metrics.serviceMonitor.scrapeTimeout | string | `""` |  |
| metrics.serviceMonitor.selector | object | `{}` |  |
| minServerId | int | `1` |  |
| nameOverride | string | `""` |  |
| namespaceOverride | string | `""` |  |
| networkPolicy.allowExternal | bool | `true` |  |
| networkPolicy.enabled | bool | `false` |  |
| nodeAffinityPreset.key | string | `""` |  |
| nodeAffinityPreset.type | string | `""` |  |
| nodeAffinityPreset.values | list | `[]` |  |
| nodeSelector | object | `{}` |  |
| persistence.accessModes[0] | string | `"ReadWriteOnce"` |  |
| persistence.annotations | object | `{}` |  |
| persistence.dataLogDir.existingClaim | string | `""` |  |
| persistence.dataLogDir.selector | object | `{}` |  |
| persistence.dataLogDir.size | string | `"8Gi"` |  |
| persistence.enabled | bool | `true` |  |
| persistence.existingClaim | string | `""` |  |
| persistence.selector | object | `{}` |  |
| persistence.size | string | `"8Gi"` |  |
| persistence.storageClass | string | `""` |  |
| podAffinityPreset | string | `""` |  |
| podAnnotations | object | `{}` |  |
| podAntiAffinityPreset | string | `"soft"` |  |
| podDisruptionBudget.maxUnavailable | int | `1` |  |
| podLabels | object | `{}` |  |
| podManagementPolicy | string | `"Parallel"` |  |
| podSecurityContext.enabled | bool | `true` |  |
| podSecurityContext.fsGroup | int | `1001` |  |
| preAllocSize | int | `65536` |  |
| priorityClassName | string | `""` |  |
| readinessProbe.enabled | bool | `true` |  |
| readinessProbe.failureThreshold | int | `6` |  |
| readinessProbe.initialDelaySeconds | int | `5` |  |
| readinessProbe.periodSeconds | int | `10` |  |
| readinessProbe.probeCommandTimeout | int | `2` |  |
| readinessProbe.successThreshold | int | `1` |  |
| readinessProbe.timeoutSeconds | int | `5` |  |
| replicaCount | int | `1` |  |
| resources.requests.cpu | string | `"250m"` |  |
| resources.requests.memory | string | `"256Mi"` |  |
| rollingUpdatePartition | string | `""` |  |
| schedulerName | string | `""` |  |
| service.annotations | object | `{}` |  |
| service.disableBaseClientPort | bool | `false` |  |
| service.electionPort | int | `3888` |  |
| service.followerPort | int | `2888` |  |
| service.headless.annotations | object | `{}` |  |
| service.loadBalancerIP | string | `""` |  |
| service.nodePorts.client | string | `""` |  |
| service.nodePorts.clientTls | string | `""` |  |
| service.port | int | `2181` |  |
| service.publishNotReadyAddresses | bool | `true` |  |
| service.tlsClientPort | int | `3181` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.automountServiceAccountToken | bool | `true` |  |
| serviceAccount.create | bool | `false` |  |
| serviceAccount.name | string | `""` |  |
| sidecars | list | `[]` |  |
| snapCount | int | `100000` |  |
| syncLimit | int | `5` |  |
| tickTime | int | `2000` |  |
| tls.client.autoGenerated | bool | `false` |  |
| tls.client.enabled | bool | `false` |  |
| tls.client.existingSecret | string | `""` |  |
| tls.client.keystorePassword | string | `""` |  |
| tls.client.keystorePath | string | `"/opt/bitnami/zookeeper/config/certs/client/zookeeper.keystore.jks"` |  |
| tls.client.passwordsSecretName | string | `""` |  |
| tls.client.truststorePassword | string | `""` |  |
| tls.client.truststorePath | string | `"/opt/bitnami/zookeeper/config/certs/client/zookeeper.truststore.jks"` |  |
| tls.quorum.autoGenerated | bool | `false` |  |
| tls.quorum.enabled | bool | `false` |  |
| tls.quorum.existingSecret | string | `""` |  |
| tls.quorum.keystorePassword | string | `""` |  |
| tls.quorum.keystorePath | string | `"/opt/bitnami/zookeeper/config/certs/quorum/zookeeper.keystore.jks"` |  |
| tls.quorum.passwordsSecretName | string | `""` |  |
| tls.quorum.truststorePassword | string | `""` |  |
| tls.quorum.truststorePath | string | `"/opt/bitnami/zookeeper/config/certs/quorum/zookeeper.truststore.jks"` |  |
| tls.resources.limits | object | `{}` |  |
| tls.resources.requests | object | `{}` |  |
| tlsContainerPort | int | `3181` |  |
| tolerations | list | `[]` |  |
| topologySpreadConstraints | object | `{}` |  |
| updateStrategy | string | `"RollingUpdate"` |  |
| volumePermissions.containerSecurityContext.runAsUser | int | `0` |  |
| volumePermissions.enabled | bool | `false` |  |
| volumePermissions.image.pullPolicy | string | `"IfNotPresent"` |  |
| volumePermissions.image.pullSecrets | list | `[]` |  |
| volumePermissions.image.registry | string | `"docker.io"` |  |
| volumePermissions.image.repository | string | `"bitnami/bitnami-shell"` |  |
| volumePermissions.image.tag | string | `"10-debian-10-r311"` |  |
| volumePermissions.resources | object | `{}` |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.5.0](https://github.com/norwoodj/helm-docs/releases/v1.5.0)
