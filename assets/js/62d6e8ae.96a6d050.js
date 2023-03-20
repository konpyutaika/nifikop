"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[2167],{35318:(e,n,t)=>{t.d(n,{Zo:()=>l,kt:()=>f});var r=t(27378);function a(e,n,t){return n in e?Object.defineProperty(e,n,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[n]=t,e}function i(e,n){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);n&&(r=r.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),t.push.apply(t,r)}return t}function o(e){for(var n=1;n<arguments.length;n++){var t=null!=arguments[n]?arguments[n]:{};n%2?i(Object(t),!0).forEach((function(n){a(e,n,t[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):i(Object(t)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(t,n))}))}return e}function s(e,n){if(null==e)return{};var t,r,a=function(e,n){if(null==e)return{};var t,r,a={},i=Object.keys(e);for(r=0;r<i.length;r++)t=i[r],n.indexOf(t)>=0||(a[t]=e[t]);return a}(e,n);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);for(r=0;r<i.length;r++)t=i[r],n.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(a[t]=e[t])}return a}var c=r.createContext({}),p=function(e){var n=r.useContext(c),t=n;return e&&(t="function"==typeof e?e(n):o(o({},n),e)),t},l=function(e){var n=p(e.components);return r.createElement(c.Provider,{value:n},e.children)},u={inlineCode:"code",wrapper:function(e){var n=e.children;return r.createElement(r.Fragment,{},n)}},m=r.forwardRef((function(e,n){var t=e.components,a=e.mdxType,i=e.originalType,c=e.parentName,l=s(e,["components","mdxType","originalType","parentName"]),m=p(t),f=a,d=m["".concat(c,".").concat(f)]||m[f]||u[f]||i;return t?r.createElement(d,o(o({ref:n},l),{},{components:t})):r.createElement(d,o({ref:n},l))}));function f(e,n){var t=arguments,a=n&&n.mdxType;if("string"==typeof e||a){var i=t.length,o=new Array(i);o[0]=m;var s={};for(var c in n)hasOwnProperty.call(n,c)&&(s[c]=n[c]);s.originalType=e,s.mdxType="string"==typeof e?e:a,o[1]=s;for(var p=2;p<i;p++)o[p]=t[p];return r.createElement.apply(null,o)}return r.createElement.apply(null,t)}m.displayName="MDXCreateElement"},87342:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>l,contentTitle:()=>c,default:()=>f,frontMatter:()=>s,metadata:()=>p,toc:()=>u});var r=t(25773),a=t(30808),i=(t(27378),t(35318)),o=["components"],s={id:"2_groups_management",title:"Groups management",sidebar_label:"Groups management"},c=void 0,p={unversionedId:"3_manage_nifi/2_manage_users_and_accesses/2_groups_management",id:"version-v1.0.0/3_manage_nifi/2_manage_users_and_accesses/2_groups_management",title:"Groups management",description:"To simplify the access management Apache NiFi allows to define groups containing a list of users, on which we apply a list of access policies.",source:"@site/versioned_docs/version-v1.0.0/3_manage_nifi/2_manage_users_and_accesses/2_groups_management.md",sourceDirName:"3_manage_nifi/2_manage_users_and_accesses",slug:"/3_manage_nifi/2_manage_users_and_accesses/2_groups_management",permalink:"/nifikop/docs/v1.0.0/3_manage_nifi/2_manage_users_and_accesses/2_groups_management",draft:!1,editUrl:"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.0.0/3_manage_nifi/2_manage_users_and_accesses/2_groups_management.md",tags:[],version:"v1.0.0",lastUpdatedBy:"Alexandre Guitton",lastUpdatedAt:1668875267,formattedLastUpdatedAt:"Nov 19, 2022",frontMatter:{id:"2_groups_management",title:"Groups management",sidebar_label:"Groups management"},sidebar:"docs",previous:{title:"Users management",permalink:"/nifikop/docs/v1.0.0/3_manage_nifi/2_manage_users_and_accesses/1_users_management"},next:{title:"Managed groups",permalink:"/nifikop/docs/v1.0.0/3_manage_nifi/2_manage_users_and_accesses/3_managed_groups"}},l={},u=[],m={toc:u};function f(e){var n=e.components,t=(0,a.Z)(e,o);return(0,i.kt)("wrapper",(0,r.Z)({},m,t,{components:n,mdxType:"MDXLayout"}),(0,i.kt)("p",null,"To simplify the access management Apache NiFi allows to define groups containing a list of users, on which we apply a list of access policies.\nThis part is supported by the operator using the ",(0,i.kt)("inlineCode",{parentName:"p"},"NifiUserGroup")," resource :"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre",className:"language-yaml"},'apiVersion: nifi.konpyutaika.com/v1\nkind: NifiUserGroup\nmetadata:\n  name: group-test\nspec:\n  # Contains the reference to the NifiCluster with the one the registry client is linked.\n  clusterRef:\n    name: nc\n    namespace: nifikop\n  # contains the list of reference to NifiUsers that are part to the group.\n  usersRef:\n    - name: nc-0-node.nc-headless.nifikop.svc.cluster.local\n#      namespace: nifikop\n    - name: nc-controller.nifikop.mgt.cluster.local\n  # defines the list of access policies that will be granted to the group.\n  accessPolicies:\n      # defines the kind of access policy, could be "global" or "component".\n    - type: global\n      # defines the kind of action that will be granted, could be "read" or "write"\n      action: read\n      # resource defines the kind of resource targeted by this access policies, please refer to the following page :\n      # https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#access-policies\n      resource: /counters\n#      # componentType is used if the type is "component", it\'s allow to define the kind of component on which is the\n#      # access policy\n#      componentType: "process-groups"\n#      # componentId is used if the type is "component", it\'s allow to define the id of the component on which is the\n#      # access policy\n#      componentId: ""\n')),(0,i.kt)("p",null,"When you create a ",(0,i.kt)("inlineCode",{parentName:"p"},"NifiUserGroup")," resource, the operator will create and manage a group named ",(0,i.kt)("inlineCode",{parentName:"p"},"${resource namespace}-${resource name}")," in Nifi.\nTo declare the users that are part of this group, you just have to declare them in the ",(0,i.kt)("a",{parentName:"p",href:"../../5_references/6_nifi_usergroup#userreference"},"NifiUserGroup.UsersRef")," field."),(0,i.kt)("admonition",{type:"important"},(0,i.kt)("p",{parentName:"admonition"},"The ",(0,i.kt)("a",{parentName:"p",href:"../../5_references/6_nifi_usergroup#userreference"},"NifiUserGroup.UsersRef")," requires to declare the name and namespace of a ",(0,i.kt)("inlineCode",{parentName:"p"},"NifiUser")," resource, so it is previously required to declare the resource."),(0,i.kt)("p",{parentName:"admonition"},"It's required to create the resource even if the user is already declared in NiFi Cluster (In that case the operator will just sync the kubernetes resource).")),(0,i.kt)("p",null,"Like for ",(0,i.kt)("inlineCode",{parentName:"p"},"NifiUser")," you can declare a list of ",(0,i.kt)("a",{parentName:"p",href:"../../5_references/2_nifi_user#accesspolicy"},"AccessPolicies")," to give a list of access to your user."),(0,i.kt)("p",null,"In the example above we are giving to users ",(0,i.kt)("inlineCode",{parentName:"p"},"nc-0-node.nc-headless.nifikop.svc.cluster.local")," and ",(0,i.kt)("inlineCode",{parentName:"p"},"nc-controller.nifikop.mgt.cluster.local")," the right to view the counters information."))}f.isMDXComponent=!0}}]);