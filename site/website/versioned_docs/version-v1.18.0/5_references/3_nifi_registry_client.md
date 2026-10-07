---
id: 3_nifi_registry_client
title: NiFi Registry Client
sidebar_label: NiFi Registry Client
---

`NifiRegistryClient` is the Schema for the NiFi registry client API. It supports connecting to a NiFi Registry server, a GitHub repository, or a GitLab repository as a flow storage backend.

## Quick examples

**NiFi Registry (default)**

```yaml
apiVersion: nifi.konpyutaika.com/v2alpha1
kind: NifiRegistryClient
metadata:
  name: squidflow
spec:
  clusterRef:
    name: nc
    namespace: nifikop
  description: "Squidflow demo"
  type: registry
  registryClientConfig:
    uri: "http://nifi-registry:18080"
```

**GitHub**

```yaml
apiVersion: nifi.konpyutaika.com/v2alpha1
kind: NifiRegistryClient
metadata:
  name: squidflow-github
spec:
  clusterRef:
    name: nc
    namespace: nifikop
  description: "Squidflow GitHub demo"
  type: github
  githubConfig:
    repositoryOwner: "my-org"
    repositoryName: "nifi-flows"
    authenticationType: PERSONAL_ACCESS_TOKEN
    personalAccessTokenSecretRef:
      name: github-pat-secret
      namespace: nifikop
      data: token
    defaultBranch: main
```

**GitLab**

```yaml
apiVersion: nifi.konpyutaika.com/v2alpha1
kind: NifiRegistryClient
metadata:
  name: squidflow-gitlab
spec:
  clusterRef:
    name: nc
    namespace: nifikop
  description: "Squidflow GitLab demo"
  type: gitlab
  gitlabConfig:
    repositoryNamespace: "my-group/my-subgroup"
    repositoryName: "nifi-flows"
    authenticationType: ACCESS_TOKEN
    accessTokenSecretRef:
      name: gitlab-token-secret
      namespace: nifikop
      data: token
    defaultBranch: main
```

## NifiRegistryClient

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|Metadata that all persisted resources must have, which includes all objects registry clients must create.|No|nil|
|spec|[NifiRegistryClientSpec](#nifiregistryclientspec)|Defines the desired state of NifiRegistryClient.|No|nil|
|status|[NifiRegistryClientStatus](#nifiregistryclientstatus)|Defines the observed state of NifiRegistryClient.|No|nil|

## NifiRegistryClientSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|description|string|Describes the registry client.|No| - |
|type|Enum=`registry`;`github`;`gitlab`|The type of flow storage backend to use.|No|`registry`|
|clusterRef|[ClusterReference](./2_nifi_user#clusterreference)|Contains the reference to the NifiCluster with which the registry client is linked.|Yes| - |
|registryClientConfig|[RegistryClientConfig](#registryclientconfig)|Configuration for a NiFi Registry backend. Required when `type` is `registry`.|No| - |
|githubConfig|[GitHubConfig](#githubconfig)|Configuration for a GitHub backend. Required when `type` is `github`.|No| - |
|gitlabConfig|[GitLabConfig](#gitlabconfig)|Configuration for a GitLab backend. Required when `type` is `gitlab`.|No| - |

:::note
CEL validation rules on the CRD enforce that the matching config block is present for the chosen `type`. For example, setting `type: github` without a `githubConfig` block will be rejected by the API server.
:::

## RegistryClientConfig

Used when `type` is `registry`.

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|uri|string|URI of the NiFi Registry server.|Yes| - |

## GitHubConfig

Used when `type` is `github`.

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|apiUrl|*string|URL of the GitHub API.|No|`https://api.github.com/`|
|repositoryOwner|string|Owner of the repository (user or organization).|Yes| - |
|repositoryName|string|Name of the repository.|Yes| - |
|authenticationType|Enum=`NONE`;`PERSONAL_ACCESS_TOKEN`;`APP_INSTALLATION`|Type of authentication to use.|No| - |
|personalAccessTokenSecretRef|[SecretConfigReference](#secretconfigreference)|Secret containing the personal access token. Required when `authenticationType` is `PERSONAL_ACCESS_TOKEN`.|No| - |
|appId|*string|Identifier of the GitHub App. Required when `authenticationType` is `APP_INSTALLATION`.|No| - |
|appPrivateKeySecretRef|[SecretConfigReference](#secretconfigreference)|Secret containing the RSA private key for the GitHub App. Required when `authenticationType` is `APP_INSTALLATION`.|No| - |
|defaultBranch|*string|Default branch of the repository.|No| - |
|repositoryPath|*string|Path within the repository for storing data.|No|Repository root|
|directoryFilterExclusion|*string|Regex pattern for directories to exclude.|No|`[.]*`|
|parameterContextValues|Enum=`RETAIN`;`REMOVE`;`IGNORE_CHANGES`|How to handle parameter context values.|No| - |

## GitLabConfig

Used when `type` is `gitlab`.

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|url|*string|URL of the GitLab instance.|No|`https://gitlab.com/`|
|apiVersion|Enum=`V4`|GitLab API version.|No| - |
|repositoryNamespace|string|Namespace of the repository (user or group/subgroup path).|Yes| - |
|repositoryName|string|Name of the repository.|Yes| - |
|authenticationType|Enum=`ACCESS_TOKEN`|Type of authentication to use.|No| - |
|accessTokenSecretRef|[SecretConfigReference](#secretconfigreference)|Secret containing the access token. Required when `authenticationType` is `ACCESS_TOKEN` or not set.|No| - |
|connectTimeout|*string|Connect timeout (e.g. `"10 seconds"`).|No| - |
|readTimeout|*string|Read timeout (e.g. `"10 seconds"`).|No| - |
|defaultBranch|*string|Default branch of the repository.|No| - |
|repositoryPath|*string|Path within the repository for storing data.|No|Repository root|
|directoryFilterExclusion|*string|Regex pattern for directories to exclude.|No|`[.]*`|
|parameterContextValues|Enum=`RETAIN`;`REMOVE`;`IGNORE_CHANGES`|How to handle parameter context values.|No| - |

## SecretConfigReference

References a Kubernetes Secret and a specific key within its data map.

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string|Name of the Kubernetes Secret.|Yes| - |
|namespace|string|Namespace of the Secret.|No|Resource namespace|
|data|string|Key within the Secret's `data` map.|Yes| - |

## NifiRegistryClientStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|id|string|The nifi registry client's id.|Yes| - |
|version|int64|The last nifi registry client revision version.|Yes| - |
|latestSecretsResourceVersion|[]SecretResourceVersion|The last observed resource versions of the referenced secrets.|No| - |

## SecretResourceVersion

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string|Name of the secret.|Yes| - |
|namespace|string|Namespace of the secret.|Yes| - |
|resourceVersion|string|ResourceVersion of the secret at last sync.|Yes| - |

## RegistryClientReference

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string| name of the NifiRegistryClient. |Yes| - |
|namespace|string| the NifiRegistryClient namespace location. |Yes| - |
