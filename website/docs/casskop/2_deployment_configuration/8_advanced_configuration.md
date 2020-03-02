---
id: 8_advanced_configuration
title: Advanced configuration
sidebar_label: Advanced configuration
---

## Docker login for private registry

If you need to use a docker registry with authentication, then you will need to create a specific kubernetes secret with
theses informations.
Then you will configure the CRD with the secret name, so that it provides the data to each Statefulsets, which in
turn propagate it to each created Pod.

Create the secret :

```sh
kubectl create secret docker-registry yoursecretname \
  --docker-server=yourdockerregistry
  --docker-username=yourlogin \
  --docker-password=yourpass \
  --docker-email=yourloginemail
```

Then we will add a **imagePullSecrets** parameter in the CRD definition with value the name of the 
previously created secret. You can give several secrets :

```yaml
imagePullSecrets:
  name: yoursecretname
```


##Management of allowed Cassandra nodes disruption

CassKop makes uses of the kubernetes PodDisruptionBudget objetc to specify how many cassandra nodes disruption is
allowed on the cluster. By default, we only tolerate 1 disrupted pod at a time and will prevent to makes actions if
there is aloready an ongling disruption on the cluster.

In some edge cases it can be useful to make force the operator to continue it's actions even if there is already a
disruption ongoing. We can tune this by updating the `spec.maxPodUnavailable` parameter of the cassandracluster CRD.

> **IMPORTANT:** it is recommanded to not touch this parameter unless you know what you are doing.
