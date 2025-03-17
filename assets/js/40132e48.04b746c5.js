"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[93752],{12671:(e,n,i)=>{i.d(n,{A:()=>t});const t=i.p+"assets/images/dataflow_lifecycle_management_schema-e39196d2242598106229e66f18e8826d.jpg"},21668:(e,n,i)=>{i.d(n,{A:()=>t});const t=i.p+"assets/images/parameter_context_reconcile_loop-0b9f053e9cb447162535e03e5f5e9fbd.jpeg"},31992:(e,n,i)=>{i.d(n,{A:()=>t});const t=i.p+"assets/images/registry_client_reconcile_loop-0b8e528b249cd93e61f4186050c59c02.jpeg"},33043:(e,n,i)=>{i.d(n,{A:()=>t});const t=i.p+"assets/images/dataflow_reconcile_loop-7b564f9232c78a2c651094a8005ba6a8.jpeg"},39933:(e,n,i)=>{i.r(n),i.d(n,{assets:()=>l,contentTitle:()=>o,default:()=>p,frontMatter:()=>r,metadata:()=>t,toc:()=>c});const t=JSON.parse('{"id":"3_manage_nifi/3_manage_dataflows/0_design_principles","title":"Design Principles","description":"The Dataflow Lifecycle management feature introduces 3 new CRDs:","source":"@site/versioned_docs/version-v1.11.4/3_manage_nifi/3_manage_dataflows/0_design_principles.md","sourceDirName":"3_manage_nifi/3_manage_dataflows","slug":"/3_manage_nifi/3_manage_dataflows/0_design_principles","permalink":"/nifikop/docs/v1.11.4/3_manage_nifi/3_manage_dataflows/0_design_principles","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.11.4/3_manage_nifi/3_manage_dataflows/0_design_principles.md","tags":[],"version":"v1.11.4","lastUpdatedBy":"Juldrixx","lastUpdatedAt":1731668582000,"frontMatter":{"id":"0_design_principles","title":"Design Principles","sidebar_label":"Design Principles"},"sidebar":"docs","previous":{"title":"Managed groups","permalink":"/nifikop/docs/v1.11.4/3_manage_nifi/2_manage_users_and_accesses/3_managed_groups"},"next":{"title":"Deploy dataflow","permalink":"/nifikop/docs/v1.11.4/3_manage_nifi/3_manage_dataflows/1_deploy_dataflow"}}');var s=i(62540),a=i(43023);const r={id:"0_design_principles",title:"Design Principles",sidebar_label:"Design Principles"},o=void 0,l={},c=[];function d(e){const n={a:"a",code:"code",img:"img",li:"li",p:"p",strong:"strong",ul:"ul",...(0,a.R)(),...e.components};return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsxs)(n.p,{children:["The ",(0,s.jsx)(n.a,{href:"../../1_concepts/3_features#dataflow-lifecycle-management-via-crd",children:"Dataflow Lifecycle management feature"})," introduces 3 new CRDs:"]}),"\n",(0,s.jsxs)(n.ul,{children:["\n",(0,s.jsxs)(n.li,{children:[(0,s.jsx)(n.strong,{children:"NiFiRegistryClient:"})," Allowing you to declare a ",(0,s.jsx)(n.a,{href:"https://nifi.apache.org/docs/nifi-registry-docs/html/getting-started.html#connect-nifi-to-the-registry",children:"NiFi registry client"}),"."]}),"\n",(0,s.jsxs)(n.li,{children:[(0,s.jsx)(n.strong,{children:"NiFiParameterContext:"})," Allowing you to create parameter context, with two kinds of parameters, a simple ",(0,s.jsx)(n.code,{children:"map[string]string"})," for non-sensitive parameters and a ",(0,s.jsx)(n.code,{children:"list of secrets"})," which contains sensitive parameters."]}),"\n",(0,s.jsxs)(n.li,{children:[(0,s.jsx)(n.strong,{children:"NiFiDataflow:"})," Allowing you to declare a Dataflow based on a ",(0,s.jsx)(n.code,{children:"NiFiRegistryClient"})," and optionally a ",(0,s.jsx)(n.code,{children:"ParameterContext"}),", which will be deployed and managed by the operator on the ",(0,s.jsx)(n.code,{children:"targeted NiFi cluster"}),"."]}),"\n"]}),"\n",(0,s.jsx)(n.p,{children:"The following diagram shows the interactions between all the components:"}),"\n",(0,s.jsx)(n.p,{children:(0,s.jsx)(n.img,{alt:"dataflow lifecycle management schema",src:i(12671).A+"",width:"1123",height:"1101"})}),"\n",(0,s.jsx)(n.p,{children:"With each CRD comes a new controller, with a reconcile loop:"}),"\n",(0,s.jsxs)(n.ul,{children:["\n",(0,s.jsx)(n.li,{children:(0,s.jsx)(n.strong,{children:"NiFiRegistryClient's controller:"})}),"\n"]}),"\n",(0,s.jsx)(n.p,{children:(0,s.jsx)(n.img,{alt:"NiFi registry client&#39;s reconcile loop",src:i(31992).A+"",width:"682",height:"642"})}),"\n",(0,s.jsxs)(n.ul,{children:["\n",(0,s.jsx)(n.li,{children:(0,s.jsx)(n.strong,{children:"NiFiParameterContext's controller:"})}),"\n"]}),"\n",(0,s.jsx)(n.p,{children:(0,s.jsx)(n.img,{alt:"NiFi parameter context&#39;s reconcile loop",src:i(21668).A+"",width:"922",height:"829"})}),"\n",(0,s.jsxs)(n.ul,{children:["\n",(0,s.jsx)(n.li,{children:(0,s.jsx)(n.strong,{children:"NiFiDataflow's controller:"})}),"\n"]}),"\n",(0,s.jsx)(n.p,{children:(0,s.jsx)(n.img,{alt:"NiFi dataflow&#39;s reconcile loop",src:i(33043).A+"",width:"3146",height:"4117"})})]})}function p(e={}){const{wrapper:n}={...(0,a.R)(),...e.components};return n?(0,s.jsx)(n,{...e,children:(0,s.jsx)(d,{...e})}):d(e)}},43023:(e,n,i)=>{i.d(n,{R:()=>r,x:()=>o});var t=i(63696);const s={},a=t.createContext(s);function r(e){const n=t.useContext(a);return t.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function o(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(s):e.components||s:r(e.components),t.createElement(a.Provider,{value:n},e.children)}}}]);