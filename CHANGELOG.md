## Unreleased

### Added

### Changed

### Deprecated

### Removed

### Bug Fixes

## v0.1.0-release

### Added

- [MR #10](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/10) - Implement TLS certificates creation with Cert-Manager
- [MR #10](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/10) - Add NifiUser custom resource for TLS users (nodes and operator)
- [MR #10](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/10) - Implement NifiUser and TLS reconciliation, with secrets injection
- [MR #10](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/10) - Add Initial Admin definition into NifCluster
- [MR #9](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/9) - Add NigoApi dependency
- [MR #9](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/9) - Implement HTTP Client wrapper for Operator
- [MR #10](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/10) - Implement multi-namespace watch logic
- [MR #10](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/10) - Documentation & tutorial on OpenId Connect configuration

### Changed

- [MR #9](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/9) - Improve rolling upgrade : on ready pods, not just running
- [MR #9](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/9) - Improve cluster task events filter
- [MR #9](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/9) - Improve Helm publication removing the cache
- [MR #10](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/10) - Move secure cluster configuration level (at Spec level)
- [MR #10](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/10) - Move repo reference to the GitLab one

### Deprecated

### Removed

### Bug Fixes

- [MR #7](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/7) - Decommission lifecycle on Coordinator node
- [MR #10](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/10) - HTTPS port selection

## v0.0.1-release

### Added

- Implement pod management lifecycle
- Implement Graceful downscale pod lifecycle management
- Implement Graceful upscale pod lifecycle management
- Implement configuration lifecycle management for : nifi.properties, zookeeper.properties, state-management.xml, login-identity-providers.xml, logback.xml, bootstrap.conf, bootstrap-notification-servces.xml
- Initiate documentations
- Implementation basic makefile for some action (debug, build, deploy, run, push, unit-test)
- Create helm chart for operator
- Add Documentation for internal deployment
- Add Gitlab CI Pipeline

### Changed

### Deprecated

### Removed

### Bug Fixes