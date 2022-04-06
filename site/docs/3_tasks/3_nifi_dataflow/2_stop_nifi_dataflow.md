---
id: 2_stop_nifi_dataflow
title: Stop NiFi Dataflows
sidebar_label: Stop NiFi Dataflows
---

NiFi Dataflows are by default started if the `syncMode` is not set to `never`. But sometimes, you want to stop your dataflow and still keep it in sync.
To allow this, the label `isStopped` can be used. If this label is set to `true` and the dataflow is in sync, the operator will stop the processors and controller services of the dataflow. Once the label, the user will no longer be able to start the dataflow, the operator will ensure this.

Here is an example of how to set this label via CLI :

```sh
kubectl label nifidataflows dataflow isStopped=true --overwrite
```

Here is an example of how to set this label via the Kubernetes resource :

```yaml
apiVersion: nifi.konpyutaika.com/v1alpha1
kind: NifiDataflow
metadata:
  name: dataflow-lifecycle
  labels:
    isStopped: "true"
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
