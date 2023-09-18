---
id: 3_external_cluster
title: External cluster
sidebar_label: External cluster
---

## Common configuration

The operator allows you to manage the Dataflow lifecycle for internal (i.e cluster managed by the operator) and external NiFi cluster.
A NiFi cluster is considered as external as soon as the `NifiCluster` resource used as reference in other NiFi resource explicitly detailed the way to communicate with the cluster.

This feature allows you :

- To automate your Dataflow CI/CD using yaml
- To manage the same way your Dataflow management wherever your cluster is, on bare metal, VMs, k8s, on-premise or on cloud.

To deploy different resources (`NifiRegistryClient`, `NifiUser`, `NifiUserGroup`, `NifiParameterContext`, `NifiDataflow`) you simply have to declare a `NifiCluster` resource explaining how to discuss with the external cluster, and refer to this resource as usual using the `Spec.ClusterRef` field.

To declare an external cluster you have to follow this kind of configuration :

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiCluster
metadata:
  name: externalcluster
spec:
  # rootProcessGroupId contains the uuid of the root process group for this cluster.
  rootProcessGroupId: 'd37bee03-017a-1000-cff7-4eaaa82266b7'
  # nodeURITemplate used to dynamically compute node uri.
  nodeURITemplate: 'nifi0%d.integ.mapreduce.m0.p.fti.net:9090'
  # all node requiresunique id
  nodes:
    - id: 1
    - id: 2
    - id: 3
  # type defines if the cluster is internal (i.e manager by the operator) or external.
  # :Enum={"external","internal"}
  type: 'external'
  # clientType defines if the operator will use basic or tls authentication to query the NiFi cluster.
  # Enum={"tls","basic"}
  clientType: 'basic'
  # secretRef reference the secret containing the informations required to authenticate to the cluster.
  secretRef:
    name: nifikop-credentials
    namespace: nifikop-nifi
```

- The `Spec.RootProcessGroupId` field is required to give the ability to the operator of managing root level policy and default deployment and policy.
- The `Spec.NodeURITemplate` field, defines the hostname template of your NiFi cluster nodes, the operator will use this information and the list of id specified in `Spec.Nodes` field to generate the hostname of the nodes (in the configuration above you will have : `nifi01.integ.mapreduce.m0.p.fti.net:9090`, `nifi02.integ.mapreduce.m0.p.fti.net:9090`, `nifi03.integ.mapreduce.m0.p.fti.net:9090`).
- The `Spec.Type` field defines the type of cluster that this resource is refering to, by default it is `internal`, in our case here we just want to use this resource to reference an existing NiFi cluster, so we set this field to `external`.
- The `Spec.ClientType` field defines how we want to authenticate to the NiFi cluster API, for now we are supporting two modes :
    - `tls` : using client TLS certificate.
    - `basic` : using a username and a password to get an access token.
- The `Spec.SecretRef` defines a reference to a secret which contains the sensitive values that will be used by the operator to authenticate to the NiFi cluster API (ie in basic mode it will contain the password and username).

:::warning
The id of node only support `int32` as type, so if the hostname of your nodes doesn't match with this, you can't use this feature.
:::

## Secret configuration for Basic authentication

When you are using the basic authentication, you have to pass some informations into the secret that is referenced into the `NifiCluster` resource:

- `username` : the username associated to the user that will be used by the operator to request the REST API.
- `password` : the password associated to the user that will be used by the operator to request the REST API.
- `ca.crt (optional)`: the certificate authority to trust the server certificate if needed

The following command shows how you can create this secret :

```console
kubectl create secret generic nifikop-credentials \
  --from-file=username=./secrets/username\
  --from-file=password=./secrets/password\
  --from-file=ca.crt=./secrets/ca.crt\
  -n nifikop-nifi
```

:::info
When you use the basic authentication, the operator will create a secret `<cluster name>-basic-secret` containing for each node an access token that will be maintained by the operator.
:::

## Secret configuration for TLS authentication

When you are using the tls authentication, you have to pass some information into the secret that is referenced into the `NifiCluster` resource:

- `tls.key` : The user private key.
- `tls.crt` : The user certificate.
- `password` : the password associated to the user that will be used by the operator to request the REST API.
- `ca.crt`: The CA certificate
- `truststore.jks`:
- `keystore.jks`: