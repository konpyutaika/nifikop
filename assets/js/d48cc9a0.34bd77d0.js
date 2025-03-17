"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[29519],{43023:(e,t,s)=>{s.d(t,{R:()=>d,x:()=>c});var n=s(63696);const r={},i=n.createContext(r);function d(e){const t=n.useContext(i);return n.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function c(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(r):e.components||r:d(e.components),n.createElement(i.Provider,{value:t},e.children)}},88603:(e,t,s)=>{s.r(t),s.d(t,{assets:()=>l,contentTitle:()=>c,default:()=>a,frontMatter:()=>d,metadata:()=>n,toc:()=>o});const n=JSON.parse('{"id":"5_references/1_nifi_cluster/6_listeners_config","title":"Listeners Config","description":"ListenersConfig defines the Nifi listener types:","source":"@site/versioned_docs/version-v1.11.3/5_references/1_nifi_cluster/6_listeners_config.md","sourceDirName":"5_references/1_nifi_cluster","slug":"/5_references/1_nifi_cluster/6_listeners_config","permalink":"/nifikop/docs/v1.11.3/5_references/1_nifi_cluster/6_listeners_config","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.11.3/5_references/1_nifi_cluster/6_listeners_config.md","tags":[],"version":"v1.11.3","lastUpdatedBy":"Juldrixx","lastUpdatedAt":1731506969000,"frontMatter":{"id":"6_listeners_config","title":"Listeners Config","sidebar_label":"Listeners Config"},"sidebar":"docs","previous":{"title":"Node state","permalink":"/nifikop/docs/v1.11.3/5_references/1_nifi_cluster/5_node_state"},"next":{"title":"External Service Config","permalink":"/nifikop/docs/v1.11.3/5_references/1_nifi_cluster/7_external_service_config"}}');var r=s(62540),i=s(43023);const d={id:"6_listeners_config",title:"Listeners Config",sidebar_label:"Listeners Config"},c=void 0,l={},o=[{value:"ListenersConfig",id:"listenersconfig",level:2},{value:"InternalListener",id:"internallistener",level:2},{value:"SSLSecrets",id:"sslsecrets",level:2}];function h(e){const t={a:"a",code:"code",h2:"h2",p:"p",pre:"pre",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,i.R)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(t.p,{children:"ListenersConfig defines the Nifi listener types:"}),"\n",(0,r.jsx)(t.pre,{children:(0,r.jsx)(t.code,{className:"language-yaml",children:'  listenersConfig:\n    internalListeners:\n      - type: "https"\n        name: "https"\n        containerPort: 8443\n      - type: "cluster"\n        name: "cluster"\n        containerPort: 6007\n      - type: "s2s"\n        name: "s2s"\n        containerPort: 10000\n      - type: "prometheus"\n        name: "prometheus"\n        containerPort: 9090\n      - type: "load-balance"\n        name: "load-balance"\n        containerPort: 6342\n      - name: "my-custom-listener-port"\n        containerPort: 1234\n        protocol: "TCP"\n    sslSecrets:\n      tlsSecretName: "test-nifikop"\n      create: true\n'})}),"\n",(0,r.jsx)(t.h2,{id:"listenersconfig",children:"ListenersConfig"}),"\n",(0,r.jsxs)(t.table,{children:[(0,r.jsx)(t.thead,{children:(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.th,{children:"Field"}),(0,r.jsx)(t.th,{children:"Type"}),(0,r.jsx)(t.th,{children:"Description"}),(0,r.jsx)(t.th,{children:"Required"}),(0,r.jsx)(t.th,{children:"Default"})]})}),(0,r.jsxs)(t.tbody,{children:[(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"internalListeners"}),(0,r.jsxs)(t.td,{children:["[\xa0]",(0,r.jsx)(t.a,{href:"#internallistener",children:"InternalListener"})]}),(0,r.jsx)(t.td,{children:"specifies settings required to access nifi internally."}),(0,r.jsx)(t.td,{children:"Yes"}),(0,r.jsx)(t.td,{children:"-"})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"sslSecrets"}),(0,r.jsx)(t.td,{children:(0,r.jsx)(t.a,{href:"#sslsecrets",children:"SSLSecrets"})}),(0,r.jsx)(t.td,{children:"contains information about ssl related kubernetes secrets if one of the listener setting type set to ssl these fields must be populated to."}),(0,r.jsx)(t.td,{children:"Yes"}),(0,r.jsx)(t.td,{children:"nil"})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"clusterDomain"}),(0,r.jsx)(t.td,{children:"string"}),(0,r.jsx)(t.td,{children:'allow to override the default cluster domain which is "cluster.local".'}),(0,r.jsx)(t.td,{children:"Yes"}),(0,r.jsx)(t.td,{children:(0,r.jsx)(t.code,{children:"cluster.local"})})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"useExternalDNS"}),(0,r.jsx)(t.td,{children:"string"}),(0,r.jsxs)(t.td,{children:["allow to manage externalDNS usage by limiting the DNS names associated to each nodes and load balancer: ",(0,r.jsx)(t.code,{children:"<cluster-name>-node-<node Id>.<cluster-name>.<service name>.<cluster domain>"})]}),(0,r.jsx)(t.td,{children:"Yes"}),(0,r.jsx)(t.td,{children:"false"})]})]})]}),"\n",(0,r.jsx)(t.h2,{id:"internallistener",children:"InternalListener"}),"\n",(0,r.jsxs)(t.table,{children:[(0,r.jsx)(t.thead,{children:(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.th,{children:"Field"}),(0,r.jsx)(t.th,{children:"Type"}),(0,r.jsx)(t.th,{children:"Description"}),(0,r.jsx)(t.th,{children:"Required"}),(0,r.jsx)(t.th,{children:"Default"})]})}),(0,r.jsxs)(t.tbody,{children:[(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"type"}),(0,r.jsxs)(t.td,{children:["enum","load-balance"]}),(0,r.jsx)(t.td,{children:"allow to specify if we are in a specific nifi listener it's allowing to define some required information such as Cluster Port, Http Port, Https Port, S2S, Load Balance port, or Prometheus port"}),(0,r.jsx)(t.td,{children:"Yes"}),(0,r.jsx)(t.td,{children:"-"})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"name"}),(0,r.jsx)(t.td,{children:"string"}),(0,r.jsx)(t.td,{children:"an identifier for the port which will be configured."}),(0,r.jsx)(t.td,{children:"Yes"}),(0,r.jsx)(t.td,{children:"-"})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"containerPort"}),(0,r.jsx)(t.td,{children:"int32"}),(0,r.jsx)(t.td,{children:"the containerPort."}),(0,r.jsx)(t.td,{children:"Yes"}),(0,r.jsx)(t.td,{children:"-"})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"protocol"}),(0,r.jsx)(t.td,{children:(0,r.jsx)(t.a,{href:"https://pkg.go.dev/k8s.io/api/core/v1#Protocol",children:"Protocol"})}),(0,r.jsx)(t.td,{children:"the network protocol for this listener. Must be one of the protocol enum values (i.e. TCP, UDP, SCTP)."}),(0,r.jsx)(t.td,{children:"No"}),(0,r.jsx)(t.td,{children:(0,r.jsx)(t.code,{children:"TCP"})})]})]})]}),"\n",(0,r.jsx)(t.h2,{id:"sslsecrets",children:"SSLSecrets"}),"\n",(0,r.jsxs)(t.table,{children:[(0,r.jsx)(t.thead,{children:(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.th,{children:"Field"}),(0,r.jsx)(t.th,{children:"Type"}),(0,r.jsx)(t.th,{children:"Description"}),(0,r.jsx)(t.th,{children:"Required"}),(0,r.jsx)(t.th,{children:"Default"})]})}),(0,r.jsxs)(t.tbody,{children:[(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"tlsSecretName"}),(0,r.jsx)(t.td,{children:"string"}),(0,r.jsx)(t.td,{children:"should contain all ssl certs required by nifi including: caCert, caKey, clientCert, clientKey serverCert, serverKey, peerCert, peerKey."}),(0,r.jsx)(t.td,{children:"Yes"}),(0,r.jsx)(t.td,{children:"-"})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"create"}),(0,r.jsx)(t.td,{children:"boolean"}),(0,r.jsx)(t.td,{children:"tells the installed cert manager to create the required certs keys."}),(0,r.jsx)(t.td,{children:"Yes"}),(0,r.jsx)(t.td,{children:"-"})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"clusterScoped"}),(0,r.jsx)(t.td,{children:"boolean"}),(0,r.jsx)(t.td,{children:"defines if the Issuer created is cluster or namespace scoped."}),(0,r.jsx)(t.td,{children:"Yes"}),(0,r.jsx)(t.td,{children:"-"})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"issuerRef"}),(0,r.jsx)(t.td,{children:(0,r.jsx)(t.a,{href:"https://docs.cert-manager.io/en/release-0.9/reference/api-docs/index.html#objectreference-v1alpha1",children:"ObjectReference"})}),(0,r.jsxs)(t.td,{children:["IssuerRef allow to use an existing issuer to act as CA: ",(0,r.jsx)(t.a,{href:"https://cert-manager.io/docs/concepts/issuer/",children:"https://cert-manager.io/docs/concepts/issuer/"})]}),(0,r.jsx)(t.td,{children:"No"}),(0,r.jsx)(t.td,{children:"-"})]}),(0,r.jsxs)(t.tr,{children:[(0,r.jsx)(t.td,{children:"pkiBackend"}),(0,r.jsxs)(t.td,{children:["enum","cert-manager"]}),(0,r.jsx)(t.td,{}),(0,r.jsx)(t.td,{children:"Yes"}),(0,r.jsx)(t.td,{children:"-"})]})]})]})]})}function a(e={}){const{wrapper:t}={...(0,i.R)(),...e.components};return t?(0,r.jsx)(t,{...e,children:(0,r.jsx)(h,{...e})}):h(e)}}}]);