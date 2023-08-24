"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[34978],{35318:(e,t,a)=>{a.d(t,{Zo:()=>p,kt:()=>c});var n=a(27378);function r(e,t,a){return t in e?Object.defineProperty(e,t,{value:a,enumerable:!0,configurable:!0,writable:!0}):e[t]=a,e}function l(e,t){var a=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),a.push.apply(a,n)}return a}function i(e){for(var t=1;t<arguments.length;t++){var a=null!=arguments[t]?arguments[t]:{};t%2?l(Object(a),!0).forEach((function(t){r(e,t,a[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(a)):l(Object(a)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(a,t))}))}return e}function o(e,t){if(null==e)return{};var a,n,r=function(e,t){if(null==e)return{};var a,n,r={},l=Object.keys(e);for(n=0;n<l.length;n++)a=l[n],t.indexOf(a)>=0||(r[a]=e[a]);return r}(e,t);if(Object.getOwnPropertySymbols){var l=Object.getOwnPropertySymbols(e);for(n=0;n<l.length;n++)a=l[n],t.indexOf(a)>=0||Object.prototype.propertyIsEnumerable.call(e,a)&&(r[a]=e[a])}return r}var u=n.createContext({}),s=function(e){var t=n.useContext(u),a=t;return e&&(a="function"==typeof e?e(t):i(i({},t),e)),a},p=function(e){var t=s(e.components);return n.createElement(u.Provider,{value:t},e.children)},d={inlineCode:"code",wrapper:function(e){var t=e.children;return n.createElement(n.Fragment,{},t)}},m=n.forwardRef((function(e,t){var a=e.components,r=e.mdxType,l=e.originalType,u=e.parentName,p=o(e,["components","mdxType","originalType","parentName"]),m=s(a),c=r,k=m["".concat(u,".").concat(c)]||m[c]||d[c]||l;return a?n.createElement(k,i(i({ref:t},p),{},{components:a})):n.createElement(k,i({ref:t},p))}));function c(e,t){var a=arguments,r=t&&t.mdxType;if("string"==typeof e||r){var l=a.length,i=new Array(l);i[0]=m;var o={};for(var u in t)hasOwnProperty.call(t,u)&&(o[u]=t[u]);o.originalType=e,o.mdxType="string"==typeof e?e:r,i[1]=o;for(var s=2;s<l;s++)i[s]=a[s];return n.createElement.apply(null,i)}return n.createElement.apply(null,a)}m.displayName="MDXCreateElement"},39798:(e,t,a)=>{a.d(t,{Z:()=>i});var n=a(27378),r=a(38944);const l="tabItem_wHwb";function i(e){var t=e.children,a=e.hidden,i=e.className;return n.createElement("div",{role:"tabpanel",className:(0,r.Z)(l,i),hidden:a},t)}},23930:(e,t,a)=>{a.d(t,{Z:()=>C});var n=a(25773),r=a(27378),l=a(38944),i=a(83457),o=a(35331),u=a(30654),s=a(70784),p=a(71819);function d(e){return function(e){var t,a;return null!=(t=null==(a=r.Children.map(e,(function(e){if(!e||(0,r.isValidElement)(e)&&(t=e.props)&&"object"==typeof t&&"value"in t)return e;var t;throw new Error("Docusaurus error: Bad <Tabs> child <"+("string"==typeof e.type?e.type:e.type.name)+'>: all children of the <Tabs> component should be <TabItem>, and every <TabItem> should have a unique "value" prop.')})))?void 0:a.filter(Boolean))?t:[]}(e).map((function(e){var t=e.props;return{value:t.value,label:t.label,attributes:t.attributes,default:t.default}}))}function m(e){var t=e.values,a=e.children;return(0,r.useMemo)((function(){var e=null!=t?t:d(a);return function(e){var t=(0,s.l)(e,(function(e,t){return e.value===t.value}));if(t.length>0)throw new Error('Docusaurus error: Duplicate values "'+t.map((function(e){return e.value})).join(", ")+'" found in <Tabs>. Every value needs to be unique.')}(e),e}),[t,a])}function c(e){var t=e.value;return e.tabValues.some((function(e){return e.value===t}))}function k(e){var t=e.queryString,a=void 0!==t&&t,n=e.groupId,l=(0,o.k6)(),i=function(e){var t=e.queryString,a=void 0!==t&&t,n=e.groupId;if("string"==typeof a)return a;if(!1===a)return null;if(!0===a&&!n)throw new Error('Docusaurus error: The <Tabs> component groupId prop is required if queryString=true, because this value is used as the search param name. You can also provide an explicit value such as queryString="my-search-param".');return null!=n?n:null}({queryString:a,groupId:n});return[(0,u._X)(i),(0,r.useCallback)((function(e){if(i){var t=new URLSearchParams(l.location.search);t.set(i,e),l.replace(Object.assign({},l.location,{search:t.toString()}))}}),[i,l])]}function f(e){var t,a,n,l,i=e.defaultValue,o=e.queryString,u=void 0!==o&&o,s=e.groupId,d=m(e),f=(0,r.useState)((function(){return function(e){var t,a=e.defaultValue,n=e.tabValues;if(0===n.length)throw new Error("Docusaurus error: the <Tabs> component requires at least one <TabItem> children component");if(a){if(!c({value:a,tabValues:n}))throw new Error('Docusaurus error: The <Tabs> has a defaultValue "'+a+'" but none of its children has the corresponding value. Available values are: '+n.map((function(e){return e.value})).join(", ")+". If you intend to show no default tab, use defaultValue={null} instead.");return a}var r=null!=(t=n.find((function(e){return e.default})))?t:n[0];if(!r)throw new Error("Unexpected error: 0 tabValues");return r.value}({defaultValue:i,tabValues:d})})),h=f[0],g=f[1],N=k({queryString:u,groupId:s}),b=N[0],y=N[1],v=(t=function(e){return e?"docusaurus.tab."+e:null}({groupId:s}.groupId),a=(0,p.Nk)(t),n=a[0],l=a[1],[n,(0,r.useCallback)((function(e){t&&l.set(e)}),[t,l])]),C=v[0],w=v[1],_=function(){var e=null!=b?b:C;return c({value:e,tabValues:d})?e:null}();return(0,r.useLayoutEffect)((function(){_&&g(_)}),[_]),{selectedValue:h,selectValue:(0,r.useCallback)((function(e){if(!c({value:e,tabValues:d}))throw new Error("Can't select invalid tab value="+e);g(e),y(e),w(e)}),[y,w,d]),tabValues:d}}var h=a(76457);const g="tabList_J5MA",N="tabItem_l0OV";function b(e){var t=e.className,a=e.block,o=e.selectedValue,u=e.selectValue,s=e.tabValues,p=[],d=(0,i.o5)().blockElementScrollPositionUntilNextRender,m=function(e){var t=e.currentTarget,a=p.indexOf(t),n=s[a].value;n!==o&&(d(t),u(n))},c=function(e){var t,a=null;switch(e.key){case"Enter":m(e);break;case"ArrowRight":var n,r=p.indexOf(e.currentTarget)+1;a=null!=(n=p[r])?n:p[0];break;case"ArrowLeft":var l,i=p.indexOf(e.currentTarget)-1;a=null!=(l=p[i])?l:p[p.length-1]}null==(t=a)||t.focus()};return r.createElement("ul",{role:"tablist","aria-orientation":"horizontal",className:(0,l.Z)("tabs",{"tabs--block":a},t)},s.map((function(e){var t=e.value,a=e.label,i=e.attributes;return r.createElement("li",(0,n.Z)({role:"tab",tabIndex:o===t?0:-1,"aria-selected":o===t,key:t,ref:function(e){return p.push(e)},onKeyDown:c,onClick:m},i,{className:(0,l.Z)("tabs__item",N,null==i?void 0:i.className,{"tabs__item--active":o===t})}),null!=a?a:t)})))}function y(e){var t=e.lazy,a=e.children,n=e.selectedValue,l=(Array.isArray(a)?a:[a]).filter(Boolean);if(t){var i=l.find((function(e){return e.props.value===n}));return i?(0,r.cloneElement)(i,{className:"margin-top--md"}):null}return r.createElement("div",{className:"margin-top--md"},l.map((function(e,t){return(0,r.cloneElement)(e,{key:t,hidden:e.props.value!==n})})))}function v(e){var t=f(e);return r.createElement("div",{className:(0,l.Z)("tabs-container",g)},r.createElement(b,(0,n.Z)({},e,t)),r.createElement(y,(0,n.Z)({},e,t)))}function C(e){var t=(0,h.Z)();return r.createElement(v,(0,n.Z)({key:String(t)},e))}},27602:(e,t,a)=>{a.r(t),a.d(t,{assets:()=>m,contentTitle:()=>p,default:()=>f,frontMatter:()=>s,metadata:()=>d,toc:()=>c});var n=a(25773),r=a(30808),l=(a(27378),a(35318)),i=a(23930),o=a(39798),u=["components"],s={id:"2_customizable_install_with_helm",title:"Customizable install with Helm",sidebar_label:"Customizable install with Helm"},p=void 0,d={unversionedId:"2_deploy_nifikop/2_customizable_install_with_helm",id:"version-v1.3.0/2_deploy_nifikop/2_customizable_install_with_helm",title:"Customizable install with Helm",description:"Prerequisites",source:"@site/versioned_docs/version-v1.3.0/2_deploy_nifikop/2_customizable_install_with_helm.md",sourceDirName:"2_deploy_nifikop",slug:"/2_deploy_nifikop/2_customizable_install_with_helm",permalink:"/nifikop/docs/v1.3.0/2_deploy_nifikop/2_customizable_install_with_helm",draft:!1,editUrl:"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.3.0/2_deploy_nifikop/2_customizable_install_with_helm.md",tags:[],version:"v1.3.0",lastUpdatedBy:"Juldrixx",lastUpdatedAt:1692604575,formattedLastUpdatedAt:"Aug 21, 2023",frontMatter:{id:"2_customizable_install_with_helm",title:"Customizable install with Helm",sidebar_label:"Customizable install with Helm"},sidebar:"docs",previous:{title:"Quick start",permalink:"/nifikop/docs/v1.3.0/2_deploy_nifikop/1_quick_start"},next:{title:"Design Principles",permalink:"/nifikop/docs/v1.3.0/3_manage_nifi/1_manage_clusters/0_design_principles"}},m={},c=[{value:"Prerequisites",id:"prerequisites",level:2},{value:"Introduction",id:"introduction",level:2},{value:"Configuration",id:"configuration",level:3},{value:"Installing the Chart",id:"installing-the-chart",level:3},{value:"Listing deployed charts",id:"listing-deployed-charts",level:3},{value:"Get Status for the helm deployment",id:"get-status-for-the-helm-deployment",level:3},{value:"Uninstaling the Charts",id:"uninstaling-the-charts",level:2},{value:"Troubleshooting",id:"troubleshooting",level:2},{value:"Install of the CRD",id:"install-of-the-crd",level:3}],k={toc:c};function f(e){var t=e.components,a=(0,r.Z)(e,u);return(0,l.kt)("wrapper",(0,n.Z)({},k,a,{components:t,mdxType:"MDXLayout"}),(0,l.kt)("h2",{id:"prerequisites"},"Prerequisites"),(0,l.kt)("ul",null,(0,l.kt)("li",{parentName:"ul"},"Perform any necessary ",(0,l.kt)("a",{parentName:"li",href:"./1_quick_start"},"plateform-specific setup")),(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("a",{parentName:"li",href:"https://github.com/helm/helm#install"},"Install a Helm client")," with a version higher than 3")),(0,l.kt)("h2",{id:"introduction"},"Introduction"),(0,l.kt)("p",null,"This Helm chart install NiFiKop the Nifi Kubernetes operator to create/configure/manage NiFi\nclusters in a Kubernetes Namespace."),(0,l.kt)("p",null,"It will use Custom Ressources Definition CRDs:"),(0,l.kt)("ul",null,(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("inlineCode",{parentName:"li"},"nificlusters.nifi.konpyutaika.com"),","),(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("inlineCode",{parentName:"li"},"nifiusers.nifi.konpyutaika.com"),","),(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("inlineCode",{parentName:"li"},"nifiusergroups.nifi.konpyutaika.com"),","),(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("inlineCode",{parentName:"li"},"nifiregistryclients.nifi.konpyutaika.com"),","),(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("inlineCode",{parentName:"li"},"nifiparametercontexts.nifi.konpyutaika.com"),","),(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("inlineCode",{parentName:"li"},"nifidataflows.nifi.konpyutaika.com"),","),(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("inlineCode",{parentName:"li"},"nifinodegroupautoscalers.nifi.konpyutaika.com"),",")),(0,l.kt)("h3",{id:"configuration"},"Configuration"),(0,l.kt)("p",null,"The following tables lists the configurable parameters of the NiFi Operator Helm chart and their default values."),(0,l.kt)("table",null,(0,l.kt)("thead",{parentName:"table"},(0,l.kt)("tr",{parentName:"thead"},(0,l.kt)("th",{parentName:"tr",align:null},"Parameter"),(0,l.kt)("th",{parentName:"tr",align:null},"Description"),(0,l.kt)("th",{parentName:"tr",align:null},"Default"))),(0,l.kt)("tbody",{parentName:"table"},(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"image.repository")),(0,l.kt)("td",{parentName:"tr",align:null},"Image"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"ghcr.io/konpyutaika/docker-images/nifikop"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"image.tag")),(0,l.kt)("td",{parentName:"tr",align:null},"Image tag"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"v1.3.0-release"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"image.pullPolicy")),(0,l.kt)("td",{parentName:"tr",align:null},"Image pull policy"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"Always"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"image.imagePullSecrets.enabled")),(0,l.kt)("td",{parentName:"tr",align:null},"Enable tue use of secret for docker image"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"false"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"image.imagePullSecrets.name")),(0,l.kt)("td",{parentName:"tr",align:null},"Name of the secret to connect to docker registry"),(0,l.kt)("td",{parentName:"tr",align:null},"-")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"certManager.enabled")),(0,l.kt)("td",{parentName:"tr",align:null},"Enable cert-manager integration"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"true"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"rbacEnable")),(0,l.kt)("td",{parentName:"tr",align:null},"If true, create & use RBAC resources"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"true"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"labels")),(0,l.kt)("td",{parentName:"tr",align:null},"Labels to add to all deployed resources"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"{}"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"annotations")),(0,l.kt)("td",{parentName:"tr",align:null},"Annotations to add to all deployed resources"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"{}"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"resources")),(0,l.kt)("td",{parentName:"tr",align:null},"Pod resource requests & limits"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"{}"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"metrics.enabled")),(0,l.kt)("td",{parentName:"tr",align:null},"deploy service for metrics"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"false"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"metrics.port")),(0,l.kt)("td",{parentName:"tr",align:null},"Set port for operator metrics"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"8081"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"logLevel")),(0,l.kt)("td",{parentName:"tr",align:null},"Log level to output"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"Info"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"logEncoding")),(0,l.kt)("td",{parentName:"tr",align:null},"Log encoding to use. Either ",(0,l.kt)("inlineCode",{parentName:"td"},"json")," or ",(0,l.kt)("inlineCode",{parentName:"td"},"console")),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"json"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"certManager.clusterScoped")),(0,l.kt)("td",{parentName:"tr",align:null},"If true setup cluster scoped resources"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"false"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"namespaces")),(0,l.kt)("td",{parentName:"tr",align:null},"List of namespaces where Operator watches for custom resources. Make sure the operator ServiceAccount is granted ",(0,l.kt)("inlineCode",{parentName:"td"},"get")," permissions on this ",(0,l.kt)("inlineCode",{parentName:"td"},"Node")," resource when using limited RBACs."),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},'""')," i.e. all namespaces")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"nodeSelector")),(0,l.kt)("td",{parentName:"tr",align:null},"Node selector configuration for operator pod"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"{}"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"affinity")),(0,l.kt)("td",{parentName:"tr",align:null},"Node affinity configuration for operator pod"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"{}"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"tolerations")),(0,l.kt)("td",{parentName:"tr",align:null},"Toleration configuration for operator pod"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"{}"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"serviceAccount.create")),(0,l.kt)("td",{parentName:"tr",align:null},"Whether the SA creation is delegated to the chart or not"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"true"))),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"serviceAccount.name")),(0,l.kt)("td",{parentName:"tr",align:null},"Name of the SA used for NiFiKop deployment"),(0,l.kt)("td",{parentName:"tr",align:null},"release name")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"webhook.enabled")),(0,l.kt)("td",{parentName:"tr",align:null},"Enable webhook migration"),(0,l.kt)("td",{parentName:"tr",align:null},(0,l.kt)("inlineCode",{parentName:"td"},"true"))))),(0,l.kt)("p",null,"Specify each parameter using the ",(0,l.kt)("inlineCode",{parentName:"p"},"--set key=value[,key=value]")," argument to ",(0,l.kt)("inlineCode",{parentName:"p"},"helm install"),". For example,"),(0,l.kt)("p",null,"Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,"),(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-console"},"$ helm install nifikop \\\n      konpyutaika/nifikop \\\n      -f values.yaml\n")),(0,l.kt)("h3",{id:"installing-the-chart"},"Installing the Chart"),(0,l.kt)("admonition",{title:"Skip CRDs",type:"important"},(0,l.kt)("p",{parentName:"admonition"},"In the case where you don't want to deploy the crds using helm (",(0,l.kt)("inlineCode",{parentName:"p"},"--skip-crds"),") you need to deploy manually the crds beforehand:"),(0,l.kt)("pre",{parentName:"admonition"},(0,l.kt)("code",{parentName:"pre",className:"language-bash"},"kubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nificlusters.yaml\nkubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiusers.yaml\nkubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiusergroups.yaml\nkubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifidataflows.yaml\nkubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiparametercontexts.yaml\nkubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiregistryclients.yaml\nkubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifinodegroupautoscalers.yaml\n"))),(0,l.kt)("admonition",{title:"Conversion webhook",type:"important"},(0,l.kt)("p",{parentName:"admonition"},"In case you keep the conversions webhook enabled (to handle the conversion of resources from ",(0,l.kt)("inlineCode",{parentName:"p"},"v1alpha1")," to ",(0,l.kt)("inlineCode",{parentName:"p"},"v1"),")\nyou will need to add the following settings to your yaml definition of CRDs:"),(0,l.kt)("pre",{parentName:"admonition"},(0,l.kt)("code",{parentName:"pre",className:"language-yaml"},"...\nannotations:\n    cert-manager.io/inject-ca-from: ${namespace}/${certificate_name}\n...\nspec:\n  ...\n  conversion:\n    strategy: Webhook\n    webhook:\n      clientConfig:\n        service:\n          namespace: ${namespace}\n          name: ${webhook_service_name}\n          path: /convert\n      conversionReviewVersions:\n        - v1\n        - v1alpha1\n  ...\n")),(0,l.kt)("p",{parentName:"admonition"},"Where : "),(0,l.kt)("ul",{parentName:"admonition"},(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("inlineCode",{parentName:"li"},"namespace"),": is the namespace in which you will deploy your helm chart."),(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("inlineCode",{parentName:"li"},"certificate_name"),": is ",(0,l.kt)("inlineCode",{parentName:"li"},"${helm release name}-webhook-cert")),(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("inlineCode",{parentName:"li"},"webhook_service_name"),": is ",(0,l.kt)("inlineCode",{parentName:"li"},"${helm release name}-webhook-cert")))),(0,l.kt)(i.Z,{defaultValue:"dryrun",values:[{label:"dry run",value:"dryrun"},{label:"release name",value:"rn"},{label:"set parameters",value:"set-params"}],mdxType:"Tabs"},(0,l.kt)(o.Z,{value:"dryrun",mdxType:"TabItem"},(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-bash"},'helm install nifikop konpyutaika/nifikop \\\n    --dry-run \\\n    --set logLevel=Debug \\\n    --set namespaces={"nifikop"}\n'))),(0,l.kt)(o.Z,{value:"rn",mdxType:"TabItem"},(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-bash"},"helm install <release name> konpyutaika/nifikop\n"))),(0,l.kt)(o.Z,{value:"set-params",mdxType:"TabItem"},(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-bash"},'helm install nifikop konpyutaika/nifikop --set namespaces={"nifikop"}\n')))),(0,l.kt)("blockquote",null,(0,l.kt)("p",{parentName:"blockquote"},"the ",(0,l.kt)("inlineCode",{parentName:"p"},"--replace")," flag allow you to reuses a charts release name")),(0,l.kt)("h3",{id:"listing-deployed-charts"},"Listing deployed charts"),(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-bash"},"helm list\n")),(0,l.kt)("h3",{id:"get-status-for-the-helm-deployment"},"Get Status for the helm deployment"),(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-bash"},"helm status nifikop\n")),(0,l.kt)("h2",{id:"uninstaling-the-charts"},"Uninstaling the Charts"),(0,l.kt)("p",null,"If you want to delete the operator from your Kubernetes cluster, the operator deployment\nshould be deleted."),(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-bash"},"helm del nifikop\n")),(0,l.kt)("p",null,"The command removes all the Kubernetes components associated with the chart and deletes the helm release."),(0,l.kt)("admonition",{type:"tip"},(0,l.kt)("p",{parentName:"admonition"},"The CRD created by the chart are not removed by default and should be manually cleaned up (if required)")),(0,l.kt)("p",null,"Manually delete the CRD:"),(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-bash"},"kubectl delete crd nificlusters.nifi.konpyutaika.com\nkubectl delete crd nifiusers.nifi.konpyutaika.com\nkubectl delete crd nifiusergroups.nifi.konpyutaika.com\nkubectl delete crd nifiregistryclients.nifi.konpyutaika.com\nkubectl delete crd nifiparametercontexts.nifi.konpyutaika.com\nkubectl delete crd nifidataflows.nifi.konpyutaika.com\n")),(0,l.kt)("admonition",{type:"warning"},(0,l.kt)("p",{parentName:"admonition"},"If you delete the CRD then\nIt will delete ",(0,l.kt)("strong",{parentName:"p"},"ALL")," Clusters that has been created using this CRD!!!\nPlease never delete a CRD without very good care")),(0,l.kt)("p",null,"Helm always keeps records of what releases happened. Need to see the deleted releases ?"),(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-bash"},"helm list --deleted\n")),(0,l.kt)("p",null,"Need to see all of the releases (deleted and currently deployed, as well as releases that\nfailed) ?"),(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-bash"},"helm list --all\n")),(0,l.kt)("p",null,"Because Helm keeps records of deleted releases, a release name cannot be re-used. (If you really need to re-use a\nrelease name, you can use the ",(0,l.kt)("inlineCode",{parentName:"p"},"--replace")," flag, but it will simply re-use the existing release and replace its\nresources.)"),(0,l.kt)("p",null,"Note that because releases are preserved in this way, you can rollback a deleted resource, and have it re-activate."),(0,l.kt)("p",null,"To purge a release"),(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre",className:"language-bash"},"helm delete --purge nifikop\n")),(0,l.kt)("h2",{id:"troubleshooting"},"Troubleshooting"),(0,l.kt)("h3",{id:"install-of-the-crd"},"Install of the CRD"),(0,l.kt)("p",null,"By default, the chart will install the CRDs, but this installation is global for the whole\ncluster, and you may want to not modify the already deployed CRDs."),(0,l.kt)("p",null,"In this case there is a parameter to say to not install the CRDs :"),(0,l.kt)("pre",null,(0,l.kt)("code",{parentName:"pre"},'$ helm install --name nifikop ./helm/nifikop --set namespaces={"nifikop"} --skip-crds\n')))}f.isMDXComponent=!0}}]);