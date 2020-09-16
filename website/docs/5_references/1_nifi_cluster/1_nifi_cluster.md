---
id: 1_nifi_cluster
title: NiFi cluster
sidebar_label: NiFi cluster
---

`NifiCluster` describes the desired state of the NiFi cluster we want to setup through the operator.

```yaml
apiVersion: nifi.orange.com/v1alpha1
kind: NifiCluster
metadata:
  name: simplenifi
spec:
  service:
    headlessEnabled: true
  zkAddresse: "zookeepercluster-client.zookeeper:2181"
  zkPath: "/simplenifi"
  clusterImage: "apache/nifi:1.11.3"
  oneNifiNodePerNode: false
  nodeConfigGroups:
    default_group:
      isNode: true
      storageConfigs:
        - mountPath: "/opt/nifi/nifi-current/logs"
          name: logs
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            storageClassName: "standard"
            resources:
              requests:
                storage: 10Gi
      serviceAccountName: "default"
      resourcesRequirements:
        limits:
          cpu: "2"
          memory: 3Gi
        requests:
          cpu: "1"
          memory: 1Gi
  nodes:
    - id: 1
      nodeConfigGroup: "default_group"
    - id: 2
      nodeConfigGroup: "default_group"
  propagateLabels: true
  nifiClusterTaskSpec:
    retryDurationMinutes: 10
  listenersConfig:
    internalListeners:
      - type: "http"
        name: "http"
        containerPort: 8080
      - type: "cluster"
        name: "cluster"
        containerPort: 6007
      - type: "s2s"
        name: "s2s"
        containerPort: 10000
```

## NifiCluster

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects users must create.|No|nil|
|spec|[NifiClusterSpec](#nificlusterspec)|defines the desired state of NifiCluster.|No|nil|
|status|[NifiClusterStatus](#nificlusterstatus)|defines the observed state of NifiCluster.|No|nil|

## NifiClusterSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|service|[ServicePolicy](#servicepolicy)| defines the policy for services owned by NiFiKop operator. |No| - |
|pod|[PodPolicy](#podpolicy)| defines the policy for pod owned by NiFiKop operator. |No| - |
|zkAddresse|string| specifies the ZooKeeper connection string in the form hostname:port where host and port are those of a Zookeeper server.|Yes|""|
|zkPath|string| specifies the Zookeeper chroot path as part of its Zookeeper connection string which puts its data under same path in the global ZooKeeper namespace.|Yes|"/"|
|initContainerImage|string|  can override the default image used into the init container to check if ZoooKeeper server is reachable.. |Yes|"busybox"|
|initContainers|\[ \]string| defines additional initContainers configurations. |No|\[ \]|
|clusterImage|string| can specify the whole nificluster image in one place. |No|""|
|clusterSecure|boolean| cluster nodes secure mode : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#cluster_common_properties. |Yes|false|
|siteToSiteSecure|boolean| site to Site properties Secure mode : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#site_to_site_properties. |Yes|false|
|oneNifiNodePerNode|boolean|if set to true every nifi node is started on a new node, if there is not enough node to do that it will stay in pending state. If set to false the operator also tries to schedule the nifi node to a unique node but if the node number is insufficient the nifi node will be scheduled to a node where a nifi node is already running.|Yes| nil |
|propagateLabels|boolean| - |Yes|false|
|initialAdminUser|string| name of the user account which will be configured as initial admin into NiFi cluster : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#initial-admin-identity. |No|""|
|readOnlyConfig|[ReadOnlyConfig](/nifikop/docs/5_references/1_nifi_cluster/2_read_only_config)| specifies the read-only type Nifi config cluster wide, all theses will be merged with node specified readOnly configurations, so it can be overwritten per node.|No| nil |
|nodeConfigGroups|map\[string\][NodeConfig](/nifikop/docs/5_references/1_nifi_cluster/3_node_config)| specifies multiple node configs with unique name|No| nil |
|nodes|\[  \][Node](/nifikop/docs/5_references/1_nifi_cluster/3_node_config)| specifies the list of cluster nodes, all node requires an image, unique id, and storageConfigs settings|Yes| nil 
|ldapConfiguration|[LdapConfiguration](#ldapconfiguration)| specifies the configuration if you want to use LDAP.|No| nil |
|nifiClusterTaskSpec|[NifiClusterTaskSpec](#nificlustertaskspec)| specifies the configuration of the nifi cluster Tasks.|No| nil |
|listenersConfig|[ListenersConfig](/nifikop/docs/5_references/1_nifi_cluster/6_listeners_config)| listenersConfig specifies nifi's listener specifig configs.|Yes| - |

## NifiClusterStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|nodesState|map\[string\][NodeState](/nifikop/docs/5_references/1_nifi_cluster/5_node_state)|Store the state of each nifi node.|No| - |
|State|[ClusterState](#clusterstate)|Store the state of each nifi node.|Yes| - |

## ServicePolicy

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|headlessEnabled|boolean| specifies if the cluster should use headlessService for Nifi or individual services using service per nodes may come an handy case of service mesh.|Yes|false|
|annotations|map\[string\]string|Annotations specifies the annotations to attach to services the NiFiKop operator creates|No|-|

## PodPolicy

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|annotations|map\[string\]string|Annotations specifies the annotations to attach to pods the NiFiKop operator creates|No|-|


## LdapConfiguration

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|enabled|boolean| if set to true, we will enable ldap usage into nifi.properties configuration.|No| false |
|url|string| space-separated list of URLs of the LDAP servers (i.e. ldap://${hostname}:${port}).|No| "" |
|searchBase|string| base DN for searching for users (i.e. CN=Users,DC=example,DC=com).|No| "" |
|searchFilter|string| Filter for searching for users against the 'User Search Base'. (i.e. sAMAccountName={0}). The user specified name is inserted into '{0}'.|No| "" |

## NifiClusterTaskSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|retryDurationMinutes|int| describes the amount of time the Operator waits for the task.|Yes| 5 |

## ClusterState

|Name|Value|Description|
|-----|----|------------|
|NifiClusterInitializing|ClusterInitializing|states that the cluster is still in initializing stage|
|NifiClusterInitialized|ClusterInitialized|states that the cluster is initialized|
|NifiClusterReconciling|ClusterReconciling|states that the cluster is still in reconciling stage|
|NifiClusterRollingUpgrading|ClusterRollingUpgrading|states that the cluster is rolling upgrading|
|NifiClusterRunning|ClusterRunning|states that the cluster is in running state|