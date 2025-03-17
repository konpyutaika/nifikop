"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[17104],{43023:(e,i,n)=>{n.d(i,{R:()=>s,x:()=>a});var t=n(63696);const r={},o=t.createContext(r);function s(e){const i=t.useContext(o);return t.useMemo((function(){return"function"==typeof e?e(i):{...i,...e}}),[i,e])}function a(e){let i;return i=e.disableParentContext?"function"==typeof e.components?e.components(r):e.components||r:s(e.components),t.createElement(o.Provider,{value:i},e.children)}},95284:(e,i,n)=>{n.r(i),n.d(i,{assets:()=>c,contentTitle:()=>a,default:()=>l,frontMatter:()=>s,metadata:()=>t,toc:()=>d});const t=JSON.parse('{"id":"3_manage_nifi/1_manage_clusters/1_deploy_cluster/5_users_authentication/1_oidc","title":"OpenId Connect","description":"To enable authentication via OpenId Connect refering to NiFi Administration guide required some configuration into nifi.properties.","source":"@site/versioned_docs/version-v1.3.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/5_users_authentication/1_oidc.md","sourceDirName":"3_manage_nifi/1_manage_clusters/1_deploy_cluster/5_users_authentication","slug":"/3_manage_nifi/1_manage_clusters/1_deploy_cluster/5_users_authentication/1_oidc","permalink":"/nifikop/docs/v1.3.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/5_users_authentication/1_oidc","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.3.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/5_users_authentication/1_oidc.md","tags":[],"version":"v1.3.0","lastUpdatedBy":"Juldrixx","lastUpdatedAt":1707144987000,"frontMatter":{"id":"1_oidc","title":"OpenId Connect","sidebar_label":"OpenId Connect"},"sidebar":"docs","previous":{"title":"SSL configuration","permalink":"/nifikop/docs/v1.3.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/4_ssl_configuration"},"next":{"title":"Custom User Authorizers","permalink":"/nifikop/docs/v1.3.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/6_users_authorization/1_custom_user_authorizer"}}');var r=n(62540),o=n(43023);const s={id:"1_oidc",title:"OpenId Connect",sidebar_label:"OpenId Connect"},a=void 0,c={},d=[];function u(e){const i={a:"a",code:"code",p:"p",pre:"pre",...(0,o.R)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsxs)(i.p,{children:["To enable authentication via OpenId Connect refering to ",(0,r.jsx)(i.a,{href:"https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html",children:"NiFi Administration guide"})," required some configuration into ",(0,r.jsx)(i.code,{children:"nifi.properties"}),"."]}),"\n",(0,r.jsxs)(i.p,{children:["In addition and to ensure multiple identity provider support, we recommended to add the following configuration to your ",(0,r.jsx)(i.code,{children:"nifi.properties"}),":"]}),"\n",(0,r.jsx)(i.pre,{children:(0,r.jsx)(i.code,{className:"language-sh",children:"nifi.security.identity.mapping.pattern.dn=CN=([^,]*)(?:, (?:O|OU)=.*)?\nnifi.security.identity.mapping.value.dn=$1\nnifi.security.identity.mapping.transform.dn=NONE\n"})}),"\n",(0,r.jsxs)(i.p,{children:["To perform this with ",(0,r.jsx)(i.code,{children:"NiFiKop"})," you just have to configure the ",(0,r.jsx)(i.code,{children:"Spec.NifiProperties.OverrideConfigs"})," field with your OIDC configuration, for example:"]}),"\n",(0,r.jsx)(i.pre,{children:(0,r.jsx)(i.code,{className:"language-yaml",children:"apiVersion: nifi.konpyutaika.com/v1\nkind: NifiCluster\n...\nspec:\n  ...\n  readOnlyConfig:\n    # NifiProperties configuration that will be applied to the node.\n    nifiProperties:\n      webProxyHosts:\n        - nifistandard2.trycatchlearn.fr:8443\n      # Additionnal nifi.properties configuration that will override the one produced based\n      # on template and configurations.\n      overrideConfigs: |\n        nifi.security.user.oidc.discovery.url=<oidc server discovery url>\n        nifi.security.user.oidc.client.id=<oidc client's id>\n        nifi.security.user.oidc.client.secret=<oidc client's secret>\n        nifi.security.identity.mapping.pattern.dn=CN=([^,]*)(?:, (?:O|OU)=.*)?\n        nifi.security.identity.mapping.value.dn=$1\n        nifi.security.identity.mapping.transform.dn=NONE\n      ...\n   ...\n...\n"})})]})}function l(e={}){const{wrapper:i}={...(0,o.R)(),...e.components};return i?(0,r.jsx)(i,{...e,children:(0,r.jsx)(u,{...e})}):u(e)}}}]);