"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[79466],{43023:(e,i,t)=>{t.d(i,{R:()=>d,x:()=>l});var s=t(63696);const n={},r=s.createContext(n);function d(e){const i=s.useContext(r);return s.useMemo((function(){return"function"==typeof e?e(i):{...i,...e}}),[i,e])}function l(e){let i;return i=e.disableParentContext?"function"==typeof e.components?e.components(n):e.components||n:d(e.components),s.createElement(r.Provider,{value:i},e.children)}},46879:(e,i,t)=>{t.r(i),t.d(i,{assets:()=>c,contentTitle:()=>l,default:()=>o,frontMatter:()=>d,metadata:()=>s,toc:()=>h});const s=JSON.parse('{"id":"5_references/3_nifi_registry_client","title":"NiFi Registry Client","description":"NifiRegistryClient is the Schema for the NiFi registry client API.","source":"@site/../docs/5_references/3_nifi_registry_client.md","sourceDirName":"5_references","slug":"/5_references/3_nifi_registry_client","permalink":"/nifikop/docs/next/5_references/3_nifi_registry_client","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/../docs/5_references/3_nifi_registry_client.md","tags":[],"version":"current","lastUpdatedBy":"Alexandre Guitton","lastUpdatedAt":1664046775000,"frontMatter":{"id":"3_nifi_registry_client","title":"NiFi Registry Client","sidebar_label":"NiFi Registry Client"},"sidebar":"docs","previous":{"title":"NiFi User","permalink":"/nifikop/docs/next/5_references/2_nifi_user"},"next":{"title":"NiFi Parameter Context","permalink":"/nifikop/docs/next/5_references/4_nifi_parameter_context"}}');var n=t(62540),r=t(43023);const d={id:"3_nifi_registry_client",title:"NiFi Registry Client",sidebar_label:"NiFi Registry Client"},l=void 0,c={},h=[{value:"NifiRegistryClient",id:"nifiregistryclient",level:2},{value:"NifiRegistryClientsSpec",id:"nifiregistryclientsspec",level:2},{value:"NifiRegistryClientStatus",id:"nifiregistryclientstatus",level:2}];function a(e){const i={a:"a",code:"code",h2:"h2",p:"p",pre:"pre",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,r.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsxs)(i.p,{children:[(0,n.jsx)(i.code,{children:"NifiRegistryClient"})," is the Schema for the NiFi registry client API."]}),"\n",(0,n.jsx)(i.pre,{children:(0,n.jsx)(i.code,{className:"language-yaml",children:'apiVersion: nifi.konpyutaika.com/v1alpha1\nkind: NifiRegistryClient\nmetadata:\n  name: squidflow\nspec:\n  clusterRef:\n    name: nc\n    namespace: nifikop\n  description: "Squidflow demo"\n  uri: "http://nifi-registry:18080"\n'})}),"\n",(0,n.jsx)(i.h2,{id:"nifiregistryclient",children:"NifiRegistryClient"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"metadata"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta",children:"ObjectMetadata"})}),(0,n.jsx)(i.td,{children:"is metadata that all persisted resources must have, which includes all objects registry clients must create."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"spec"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"#nifiregistryclientspec",children:"NifiRegistryClientSpec"})}),(0,n.jsx)(i.td,{children:"defines the desired state of NifiRegistryClient."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"status"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"#nifiregistryclientstatus",children:"NifiRegistryClientStatus"})}),(0,n.jsx)(i.td,{children:"defines the observed state of NifiRegistryClient."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"nil"})]})]})]}),"\n",(0,n.jsx)(i.h2,{id:"nifiregistryclientsspec",children:"NifiRegistryClientsSpec"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"description"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"describes the Registry client."}),(0,n.jsx)(i.td,{children:"No"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"uri"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"URI of the NiFi registry that should be used for pulling the flow."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"clusterRef"}),(0,n.jsx)(i.td,{children:(0,n.jsx)(i.a,{href:"./2_nifi_user#clusterreference",children:"ClusterReference"})}),(0,n.jsx)(i.td,{children:"contains the reference to the NifiCluster with the one the user is linked."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"-"})]})]})]}),"\n",(0,n.jsx)(i.h2,{id:"nifiregistryclientstatus",children:"NifiRegistryClientStatus"}),"\n",(0,n.jsxs)(i.table,{children:[(0,n.jsx)(i.thead,{children:(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.th,{children:"Field"}),(0,n.jsx)(i.th,{children:"Type"}),(0,n.jsx)(i.th,{children:"Description"}),(0,n.jsx)(i.th,{children:"Required"}),(0,n.jsx)(i.th,{children:"Default"})]})}),(0,n.jsxs)(i.tbody,{children:[(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"id"}),(0,n.jsx)(i.td,{children:"string"}),(0,n.jsx)(i.td,{children:"nifi registry client's id."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"-"})]}),(0,n.jsxs)(i.tr,{children:[(0,n.jsx)(i.td,{children:"version"}),(0,n.jsx)(i.td,{children:"int64"}),(0,n.jsx)(i.td,{children:"the last nifi registry client revision version catched."}),(0,n.jsx)(i.td,{children:"Yes"}),(0,n.jsx)(i.td,{children:"-"})]})]})]})]})}function o(e={}){const{wrapper:i}={...(0,r.R)(),...e.components};return i?(0,n.jsx)(i,{...e,children:(0,n.jsx)(a,{...e})}):a(e)}}}]);