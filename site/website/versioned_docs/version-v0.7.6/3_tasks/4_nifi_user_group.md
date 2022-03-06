---
id: 4_nifi_user_group
title: Provisioning NiFi Users and Groups
sidebar_label: NiFi Users and Groups
---

## User management

The `NifiUser` resource was already introduced for the [SSL credentials](./2_security/1_ssl.md#create-ssl-credentials) concerns.
What we are covering here is the NiFi user management part introduced in this resource.

When you create a `NifiUser` resource the operator will :

1. Try to check if a user already exists with the same name on the NiFi cluster, if it does, the operator will set [NifiUser.Status.Id](./2_security/1_ssl.md#create-ssl-credentials) to bind it with the kubernetes resource.
2. If no user is found, the operator will create and manage it (i.e it will ensure the synchronisation with the NiFi Cluster).

```yaml
apiVersion: nifi.orange.com/v1alpha1
kind: NifiUser
metadata:
  name: aguitton
spec:
  # identity field is use to define the user identity on NiFi cluster side,
  #	it use full when the user's name doesn't suite with Kubernetes resource name.
  identity: alexandre.guitton@orange.com
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
      componentType: 'process-groups'
      # componentId is used if the type is "component", it's allow to define the id of the component on which is the
      # access policy
      componentId: ''
```

By default the user name that will be used is the name of the resource.

But as there are some constraints on this name (e.g [RFC 1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names)) that doesn't match with those applied on NiFi, you can override it with the `NifiUser.Spec.Identity` field which is more permissive.
In the example above the kubernetes resource name will be `aguitton` but the NiFi use created on the cluster will be `alexandre.guitton@orange.com`.

In the case the user will not authenticate himself using TLS authentication, the operator doesn't have to create a certificate, so just set `NifiUser.Spec.CreateCert` to false.

For each user, you have the ability to define a list of [AccessPolicies](../5_references/2_nifi_user.md#accesspolicy) to give a list of access to your user.
In the example above we are giving to user `alexandre.guitton@orange.com` the right to view metadata et content for the root process group in flowfile queues in outbound connections and through provenance events.

## UserGroup management

To simplify the access management Apache NiFi allows to define groups containing a list of users, on which we apply a list of access policies.
This part is supported by the operator using the `NifiUserGroup` resource :

```yaml
apiVersion: nifi.orange.com/v1alpha1
kind: NifiUserGroup
metadata:
  name: group-test
spec:
  # Contains the reference to the NifiCluster with the one the registry client is linked.
  clusterRef:
    name: nc
    namespace: nifikop
  # contains the list of reference to NifiUsers that are part to the group.
  usersRef:
    - name: nc-0-node.nc-headless.nifikop.svc.cluster.local
    #      namespace: nifikop
    - name: nc-controller.nifikop.mgt.cluster.local
  # defines the list of access policies that will be granted to the group.
  accessPolicies:
    # defines the kind of access policy, could be "global" or "component".
    - type: global
      # defines the kind of action that will be granted, could be "read" or "write"
      action: read
      # resource defines the kind of resource targeted by this access policies, please refer to the following page :
      #	https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#access-policies
      resource: /counters
#      # componentType is used if the type is "component", it's allow to define the kind of component on which is the
#      # access policy
#      componentType: "process-groups"
#      # componentId is used if the type is "component", it's allow to define the id of the component on which is the
#      # access policy
#      componentId: ""
```

When you create a `NifiUserGroup` resource, the operator will create and manage a group named `${resource namespace}-${resource name}` in Nifi.
To declare the users that are part of this group, you just have to declare them in the [NifiUserGroup.UsersRef](../5_references/6_nifi_usergroup.md#userreference) field.

:::important
The [NifiUserGroup.UsersRef](../5_references/6_nifi_usergroup.md#userreference) requires to declare the name and namespace of a `NifiUser` resource, so it is previously required to declare the resource.

It's required to create the resource even if the user is already declared in NiFi Cluster (In that case the operator will just sync the kubernetes resource).
:::

Like for `NifiUser` you can declare a list of [AccessPolicies](../5_references/2_nifi_user.md#accesspolicy) to give a list of access to your user.

In the example above we are giving to users `nc-0-node.nc-headless.nifikop.svc.cluster.local` and `nc-controller.nifikop.mgt.cluster.local` the right to view the counters informations.

## Managed groups for simple setup

In some case these two features could be heavy to define, for example when you have 10 dataflows with one cluster for each of them, it will lead in a lot of `.yaml` files ...
To simplify this, we implement in the operator 2 `managed groups` :

- **Admins :** a group giving access to everything on the NiFi Cluster,
- **Readers :** a group giving access as viewer on the NiFi Cluster.

You can directly define the list of users who belong to each of them in the `NifiCluster.Spec` field :

```yaml
apiVersion: nifi.orange.com/v1alpha1
kind: NifiCluster
metadata:
  name: mynifi
spec:
  ...
  oneNifiNodePerNode: false
  #
  propagateLabels: true
  managedAdminUsers:
    -  identity : "alexandre.guitton@orange.com"
       name: "aguitton"
    -  identity : "nifiuser@orange.com"
       name: "nifiuser"
  managedReaderUsers:
    -  identity : "toto@orange.com"
       name: "toto"
    ...
```

In this example the operator will create and manage 3 `NifiUsers` :

- **aguitton**, with the identity : `alexandre.guitton@orange.com`
- **nifiuser**, with the identity : `nifiuser@orange.com`
- **toto**, with the identity : `toto@orange.com`

And create and manage two groups :

- **managed-admins :** that will contain 3 users (**aguitton**, **nifiuser**, **nc-controller.nifikop.mgt.cluster.local** which is the controller user).
- **managed-readers :** that will contain 1 user (**toto**)

And the rest of the stuff will be reconciled and managed as described for `NifiUsers` and `NifiUserGroups`.

:::note
There is one more group that is created and managed by the operator, this is the **managed-nodes** group, for each node a `NifiUser` is created, and we automatically add them to this group to give them the right list of accesses.

To get the list of managed groups just check the list of `NifiUserGroup` :

```console
kubectl get -n nifikop nifiusergroups.nifi.orange.com
NAME              AGE
managed-admins    6d7h
managed-nodes     6d7h
managed-readers   6d7h
```

:::
