---
id: 0_design_principles
title: Design Principles
sidebar_label: Design Principles
---

The [Dataflow Lifecycle management feature](../../1_concepts/3_features#dataflow-lifecycle-management-via-crd) introduces 3 new CRDs :

- **NiFiRegistryClient :** Allowing you to declare a [NiFi registry client](https://nifi.apache.org/docs/nifi-registry-docs/html/getting-started.html#connect-nifi-to-the-registry).
- **NiFiParameterContext :** Allowing you to create parameter context, with two kinds of parameters, a simple `map[string]string` for non-sensitive parameters and a `list of secrets` which contains sensitive parameters.
- **NiFiDataflow :** Allowing you to declare a Dataflow based on a `NiFiRegistryClient` and optionally a `ParameterContext`, which will be deployed and managed by the operator on the `targeted NiFi cluster`.

The following diagram shows the interactions between all the components :

![dataflow lifecycle management schema](/img/1_concepts/2_design_principes/dataflow_lifecycle_management_schema.jpg)

With each CRD comes a new controller, with a reconcile loop :

- **NiFiRegistryClient's controller :**

![NiFi registry client's reconcile loop](/img/1_concepts/2_design_principes/registry_client_reconcile_loop.jpeg)

- **NiFiParameterContext's controller :**

![NiFi parameter context's reconcile loop](/img/1_concepts/2_design_principes/parameter_context_reconcile_loop.jpeg)

- **NiFiDataflow's controller :**

![NiFi dataflow's reconcile loop](/img/1_concepts/2_design_principes/dataflow_reconcile_loop.jpeg)