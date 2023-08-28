---
id: 1_deploy_dataflow
title: Deploy dataflow
sidebar_label: Deploy dataflow
---

You can create NiFi dataflows either :

* directly against the cluster through its REST API (using UI or some home made scripts), or
* via the `NifiDataflow` CRD.

If you want more details about the design, just have a look on the [design page](./0_design_principles#dataflow-lifecycle-management)

To deploy a [NifiDataflow] you have to start by deploying a [NifiRegistryClient] because **NiFiKop** manages dataflow using the [NiFi Registry feature](https://nifi.apache.org/registry).

Below is an example of [NifiRegistryClient] :

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiRegistryClient
metadata:
  name: registry-client-example
  namespace: nifikop
spec:
  clusterRef:
    name: nc
    namespace: nifikop
  description: "Registry client managed by NiFiKop"
  uri: "http://nifi.hostname.com:18080"
```

Once you have deployed your [NifiRegistryClient], you have the possibility of defining a configuration that you will apply to your [NifiDataflow].

This configuration is defined using the [NifiParameterContext] CRD, which NiFiKop will convert into a [Parameter context](https://nifi.apache.org/docs/nifi-docs/html/user-guide.html#parameter-contexts).


Below is an example of [NifiParameterContext]:

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiParameterContext
metadata:
  name: dataflow-lifecycle
  namespace: demo
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
      value: toto
      description: toto
```

As you can see, in the [NifiParameterContext] you can refer to some secrets that will be converted into [sensitive parameter](https://nifi.apache.org/docs/nifi-docs/html/user-guide.html#using-parameters-with-sensitive-properties).

Here is an example of secret that you can create that will be used by the configuration above :

```console
kubectl create secret generic secret-params \
    --from-literal=secret1=yop \
    --from-literal=secret2=yep \
    -n nifikop
```

:::warning
As a sensitive value cannot be retrieved through the Rest API, to update the value of a sensitive parameter, you have to :

- remove it from the secret
- wait for the next loop
- insert the parameter with the new value inside the secret

or you can simply create a new [NifiParameterContext] and refer it into your [NifiDataflow].
:::

You can now deploy your [NifiDataflow] by referencing the previous objects :

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiDataflow
metadata:
  name: dataflow-lifecycle
spec:
  parentProcessGroupID: "16cfd2ec-0174-1000-0000-00004b9b35cc"
  bucketId: "01ced6cc-0378-4893-9403-f6c70d080d4f"
  flowId: "9b2fb465-fb45-49e7-94fe-45b16b642ac9"
  flowVersion: 2
  syncMode: always
  skipInvalidControllerService: true
  skipInvalidComponent: true
  clusterRef:
    name: nc
    namespace: nifikop
  registryClientRef:
    name: registry-client-example
    namespace: nifikop
  parameterContextRef:
    name: dataflow-lifecycle
    namespace: demo
  updateStrategy: drain
```

To find details about the versioned flow information required check the [official documentation](https://nifi.apache.org/docs/nifi-registry-docs/index.html)

You have two modes of control from your dataflow by the operator :

1 - `Spec.SyncMode == never` : The operator will deploy the dataflow as described in the resource, and never control it (unless you change the field to `always`). It is useful when you want to deploy your dataflow without starting it.

2 - `Spec.SyncMode == once` : The operator will deploy the dataflow as described in the resource, run it once, and never control it again (unless you change the field to `always`). It is useful when you want to deploy your dataflow in a dev environment, and you want to update the dataflow.

3 - `Spec.SyncMode == always` : The operator will deploy and ensure the dataflow lifecycle, it will avoid all manual modification directly from the Cluster (e.g remove the process group, remove the versioning, update the parent process group, make some local changes ...). If you want to perform update, rollback or stuff like this, you have to simply update the [NifiDataflow] resource.

:::important
More information about `Spec.UpdateStrategy` [here](../../5_references/5_nifi_dataflow#componentupdatestrategy)
:::

[NifiDataflow]: ../../5_references/5_nifi_dataflow
[NifiRegistryClient]: ../../5_references/3_nifi_registry_client
[NifiParameterContext]: ../../5_references/4_nifi_parameter_context