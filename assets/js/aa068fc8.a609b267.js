"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[10364],{43023:(e,t,n)=>{n.d(t,{R:()=>s,x:()=>a});var o=n(63696);const i={},r=o.createContext(i);function s(e){const t=o.useContext(r);return o.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function a(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(i):e.components||i:s(e.components),o.createElement(r.Provider,{value:t},e.children)}},94405:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>c,contentTitle:()=>a,default:()=>h,frontMatter:()=>s,metadata:()=>o,toc:()=>l});const o=JSON.parse('{"id":"1_concepts/2_design_principles","title":"Design Principles","description":"This operator is built following the logic implied by the [operator sdk framework] (https://sdk.operatorframework.io/).","source":"@site/versioned_docs/version-v1.1.0/1_concepts/2_design_principles.md","sourceDirName":"1_concepts","slug":"/1_concepts/2_design_principles","permalink":"/nifikop/docs/v1.1.0/1_concepts/2_design_principles","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.1.0/1_concepts/2_design_principles.md","tags":[],"version":"v1.1.0","lastUpdatedBy":"Juldrixx","lastUpdatedAt":1679343499000,"frontMatter":{"id":"2_design_principles","title":"Design Principles","sidebar_label":"Design Principles"},"sidebar":"docs","previous":{"title":"Start here","permalink":"/nifikop/docs/v1.1.0/1_concepts/1_start_here"},"next":{"title":"Features","permalink":"/nifikop/docs/v1.1.0/1_concepts/3_features"}}');var i=n(62540),r=n(43023);const s={id:"2_design_principles",title:"Design Principles",sidebar_label:"Design Principles"},a=void 0,c={},l=[{value:"Separation of concerns",id:"separation-of-concerns",level:2},{value:"One controller per resource",id:"one-controller-per-resource",level:2}];function d(e){const t={a:"a",code:"code",h2:"h2",li:"li",p:"p",ul:"ul",...(0,r.R)(),...e.components};return(0,i.jsxs)(i.Fragment,{children:[(0,i.jsxs)(t.p,{children:["This operator is built following the logic implied by the [operator sdk framework] (",(0,i.jsx)(t.a,{href:"https://sdk.operatorframework.io/",children:"https://sdk.operatorframework.io/"}),").\nWhat we want to offer with NiFiKop is that we provide as much automation as possible to manage NiFi at scale and in the most automated way possible."]}),"\n",(0,i.jsx)(t.h2,{id:"separation-of-concerns",children:"Separation of concerns"}),"\n",(0,i.jsx)(t.p,{children:"Kubernetes is designed for automation. Right out of the box, the Kubernetes core has a lot of automation features built in. You can use Kubernetes to automate the deployment and execution of workloads, and you can automate how Kubernetes does it."}),"\n",(0,i.jsx)(t.p,{children:"The Kubernetes operator model concept allows you to extend cluster behavior without changing the Kubernetes code itself by binding controllers to one or more custom resources. Operators are clients of the Kubernetes API that act as controllers for a custom resource."}),"\n",(0,i.jsx)(t.p,{children:"There are two things we can think of when we talk about operators in Kubernetes:"}),"\n",(0,i.jsxs)(t.ul,{children:["\n",(0,i.jsx)(t.li,{children:"Automate the deployment of my entire stack."}),"\n",(0,i.jsx)(t.li,{children:"Automate the actions required by the deployment."}),"\n"]}),"\n",(0,i.jsx)(t.p,{children:"For NiFiKop, we focus primarily on NiFi for the stack concept, what does that mean?"}),"\n",(0,i.jsxs)(t.ul,{children:["\n",(0,i.jsx)(t.li,{children:"We do not manage other components that can be integrated with NiFi Cluster like Prometheus, Zookeeper, NiFi registry etc."}),"\n",(0,i.jsx)(t.li,{children:"We want to provide as many tools as possible to automate the work on NiFi (cluster deployment, data flow and user management, etc.)."}),"\n"]}),"\n",(0,i.jsx)(t.p,{children:"We consider that for NiFiKop to be a low-level operator, focused on NiFi and only NiFi, and if we want to manage a complex stack with e.g. Zookeeper, NiFi Registry, Prometheus etc. we need something else working at a higher level, like Helm charts or another operator controlling NiFiKop and other resources."}),"\n",(0,i.jsx)(t.h2,{id:"one-controller-per-resource",children:"One controller per resource"}),"\n",(0,i.jsx)(t.p,{children:"The operator should reflect as much as possible the behavior of the solution we want to automate. If we take our case with NiFi, what we can say is that:"}),"\n",(0,i.jsxs)(t.ul,{children:["\n",(0,i.jsx)(t.li,{children:"You can have one or more NiFi clusters"}),"\n",(0,i.jsx)(t.li,{children:"On your cluster you can define a NiFi registry client, but it is not mandatory."}),"\n",(0,i.jsx)(t.li,{children:"You can also define users and groups and deploy a DataFlow if you want."}),"\n"]}),"\n",(0,i.jsx)(t.p,{children:"This means that your cluster is not defined by what is deployed on it, and what you deploy on it does not depend on it.\nTo be more explicit, just because I deploy a NiFi cluster doesn't mean the DataFlow deployed on it will stay there, we can move the DataFlow from one cluster to another."}),"\n",(0,i.jsxs)(t.p,{children:["To manage this, we need to create different kinds of resources (",(0,i.jsx)(t.a,{href:"../5_references/1_nifi_cluster",children:"NifiCluster"}),", ",(0,i.jsx)(t.a,{href:"../5_references/5_nifi_dataflow",children:"NifiDataflow"}),", ",(0,i.jsx)(t.a,{href:"../5_references/4_nifi_parameter_context",children:"NifiParameterContext"}),", ",(0,i.jsx)(t.a,{href:"../5_references/2_nifi_user",children:"NifiUser"}),", ",(0,i.jsx)(t.a,{href:"../5_references/6_nifi_usergroup",children:"NifiUserGroup"}),", ",(0,i.jsx)(t.a,{href:"../5_references/3_nifi_registry_client",children:"NifiRegistryClient"}),", ",(0,i.jsx)(t.a,{href:"../5_references/7_nifi_nodegroup_autoscaler",children:"NifiNodeGroupAutoscaler"}),") with one controller per resource that will manage its own resource.\nIn this way, we can say:"]}),"\n",(0,i.jsxs)(t.ul,{children:["\n",(0,i.jsx)(t.li,{children:"I deploy a NiFiCluster"}),"\n",(0,i.jsxs)(t.li,{children:["I define a NiFiDataflow that will be deployed on this cluster, and if I want to change cluster, I just have to change the ",(0,i.jsx)(t.code,{children:"ClusterRef"})," to change cluster"]}),"\n"]})]})}function h(e={}){const{wrapper:t}={...(0,r.R)(),...e.components};return t?(0,i.jsx)(t,{...e,children:(0,i.jsx)(d,{...e})}):d(e)}}}]);