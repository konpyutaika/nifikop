---
id: 2_cassandra_config
title: Cassandra configuration
sidebar_label: Cassandra configuration
---

## Casskop >= 0.4.0

Starting with CassKop 0.4.0, there is no more obligation to rely on the CassKop Cassandra docker image.

The specific part of configuring Cassandra to works with CassKoçp has been mooved out from the image to a specific
init-container which is executed prior to the Cassandra image.

There is a specific bootstrap image that is build from the docker directory, and contains all required files or scripts
to works with CassKop.

### Cassandra bootstrap 

The configuration of the Cassandra image has been delegated to init-containers that are executed prior to the Cassandra
container when the pod is started.

### Initcontainer 1 : init-config

The init containers is responsible for the foloowing actions :
- copy the default Cassandra configuration from user's provided Cassandra image.

Configuration:
- The init-config container is by default the baseImage Cassandra image that can be changed using
`Spec.initContainerImage`.
- The default command executed by the init-container is `cp -vr /etc/cassandra/* /bootstrap"` and can be changed using
  `Spec.initContainerCmd`

### Initcontainer 2 : bootstrap

The bootstrap Container : 
- applying files and additional jar from the bootstrap image to the default configuration
- applying the user's configmap custom configuration (if any) on top of the default configuration
- modifying the configuration to be suitable to run with Casskop:
  - update cluster name 
  - configure dc/rack properties
  - applying seedlist
  - add cassadnra exporter and jolokia agent
  - ..

We provide the bootstrap image, but you can change it using `Spec.bootstrapImage` but you need to comply with the
required actions, see [Bootstrap](../docker/bootstrap).

## OLD Cassandra docker image

CassKop (< 0.4.0) relies on specific Docker Cassandra Image, where we provide specific startup script
that will be used to make live configuration on the Cassandra node we deploy.

The actual associated Docker image can be found in the gitlab repo :
[Cassandra for k8s](https://github.com/Orange-OpenSource/cassandra-image/).


You can manage the url and version of the docker image using the `spec.baseImage` and `spec.version` parameters of the
cassandracluster CRD.

> The Cassandra Image tag name must follow a specific semantic to allow CassKop to detect major
> changes in Cassandra version which may need to perform specific actions such as `nodetool upgradesstables`.

The actual naming of the Cassandra Image is as follow :

`<cassandra_version>-<java_version>-<git-tag-name>`

Example :
`3.11.3-8u201-0.3.0`

When Changing one of thoses fields, CassKop will triger an
[UpdateDockerImage](../documentation/operations.md#updatedockerimage)

## NodesPerRacks

One of the requirements for CassKop is to always keep the same number of nodes in each of it's racks, per Cassandra
DCs. The number of nodes used for the Cassandra Cluster is configured using the `CassandraCluster.spec.nodesPerRacks`
property.

> If you have not specify a Cluster Topology, then you'll have a default datacenter named `dc1`and a default rack named
> `rack1` 

It is a good practice for Cassandra to keep the same number of nodes in each Cassandra Racks. CassKop guarantees
that, but you can define different numbers of replicas for racks in different Cassandra DataCenters. 

This is done using the `nodesPerRacks` property in `CassandraCluster.spec.topology.dc[<idx>].nodesPerRacks`. If
specified on the datacenter level, this parameter takes priority over the global `CassandraCluster.spec.nodesPerRacks`.

Example:
example to scale up the nodesPerRacks in DC2 :

```yaml
  topology:
    dc:
      - name: dc1
        rack:
          - name: rack1
          - name: rack2
      - name: dc2
        nodesPerRacks: 3        <--- We increase by one this value
        rack:
          - name: rack1
```

> The number of Cassandra nodes will be the multiplication of the number of racks * the nodesPerRacks value.

If we changes on of theses properties then CassKop will triger either a [ScaleUp](../documentation/operations.md#scaleup)
or a [ScaleDown](../documentation/operations.md#scaledown) operation.

## Configuration override using configMap

CassKop allows you to customize the configuration of Apache Cassandra nodes by specifying a dedicated `ConfigMap`
name in the `CassandraCluster.spec.configMapName` containing configuration files to be overwritten above the default
configuration files of the dedicated Docker Image.

We have a specific Cassandra Docker image startup script that will overwrite each files in the directory
`/etc/cassandra`, from the one specified in the configMap if they exists.


>You can surcharge any files in the docker image `/etc/cassandra` with the ConfigMap.

Typical overwriting files may be :
- cassandra.yaml
- jvm.options
- specifying a pre_run.sh script

You can find an example with [this example](../samples/cassandra-configmap-v1.yaml)
```
$ kubectl apply -f config/cassandra-configmap-v1.yaml
configmap "cassandra-configmap-v1" created
```

Now you can add the `configMapName: cassandra-configmap-v1` to the Spec section of your CassandraCluster definition
[example](../samples/cassandracluster.yaml)

If you edit the ConfigMap it won't be detected neither by CassKop nor by the statefulsets/pods (unless you reboot the
pods). 
It is recommanded for configuration changes, to version the configmap and to create apply new configmap in the CRD, this
will trigger a rollingRestart of the whole cluster applying the new configuration.

> **IMPORTANT:** each time you specify a new configMap CassKop will start a `rollingUpdate` of all nodes
> in the cluster. more info on [UpdateConfigMap](../documentation/operations.md#updateconfigmap)

> **IMPORTANT:** At this time CassKop won't allow you to specify only excerpt of the configurations files, your
> ConfigMap **MUST** contain valid and complete configuration files 

## Configuration pre_run.sh script

In case you need to make some specific actions on a particular node, such as make uses of the **CASSANDRA_REPLACE_NODE**
variable, you can uses the pre_run.sh script in the ConfigMap. If present, the cassandra docker will execute this script
prior to the `run.sh` script from the docker image.

example of a configMap with the pre_run.sh script :

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cassandra-configmap-pre-run
data:
  pre_run.sh: |-
    echo "** this is a pre-scrip for run.sh that can be edit with configmap"
    test "$(hostname)" == 'cassandra-demo-dc1-rack1-0' && export CASSANDRA_REPLACE_NODE=10.233.93.174
    echo "** end of pre_run.sh script, continue with run.sh"
```

> **IMPORTANT:** In the case you use the configmap for one-time specific action, don't forget to edit again to remove
> the specific treatment once it is no more needed.

## JVM options

### Memory

Apache Cassandra are running inside of a Java Virtual Machine (JVM). JVM has many configuration options to optimize the
performance for different platforms and architectures. 

CassKop allows configuring theses values by adding a `jvm.options` in the user `ConfigMap`.

The default value used for `-Xmx` depends on whether there is a memory request configured for the container :
- If there is a memory request, the JVM's maximum memory must be set to a value corresponding to the limit.
- When there is no memory request, CassKop will limit it to "2048M".

- set the memory request and the memory limit to the same value, so that the pod is in guarantee mode

> CassKop will automatically compute the env var CASSANDRA_MAX_HEAP which is used to define `-Xms` and `-Xmx` in the
> `/run.sh` docker image script, from 1/4 of container Memory Limit.

### GarbageCollector output

We have a specific parameter in the CRD `spec.gcStdout: true/false` which specify if we wants to send the JVM garbage collector logs
in the stdout of the container, or inside a specific file in the container.

Default value is true, so it sends GC logs in stdout along with cassandra's logs.

## Authentication and authorizations

CassKop uses Jolokia from the cassandra-image to communicate. We can add
authentication on Jolokia by defining a secret :

Example:

```console
apiVersion: v1
kind: Secret
metadata:
  name: jolokia-auth
type: Opaque
data:
  password: TTBucDQ1NXcwcmQ=
  username: am9sb2tpYS11c2Vy
```

and in the CRD you wille define `spec.imageJolokiaSecret`

```console
...
  imageJolokiaSecret:
    name: jolokia-auth
...
```

CassKop will propagate the secrets in Cassandra so that it can configure
Jolokia, and uses it to connect.