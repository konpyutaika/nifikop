"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[90625],{73633:(e,n,i)=>{i.r(n),i.d(n,{assets:()=>c,contentTitle:()=>s,default:()=>h,frontMatter:()=>r,metadata:()=>d,toc:()=>a});var t=i(62540),o=i(43023);const r={id:"4_node",title:"Node",sidebar_label:"Node"},s=void 0,d={id:"5_references/1_nifi_cluster/4_node",title:"Node",description:"Node defines the nifi node basic configuration",source:"@site/versioned_docs/version-v1.10.0/5_references/1_nifi_cluster/4_node.md",sourceDirName:"5_references/1_nifi_cluster",slug:"/5_references/1_nifi_cluster/4_node",permalink:"/nifikop/docs/v1.10.0/5_references/1_nifi_cluster/4_node",draft:!1,unlisted:!1,editUrl:"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.10.0/5_references/1_nifi_cluster/4_node.md",tags:[],version:"v1.10.0",lastUpdatedBy:"Juldrixx",lastUpdatedAt:1722603841e3,frontMatter:{id:"4_node",title:"Node",sidebar_label:"Node"},sidebar:"docs",previous:{title:"Node configuration",permalink:"/nifikop/docs/v1.10.0/5_references/1_nifi_cluster/3_node_config"},next:{title:"Node state",permalink:"/nifikop/docs/v1.10.0/5_references/1_nifi_cluster/5_node_state"}},c={},a=[{value:"Node",id:"node",level:2}];function l(e){const n={a:"a",code:"code",h2:"h2",p:"p",pre:"pre",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,o.R)(),...e.components};return(0,t.jsxs)(t.Fragment,{children:[(0,t.jsx)(n.p,{children:"Node defines the nifi node basic configuration"}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-yaml",children:'    - id: 0\n      # nodeConfigGroup can be used to ease the node configuration, if set only the id is required\n      nodeConfigGroup: "default_group"\n      # readOnlyConfig can be used to pass Nifi node config\n      # which has type read-only these config changes will trigger rolling upgrade\n      readOnlyConfig:\n        nifiProperties:\n          overrideConfigs: |\n            nifi.ui.banner.text=NiFiKop - Node 0\n      # node configuration\n#       nodeConfig:\n    - id: 2\n      # readOnlyConfig can be used to pass Nifi node config\n      # which has type read-only these config changes will trigger rolling upgrade\n      readOnlyConfig:\n        overrideConfigs: |\n          nifi.ui.banner.text=NiFiKop - Node 2\n      # node configuration\n      nodeConfig:\n        resourcesRequirements:\n          limits:\n            cpu: "2"\n            memory: 3Gi\n          requests:\n            cpu: "1"\n            memory: 1Gi\n        storageConfigs:\n          # Name of the storage config, used to name PV to reuse into sidecars for example.\n          - name: provenance-repository\n            # Path where the volume will be mount into the main nifi container inside the pod.\n            mountPath: "/opt/nifi/provenance_repository"\n            # Metadata to attach to the PVC that gets created\n            metadata:\n              labels:\n                my-label: my-value\n              annotations:\n                my-annotation: my-value\n            # Kubernetes PVC spec\n            # https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#create-a-persistentvolumeclaim\n            pvcSpec:\n              accessModes:\n                - ReadWriteOnce\n              storageClassName: "standard"\n              resources:\n                requests:\n                  storage: 8Gi\n'})}),"\n",(0,t.jsx)(n.h2,{id:"node",children:"Node"}),"\n",(0,t.jsxs)(n.table,{children:[(0,t.jsx)(n.thead,{children:(0,t.jsxs)(n.tr,{children:[(0,t.jsx)(n.th,{children:"Field"}),(0,t.jsx)(n.th,{children:"Type"}),(0,t.jsx)(n.th,{children:"Description"}),(0,t.jsx)(n.th,{children:"Required"}),(0,t.jsx)(n.th,{children:"Default"})]})}),(0,t.jsxs)(n.tbody,{children:[(0,t.jsxs)(n.tr,{children:[(0,t.jsx)(n.td,{children:"id"}),(0,t.jsx)(n.td,{children:"int32"}),(0,t.jsx)(n.td,{children:"unique Node id."}),(0,t.jsx)(n.td,{children:"Yes"}),(0,t.jsx)(n.td,{children:"-"})]}),(0,t.jsxs)(n.tr,{children:[(0,t.jsx)(n.td,{children:"nodeConfigGroup"}),(0,t.jsx)(n.td,{children:"string"}),(0,t.jsx)(n.td,{children:"can be used to ease the node configuration, if set only the id is required"}),(0,t.jsx)(n.td,{children:"No"}),(0,t.jsx)(n.td,{children:'""'})]}),(0,t.jsxs)(n.tr,{children:[(0,t.jsx)(n.td,{children:"readOnlyConfig"}),(0,t.jsx)(n.td,{children:(0,t.jsx)(n.a,{href:"./2_read_only_config",children:"ReadOnlyConfig"})}),(0,t.jsx)(n.td,{children:"readOnlyConfig can be used to pass Nifi node config which has type read-only these config changes will trigger rolling upgrade."}),(0,t.jsx)(n.td,{children:"No"}),(0,t.jsx)(n.td,{children:"nil"})]}),(0,t.jsxs)(n.tr,{children:[(0,t.jsx)(n.td,{children:"nodeConfig"}),(0,t.jsx)(n.td,{children:(0,t.jsx)(n.a,{href:"./3_node_config",children:"NodeConfig"})}),(0,t.jsx)(n.td,{children:"node configuration."}),(0,t.jsx)(n.td,{children:"No"}),(0,t.jsx)(n.td,{children:"nil"})]})]})]})]})}function h(e={}){const{wrapper:n}={...(0,o.R)(),...e.components};return n?(0,t.jsx)(n,{...e,children:(0,t.jsx)(l,{...e})}):l(e)}},43023:(e,n,i)=>{i.d(n,{R:()=>s,x:()=>d});var t=i(63696);const o={},r=t.createContext(o);function s(e){const n=t.useContext(r);return t.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function d(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(o):e.components||o:s(e.components),t.createElement(r.Provider,{value:n},e.children)}}}]);