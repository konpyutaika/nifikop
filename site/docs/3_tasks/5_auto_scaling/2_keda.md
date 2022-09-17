---
id: 2_keda
title: Auto scaling with KEDA
sidebar_label: Auto scaling with KEDA
---

## What is KEDA ?

[KEDA] is a Kubernetes-based Event Driven Autoscaler. With [KEDA], you can drive the scaling of any container in Kubernetes based on the number of events needing to be processed.

[KEDA] is a single-purpose and lightweight component that can be added into any Kubernetes cluster. [KEDA] works alongside standard Kubernetes components like the Horizontal Pod Autoscaler and can extend functionality without overwriting or duplication. With [KEDA] you can explicitly map the apps you want to use event-driven scale, with other apps continuing to function. This makes [KEDA] a flexible and safe option to run alongside any number of any other Kubernetes applications or frameworks.


## Deploying KEDA

Following the [documentation](https://keda.sh/docs/2.8/deploy/) here are the steps to deploy KEDA.


Deploying KEDA with Helm is very simple:

- Add Helm repo

````console
helm repo add kedacore https://kedacore.github.io/charts
````

- Update Helm repo

````console
helm repo update
````

- Install keda Helm chart

```console
kubectl create namespace keda
helm install keda kedacore/keda --namespace keda
```



:::important
More information about `Spec.UpdateStrategy` [here](../5_references/5_nifi_dataflow.md#dataflowupdatestrategy)
:::

[KEDA]: https://keda.sh/
[NifiDataflow]: ../5_references/5_nifi_dataflow.md
[NifiRegistryClient]: ../5_references/3_nifi_registry_client.md
[NifiParameterContext]: ../5_references/4_nifi_parameter_context.md