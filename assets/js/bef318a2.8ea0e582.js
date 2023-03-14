"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[903],{35318:(e,t,n)=>{n.d(t,{Zo:()=>p,kt:()=>d});var i=n(27378);function a(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function r(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);t&&(i=i.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,i)}return n}function o(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?r(Object(n),!0).forEach((function(t){a(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):r(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function s(e,t){if(null==e)return{};var n,i,a=function(e,t){if(null==e)return{};var n,i,a={},r=Object.keys(e);for(i=0;i<r.length;i++)n=r[i],t.indexOf(n)>=0||(a[n]=e[n]);return a}(e,t);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);for(i=0;i<r.length;i++)n=r[i],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(a[n]=e[n])}return a}var l=i.createContext({}),c=function(e){var t=i.useContext(l),n=t;return e&&(n="function"==typeof e?e(t):o(o({},t),e)),n},p=function(e){var t=c(e.components);return i.createElement(l.Provider,{value:t},e.children)},u={inlineCode:"code",wrapper:function(e){var t=e.children;return i.createElement(i.Fragment,{},t)}},m=i.forwardRef((function(e,t){var n=e.components,a=e.mdxType,r=e.originalType,l=e.parentName,p=s(e,["components","mdxType","originalType","parentName"]),m=c(n),d=a,f=m["".concat(l,".").concat(d)]||m[d]||u[d]||r;return n?i.createElement(f,o(o({ref:t},p),{},{components:n})):i.createElement(f,o({ref:t},p))}));function d(e,t){var n=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var r=n.length,o=new Array(r);o[0]=m;var s={};for(var l in t)hasOwnProperty.call(t,l)&&(s[l]=t[l]);s.originalType=e,s.mdxType="string"==typeof e?e:a,o[1]=s;for(var c=2;c<r;c++)o[c]=n[c];return i.createElement.apply(null,o)}return i.createElement.apply(null,n)}m.displayName="MDXCreateElement"},7894:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>p,contentTitle:()=>l,default:()=>d,frontMatter:()=>s,metadata:()=>c,toc:()=>u});var i=n(25773),a=n(30808),r=(n(27378),n(35318)),o=["components"],s={id:"2_istio_service_mesh",title:"Istio service mesh",sidebar_label:"Istio service mesh"},l=void 0,c={unversionedId:"3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/2_istio_service_mesh",id:"version-v1.0.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/2_istio_service_mesh",title:"Istio service mesh",description:"The purpose of this section is to explain how to expose your NiFi cluster using Istio on Kubernetes.",source:"@site/versioned_docs/version-v1.0.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/2_istio_service_mesh.md",sourceDirName:"3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster",slug:"/3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/2_istio_service_mesh",permalink:"/nifikop/docs/3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/2_istio_service_mesh",draft:!1,editUrl:"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.0.0/3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/2_istio_service_mesh.md",tags:[],version:"v1.0.0",lastUpdatedBy:"Giuseppe Gerla",lastUpdatedAt:1678782172,formattedLastUpdatedAt:"Mar 14, 2023",frontMatter:{id:"2_istio_service_mesh",title:"Istio service mesh",sidebar_label:"Istio service mesh"},sidebar:"docs",previous:{title:"Kubernetes service",permalink:"/nifikop/docs/3_manage_nifi/1_manage_clusters/1_deploy_cluster/3_expose_cluster/1_kubernetes_service"},next:{title:"SSL configuration",permalink:"/nifikop/docs/3_manage_nifi/1_manage_clusters/1_deploy_cluster/4_ssl_configuration"}},p={},u=[{value:"Istio configuration for HTTP",id:"istio-configuration-for-http",level:2},{value:"Istio configuration for HTTPS",id:"istio-configuration-for-https",level:2}],m={toc:u};function d(e){var t=e.components,n=(0,a.Z)(e,o);return(0,r.kt)("wrapper",(0,i.Z)({},m,n,{components:t,mdxType:"MDXLayout"}),(0,r.kt)("p",null,"The purpose of this section is to explain how to expose your NiFi cluster using Istio on Kubernetes."),(0,r.kt)("h2",{id:"istio-configuration-for-http"},"Istio configuration for HTTP"),(0,r.kt)("p",null,"To access to the NiFi cluster from the external world, it is needed to configure a ",(0,r.kt)("inlineCode",{parentName:"p"},"Gateway")," and a ",(0,r.kt)("inlineCode",{parentName:"p"},"VirtualService")," on Istio."),(0,r.kt)("p",null,"The following example show how to define a ",(0,r.kt)("inlineCode",{parentName:"p"},"Gateway")," that will be able to intercept all requests for a specific domain host on HTTP port 80."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml"},"apiVersion: networking.istio.io/v1beta1\nkind: Gateway\nmetadata:\n  name: nifi-gateway\nspec:\n  selector:\n    istio: ingressgateway\n  servers:\n  - port:\n      number: 80\n      name: http\n      protocol: HTTP\n    hosts:\n    - nifi.my-domain.com\n")),(0,r.kt)("p",null,"In combination, we need to define also a ",(0,r.kt)("inlineCode",{parentName:"p"},"VirtualService")," that redirect all requests itercepted by the ",(0,r.kt)("inlineCode",{parentName:"p"},"Gateway")," to a specific service. "),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml"},"apiVersion: networking.istio.io/v1beta1\nkind: VirtualService\nmetadata:\n  name: nifi-vs\nspec:\n  gateways:\n  - nifi-gateway\n  hosts:\n  - nifi.my-domain.com\n  http:\n  - match:\n    - uri:\n        prefix: /\n    route:\n    - destination:\n        host: nifi\n        port:\n          number: 8080\n")),(0,r.kt)("h2",{id:"istio-configuration-for-https"},"Istio configuration for HTTPS"),(0,r.kt)("p",null,"In case you are deploying a cluster and you want to enable the HTTPS protocol to manage user authentication, the configuration is more complex. To understand the reason of this, it is important to explain that in this scenario the HTTPS protocol with all certificates is managed directly by NiFi. This means that all requests passes through all Istio services in an encrypted way, so Istio can't manage a sticky session.\nTo solve this issue, the tricky is to limit the HTTPS session till the ",(0,r.kt)("inlineCode",{parentName:"p"},"Gateway"),", then decrypt all requests in HTTP, manage the sticky session and then encrypt again in HTTPS before reach the NiFi node.\nIstio allows to do this using a destination rule combined with the ",(0,r.kt)("inlineCode",{parentName:"p"},"VirtualService"),". In the next example, we define a ",(0,r.kt)("inlineCode",{parentName:"p"},"Gateway")," that accepts HTTPS traffic and transform it in HTTP."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml"},"apiVersion: networking.istio.io/v1beta1\nkind: Gateway\nmetadata:\n  name: nifi-gateway\nspec:\n  selector:\n    istio: ingressgateway\n  servers:\n  - port:\n      number: 443\n      name: https\n      protocol: HTTPS\n    tls:\n      mode: SIMPLE\n      credentialName: my-secret\n    hosts:\n    - nifi.my-domain.com\n")),(0,r.kt)("p",null,"In combination, we need to define also a ",(0,r.kt)("inlineCode",{parentName:"p"},"VirtualService")," that redirect all HTTP traffic to a specific the ",(0,r.kt)("inlineCode",{parentName:"p"},"ClusterIP")," Service. "),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml"},"apiVersion: networking.istio.io/v1beta1\nkind: VirtualService\nmetadata:\n  name: nifi-vs\nspec:\n  gateways:\n  - nifi-gateway\n  hosts:\n  - nifi.my-domain.com\n  http:\n  - match:\n    - uri:\n        prefix: /\n    route:\n    - destination:\n        host: <service-name>.<namespace>.svc.cluster.local\n        port:\n          number: 8443\n")),(0,r.kt)("p",null,"Please note that the service name configured as destination of the ",(0,r.kt)("inlineCode",{parentName:"p"},"VirtualService")," is the name of the Service created with the following section in the cluster Deployment YAML"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml"},'spec:  \n  externalServices:  \n    - name: "nifi-cluster"\n      spec:\n        type: ClusterIP\n        portConfigs:\n          - port: 8443\n            internalListenerName: "https"\n')),(0,r.kt)("p",null,"Finally the destination rule that redirect all HTTP traffic destinated to the ",(0,r.kt)("inlineCode",{parentName:"p"},"ClusterIP")," Service to HTTPS encrypting it."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml"},"apiVersion: networking.istio.io/v1beta1\nkind: DestinationRule\nmetadata:\n  name: nifi-dr\nspec:\n  host: <service-name>.<namespace>.svc.cluster.local\n  trafficPolicy:\n    tls:\n      mode: SIMPLE\n    loadBalancer:\n      consistentHash:\n        httpCookie:\n          name: __Secure-Authorization-Bearer\n          ttl: 0s\n")),(0,r.kt)("p",null,"As you can see in the previous example, the destination rule also define how manage the sticky session based on httpCookie property called __Secure-Authorization-Bearer."))}d.isMDXComponent=!0}}]);