---
id: 1_nifi_cluster
title: NiFi cluster
sidebar_label: NiFi cluster
---

`NifiCluster` describes the desired state of the NiFi cluster we want to setup through the operator.

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiCluster
metadata:
  name: simplenifi
spec:
  service:
    headlessEnabled: true
    annotations:
      tyty: ytyt
    labels:
      cluster-name: simplenifi
      tete: titi
  clusterManager: zookeeper
  zkAddress: "zookeeper.zookeeper:2181"
  zkPath: /simplenifi
  externalServices:
    - metadata:
        annotations:
          toto: tata
        labels:
          cluster-name: driver-simplenifi
          titi: tutu
      name: driver-ip
      spec:
        portConfigs:
          - internalListenerName: http
            port: 8080
        type: ClusterIP
  clusterImage: "apache/nifi:1.28.0"
  initContainerImage: "bash:5.2.2"
  oneNifiNodePerNode: true
  readOnlyConfig:
    nifiProperties:
      overrideConfigs: |
        nifi.sensitive.props.key=thisIsABadSensitiveKeyPassword
  pod:
    annotations:
      toto: tata
    labels:
      cluster-name: simplenifi
      titi: tutu
  nodeConfigGroups:
    default_group:
      imagePullPolicy: IfNotPresent
      isNode: true
      serviceAccountName: default
      storageConfigs:
        - mountPath: "/opt/nifi/nifi-current/logs"
          name: logs
          reclaimPolicy: Delete
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            storageClassName: "standard"
            resources:
              requests:
                storage: 10Gi
      resourcesRequirements:
        limits:
          cpu: "1"
          memory: 2Gi
        requests:
          cpu: "1"
          memory: 2Gi
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
      - containerPort: 8080
        type: http
        name: http
      - containerPort: 6007
        type: cluster
        name: cluster
      - containerPort: 10000
        type: s2s
        name: s2s
      - containerPort: 9090
        type: prometheus
        name: prometheus
      - containerPort: 6342
        type: load-balance
        name: load-balance

```

## NifiCluster

| Field    | Type                                                                                | Description                                                                                       | Required | Default |
| -------- | ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- | -------- | ------- |
| metadata | [ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta) | is metadata that all persisted resources must have, which includes all objects users must create. | No       | nil     |
| spec     | [NifiClusterSpec](#nificlusterspec)                                                 | defines the desired state of NifiCluster.                                                         | No       | nil     |
| status   | [NifiClusterStatus](#nificlusterstatus)                                             | defines the observed state of NifiCluster.                                                        | No       | nil     |

## NifiClusterSpec

| Field                     | Type                                                                                           | Description                                                                                                                                                                                                                                                                                                                              | Required         | Default                    |
| ------------------------- | ---------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------- | -------------------------- |
| clientType                | Enum={"tls","basic"}                                                                           | defines if the operator will use basic or tls authentication to query the NiFi cluster.                                                                                                                                                                                                                                                  | No               | `tls`                      |
| type                      | Enum={"external","internal"}                                                                   | defines if the cluster is internal (i.e manager by the operator) or external.                                                                                                                                                                                                                                                            | No               | `internal`                 |
| nodeURITemplate           | string                                                                                         | used to dynamically compute node uri.                                                                                                                                                                                                                                                                                                    | if external type | -                          |
| nifiURI                   | stringused access through a LB uri.                                                            | if external type                                                                                                                                                                                                                                                                                                                         | -                |
| rootProcessGroupId        | string                                                                                         | contains the uuid of the root process group for this cluster.                                                                                                                                                                                                                                                                            | if external type | -                          |
| secretRef                 | \[&nbsp;\][SecretReference](../4_nifi_parameter_context#secretreference)                            | reference the secret containing the informations required to authentiticate to the cluster.                                                                                                                                                                                                                                              | if external type | -                          |
| proxyUrl                  | string                                                                                         | defines the proxy required to query the NiFi cluster.                                                                                                                                                                                                                                                                                    | if external type | -                          |
| service                   | [ServicePolicy](#servicepolicy)                                                                | defines the policy for services owned by NiFiKop operator.                                                                                                                                                                                                                                                                               | No               | -                          |
| pod                       | [PodPolicy](#podpolicy)                                                                        | defines the policy for pod owned by NiFiKop operator.                                                                                                                                                                                                                                                                                    | No               | -                          |
| clusterManager            | [ClusterManagerType](#clustermanagertype)                                                           | specifies which manager will handle the cluster leader election and state management.                                                                                                                                                                                                                                                    | No               | zookeeper                  |
| zkAddress                 | string                                                                                         | specifies the ZooKeeper connection string in the form hostname:port where host and port are those of a Zookeeper server.                                                                                                                                                                                                                 | No               | ""                         |
| zkPath                    | string                                                                                         | specifies the Zookeeper chroot path as part of its Zookeeper connection string which puts its data under same path in the global ZooKeeper namespace.                                                                                                                                                                                    | Yes              | "/"                        |
| initContainerImage        | string                                                                                         | can override the default image used into the init container to check if ZoooKeeper server is reachable.                                                                                                                                                                                                                                 | Yes              | "bash"                     |
| initContainers            | \[&nbsp;\]string                                                                                    | defines additional initContainers configurations.                                                                                                                                                                                                                                                                                        | No               | \[&nbsp;\]                      |
| clusterImage              | string                                                                                         | can specify the whole nificluster image in one place.                                                                                                                                                                                                                                                                                    | No               | ""                         |
| oneNifiNodePerNode        | boolean                                                                                        | if set to true every nifi node is started on a new node, if there is not enough node to do that it will stay in pending state. If set to false the operator also tries to schedule the nifi node to a unique node but if the node number is insufficient the nifi node will be scheduled to a node where a nifi node is already running. | No               | nil                        |
| propagateLabels           | boolean                                                                                        | whether the labels defined on the `NifiCluster` metadata will be propagated to resources created by the operator or not.                                                                                                                                                                                                                 | Yes              | false                      |
| managedAdminUsers         | \[&nbsp;\][ManagedUser](#managedusers)                                                              | contains the list of users that will be added to the managed admin group (with all rights).                                                                                                                                                                                                                                              | No               | []                         |
| managedReaderUsers        | \[&nbsp;\][ManagedUser](#managedusers)                                                              | contains the list of users that will be added to the managed admin group (with all rights).                                                                                                                                                                                                                                              | No               | []                         |
| readOnlyConfig            | [ReadOnlyConfig](./2_read_only_config)                                                         | specifies the read-only type Nifi config cluster wide, all theses will be merged with node specified readOnly configurations, so it can be overwritten per node.                                                                                                                                                                         | No               | nil                        |
| nodeUserIdentityTemplate  | string                                                                                         | specifies the template to be used when naming the node user identity (e.g. node-%d-mysuffix)                                                                                                                                                                                                                                             | Yes              | "node-%d-\<cluster-name\>" |
| nodeConfigGroups          | map\[string\][NodeConfig](./3_node_config)                                                     | specifies multiple node configs with unique name                                                                                                                                                                                                                                                                                         | No               | nil                        |
| nodes                     | \[&nbsp;\][Node](./3_node_config)                                                                   | specifies the list of cluster nodes, all node requires an image, unique id, and storageConfigs settings                                                                                                                                                                                                                                  | Yes              | nil                        |
| disruptionBudget          | [DisruptionBudget](#disruptionbudget)                                                          | defines the configuration for PodDisruptionBudget.                                                                                                                                                                                                                                                                                       | No               | nil                        |
| ldapConfiguration         | [LdapConfiguration](#ldapconfiguration)                                                        | specifies the configuration if you want to use LDAP.                                                                                                                                                                                                                                                                                     | No               | nil                        |
| singleUserConfiguration   | [SingleUserConfiguration](#singleuserconfiguration)                                            | specifies the configuration if you want to use SingleUser.                                                                                                                                                                                                                                                                               | No               | nil                        |
| nifiClusterTaskSpec       | [NifiClusterTaskSpec](#nificlustertaskspec)                                                    | specifies the configuration of the nifi cluster Tasks.                                                                                                                                                                                                                                                                                   | No               | nil                        |
| listenersConfig           | [ListenersConfig](./6_listeners_config)                                                        | specifies nifi's listener specifig configs.                                                                                                                                                                                                                                                                                              | No               | -                          |
| sidecarConfigs            | \[&nbsp;\][Container](https://godoc.org/k8s.io/api/core/v1#Container)                               | Defines additional sidecar configurations. [Check documentation for more informations]                                                                                                                                                                                                                                                   |
| externalServices          | \[&nbsp;\][ExternalServiceConfigs](./7_external_service_config)                                     | specifies settings required to access nifi externally.                                                                                                                                                                                                                                                                                   | No               | -                          |
| topologySpreadConstraints | \[&nbsp;\][TopologySpreadConstraint](https://godoc.org/k8s.io/api/core/v1#TopologySpreadConstraint) | specifies any TopologySpreadConstraint objects to be applied to all nodes.                                                                                                                                                                                                                                                               | No               | nil                        |
| nifiControllerTemplate    | string                                                                                         | NifiControllerTemplate specifies the template to be used when naming the node controller (e.g. %s-mysuffix) **Warning: once defined don't change this value either the operator will no longer be able to manage the cluster**                                                                                                           | Yes              | "%s-controller"            |
| controllerUserIdentity    | string                                                                                         | ControllerUserIdentity specifies what to call the static admin user's identity **Warning: once defined don't change this value either the operator will no longer be able to manage the cluster**                                                                                                                                        | Yes              | false                      |


## NifiClusterStatus

| Field              | Type                                     | Description                                                   | Required | Default |
| ------------------ | ---------------------------------------- | ------------------------------------------------------------- | -------- | ------- |
| nodesState         | map\[string\][NodeState](./5_node_state) | Store the state of each nifi node.                            | No       | -       |
| State              | [ClusterState](#clusterstate)            | Store the state of each nifi node.                            | Yes      | -       |
| rootProcessGroupId | string                                   | contains the uuid of the root process group for this cluster. | No       | -       |

## ServicePolicy

| Field           | Type                | Description                                                                                                                                         | Required | Default                                                   |
| --------------- | ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | --------------------------------------------------------- |
| headlessEnabled | boolean             | specifies if the cluster should use headlessService for Nifi or individual services using service per nodes may come an handy case of service mesh. | Yes      | false                                                     |
| serviceTemplate | string              | specifies the template to be used when naming the service.                                                                                          | Yes      | If headlessEnabled = true ? "%s-headless" = "%s-all-node" |
| annotations     | map\[string\]string | Annotations specifies the annotations to attach to services the NiFiKop operator creates                                                            | No       | -                                                         |
| labels          | map\[string\]string | Labels specifies the labels to attach to services the NiFiKop operator creates                                                                      | No       | -                                                         |


## PodPolicy

| Field          | Type                                                             | Description                                                                                                           | Required | Default |
| -------------- | ---------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- | -------- | ------- |
| annotations    | map\[string\]string                                              | Annotations specifies the annotations to attach to pods the NiFiKop operator creates                                  | No       | -       |
| labels         | map\[string\]string                                              | Labels specifies the Labels to attach to pods the NiFiKop operator creates                                            | No       | -       |
| hostAliases    | \[&nbsp;\][HostAlias](https://pkg.go.dev/k8s.io/api/core/v1#HostAlias) | A list of host aliases to include in every pod's /etc/hosts configuration in the scenario where DNS is not available. | No       | \[&nbsp;\]    |
| readinessProbe | [Probe](https://pkg.go.dev/k8s.io/api/core/v1#Probe)             | The readiness probe that the `Pod` is configured with. If not provided, a default will be used.                       | No       | nil     |
| livenessProbe  | [Probe](https://pkg.go.dev/k8s.io/api/core/v1#Probe)             | The liveness probe that the `Pod` is configured with. If not provided, a default will be used.                        | No       | nil     |

## ManagedUsers

| Field    | Type   | Description                                                                                                                                           | Required | Default |
| -------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------- |
| identity | string | identity field is use to define the user identity on NiFi cluster side, it use full when the user's name doesn't suite with Kubernetes resource name. | No       | -       |
| name     | string | name field is use to name the NifiUser resource, if not identity is provided it will be used to name the user on NiFi cluster side.                   | Yes      | -       |

## DisruptionBudget

| Field  | Type   | Description                                                                 | Required | Default |
| ------ | ------ | --------------------------------------------------------------------------- | -------- | ------- |
| create | bool   | if set to true, will create a podDisruptionBudget.                          | No       | -       |
| budget | string | the budget to set for the PDB, can either be static number or a percentage. | Yes      | -       |

## LdapConfiguration

| Field                   | Type    | Description                                                                                                                               | Required | Default     |
| ----------------------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------------- | -------- | ----------- |
| enabled                 | boolean | if set to true, we will enable ldap usage into nifi.properties configuration.                                                             | No       | false       |
| url                     | string  | space-separated list of URLs of the LDAP servers (i.e. ldap://$\{hostname}:$\{port}).                                                     | No       | ""          |
| searchBase              | string  | base DN for searching for users (i.e. CN=Users,DC=example,DC=com).                                                                        | No       | ""          |
| searchFilter            | string  | Filter for searching for users against the 'User Search Base'. (i.e. sAMAccountName={0}). The user specified name is inserted into '{0}'. | No       | ""          |
| authenticationStrategy  | string | How the connection to the LDAP server is authenticated. Possible values are ANONYMOUS, SIMPLE, LDAPS, or START_TLS.                        | No       | START_TLS   |
| managerDn               | string | The DN of the manager that is used to bind to the LDAP server to search for users.                                                         | No       | ""          |
| managerPassword         | string | The password of the manager that is used to bind to the LDAP server to search for users.                                                   | No       | ""          |
| tlsKeystore             | string | Path to the Keystore that is used when connecting to LDAP using LDAPS or START_TLS. Not required for LDAPS. Only used for mutual TLS       | No       | ""          |
| tlsKeystorePassword     | string | Password for the Keystore that is used when connecting to LDAP using LDAPS or START_TLS.                                                   | No       | ""          |
| tlsKeystoreType         | string | Type of the Keystore that is used when connecting to LDAP using LDAPS or START_TLS (i.e. JKS or PKCS12).                                   | No       | ""          |
| tlsTruststore           | string | Path to the Truststore that is used when connecting to LDAP using LDAPS or START_TLS. Required for LDAPS                                   | No       | ""          |
| tlsTruststorePassword   | string | Password for the Truststore that is used when connecting to LDAP using LDAPS or START_TLS.                                                 | No       | ""          |
| tlsTruststoreType       | string | Type of the Truststore that is used when connecting to LDAP using LDAPS or START_TLS (i.e. JKS or PKCS12).                                 | No       | ""          |
| clientAuth              | string | Client authentication policy when connecting to LDAP using LDAPS or START_TLS. Possible values are REQUIRED, WANT, NONE.                   | No       | ""          |
| protocol                | string | Protocol to use when connecting to LDAP using LDAPS or START_TLS. (i.e. TLS, TLSv1.1, TLSv1.2, etc).                                       | No       | ""          |
| shutdownGracefully      | string | Specifies whether the TLS should be shut down gracefully before the target context is closed. Defaults to false.                           | No       | ""          |
| referralStrategy        | string | Strategy for handling referrals. Possible values are FOLLOW, IGNORE, THROW.                                                                | No       | FOLLOW          |
| identityStrategy        | string | Strategy to identify users. Possible values are USE_DN and USE_USERNAME.                                                                   | No       | USE_DN      |

## SingleUserConfiguration

| Field             | Type                                                           | Description                                                                                                                                                                                                                         | Required | Default                                    |
| ----------------- | -------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------------------------ |
| enabled           | boolean                                                        | specifies whether or not the cluster should use single user authentication for Nifi                                                                                                                                                 | No       | false                                      |
| authorizerEnabled | boolean                                                        | specifies if the cluster should use use the single-user-authorizer instead of the managed-authorizer (if enabled, the creation of users and user groups will not work in NiFi, and the single user will have no rights by default.) | No       | true                                       |
| secretRef         | [SecretReference](../4_nifi_parameter_context#secretreference) | references the secret containing the informations required to authentiticate to the cluster                                                                                                                                         | No       | nil                                        |
| secretKeys        | [UserSecretKeys](#usersecretkeys)                              | references the keys from the secret containing the user name and password.                                                                                                                                                          | No       | \{username:"username", password:"password"} |

## NifiClusterTaskSpec

| Field                | Type | Description                                                                                                                                | Required | Default |
| -------------------- | ---- | ------------------------------------------------------------------------------------------------------------------------------------------ | -------- | ------- |
| retryDurationMinutes | int  | describes the time the operator waits before going back and retrying a cluster task, which can be: scale up, scale down, rolling upgrade..| Yes      | 5       |

## ClusterState

| Name                        | Value                   | Description                                            |
| --------------------------- | ----------------------- | ------------------------------------------------------ |
| NifiClusterInitializing     | ClusterInitializing     | states that the cluster is still in initializing stage |
| NifiClusterInitialized      | ClusterInitialized      | states that the cluster is initialized                 |
| NifiClusterReconciling      | ClusterReconciling      | states that the cluster is still in reconciling stage  |
| NifiClusterRollingUpgrading | ClusterRollingUpgrading | states that the cluster is rolling upgrading           |
| NifiClusterRunning          | ClusterRunning          | states that the cluster is in running state            |
| NifiClusterNoNodes          | NifiClusterNoNodes      | states that the cluster has no nodes                   |

## UserSecretKeys

| Field    | Type   | Description                                                        | Required | Default  |
| -------- | ------ | ------------------------------------------------------------------ | -------- | -------- |
| username | string | specifies he name of the secret key to retrieve the user name.     | No       | username |
| password | string | specifies he name of the secret key to retrieve the user password. | No       | password |

## ClusterManagerType

| Name                     | Value      | Description                                                                                                                                             |
| ------------------------ | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ZookeeperClusterManager  | zookeeper  | indicates that the cluster leader election and state management will be managed with ZooKeeper. When Zookeeper is configured, you must also configure `NifiCluster.spec.zkPath` and `NifiCluster.spec.zkAddress`. |
| KubernetesClusterManager | kubernetes | indicates that the cluster leader election and state management will be managed with Kubernetes resources, with `Leases` and `ConfigMaps` respectively.                                                           |