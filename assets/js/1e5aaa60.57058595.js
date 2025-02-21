"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[6695],{46177:(e,n,i)=>{i.r(n),i.d(n,{assets:()=>l,contentTitle:()=>c,default:()=>h,frontMatter:()=>a,metadata:()=>s,toc:()=>r});const s=JSON.parse('{"id":"3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/0_design_principles","title":"Design Principles","description":"These feature have been scoped by the community, please find the discussion and technical scoping here.","source":"@site/versioned_docs/version-v1.11.0/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/0_design_principles.md","sourceDirName":"3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling","slug":"/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/0_design_principles","permalink":"/nifikop/docs/v1.11.0/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/0_design_principles","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.11.0/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/0_design_principles.md","tags":[],"version":"v1.11.0","lastUpdatedBy":"Julien Guitton","lastUpdatedAt":1740132642000,"frontMatter":{"id":"0_design_principles","title":"Design Principles","sidebar_label":"Design Principles"},"sidebar":"docs","previous":{"title":"Scaling mechanism","permalink":"/nifikop/docs/v1.11.0/3_manage_nifi/1_manage_clusters/2_cluster_scaling/1_scaling_mechanism"},"next":{"title":"Using KEDA","permalink":"/nifikop/docs/v1.11.0/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/1_using_keda"}}');var t=i(62540),o=i(43023);const a={id:"0_design_principles",title:"Design Principles",sidebar_label:"Design Principles"},c=void 0,l={},r=[{value:"Design reflexion",id:"design-reflexion",level:2},{value:"Implementation",id:"implementation",level:2}];function d(e){const n={a:"a",admonition:"admonition",code:"code",h2:"h2",img:"img",li:"li",p:"p",ul:"ul",...(0,o.R)(),...e.components};return(0,t.jsxs)(t.Fragment,{children:[(0,t.jsx)(n.admonition,{type:"info",children:(0,t.jsxs)(n.p,{children:["These feature have been scoped by the community, please find the discussion and technical scoping ",(0,t.jsx)(n.a,{href:"https://docs.google.com/document/d/1QNGSNNjWx4CGt5-NvX9ZArQMfyrwjw-B95f54GUNdB0/edit#heading=h.t9xh94v7viuj",children:"here"}),"."]})}),"\n",(0,t.jsx)(n.h2,{id:"design-reflexion",children:"Design reflexion"}),"\n",(0,t.jsxs)(n.p,{children:["If you read the technical scoping above, we explored many options for enabling automatic scaling of NiFi clusters.\nAfter much discussion, it turned out that we wanted to mimic the approach and design behind auto-scaling a deployment with ",(0,t.jsx)(n.a,{href:"https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",children:"HPA"}),"."]}),"\n",(0,t.jsxs)(n.p,{children:["If we look at how this works, you define a ",(0,t.jsx)(n.code,{children:"Deployment"}),", which will manage a ",(0,t.jsx)(n.code,{children:"ReplicaSet"})," which will manage ",(0,t.jsx)(n.code,{children:"Pods"}),". And you define your ",(0,t.jsx)(n.code,{children:"HPA"})," which will manage the scale field of the ",(0,t.jsx)(n.code,{children:"Deployment"}),".\nFor our ",(0,t.jsx)(n.code,{children:"NiFiCluster"})," we considered the same kind of separation of concerns: we define a new resource ",(0,t.jsx)(n.code,{children:"NifiNodeGroupAutoScaler"})," that manages the ",(0,t.jsx)(n.code,{children:"NifiCluster"})," that will manage the ",(0,t.jsx)(n.code,{children:"Pods"}),". And you define your ",(0,t.jsx)(n.code,{children:"HPA"})," which will manage the scale field of the ",(0,t.jsx)(n.code,{children:"Deployment"}),"."]}),"\n",(0,t.jsx)(n.p,{children:"This is the basis of the thinking. There was another inspiration for designing the functionality, which is that we wanted to keep the possibility of different types of node groups and manage them separately, so we pushed by thinking about similar existing models, and we thought about how in the Kubernetes Cloud Cluster (EKS, GKE etc.) nodes can be managed.\nYou can define fixed groups of nodes, you can auto-scale others."}),"\n",(0,t.jsxs)(n.p,{children:["And finally, we wanted to separate the ",(0,t.jsx)(n.code,{children:"NifiCluster"})," itself from the ",(0,t.jsx)(n.code,{children:"autoscaling management"})," and allow mixing the two, allowing you to have a cluster initially with no scaling at all, add scaling from a subset of nodes with a given configuration, and finally disable autoscaling without any impact."]}),"\n",(0,t.jsx)(n.h2,{id:"implementation",children:"Implementation"}),"\n",(0,t.jsxs)(n.p,{children:["Referring to the official guideline, the recommended approach is to implement ",(0,t.jsx)(n.a,{href:"https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#scale-subresource",children:"the sub resource scale in the CRD"}),"."]}),"\n",(0,t.jsx)(n.p,{children:"This approach requires to define:"}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsxs)(n.li,{children:[(0,t.jsx)(n.code,{children:"specReplicasPath"})," defines the JSONPath inside of a custom resource that corresponds to ",(0,t.jsx)(n.code,{children:"scale.spec.replicas"})]}),"\n",(0,t.jsxs)(n.li,{children:[(0,t.jsx)(n.code,{children:"statusReplicasPath"})," defines the JSONPath inside of a custom resource that corresponds to ",(0,t.jsx)(n.code,{children:"scale.status.replicas"})]}),"\n",(0,t.jsxs)(n.li,{children:[(0,t.jsx)(n.code,{children:"labelSelectorPath"})," defines the JSONPath inside of a custom resource that corresponds to ",(0,t.jsx)(n.code,{children:"scale.Status.Selector"})]}),"\n"]}),"\n",(0,t.jsxs)(n.p,{children:["we add a new resource: ",(0,t.jsx)(n.a,{href:"../../../../5_references/7_nifi_nodegroup_autoscaler",children:"NifiNodeGroupAutoScaler"}),", with the following fields:"]}),"\n",(0,t.jsxs)(n.ul,{children:["\n",(0,t.jsxs)(n.li,{children:[(0,t.jsx)(n.code,{children:"spec.nifiClusterRef"}),": reference to the NiFi cluster resource that will be autoscaled"]}),"\n",(0,t.jsxs)(n.li,{children:[(0,t.jsx)(n.code,{children:"spec.nodeConfigGroupId"}),": reference to the nodeConfigGroup that will be used for nodes managed by the auto scaling."]}),"\n",(0,t.jsxs)(n.li,{children:[(0,t.jsx)(n.code,{children:"spec.readOnlyConfig"}),": defines a readOnlyConfig to apply to each node in this node group. Any settings here will override those set in the configured ",(0,t.jsx)(n.code,{children:"NifiCluster"}),"."]}),"\n",(0,t.jsxs)(n.li,{children:[(0,t.jsx)(n.code,{children:"spec.nodeConfig"}),": defines a nodeConfig to apply to each node in this node group. Any settings here will override those set in the configured ",(0,t.jsx)(n.code,{children:"nodeConfigGroupId"}),"."]}),"\n",(0,t.jsxs)(n.li,{children:[(0,t.jsx)(n.code,{children:"spec.replicas"}),": current number of replicas expected for the node config group"]}),"\n",(0,t.jsxs)(n.li,{children:[(0,t.jsx)(n.code,{children:"spec.upscaleStrategy"}),": strategy used to upscale (simple)"]}),"\n",(0,t.jsxs)(n.li,{children:[(0,t.jsx)(n.code,{children:"spec.downscaleStrategy"}),": strategy used to downscale (lifo)"]}),"\n"]}),"\n",(0,t.jsx)(n.p,{children:"Here is a representation  of dependencies:"}),"\n",(0,t.jsx)(n.p,{children:(0,t.jsx)(n.img,{alt:"auto scaling schema",src:i(9989).A+"",width:"1741",height:"921"})})]})}function h(e={}){const{wrapper:n}={...(0,o.R)(),...e.components};return n?(0,t.jsx)(n,{...e,children:(0,t.jsx)(d,{...e})}):d(e)}},9989:(e,n,i)=>{i.d(n,{A:()=>s});const s=i.p+"assets/images/auto_scaling-efb9955ed5598d76dc318ad7e8df9e2f.jpg"},43023:(e,n,i)=>{i.d(n,{R:()=>a,x:()=>c});var s=i(63696);const t={},o=s.createContext(t);function a(e){const n=s.useContext(o);return s.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function c(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(t):e.components||t:a(e.components),s.createElement(o.Provider,{value:n},e.children)}}}]);