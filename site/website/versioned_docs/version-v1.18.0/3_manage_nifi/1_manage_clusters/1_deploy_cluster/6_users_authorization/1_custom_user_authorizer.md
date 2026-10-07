---
id: 1_custom_user_authorizer
title: Custom User Authorizers
sidebar_label: Custom User Authorizers
---

:::info
This is an advanced configuration topic. In most cases, the default NiFi authorizer configuration is sufficient.
:::

According to the NiFi Admin Guide, an [Authorizer](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#authorizer-configuration) grants users the privileges to manage users and policies by creating preliminary authorizations at startup. By default, the `StandardManagedAuthorizer` leverages a `FileUserGroupProvider` and a `FileAccessPolicyProvider` which are file-based rules for each user you allow to interact with your NiFi cluster.

In many cases, the default authorizer configuration is enough to control access to a NiFi cluster. However, there may be advanced cases where the default [`managed-authorizer`](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#standardmanagedauthorizer) isn't sufficient to make every authorization decision you need. In this case, you can provide a custom authorizer extension and use that instead.

Suppose a custom Authorizer is written and deployed with NiFi that reads the rules from a remote database rather than a local file. We'll call this `DatabaseAuthorizer`. Also suppose it is composed of a `DatabaseUserGroupProvider` and a `DatabaseAccessPolicyProvider`. In order to leverage these, they must end up on NiFi's classpath.

In order to use this authorizer, you need to update NiFi's `authorizers.xml` configuration. This can be done through NiFiKOp by setting either the `Spec.readOnlyConfig.authorizerConfig.replaceTemplateConfigMap` or `Spec.readOnlyConfig.authorizerConfig.replaceTemplateSecretConfig`. The NiFiKOp deployment is dynamic in that node identities are determined at deploy time, so the authorizer configuration is templated to account for this. This means that the replacement ConfigMap or Secret must also be templated.

Following the example, the below would be a sufficient authorizer template replacement:

```yaml
{{- $nodeList := .NodeList }}
{{- $clusterName := .ClusterName }}
{{- $namespace := .Namespace }}<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<authorizers>
    <userGroupProvider>
        <identifier>file-user-group-provider</identifier>
        <class>org.apache.nifi.authorization.FileUserGroupProvider</class>
        <property name="Users File">../data/users.xml</property>
        <property name="Legacy Authorized Users File"></property>
        <property name="Initial User Identity admin">{{ .ControllerUser }}</property>
{{- range $i, $host := .NodeList }}
        <property name="Initial User Identity {{ $i }}">{{ $host }}</property>
{{- end }}
    </userGroupProvider>
    <userGroupProvider>
        <identifier>database-user-group-provider</identifier>
        <class>my.custom.DatabaseUserGroupProvider</class>
        <!-- Any extra configuration for this provider goes here -->
{{- range $i, $host := .NodeList }}
        <property name="Initial User Identity {{ $i }}">{{ $host }}</property>
{{- end }}
    </userGroupProvider>
    <accessPolicyProvider>
        <identifier>file-access-policy-provider</identifier>
        <class>org.apache.nifi.authorization.FileAccessPolicyProvider</class>
        <property name="User Group Provider">file-user-group-provider</property>
        <property name="Authorizations File">../data/authorizations.xml</property>
        <property name="Initial Admin Identity">{{ .ControllerUser }}</property>
        <property name="Legacy Authorized Users File"></property>
{{- range $i, $host := .NodeList }}
        <property name="Node Identity {{ $i }}">{{ $host }}</property>
{{- end }}
		<property name="Node Group"></property>
    </accessPolicyProvider>
    <accessPolicyProvider>
        <identifier>database-access-policy-provider</identifier>
        <class>my.custom.DatabaseAccessPolicyProvider</class>
        <!-- Any extra configuration for this provider goes here -->
{{- range $i, $host := .NodeList }}
        <property name="Node Identity {{ $i }}">{{ $host }}</property>
{{- end }}
		<property name="Node Group"></property>
    </accessPolicyProvider>
    <authorizer>
        <identifier>managed-authorizer</identifier>
        <class>org.apache.nifi.authorization.StandardManagedAuthorizer</class>
        <property name="Access Policy Provider">file-access-policy-provider</property>
    </authorizer>
    <authorizer>
        <identifier>custom-database-authorizer</identifier>
        <class>my.custom.DatabaseAuthorizer</class>
        <property name="Access Policy Provider">database-access-policy-provider</property>
    </authorizer>
</authorizers>
```

And finally, the NiFi property `nifi.security.user.authorizer` indicates which of the configured authorizers in the authorizers.xml file to use. Following the example, we'd set the property to:

```sh
nifi.security.user.authorizer=custom-database-authorizer
```

