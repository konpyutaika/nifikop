"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[24397],{35318:(e,n,t)=>{t.d(n,{Zo:()=>p,kt:()=>m});var r=t(27378);function o(e,n,t){return n in e?Object.defineProperty(e,n,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[n]=t,e}function a(e,n){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);n&&(r=r.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),t.push.apply(t,r)}return t}function i(e){for(var n=1;n<arguments.length;n++){var t=null!=arguments[n]?arguments[n]:{};n%2?a(Object(t),!0).forEach((function(n){o(e,n,t[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):a(Object(t)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(t,n))}))}return e}function s(e,n){if(null==e)return{};var t,r,o=function(e,n){if(null==e)return{};var t,r,o={},a=Object.keys(e);for(r=0;r<a.length;r++)t=a[r],n.indexOf(t)>=0||(o[t]=e[t]);return o}(e,n);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(r=0;r<a.length;r++)t=a[r],n.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(o[t]=e[t])}return o}var l=r.createContext({}),c=function(e){var n=r.useContext(l),t=n;return e&&(t="function"==typeof e?e(n):i(i({},n),e)),t},p=function(e){var n=c(e.components);return r.createElement(l.Provider,{value:n},e.children)},u={inlineCode:"code",wrapper:function(e){var n=e.children;return r.createElement(r.Fragment,{},n)}},d=r.forwardRef((function(e,n){var t=e.components,o=e.mdxType,a=e.originalType,l=e.parentName,p=s(e,["components","mdxType","originalType","parentName"]),d=c(t),m=o,g=d["".concat(l,".").concat(m)]||d[m]||u[m]||a;return t?r.createElement(g,i(i({ref:n},p),{},{components:t})):r.createElement(g,i({ref:n},p))}));function m(e,n){var t=arguments,o=n&&n.mdxType;if("string"==typeof e||o){var a=t.length,i=new Array(a);i[0]=d;var s={};for(var l in n)hasOwnProperty.call(n,l)&&(s[l]=n[l]);s.originalType=e,s.mdxType="string"==typeof e?e:o,i[1]=s;for(var c=2;c<a;c++)i[c]=t[c];return r.createElement.apply(null,i)}return r.createElement.apply(null,t)}d.displayName="MDXCreateElement"},61171:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>p,contentTitle:()=>l,default:()=>m,frontMatter:()=>s,metadata:()=>c,toc:()=>u});var r=t(25773),o=t(30808),a=(t(27378),t(35318)),i=["components"],s={id:"1_v0.7.x_to_v0.8.0",title:"v0.7.x to v0.8.0",sidebar_label:"v0.7.x to v0.8.0"},l=void 0,c={unversionedId:"7_upgrade_guides/1_v0.7.x_to_v0.8.0",id:"version-v1.3.0/7_upgrade_guides/1_v0.7.x_to_v0.8.0",title:"v0.7.x to v0.8.0",description:"Guide to migrate operator resources built using nifi.orange.com/v1alpha1 to nifi.konpyutaika/v1alpha1.",source:"@site/versioned_docs/version-v1.3.0/7_upgrade_guides/1_v0.7.x_to_v0.8.0.md",sourceDirName:"7_upgrade_guides",slug:"/7_upgrade_guides/1_v0.7.x_to_v0.8.0",permalink:"/nifikop/docs/v1.3.0/7_upgrade_guides/1_v0.7.x_to_v0.8.0",draft:!1,editUrl:"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.3.0/7_upgrade_guides/1_v0.7.x_to_v0.8.0.md",tags:[],version:"v1.3.0",lastUpdatedBy:"Juldrixx",lastUpdatedAt:1692604575,formattedLastUpdatedAt:"Aug 21, 2023",frontMatter:{id:"1_v0.7.x_to_v0.8.0",title:"v0.7.x to v0.8.0",sidebar_label:"v0.7.x to v0.8.0"},sidebar:"docs",previous:{title:"Credits",permalink:"/nifikop/docs/v1.3.0/6_contributing/3_credits"},next:{title:"v0.14.1 to v0.15.0",permalink:"/nifikop/docs/v1.3.0/7_upgrade_guides/2_v0.14.1_to_v0.15.0"}},p={},u=[{value:"Getting started",id:"getting-started",level:2},{value:"Prerequisites",id:"prerequisites",level:2},{value:"Initial setup",id:"initial-setup",level:2},{value:"Script setup",id:"script-setup",level:2},{value:"Run script",id:"run-script",level:2}],d={toc:u};function m(e){var n=e.components,t=(0,o.Z)(e,i);return(0,a.kt)("wrapper",(0,r.Z)({},d,t,{components:n,mdxType:"MDXLayout"}),(0,a.kt)("p",null,"Guide to migrate operator resources built using ",(0,a.kt)("inlineCode",{parentName:"p"},"nifi.orange.com/v1alpha1")," to ",(0,a.kt)("inlineCode",{parentName:"p"},"nifi.konpyutaika/v1alpha1"),"."),(0,a.kt)("h2",{id:"getting-started"},"Getting started"),(0,a.kt)("p",null,"The goal is to migrate your NiFiKop resources from the old CRDs to the new ones without any service interruption."),(0,a.kt)("p",null,"To do this, it is necessary to have both versions of CRDs available on Kubernetes and to have the old operator stopped (to prevent any manipulation on the resources).\nThen launch the script developed in nodejs presented in the following. The script will copy the resources in the old CRDs to the new CRDs keeping only the relevant fields (labels, annotations, name and spec) and then copy the status."),(0,a.kt)("h2",{id:"prerequisites"},"Prerequisites"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"https://nodejs.org/en/download/"},"nodejs")," version 15.3.0+"),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"https://docs.npmjs.com/cli/v7/configuring-npm/install"},"npm")," version 7.0.14+")),(0,a.kt)("h2",{id:"initial-setup"},"Initial setup"),(0,a.kt)("p",null,"Create a nodejs project and download the required dependencies:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},"npm init -y\nnpm install @kubernetes/client-node@0.16.3 minimist@1.2.6\n")),(0,a.kt)("p",null,"In ",(0,a.kt)("inlineCode",{parentName:"p"},"package.json")," add the following script:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-json"},'"start": "node --no-warnings index.js"\n')),(0,a.kt)("p",null,"Your ",(0,a.kt)("inlineCode",{parentName:"p"},"package.json")," should look like that:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-json"},'{\n  "name": "nifikop_crd_migration",\n  "version": "1.0.0",\n  "description": "Script to migrate from the old CRDs to the new CRDs.",\n  "main": "index.js",\n  "scripts": {\n    "start": "node --no-warnings index.js",\n    "test": "echo \\"Error: no test specified\\" && exit 1"\n  },\n  "keywords": [\n    "K8S",\n    "NiFiKop",\n    "CRDs"\n  ],\n  "license": "ISC",\n  "dependencies": {\n    "@kubernetes/client-node": "^0.16.3",\n    "minimist": "^1.2.6"\n  }\n}\n')),(0,a.kt)("h2",{id:"script-setup"},"Script setup"),(0,a.kt)("p",null,"Create the file ",(0,a.kt)("inlineCode",{parentName:"p"},"index.js")," with the following content:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-js"},"process.env['NODE_TLS_REJECT_UNAUTHORIZED'] = 0;\nconst k8s = require('@kubernetes/client-node');\n\nconst kc = new k8s.KubeConfig();\nkc.loadFromDefault();\n\nconst k8sApi = kc.makeApiClient(k8s.CustomObjectsApi);\n\nconst KONPYUTAIKA_GROUP = 'nifi.konpyutaika.com';\nconst KONPYUTAIKA_GROUP_VERSION = 'v1alpha1';\nconst ORANGE_GROUP = 'nifi.orange.com';\nconst ORANGE_GROUP_VERSION = 'v1alpha1';\n\nconst call = async (SRC_GRP, SRC_GRP_VER, DST_GRP, DST_GRP_VER, KIND_PLURAL, NAMESPACE) => {\n    console.log(`Listing ${KIND_PLURAL} of ${SRC_GRP}/${SRC_GRP_VER} in ${NAMESPACE}...`);\n    const listResources = (await k8sApi.listNamespacedCustomObject(SRC_GRP, SRC_GRP_VER, NAMESPACE, KIND_PLURAL)).body.items;\n    return Promise.all(listResources.map(async (resource) => {\n        try {\n            console.log(`Found ${resource.kind} \"${resource.metadata.name}\" of ${resource.apiVersion} in ${NAMESPACE}`);\n\n            if (resource.metadata.ownerReferences) {\n                console.log(`${resource.kind} ${resource.metadata.name} mananged by something else (ownerRefereces is set).`);\n                return;\n            }\n\n            const bodyResource = {\n                apiVersion: `${DST_GRP}/${DST_GRP_VER}`,\n                kind: resource.kind,\n                metadata: {\n                    name: resource.metadata.name,\n                    annotations: resource.metadata.annotations,\n                    labels: resource.metadata.labels\n                },\n                spec: resource.spec\n            };\n\n            console.log(`Creating ${bodyResource.kind} \"${bodyResource.metadata.name}\" of ${bodyResource.apiVersion} in ${NAMESPACE}...`);\n            const newResource = (await k8sApi.createNamespacedCustomObject(DST_GRP, DST_GRP_VER, NAMESPACE, KIND_PLURAL, bodyResource)).body;\n            console.log('...done creating.');\n\n            const bodyStatus = {\n                apiVersion: newResource.apiVersion,\n                kind: newResource.kind,\n                metadata: {\n                    name: newResource.metadata.name,\n                    resourceVersion: newResource.metadata.resourceVersion\n                },\n                status: resource.status\n            };\n\n            console.log(`Copying status from ${resource.kind} \"${resource.metadata.name}\" of ${newResource.apiVersion} to ${newResource.kind} \"${newResource.metadata.name}\" of ${newResource.apiVersion} in ${NAMESPACE}...`);\n            const newResourceWithStatus = (await k8sApi.replaceNamespacedCustomObjectStatus(DST_GRP, DST_GRP_VER, NAMESPACE, KIND_PLURAL, bodyStatus.metadata.name, bodyStatus)).body;\n            console.log('...done copying.');\n            return newResourceWithStatus;\n        }\n        catch (e) {\n            console.error(e.body ? e.body.message ? e.body.message : e.body : e);\n        }\n    }));\n};\n\nconst argv = require('minimist')(process.argv.slice(2));\n\nlet NAMESPACE = argv.namespace ? argv.namespace.length > 0 ? argv.namespace : 'default' : 'default';\nlet KIND_PLURAL = {\n    cluster: 'nificlusters',\n    dataflow: 'nifidataflows',\n    parametercontext: 'nifiparametercontexts',\n    registryclient: 'nifiregistryclients',\n    user: 'nifiusers',\n    usergroup: 'nifiusergroups',\n};\n\nif (!argv.type) {\n    console.error('Type not provided');\n    process.exit(1);\n}\n\nif (!KIND_PLURAL[argv.type]) {\n    console.error(`Type ${argv.type} is not one of the following types: ${Object.keys(KIND_PLURAL)}`);\n    process.exit(1);\n}\n\nconsole.log(`########### START: ${KIND_PLURAL[argv.type]} ###########`);\ncall( ORANGE_GROUP, ORANGE_GROUP_VERSION, KONPYUTAIKA_GROUP, KONPYUTAIKA_GROUP_VERSION, KIND_PLURAL[argv.type], NAMESPACE)\n    .then(r => console.log('############ END ############'))\n    .catch(e => console.error(e));\n")),(0,a.kt)("h2",{id:"run-script"},"Run script"),(0,a.kt)("p",null,"To migrate the resources, run the following command:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},"npm start -- --type=<NIFIKOP_RESOURCE> --namespace=<K8S_NAMESPACE>\n")),(0,a.kt)("p",null,"with"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("inlineCode",{parentName:"li"},"<NIFIKOP_RESOURCE>"),": NiFiKop resource type (cluster, dataflow, user, usergroup, parametercontext or registryclient)"),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("inlineCode",{parentName:"li"},"<K8S_NAMESPACE>:")," Kubernetes namespace where the resources will be migrated")))}m.isMDXComponent=!0}}]);