---
id: 4_nifi_parameter_context
title: NiFi Parameter Context
sidebar_label: NiFi Parameter Context
---

`NifiParameterContext` is the Schema for the NiFi parameter context API.

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiParameterContext
metadata:
  name: dataflow-lifecycle
spec:
  description: "It is a test"
  clusterRef:
    name: nc
    namespace: nifikop
  secretRefs:
    - name: secret-params
      namespace: nifikop
  parameters:
    - name: test
      value: toto
      description: tutu
    - name: test2
      description: toto
      sensistive: true
---
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiParameterContext
metadata:
  name: dataflow-lifecycle-child
spec:
  description: "It is a child test"
  clusterRef:
    name: nc
    namespace: nifikop
  secretRefs:
    - name: secret-params
      namespace: nifikop
  inheritedParameterContexts:
    - name: dataflow-lifecycle
  parameters:
    - name: test
      value: toto-child
      description: tutu (child)
```

## NifiParameterContext

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects parameter contexts must create.|No|nil|
|spec|[NifiParameterContextSpec](#NifiParameterContextspec)|defines the desired state of NifiParameterContext.|No|nil|
|status|[NifiParameterContextStatus](#NifiParameterContextstatus)|defines the observed state of NifiParameterContext.|No|nil|

## NifiParameterContextsSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|description|string| describes the Parameter Context. |No| - |
|parameters|\[&nbsp;\][Parameter](#parameter)| a list of non-sensitive Parameters. |Yes| - |
|secretRefs|\[&nbsp;\][SecretReference](#secretreference)| a list of secret containing sensitive parameters (the key will name of the parameter) |No| - |
|clusterRef|[ClusterReference](./2_nifi_user#clusterreference)| contains the reference to the NifiCluster with the one the user is linked. |Yes| - |
|inheritedParameterContext|[ParameterContextReference](#parametercontextreference)| contains the reference(s) to the NiFiParameterContext it should inherit from. |No| - |
|disableTakeOver|bool| whether or not the operator should take over an existing parameter context if its name is the same. |No| - |

## NifiParameterContextStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|id|string| nifi parameter context's id. |Yes| - |
|version|int64| the last nifi parameter context revision version catched. |Yes| - |
|latestUpdateRequest|[ParameterContextUpdateRequest](#parametercontextupdaterequest)|the latest update request. |Yes| - |
|version|int64| the last nifi parameter context revision version catched. |Yes| - |

## Parameter

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string| the name of the Parameter. |Yes| - |
|value|string| the value of the Parameter. |No| - |
|description|string| the description of the Parameter. |No| - |
|sensitive|string| Whether the parameter is sensitive or not. |No| false |

## SecretReference

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string| name of the secret. |Yes| - |
|namespace|string| the secret namespace location. |Yes| - |


## ParameterContextUpdateRequest

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|id|string| the id of the update request. |Yes| - |
|uri|string| the uri for this request. |Yes| - |
|submissionTime|string|  the timestamp of when the request was submitted This property is read only. |Yes| - |
|lastUpdated|string| the timestamp of when the request was submitted This property is read only. |Yes| - |
|complete|bool| whether or not this request has completed. |Yes| false |
|failureReason|string| an explication of why the request failed, or null if this request has not failed. |Yes| - |
|percentCompleted|int32| the percentage complete of the request, between 0 and 100. |Yes| - |
|state|string| the state of the request. |Yes| - |

## ParameterContextReference

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string| name of the NifiParameterContext. |Yes| - |
|namespace|string| the NifiParameterContext namespace location. |No| - |
