## Unreleased

### Added

- [PR #55](https://github.com/Orange-OpenSource/nifikop/pull/55) - **[Operator/NifiCluster]** Add the ability to define additional sidecar to each NiFi nodes.
- [PR #59](https://github.com/Orange-OpenSource/nifikop/pull/59) - **[Operator/NifiCluster]** Give the ability to define kubernetes services to expose the NiFi cluster.
- [PR #55](https://github.com/Orange-OpenSource/nifikop/pull/55) - **[Documentation]** Upgdrade to Docusaurus 2.0.0-alpha70 and enable versioned feature.

### Changed

### Deprecated

### Removed

- [PR #59](https://github.com/Orange-OpenSource/nifikop/pull/59) - **[Operator/NifiCluster]** No more default service to expose the NiFi cluster, you have to explicitly define ExternalService.

### Fixed Bugs

- [PR #62](https://github.com/Orange-OpenSource/nifikop/pull/62) - **[Operator/NifiCluster]** Fix DNS names for services in all-node mode.

## v0.4.3

### Added

- [PR #53](https://github.com/Orange-OpenSource/nifikop/pull/53) - **[Operator/NifiUser]** Cert-manager integration can now be disabled (it's still required for secured cluster).

### Changed

- [PR #53](https://github.com/Orange-OpenSource/nifikop/pull/53) - **[Operator]** Upgrade operator-sdk from v0.18.0 to v.1.3.0, which upgrade k8s dependencies to 0.19.4 and migrate to Kubebuilder aligned project layout.
- [PR #53](https://github.com/Orange-OpenSource/nifikop/pull/53) - **[CI]** Update steps with new Makefile commands.

### Deprecated

- [PR #53](https://github.com/Orange-OpenSource/nifikop/pull/53) - **[Operator/CRD]** No more support for Kubernetes cluster under version 1.16 (we no longer provide crds in version v1beta1)

### Fixed Bugs

- [PR #53](https://github.com/Orange-OpenSource/nifikop/pull/53) - **[Operator]** Upgrade k8s dependencies to match with new version requirement : [#52](https://github.com/Orange-OpenSource/nifikop/issues/52) [#51](https://github.com/Orange-OpenSource/nifikop/issues/51) [#33](https://github.com/Orange-OpenSource/nifikop/issues/33) 
- [PR #53](https://github.com/Orange-OpenSource/nifikop/pull/53) - **[Operator]** Fix the users used into Reader user group
- [PR #53](https://github.com/Orange-OpenSource/nifikop/pull/53) - **[Documentation]** Fix the chart version informations : [#51](https://github.com/Orange-OpenSource/nifikop/issues/51)

## v0.4.2-alpha-release

- [PR #41](https://github.com/Orange-OpenSource/nifikop/pull/42) - **[Operator]** Access policies enum type list

### Added

- [PR #41](https://github.com/Orange-OpenSource/nifikop/pull/41) - **[Operator/NifiUser]** Manage NiFi's users into NiFi Cluster
- [PR #41](https://github.com/Orange-OpenSource/nifikop/pull/41) - **[Operator/NifiUserGroup]** Manage NiFi's user groups into NiFi Cluster
- [PR #41](https://github.com/Orange-OpenSource/nifikop/pull/41) - **[Operator]** Manage NiFi's access policies
- [PR #41](https://github.com/Orange-OpenSource/nifikop/pull/41) - **[Operator/NifiCluster]** Create three defaults groups : admins, readers, nodes
- [PR #41](https://github.com/Orange-OpenSource/nifikop/pull/41) - **[Operator/NifiCluster]** Add pod disruption budget support


### Changed

- [PR #41](https://github.com/Orange-OpenSource/nifikop/pull/41) - **[Helm Chart]** Add CRDs
- [PR #41](https://github.com/Orange-OpenSource/nifikop/pull/41) - **[Operator/NifiCluster]** Manage default process group id if not defined, using the root process group one.
- [PR #41](https://github.com/Orange-OpenSource/nifikop/pull/41) - **[Operator/NifiCluster]** Rename `zkAddresse` to `zkAddress`

### Removed

- [PR #41](https://github.com/Orange-OpenSource/nifikop/pull/41) - **[Operator/NifiCluster]** Remove `ClusterSecure` and `SiteToSiteSecure` by only checking if `SSLSecret` is set.


### Fixed Bugs

- [PR #30](https://github.com/Orange-OpenSource/nifikop/pull/40) - **[Documentation]** Fix getting started

## v0.3.1-release

### Fixed Bugs

- [PR #37](https://github.com/Orange-OpenSource/nifikop/pull/37) - **[Operator]** nifi.properties merge

## v0.3.0-release

### Added

- [PR #31](https://github.com/Orange-OpenSource/nifikop/pull/31) - **[Operator]** Dataflow lifecycle management

### Fixed Bugs

- [PR #30](https://github.com/Orange-OpenSource/nifikop/pull/30) - [Documentation] Fix slack link

## v0.2.1-release

### Added

- [PR #25](https://github.com/Orange-OpenSource/nifikop/pull/25) -  [Helm Chart] Add support for iterating over namespaces
- [PR #18](https://github.com/Orange-OpenSource/nifikop/pull/18) - [Operator] NiFiKop CRDs in version `v1beta1` of CustomResourceDefinition object.

### Changed

### Deprecated

### Removed

- [PR #18](https://github.com/Orange-OpenSource/nifikop/pull/18) - [Helm Chart] Remove CRDs deployment and append documentation.

### Fixed Bugs

- [PR #24](https://github.com/Orange-OpenSource/nifikop/pull/24) - [Documentation] Correct patterns & versions.
- [PR #28](https://github.com/Orange-OpenSource/nifikop/pull/28) - [Operator] PKI finalize bad namespace for ca cert.

## v0.2.0-release

### Added

- [MR #16](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/16) - Allow to override cluster domain
- [MR #16](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/16) - Support external DNS
- [MR #16](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/16) - Add `Spec.Service` field, allowing to add service's annotations.
- [MR #16](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/16) - Add `Spec.Pod` field, allowing to add pod's annotations.
- [MR #16](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/16) - Documentation & blog article about external dns and Let's encrypt usage.
- [MR #16](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/16) - Documentation on RicKaaStley deployment.
- [MR #16](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/16) - Improve test unit coverage.

### Changed

- [MR #17](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/17) - Upgrade dependencies
- [MR #17](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/17) - CRD generated under `apiextensions.k8s.io/v1`
- [MR #16](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/16) - Set binami zookeeper helm chart as recommended solution for 
ZooKeeper.
- [MR #16](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/16) - Improve terraform setup for articles.
- [MR #18](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/18) - Add ability to define if cert-manager is cluster scoped or not.
- [MR #18](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/18) - Open source changes

### Deprecated

- [MR #16](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/16) - Change `Spec.HeadlessServiceEnabled` to`Spec.Service.HeadlessEnabled`

### Removed

### Fixed Bugs

- [MR #13](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/13) - Fix scale out init scope in TLS cluster.

## v0.1.0-release

### Added

- [MR #10](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/10) - Implement TLS certificates creation with Cert-Manager
- [MR #10](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/10) - Add NifiUser custom resource for TLS users (nodes and operator)
- [MR #10](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/10) - Implement NifiUser and TLS reconciliation, with secrets injection
- [MR #10](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/10) - Add Initial Admin definition into NifCluster
- [MR #9](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/9) - Add NigoApi dependency
- [MR #9](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/9) - Implement HTTP Client wrapper for Operator
- [MR #10](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/10) - Implement multi-namespace watch logic
- [MR #10](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/10) - Documentation & tutorial on OpenId Connect configuration

### Changed

- [MR #9](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/9) - Improve rolling upgrade : on ready pods, not just running
- [MR #9](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/9) - Improve cluster task events filter
- [MR #9](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/9) - Improve Helm publication removing the cache
- [MR #10](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/10) - Move secure cluster configuration level (at Spec level)
- [MR #10](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/10) - Move repo reference to the GitLab one

### Deprecated

### Removed

### Fixed Bugs

- [MR #7](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/7) - Decommission lifecycle on Coordinator node
- [MR #10](https://github.com/Orange-OpenSource/nifikop/-/merge_requests/10) - HTTPS port selection

## v0.0.1-release

### Added

- Implement pod management lifecycle
- Implement Graceful downscale pod lifecycle management
- Implement Graceful upscale pod lifecycle management
- Implement configuration lifecycle management for : nifi.properties, zookeeper.properties, state-management.xml, login-identity-providers.xml, logback.xml, bootstrap.conf, bootstrap-notification-servces.xml
- Initiate documentations
- Implementation basic makefile for some actions (debug, build, deploy, run, push, unit-test)
- Create helm chart for operator
- Add Documentation for internal deployment
- Add Gitlab CI Pipeline

### Changed

### Deprecated

### Removed

### Fixed Bugs