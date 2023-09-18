---
id: 1_deploy_connection
title: Deploy connection
sidebar_label: Deploy connection
---

You can create NiFi connections either :

* directly against the cluster through its REST API (using UI or some home made scripts), or
* via the `NifiConnection` CRD.

To deploy a [NifiConnection] you have to start by deploying at least 2 [NifiDataflows] because **NiFiKop** manages connection between 2 [NifiDataflows].

If you want more details about how to deploy [NifiDataflow], just have a look on the [how to deploy dataflow page](../3_manage_dataflows/1_deploy_dataflow).

Below is an example of 2 [NifiDataflows] named respectively `input` and `output`:

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiDataflow
metadata:
  name: input
  namespace: nifikop
spec:
  clusterRef:
    name: nc
    namespace: nifikop
  bucketId: deedb9f6-65a4-44e9-a1c9-722008fcd0ba
  flowId: ab95431d-980d-41bd-904a-fac4bd864ba6
  flowVersion: 1
  registryClientRef:
    name: registry-client-example
    namespace: nifikop
  skipInvalidComponent: true
  skipInvalidControllerService: true
  syncMode: always
  updateStrategy: drain
  flowPosition:
    posX: 0
    posY: 0
---
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiDataflow
metadata:
  name: output
  namespace: nifikop
spec:
  clusterRef:
    name: nc
    namespace: nifikop
  bucketId: deedb9f6-65a4-44e9-a1c9-722008fcd0ba
  flowId: fc5363eb-801e-432f-aa94-488838674d07
  flowVersion: 2
  registryClientRef:
    name: registry-client-example
    namespace: nifikop
  skipInvalidComponent: true
  skipInvalidControllerService: true
  syncMode: always
  updateStrategy: drain
  flowPosition:
    posX: 750
    posY: 0
```

We will obtain the following initial setup:
![Initial setup](/img/3_tasks/4_manage_connections/1_deploy_connections/initial_setup.jpg)

:::important
The `input` dataflow must have an `output port` and the `output` dataflow must have an `input port`.
:::

Now that we have 2 [NifiDataflows], we can connect them with a [NifiConnection].

Below is an example of a [NifiConnection] named `connection` between the 2 previously deployed dataflows:

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiConnection
metadata:
  name: connection
  namespace: nifikop
spec:
  source:
    name: input
    namespace: nifikop
    subName: output
    type: dataflow
  destination:
    name: output
    namespace: nifikop
    subName: input
    type: dataflow
  configuration:
    backPressureDataSizeThreshold: 100 GB
    backPressureObjectThreshold: 10000
    flowFileExpiration: 1 hour
    labelIndex: 0
    bends:
      - posX: 550
        posY: 550
      - posX: 550
        posY: 440
      - posX: 550
        posY: 88
  updateStrategy: drain
```

You will obtain the following setup:
![Connection setup](/img/3_tasks/4_manage_connections/1_deploy_connections/connection_setup.jpg)

The `prioritizers` field takes a list of prioritizers, and the order of the list matters in NiFi so it matters in the resource.

- `prioriters=[NewestFlowFileFirstPrioritizer, FirstInFirstOutPrioritizer, OldestFlowFileFirstPrioritizer]` ![Connection prioritizers 0](/img/3_tasks/4_manage_connections/1_deploy_connections/connection_prioritizers_0.jpg)
- `prioriters=[FirstInFirstOutPrioritizer, NewestFlowFileFirstPrioritizer, OldestFlowFileFirstPrioritizer]` ![Connection prioritizers 1](/img/3_tasks/4_manage_connections/1_deploy_connections/connection_prioritizers_0.jpg)
- `prioriters=[PriorityAttributePrioritizer]` ![Connection prioritizers 2](/img/3_tasks/4_manage_connections/1_deploy_connections/connection_prioritizers_0.jpg)

The `labelIndex` field will place the label of the connection according to the bends.
If we take the previous bending configuration, you will get this setup for these labelIndex:

- `labelIndex=0`: ![Connection labelIndex 0](/img/3_tasks/4_manage_connections/1_deploy_connections/connection_labelindex_0.jpg)
- `labelIndex=1`: ![Connection labelIndex 1](/img/3_tasks/4_manage_connections/1_deploy_connections/connection_labelindex_1.jpg)
- `labelIndex=2`: ![Connection labelIndex 2](/img/3_tasks/4_manage_connections/1_deploy_connections/connection_labelindex_2.jpg)

[NifiDataflow]: ../../5_references/5_nifi_dataflow
[NifiDataflows]: ../../5_references/5_nifi_dataflow
[NifiConnection]: ../../5_references/8_nifi_connection