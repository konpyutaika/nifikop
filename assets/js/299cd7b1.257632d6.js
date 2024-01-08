"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[58745],{61816:(e,n,i)=>{i.r(n),i.d(n,{assets:()=>r,contentTitle:()=>a,default:()=>u,frontMatter:()=>o,metadata:()=>l,toc:()=>c});var s=i(24246),t=i(71670);const o={id:"1_scaling_mechanism",title:"Scaling mechanism",sidebar_label:"Scaling mechanism"},a=void 0,l={id:"3_manage_nifi/1_manage_clusters/2_cluster_scaling/1_scaling_mechanism",title:"Scaling mechanism",description:"This tasks shows you how to perform a gracefull cluster scale up and scale down.",source:"@site/versioned_docs/version-v1.4.0/3_manage_nifi/1_manage_clusters/2_cluster_scaling/1_scaling_mechanism.md",sourceDirName:"3_manage_nifi/1_manage_clusters/2_cluster_scaling",slug:"/3_manage_nifi/1_manage_clusters/2_cluster_scaling/1_scaling_mechanism",permalink:"/nifikop/docs/v1.4.0/3_manage_nifi/1_manage_clusters/2_cluster_scaling/1_scaling_mechanism",draft:!1,unlisted:!1,editUrl:"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.4.0/3_manage_nifi/1_manage_clusters/2_cluster_scaling/1_scaling_mechanism.md",tags:[],version:"v1.4.0",lastUpdatedBy:"Juldrixx",lastUpdatedAt:1704729012,formattedLastUpdatedAt:"Jan 8, 2024",frontMatter:{id:"1_scaling_mechanism",title:"Scaling mechanism",sidebar_label:"Scaling mechanism"},sidebar:"docs",previous:{title:"Custom User Authorizers",permalink:"/nifikop/docs/v1.4.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/6_users_authorization/1_custom_user_authorizer"},next:{title:"Design Principles",permalink:"/nifikop/docs/v1.4.0/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/0_design_principles"}},r={},c=[{value:"Before you begin",id:"before-you-begin",level:2},{value:"About this task",id:"about-this-task",level:2},{value:"Scale up : Add a new node",id:"scale-up--add-a-new-node",level:2},{value:"Scaledown : Gracefully remove node",id:"scaledown--gracefully-remove-node",level:2}];function d(e){const n={a:"a",admonition:"admonition",code:"code",h2:"h2",img:"img",li:"li",ol:"ol",p:"p",pre:"pre",strong:"strong",ul:"ul",...(0,t.a)(),...e.components};return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(n.p,{children:"This tasks shows you how to perform a gracefull cluster scale up and scale down."}),"\n",(0,s.jsx)(n.h2,{id:"before-you-begin",children:"Before you begin"}),"\n",(0,s.jsxs)(n.ul,{children:["\n",(0,s.jsxs)(n.li,{children:["Setup NiFiKop by following the instructions in the ",(0,s.jsx)(n.a,{href:"../../../2_deploy_nifikop/1_quick_start",children:"Installation guide"}),"."]}),"\n",(0,s.jsxs)(n.li,{children:["Deploy the ",(0,s.jsx)(n.a,{href:"../1_deploy_cluster/1_quick_start",children:"Simple NiFi"})," sample cluster."]}),"\n",(0,s.jsxs)(n.li,{children:["Review the ",(0,s.jsx)(n.a,{href:"../../../5_references/1_nifi_cluster/4_node",children:"Node"})," references doc."]}),"\n"]}),"\n",(0,s.jsx)(n.h2,{id:"about-this-task",children:"About this task"}),"\n",(0,s.jsxs)(n.p,{children:["The ",(0,s.jsx)(n.a,{href:"../1_deploy_cluster/1_quick_start",children:"Simple NiFi"})," example consists of a three nodes NiFi cluster.\nA node decommission must follow a strict procedure, described in the ",(0,s.jsx)(n.a,{href:"https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#decommission-nodes",children:"NiFi documentation"})," :"]}),"\n",(0,s.jsxs)(n.ol,{children:["\n",(0,s.jsx)(n.li,{children:"Disconnect the node"}),"\n",(0,s.jsx)(n.li,{children:"Once disconnect completes, offload the node."}),"\n",(0,s.jsx)(n.li,{children:"Once offload completes, delete the node."}),"\n",(0,s.jsx)(n.li,{children:"Once the delete request has finished, stop/remove the NiFi service on the host."}),"\n"]}),"\n",(0,s.jsx)(n.p,{children:"For the moment, we have implemented it as follows in the operator :"}),"\n",(0,s.jsxs)(n.ol,{children:["\n",(0,s.jsx)(n.li,{children:"Disconnect the node"}),"\n",(0,s.jsx)(n.li,{children:"Once disconnect completes, offload the node."}),"\n",(0,s.jsx)(n.li,{children:"Once offload completes, delete the pod."}),"\n",(0,s.jsx)(n.li,{children:"Once the pod deletion completes, delete the node."}),"\n",(0,s.jsx)(n.li,{children:"Once the delete request has finished, remove the node from the NifiCluster status."}),"\n"]}),"\n",(0,s.jsx)(n.p,{children:"In addition, we have a regular check that ensure that all nodes have been removed."}),"\n",(0,s.jsx)(n.p,{children:"In this task, you will first perform a scale up, in adding an new node. Then, you will remove another node that the one created, and observe the decommission's steps."}),"\n",(0,s.jsx)(n.h2,{id:"scale-up--add-a-new-node",children:"Scale up : Add a new node"}),"\n",(0,s.jsxs)(n.p,{children:["For this task, we will simply add a node with the same configuration than the other ones, if you want to know more about how to add a node with an other configuration let's have a look to the ",(0,s.jsx)(n.a,{href:"../1_deploy_cluster/2_nodes_configuration",children:"Node configuration"})," documentation page."]}),"\n",(0,s.jsxs)(n.ol,{children:["\n",(0,s.jsx)(n.li,{children:"Add and run a dataflow as the example :"}),"\n"]}),"\n",(0,s.jsx)(n.p,{children:(0,s.jsx)(n.img,{alt:"Scaling dataflow",src:i(90192).Z+"",width:"832",height:"660"})}),"\n",(0,s.jsxs)(n.ol,{start:"2",children:["\n",(0,s.jsxs)(n.li,{children:["Add a new node to the list of ",(0,s.jsx)(n.code,{children:"NifiCluster.Spec.Nodes"})," field, by following the ",(0,s.jsx)(n.a,{href:"../../../5_references/1_nifi_cluster/4_node",children:"Node object definition"})," documentation:"]}),"\n"]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-yaml",children:'apiVersion: nifi.konpyutaika.com/v1\nkind: NifiCluster\nmetadata:\n  name: simplenifi\nspec:\n  service:\n    headlessEnabled: true\n  zkAddress: "zookeepercluster-client.zookeeper:2181"\n  zkPath: "/simplenifi"\n  clusterImage: "apache/nifi:1.12.1"\n  oneNifiNodePerNode: false\n  nodeConfigGroups:\n    default_group:\n      isNode: true\n      storageConfigs:\n        - mountPath: "/opt/nifi/nifi-current/logs"\n          name: logs\n          metadata:\n            labels:\n              my-label: my-value\n            annotations:\n              my-annotation: my-value\n          pvcSpec:\n            accessModes:\n              - ReadWriteOnce\n            storageClassName: "standard"\n            resources:\n              requests:\n                storage: 10Gi\n      serviceAccountName: "default"\n      resourcesRequirements:\n        limits:\n          cpu: "2"\n          memory: 3Gi\n        requests:\n          cpu: "1"\n          memory: 1Gi\n  nodes:\n    - id: 0\n      nodeConfigGroup: "default_group"\n    - id: 1\n      nodeConfigGroup: "default_group"\n    - id: 2\n      nodeConfigGroup: "default_group"\n# >>>> START: The new node\n    - id: 25\n      nodeConfigGroup: "default_group"\n# <<<< END\n  propagateLabels: true\n  nifiClusterTaskSpec:\n    retryDurationMinutes: 10\n  listenersConfig:\n    internalListeners:\n      - type: "http"\n        name: "http"\n        containerPort: 8080\n      - type: "cluster"\n        name: "cluster"\n        containerPort: 6007\n      - type: "s2s"\n        name: "s2s"\n        containerPort: 10000\n'})}),"\n",(0,s.jsx)(n.admonition,{type:"important",children:(0,s.jsxs)(n.p,{children:[(0,s.jsx)(n.strong,{children:"Note :"})," The ",(0,s.jsx)(n.code,{children:"Node.Id"})," field must be unique in the ",(0,s.jsx)(n.code,{children:"NifiCluster.Spec.Nodes"})," list."]})}),"\n",(0,s.jsxs)(n.ol,{start:"3",children:["\n",(0,s.jsxs)(n.li,{children:["Apply the new ",(0,s.jsx)(n.code,{children:"NifiCluster"})," configuration :"]}),"\n"]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-sh",children:"kubectl -n nifi apply -f config/samples/simplenificluster.yaml\n"})}),"\n",(0,s.jsxs)(n.ol,{start:"4",children:["\n",(0,s.jsx)(n.li,{children:"You should now have the following resources into kubernetes :"}),"\n"]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-console",children:"kubectl get pods,configmap,pvc -l nodeId=25\nNAME                          READY   STATUS    RESTARTS   AGE\npod/simplenifi-25-nodem5jh4   1/1     Running   0          11m\n\nNAME                             DATA   AGE\nconfigmap/simplenifi-config-25   7      11m\n\nNAME                                               STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE\npersistentvolumeclaim/simplenifi-25-storagehwn24   Bound    pvc-7da86076-728e-11ea-846d-42010a8400f2   10Gi       RWO            standard       11m\n"})}),"\n",(0,s.jsx)(n.p,{children:"And if you go on the NiFi UI, in the cluster administration page :"}),"\n",(0,s.jsx)(n.p,{children:(0,s.jsx)(n.img,{alt:"Scale up, cluster list",src:i(17530).Z+"",width:"1880",height:"357"})}),"\n",(0,s.jsxs)(n.ol,{start:"5",children:["\n",(0,s.jsx)(n.li,{children:"You now have data on the new node :"}),"\n"]}),"\n",(0,s.jsx)(n.p,{children:(0,s.jsx)(n.img,{alt:"Scale up, cluster distribution",src:i(70486).Z+"",width:"766",height:"568"})}),"\n",(0,s.jsx)(n.h2,{id:"scaledown--gracefully-remove-node",children:"Scaledown : Gracefully remove node"}),"\n",(0,s.jsx)(n.p,{children:"For this task, we will simply remove a node and look at that the decommissions steps."}),"\n",(0,s.jsxs)(n.ol,{children:["\n",(0,s.jsxs)(n.li,{children:["Remove the node from the list of ",(0,s.jsx)(n.code,{children:"NifiCluster.Spec.Nodes"})," field :"]}),"\n"]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-yaml",children:'apiVersion: nifi.konpyutaika.com/v1\nkind: NifiCluster\nmetadata:\n  name: simplenifi\nspec:\n  headlessServiceEnabled: true\n  zkAddress: "zookeepercluster-client.zookeeper:2181"\n  zkPath: "/simplenifi"\n  clusterImage: "apache/nifi:1.11.3"\n  oneNifiNodePerNode: false\n  nodeConfigGroups:\n    default_group:\n      isNode: true\n      storageConfigs:\n        - mountPath: "/opt/nifi/nifi-current/logs"\n          name: logs\n          metadata:\n            labels:\n              my-label: my-value\n            annotations:\n              my-annotation: my-value\n          pvcSpec:\n            accessModes:\n              - ReadWriteOnce\n            storageClassName: "standard"\n            resources:\n              requests:\n                storage: 10Gi\n      serviceAccountName: "default"\n      resourcesRequirements:\n        limits:\n          cpu: "2"\n          memory: 3Gi\n        requests:\n          cpu: "1"\n          memory: 1Gi\n  nodes:\n    - id: 0\n      nodeConfigGroup: "default_group"\n    - id: 1\n      nodeConfigGroup: "default_group"\n# >>>> START: node removed\n#    - id: 2\n#      nodeConfigGroup: "default_group"\n# <<<< END\n    - id: 25\n      nodeConfigGroup: "default_group"\n  propagateLabels: true\n  nifiClusterTaskSpec:\n    retryDurationMinutes: 10\n  listenersConfig:\n    internalListeners:\n      - type: "http"\n        name: "http"\n        containerPort: 8080\n      - type: "cluster"\n        name: "cluster"\n        containerPort: 6007\n      - type: "s2s"\n        name: "s2s"\n        containerPort: 10000\n'})}),"\n",(0,s.jsxs)(n.ol,{start:"2",children:["\n",(0,s.jsxs)(n.li,{children:["Apply the new ",(0,s.jsx)(n.code,{children:"NifiCluster"})," configuration :"]}),"\n"]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-sh",children:"kubectl -n nifi apply -f config/samples/simplenificluster.yaml\n"})}),"\n",(0,s.jsxs)(n.ol,{start:"3",children:["\n",(0,s.jsxs)(n.li,{children:["You can follow the node's action step status in the ",(0,s.jsx)(n.code,{children:"NifiCluster.Status"})," description :"]}),"\n"]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-console",children:"kubectl describe nificluster simplenifi\n\n...\nStatus:\n  Nodes State:\n    ...\n    2:\n      Configuration State:  ConfigInSync\n      Graceful Action State:\n        Action State:   GracefulDownscaleRequired\n        Error Message:\n    ...\n...\n"})}),"\n",(0,s.jsx)(n.admonition,{type:"tip",children:(0,s.jsxs)(n.p,{children:["The list of decommisions step and their corresponding value for the ",(0,s.jsx)(n.code,{children:"Nifi Cluster.Status.Node State.Graceful ActionState.ActionStep"})," field is described into the ",(0,s.jsx)(n.a,{href:"../../../5_references/1_nifi_cluster/5_node_state#actionstep",children:"Node State page"})]})}),"\n",(0,s.jsxs)(n.ol,{start:"4",children:["\n",(0,s.jsxs)(n.li,{children:["Once the scaledown successfully performed, you should have the data offloaded on the other nodes, and the node state removed from the ",(0,s.jsx)(n.code,{children:"NifiCluster.Status.NodesState"})," list :"]}),"\n"]}),"\n",(0,s.jsx)(n.admonition,{type:"warning",children:(0,s.jsxs)(n.p,{children:["Keep in mind that the ",(0,s.jsx)(n.a,{href:"/nifikop/docs/v1.4.0/5_references/1_nifi_cluster/#nificlustertaskspec",children:(0,s.jsx)(n.code,{children:"NifiCluster.Spec.nifiClusterTaskSpec.retryDurationMinutes"})})," should be long enough to perform the whole procedure, or you will have some rollback and retry loop."]})})]})}function u(e={}){const{wrapper:n}={...(0,t.a)(),...e.components};return n?(0,s.jsx)(n,{...e,children:(0,s.jsx)(d,{...e})}):d(e)}},17530:(e,n,i)=>{i.d(n,{Z:()=>s});const s=i.p+"assets/images/scaleup_cluster_list-35ad91fb8c072c4235a969eb9acfcdae.png"},70486:(e,n,i)=>{i.d(n,{Z:()=>s});const s=i.p+"assets/images/scaleup_distribution-e8a1d9e0e4ce70f4fe16965ebeee7a32.png"},90192:(e,n,i)=>{i.d(n,{Z:()=>s});const s=i.p+"assets/images/scaling_dataflow-5966160565dedb1d2c691ae255bae15c.png"},71670:(e,n,i)=>{i.d(n,{Z:()=>l,a:()=>a});var s=i(27378);const t={},o=s.createContext(t);function a(e){const n=s.useContext(o);return s.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function l(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(t):e.components||t:a(e.components),s.createElement(o.Provider,{value:n},e.children)}}}]);