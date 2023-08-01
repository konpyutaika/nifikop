<p align="center"><img src="docs/img/nifikop.png" width="160"></p>

<p align="center">
  <a href="https://github.com/konpyutaika/nifikop/pkgs/container/docker-images%2Fnifikop">
    <img src="https://img.shields.io/github/v/release/konpyutaika/nifikop?style=shield&logo=docker" alt="Docker Automated build">
  </a>

  <a href="https://github.com/konpyutaika/nifikop/releases">
    <img src="https://img.shields.io/github/v/release/konpyutaika/nifikop?style=shield&logo=github" alt="Github Releases">
  </a>

  <a href="https://join.slack.com/t/konpytika/shared_invite/zt-1f0z06xv7-VjnKwrUBcyE2I_BwOukIUQ">
    <img src="https://img.shields.io/badge/slack-nifikop-brightgreen.svg?style=shield&logo=slack" alt="Slack chat">
  </a>

  <a href="https://circleci.com/gh/konpyutaika/nifikop">
    <img src="https://img.shields.io/circleci/build/github/konpyutaika/nifikop/master?logo=circleci&style=shield" alt="CircleCI">
  </a>

  <a href="https://goreportcard.com/report/github.com/konpyutaika/nifikop">
    <img src="https://goreportcard.com/badge/github.com/konpyutaika/nifikop?style=shield&logo=go" alt="Go Report Card">
  </a>

  <a href="https://github.com/konpyutaika/nifikop/">
    <img src="https://img.shields.io/badge/license-Apache%20v2-orange.svg?style=shield&logo=apache" alt="license">
  </a>
</p>

# NiFiKop

You can access to the full documentation on the [NiFiKop Documentation](https://konpyutaika.github.io/nifikop/)

The Konpyūtāika NiFi operator is a Kubernetes operator to automate provisioning, management, autoscaling and operations of [Apache NiFi](https://nifi.apache.org/) clusters deployed to K8s.

## Overview

Apache NiFi is an open-source solution that supports powerful and scalable directed graphs of data routing, transformation, and system mediation logic. 
Some of the high-level capabilities and objectives of Apache NiFi include, and some of the main features of the **NiFiKop** are:

- **Fine grained** node configuration support
- Graceful rolling upgrade
- graceful NiFi cluster **scaling**
- encrypted communication using SSL
- the provisioning of secure NiFi clusters
- Advanced Dataflow and user management via CRD

Some of the roadmap features :

- Monitoring via **Prometheus**
- Automatic reaction and self healing based on alerts (plugin system, with meaningful default alert plugins)
- graceful NiFi cluster **scaling and rebalancing**

## Motivation

There are already some approaches to operating NiFi on Kubernetes, however, we did not find them appropriate for use in a highly dynamic environment, nor capable of meeting our needs.

- [Helm chart](https://github.com/cetic/helm-nifi)
- [Cloudera Nifi Operator](https://blog.cloudera.com/cloudera-flow-management-goes-cloud-native-with-apache-nifi-on-red-hat-openshift-kubernetes-platform/)

Finally, our motivation is to build an open source solution and a community which drives the innovation and features of this operator.

## Installation

To get up and running quickly, check our [Getting Started page](https://konpyutaika.github.io/nifikop/docs/2_deploy_nifikop/1_quick_start)

## Development

Checkout out the [Developer page](https://konpyutaika.github.io/nifikop/docs/6_contributing/1_developer_guide)

## Features

Check out the [Supported Features Page](https://konpyutaika.github.io/nifikop/docs/1_concepts/3_features)

## Issues, feature requests and roadmap

Please note that the NiFi operator is constantly under development and new releases might introduce breaking changes. We are striving to keep backward compatibility as much as possible while adding new features at a fast pace. Issues, new features or bugs are tracked on the projects [GitHub page](https://github.com/konpyutaika/nifikop/issues) - please feel free to add yours!

To track some of the significant features and future items from the roadmap please visit the [roadmap doc](https://konpyutaika.github.io/nifikop/docs/1_concepts/4_roadmap).

## Contributing 

If you find this project useful here's how you can help:

- Send a pull request with your new features and bug fixes
- Help new users with issues they may encounter
- Support the development of this project and star this repo!

## Community

If you have any questions about the NiFi operator, and would like to talk to us and the other members of the community, please join our [Slack](https://join.slack.com/t/konpytika/shared_invite/zt-14md072lv-Jr8mqYoeUrqzfZF~YGUpXA).

If you find this project useful, help us:

- Support the development of this project and star this repo! :star:
- If you use the Nifi operator in a production environment, add yourself to the list of production [adopters](ADOPTERS.md). :metal: <br>
- Help new users with issues they may encounter :muscle:
- Send a pull request with your new features and bug fixes :rocket:

## Companies

If you are a company that uses or contributes to the Operator and you would like to help the operator to get more visibility, 
you can create a Pull Request to add you to the [list of companies](https://github.com/konpyutaika/nifikop/blob/master/site/website/src/pages/index.js#L104) working on or with NiFiKop !

## Credits

- Operator implementation based on [banzaicloud/kafka-operator](https://github.com/banzaicloud/kafka-operator)
- NiFi kubernetes setup configuration inspired from [cetic/helm-nifi](https://github.com/cetic/helm-nifi)
- Implementation is based on [Operator SDK](https://github.com/operator-framework/operator-sdk)

## License

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
