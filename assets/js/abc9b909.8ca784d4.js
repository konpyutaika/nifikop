"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[53863],{43023:(e,n,t)=>{t.d(n,{R:()=>a,x:()=>r});var i=t(63696);const s={},o=i.createContext(s);function a(e){const n=i.useContext(o);return i.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function r(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(s):e.components||s:a(e.components),i.createElement(o.Provider,{value:n},e.children)}},92525:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>c,contentTitle:()=>r,default:()=>h,frontMatter:()=>a,metadata:()=>i,toc:()=>l});const i=JSON.parse('{"id":"3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/2_istio_service_mesh","title":"Istio service mesh","description":"The purpose of this section is to explain how to expose your NiFi cluster using Istio on Kubernetes.","source":"@site/versioned_docs/version-v1.6.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/2_istio_service_mesh.md","sourceDirName":"3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster","slug":"/3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/2_istio_service_mesh","permalink":"/nifikop/docs/v1.6.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/2_istio_service_mesh","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.6.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/2_istio_service_mesh.md","tags":[],"version":"v1.6.0","lastUpdatedBy":"Juldrixx","lastUpdatedAt":1702899203000,"frontMatter":{"id":"2_istio_service_mesh","title":"Istio service mesh","sidebar_label":"Istio service mesh"},"sidebar":"docs","previous":{"title":"Kubernetes service","permalink":"/nifikop/docs/v1.6.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/1_kubernetes_service"},"next":{"title":"SSL configuration","permalink":"/nifikop/docs/v1.6.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/4_ssl_configuration"}}');var s=t(62540),o=t(43023);const a={id:"2_istio_service_mesh",title:"Istio service mesh",sidebar_label:"Istio service mesh"},r=void 0,c={},l=[{value:"Istio configuration for HTTP",id:"istio-configuration-for-http",level:2},{value:"Istio configuration for HTTPS",id:"istio-configuration-for-https",level:2}];function d(e){const n={code:"code",h2:"h2",p:"p",pre:"pre",...(0,o.R)(),...e.components};return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(n.p,{children:"The purpose of this section is to explain how to expose your NiFi cluster using Istio on Kubernetes."}),"\n",(0,s.jsx)(n.h2,{id:"istio-configuration-for-http",children:"Istio configuration for HTTP"}),"\n",(0,s.jsxs)(n.p,{children:["To access to the NiFi cluster from the external world, it is needed to configure a ",(0,s.jsx)(n.code,{children:"Gateway"})," and a ",(0,s.jsx)(n.code,{children:"VirtualService"})," on Istio."]}),"\n",(0,s.jsxs)(n.p,{children:["The following example show how to define a ",(0,s.jsx)(n.code,{children:"Gateway"})," that will be able to intercept all requests for a specific domain host on HTTP port 80."]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-yaml",children:"apiVersion: networking.istio.io/v1beta1\nkind: Gateway\nmetadata:\n  name: nifi-gateway\nspec:\n  selector:\n    istio: ingressgateway\n  servers:\n  - port:\n      number: 80\n      name: http\n      protocol: HTTP\n    hosts:\n    - nifi.my-domain.com\n"})}),"\n",(0,s.jsxs)(n.p,{children:["In combination, we need to define also a ",(0,s.jsx)(n.code,{children:"VirtualService"})," that redirect all requests itercepted by the ",(0,s.jsx)(n.code,{children:"Gateway"})," to a specific service."]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-yaml",children:"apiVersion: networking.istio.io/v1beta1\nkind: VirtualService\nmetadata:\n  name: nifi-vs\nspec:\n  gateways:\n  - nifi-gateway\n  hosts:\n  - nifi.my-domain.com\n  http:\n  - match:\n    - uri:\n        prefix: /\n    route:\n    - destination:\n        host: nifi\n        port:\n          number: 8080\n"})}),"\n",(0,s.jsx)(n.h2,{id:"istio-configuration-for-https",children:"Istio configuration for HTTPS"}),"\n",(0,s.jsxs)(n.p,{children:["In case you are deploying a cluster and you want to enable the HTTPS protocol to manage user authentication, the configuration is more complex. To understand the reason of this, it is important to explain that in this scenario the HTTPS protocol with all certificates is managed directly by NiFi. This means that all requests passes through all Istio services in an encrypted way, so Istio can't manage a sticky session.\nTo solve this issue, the tricky is to limit the HTTPS session till the ",(0,s.jsx)(n.code,{children:"Gateway"}),", then decrypt all requests in HTTP, manage the sticky session and then encrypt again in HTTPS before reach the NiFi node.\nIstio allows to do this using a destination rule combined with the ",(0,s.jsx)(n.code,{children:"VirtualService"}),". In the next example, we define a ",(0,s.jsx)(n.code,{children:"Gateway"})," that accepts HTTPS traffic and transform it in HTTP."]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-yaml",children:"apiVersion: networking.istio.io/v1beta1\nkind: Gateway\nmetadata:\n  name: nifi-gateway\nspec:\n  selector:\n    istio: ingressgateway\n  servers:\n  - port:\n      number: 443\n      name: https\n      protocol: HTTPS\n    tls:\n      mode: SIMPLE\n      credentialName: my-secret\n    hosts:\n    - nifi.my-domain.com\n"})}),"\n",(0,s.jsxs)(n.p,{children:["In combination, we need to define also a ",(0,s.jsx)(n.code,{children:"VirtualService"})," that redirect all HTTP traffic to a specific the ",(0,s.jsx)(n.code,{children:"ClusterIP"})," Service."]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-yaml",children:"apiVersion: networking.istio.io/v1beta1\nkind: VirtualService\nmetadata:\n  name: nifi-vs\nspec:\n  gateways:\n  - nifi-gateway\n  hosts:\n  - nifi.my-domain.com\n  http:\n  - match:\n    - uri:\n        prefix: /\n    route:\n    - destination:\n        host: <service-name>.<namespace>.svc.cluster.local\n        port:\n          number: 8443\n"})}),"\n",(0,s.jsxs)(n.p,{children:["Please note that the service name configured as destination of the ",(0,s.jsx)(n.code,{children:"VirtualService"})," is the name of the Service created with the following section in the cluster Deployment YAML"]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-yaml",children:'spec:  \n  externalServices:  \n    - name: "nifi-cluster"\n      spec:\n        type: ClusterIP\n        portConfigs:\n          - port: 8443\n            internalListenerName: "https"\n'})}),"\n",(0,s.jsxs)(n.p,{children:["Finally the destination rule that redirect all HTTP traffic destinated to the ",(0,s.jsx)(n.code,{children:"ClusterIP"})," Service to HTTPS encrypting it."]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-yaml",children:"apiVersion: networking.istio.io/v1beta1\nkind: DestinationRule\nmetadata:\n  name: nifi-dr\nspec:\n  host: <service-name>.<namespace>.svc.cluster.local\n  trafficPolicy:\n    tls:\n      mode: SIMPLE\n    loadBalancer:\n      consistentHash:\n        httpCookie:\n          name: __Secure-Authorization-Bearer\n          ttl: 0s\n"})}),"\n",(0,s.jsx)(n.p,{children:"As you can see in the previous example, the destination rule also define how manage the sticky session based on httpCookie property called __Secure-Authorization-Bearer."})]})}function h(e={}){const{wrapper:n}={...(0,o.R)(),...e.components};return n?(0,s.jsx)(n,{...e,children:(0,s.jsx)(d,{...e})}):d(e)}}}]);