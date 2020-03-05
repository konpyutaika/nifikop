<p align="center"><img src="docs/img/nifikop.png" width="160"></p>

<p align="center">
  <a href="https://hub.docker.com/r/orangeopensource/nifikop/">
    <img src="https://img.shields.io/docker/cloud/automated/orangeopensource/nifikop.svg" alt="Docker Automated build">
  </a>

  <a href="https://circleci.com/gh/orangeopensource/nifikop">
    <img src="https://circleci.com/gh/orangeopensource/nifikopr/tree/master.svg?style=shield" alt="CircleCI">
  </a>

  <a href="https://goreportcard.com/report/github.com/erdrix/nifikop">
    <img src="https://goreportcard.com/badge/github.com/erdrix/nifikop" alt="Go Report Card">
  </a>

  <a href="https://github.com/erdrix/nifikop/">
    <img src="https://img.shields.io/badge/license-Apache%20v2-orange.svg" alt="license">
  </a>
</p>

# NiFiKop

You can access to the full documentation on [NiFiKop Documentation website](https://kubernetes.pages.gitlab.si.francetelecom.fr/nifikop/)

The Orange NiFi operator is a Kubernetes operator to automate provisioning, management, autoscaling and operations of [Apache NiFi](https://nifi.apache.org/) clusters deployed to K8s.

## Overview

Apache NiFi is an open-source solution that support powerful and scalable directed graphs of data routing, transformation, and system mediation logic. 
Some of the high-level capabilities and objectives of Apache NiFi include, and some of the main features of the **NiFiKop** are:

- **Fine grained** node configuration support
- Graceful rolling upgrade
- graceful NiFi cluster **scaling**

Some of the roadmap features :

- Monitoring via **Prometheus**
- Automatic reaction and self healing based on alerts (plugin system, with meaningful default alert plugins)
- Encrypted communication using SSL
- graceful NiFi cluster **scaling and rebalancing**
- Advanced Dataflow and user management via CRD
- the provisioning of secure NiFi clusters

## Motivation

At [Orange](https://opensource.orange.com/fr/accueil/) we are building some [Kubernetes operator](https://github.com/erdrix?utf8=%E2%9C%93&q=operator&type=&language=), that operate NiFi and Cassandra clusters (among other types) for our business cases.

There are already some approaches to operating NiFi on Kubernetes, however, we did not find them appropriate for use in a highly dynamic environment, nor capable of meeting our needs.

- [Helm chart](https://github.com/cetic/helm-nifi)
- [Cloudera Nifi Operator](https://blog.cloudera.com/cloudera-flow-management-goes-cloud-native-with-apache-nifi-on-red-hat-openshift-kubernetes-platform/)

Finally, our motivation is to build an open source solution and a community which drives the innovation and features of this operator.


## Installation

The operator installs the 1.11.2 version of Apache NiFi, and can run on Minikube v0.33.1+ and Kubernetes 1.12.0+.

> The operator supports NiFi 1.11.0+

As a pre-requisite it needs a Kubernetes cluster. Also, NiFi requires Zookeeper so you need to first have a Zookeeper cluster if you don't already have one.

> We believe in the `separation of concerns` principle, thus the NiFi operator does not install nor manage Zookeeper.

### Install Zookeeper

To install Zookeeper we recommend using the [Pravega's Zookeeper Operator](https://github.com/pravega/zookeeper-operator).
You can deploy Zookeeper by using the Helm chart.

```bash
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com/
# Using helm3
# You have to create the namespace before executing following command
helm install zookeeper-operator --namespace=zookeeper banzaicloud-stable/zookeeper-operator
# Using previous versions of helm
helm install --name zookeeper-operator --namespace=zookeeper banzaicloud-stable/zookeeper-operator
kubectl create --namespace zookeeper -f - <<EOF
apiVersion: zookeeper.pravega.io/v1beta1
kind: ZookeeperCluster
metadata:
  name: zookeepercluster
  namespace: zookeeper
spec:
  replicas: 3
EOF
```

### Installation

We recommend to use a **custom StorageClass** to leverage the volume binding mode `WaitForFirstConsumer`

```bash
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: exampleStorageclass
parameters:
  type: pd-standard
provisioner: kubernetes.io/gce-pd
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```
> Remember to set your NiFiCluster CR properly to use the newly created StorageClass.

1. Set `KUBECONFIG` pointing towards your cluster
2. Run `make deploy` (deploys the operator in the current namespace into the cluster)
3. Set your NiFi configurations in a Kubernetes custom resource (sample: `config/samples/simplenificluster.yaml`) and run this command to deploy the NiFi components:

```bash
# Add your zookeeper svc name to the configuration
kubectl create -n nifi -f config/samples/simplenificluster.yaml
```

### Easy way: installing with Helm

Alternatively, if you are using Helm, you can deploy the operator using a Helm chart [Helm chart](https://github.com/erdrix/nifikop/tree/master/helm):

> To install the an other version of the operator use `helm install --name=nifikop --namespace=nifi --set operator.image.tag=x.y.z orange-incubator/nifikop`

```bash
helm repo add orange-incubator https://orange-kubernetes-charts-incubator.storage.googleapis.com/

# Using helm3
# You have to create the namespace before executing following command
helm install nifikop --namespace=nifi orange-incubator/nifikop
# Using previous versions of helm
helm install --name=nifikop --namespace=nifi orange-incubator/nifikop

# Add your zookeeper svc name to the configuration
kubectl create -n nifi -f config/samples/simplenificluster.yaml
```

## Test Your Deployment

## Development

Checkout out the [developer docs](docs/dev/developer_guide.md)

## Features

Check out the [supported features](docs/features.md)

## Issues, feature requests and roadmap

Please note that the NiFi operator is constantly under development and new releases might introduce breaking changes. We are striving to keep backward compatibility as much as possible while adding new features at a fast pace. Issues, new features or bugs are tracked on the projects [GitHub page](https://github.com/erdrix/nifikop/issues) - please feel free to add yours!

To track some of the significant features and future items from the roadmap please visit the [roadmap doc](docs/roadmap.md).

## Contributing 

If you find this project useful here's how you can help:

- Send a pull request with your new features and bug fixes
- Help new users with issues they may encounter
- Support the development of this project and star this repo!

## Community

If you have any questions about the NiFi operator, and would like to talk to us and the other members of the community, please join our [Slack](https://slack.nifikop.io/).

If you find this project useful, help us:

- Support the development of this project and star this repo! :star:
- If you use the Nifi operator in a production environment, add yourself to the list of production [adopters](https://github.com/erdrix/nifikop/blob/master/ADOPTERS.md). :metal: <br>
- Help new users with issues they may encounter :muscle:
- Send a pull request with your new features and bug fixes :rocket:

## Credits

- Operator implementation based on [banzaicloud/kafka-operator](https://github.com/banzaicloud/kafka-operator)
- NiFi kubernetes setup configuration inspired from [cetic/helm-nifi](https://github.com/cetic/helm-nifi)
- Implementation is based on [Operator SDK](https://github.com/operator-framework/operator-sdk)

## License

Copyright (c) 2019 [Orange, Inc.](https://opensource.orange.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.