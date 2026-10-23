---
id: 4_cluster_hibernation
title: Cluster hibernation
sidebar_label: Cluster hibernation
---

Hibernation stops every NiFi pod of a `NifiCluster` without deleting the cluster or its data. Use it to free
resources or pause processing for a while.

## Hibernate a cluster

Set `spec.hibernate` to `true`:

```bash
kubectl patch nificluster simplenifi --type merge -p '{"spec":{"hibernate":true}}'
```

The operator then:

- deletes every NiFi pod of the cluster;
- keeps every PVC, whatever the `reclaimPolicy` of its `storageConfigs`;
- keeps the node configuration secrets, the node services and the node status;
- stops all calls to the NiFi API for this cluster.

The cluster state moves to `ClusterHibernating` while pods are stopping, then to `ClusterHibernated`:

```bash
kubectl get nificluster simplenifi -o jsonpath='{.status.state}'
```

If a node upscale or downscale is in progress, hibernation waits for it to finish before stopping any pod. The
cluster keeps running meanwhile, and the operator records a `NifiClusterHibernationPending` event.

Each pod gets `terminationGracePeriodSeconds` (default `120`) to shut down. Its `preStop` hook runs
`nifi.sh stop`, which waits up to `graceful.shutdown.seconds` (default `20`) for NiFi to stop. Keep
`terminationGracePeriodSeconds` higher than `graceful.shutdown.seconds`.

## Wake up a cluster

Set `spec.hibernate` back to `false`:

```bash
kubectl patch nificluster simplenifi --type merge -p '{"spec":{"hibernate":false}}'
```

The operator recreates the pods, which reattach to their existing PVCs. The cluster state moves to
`ClusterInitialized`, then to `ClusterRunning` once all nodes are ready. Flows, queued FlowFiles and
repositories are restored from the volumes.

## Resources that depend on the cluster

`NifiDataflow`, `NifiUser`, `NifiUserGroup`, `NifiParameterContext`, `NifiRegistryClient`, `NifiConnection`
and `NifiResource` wait while the cluster is not ready. They resume after wake-up.

`NifiNodeGroupAutoscaler` does not change `spec.nodes` while the cluster is hibernated.

## Limitations

- Do not remove nodes from `spec.nodes` while the cluster is hibernated. Remove them after wake-up, so the
  operator can run the graceful downscale.
- Deleting a dependent resource (for example a `NifiDataflow`) while the cluster is hibernated blocks on its
  finalizer until the cluster is awake, because the operator must remove it from NiFi.

:::warning
Do not use `nodes: []` to stop a cluster. Removing nodes from `spec.nodes` is a scale-down: the operator deletes
the data PVCs of every removed node whose `reclaimPolicy` is `Delete`, which is the default.
:::
