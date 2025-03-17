"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[88036],{43023:(e,n,s)=>{s.d(n,{R:()=>i,x:()=>r});var a=s(63696);const t={},o=a.createContext(t);function i(e){const n=a.useContext(o);return a.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function r(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(t):e.components||t:i(e.components),a.createElement(o.Provider,{value:n},e.children)}},96995:(e,n,s)=>{s.r(n),s.d(n,{assets:()=>l,contentTitle:()=>r,default:()=>p,frontMatter:()=>i,metadata:()=>a,toc:()=>c});const a=JSON.parse('{"id":"3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/1_using_keda","title":"Using KEDA","description":"Deploy KDEA","source":"@site/versioned_docs/version-v1.1.1/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/1_using_keda.md","sourceDirName":"3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling","slug":"/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/1_using_keda","permalink":"/nifikop/docs/v1.1.1/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/1_using_keda","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.1.1/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/1_using_keda.md","tags":[],"version":"v1.1.1","lastUpdatedBy":"Juldrixx","lastUpdatedAt":1707144987000,"frontMatter":{"id":"1_using_keda","title":"Using KEDA","sidebar_label":"Using KEDA"},"sidebar":"docs","previous":{"title":"Design Principles","permalink":"/nifikop/docs/v1.1.1/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/0_design_principles"},"next":{"title":"External cluster","permalink":"/nifikop/docs/v1.1.1/3_manage_nifi/1_manage_clusters/3_external_cluster"}}');var t=s(62540),o=s(43023);const i={id:"1_using_keda",title:"Using KEDA",sidebar_label:"Using KEDA"},r=void 0,l={},c=[{value:"Deploy KDEA",id:"deploy-kdea",level:2},{value:"What is KEDA ?",id:"what-is-keda-",level:3},{value:"Deployment",id:"deployment",level:3},{value:"Deploy NiFI cluster",id:"deploy-nifi-cluster",level:3},{value:"Deploy NiFi cluster autoscaling group",id:"deploy-nifi-cluster-autoscaling-group",level:3},{value:"Deploy Prometheus",id:"deploy-prometheus",level:3},{value:"Deploy Scale object",id:"deploy-scale-object",level:3}];function d(e){const n={a:"a",code:"code",h2:"h2",h3:"h3",li:"li",p:"p",pre:"pre",ul:"ul",...(0,o.R)(),...e.components};return(0,t.jsxs)(t.Fragment,{children:[(0,t.jsx)(n.h2,{id:"deploy-kdea",children:"Deploy KDEA"}),"\n",(0,t.jsx)(n.h3,{id:"what-is-keda-",children:"What is KEDA ?"}),"\n",(0,t.jsxs)(n.p,{children:[(0,t.jsx)(n.a,{href:"https://keda.sh/",children:"KEDA"})," is a Kubernetes-based Event Driven Autoscaler. With ",(0,t.jsx)(n.a,{href:"https://keda.sh/",children:"KEDA"}),", you can drive the scaling of any container in Kubernetes based on the number of events needing to be processed."]}),"\n",(0,t.jsxs)(n.p,{children:[(0,t.jsx)(n.a,{href:"https://keda.sh/",children:"KEDA"})," is a single-purpose and lightweight component that can be added into any Kubernetes cluster. ",(0,t.jsx)(n.a,{href:"https://keda.sh/",children:"KEDA"})," works alongside standard Kubernetes components like the Horizontal Pod Autoscaler and can extend functionality without overwriting or duplication. With ",(0,t.jsx)(n.a,{href:"https://keda.sh/",children:"KEDA"})," you can explicitly map the apps you want to use event-driven scale, with other apps continuing to function. This makes ",(0,t.jsx)(n.a,{href:"https://keda.sh/",children:"KEDA"})," a flexible and safe option to run alongside any number of any other Kubernetes applications or frameworks."]}),"\n",(0,t.jsxs)(n.p,{children:[(0,t.jsx)(n.a,{href:"https://keda.sh/",children:"KEDA"})," can be a very powerful tool for integration with NiFi because we can auto-scale based on a service that your DataFlow will consume (e.g. PubSub, etc.) or with NiFi metrics exposed using Prometheus."]}),"\n",(0,t.jsx)(n.h3,{id:"deployment",children:"Deployment"}),"\n",(0,t.jsxs)(n.p,{children:["Following the ",(0,t.jsx)(n.a,{href:"https://keda.sh/docs/2.8/deploy/",children:"documentation"})," here are the steps to deploy KEDA."]}),"\n",(0,t.jsx)(n.p,{children:"Deploying KEDA with Helm is very simple:"}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsx)(n.li,{children:"Add Helm repo"}),"\n"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-console",children:"helm repo add kedacore https://kedacore.github.io/charts\n"})}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsx)(n.li,{children:"Update Helm repo"}),"\n"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-console",children:"helm repo update\n"})}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsx)(n.li,{children:"Install keda Helm chart"}),"\n"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-console",children:"kubectl create namespace keda\nhelm install keda kedacore/keda --namespace keda\n"})}),"\n",(0,t.jsx)(n.h3,{id:"deploy-nifi-cluster",children:"Deploy NiFI cluster"}),"\n",(0,t.jsxs)(n.p,{children:["Use your own NiFi cluster deployment, for this example we will add a specific ",(0,t.jsx)(n.code,{children:"NodeConfigGroup"})," which will be used for auto-scaling nodes, and add the configuration for Prometheus:"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-yaml",children:'...\nspec:\n  ...\n  listenersConfig:\n    internalListeners:\n    ...\n    - containerPort: 9090\n      name: prometheus\n      type: prometheus \n    ...\n  ...\n  nodeConfigGroups:\n    auto_scaling:\n      isNode: true\n      resourcesRequirements:\n        limits:\n          cpu: "2"\n          memory: 2Gi\n        requests:\n          cpu: "1"\n          memory: 1Gi\n      serviceAccountName: external-dns\n      storageConfigs:\n        - mountPath: /opt/nifi/nifi-current/logs\n          name: logs\n          pvcSpec:\n            accessModes:\n              - ReadWriteOnce\n            resources:\n              requests:\n                storage: 10Gi\n            storageClassName: ssd-wait\n        - mountPath: /opt/nifi/data\n          name: data\n          pvcSpec:\n            accessModes:\n              - ReadWriteOnce\n            resources:\n              requests:\n                storage: 10Gi\n            storageClassName: ssd-wait\n        - mountPath: /opt/nifi/extensions\n          name: extensions-repository\n          pvcSpec:\n            accessModes:\n              - ReadWriteOnce\n            resources:\n              requests:\n                storage: 10Gi\n            storageClassName: ssd-wait\n        - mountPath: /opt/nifi/flowfile_repository\n          name: flowfile-repository\n          pvcSpec:\n            accessModes:\n              - ReadWriteOnce\n            resources:\n              requests:\n                storage: 10Gi\n            storageClassName: ssd-wait\n        - mountPath: /opt/nifi/nifi-current/conf\n          name: conf\n          pvcSpec:\n            accessModes:\n              - ReadWriteOnce\n            resources:\n              requests:\n                storage: 10Gi\n            storageClassName: ssd-wait\n        - mountPath: /opt/nifi/content_repository\n          name: content-repository\n          pvcSpec:\n            accessModes:\n              - ReadWriteOnce\n            resources:\n              requests:\n                storage: 10Gi\n            storageClassName: ssd-wait\n        - mountPath: /opt/nifi/provenance_repository\n          name: provenance-repository\n          pvcSpec:\n            accessModes:\n              - ReadWriteOnce\n            resources:\n              requests:\n                storage: 10Gi\n            storageClassName: ssd-wait\n  ...\n'})}),"\n",(0,t.jsx)(n.h3,{id:"deploy-nifi-cluster-autoscaling-group",children:"Deploy NiFi cluster autoscaling group"}),"\n",(0,t.jsxs)(n.p,{children:["Now we will deploy a ",(0,t.jsx)(n.code,{children:"NifiNodeGroupAutoscaler"})," to define how and what we want to autoscale:"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-yaml",children:"apiVersion: nifi.konpyutaika.com/v1alpha1\nkind: NifiNodeGroupAutoscaler\nmetadata:\n  name: nifinodegroupautoscaler-sample\n  namespace: clusters\nspec:\n  # contains the reference to the NifiCluster with the one the node group autoscaler is linked.\n  clusterRef:\n    name: cluster\n    namespace: clusters\n  # defines the id of the NodeConfig contained in NifiCluster.Spec.NodeConfigGroups\n  nodeConfigGroupId: auto_scaling\n  # readOnlyConfig can be used to pass Nifi node config\n  # which has type read-only these config changes will trigger rolling upgrade\n  readOnlyConfig:\n    nifiProperties:\n      overrideConfigs: |\n        nifi.ui.banner.text=NiFiKop - Scale Group\n  # This is an example of a node config you can apply to each replica in this node group.\n  # Any settings here will override those in the configured nodeConfigGroupId\n#  nodeConfig:\n#    nodeSelector:\n#      node_type: high-mem\n  # The selector used to identify nodes in NifiCluster.Spec.Nodes this autoscaler will manage\n  # Use Node.Labels in combination with this selector to clearly define which nodes will be managed by this autoscaler \n  nodeLabelsSelector: \n    matchLabels:\n      nifi_cr: cluster\n      nifi_node_group: auto-scaling\n  # the strategy used to decide how to add nodes to a nifi cluster\n  upscaleStrategy: simple\n  # the strategy used to decide how to remove nodes from an existing cluster\n  downscaleStrategy: lifo\n"})}),"\n",(0,t.jsxs)(n.p,{children:["Here we will autoscale using the ",(0,t.jsx)(n.code,{children:"NodeConfigGroup"}),": auto_scaling."]}),"\n",(0,t.jsx)(n.h3,{id:"deploy-prometheus",children:"Deploy Prometheus"}),"\n",(0,t.jsxs)(n.p,{children:["For this example, we will base the autoscaling on some metrics of NiFi cluster, to deploy Prometheus we use ",(0,t.jsx)(n.a,{href:"https://github.com/prometheus-operator/prometheus-operator",children:"prometheus operator"}),"."]}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsx)(n.li,{children:"Create dedicated namespace:"}),"\n"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-console",children:"kubectl create namespace monitoring-system\n"})}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsx)(n.li,{children:"Add Helm repo"}),"\n"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-console",children:"helm repo add prometheus https://prometheus-community.github.io/helm-charts\n"})}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsx)(n.li,{children:"Update Helm repo"}),"\n"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-console",children:"helm repo update\n"})}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsx)(n.li,{children:"Deploy prometheus operator"}),"\n"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-console",children:"helm install prometheus prometheus/kube-prometheus-stack --namespace monitoring-system \\\n    --set prometheusOperator.createCustomResource=false \\\n    --set prometheusOperator.logLevel=debug \\\n    --set prometheusOperator.alertmanagerInstanceNamespaces=monitoring-system \\\n    --set prometheusOperator.namespaces.additional[0]=monitoring-system \\\n    --set prometheusOperator.prometheusInstanceNamespaces=monitoring-system \\\n    --set prometheusOperator.thanosRulerInstanceNamespaces=monitoring-system \\\n    --set defaultRules.enable=false \\\n    --set alertmanager.enabled=false \\\n    --set grafana.enabled=false \\\n    --set kubeApiServer.enabled=false \\\n    --set kubelet.enabled=false \\\n    --set kubeControllerManager.enabled=false \\\n    --set coreDNS.enabled=false \\\n    --set kubeEtcd.enabled=false \\\n    --set kubeScheduler.enabled=false \\\n    --set kubeProxy.enabled=false \\\n    --set kubeStateMetrics.enabled=false \\\n    --set prometheus.enabled=false\n"})}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsx)(n.li,{children:"Deploy the prometheus resource:"}),"\n"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-yaml",children:"apiVersion: monitoring.coreos.com/v1\nkind: Prometheus\nmetadata:\n  name: prometheus\n  namespace: monitoring-system\nspec:\n  enableAdminAPI: false\n  evaluationInterval: 30s\n  logLevel: debug\n  podMonitorSelector:\n    matchExpressions:\n    - key: release\n      operator: In\n      values:\n      - prometheus\n  resources:\n    requests:\n      memory: 400Mi\n  scrapeInterval: 30s\n  serviceAccountName: prometheus\n  serviceMonitorSelector:\n    matchExpressions:\n    - key: app\n      operator: In\n      values:\n      - nifi-cluster\n"})}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsxs)(n.li,{children:["Deploy ",(0,t.jsx)(n.code,{children:"ServiceMonitor"})," resource:"]}),"\n"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-yaml",children:"apiVersion: monitoring.coreos.com/v1\nkind: ServiceMonitor\nmetadata:\n  name: cluster\n  namespace: monitoring-system\n  labels:\n    app: nifi-cluster\n    nifi_cr: cluster\nspec:\n  selector:\n    matchLabels:\n      app: nifi\n      nifi_cr: cluster\n  namespaceSelector:\n    matchNames:\n      - clusters\n  endpoints:\n    - interval: 10s\n      port: prometheus\n      path: /metrics\n      honorLabels: true\n      relabelings:\n        - sourceLabels: [__meta_kubernetes_pod_ip]\n          separator: ;\n          regex: (.*)\n          targetLabel: pod_ip\n          replacement: $1\n          action: replace\n        - sourceLabels: [__meta_kubernetes_pod_label_nodeId]\n          separator: ;\n          regex: (.*)\n          targetLabel: nodeId\n          replacement: $1\n          action: replace\n        - sourceLabels: [__meta_kubernetes_pod_label_nifi_cr]\n          separator: ;\n          regex: (.*)\n          targetLabel: nifi_cr\n          replacement: $1\n          action: replace\n"})}),"\n",(0,t.jsxs)(n.p,{children:["You should now have a ",(0,t.jsx)(n.code,{children:"prometheus-prometheus-0"})," pod and a ",(0,t.jsx)(n.code,{children:"prometheus-operated"})," service, you can access your prometheus using port forwarding:"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-console",children:"kubectl port-forward -n monitoring-system service/prometheus-operated 9090:9090\n"})}),"\n",(0,t.jsxs)(n.p,{children:["You should be able to connect to your prometheus instance on ",(0,t.jsx)(n.code,{children:"http://localhost:9090"}),", check that you can query your NiFi metrics correctly."]}),"\n",(0,t.jsx)(n.h3,{id:"deploy-scale-object",children:"Deploy Scale object"}),"\n",(0,t.jsxs)(n.p,{children:["The last step is to deploy your ",(0,t.jsx)(n.a,{href:"https://keda.sh/docs/2.8/concepts/scaling-deployments/#scaledobject-spec",children:"ScaledObject"})," to define how to scale your NiFi node:"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-yaml",children:"apiVersion: keda.sh/v1alpha1\nkind: ScaledObject\nmetadata:\n  name: cluster\n  namespace: clusters\nspec:\n  scaleTargetRef:\n    apiVersion:    nifi.konpyutaika.com/v1alpha1     # Optional. Default: apps/v1\n    kind:          NifiNodeGroupAutoscaler           # Optional. Default: Deployment\n    name:          nifinodegroupautoscaler-sample    # Mandatory. Must be in the same namespace as the ScaledObject\n    envSourceContainerName: nifi                     # Optional. Default: .spec.template.spec.containers[0]\n  pollingInterval:  30                               # Optional. Default: 30 seconds\n  cooldownPeriod:   300                              # Optional. Default: 300 seconds\n  idleReplicaCount: 0                                # Optional. Default: ignored, must be less than minReplicaCount \n  minReplicaCount:  1                                # Optional. Default: 0\n  maxReplicaCount:  3                                # Optional. Default: 100\n  fallback:                                          # Optional. Section to specify fallback options\n    failureThreshold: 5                              # Mandatory if fallback section is included\n    replicas: 1                                      # Mandatory if fallback section is included\n  # advanced:                                          # Optional. Section to specify advanced options\n  #   restoreToOriginalReplicaCount: true              # Optional. Whether the target resource should be scaled back to original replicas count, after the ScaledObject is deleted\n  #   horizontalPodAutoscalerConfig:                   # Optional. Section to specify HPA related options\n  #     name: {name-of-hpa-resource}                   # Optional. Default: keda-hpa-{scaled-object-name}\n  #     behavior:                                      # Optional. Use to modify HPA's scaling behavior\n  #       scaleDown:\n  #         stabilizationWindowSeconds: 300 <--- \n  #         policies:\n  #         - type: Percent\n  #           value: 100\n  #           periodSeconds: 15\n  triggers:\n    - type: prometheus\n      metadata:\n        serverAddress: http://prometheus-operated.monitoring-system.svc:9090\n        metricName: http_requests_total\n        threshold: <threshold value>\n        query: <prometheus query>\n"})}),"\n",(0,t.jsxs)(n.p,{children:["Now everything is ready, you must have an ",(0,t.jsx)(n.code,{children:"HPA"})," deployed that manage you ",(0,t.jsx)(n.code,{children:"NifiNodeGroupAutoscaler"})]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-console",children:"kubectl get -n clusters hpa\nNAME                REFERENCE                                                TARGETS         MINPODS   MAXPODS   REPLICAS   AGE\nkeda-hpa-cluster    NifiNodeGroupAutoscaler/nifinodegroupautoscaler-sample     1833m/2 (avg)   1         3         2          25d\n"})})]})}function p(e={}){const{wrapper:n}={...(0,o.R)(),...e.components};return n?(0,t.jsx)(n,{...e,children:(0,t.jsx)(d,{...e})}):d(e)}}}]);