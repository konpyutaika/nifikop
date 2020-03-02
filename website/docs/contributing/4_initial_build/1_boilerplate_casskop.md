---
id: 1_boilerplate_casskop
title: Boilerplate CassKop
sidebar_label: Boilerplate CassKop
---

We used the SDK to create the repository layout. This command is for memory ;) (or for applying sdk upgrades)

> You need to have first install the SDK.

```sh
#old version
 operator-sdk new casskop --api-version=db.orange.com/v1alpha1 --kind=CassandraCluster
#new version
operator-sdk new casskop --dep-manager=modules --repo=github.com.Orange-OpenSource/casskop --type=go
```

Then you want to add managers:

```sh
# Add a new API for the custom resource CassandraCluster
$ operator-sdk add api --api-version=db.orange.com/v1alpha1 --kind=CassandraCluster

# Add a new controller that watches for CassandraCluster
$ operator-sdk add controller --api-version=db.orange.com/v1alpha1 --kind=CassandraCluster
```