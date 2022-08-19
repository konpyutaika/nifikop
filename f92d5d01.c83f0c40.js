(window.webpackJsonp=window.webpackJsonp||[]).push([[514],{581:function(e,n,t){"use strict";t.r(n),t.d(n,"frontMatter",(function(){return s})),t.d(n,"metadata",(function(){return c})),t.d(n,"rightToc",(function(){return l})),t.d(n,"default",(function(){return u}));var r=t(2),o=t(6),a=(t(0),t(599)),i=["components"],s={id:"1_v0.7.x_to_v0.8.0",title:"v0.7.x to v0.8.0",sidebar_label:"v0.7.x to v0.8.0"},c={unversionedId:"7_upgrade/1_v0.7.x_to_v0.8.0",id:"version-v0.11.0/7_upgrade/1_v0.7.x_to_v0.8.0",isDocsHomePage:!1,title:"v0.7.x to v0.8.0",description:"Guide to migrate operator resources built using nifi.orange.com/v1alpha1 to nifi.konpyutaika/v1alpha1.",source:"@site/versioned_docs/version-v0.11.0/7_upgrade/1_v0.7.x_to_v0.8.0.md",slug:"/7_upgrade/1_v0.7.x_to_v0.8.0",permalink:"/nifikop/docs/v0.11.0/7_upgrade/1_v0.7.x_to_v0.8.0",editUrl:"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v0.11.0/7_upgrade/1_v0.7.x_to_v0.8.0.md",version:"v0.11.0",lastUpdatedBy:"Juldrixx",lastUpdatedAt:1656349755,sidebar_label:"v0.7.x to v0.8.0",sidebar:"version-v0.11.0/docs",previous:{title:"Credits",permalink:"/nifikop/docs/v0.11.0/6_contributing/3_credits"}},l=[{value:"Getting started",id:"getting-started",children:[]},{value:"Prerequisites",id:"prerequisites",children:[]},{value:"Initial setup",id:"initial-setup",children:[]},{value:"Script setup",id:"script-setup",children:[]},{value:"Run script",id:"run-script",children:[]}],p={rightToc:l};function u(e){var n=e.components,t=Object(o.a)(e,i);return Object(a.b)("wrapper",Object(r.a)({},p,t,{components:n,mdxType:"MDXLayout"}),Object(a.b)("p",null,"Guide to migrate operator resources built using ",Object(a.b)("inlineCode",{parentName:"p"},"nifi.orange.com/v1alpha1")," to ",Object(a.b)("inlineCode",{parentName:"p"},"nifi.konpyutaika/v1alpha1"),"."),Object(a.b)("h2",{id:"getting-started"},"Getting started"),Object(a.b)("p",null,"The goal is to migrate your NiFiKop resources from the old CRDs to the new ones without any service interruption."),Object(a.b)("p",null,"To do this, it is necessary to have both versions of CRDs available on Kubernetes and to have the old operator stopped (to prevent any manipulation on the resources).\nThen launch the script developed in nodejs presented in the following. The script will copy the resources in the old CRDs to the new CRDs keeping only the relevant fields (labels, annotations, name and spec) and then copy the status."),Object(a.b)("h2",{id:"prerequisites"},"Prerequisites"),Object(a.b)("ul",null,Object(a.b)("li",{parentName:"ul"},Object(a.b)("a",{parentName:"li",href:"https://nodejs.org/en/download/"},"nodejs")," version 15.3.0+"),Object(a.b)("li",{parentName:"ul"},Object(a.b)("a",{parentName:"li",href:"https://docs.npmjs.com/cli/v7/configuring-npm/install"},"npm")," version 7.0.14+")),Object(a.b)("h2",{id:"initial-setup"},"Initial setup"),Object(a.b)("p",null,"Create a nodejs project and download the required dependencies:"),Object(a.b)("pre",null,Object(a.b)("code",{parentName:"pre",className:"language-bash"},"npm init -y\nnpm install @kubernetes/client-node@0.16.3 minimist@1.2.6\n")),Object(a.b)("p",null,"In ",Object(a.b)("inlineCode",{parentName:"p"},"package.json")," add the following script:"),Object(a.b)("pre",null,Object(a.b)("code",{parentName:"pre",className:"language-json"},'"start": "node --no-warnings index.js"\n')),Object(a.b)("p",null,"Your ",Object(a.b)("inlineCode",{parentName:"p"},"package.json")," should look like that:"),Object(a.b)("pre",null,Object(a.b)("code",{parentName:"pre",className:"language-json"},'{\n  "name": "nifikop_crd_migration",\n  "version": "1.0.0",\n  "description": "Script to migrate from the old CRDs to the new CRDs.",\n  "main": "index.js",\n  "scripts": {\n    "start": "node --no-warnings index.js",\n    "test": "echo \\"Error: no test specified\\" && exit 1"\n  },\n  "keywords": [\n    "K8S",\n    "NiFiKop",\n    "CRDs"\n  ],\n  "license": "ISC",\n  "dependencies": {\n    "@kubernetes/client-node": "^0.16.3",\n    "minimist": "^1.2.6"\n  }\n}\n')),Object(a.b)("h2",{id:"script-setup"},"Script setup"),Object(a.b)("p",null,"Create the file ",Object(a.b)("inlineCode",{parentName:"p"},"index.js")," with the following content:"),Object(a.b)("pre",null,Object(a.b)("code",{parentName:"pre",className:"language-js"},"process.env['NODE_TLS_REJECT_UNAUTHORIZED'] = 0;\nconst k8s = require('@kubernetes/client-node');\n\nconst kc = new k8s.KubeConfig();\nkc.loadFromDefault();\n\nconst k8sApi = kc.makeApiClient(k8s.CustomObjectsApi);\n\nconst KONPYUTAIKA_GROUP = 'nifi.konpyutaika.com';\nconst KONPYUTAIKA_GROUP_VERSION = 'v1alpha1';\nconst ORANGE_GROUP = 'nifi.orange.com';\nconst ORANGE_GROUP_VERSION = 'v1alpha1';\n\nconst call = async (SRC_GRP, SRC_GRP_VER, DST_GRP, DST_GRP_VER, KIND_PLURAL, NAMESPACE) => {\n    console.log(`Listing ${KIND_PLURAL} of ${SRC_GRP}/${SRC_GRP_VER} in ${NAMESPACE}...`);\n    const listResources = (await k8sApi.listNamespacedCustomObject(SRC_GRP, SRC_GRP_VER, NAMESPACE, KIND_PLURAL)).body.items;\n    return Promise.all(listResources.map(async (resource) => {\n        try {\n            console.log(`Found ${resource.kind} \"${resource.metadata.name}\" of ${resource.apiVersion} in ${NAMESPACE}`);\n\n            if (resource.metadata.ownerReferences) {\n                console.log(`${resource.kind} ${resource.metadata.name} mananged by something else (ownerRefereces is set).`);\n                return;\n            }\n\n            const bodyResource = {\n                apiVersion: `${DST_GRP}/${DST_GRP_VER}`,\n                kind: resource.kind,\n                metadata: {\n                    name: resource.metadata.name,\n                    annotations: resource.metadata.annotations,\n                    labels: resource.metadata.labels\n                },\n                spec: resource.spec\n            };\n\n            console.log(`Creating ${bodyResource.kind} \"${bodyResource.metadata.name}\" of ${bodyResource.apiVersion} in ${NAMESPACE}...`);\n            const newResource = (await k8sApi.createNamespacedCustomObject(DST_GRP, DST_GRP_VER, NAMESPACE, KIND_PLURAL, bodyResource)).body;\n            console.log('...done creating.');\n\n            const bodyStatus = {\n                apiVersion: newResource.apiVersion,\n                kind: newResource.kind,\n                metadata: {\n                    name: newResource.metadata.name,\n                    resourceVersion: newResource.metadata.resourceVersion\n                },\n                status: resource.status\n            };\n\n            console.log(`Copying status from ${resource.kind} \"${resource.metadata.name}\" of ${newResource.apiVersion} to ${newResource.kind} \"${newResource.metadata.name}\" of ${newResource.apiVersion} in ${NAMESPACE}...`);\n            const newResourceWithStatus = (await k8sApi.replaceNamespacedCustomObjectStatus(DST_GRP, DST_GRP_VER, NAMESPACE, KIND_PLURAL, bodyStatus.metadata.name, bodyStatus)).body;\n            console.log('...done copying.');\n            return newResourceWithStatus;\n        }\n        catch (e) {\n            console.error(e.body ? e.body.message ? e.body.message : e.body : e);\n        }\n    }));\n};\n\nconst argv = require('minimist')(process.argv.slice(2));\n\nlet NAMESPACE = argv.namespace ? argv.namespace.length > 0 ? argv.namespace : 'default' : 'default';\nlet KIND_PLURAL = {\n    cluster: 'nificlusters',\n    dataflow: 'nifidataflows',\n    parametercontext: 'nifiparametercontexts',\n    registryclient: 'nifiregistryclients',\n    user: 'nifiusers',\n    usergroup: 'nifiusergroups',\n};\n\nif (!argv.type) {\n    console.error('Type not provided');\n    process.exit(1);\n}\n\nif (!KIND_PLURAL[argv.type]) {\n    console.error(`Type ${argv.type} is not one of the following types: ${Object.keys(KIND_PLURAL)}`);\n    process.exit(1);\n}\n\nconsole.log(`########### START: ${KIND_PLURAL[argv.type]} ###########`);\ncall( ORANGE_GROUP, ORANGE_GROUP_VERSION, KONPYUTAIKA_GROUP, KONPYUTAIKA_GROUP_VERSION, KIND_PLURAL[argv.type], NAMESPACE)\n    .then(r => console.log('############ END ############'))\n    .catch(e => console.error(e));\n")),Object(a.b)("h2",{id:"run-script"},"Run script"),Object(a.b)("p",null,"To migrate the resources, run the following command:"),Object(a.b)("pre",null,Object(a.b)("code",{parentName:"pre",className:"language-bash"},"npm start -- --type=<NIFIKOP_RESOURCE> --namespace=<K8S_NAMESPACE>\n")),Object(a.b)("p",null,"with"),Object(a.b)("ul",null,Object(a.b)("li",{parentName:"ul"},Object(a.b)("inlineCode",{parentName:"li"},"<NIFIKOP_RESOURCE>"),": NiFiKop resource type (cluster, dataflow, user, usergroup, parametercontext or registryclient)"),Object(a.b)("li",{parentName:"ul"},Object(a.b)("inlineCode",{parentName:"li"},"<K8S_NAMESPACE>:")," Kubernetes namespace where the resources will be migrated")))}u.isMDXComponent=!0},599:function(e,n,t){"use strict";t.d(n,"a",(function(){return u})),t.d(n,"b",(function(){return m}));var r=t(0),o=t.n(r);function a(e,n,t){return n in e?Object.defineProperty(e,n,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[n]=t,e}function i(e,n){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);n&&(r=r.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),t.push.apply(t,r)}return t}function s(e){for(var n=1;n<arguments.length;n++){var t=null!=arguments[n]?arguments[n]:{};n%2?i(Object(t),!0).forEach((function(n){a(e,n,t[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):i(Object(t)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(t,n))}))}return e}function c(e,n){if(null==e)return{};var t,r,o=function(e,n){if(null==e)return{};var t,r,o={},a=Object.keys(e);for(r=0;r<a.length;r++)t=a[r],n.indexOf(t)>=0||(o[t]=e[t]);return o}(e,n);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(r=0;r<a.length;r++)t=a[r],n.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(o[t]=e[t])}return o}var l=o.a.createContext({}),p=function(e){var n=o.a.useContext(l),t=n;return e&&(t="function"==typeof e?e(n):s(s({},n),e)),t},u=function(e){var n=p(e.components);return o.a.createElement(l.Provider,{value:n},e.children)},d={inlineCode:"code",wrapper:function(e){var n=e.children;return o.a.createElement(o.a.Fragment,{},n)}},b=o.a.forwardRef((function(e,n){var t=e.components,r=e.mdxType,a=e.originalType,i=e.parentName,l=c(e,["components","mdxType","originalType","parentName"]),u=p(t),b=r,m=u["".concat(i,".").concat(b)]||u[b]||d[b]||a;return t?o.a.createElement(m,s(s({ref:n},l),{},{components:t})):o.a.createElement(m,s({ref:n},l))}));function m(e,n){var t=arguments,r=n&&n.mdxType;if("string"==typeof e||r){var a=t.length,i=new Array(a);i[0]=b;var s={};for(var c in n)hasOwnProperty.call(n,c)&&(s[c]=n[c]);s.originalType=e,s.mdxType="string"==typeof e?e:r,i[1]=s;for(var l=2;l<a;l++)i[l]=t[l];return o.a.createElement.apply(null,i)}return o.a.createElement.apply(null,t)}b.displayName="MDXCreateElement"}}]);