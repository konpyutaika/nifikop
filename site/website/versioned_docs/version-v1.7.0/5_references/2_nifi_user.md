---
id: 2_nifi_user
title: NiFi User
sidebar_label: NiFi User
---

`NifiUser` is the Schema for the nifi users API.

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiUser
metadata:
  name: aguitton
spec:
  identity: alexandre.guitton@konpyutaika.com
  clusterRef:
    name: nc
    namespace: nifikop
  createCert: false
```

## NifiUser
|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects users must create.|No|nil|
|spec|[NifiUserSpec](#nifiuserspec)|defines the desired state of NifiUser.|No|nil|
|status|[NifiUserStatus](#nifiuserstatus)|defines the observed state of NifiUser.|No|nil|

## NifiUserSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|identity|string| used to define the user identity on NiFi cluster side, when the user's name doesn't suit with Kubernetes resource name. |No| - |
|secretName|string| name of the secret where all cert resources will be stored. |No| - |
|clusterRef|[ClusterReference](#clusterreference)|  contains the reference to the NifiCluster with the one the user is linked. |Yes| - |
|DNSNames|\[&nbsp;\]string| list of DNSNames that the user will used to request the NifiCluster (allowing to create the right certificates associated). |Yes| - |
|includeJKS|boolean| whether or not the the operator also include a Java keystore format (JKS) with you secret. |Yes| - |
|createCert|boolean| whether or not a certificate will be created for this user. |No| - |
|accessPolicies|\[&nbsp;\][AccessPolicy](#accesspolicy)| defines the list of access policies that will be granted to the group. |No| [] |


## NifiUserStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|id|string| the nifi user's node id.|Yes| - |
|version|string| the last nifi  user's node revision version catched.|Yes| - |

## ClusterReference

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string|  name of the NifiCluster. |Yes| - |
|namespace|string|  the NifiCluster namespace location. |Yes| - |

## AccessPolicy

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|type|[AccessPolicyType](#accesspolicytype)| defines the kind of access policy, could be "global" or "component". |Yes| - |
|action|[AccessPolicyAction](#accesspolicyaction)| defines the kind of action that will be granted, could be "read" or "write". |Yes| - |
|resource|[AccessPolicyResource](#accesspolicyresource)| defines the kind of resource targeted by this access policies, please refer to the following page : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#access-policies |Yes| - |
|componentType|string| used if the type is "component", it allows to define the kind of component on which is the access policy. |No| - |
|componentId|string| used if the type is "component", it allows to define the id of the component on which is the access policy. |No| - |

## AccessPolicyType

|Name|Value|Description|
|-----|----|------------|
|GlobalAccessPolicyType|global|Global access policies govern the following system level authorizations|
|ComponentAccessPolicyType|component|Component level access policies govern the following component level authorizations|

## AccessPolicyAction

|Name|Value|Description|
|-----|----|------------|
|ReadAccessPolicyAction|read|Allows users to view|
|WriteAccessPolicyAction|write|Allows users to modify|

## AccessPolicyResource

|Name|Value|Description|
|-----|----|------------|
|FlowAccessPolicyResource|/flow|About the UI|
|ControllerAccessPolicyResource|/controller| about the controller including Reporting Tasks, Controller Services, Parameter Contexts and Nodes in the Cluster|
|ParameterContextAccessPolicyResource|/parameter-context|About the Parameter Contexts. Access to Parameter Contexts are inherited from the "access the controller" policies unless overridden.|
|ProvenanceAccessPolicyResource|/provenance|Allows users to submit a Provenance Search and request Event Lineage|
|RestrictedComponentsAccessPolicyResource|/restricted-components|About the restricted components assuming other permissions are sufficient. The restricted components may indicate which specific permissions are required. Permissions can be granted for specific restrictions or be granted regardless of restrictions. If permission is granted regardless of restrictions, the user can create/modify all restricted components.|
|PoliciesAccessPolicyResource|/policies|About the policies for all components|
|TenantsAccessPolicyResource|/tenants| About the users and user groups|
|SiteToSiteAccessPolicyResource|/site-to-site|Allows other NiFi instances to retrieve Site-To-Site details|
|SystemAccessPolicyResource|/system|Allows users to view System Diagnostics|
|ProxyAccessPolicyResource|/proxy|Allows proxy machines to send requests on the behalf of others|
|CountersAccessPolicyResource|/counters|About counters|
|ComponentsAccessPolicyResource|/| About the component configuration details|
|OperationAccessPolicyResource|/operation|to operate components by changing component run status (start/stop/enable/disable), remote port transmission status, or terminating processor threads|
|ProvenanceDataAccessPolicyResource|/provenance-data|to view provenance events generated by this component|
|DataAccessPolicyResource|/data|About metadata and content for this component in flowfile queues in outbound connections and through provenance events|
|PoliciesComponentAccessPolicyResource|/policies|-|
|DataTransferAccessPolicyResource|/data-transfer|Allows a port to receive data from NiFi instances|

