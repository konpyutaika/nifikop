---
id: 1_gke
title: Google Kubernetes Engine
sidebar_label: Google Kubernetes Engine
---

Follow these instructions to prepare a GKE cluster for NiFiKop

1. Setup environment variables.

```sh
export GCP_PROJECT=<project_id>
export GCP_ZONE=<zone>
export CLUSTER_NAME=<cluster-name>
```

2. Create a new cluster.

```sh
gcloud container clusters create $CLUSTER_NAME \
  --cluster-version latest \
  --machine-type=n1-standard-1 \
  --num-nodes 4 \
  --zone $GCP_ZONE \
  --project $GCP_PROJECT
```

3. Retrieve your credentials for `kubectl`.

```sh 
cloud container clusters get-credentials $CLUSTER_NAME \
    --zone $GCP_ZONE \
    --project $GCP_PROJECT
```

4. Grant cluster administrator (admin) permissions to the current user. To create the necessary RBAC rules for NiFiKop, the current user requires admin permissions.

```sh 
kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole=cluster-admin \
    --user=$(gcloud config get-value core/account)
```
