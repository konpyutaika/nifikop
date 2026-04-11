---
id: 1_using_keda
title: Using KEDA
sidebar_label: Using KEDA
---

## Deploy KDEA

### What is KEDA ?

[KEDA] is a Kubernetes-based Event Driven Autoscaler. With [KEDA], you can drive the scaling of any container in Kubernetes based on the number of events needing to be processed.

[KEDA] is a single-purpose and lightweight component that can be added into any Kubernetes cluster. [KEDA] works alongside standard Kubernetes components like the Horizontal Pod Autoscaler and can extend functionality without overwriting or duplication. With [KEDA] you can explicitly map the apps you want to use event-driven scale, with other apps continuing to function. This makes [KEDA] a flexible and safe option to run alongside any number of any other Kubernetes applications or frameworks.

[KEDA] can be a very powerful tool for integration with NiFi because we can auto-scale based on a service that your DataFlow will consume (e.g. PubSub, etc.) or with NiFi metrics exposed using Prometheus.

### Deployment

Following the [documentation](https://keda.sh/docs/2.8/deploy/) here are the steps to deploy KEDA.

Deploying KEDA with Helm is very simple:

- Add Helm repo

````console
helm repo add kedacore https://kedacore.github.io/charts
````

- Update Helm repo

````console
helm repo update
````

- Install keda Helm chart

```console
kubectl create namespace keda
helm install keda kedacore/keda --namespace keda
```

[KEDA]: https://keda.sh/

### Deploy NiFI cluster

Use your own NiFi cluster deployment, for this example we will add a specific `NodeConfigGroup` which will be used for auto-scaling nodes, and add the configuration for Prometheus:

```yaml
...
spec:
  ...
  listenersConfig:
    internalListeners:
    ...
    - containerPort: 9090
      name: prometheus
      type: prometheus 
    ...
  ...
  nodeConfigGroups:
    auto_scaling:
      isNode: true
      resourcesRequirements:
        limits:
          cpu: "2"
          memory: 2Gi
        requests:
          cpu: "1"
          memory: 1Gi
      serviceAccountName: external-dns
      storageConfigs:
        - mountPath: /opt/nifi/nifi-current/logs
          name: logs
          metadata:
            labels:
              my-label: my-value
            annotations:
              my-annotation: my-value
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            resources:
              requests:
                storage: 10Gi
            storageClassName: ssd-wait
        - mountPath: /opt/nifi/data
          name: data
          metadata:
            labels:
              my-label: my-value
            annotations:
              my-annotation: my-value
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            resources:
              requests:
                storage: 10Gi
            storageClassName: ssd-wait
        - mountPath: /opt/nifi/extensions
          name: extensions-repository
          metadata:
            labels:
              my-label: my-value
            annotations:
              my-annotation: my-value
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            resources:
              requests:
                storage: 10Gi
            storageClassName: ssd-wait
        - mountPath: /opt/nifi/flowfile_repository
          name: flowfile-repository
          metadata:
            labels:
              my-label: my-value
            annotations:
              my-annotation: my-value
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            resources:
              requests:
                storage: 10Gi
            storageClassName: ssd-wait
        - mountPath: /opt/nifi/nifi-current/conf
          name: conf
          metadata:
            labels:
              my-label: my-value
            annotations:
              my-annotation: my-value
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            resources:
              requests:
                storage: 10Gi
            storageClassName: ssd-wait
        - mountPath: /opt/nifi/content_repository
          name: content-repository
          metadata:
            labels:
              my-label: my-value
            annotations:
              my-annotation: my-value
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            resources:
              requests:
                storage: 10Gi
            storageClassName: ssd-wait
        - mountPath: /opt/nifi/provenance_repository
          name: provenance-repository
          metadata:
            labels:
              my-label: my-value
            annotations:
              my-annotation: my-value
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            resources:
              requests:
                storage: 10Gi
            storageClassName: ssd-wait
  ...
```

### Deploy NiFi cluster autoscaling group

Now we will deploy a `NifiNodeGroupAutoscaler` to define how and what we want to autoscale: 

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiNodeGroupAutoscaler
metadata:
  name: nifinodegroupautoscaler-sample
  namespace: clusters
spec:
  # contains the reference to the NifiCluster with the one the node group autoscaler is linked.
  clusterRef:
    name: cluster
    namespace: clusters
  # defines the id of the NodeConfig contained in NifiCluster.Spec.NodeConfigGroups
  nodeConfigGroupId: auto_scaling
  # readOnlyConfig can be used to pass Nifi node config
  # which has type read-only these config changes will trigger rolling upgrade
  readOnlyConfig:
    nifiProperties:
      overrideConfigs: |
        nifi.ui.banner.text=NiFiKop - Scale Group
  # This is an example of a node config you can apply to each replica in this node group.
  # Any settings here will override those in the configured nodeConfigGroupId
#  nodeConfig:
#    nodeSelector:
#      node_type: high-mem
  # The selector used to identify nodes in NifiCluster.Spec.Nodes this autoscaler will manage
  # Use Node.Labels in combination with this selector to clearly define which nodes will be managed by this autoscaler 
  nodeLabelsSelector: 
    matchLabels:
      nifi_cr: cluster
      nifi_node_group: auto-scaling
  # the strategy used to decide how to add nodes to a nifi cluster
  upscaleStrategy: simple
  # the strategy used to decide how to remove nodes from an existing cluster
  downscaleStrategy: lifo
```

Here we will autoscale using the `NodeConfigGroup`: auto_scaling.

### Deploy Prometheus

For this example, we will base the autoscaling on some metrics of NiFi cluster, to deploy Prometheus we use [prometheus operator](https://github.com/prometheus-operator/prometheus-operator).

- Create dedicated namespace: 

```console
kubectl create namespace monitoring-system
```

- Add Helm repo

````console
helm repo add prometheus https://prometheus-community.github.io/helm-charts
````

- Update Helm repo

````console
helm repo update
````

- Deploy prometheus operator

```console
helm install prometheus prometheus/kube-prometheus-stack --namespace monitoring-system \
    --set prometheusOperator.createCustomResource=false \
    --set prometheusOperator.logLevel=debug \
    --set prometheusOperator.alertmanagerInstanceNamespaces=monitoring-system \
    --set prometheusOperator.namespaces.additional[0]=monitoring-system \
    --set prometheusOperator.prometheusInstanceNamespaces=monitoring-system \
    --set prometheusOperator.thanosRulerInstanceNamespaces=monitoring-system \
    --set defaultRules.enable=false \
    --set alertmanager.enabled=false \
    --set grafana.enabled=false \
    --set kubeApiServer.enabled=false \
    --set kubelet.enabled=false \
    --set kubeControllerManager.enabled=false \
    --set coreDNS.enabled=false \
    --set kubeEtcd.enabled=false \
    --set kubeScheduler.enabled=false \
    --set kubeProxy.enabled=false \
    --set kubeStateMetrics.enabled=false \
    --set prometheus.enabled=false
```

- Deploy the `ServiceAccount`, `ClusterRole` and `ClusterRoleBinding` resources:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus
  namespace: monitoring-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus
rules:
- apiGroups: [""]
  resources:
  - nodes
  - nodes/metrics
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources:
  - configmaps
  verbs: ["get"]
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses
  verbs: ["get", "list", "watch"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prometheus
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus
subjects:
- kind: ServiceAccount
  name: prometheus
  namespace: monitoring-system
```

- Deploy the `Prometheus` resource: 

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring-system
spec:
  enableAdminAPI: false
  evaluationInterval: 30s
  logLevel: debug
  podMonitorSelector:
    matchExpressions:
    - key: release
      operator: In
      values:
      - prometheus
  resources:
    requests:
      memory: 400Mi
  scrapeInterval: 30s
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchExpressions:
    - key: app
      operator: In
      values:
      - nifi-cluster
```

- Deploy the `ServiceMonitor` resource:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cluster
  namespace: monitoring-system
  labels:
    app: nifi-cluster
    nifi_cr: cluster
spec:
  selector:
    matchLabels:
      app: nifi
      nifi_cr: cluster
  namespaceSelector:
    matchNames:
      - clusters
  endpoints:
    - interval: 10s
      port: prometheus
      path: /metrics
      honorLabels: true
      relabelings:
        - sourceLabels: [__meta_kubernetes_pod_ip]
          separator: ;
          regex: (.*)
          targetLabel: pod_ip
          replacement: $1
          action: replace
        - sourceLabels: [__meta_kubernetes_pod_label_nodeId]
          separator: ;
          regex: (.*)
          targetLabel: nodeId
          replacement: $1
          action: replace
        - sourceLabels: [__meta_kubernetes_pod_label_nifi_cr]
          separator: ;
          regex: (.*)
          targetLabel: nifi_cr
          replacement: $1
          action: replace
```

You should now have a `prometheus-prometheus-0` pod and a `prometheus-operated` service, you can access your prometheus using port forwarding:

```console
kubectl port-forward -n monitoring-system service/prometheus-operated 9090:9090
```

You should be able to connect to your prometheus instance on `http://localhost:9090`, check that you can query your NiFi metrics correctly.

### Deploy Scale object

The last step is to deploy your [ScaledObject](https://keda.sh/docs/2.10/concepts/scaling-deployments/#scaledobject-spec) to define how to scale your NiFi node: 

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: cluster
  namespace: clusters
spec:
  scaleTargetRef:
    apiVersion:    nifi.konpyutaika.com/v1alpha1     # Optional. Default: apps/v1
    kind:          NifiNodeGroupAutoscaler           # Optional. Default: Deployment
    name:          nifinodegroupautoscaler-sample    # Mandatory. Must be in the same namespace as the ScaledObject
    envSourceContainerName: nifi                     # Optional. Default: .spec.template.spec.containers[0]
  pollingInterval:  30                               # Optional. Default: 30 seconds
  cooldownPeriod:   300                              # Optional. Default: 300 seconds
  idleReplicaCount: 0                                # Optional. Default: ignored, must be less than minReplicaCount 
  minReplicaCount:  1                                # Optional. Default: 0
  maxReplicaCount:  3                                # Optional. Default: 100
  fallback:                                          # Optional. Section to specify fallback options
    failureThreshold: 5                              # Mandatory if fallback section is included
    replicas: 1                                      # Mandatory if fallback section is included
  # advanced:                                          # Optional. Section to specify advanced options
  #   restoreToOriginalReplicaCount: true              # Optional. Whether the target resource should be scaled back to original replicas count, after the ScaledObject is deleted
  #   horizontalPodAutoscalerConfig:                   # Optional. Section to specify HPA related options
  #     name: {name-of-hpa-resource}                   # Optional. Default: keda-hpa-{scaled-object-name}
  #     behavior:                                      # Optional. Use to modify HPA's scaling behavior
  #       scaleDown:
  #         stabilizationWindowSeconds: 300 <--- 
  #         policies:
  #         - type: Percent
  #           value: 100
  #           periodSeconds: 15
  triggers:
    - type: prometheus
      metadata:
        serverAddress: http://prometheus-operated.monitoring-system.svc:9090
        metricName: http_requests_total
        threshold: <threshold value>
        query: <prometheus query>
```

Now everything is ready, you must have an `HPA` deployed that manage your `NifiNodeGroupAutoscaler`

```console
kubectl get -n clusters hpa
NAME                REFERENCE                                                TARGETS         MINPODS   MAXPODS   REPLICAS   AGE
keda-hpa-cluster    NifiNodeGroupAutoscaler/nifinodegroupautoscaler-sample     1833m/2 (avg)   1         3         2          25d
```