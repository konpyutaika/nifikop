(window.webpackJsonp=window.webpackJsonp||[]).push([[126],{190:function(e,n,t){"use strict";t.r(n),t.d(n,"frontMatter",(function(){return o})),t.d(n,"metadata",(function(){return c})),t.d(n,"rightToc",(function(){return l})),t.d(n,"default",(function(){return s}));var r=t(2),a=t(6),i=(t(0),t(521)),o={id:"4_node",title:"Node",sidebar_label:"Node"},c={unversionedId:"5_references/1_nifi_cluster/4_node",id:"version-v0.10.0/5_references/1_nifi_cluster/4_node",isDocsHomePage:!1,title:"Node",description:"Node defines the nifi node basic configuration",source:"@site/versioned_docs/version-v0.10.0/5_references/1_nifi_cluster/4_node.md",slug:"/5_references/1_nifi_cluster/4_node",permalink:"/nifikop/docs/5_references/1_nifi_cluster/4_node",editUrl:"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v0.10.0/5_references/1_nifi_cluster/4_node.md",version:"v0.10.0",lastUpdatedBy:"Juldrixx",lastUpdatedAt:1648141996,sidebar_label:"Node",sidebar:"version-v0.10.0/docs",previous:{title:"Node configuration",permalink:"/nifikop/docs/5_references/1_nifi_cluster/3_node_config"},next:{title:"Node state",permalink:"/nifikop/docs/5_references/1_nifi_cluster/5_node_state"}},l=[{value:"Node",id:"node",children:[]}],d={rightToc:l};function s(e){var n=e.components,t=Object(a.a)(e,["components"]);return Object(i.b)("wrapper",Object(r.a)({},d,t,{components:n,mdxType:"MDXLayout"}),Object(i.b)("p",null,"Node defines the nifi node basic configuration"),Object(i.b)("pre",null,Object(i.b)("code",Object(r.a)({parentName:"pre"},{className:"language-yaml"}),'    - id: 0\n      # nodeConfigGroup can be used to ease the node configuration, if set only the id is required\n      nodeConfigGroup: "default_group"\n      # readOnlyConfig can be used to pass Nifi node config\n      # which has type read-only these config changes will trigger rolling upgrade\n      readOnlyConfig:\n        nifiProperties:\n          overrideConfigs: |\n            nifi.ui.banner.text=NiFiKop - Node 0\n      # node configuration\n#       nodeConfig:\n    - id: 2\n      # readOnlyConfig can be used to pass Nifi node config\n      # which has type read-only these config changes will trigger rolling upgrade\n      readOnlyConfig:\n        overrideConfigs: |\n          nifi.ui.banner.text=NiFiKop - Node 2\n      # node configuration\n      nodeConfig:\n        resourcesRequirements:\n          limits:\n            cpu: "2"\n            memory: 3Gi\n          requests:\n            cpu: "1"\n            memory: 1Gi\n        storageConfigs:\n          # Name of the storage config, used to name PV to reuse into sidecars for example.\n          - name: provenance-repository\n            # Path where the volume will be mount into the main nifi container inside the pod.\n            mountPath: "/opt/nifi/provenance_repository"\n            # Kubernetes PVC spec\n            # https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#create-a-persistentvolumeclaim\n            pvcSpec:\n              accessModes:\n                - ReadWriteOnce\n              storageClassName: "standard"\n              resources:\n                requests:\n                  storage: 8Gi\n')),Object(i.b)("h2",{id:"node"},"Node"),Object(i.b)("table",null,Object(i.b)("thead",{parentName:"table"},Object(i.b)("tr",{parentName:"thead"},Object(i.b)("th",Object(r.a)({parentName:"tr"},{align:null}),"Field"),Object(i.b)("th",Object(r.a)({parentName:"tr"},{align:null}),"Type"),Object(i.b)("th",Object(r.a)({parentName:"tr"},{align:null}),"Description"),Object(i.b)("th",Object(r.a)({parentName:"tr"},{align:null}),"Required"),Object(i.b)("th",Object(r.a)({parentName:"tr"},{align:null}),"Default"))),Object(i.b)("tbody",{parentName:"table"},Object(i.b)("tr",{parentName:"tbody"},Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"id"),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"int32"),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"unique Node id."),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"Yes"),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"-")),Object(i.b)("tr",{parentName:"tbody"},Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"nodeConfigGroup"),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"string"),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"can be used to ease the node configuration, if set only the id is required"),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"No"),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),'""')),Object(i.b)("tr",{parentName:"tbody"},Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"readOnlyConfig"),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),Object(i.b)("a",Object(r.a)({parentName:"td"},{href:"/nifikop/docs/5_references/1_nifi_cluster/2_read_only_config"}),"ReadOnlyConfig")),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"readOnlyConfig can be used to pass Nifi node config which has type read-only these config changes will trigger rolling upgrade."),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"No"),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"nil")),Object(i.b)("tr",{parentName:"tbody"},Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"nodeConfig"),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),Object(i.b)("a",Object(r.a)({parentName:"td"},{href:"/nifikop/docs/5_references/1_nifi_cluster/3_node_config"}),"NodeConfig")),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"node configuration."),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"No"),Object(i.b)("td",Object(r.a)({parentName:"tr"},{align:null}),"nil")))))}s.isMDXComponent=!0},521:function(e,n,t){"use strict";t.d(n,"a",(function(){return b})),t.d(n,"b",(function(){return f}));var r=t(0),a=t.n(r);function i(e,n,t){return n in e?Object.defineProperty(e,n,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[n]=t,e}function o(e,n){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);n&&(r=r.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),t.push.apply(t,r)}return t}function c(e){for(var n=1;n<arguments.length;n++){var t=null!=arguments[n]?arguments[n]:{};n%2?o(Object(t),!0).forEach((function(n){i(e,n,t[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):o(Object(t)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(t,n))}))}return e}function l(e,n){if(null==e)return{};var t,r,a=function(e,n){if(null==e)return{};var t,r,a={},i=Object.keys(e);for(r=0;r<i.length;r++)t=i[r],n.indexOf(t)>=0||(a[t]=e[t]);return a}(e,n);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);for(r=0;r<i.length;r++)t=i[r],n.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(a[t]=e[t])}return a}var d=a.a.createContext({}),s=function(e){var n=a.a.useContext(d),t=n;return e&&(t="function"==typeof e?e(n):c(c({},n),e)),t},b=function(e){var n=s(e.components);return a.a.createElement(d.Provider,{value:n},e.children)},u={inlineCode:"code",wrapper:function(e){var n=e.children;return a.a.createElement(a.a.Fragment,{},n)}},p=a.a.forwardRef((function(e,n){var t=e.components,r=e.mdxType,i=e.originalType,o=e.parentName,d=l(e,["components","mdxType","originalType","parentName"]),b=s(t),p=r,f=b["".concat(o,".").concat(p)]||b[p]||u[p]||i;return t?a.a.createElement(f,c(c({ref:n},d),{},{components:t})):a.a.createElement(f,c({ref:n},d))}));function f(e,n){var t=arguments,r=n&&n.mdxType;if("string"==typeof e||r){var i=t.length,o=new Array(i);o[0]=p;var c={};for(var l in n)hasOwnProperty.call(n,l)&&(c[l]=n[l]);c.originalType=e,c.mdxType="string"==typeof e?e:r,o[1]=c;for(var d=2;d<i;d++)o[d]=t[d];return a.a.createElement.apply(null,o)}return a.a.createElement.apply(null,t)}p.displayName="MDXCreateElement"}}]);