"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[78828],{2136:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>d,contentTitle:()=>r,default:()=>u,frontMatter:()=>s,metadata:()=>i,toc:()=>c});const i=JSON.parse('{"id":"7_upgrade_guides/2_v0.14.1_to_v0.15.0","title":"v0.14.1 to v0.15.0","description":"PR #189 changed the default Zookeeper init container image changed from busybox to bash. If you have overridden the NifiCluster.Spec.InitContainerImage then you need to change it to bash or one that contains a bash shell.","source":"@site/versioned_docs/version-v1.13.0/7_upgrade_guides/2_v0.14.1_to_v0.15.0.md","sourceDirName":"7_upgrade_guides","slug":"/7_upgrade_guides/2_v0.14.1_to_v0.15.0","permalink":"/nifikop/docs/7_upgrade_guides/2_v0.14.1_to_v0.15.0","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.13.0/7_upgrade_guides/2_v0.14.1_to_v0.15.0.md","tags":[],"version":"v1.13.0","lastUpdatedBy":"Julien Guitton","lastUpdatedAt":1742252278000,"frontMatter":{"id":"2_v0.14.1_to_v0.15.0","title":"v0.14.1 to v0.15.0","sidebar_label":"v0.14.1 to v0.15.0"},"sidebar":"docs","previous":{"title":"v0.7.x to v0.8.0","permalink":"/nifikop/docs/7_upgrade_guides/1_v0.7.x_to_v0.8.0"},"next":{"title":"v0.16.0 to v1.0.0","permalink":"/nifikop/docs/7_upgrade_guides/3_v0.16.0_to_v1.0.0"}}');var a=n(62540),o=n(43023);const s={id:"2_v0.14.1_to_v0.15.0",title:"v0.14.1 to v0.15.0",sidebar_label:"v0.14.1 to v0.15.0"},r=void 0,d={},c=[{value:"Getting started",id:"getting-started",level:2}];function l(e){const t={a:"a",code:"code",h2:"h2",p:"p",pre:"pre",...(0,o.R)(),...e.components};return(0,a.jsxs)(a.Fragment,{children:[(0,a.jsxs)(t.p,{children:[(0,a.jsx)(t.a,{href:"https://github.com/konpyutaika/nifikop/pull/189",children:"PR #189"})," changed the default Zookeeper init container image changed from ",(0,a.jsx)(t.code,{children:"busybox"})," to ",(0,a.jsx)(t.code,{children:"bash"}),". If you have overridden the ",(0,a.jsx)(t.code,{children:"NifiCluster.Spec.InitContainerImage"})," then you need to change it to ",(0,a.jsx)(t.code,{children:"bash"})," or one that contains a bash shell."]}),"\n",(0,a.jsx)(t.h2,{id:"getting-started",children:"Getting started"}),"\n",(0,a.jsxs)(t.p,{children:["If you haven't overridden the default ",(0,a.jsx)(t.code,{children:"NifiCluster.Spec.InitContainerImage"}),", then there are no special upgrade instructions. If you have, like for example below:"]}),"\n",(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-yaml",children:'apiVersion: nifi.konpyutaika.com/v1alpha1\nkind: NifiCluster\nmetadata:\n  name: mynifi\nspec:\n  initContainerImage:\n    repository: busybox\n    tag: "1.34.0"\n'})}),"\n",(0,a.jsxs)(t.p,{children:["Then you must change it to ",(0,a.jsx)(t.code,{children:"bash"})," or an image that contains a bash shell:"]}),"\n",(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-yaml",children:'apiVersion: nifi.konpyutaika.com/v1alpha1\nkind: NifiCluster\nmetadata:\n  name: mynifi\nspec:\n  initContainerImage:\n    repository: bash\n    tag: "5.2.2"\n'})})]})}function u(e={}){const{wrapper:t}={...(0,o.R)(),...e.components};return t?(0,a.jsx)(t,{...e,children:(0,a.jsx)(l,{...e})}):l(e)}},43023:(e,t,n)=>{n.d(t,{R:()=>s,x:()=>r});var i=n(63696);const a={},o=i.createContext(a);function s(e){const t=i.useContext(o);return i.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function r(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(a):e.components||a:s(e.components),i.createElement(o.Provider,{value:t},e.children)}}}]);