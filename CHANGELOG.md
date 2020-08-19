## Unreleased

### Added

- [PR #25](https://github.com/Orange-OpenSource/nifikop/pull/25) -  [Helm Chart] Add support for iterating over namespaces

- [PR #18](https://github.com/Orange-OpenSource/nifikop/pull/18) - [Operator] NiFiKop CRDs in version `v1beta1` of CustomResourceDefinition object.

### Changed

### Deprecated

### Removed

- [PR #18](https://github.com/Orange-OpenSource/nifikop/pull/18) - [Helm Chart] Remove CRDs deployment and append documentation.

### Fixed Bugs

- [PR #24](https://github.com/Orange-OpenSource/nifikop/pull/24) - [Documentation] Correct patterns & versions.

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