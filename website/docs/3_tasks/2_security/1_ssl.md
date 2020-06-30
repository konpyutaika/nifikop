---
id: 1_ssl
title: Securing NiFi with SSL
sidebar_label: SSL
---

The `NiFi operator` makes securing your NiFi cluster with SSL. You may provide your own certificates, or instruct the operator to create them for from your cluster configuration.

Below this is an example configuration required to secure your cluster with SSL : 

```yaml
apiVersion: nifi.orange.com/v1alpha1
kind: NifiCluster
...
spec:
  ...
  clusterSecure: true
  siteToSiteSecure: true
  initialAdminUser: aguitton.ext@orange.com
  ...
  readOnlyConfig:
    # NifiProperties configuration that will be applied to the node.
    nifiProperties:
      webProxyHosts:
        - nifistandard2.trycatchlearn.fr:8443
        ...
  ...
  listenersConfig:
    internalListeners:
      - type: "https"
        name: "https"
        containerPort: 8443
      - type: "cluster"
        name: "cluster"
        containerPort: 6007
      - type: "s2s"
        name: "s2s"
        containerPort: 10000
    sslSecrets:
      tlsSecretName: "test-nifikop"
      create: true
```

- `clusterSecure` : cluster nodes secure mode : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#cluster_common_properties.
- `siteToSiteSecure` : site to Site properties Secure mode : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#site_to_site_properties.
- `initialAdminUser` : name of the user account which will be configured as initial admin into NiFi cluster : https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#initial-admin-identity
- `readOnlyConfig.nifiProperties.webProxyHosts` : A list of allowed HTTP Host header values to consider when NiFi is running securely and will be receiving requests to a different host[:port] than it is bound to. [web-properties](https://nifi.apache.org/docs/nifi-docs/html/administration-guide.html#web-properties)

If `listenersConfig.sslSecrets.create` is set to `false`, the operator will look for the secret at `listenersConfig.sslSecrets.tlsSecretName` and expect these values :

| key | value |
|-----|-------|
| caCert | The CA certificate |
| caKey | The CA private key |
| clientCert | A client certificate (this will be used by operator for NiFI operations) |
| clientKey | The private key for clientCert |

## Using an existing Issuer

As described in the [Reference section](/nifikop/docs/5_references/1_nifi_cluster/6_listeners_config#sslsecrets), instead of using a self-signed certificate as CA, you can use an existing one.
In order to do so, you only have to refer it into your `Spec.ListenerConfig.SslSecrets.IssuerRef` field. 

### Example : Let's encrypt 

Let's say you have an existing DNS server, with [external dns](https://github.com/kubernetes-sigs/external-dns) deployed into your cluster's namespace.
You can easily use Let's encrypt as authority for your certificate. 

To do this, you have to : 

1. Create an issuer : 

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Issuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # You must replace this email address with your own.
    # Let's Encrypt will use this to contact you about expiring
    # certificates, and issues related to your account.
    email: <your email address>
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      # Secret resource used to store the account's private key.
      name: example-issuer-account-key
    # Add a single challenge solver, HTTP01 using nginx
    solvers:
      - http01:
          ingress:
            ingressTemplate:
              metadata:
                annotations:
                  "external-dns.alpha.kubernetes.io/ttl": "5"
```

2. Setup External dns and correctly create your issuer into your cluster configuration : 

```yaml 
apiVersion: nifi.orange.com/v1alpha1
kind: NifiCluster
...
spec:
  ...
  clusterSecure: true
  siteToSiteSecure: true
  ...
  listenersConfig:
    clusterDomain: <DNS zone name>
    useExternalDNS: true
    ...
    sslSecrets:
      tlsSecretName: "test-nifikop"
      create: true
      issuerRef:
        name: letsencrypt-staging
        kind: Issuer
```

## Create SSL credentials

You may use `NifiUser` resource to create new certificates for your applications, allowing them to query your Nifi cluster.

To create a new client you will need to generate new certificates sign by the CA. The operator can automate this for you using the `NifiUser` CRD : 

```console
cat << EOF | kubectl apply -n nifi -f -
apiVersion:  nifi.orange.com/v1alpha1
kind: NifiUser
metadata:
  name: example-client
  namespace: nifi
spec:
  clusterRef:
    name: nifi
  secretName: example-client-secret
EOF
```

This will create a user and store its credentials in the secret `example-client-secret`. The secret contains these fields : 

| key | value |
|-----|-------|
| ca.crt | The CA certificate |
| tls.crt | The user certificate |
| tls.key | The user private key |

You can then mount these secret to your pod. Alternatively, you can write them to your local machine by running:

```console
kubectl get secret example-client-secret -o jsonpath="{['data']['ca\.crt']}" | base64 -d > ca.crt
kubectl get secret example-client-secret -o jsonpath="{['data']['tls\.crt']}" | base64 -d > tls.crt
kubectl get secret example-client-secret -o jsonpath="{['data']['tls\.key']}" | base64 -d > tls.key
```

The operator can also include a Java keystore format (JKS) with your user secret if you'd like. Add `includeJKS`: `true` to the `spec` like shown above, and then the user-secret will gain these additional fields :

| key | value |
|-----|-------|
| tls.jks | The java keystore containing both the user keys and the CA (use this for your keystore AND truststore) |
| pass.txt | The password to decrypt the JKS (this will be randomly generated) |

## Current limitations

### Operator access policies

For the current version of the operator, the access policies are still not managed (refer to the [Roadmap page](/nifikop/docs/1_concepts/4_roadmap#authentification-management)).
That's why when you deploy a cluster in a secure mode, you have to define an `Initial admin`, which is a user account, that you will use to access to the cluster and add the `view` and `modify` access to the `access the controller` policy.

If you check the controller log you should have the following error : 

```json
{
  "level":"error",
  "ts":1589974514.048613,
  "logger":"nifi_client",
  "msg":"Error during talking to nifi node",
  "error":"403 Forbidden",
  "stacktrace":"github.com/go-logr/zapr.(*zapLogger).Error\n\t/go/pkg/mod/github.com/go-logr/zapr@v0.1.1/zapr.go:128\ngitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/nificlient.(*nifiClient).DescribeCluster\n\tnifikop/pkg/nificlient/system.go:22\ngitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/nificlient.(*nifiClient).Build\n\tnifikop/pkg/nificlient/client.go:70\ngitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/nificlient.NewFromCluster\n\tnifikop/pkg/nificlient/client.go:92\ngitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/controller/common.NewNodeConnection\n\tnifikop/pkg/controller/common/controller_common.go:68\ngitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/scale.EnsureRemovedNodes\n\tnifikop/pkg/scale/scale.go:210\ngitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/resources/nifi.(*Reconciler).Reconcile\n\tnifikop/pkg/resources/nifi/nifi.go:186\ngitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/controller/nificluster.(*ReconcileNifiCluster).Reconcile\n\tnifikop/pkg/controller/nificluster/nificluster_controller.go:146\nsigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).reconcileHandler\n\t/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.4.0/pkg/internal/controller/controller.go:256\nsigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem\n\t/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.4.0/pkg/internal/controller/controller.go:232\nsigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).worker\n\t/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.4.0/pkg/internal/controller/controller.go:211\nk8s.io/apimachinery/pkg/util/wait.JitterUntil.func1\n\t/go/pkg/mod/k8s.io/apimachinery@v0.0.0-20191004115801-a2eda9f80ab8/pkg/util/wait/wait.go:152\nk8s.io/apimachinery/pkg/util/wait.JitterUntil\n\t/go/pkg/mod/k8s.io/apimachinery@v0.0.0-20191004115801-a2eda9f80ab8/pkg/util/wait/wait.go:153\nk8s.io/apimachinery/pkg/util/wait.Until\n\t/go/pkg/mod/k8s.io/apimachinery@v0.0.0-20191004115801-a2eda9f80ab8/pkg/util/wait/wait.go:88"
}
```

1. To do this, you have to connect on your cluster and go into the users page : 

![users page](/nifikop/img/3_tasks/2_security/1_ssl/users.png)

2. Create a new user entry with `<NifiCluster's name>-controller.<NifiCluster's namespace>.mgt.cluster.local` name, this the name of the `NifiUser` associated to the operator : 
For example for the the `NifiCluster` **sslnifi** deployed into the namespace **nifi** we have the following configuration : 

![add controller's user](/nifikop/img/3_tasks/2_security/1_ssl/add_user_controller.png)

3. Now we will add the required policies, so go to the `Access Policies` page : 

![Access policies page](/nifikop/img/3_tasks/2_security/1_ssl/access_policies_page.png)

4. And add the operator's user the right to `view` and `modify` to the `access the controller` policy.

![Access Controller view](/nifikop/img/3_tasks/2_security/1_ssl/access_conrtoller_view.png)

![Access Controller modify](/nifikop/img/3_tasks/2_security/1_ssl/access_conrtoller_modify.png)

If you check once again the logs, you no longer have the error.

### Scale up - Node declaration

For the moment the operator is not able to add a new user inside NiFi cluster. So when you scale up the cluster, the node will join the cluster but will not create the user and associated policies.
This will be covered by the [issue #9](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/issues/9), while waiting for this feature, you have to :

1. Go on users page : 

![users page](/nifikop/img/3_tasks/2_security/1_ssl/users.png)

2. Create a new user entry with `<NifiCluster's name>-<NiFi node id>-node.<NifiCluster's name>-headless.<NifiCluster's namespace>` name if you setup **Headless**, else use `<NifiCluster's name>-<NiFi node id>-node.<NifiCluster's namespace>` as name : 

![add node's user](/nifikop/img/3_tasks/2_security/1_ssl/add_node_user.png)

3. The node require `proxy user requests` access policy, so go to the `Access Policies` page : 

![Access policies page](/nifikop/img/3_tasks/2_security/1_ssl/access_policies_page.png)

4. Add the node user to this policy : 

![Access policy Node](/nifikop/img/3_tasks/2_security/1_ssl/access_policy_node.png)