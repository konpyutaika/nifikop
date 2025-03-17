"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[66975],{43023:(e,t,n)=>{n.d(t,{R:()=>o,x:()=>a});var i=n(63696);const s={},r=i.createContext(s);function o(e){const t=i.useContext(r);return i.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function a(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(s):e.components||s:o(e.components),i.createElement(r.Provider,{value:t},e.children)}},99198:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>c,contentTitle:()=>a,default:()=>u,frontMatter:()=>o,metadata:()=>i,toc:()=>l});const i=JSON.parse('{"id":"3_manage_nifi/1_manage_clusters/3_external_cluster","title":"External cluster","description":"Common configuration","source":"@site/versioned_docs/version-v1.11.2/3_manage_nifi/1_manage_clusters/3_external_cluster.md","sourceDirName":"3_manage_nifi/1_manage_clusters","slug":"/3_manage_nifi/1_manage_clusters/3_external_cluster","permalink":"/nifikop/docs/v1.11.2/3_manage_nifi/1_manage_clusters/3_external_cluster","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.11.2/3_manage_nifi/1_manage_clusters/3_external_cluster.md","tags":[],"version":"v1.11.2","lastUpdatedBy":"Juldrixx","lastUpdatedAt":1731434746000,"frontMatter":{"id":"3_external_cluster","title":"External cluster","sidebar_label":"External cluster"},"sidebar":"docs","previous":{"title":"Using KEDA","permalink":"/nifikop/docs/v1.11.2/3_manage_nifi/1_manage_clusters/2_cluster_scaling/2_auto_scaling/1_using_keda"},"next":{"title":"Users management","permalink":"/nifikop/docs/v1.11.2/3_manage_nifi/2_manage_users_and_accesses/1_users_management"}}');var s=n(62540),r=n(43023);const o={id:"3_external_cluster",title:"External cluster",sidebar_label:"External cluster"},a=void 0,c={},l=[{value:"Common configuration",id:"common-configuration",level:2},{value:"Secret configuration for Basic authentication",id:"secret-configuration-for-basic-authentication",level:2},{value:"Secret configuration for TLS authentication",id:"secret-configuration-for-tls-authentication",level:2}];function d(e){const t={admonition:"admonition",code:"code",h2:"h2",li:"li",p:"p",pre:"pre",ul:"ul",...(0,r.R)(),...e.components};return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(t.h2,{id:"common-configuration",children:"Common configuration"}),"\n",(0,s.jsxs)(t.p,{children:["The operator allows you to manage the Dataflow lifecycle for internal (i.e cluster managed by the operator) and external NiFi cluster.\nA NiFi cluster is considered as external as soon as the ",(0,s.jsx)(t.code,{children:"NifiCluster"})," resource used as reference in other NiFi resource explicitly detailed the way to communicate with the cluster."]}),"\n",(0,s.jsx)(t.p,{children:"This feature allows you:"}),"\n",(0,s.jsxs)(t.ul,{children:["\n",(0,s.jsx)(t.li,{children:"To automate your Dataflow CI/CD using yaml"}),"\n",(0,s.jsx)(t.li,{children:"To manage the same way your Dataflow management wherever your cluster is, on bare metal, VMs, k8s, on-premise or on cloud."}),"\n"]}),"\n",(0,s.jsxs)(t.p,{children:["To deploy different resources (",(0,s.jsx)(t.code,{children:"NifiRegistryClient"}),", ",(0,s.jsx)(t.code,{children:"NifiUser"}),", ",(0,s.jsx)(t.code,{children:"NifiUserGroup"}),", ",(0,s.jsx)(t.code,{children:"NifiParameterContext"}),", ",(0,s.jsx)(t.code,{children:"NifiDataflow"}),") you simply have to declare a ",(0,s.jsx)(t.code,{children:"NifiCluster"})," resource explaining how to discuss with the external cluster, and refer to this resource as usual using the ",(0,s.jsx)(t.code,{children:"Spec.ClusterRef"})," field."]}),"\n",(0,s.jsx)(t.p,{children:"To declare an external cluster you have to follow this kind of configuration:"}),"\n",(0,s.jsx)(t.pre,{children:(0,s.jsx)(t.code,{className:"language-yaml",children:"apiVersion: nifi.konpyutaika.com/v1\nkind: NifiCluster\nmetadata:\n  name: externalcluster\nspec:\n  # rootProcessGroupId contains the uuid of the root process group for this cluster.\n  rootProcessGroupId: 'd37bee03-017a-1000-cff7-4eaaa82266b7'\n  # nodeURITemplate used to dynamically compute node uri.\n  nodeURITemplate: 'nifi0%d.integ.mapreduce.m0.p.fti.net:9090'\n  # all node requiresunique id\n  nodes:\n    - id: 1\n    - id: 2\n    - id: 3\n  # type defines if the cluster is internal (i.e manager by the operator) or external.\n  # :Enum={\"external\",\"internal\"}\n  type: 'external'\n  # clientType defines if the operator will use basic or tls authentication to query the NiFi cluster.\n  # Enum={\"tls\",\"basic\"}\n  clientType: 'basic'\n  # secretRef reference the secret containing the informations required to authenticate to the cluster.\n  secretRef:\n    name: nifikop-credentials\n    namespace: nifikop-nifi\n"})}),"\n",(0,s.jsxs)(t.ul,{children:["\n",(0,s.jsxs)(t.li,{children:["The ",(0,s.jsx)(t.code,{children:"Spec.RootProcessGroupId"})," field is required to give the ability to the operator of managing root level policy and default deployment and policy."]}),"\n",(0,s.jsxs)(t.li,{children:["The ",(0,s.jsx)(t.code,{children:"Spec.NodeURITemplate"})," field, defines the hostname template of your NiFi cluster nodes, the operator will use this information and the list of id specified in ",(0,s.jsx)(t.code,{children:"Spec.Nodes"})," field to generate the hostname of the nodes (in the configuration above you will have: ",(0,s.jsx)(t.code,{children:"nifi01.integ.mapreduce.m0.p.fti.net:9090"}),", ",(0,s.jsx)(t.code,{children:"nifi02.integ.mapreduce.m0.p.fti.net:9090"}),", ",(0,s.jsx)(t.code,{children:"nifi03.integ.mapreduce.m0.p.fti.net:9090"}),")."]}),"\n",(0,s.jsxs)(t.li,{children:["The ",(0,s.jsx)(t.code,{children:"Spec.Type"})," field defines the type of cluster that this resource is refering to, by default it is ",(0,s.jsx)(t.code,{children:"internal"}),", in our case here we just want to use this resource to reference an existing NiFi cluster, so we set this field to ",(0,s.jsx)(t.code,{children:"external"}),"."]}),"\n",(0,s.jsxs)(t.li,{children:["The ",(0,s.jsx)(t.code,{children:"Spec.ClientType"})," field defines how we want to authenticate to the NiFi cluster API, for now we are supporting two modes:","\n",(0,s.jsxs)(t.ul,{children:["\n",(0,s.jsxs)(t.li,{children:[(0,s.jsx)(t.code,{children:"tls"}),": using client TLS certificate."]}),"\n",(0,s.jsxs)(t.li,{children:[(0,s.jsx)(t.code,{children:"basic"}),": using a username and a password to get an access token."]}),"\n"]}),"\n"]}),"\n",(0,s.jsxs)(t.li,{children:["The ",(0,s.jsx)(t.code,{children:"Spec.SecretRef"})," defines a reference to a secret which contains the sensitive values that will be used by the operator to authenticate to the NiFi cluster API (ie in basic mode it will contain the password and username)."]}),"\n"]}),"\n",(0,s.jsx)(t.admonition,{type:"warning",children:(0,s.jsxs)(t.p,{children:["The id of node only support ",(0,s.jsx)(t.code,{children:"int32"})," as type, so if the hostname of your nodes doesn't match with this, you can't use this feature."]})}),"\n",(0,s.jsx)(t.h2,{id:"secret-configuration-for-basic-authentication",children:"Secret configuration for Basic authentication"}),"\n",(0,s.jsxs)(t.p,{children:["When you are using the basic authentication, you have to pass some informations into the secret that is referenced into the ",(0,s.jsx)(t.code,{children:"NifiCluster"})," resource:"]}),"\n",(0,s.jsxs)(t.ul,{children:["\n",(0,s.jsxs)(t.li,{children:[(0,s.jsx)(t.code,{children:"username"}),": the username associated to the user that will be used by the operator to request the REST API."]}),"\n",(0,s.jsxs)(t.li,{children:[(0,s.jsx)(t.code,{children:"password"}),": the password associated to the user that will be used by the operator to request the REST API."]}),"\n",(0,s.jsxs)(t.li,{children:[(0,s.jsx)(t.code,{children:"ca.crt (optional)"}),": the certificate authority to trust the server certificate if needed"]}),"\n"]}),"\n",(0,s.jsx)(t.p,{children:"The following command shows how you can create this secret:"}),"\n",(0,s.jsx)(t.pre,{children:(0,s.jsx)(t.code,{className:"language-console",children:"kubectl create secret generic nifikop-credentials \\\n  --from-file=username=./secrets/username\\\n  --from-file=password=./secrets/password\\\n  --from-file=ca.crt=./secrets/ca.crt\\\n  -n nifikop-nifi\n"})}),"\n",(0,s.jsx)(t.admonition,{type:"info",children:(0,s.jsxs)(t.p,{children:["When you use the basic authentication, the operator will create a secret ",(0,s.jsx)(t.code,{children:"<cluster name>-basic-secret"})," containing for each node an access token that will be maintained by the operator."]})}),"\n",(0,s.jsx)(t.h2,{id:"secret-configuration-for-tls-authentication",children:"Secret configuration for TLS authentication"}),"\n",(0,s.jsxs)(t.p,{children:["When you are using the tls authentication, you have to pass some information into the secret that is referenced into the ",(0,s.jsx)(t.code,{children:"NifiCluster"})," resource:"]}),"\n",(0,s.jsxs)(t.ul,{children:["\n",(0,s.jsxs)(t.li,{children:[(0,s.jsx)(t.code,{children:"tls.key"}),": The user private key."]}),"\n",(0,s.jsxs)(t.li,{children:[(0,s.jsx)(t.code,{children:"tls.crt"}),": The user certificate."]}),"\n",(0,s.jsxs)(t.li,{children:[(0,s.jsx)(t.code,{children:"password"}),": the password associated to the user that will be used by the operator to request the REST API."]}),"\n",(0,s.jsxs)(t.li,{children:[(0,s.jsx)(t.code,{children:"ca.crt"}),": The CA certificate"]}),"\n",(0,s.jsxs)(t.li,{children:[(0,s.jsx)(t.code,{children:"truststore.jks"}),":"]}),"\n",(0,s.jsxs)(t.li,{children:[(0,s.jsx)(t.code,{children:"keystore.jks"}),":"]}),"\n"]})]})}function u(e={}){const{wrapper:t}={...(0,r.R)(),...e.components};return t?(0,s.jsx)(t,{...e,children:(0,s.jsx)(d,{...e})}):d(e)}}}]);