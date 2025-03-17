"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[20658],{43023:(e,i,t)=>{t.d(i,{R:()=>d,x:()=>l});var s=t(63696);const n={},r=s.createContext(n);function d(e){const i=s.useContext(r);return s.useMemo((function(){return"function"==typeof e?e(i):{...i,...e}}),[i,e])}function l(e){let i;return i=e.disableParentContext?"function"==typeof e.components?e.components(n):e.components||n:d(e.components),s.createElement(r.Provider,{value:i},e.children)}},46375:(e,i,t)=>{t.r(i),t.d(i,{assets:()=>c,contentTitle:()=>l,default:()=>a,frontMatter:()=>d,metadata:()=>s,toc:()=>h});const s=JSON.parse('{"id":"5_references/1_nifi_cluster/1_nifi_cluster","title":"NiFi cluster","description":"NifiCluster describes the desired state of the NiFi cluster we want to setup through the operator.","source":"@site/versioned_docs/version-v0.12.0/5_references/1_nifi_cluster/1_nifi_cluster.md","sourceDirName":"5_references/1_nifi_cluster","slug":"/5_references/1_nifi_cluster/","permalink":"/nifikop/docs/v0.12.0/5_references/1_nifi_cluster/","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v0.12.0/5_references/1_nifi_cluster/1_nifi_cluster.md","tags":[],"version":"v0.12.0","lastUpdatedBy":"Juldrixx","lastUpdatedAt":1707144987000,"frontMatter":{"id":"1_nifi_cluster","title":"NiFi cluster","sidebar_label":"NiFi cluster"},"sidebar":"docs","previous":{"title":"NiFi Users and Groups","permalink":"/nifikop/docs/v0.12.0/3_tasks/4_nifi_user_group"},"next":{"title":"Read only configurations","permalink":"/nifikop/docs/v0.12.0/5_references/1_nifi_cluster/2_read_only_config"}}');var n=t(62540),r=t(43023);const d={id:"1_nifi_cluster",title:"NiFi cluster",sidebar_label:"NiFi cluster"},l=void 0,c={},h=[{value:"NifiCluster",id:"nificluster",level:2},{value:"NifiClusterSpec",id:"nificlusterspec",level:2},{value:"NifiClusterStatus",id:"nificlusterstatus",level:2},{value:"ServicePolicy",id:"servicepolicy",level:2},{value:"PodPolicy",id:"podpolicy",level:2},{value:"ManagedUsers",id:"managedusers",level:2},{value:"DisruptionBudget",id:"disruptionbudget",level:2},{value:"LdapConfiguration",id:"ldapconfiguration",level:2},{value:"NifiClusterTaskSpec",id:"nificlustertaskspec",level:2},{value:"ClusterState",id:"clusterstate",level:2}];function o(e){const i={a:"a",code:"code",h2:"h2",p:"p",pre:"pre",strong:"strong",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,r.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsxs)(i.p,{children:[(0,n.jsx)(i.code,{children:"NifiCluster"})," describes the desired state of the NiFi cluster we want to setup through the operator."]}),"\n",(0,n.jsx)(i.pre,{children:(0,n.jsx)(i.code,{className:"language-yaml",children:"apiVersion: nifi.konpyutaika.com/v1alpha1\nkind: NifiCluster\nmetadata:\n  name: simplenifi\nspec:\n  service:\n    headlessEnabled: true\n    annotations:\n      tyty: ytyt\n    labels:\n      tete: titi  \n  pod:\n    annotations:\n      toto: tata\n    labels:\n      titi: tutu\n  zkAddress: 'zookeepercluster-client.zookeeper:2181'\n  zkPath: '/simplenifi'\n  clusterImage: 'apache/nifi:1.11.3'\n  oneNifiNodePerNode: false\n  nodeConfigGroups:\n    default_group:\n      isNode: true\n      podMetadata:\n        annotations:\n          node-annotation: \"node-annotation-value\"\n        labels:\n          node-label: \"node-label-value\"\n      externalVolumeConfigs:\n        - name: example-volume\n          mountPath: \"/opt/nifi/example\"\n          secret:\n            secretName: \"raw-controller\"\n      storageConfigs:\n        - mountPath: '/opt/nifi/nifi-current/logs'\n          name: logs\n          pvcSpec:\n            accessModes:\n              - ReadWriteOnce\n            storageClassName: 'standard'\n            resources:\n              requests:\n                storage: 10Gi\n      serviceAccountName: 'default'\n      resourcesRequirements:\n        limits:\n          cpu: '2'\n          memory: 3Gi\n        requests:\n          cpu: '1'\n          memory: 1Gi\n  nodes:\n    - id: 1\n      nodeConfigGroup: 'default_group'\n    - id: 2\n      nodeConfigGroup: 'default_group'\n  propagateLabels: true\n  nifiClusterTaskSpec:\n    retryDurationMinutes: 10\n  listenersConfig:\n    internalListeners:\n      - type: 'http'\n        name: 'http'\n        containerPort: 8080\n      - type: 'cluster'\n        name: 'cluster'\n        containerPort: 6007\n      - type: 's2s'\n        name: 's2s'\n        containerPort: 10000\n  externalServices:\n    - name: 'clusterip'\n      spec:\n        type: ClusterIP\n        portConfigs:\n          - port: 8080\n            internalListenerName: 'http'\n      metadata:\n        annotations:\n          toto: tata\n        labels:\n          titi: tutu\n"})}),"\n",(0,n.jsx)(i.h2,{id:"nificluster",children:"NifiCluster"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"metadata"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta",children:"ObjectMetadata"})}),(0,n.jsx)(i.td,{children:"is metadata that all persisted resources must have, which includes all objects users must create."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"spec"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"#nificlusterspec",children:"NifiClusterSpec"})}),(0,n.jsx)(i.td,{children:"defines the desired state of NifiCluster."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"status"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"#nificlusterstatus",children:"NifiClusterStatus"})}),(0,n.jsx)(i.td,{children:"defines the observed state of NifiCluster."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]})]})]}),"\n",(0,n.jsx)(i.h2,{id:"nificlusterspec",children:"NifiClusterSpec"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"clientType"}),(0,n.jsxs)(i.td,{children:["Enum=","basic"]}),(0,n.jsx)(i.td,{children:"defines if the operator will use basic or tls authentication to query the NiFi cluster."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"tls"})})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"type"}),(0,n.jsxs)(i.td,{children:["Enum=","internal"]}),(0,n.jsx)(i.td,{children:"defines if the cluster is internal (i.e manager by the operator) or external."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.code,{children:"internal"})})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"nodeURITemplate"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"used to dynamically compute node uri."}),(0,n.jsx)(i.td,{children:"if external type"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"nifiURI"}),(0,n.jsx)(i.td,{children:"stringused access through a LB uri."}),(0,n.jsx)(i.td,{children:"if external type"}),(0,n.jsx)(i.td,{children:"-"}),(0,n.jsx)(i.td,{})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"rootProcessGroupId"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"contains the uuid of the root process group for this cluster."}),(0,n.jsx)(i.td,{children:"if external type"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"secretRef"}),(0,n.jsxs)(i.td,{children:["[\xa0]",(0,n.jsx)(i.a,{href:"../4_nifi_parameter_context#secretreference",children:"SecretReference"})]}),(0,n.jsx)(i.td,{children:"reference the secret containing the informations required to authentiticate to the cluster."}),(0,n.jsx)(i.td,{children:"if external type"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"proxyUrl"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"defines the proxy required to query the NiFi cluster."}),(0,n.jsx)(i.td,{children:"if external type"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"service"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"#servicepolicy",children:"ServicePolicy"})}),(0,n.jsx)(i.td,{children:"defines the policy for services owned by NiFiKop operator."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"pod"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"#podpolicy",children:"PodPolicy"})}),(0,n.jsx)(i.td,{children:"defines the policy for pod owned by NiFiKop operator."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"zkAddress"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsxs)(i.td,{children:["specifies the ZooKeeper connection string in the form hostname",":port"," where host and port are those of a Zookeeper server."]}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:'""'})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"zkPath"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"specifies the Zookeeper chroot path as part of its Zookeeper connection string which puts its data under same path in the global ZooKeeper namespace."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:'"/"'})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"initContainerImage"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"can override the default image used into the init container to check if ZoooKeeper server is reachable."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:'"busybox"'})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"initContainers"}),(0,n.jsx)(i.td,{children:"[\xa0]string"}),(0,n.jsx)(i.td,{children:"defines additional initContainers configurations."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"[\xa0]"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"clusterImage"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"can specify the whole nificluster image in one place."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:'""'})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"oneNifiNodePerNode"}),(0,n.jsx)(i.td,{children:"boolean"}),(0,n.jsx)(i.td,{children:"if set to true every nifi node is started on a new node, if there is not enough node to do that it will stay in pending state. If set to false the operator also tries to schedule the nifi node to a unique node but if the node number is insufficient the nifi node will be scheduled to a node where a nifi node is already running."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"propagateLabels"}),(0,n.jsx)(i.td,{children:"boolean"}),(0,n.jsx)(i.td,{children:"-"}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"false"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"managedAdminUsers"}),(0,n.jsxs)(i.td,{children:["[\xa0]",(0,n.jsx)(i.a,{href:"#managedusers",children:"ManagedUser"})]}),(0,n.jsx)(i.td,{children:"contains the list of users that will be added to the managed admin group (with all rights)."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"[]"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"managedReaderUsers"}),(0,n.jsxs)(i.td,{children:["[\xa0]",(0,n.jsx)(i.a,{href:"#managedusers",children:"ManagedUser"})]}),(0,n.jsx)(i.td,{children:"contains the list of users that will be added to the managed admin group (with all rights)."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"[]"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"readOnlyConfig"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"/nifikop/docs/v0.12.0/5_references/1_nifi_cluster/2_read_only_config",children:"ReadOnlyConfig"})}),(0,n.jsx)(i.td,{children:"specifies the read-only type Nifi config cluster wide, all theses will be merged with node specified readOnly configurations, so it can be overwritten per node."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"nodeUserIdentityTemplate"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"specifies the template to be used when naming the node user identity (e.g. node-%d-mysuffix)"}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:'"node-%d-<cluster-name>"'})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"nodeConfigGroups"}),(0,n.jsxs)(i.td,{children:["map[string]",(0,n.jsx)(i.a,{href:"/nifikop/docs/v0.12.0/5_references/1_nifi_cluster/3_node_config",children:"NodeConfig"})]}),(0,n.jsx)(i.td,{children:"specifies multiple node configs with unique name"}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"nodes"}),(0,n.jsxs)(i.td,{children:["[\xa0]",(0,n.jsx)(i.a,{href:"/nifikop/docs/v0.12.0/5_references/1_nifi_cluster/3_node_config",children:"Node"})]}),(0,n.jsx)(i.td,{children:"specifies the list of cluster nodes, all node requires an image, unique id, and storageConfigs settings"}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"disruptionBudget"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"#disruptionbudget",children:"DisruptionBudget"})}),(0,n.jsx)(i.td,{children:"defines the configuration for PodDisruptionBudget."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"ldapConfiguration"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"#ldapconfiguration",children:"LdapConfiguration"})}),(0,n.jsx)(i.td,{children:"specifies the configuration if you want to use LDAP."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"nifiClusterTaskSpec"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"#nificlustertaskspec",children:"NifiClusterTaskSpec"})}),(0,n.jsx)(i.td,{children:"specifies the configuration of the nifi cluster Tasks."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"listenersConfig"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"/nifikop/docs/v0.12.0/5_references/1_nifi_cluster/6_listeners_config",children:"ListenersConfig"})}),(0,n.jsx)(i.td,{children:"specifies nifi's listener specifig configs."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"sidecarConfigs"}),(0,n.jsxs)(i.td,{children:["[\xa0]",(0,n.jsx)(i.a,{href:"https://godoc.org/k8s.io/api/core/v1#Container",children:"Container"})]}),(0,n.jsx)(i.td,{children:"Defines additional sidecar configurations. [Check documentation for more informations]"}),(0,n.jsx)(i.td,{}),(0,n.jsx)(i.td,{})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"externalServices"}),(0,n.jsxs)(i.td,{children:["[\xa0]",(0,n.jsx)(i.a,{href:"/nifikop/docs/v0.12.0/5_references/1_nifi_cluster/7_external_service_config",children:"ExternalServiceConfigs"})]}),(0,n.jsx)(i.td,{children:"specifies settings required to access nifi externally."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"topologySpreadConstraints"}),(0,n.jsxs)(i.td,{children:["[\xa0]",(0,n.jsx)(i.a,{href:"https://godoc.org/k8s.io/api/core/v1#TopologySpreadConstraint",children:"TopologySpreadConstraint"})]}),(0,n.jsx)(i.td,{children:"specifies any TopologySpreadConstraint objects to be applied to all nodes."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"nifiControllerTemplate"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsxs)(i.td,{children:["NifiControllerTemplate specifies the template to be used when naming the node controller (e.g. %s-mysuffix) ",(0,n.jsx)(i.strong,{children:"Warning: once defined don't change this value either the operator will no longer be able to manage the cluster"})]}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:'"%s-controller"'})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"controllerUserIdentity"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsxs)(i.td,{children:["ControllerUserIdentity specifies what to call the static admin user's identity ",(0,n.jsx)(i.strong,{children:"Warning: once defined don't change this value either the operator will no longer be able to manage the cluster"})]}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"false"})]})]})]}),"\n",(0,n.jsx)(i.h2,{id:"nificlusterstatus",children:"NifiClusterStatus"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"nodesState"}),(0,n.jsxs)(i.td,{children:["map[string]",(0,n.jsx)(i.a,{href:"/nifikop/docs/v0.12.0/5_references/1_nifi_cluster/5_node_state",children:"NodeState"})]}),(0,n.jsx)(i.td,{children:"Store the state of each nifi node."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"State"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"#clusterstate",children:"ClusterState"})}),(0,n.jsx)(i.td,{children:"Store the state of each nifi node."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"rootProcessGroupId"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"contains the uuid of the root process group for this cluster."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]})]})]}),"\n",(0,n.jsx)(i.h2,{id:"servicepolicy",children:"ServicePolicy"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"headlessEnabled"}),(0,n.jsx)(i.td,{children:"boolean"}),(0,n.jsx)(i.td,{children:"specifies if the cluster should use headlessService for Nifi or individual services using service per nodes may come an handy case of service mesh."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"false"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"serviceTemplate"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"specifies the template to be used when naming the service."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:'If headlessEnabled = true ? "%s-headless" = "%s-all-node"'})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"annotations"}),(0,n.jsx)(i.td,{children:"map[string]string"}),(0,n.jsx)(i.td,{children:"Annotations specifies the annotations to attach to services the NiFiKop operator creates"}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"labels"}),(0,n.jsx)(i.td,{children:"map[string]string"}),(0,n.jsx)(i.td,{children:"Labels specifies the labels to attach to services the NiFiKop operator creates"}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]})]})]}),"\n",(0,n.jsx)(i.h2,{id:"podpolicy",children:"PodPolicy"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"annotations"}),(0,n.jsx)(i.td,{children:"map[string]string"}),(0,n.jsx)(i.td,{children:"Annotations specifies the annotations to attach to pods the NiFiKop operator creates"}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"labels"}),(0,n.jsx)(i.td,{children:"map[string]string"}),(0,n.jsx)(i.td,{children:"Labels specifies the Labels to attach to pods the NiFiKop operator creates"}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"hostAliases"}),(0,n.jsxs)(i.td,{children:["[\xa0]",(0,n.jsx)(i.a,{href:"https://pkg.go.dev/k8s.io/api/core/v1#HostAlias",children:"HostAlias"})]}),(0,n.jsx)(i.td,{children:"A list of host aliases to include in every pod's /etc/hosts configuration in the scenario where DNS is not available."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"[\xa0]"})]})]})]}),"\n",(0,n.jsx)(i.h2,{id:"managedusers",children:"ManagedUsers"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"identity"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"identity field is use to define the user identity on NiFi cluster side, it use full when the user's name doesn't suite with Kubernetes resource name."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"name"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"name field is use to name the NifiUser resource, if not identity is provided it will be used to name the user on NiFi cluster side."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"-"})]})]})]}),"\n",(0,n.jsx)(i.h2,{id:"disruptionbudget",children:"DisruptionBudget"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"create"}),(0,n.jsx)(i.td,{children:"bool"}),(0,n.jsx)(i.td,{children:"if set to true, will create a podDisruptionBudget."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"budget"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"the budget to set for the PDB, can either be static number or a percentage."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"-"})]})]})]}),"\n",(0,n.jsx)(i.h2,{id:"ldapconfiguration",children:"LdapConfiguration"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"enabled"}),(0,n.jsx)(i.td,{children:"boolean"}),(0,n.jsx)(i.td,{children:"if set to true, we will enable ldap usage into nifi.properties configuration."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"false"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"url"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"space-separated list of URLs of the LDAP servers (i.e. ldap://${hostname}:${port})."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:'""'})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"searchBase"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"base DN for searching for users (i.e. CN=Users,DC=example,DC=com)."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:'""'})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"searchFilter"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsxs)(i.td,{children:["Filter for searching for users against the 'User Search Base'. (i.e. sAMAccountName=",0,"). The user specified name is inserted into '",0,"'."]}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:'""'})]})]})]}),"\n",(0,n.jsx)(i.h2,{id:"nificlustertaskspec",children:"NifiClusterTaskSpec"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsx)(i.tbody,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"retryDurationMinutes"}),(0,n.jsx)(i.td,{children:"int"}),(0,n.jsx)(i.td,{children:"describes the amount of time the Operator waits for the task."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"5"})]})})]}),"\n",(0,n.jsx)(i.h2,{id:"clusterstate",children:"ClusterState"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Name"}),(0,n.jsx)(i.th,{children:"Value"}),(0,n.jsx)(i.th,{children:"Description"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"NifiClusterInitializing"}),(0,n.jsx)(i.td,{children:"ClusterInitializing"}),(0,n.jsx)(i.td,{children:"states that the cluster is still in initializing stage"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"NifiClusterInitialized"}),(0,n.jsx)(i.td,{children:"ClusterInitialized"}),(0,n.jsx)(i.td,{children:"states that the cluster is initialized"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"NifiClusterReconciling"}),(0,n.jsx)(i.td,{children:"ClusterReconciling"}),(0,n.jsx)(i.td,{children:"states that the cluster is still in reconciling stage"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"NifiClusterRollingUpgrading"}),(0,n.jsx)(i.td,{children:"ClusterRollingUpgrading"}),(0,n.jsx)(i.td,{children:"states that the cluster is rolling upgrading"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"NifiClusterRunning"}),(0,n.jsx)(i.td,{children:"ClusterRunning"}),(0,n.jsx)(i.td,{children:"states that the cluster is in running state"})]})]})]})]})}function a(e={}){const{wrapper:i}={...(0,r.R)(),...e.components};return i?(0,n.jsx)(i,{...e,children:(0,n.jsx)(o,{...e})}):o(e)}}}]);