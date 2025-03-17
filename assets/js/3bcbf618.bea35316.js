"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[78112],{43023:(e,n,t)=>{t.d(n,{R:()=>s,x:()=>o});var i=t(63696);const r={},a=i.createContext(r);function s(e){const n=i.useContext(a);return i.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function o(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(r):e.components||r:s(e.components),i.createElement(a.Provider,{value:n},e.children)}},49936:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>u,contentTitle:()=>c,default:()=>p,frontMatter:()=>l,metadata:()=>i,toc:()=>d});const i=JSON.parse('{"id":"2_deploy_nifikop/1_quick_start","title":"Quick start","description":"For information about versions compatibilty of the operator features with Kubernetes and Apache NiFi, let\'s have look of the version compatibility page","source":"@site/versioned_docs/version-v1.1.1/2_deploy_nifikop/1_quick_start.md","sourceDirName":"2_deploy_nifikop","slug":"/2_deploy_nifikop/1_quick_start","permalink":"/nifikop/docs/v1.1.1/2_deploy_nifikop/1_quick_start","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.1.1/2_deploy_nifikop/1_quick_start.md","tags":[],"version":"v1.1.1","lastUpdatedBy":"Juldrixx","lastUpdatedAt":1707144987000,"frontMatter":{"id":"1_quick_start","title":"Quick start","sidebar_label":"Quick start"},"sidebar":"docs","previous":{"title":"Roadmap","permalink":"/nifikop/docs/v1.1.1/1_concepts/4_roadmap"},"next":{"title":"Customizable install with Helm","permalink":"/nifikop/docs/v1.1.1/2_deploy_nifikop/2_customizable_install_with_helm"}}');var r=t(62540),a=t(43023),s=t(65457),o=t(97265);const l={id:"1_quick_start",title:"Quick start",sidebar_label:"Quick start"},c=void 0,u={},d=[{value:"Getting Started",id:"getting-started",level:2},{value:"Cluster Setup",id:"cluster-setup",level:3},{value:"Install kubectl",id:"install-kubectl",level:3},{value:"Configure kubectl",id:"configure-kubectl",level:3},{value:"Install cert-manager",id:"install-cert-manager",level:2},{value:"Deploy NiFiKop",id:"deploy-nifikop",level:2}];function h(e){const n={a:"a",admonition:"admonition",blockquote:"blockquote",code:"code",h2:"h2",h3:"h3",li:"li",p:"p",pre:"pre",ul:"ul",...(0,a.R)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsxs)(n.p,{children:["For information about versions compatibilty of the operator features with Kubernetes and Apache NiFi, let's have look of the ",(0,r.jsx)(n.a,{href:"../4_compatibility_versions",children:"version compatibility page"})]}),"\n",(0,r.jsx)(n.h2,{id:"getting-started",children:"Getting Started"}),"\n",(0,r.jsx)(n.h3,{id:"cluster-setup",children:"Cluster Setup"}),"\n",(0,r.jsx)(n.p,{children:"For local testing we recommend following one of the following setup guides:"}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsx)(n.li,{children:(0,r.jsx)(n.a,{href:"https://docs.docker.com/desktop/kubernetes",children:"Docker Desktop (Mac)"})}),"\n",(0,r.jsxs)(n.li,{children:[(0,r.jsx)(n.a,{href:"https://minikube.sigs.k8s.io/docs/start",children:"Minikube"}),"\n",(0,r.jsx)(n.admonition,{type:"note",children:(0,r.jsxs)(n.p,{children:["Start Minikube with at least 4gb RAM with ",(0,r.jsx)(n.code,{children:"minikube start --memory=4000"})]})}),"\n"]}),"\n",(0,r.jsx)(n.li,{children:(0,r.jsx)(n.a,{href:"https://kind.sigs.k8s.io/docs/user/quick-start/",children:"Kind"})}),"\n",(0,r.jsxs)(n.li,{children:["For testing on GKE you can ",(0,r.jsx)(n.a,{href:"https://cloud.google.com/kubernetes-engine/docs/how-to/creating-a-zonal-cluster",children:"create a cluster with the command line or the Cloud Console UI"}),"."]}),"\n",(0,r.jsxs)(n.li,{children:["For testing on EKS you can ",(0,r.jsx)(n.a,{href:"https://eksctl.io/introduction/",children:"install eksctl"})," and run ",(0,r.jsx)(n.code,{children:"eksctl create cluster"})," to create an EKS cluster/VPC/subnets/etc. This process should take 10-15 minutes."]}),"\n"]}),"\n",(0,r.jsx)(n.h3,{id:"install-kubectl",children:"Install kubectl"}),"\n",(0,r.jsxs)(n.p,{children:["If you do not already have the CLI tool ",(0,r.jsx)(n.code,{children:"kubectl"})," installed, please follow ",(0,r.jsx)(n.a,{href:"https://kubernetes.io/docs/tasks/tools/",children:"these instructions to install"}),"."]}),"\n",(0,r.jsx)(n.h3,{id:"configure-kubectl",children:"Configure kubectl"}),"\n",(0,r.jsxs)(n.p,{children:["Configure ",(0,r.jsx)(n.code,{children:"kubectl"})," to connect to your cluster by using ",(0,r.jsx)(n.code,{children:"kubectl config use-context my-cluster-name"}),"."]}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:["For GKE","\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:["Configure ",(0,r.jsx)(n.code,{children:"gcloud"})," with ",(0,r.jsx)(n.code,{children:"gcloud auth login"}),"."]}),"\n",(0,r.jsxs)(n.li,{children:["On the Google Cloud Console, the cluster page will have a ",(0,r.jsx)(n.code,{children:"Connect"})," button, which will give a command to run locally that looks like"]}),"\n"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-console",children:"gcloud container clusters get-credentials CLUSTER_NAME --zone ZONE_NAME --project PROJECT_NAME.\n"})}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:["Use ",(0,r.jsx)(n.code,{children:"kubectl config get-contexts"})," to show the contexts available."]}),"\n",(0,r.jsxs)(n.li,{children:["Run ",(0,r.jsx)(n.code,{children:"kubectl config use-context ${gke context}"})," to access the cluster from ",(0,r.jsx)(n.code,{children:"kubectl"}),"."]}),"\n"]}),"\n"]}),"\n",(0,r.jsxs)(n.li,{children:["For EKS","\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:[(0,r.jsx)(n.a,{href:"https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html",children:"Configure your AWS CLI"})," to connect to your project."]}),"\n",(0,r.jsxs)(n.li,{children:["Install ",(0,r.jsx)(n.a,{href:"https://eksctl.io/introduction/",children:"eksctl"})]}),"\n",(0,r.jsxs)(n.li,{children:["Run ",(0,r.jsx)(n.code,{children:"eksctl utils write-kubeconfig --cluster=${CLUSTER NAME}"})," to make the context available to ",(0,r.jsx)(n.code,{children:"kubectl"})]}),"\n",(0,r.jsxs)(n.li,{children:["Use ",(0,r.jsx)(n.code,{children:"kubectl config get-context"}),"s to show the contexts available."]}),"\n",(0,r.jsxs)(n.li,{children:["Run ",(0,r.jsx)(n.code,{children:"kubectl config use-context ${eks context}"})," to access the cluster with ",(0,r.jsx)(n.code,{children:"kubectl"}),"."]}),"\n"]}),"\n"]}),"\n"]}),"\n",(0,r.jsx)(n.h2,{id:"install-cert-manager",children:"Install cert-manager"}),"\n",(0,r.jsxs)(n.p,{children:["The NiFiKop operator uses ",(0,r.jsx)(n.code,{children:"cert-manager"})," for issuing certificates to users and and nodes, so you'll need to have it setup in case you want to deploy a secured cluster with authentication enabled. The minimum supported cert-manager version is v1.0."]}),"\n",(0,r.jsxs)(s.A,{defaultValue:"directly",values:[{label:"Directly",value:"directly"},{label:"helm 3",value:"helm3"}],children:[(0,r.jsx)(o.A,{value:"directly",children:(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-bash",children:"# Install the CustomResourceDefinitions and cert-manager itself\nkubectl apply -f \\\n    https://github.com/jetstack/cert-manager/releases/download/v1.7.2/cert-manager.yaml\n"})})}),(0,r.jsx)(o.A,{value:"helm3",children:(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-bash",children:"# Install CustomResourceDefinitions first\nkubectl apply --validate=false -f \\\n   https://github.com/jetstack/cert-manager/releases/download/v1.7.2/cert-manager.crds.yaml\n\n# Add the jetstack helm repo\nhelm repo add jetstack https://charts.jetstack.io\nhelm repo update\n\n# You have to create the namespace before executing following command\nhelm install cert-manager \\\n    --namespace cert-manager \\\n    --version v1.7.2 jetstack/cert-manager\n"})})})]}),"\n",(0,r.jsx)(n.h2,{id:"deploy-nifikop",children:"Deploy NiFiKop"}),"\n",(0,r.jsxs)(n.p,{children:["You can deploy the operator using a Helm chart ",(0,r.jsx)(n.a,{href:"https://github.com/konpyutaika/nifikop/tree/master/helm",children:"Helm chart"}),":"]}),"\n",(0,r.jsxs)(n.blockquote,{children:["\n",(0,r.jsxs)(n.p,{children:["To install an other version of the operator use ",(0,r.jsx)(n.code,{children:"helm install --name=nifikop --namespace=nifi --set operator.image.tag=x.y.z konpyutaika-incubator/nifikop"})]}),"\n"]}),"\n",(0,r.jsxs)(n.p,{children:["In the case where you don't want to deploy the crds using helm (",(0,r.jsx)(n.code,{children:"--skip-crds"}),"), you have to deploy manually the crds:"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-bash",children:"kubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nificlusters.yaml\nkubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiusers.yaml\nkubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiusergroups.yaml\nkubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifidataflows.yaml\nkubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiparametercontexts.yaml\nkubectl apply -f https://raw.githubusercontent.com/konpyutaika/nifikop/master/config/crd/bases/nifi.konpyutaika.com_nifiregistryclients.yaml\n"})}),"\n",(0,r.jsxs)(n.admonition,{title:"Conversion webhook",type:"important",children:[(0,r.jsxs)(n.p,{children:["In case you keep the conversion webhook enabled (to handle the conversion of resources from ",(0,r.jsx)(n.code,{children:"v1alpha1"})," to ",(0,r.jsx)(n.code,{children:"v1"}),")\nyou will need to add the following settings to your yaml definition of CRDs:"]}),(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-yaml",children:"...\nannotations:\n    cert-manager.io/inject-ca-from: ${namespace}/${certificate_name}\n...\nspec:\n  ...\n  conversion:\n    strategy: Webhook\n    webhook:\n      clientConfig:\n        service:\n          namespace: ${namespace}\n          name: ${webhook_service_name}\n          path: /convert\n      conversionReviewVersions:\n        - v1\n        - v1alpha1\n  ...\n"})}),(0,r.jsx)(n.p,{children:"Where:"}),(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:[(0,r.jsx)(n.code,{children:"namespace"}),": is the namespace in which you will deploy your helm chart."]}),"\n",(0,r.jsxs)(n.li,{children:[(0,r.jsx)(n.code,{children:"certificate_name"}),": is ",(0,r.jsx)(n.code,{children:"${helm release name}-webhook-cert"})]}),"\n",(0,r.jsxs)(n.li,{children:[(0,r.jsx)(n.code,{children:"webhook_service_name"}),": is ",(0,r.jsx)(n.code,{children:"${helm release name}-webhook-cert"})]}),"\n"]})]}),"\n",(0,r.jsx)(n.p,{children:"Now deploy the helm chart:"}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-bash",children:'# You have to create the namespace before executing following command\nhelm install nifikop \\\n    oci://ghcr.io/konpyutaika/helm-charts/nifikop \\\n    --namespace=nifi \\\n    --version 1.1.1 \\\n    --set image.tag=v1.1.1-release \\\n    --set resources.requests.memory=256Mi \\\n    --set resources.requests.cpu=250m \\\n    --set resources.limits.memory=256Mi \\\n    --set resources.limits.cpu=250m \\\n    --set namespaces={"nifi"}\n'})}),"\n",(0,r.jsx)(n.admonition,{type:"note",children:(0,r.jsxs)(n.p,{children:["Add the following parameter if you are using this instance to only deploy unsecured clusters: ",(0,r.jsx)(n.code,{children:"--set certManager.enabled=false"})]})})]})}function p(e={}){const{wrapper:n}={...(0,a.R)(),...e.components};return n?(0,r.jsx)(n,{...e,children:(0,r.jsx)(h,{...e})}):h(e)}},65457:(e,n,t)=>{t.d(n,{A:()=>y});var i=t(63696),r=t(11750),a=t(93707),s=t(49519),o=t(83604),l=t(95196),c=t(76229),u=t(88030);function d(e){var n,t;return null!=(n=null==(t=i.Children.toArray(e).filter((function(e){return"\n"!==e})).map((function(e){if(!e||(0,i.isValidElement)(e)&&((n=e.props)&&"object"==typeof n&&"value"in n))return e;var n;throw new Error("Docusaurus error: Bad <Tabs> child <"+("string"==typeof e.type?e.type:e.type.name)+'>: all children of the <Tabs> component should be <TabItem>, and every <TabItem> should have a unique "value" prop.')})))?void 0:t.filter(Boolean))?n:[]}function h(e){var n=e.values,t=e.children;return(0,i.useMemo)((function(){var e=null!=n?n:function(e){return d(e).map((function(e){var n=e.props;return{value:n.value,label:n.label,attributes:n.attributes,default:n.default}}))}(t);return function(e){var n=(0,c.XI)(e,(function(e,n){return e.value===n.value}));if(n.length>0)throw new Error('Docusaurus error: Duplicate values "'+n.map((function(e){return e.value})).join(", ")+'" found in <Tabs>. Every value needs to be unique.')}(e),e}),[n,t])}function p(e){var n=e.value;return e.tabValues.some((function(e){return e.value===n}))}function m(e){var n=e.queryString,t=void 0!==n&&n,r=e.groupId,a=(0,s.W6)(),o=function(e){var n=e.queryString,t=void 0!==n&&n,i=e.groupId;if("string"==typeof t)return t;if(!1===t)return null;if(!0===t&&!i)throw new Error('Docusaurus error: The <Tabs> component groupId prop is required if queryString=true, because this value is used as the search param name. You can also provide an explicit value such as queryString="my-search-param".');return null!=i?i:null}({queryString:t,groupId:r});return[(0,l.aZ)(o),(0,i.useCallback)((function(e){if(o){var n=new URLSearchParams(a.location.search);n.set(o,e),a.replace(Object.assign({},a.location,{search:n.toString()}))}}),[o,a])]}function f(e){var n,t,r,a,s=e.defaultValue,l=e.queryString,c=void 0!==l&&l,d=e.groupId,f=h(e),b=(0,i.useState)((function(){return function(e){var n,t=e.defaultValue,i=e.tabValues;if(0===i.length)throw new Error("Docusaurus error: the <Tabs> component requires at least one <TabItem> children component");if(t){if(!p({value:t,tabValues:i}))throw new Error('Docusaurus error: The <Tabs> has a defaultValue "'+t+'" but none of its children has the corresponding value. Available values are: '+i.map((function(e){return e.value})).join(", ")+". If you intend to show no default tab, use defaultValue={null} instead.");return t}var r=null!=(n=i.find((function(e){return e.default})))?n:i[0];if(!r)throw new Error("Unexpected error: 0 tabValues");return r.value}({defaultValue:s,tabValues:f})})),k=b[0],g=b[1],v=m({queryString:c,groupId:d}),x=v[0],j=v[1],y=(n=function(e){return e?"docusaurus.tab."+e:null}({groupId:d}.groupId),t=(0,u.Dv)(n),r=t[0],a=t[1],[r,(0,i.useCallback)((function(e){n&&a.set(e)}),[n,a])]),w=y[0],_=y[1],C=function(){var e=null!=x?x:w;return p({value:e,tabValues:f})?e:null}();return(0,o.A)((function(){C&&g(C)}),[C]),{selectedValue:k,selectValue:(0,i.useCallback)((function(e){if(!p({value:e,tabValues:f}))throw new Error("Can't select invalid tab value="+e);g(e),j(e),_(e)}),[j,_,f]),tabValues:f}}var b=t(95200);const k={tabList:"tabList_J5MA",tabItem:"tabItem_l0OV"};var g=t(62540);function v(e){var n=e.className,t=e.block,i=e.selectedValue,s=e.selectValue,o=e.tabValues,l=[],c=(0,a.a_)().blockElementScrollPositionUntilNextRender,u=function(e){var n=e.currentTarget,t=l.indexOf(n),r=o[t].value;r!==i&&(c(n),s(r))},d=function(e){var n,t=null;switch(e.key){case"Enter":u(e);break;case"ArrowRight":var i,r=l.indexOf(e.currentTarget)+1;t=null!=(i=l[r])?i:l[0];break;case"ArrowLeft":var a,s=l.indexOf(e.currentTarget)-1;t=null!=(a=l[s])?a:l[l.length-1]}null==(n=t)||n.focus()};return(0,g.jsx)("ul",{role:"tablist","aria-orientation":"horizontal",className:(0,r.A)("tabs",{"tabs--block":t},n),children:o.map((function(e){var n=e.value,t=e.label,a=e.attributes;return(0,g.jsx)("li",Object.assign({role:"tab",tabIndex:i===n?0:-1,"aria-selected":i===n,ref:function(e){l.push(e)},onKeyDown:d,onClick:u},a,{className:(0,r.A)("tabs__item",k.tabItem,null==a?void 0:a.className,{"tabs__item--active":i===n}),children:null!=t?t:n}),n)}))})}function x(e){var n=e.lazy,t=e.children,a=e.selectedValue,s=(Array.isArray(t)?t:[t]).filter(Boolean);if(n){var o=s.find((function(e){return e.props.value===a}));return o?(0,i.cloneElement)(o,{className:(0,r.A)("margin-top--md",o.props.className)}):null}return(0,g.jsx)("div",{className:"margin-top--md",children:s.map((function(e,n){return(0,i.cloneElement)(e,{key:n,hidden:e.props.value!==a})}))})}function j(e){var n=f(e);return(0,g.jsxs)("div",{className:(0,r.A)("tabs-container",k.tabList),children:[(0,g.jsx)(v,Object.assign({},n,e)),(0,g.jsx)(x,Object.assign({},n,e))]})}function y(e){var n=(0,b.A)();return(0,g.jsx)(j,Object.assign({},e,{children:d(e.children)}),String(n))}},97265:(e,n,t)=>{t.d(n,{A:()=>s});t(63696);var i=t(11750);const r={tabItem:"tabItem_wHwb"};var a=t(62540);function s(e){var n=e.children,t=e.hidden,s=e.className;return(0,a.jsx)("div",{role:"tabpanel",className:(0,i.A)(r.tabItem,s),hidden:t,children:n})}}}]);