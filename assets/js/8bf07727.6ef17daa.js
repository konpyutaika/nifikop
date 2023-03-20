"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[3958],{35318:(e,t,n)=>{n.d(t,{Zo:()=>s,kt:()=>m});var a=n(27378);function r(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function l(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,a)}return n}function i(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?l(Object(n),!0).forEach((function(t){r(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):l(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function o(e,t){if(null==e)return{};var n,a,r=function(e,t){if(null==e)return{};var n,a,r={},l=Object.keys(e);for(a=0;a<l.length;a++)n=l[a],t.indexOf(n)>=0||(r[n]=e[n]);return r}(e,t);if(Object.getOwnPropertySymbols){var l=Object.getOwnPropertySymbols(e);for(a=0;a<l.length;a++)n=l[a],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(r[n]=e[n])}return r}var d=a.createContext({}),u=function(e){var t=a.useContext(d),n=t;return e&&(n="function"==typeof e?e(t):i(i({},t),e)),n},s=function(e){var t=u(e.components);return a.createElement(d.Provider,{value:t},e.children)},p={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},c=a.forwardRef((function(e,t){var n=e.components,r=e.mdxType,l=e.originalType,d=e.parentName,s=o(e,["components","mdxType","originalType","parentName"]),c=u(n),m=r,g=c["".concat(d,".").concat(m)]||c[m]||p[m]||l;return n?a.createElement(g,i(i({ref:t},s),{},{components:n})):a.createElement(g,i({ref:t},s))}));function m(e,t){var n=arguments,r=t&&t.mdxType;if("string"==typeof e||r){var l=n.length,i=new Array(l);i[0]=c;var o={};for(var d in t)hasOwnProperty.call(t,d)&&(o[d]=t[d]);o.originalType=e,o.mdxType="string"==typeof e?e:r,i[1]=o;for(var u=2;u<l;u++)i[u]=n[u];return a.createElement.apply(null,i)}return a.createElement.apply(null,n)}c.displayName="MDXCreateElement"},29064:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>s,contentTitle:()=>d,default:()=>m,frontMatter:()=>o,metadata:()=>u,toc:()=>p});var a=n(25773),r=n(30808),l=(n(27378),n(35318)),i=["components"],o={id:"7_nifi_nodegroup_autoscaler",title:"NiFi NodeGroup Autoscaler",sidebar_label:"NiFi NodeGroup Autoscaler"},d=void 0,u={unversionedId:"5_references/7_nifi_nodegroup_autoscaler",id:"version-v1.0.0/5_references/7_nifi_nodegroup_autoscaler",title:"NiFi NodeGroup Autoscaler",description:"NifiNodeGroupAutoscaler is the Schema through which you configure automatic scaling of NifiCluster deployments.",source:"@site/versioned_docs/version-v1.0.0/5_references/7_nifi_nodegroup_autoscaler.md",sourceDirName:"5_references",slug:"/5_references/7_nifi_nodegroup_autoscaler",permalink:"/nifikop/docs/v1.0.0/5_references/7_nifi_nodegroup_autoscaler",draft:!1,editUrl:"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.0.0/5_references/7_nifi_nodegroup_autoscaler.md",tags:[],version:"v1.0.0",lastUpdatedBy:"Alexandre Guitton",lastUpdatedAt:1668875267,formattedLastUpdatedAt:"Nov 19, 2022",frontMatter:{id:"7_nifi_nodegroup_autoscaler",title:"NiFi NodeGroup Autoscaler",sidebar_label:"NiFi NodeGroup Autoscaler"},sidebar:"docs",previous:{title:"NiFi UserGroup",permalink:"/nifikop/docs/v1.0.0/5_references/6_nifi_usergroup"},next:{title:"Contribution organization",permalink:"/nifikop/docs/v1.0.0/6_contributing/0_contribution_organization"}},s={},p=[{value:"NifiNodeGroupAutoscaler",id:"nifinodegroupautoscaler",level:2},{value:"NifiNodeGroupAutoscalerSpec",id:"nifinodegroupautoscalerspec",level:2},{value:"NifiNodeGroupAutoscalerStatus",id:"nifinodegroupautoscalerstatus",level:2}],c={toc:p};function m(e){var t=e.components,n=(0,r.Z)(e,i);return(0,l.kt)("wrapper",(0,a.Z)({},c,n,{components:t,mdxType:"MDXLayout"}),(0,l.kt)("p",null,(0,l.kt)("inlineCode",{parentName:"p"},"NifiNodeGroupAutoscaler")," is the Schema through which you configure automatic scaling of ",(0,l.kt)("inlineCode",{parentName:"p"},"NifiCluster")," deployments."),(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-yaml"},"apiVersion: nifi.konpyutaika.com/v1alpha1\nkind: NifiNodeGroupAutoscaler\nmetadata:\n  name: nifinodegroupautoscaler-sample\nspec:\n  # contains the reference to the NifiCluster with the one the node group autoscaler is linked.\n  clusterRef:\n    name: nificluster-name\n    namespace: nifikop\n  # defines the id of the NodeConfig contained in NifiCluster.Spec.NodeConfigGroups\n  nodeConfigGroupId: default-node-group\n  # The selector used to identify nodes in NifiCluster.Spec.Nodes this autoscaler will manage\n  # Use Node.Labels in combination with this selector to clearly define which nodes will be managed by this autoscaler \n  nodeLabelsSelector: \n    matchLabels:\n      nifi_cr: nificluster-name\n      nifi_node_group: default-node-group\n  # the strategy used to decide how to add nodes to a nifi cluster\n  upscaleStrategy: simple\n  # the strategy used to decide how to remove nodes from an existing cluster\n  downscaleStrategy: lifo\n")),(0,l.kt)("h2",{id:"nifinodegroupautoscaler"},"NifiNodeGroupAutoscaler"),(0,l.kt)("table",null,(0,l.kt)("thead",{parentName:"table"},(0,l.kt)("tr",{parentName:"thead"},(0,l.kt)("th",{parentName:"tr",align:null},"Field"),(0,l.kt)("th",{parentName:"tr",align:null},"Type"),(0,l.kt)("th",{parentName:"tr",align:null},"Description"),(0,l.kt)("th",{parentName:"tr",align:null},"Required"),(0,l.kt)("th",{parentName:"tr",align:null},"Default"))),(0,l.kt)("tbody",{parentName:"table"},(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"metadata"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("a",{parentName:"td",href:"https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta"},"ObjectMetadata")),(0,l.kt)("td",{parentName:"tr",align:null},"is metadata that all persisted resources must have, which includes all objects nodegroupautoscalers must create."),(0,l.kt)("td",{parentName:"tr",align:null},"No"),(0,l.kt)("td",{parentName:"tr",align:null},"nil")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"spec"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("a",{parentName:"td",href:"#nifinodegroupautoscalerspec"},"NifiNodeGroupAutoscalerSpec")),(0,l.kt)("td",{parentName:"tr",align:null},"defines the desired state of NifiNodeGroupAutoscaler."),(0,l.kt)("td",{parentName:"tr",align:null},"No"),(0,l.kt)("td",{parentName:"tr",align:null},"nil")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"status"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("a",{parentName:"td",href:"#nifinodegroupautoscalerstatus"},"NifiNodeGroupAutoscalerStatus")),(0,l.kt)("td",{parentName:"tr",align:null},"defines the observed state of NifiNodeGroupAutoscaler."),(0,l.kt)("td",{parentName:"tr",align:null},"No"),(0,l.kt)("td",{parentName:"tr",align:null},"nil")))),(0,l.kt)("h2",{id:"nifinodegroupautoscalerspec"},"NifiNodeGroupAutoscalerSpec"),(0,l.kt)("table",null,(0,l.kt)("thead",{parentName:"table"},(0,l.kt)("tr",{parentName:"thead"},(0,l.kt)("th",{parentName:"tr",align:null},"Field"),(0,l.kt)("th",{parentName:"tr",align:null},"Type"),(0,l.kt)("th",{parentName:"tr",align:null},"Description"),(0,l.kt)("th",{parentName:"tr",align:null},"Required"),(0,l.kt)("th",{parentName:"tr",align:null},"Default"))),(0,l.kt)("tbody",{parentName:"table"},(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"clusterRef"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("a",{parentName:"td",href:"./2_nifi_user#clusterreference"},"ClusterReference")),(0,l.kt)("td",{parentName:"tr",align:null},"contains the reference to the NifiCluster containing the node group this autoscaler should manage."),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"-")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"nodeConfigGroupId"),(0,l.kt)("td",{parentName:"tr",align:null},"string"),(0,l.kt)("td",{parentName:"tr",align:null},"defines the id of the ",(0,l.kt)("a",{parentName:"td",href:"./1_nifi_cluster/3_node_config"},"NodeConfig")," contained in ",(0,l.kt)("inlineCode",{parentName:"td"},"NifiCluster.Spec.NodeConfigGroups"),"."),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"-")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"nodeLabelsSelector"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("a",{parentName:"td",href:"https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#LabelSelector"},"LabelSelector")),(0,l.kt)("td",{parentName:"tr",align:null},"defines the set of labels used to identify nodes in a ",(0,l.kt)("inlineCode",{parentName:"td"},"NifiCluster")," node config group. Use ",(0,l.kt)("inlineCode",{parentName:"td"},"Node.Labels")," in combination with this selector to clearly define which nodes will be managed by this autoscaler. Take care to avoid having mutliple autoscalers managing the same nodes."),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"-")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"readOnlyConfig"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("a",{parentName:"td",href:"./1_nifi_cluster/2_read_only_config"},"ReadOnlyConfig")),(0,l.kt)("td",{parentName:"tr",align:null},"defines a readOnlyConfig to apply to each node in this node group. Any settings here will override those set in the configured ",(0,l.kt)("inlineCode",{parentName:"td"},"nodeConfigGroupId"),"."),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"-")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"nodeConfig"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("a",{parentName:"td",href:"./1_nifi_cluster/3_node_config"},"NodeConfig")),(0,l.kt)("td",{parentName:"tr",align:null},"defines a nodeConfig to apply to each node in this node group. Any settings here will override those set in the configured ",(0,l.kt)("inlineCode",{parentName:"td"},"nodeConfigGroupId"),"."),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"-")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"upscaleStrategy"),(0,l.kt)("td",{parentName:"tr",align:null},"string"),(0,l.kt)("td",{parentName:"tr",align:null},"The strategy NiFiKop will use to scale up the nodes managed by this autoscaler. Must be one of {",(0,l.kt)("inlineCode",{parentName:"td"},"simple"),"}."),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"-")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"downscaleStrategy"),(0,l.kt)("td",{parentName:"tr",align:null},"string"),(0,l.kt)("td",{parentName:"tr",align:null},"The strategy NiFiKop will use to scale down the nodes managed by this autoscaler. Must be one of {",(0,l.kt)("inlineCode",{parentName:"td"},"lifo"),"}."),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"-")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"replicas"),(0,l.kt)("td",{parentName:"tr",align:null},"int"),(0,l.kt)("td",{parentName:"tr",align:null},"the initial number of replicas to configure the ",(0,l.kt)("inlineCode",{parentName:"td"},"HorizontalPodAutoscaler")," with. After the initial configuration, this ",(0,l.kt)("inlineCode",{parentName:"td"},"replicas")," configuration will be automatically updated by the Kubernetes ",(0,l.kt)("inlineCode",{parentName:"td"},"HorizontalPodAutoscaler")," controller."),(0,l.kt)("td",{parentName:"tr",align:null},"No"),(0,l.kt)("td",{parentName:"tr",align:null},"1")))),(0,l.kt)("h2",{id:"nifinodegroupautoscalerstatus"},"NifiNodeGroupAutoscalerStatus"),(0,l.kt)("table",null,(0,l.kt)("thead",{parentName:"table"},(0,l.kt)("tr",{parentName:"thead"},(0,l.kt)("th",{parentName:"tr",align:null},"Field"),(0,l.kt)("th",{parentName:"tr",align:null},"Type"),(0,l.kt)("th",{parentName:"tr",align:null},"Description"),(0,l.kt)("th",{parentName:"tr",align:null},"Required"),(0,l.kt)("th",{parentName:"tr",align:null},"Default"))),(0,l.kt)("tbody",{parentName:"table"},(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"state"),(0,l.kt)("td",{parentName:"tr",align:null},"string"),(0,l.kt)("td",{parentName:"tr",align:null},"the state of the nodegroup autoscaler. This is set by the autoscaler."),(0,l.kt)("td",{parentName:"tr",align:null},"No"),(0,l.kt)("td",{parentName:"tr",align:null},"-")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"replicas"),(0,l.kt)("td",{parentName:"tr",align:null},"int"),(0,l.kt)("td",{parentName:"tr",align:null},"the current number of replicas running in the node group this autoscaler is managing. This is set by the autoscaler."),(0,l.kt)("td",{parentName:"tr",align:null},"No"),(0,l.kt)("td",{parentName:"tr",align:null},"-")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"selector"),(0,l.kt)("td",{parentName:"tr",align:null},"string"),(0,l.kt)("td",{parentName:"tr",align:null},"the ",(0,l.kt)("a",{parentName:"td",href:"https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/"},"selector")," used by the ",(0,l.kt)("inlineCode",{parentName:"td"},"HorizontalPodAutoscaler")," controller to identify the replicas in this node group. This is set by the autoscaler."),(0,l.kt)("td",{parentName:"tr",align:null},"No"),(0,l.kt)("td",{parentName:"tr",align:null},"-")))))}m.isMDXComponent=!0}}]);