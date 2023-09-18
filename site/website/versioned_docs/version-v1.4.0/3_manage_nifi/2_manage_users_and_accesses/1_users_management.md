---
id: 1_users_management
title: Users management
sidebar_label: Users management
---

The `NifiUser` resource was already introduced for the [SSL credentials](../1_manage_clusters/1_deploy_cluster/4_ssl_configuration#create-ssl-credentials) concerns.
What we are covering here is the NiFi user management part introduced in this resource.

When you create a `NifiUser` resource the operator will :

1. Try to check if a user already exists with the same name on the NiFi cluster, if it does, the operator will set [NifiUser.Status.Id](../1_manage_clusters/1_deploy_cluster/4_ssl_configuration#create-ssl-credentials) to bind it with the kubernetes resource.
2. If no user is found, the operator will create and manage it (i.e it will ensure the synchronisation with the NiFi Cluster).

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiUser
metadata:
  name: aguitton
spec:
  # identity field is use to define the user identity on NiFi cluster side,
  #	it use full when the user's name doesn't suite with Kubernetes resource name.
  identity: alexandre.guitton@konpyutaika.com
  # Contains the reference to the NifiCluster with the one the registry client is linked.
  clusterRef:
    name: nc
    namespace: nifikop
  # Whether or not the the operator also include a Java keystore format (JKS) with you secret
  includeJKS: false
  # Whether or not a certificate will be created for this user.
  createCert: false
  # defines the list of access policies that will be granted to the group.
  accessPolicies:
      # defines the kind of access policy, could be "global" or "component".
    - type: component
      # defines the kind of action that will be granted, could be "read" or "write"
      action: read
      # resource defines the kind of resource targeted by this access policies, please refer to the following page :
      #	https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#access-policies
      resource: /data
      # componentType is used if the type is "component", it's allow to define the kind of component on which is the
      # access policy
      componentType: "process-groups"
      # componentId is used if the type is "component", it's allow to define the id of the component on which is the
      # access policy
      componentId: ""
```

By default the user name that will be used is the name of the resource.

But as there are some constraints on this name (e.g [RFC 1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names)) that doesn't match with those applied on NiFi, you can override it with the `NifiUser.Spec.Identity` field which is more permissive.
In the example above the kubernetes resource name will be `aguitton` but the NiFi use created on the cluster will be `alexandre.guitton@konpyutaika.com`.

In the case the user will not authenticate himself using TLS authentication, the operator doesn't have to create a certificate, so just set `NifiUser.Spec.CreateCert` to false.

For each user, you have the ability to define a list of [AccessPolicies](../../5_references/2_nifi_user#accesspolicy) to give a list of access to your user.
In the example above we are giving to user `alexandre.guitton@konpyutaika.com` the right to view metadata et content for the root process group in flowfile queues in outbound connections and through provenance events.