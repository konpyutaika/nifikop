---
id: 1_deploy_process_groups
title: Deploy process groups
sidebar_label: Deploy process groups
---

You can create NiFi process groups either:

* directly against the cluster through its REST API (using UI or some home made scripts), or
* via the `NifiResource` CRD.

To deploy a Process Group you must have a NifiCluster already configured, this is defined using the [NifiCluster] CRD which is assumed to have already been deployed.

This configuration is defined using the [NifiParameterContext] CRD, which NiFiKop will convert into a [Parameter context](https://nifi.apache.org/docs/nifi-docs/html/user-guide.html#parameter-contexts).


Below is an example of [NifiParameterContext]:

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiParameterContext
metadata:
  name: dataflow-lifecycle
  namespace: nifikop
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

Here is an example of secret that you can create that will be used by the configuration above:

```console
kubectl create secret generic secret-params \
    --from-literal=secret1=yop \
    --from-literal=secret2=yep \
    -n nifikop
```

:::warning
As a sensitive value cannot be retrieved through the Rest API, to update the value of a sensitive parameter, you have to:

- remove it from the secret
- wait for the next loop
- insert the parameter with the new value inside the secret

or you can simply create a new [NifiParameterContext] and refer it into your [NifiResource].
:::

You can now deploy your [NifiResource] by referencing the previous objects:

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiResource
metadata:
  name: processgroup
  namespace: nifikop
spec:
  parentProcessGroupID: "16cfd2ec-0174-1000-0000-00004b9b35cc"
  name: Process Group Instance
  type: process-group
  comments: Example Process Group
  configuration:
    position:
      posX: 0
      posY: 0
    parameterContextRef:
      name: dataflow-lifecycle
      namespace: nifikop
    updateStrategy: drain
  clusterRef:
    name: nc
    namespace: nifikop
```

To find details about the process group information required check the [official documentation](https://nifi.apache.org/docs/nifi-docs/html/user-guide.html#Configuring_a_ProcessGroup)

The Process Group then can be used by a [NifiDataflow] as its Parent Process Group.  You can use the below snippet to link together the [NifiResource] and the [NifiDataflow]

```yaml
apiVersion: nifi.konpyutaika.com/v1
kind: NifiDataflow
metadata:
  name: dataflow-lifecycle
  namespace: nifikop
spec:
  parentProcessGroupRef:
    name: processgroup
    namespace: nifikop
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
  updateStrategy: drain
```

[NifiParameterContext]: ../../5_references/4_nifi_parameter_context/
[NifiCluster]: ../../5_references/1_nifi_cluster/1_nifi_cluster