"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[3011],{35318:(e,t,n)=>{n.d(t,{Zo:()=>p,kt:()=>f});var a=n(27378);function i(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function r(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,a)}return n}function o(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?r(Object(n),!0).forEach((function(t){i(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):r(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function l(e,t){if(null==e)return{};var n,a,i=function(e,t){if(null==e)return{};var n,a,i={},r=Object.keys(e);for(a=0;a<r.length;a++)n=r[a],t.indexOf(n)>=0||(i[n]=e[n]);return i}(e,t);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);for(a=0;a<r.length;a++)n=r[a],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(i[n]=e[n])}return i}var s=a.createContext({}),c=function(e){var t=a.useContext(s),n=t;return e&&(n="function"==typeof e?e(t):o(o({},t),e)),n},p=function(e){var t=c(e.components);return a.createElement(s.Provider,{value:t},e.children)},m={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},d=a.forwardRef((function(e,t){var n=e.components,i=e.mdxType,r=e.originalType,s=e.parentName,p=l(e,["components","mdxType","originalType","parentName"]),d=c(n),f=i,g=d["".concat(s,".").concat(f)]||d[f]||m[f]||r;return n?a.createElement(g,o(o({ref:t},p),{},{components:n})):a.createElement(g,o({ref:t},p))}));function f(e,t){var n=arguments,i=t&&t.mdxType;if("string"==typeof e||i){var r=n.length,o=new Array(r);o[0]=d;var l={};for(var s in t)hasOwnProperty.call(t,s)&&(l[s]=t[s]);l.originalType=e,l.mdxType="string"==typeof e?e:i,o[1]=l;for(var c=2;c<r;c++)o[c]=n[c];return a.createElement.apply(null,o)}return a.createElement.apply(null,n)}d.displayName="MDXCreateElement"},10512:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>p,contentTitle:()=>s,default:()=>f,frontMatter:()=>l,metadata:()=>c,toc:()=>m});var a=n(25773),i=n(30808),r=(n(27378),n(35318)),o=["components"],l={id:"0_design_principles",title:"Design Principles",sidebar_label:"Design Principles"},s=void 0,c={unversionedId:"3_manage_nifi/3_manage_dataflows/0_design_principles",id:"version-v1.0.0/3_manage_nifi/3_manage_dataflows/0_design_principles",title:"Design Principles",description:"The Dataflow Lifecycle management feature introduces 3 new CRDs :",source:"@site/versioned_docs/version-v1.0.0/3_manage_nifi/3_manage_dataflows/0_design_principles.md",sourceDirName:"3_manage_nifi/3_manage_dataflows",slug:"/3_manage_nifi/3_manage_dataflows/0_design_principles",permalink:"/nifikop/docs/v1.0.0/3_manage_nifi/3_manage_dataflows/0_design_principles",draft:!1,editUrl:"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.0.0/3_manage_nifi/3_manage_dataflows/0_design_principles.md",tags:[],version:"v1.0.0",lastUpdatedBy:"Alexandre Guitton",lastUpdatedAt:1668875267,formattedLastUpdatedAt:"Nov 19, 2022",frontMatter:{id:"0_design_principles",title:"Design Principles",sidebar_label:"Design Principles"},sidebar:"docs",previous:{title:"Managed groups",permalink:"/nifikop/docs/v1.0.0/3_manage_nifi/2_manage_users_and_accesses/3_managed_groups"},next:{title:"Deploy dataflow",permalink:"/nifikop/docs/v1.0.0/3_manage_nifi/3_manage_dataflows/1_deploy_dataflow"}},p={},m=[],d={toc:m};function f(e){var t=e.components,l=(0,i.Z)(e,o);return(0,r.kt)("wrapper",(0,a.Z)({},d,l,{components:t,mdxType:"MDXLayout"}),(0,r.kt)("p",null,"The ",(0,r.kt)("a",{parentName:"p",href:"../../1_concepts/3_features#dataflow-lifecycle-management-via-crd"},"Dataflow Lifecycle management feature")," introduces 3 new CRDs :"),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"NiFiRegistryClient :")," Allowing you to declare a ",(0,r.kt)("a",{parentName:"li",href:"https://nifi.apache.org/docs/nifi-registry-docs/html/getting-started.html#connect-nifi-to-the-registry"},"NiFi registry client"),"."),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"NiFiParameterContext :")," Allowing you to create parameter context, with two kinds of parameters, a simple ",(0,r.kt)("inlineCode",{parentName:"li"},"map[string]string")," for non-sensitive parameters and a ",(0,r.kt)("inlineCode",{parentName:"li"},"list of secrets")," which contains sensitive parameters."),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"NiFiDataflow :")," Allowing you to declare a Dataflow based on a ",(0,r.kt)("inlineCode",{parentName:"li"},"NiFiRegistryClient")," and optionally a ",(0,r.kt)("inlineCode",{parentName:"li"},"ParameterContext"),", which will be deployed and managed by the operator on the ",(0,r.kt)("inlineCode",{parentName:"li"},"targeted NiFi cluster"),".")),(0,r.kt)("p",null,"The following diagram shows the interactions between all the components :"),(0,r.kt)("p",null,(0,r.kt)("img",{alt:"dataflow lifecycle management schema",src:n(58643).Z,width:"1123",height:"1101"})),(0,r.kt)("p",null,"With each CRD comes a new controller, with a reconcile loop :"),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"NiFiRegistryClient's controller :"))),(0,r.kt)("p",null,(0,r.kt)("img",{alt:"NiFi registry client&#39;s reconcile loop",src:n(74219).Z,width:"682",height:"642"})),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"NiFiParameterContext's controller :"))),(0,r.kt)("p",null,(0,r.kt)("img",{alt:"NiFi parameter context&#39;s reconcile loop",src:n(34753).Z,width:"922",height:"829"})),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("strong",{parentName:"li"},"NiFiDataflow's controller :"))),(0,r.kt)("p",null,(0,r.kt)("img",{alt:"NiFi dataflow&#39;s reconcile loop",src:n(52614).Z,width:"3146",height:"4117"})))}f.isMDXComponent=!0},58643:(e,t,n)=>{n.d(t,{Z:()=>a});const a=n.p+"assets/images/dataflow_lifecycle_management_schema-e39196d2242598106229e66f18e8826d.jpg"},52614:(e,t,n)=>{n.d(t,{Z:()=>a});const a=n.p+"assets/images/dataflow_reconcile_loop-7b564f9232c78a2c651094a8005ba6a8.jpeg"},34753:(e,t,n)=>{n.d(t,{Z:()=>a});const a=n.p+"assets/images/parameter_context_reconcile_loop-0b9f053e9cb447162535e03e5f5e9fbd.jpeg"},74219:(e,t,n)=>{n.d(t,{Z:()=>a});const a=n.p+"assets/images/registry_client_reconcile_loop-0b8e528b249cd93e61f4186050c59c02.jpeg"}}]);