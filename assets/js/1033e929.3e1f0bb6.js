"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[58499],{43023:(e,n,t)=>{t.d(n,{R:()=>d,x:()=>c});var i=t(63696);const s={},r=i.createContext(s);function d(e){const n=i.useContext(r);return i.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function c(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(s):e.components||s:d(e.components),i.createElement(r.Provider,{value:n},e.children)}},75930:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>o,contentTitle:()=>c,default:()=>a,frontMatter:()=>d,metadata:()=>i,toc:()=>l});const i=JSON.parse('{"id":"5_references/8_nifi_connection","title":"NiFi Connection","description":"NifiConnection is the Schema for the NiFi connection API.","source":"@site/versioned_docs/version-v1.10.0/5_references/8_nifi_connection.md","sourceDirName":"5_references","slug":"/5_references/8_nifi_connection","permalink":"/nifikop/docs/v1.10.0/5_references/8_nifi_connection","draft":false,"unlisted":false,"editUrl":"https://github.com/konpyutaika/nifikop/edit/master/site/website/versioned_docs/version-v1.10.0/5_references/8_nifi_connection.md","tags":[],"version":"v1.10.0","lastUpdatedBy":"Juldrixx","lastUpdatedAt":1722603841000,"frontMatter":{"id":"8_nifi_connection","title":"NiFi Connection","sidebar_label":"NiFi Connection"},"sidebar":"docs","previous":{"title":"NiFi NodeGroup Autoscaler","permalink":"/nifikop/docs/v1.10.0/5_references/7_nifi_nodegroup_autoscaler"},"next":{"title":"Contribution organization","permalink":"/nifikop/docs/v1.10.0/6_contributing/0_contribution_organization"}}');var s=t(62540),r=t(43023);const d={id:"8_nifi_connection",title:"NiFi Connection",sidebar_label:"NiFi Connection"},c=void 0,o={},l=[{value:"NifiDataflow",id:"nifidataflow",level:2},{value:"NifiConnectionSpec",id:"nificonnectionspec",level:2},{value:"NifiConnectionStatus",id:"nificonnectionstatus",level:2},{value:"ComponentUpdateStrategy",id:"componentupdatestrategy",level:2},{value:"ConnectionState",id:"connectionstate",level:2},{value:"ComponentReference",id:"componentreference",level:2},{value:"ComponentType",id:"componenttype",level:2},{value:"ConnectionConfiguration",id:"connectionconfiguration",level:2},{value:"ConnectionLoadBalanceStrategy",id:"connectionloadbalancestrategy",level:2},{value:"ConnectionLoadBalanceCompression",id:"connectionloadbalancecompression",level:2},{value:"ConnectionPrioritizer",id:"connectionprioritizer",level:2},{value:"ConnectionBend",id:"connectionbend",level:2}];function h(e){const n={a:"a",code:"code",h2:"h2",p:"p",pre:"pre",strong:"strong",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,r.R)(),...e.components};return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsxs)(n.p,{children:[(0,s.jsx)(n.code,{children:"NifiConnection"})," is the Schema for the NiFi connection API."]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-yaml",children:"apiVersion: nifi.konpyutaika.com/v1alpha1\nkind: NifiConnection\nmetadata:\n  name: connection\n  namespace: instances\nspec:\n  source:\n    name: input\n    namespace: instances\n    subName: output_1\n    type: dataflow\n  destination:\n    name: output\n    namespace: instances\n    subName: input_1\n    type: dataflow\n  configuration:\n    flowFileExpiration: 1 hour\n    backPressureDataSizeThreshold: 100 GB\n    backPressureObjectThreshold: 10000\n    loadBalanceStrategy: PARTITION_BY_ATTRIBUTE\n    loadBalancePartitionAttribute: partition_attribute\n    loadBalanceCompression: DO_NOT_COMPRESS\n    prioritizers: \n      - NewestFlowFileFirstPrioritizer\n      - FirstInFirstOutPrioritizer\n    labelIndex: 0\n    bends:\n      - posX: 550\n        posY: 550\n      - posX: 550\n        posY: 440\n      - posX: 550\n        posY: 88\n  updateStrategy: drain\n"})}),"\n",(0,s.jsx)(n.h2,{id:"nifidataflow",children:"NifiDataflow"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Field"}),(0,s.jsx)(n.th,{children:"Type"}),(0,s.jsx)(n.th,{children:"Description"}),(0,s.jsx)(n.th,{children:"Required"}),(0,s.jsx)(n.th,{children:"Default"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"metadata"}),(0,s.jsx)(n.td,{children:(0,s.jsx)(n.a,{href:"https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta",children:"ObjectMetadata"})}),(0,s.jsx)(n.td,{children:"is metadata that all persisted resources must have, which includes all objects dataflows must create."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"nil"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"spec"}),(0,s.jsx)(n.td,{children:(0,s.jsx)(n.a,{href:"#nificonnectionspec",children:"NifiConnectionSpec"})}),(0,s.jsx)(n.td,{children:"defines the desired state of NifiDataflow."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"nil"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"status"}),(0,s.jsx)(n.td,{children:(0,s.jsx)(n.a,{href:"#nificonnectionstatus",children:"NifiConnectionStatus"})}),(0,s.jsx)(n.td,{children:"defines the observed state of NifiDataflow."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"nil"})]})]})]}),"\n",(0,s.jsx)(n.h2,{id:"nificonnectionspec",children:"NifiConnectionSpec"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Field"}),(0,s.jsx)(n.th,{children:"Type"}),(0,s.jsx)(n.th,{children:"Description"}),(0,s.jsx)(n.th,{children:"Required"}),(0,s.jsx)(n.th,{children:"Default"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"source"}),(0,s.jsx)(n.td,{children:(0,s.jsx)(n.a,{href:"#componentreference",children:"ComponentReference"})}),(0,s.jsx)(n.td,{children:"the Source component of the connection."}),(0,s.jsx)(n.td,{children:"Yes"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"destination"}),(0,s.jsx)(n.td,{children:(0,s.jsx)(n.a,{href:"#componentreference",children:"ComponentReference"})}),(0,s.jsx)(n.td,{children:"the Destination component of the connection."}),(0,s.jsx)(n.td,{children:"Yes"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"configuration"}),(0,s.jsx)(n.td,{children:(0,s.jsx)(n.a,{href:"#connectionconfiguration",children:"ConnectionConfiguration"})}),(0,s.jsx)(n.td,{children:"the version of the flow to run."}),(0,s.jsx)(n.td,{children:"Yes"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"updateStrategy"}),(0,s.jsx)(n.td,{children:(0,s.jsx)(n.a,{href:"#componentupdatestrategy",children:"ComponentUpdateStrategy"})}),(0,s.jsx)(n.td,{children:"describes the way the operator will deal with data when a connection will be deleted: Drop or Drain"}),(0,s.jsx)(n.td,{children:"Yes"}),(0,s.jsx)(n.td,{children:"drain"})]})]})]}),"\n",(0,s.jsx)(n.h2,{id:"nificonnectionstatus",children:"NifiConnectionStatus"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Field"}),(0,s.jsx)(n.th,{children:"Type"}),(0,s.jsx)(n.th,{children:"Description"}),(0,s.jsx)(n.th,{children:"Required"}),(0,s.jsx)(n.th,{children:"Default"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"connectionID"}),(0,s.jsx)(n.td,{children:"string"}),(0,s.jsx)(n.td,{children:"connection ID."}),(0,s.jsx)(n.td,{children:"Yes"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"state"}),(0,s.jsx)(n.td,{children:(0,s.jsx)(n.a,{href:"#connectionstate",children:"ConnectionState"})}),(0,s.jsx)(n.td,{children:"the connection current state."}),(0,s.jsx)(n.td,{children:"Yes"}),(0,s.jsx)(n.td,{children:"-"})]})]})]}),"\n",(0,s.jsx)(n.h2,{id:"componentupdatestrategy",children:"ComponentUpdateStrategy"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Name"}),(0,s.jsx)(n.th,{children:"Value"}),(0,s.jsx)(n.th,{children:"Description"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"DrainStrategy"}),(0,s.jsx)(n.td,{children:"drain"}),(0,s.jsx)(n.td,{children:"leads to block stopping of input/output component until they are empty."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"DropStrategy"}),(0,s.jsx)(n.td,{children:"drop"}),(0,s.jsx)(n.td,{children:"leads to dropping all flowfiles from the connection."})]})]})]}),"\n",(0,s.jsx)(n.h2,{id:"connectionstate",children:"ConnectionState"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Name"}),(0,s.jsx)(n.th,{children:"Value"}),(0,s.jsx)(n.th,{children:"Description"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"ConnectionStateCreated"}),(0,s.jsx)(n.td,{children:"Created"}),(0,s.jsx)(n.td,{children:"describes the status of a NifiConnection as created."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"ConnectionStateOutOfSync"}),(0,s.jsx)(n.td,{children:"OutOfSync"}),(0,s.jsx)(n.td,{children:"describes the status of a NifiConnection as out of sync."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"ConnectionStateInSync"}),(0,s.jsx)(n.td,{children:"InSync"}),(0,s.jsx)(n.td,{children:"describes the status of a NifiConnection as in sync."})]})]})]}),"\n",(0,s.jsx)(n.h2,{id:"componentreference",children:"ComponentReference"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Name"}),(0,s.jsx)(n.th,{children:"Value"}),(0,s.jsx)(n.th,{children:"Description"}),(0,s.jsx)(n.th,{children:"Required"}),(0,s.jsx)(n.th,{children:"Default"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"name"}),(0,s.jsx)(n.td,{children:"string"}),(0,s.jsx)(n.td,{children:"the name of the component."}),(0,s.jsx)(n.td,{children:"Yes"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"namespace"}),(0,s.jsx)(n.td,{children:"string"}),(0,s.jsx)(n.td,{children:"the namespace of the component."}),(0,s.jsx)(n.td,{children:"Yes"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"type"}),(0,s.jsx)(n.td,{children:(0,s.jsx)(n.a,{href:"#componenttype",children:"ComponentType"})}),(0,s.jsx)(n.td,{children:"the type of the component (e.g. nifidataflow)."}),(0,s.jsx)(n.td,{children:"Yes"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"subName"}),(0,s.jsx)(n.td,{children:"string"}),(0,s.jsx)(n.td,{children:"the name of the sub component (e.g. queue or port name)."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"-"})]})]})]}),"\n",(0,s.jsx)(n.h2,{id:"componenttype",children:"ComponentType"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Name"}),(0,s.jsx)(n.th,{children:"Value"}),(0,s.jsx)(n.th,{children:"Description"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"ComponentDataflow"}),(0,s.jsx)(n.td,{children:"dataflow"}),(0,s.jsx)(n.td,{children:"indicates that the component is a NifiDataflow."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"ComponentInputPort"}),(0,s.jsx)(n.td,{children:"input-port"}),(0,s.jsxs)(n.td,{children:["indicates that the component is a NifiInputPort. ",(0,s.jsx)(n.strong,{children:"(not implemented)"})]})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"ComponentOutputPort"}),(0,s.jsx)(n.td,{children:"output-port"}),(0,s.jsxs)(n.td,{children:["indicates that the component is a NifiOutputPort. ",(0,s.jsx)(n.strong,{children:"(not implemented)"})]})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"ComponentProcessor"}),(0,s.jsx)(n.td,{children:"processor"}),(0,s.jsxs)(n.td,{children:["indicates that the component is a NifiProcessor. ",(0,s.jsx)(n.strong,{children:"(not implemented)"})]})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"ComponentFunnel"}),(0,s.jsx)(n.td,{children:"funnel"}),(0,s.jsxs)(n.td,{children:["indicates that the component is a NifiFunnel. ",(0,s.jsx)(n.strong,{children:"(not implemented)"})]})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"ComponentProcessGroup"}),(0,s.jsx)(n.td,{children:"process-group"}),(0,s.jsxs)(n.td,{children:["indicates that the component is a NifiProcessGroup. ",(0,s.jsx)(n.strong,{children:"(not implemented)"})]})]})]})]}),"\n",(0,s.jsx)(n.h2,{id:"connectionconfiguration",children:"ConnectionConfiguration"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Name"}),(0,s.jsx)(n.th,{children:"Value"}),(0,s.jsx)(n.th,{children:"Description"}),(0,s.jsx)(n.th,{children:"Required"}),(0,s.jsx)(n.th,{children:"Default"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"flowFileExpiration"}),(0,s.jsx)(n.td,{children:"string"}),(0,s.jsx)(n.td,{children:"the maximum amount of time an object may be in the flow before it will be automatically aged out of the flow."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"backPressureDataSizeThreshold"}),(0,s.jsx)(n.td,{children:"string"}),(0,s.jsx)(n.td,{children:"the maximum data size of objects that can be queued before back pressure is applied."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"1 GB"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"backPressureObjectThreshold"}),(0,s.jsx)(n.td,{children:"*int64"}),(0,s.jsx)(n.td,{children:"the maximum number of objects that can be queued before back pressure is applied."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"10000"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"loadBalanceStrategy"}),(0,s.jsx)(n.td,{children:(0,s.jsx)(n.a,{href:"#connectionloadbalancestrategy",children:"ConnectionLoadBalanceStrategy"})}),(0,s.jsx)(n.td,{children:"how to load balance the data in this Connection across the nodes in the cluster."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"DO_NOT_LOAD_BALANCE"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"loadBalancePartitionAttribute"}),(0,s.jsx)(n.td,{children:"string"}),(0,s.jsx)(n.td,{children:"the FlowFile Attribute to use for determining which node a FlowFile will go to."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"loadBalanceCompression"}),(0,s.jsx)(n.td,{children:(0,s.jsx)(n.a,{href:"#connectionloadbalancecompression",children:"ConnectionLoadBalanceCompression"})}),(0,s.jsx)(n.td,{children:"whether or not data should be compressed when being transferred between nodes in the cluster."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"DO_NOT_COMPRESS"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"prioritizers"}),(0,s.jsxs)(n.td,{children:["[\xa0]",(0,s.jsx)(n.a,{href:"#connectionprioritizer",children:"ConnectionPrioritizer"})]}),(0,s.jsx)(n.td,{children:"the comparators used to prioritize the queue."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"labelIndex"}),(0,s.jsx)(n.td,{children:"*int32"}),(0,s.jsx)(n.td,{children:"the index of the bend point where to place the connection label."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"bends"}),(0,s.jsxs)(n.td,{children:["[\xa0]",(0,s.jsx)(n.a,{href:"#connectionbend",children:"ConnectionBend"})]}),(0,s.jsx)(n.td,{children:"the bend points on the connection."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"-"})]})]})]}),"\n",(0,s.jsx)(n.h2,{id:"connectionloadbalancestrategy",children:"ConnectionLoadBalanceStrategy"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Name"}),(0,s.jsx)(n.th,{children:"Value"}),(0,s.jsx)(n.th,{children:"Description"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"StrategyDoNotLoadBalance"}),(0,s.jsx)(n.td,{children:"DO_NOT_LOAD_BALANCE"}),(0,s.jsx)(n.td,{children:"do not load balance FlowFiles between nodes in the cluster."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"StrategyPartitionByAttribute"}),(0,s.jsx)(n.td,{children:"PARTITION_BY_ATTRIBUTE"}),(0,s.jsx)(n.td,{children:"determine which node to send a given FlowFile to based on the value of a user-specified FlowFile Attribute. All FlowFiles that have the same value for said Attribute will be sent to the same node in the cluster."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"StrategyRoundRobin"}),(0,s.jsx)(n.td,{children:"ROUND_ROBIN"}),(0,s.jsx)(n.td,{children:"flowFiles will be distributed to nodes in the cluster in a Round-Robin fashion. However, if a node in the cluster is not able to receive data as fast as other nodes, that node may be skipped in one or more iterations in order to maximize throughput of data distribution across the cluster."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"StrategySingle"}),(0,s.jsx)(n.td,{children:"SINGLE"}),(0,s.jsx)(n.td,{children:"all FlowFiles will be sent to the same node. Which node they are sent to is not defined."})]})]})]}),"\n",(0,s.jsx)(n.h2,{id:"connectionloadbalancecompression",children:"ConnectionLoadBalanceCompression"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Name"}),(0,s.jsx)(n.th,{children:"Value"}),(0,s.jsx)(n.th,{children:"Description"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"CompressionDoNotCompress"}),(0,s.jsx)(n.td,{children:"DO_NOT_COMPRESS"}),(0,s.jsx)(n.td,{children:"flowFiles will not be compressed."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"CompressionCompressAttributesOnly"}),(0,s.jsx)(n.td,{children:"COMPRESS_ATTRIBUTES_ONLY"}),(0,s.jsx)(n.td,{children:"flowFiles' attributes will be compressed, but the flowFiles' contents will not be."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"CompressionCompressAttributesAndContent"}),(0,s.jsx)(n.td,{children:"COMPRESS_ATTRIBUTES_AND_CONTENT"}),(0,s.jsx)(n.td,{children:"flowFiles' attributes and content will be compressed."})]})]})]}),"\n",(0,s.jsx)(n.h2,{id:"connectionprioritizer",children:"ConnectionPrioritizer"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Name"}),(0,s.jsx)(n.th,{children:"Value"}),(0,s.jsx)(n.th,{children:"Description"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"PrioritizerFirstInFirstOutPrioritizer"}),(0,s.jsx)(n.td,{children:"FirstInFirstOutPrioritizer"}),(0,s.jsx)(n.td,{children:"given two FlowFiles, the one that reached the connection first will be processed first."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"PrioritizerNewestFlowFileFirstPrioritizer"}),(0,s.jsx)(n.td,{children:"NewestFlowFileFirstPrioritizer"}),(0,s.jsx)(n.td,{children:"given two FlowFiles, the one that is newest in the dataflow will be processed first."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"PrioritizerOldestFlowFileFirstPrioritizer"}),(0,s.jsx)(n.td,{children:"OldestFlowFileFirstPrioritizer"}),(0,s.jsx)(n.td,{children:"given two FlowFiles, the one that is oldest in the dataflow will be processed first. 'This is the default scheme that is used if no prioritizers are selected'."})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"PrioritizerPriorityAttributePrioritizer"}),(0,s.jsx)(n.td,{children:"PriorityAttributePrioritizer"}),(0,s.jsx)(n.td,{children:"given two FlowFiles, an attribute called \u201cpriority\u201d will be extracted. The one that has the lowest priority value will be processed first."})]})]})]}),"\n",(0,s.jsx)(n.h2,{id:"connectionbend",children:"ConnectionBend"}),"\n",(0,s.jsxs)(n.table,{children:[(0,s.jsx)(n.thead,{children:(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.th,{children:"Name"}),(0,s.jsx)(n.th,{children:"Value"}),(0,s.jsx)(n.th,{children:"Description"}),(0,s.jsx)(n.th,{children:"Required"}),(0,s.jsx)(n.th,{children:"Default"})]})}),(0,s.jsxs)(n.tbody,{children:[(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"posX"}),(0,s.jsx)(n.td,{children:"*int64"}),(0,s.jsx)(n.td,{children:"the x coordinate."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"-"})]}),(0,s.jsxs)(n.tr,{children:[(0,s.jsx)(n.td,{children:"posY"}),(0,s.jsx)(n.td,{children:"*int64"}),(0,s.jsx)(n.td,{children:"the y coordinate."}),(0,s.jsx)(n.td,{children:"No"}),(0,s.jsx)(n.td,{children:"-"})]})]})]})]})}function a(e={}){const{wrapper:n}={...(0,r.R)(),...e.components};return n?(0,s.jsx)(n,{...e,children:(0,s.jsx)(h,{...e})}):h(e)}}}]);