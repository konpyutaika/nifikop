---
id: 5_openshift_helm_modes
title: OpenShift Helm Modes
sidebar_label: OpenShift Helm Modes
---

Both the `nifikop` operator chart and the `nifi-cluster` chart support OpenShift via optional SecurityContextConstraints (SCC) resources.

## SCC mode

### Creating an SCC from the chart

Set `openshift.scc.create: true` (operator) or `cluster.openshift.scc.create: true` (cluster) to have the chart create a dedicated SCC and bind it to the workload service account.

The SCC is only rendered when the cluster exposes the `security.openshift.io/v1` API, so the same values file works on both vanilla Kubernetes and OpenShift.

### Using a pre-existing SCC

If your cluster already has an approved SCC managed by an administrator:

- Set `openshift.scc.create: false`
- Set `openshift.scc.existingName: <your-scc-name>`

The chart will skip SCC creation but still create a Role and RoleBinding granting the workload service account permission to `use` the named SCC.

## Service accounts

When `cluster.openshift.scc.create` or `cluster.openshift.scc.existingName` is set, the chart creates a dedicated service account for SCC binding (unless `cluster.manager=kubernetes`, in which case the manager service account is reused).

You can control the service account name via `cluster.openshift.scc.serviceAccount.name` or disable its creation with `cluster.openshift.scc.serviceAccount.create: false`.

## Examples

- `helm/nifikop/examples/values-openshift.yaml` — operator chart with SCC
- `helm/nifi-cluster/examples/values-openshift-scc.yaml` — cluster chart with SCC
