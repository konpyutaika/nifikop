## Unreleased

### Added

- Implement TLS certificates creation with Cert-Manager
- Add NifiUser custom resource for TLS users (nodes and operator)
- Implement NifiUser and TLS reconciliation, with secrets injection
- Add Initial Admin definition into NifCluster
- Add NioApi dependency
- Implement HTTP Client wrapper for Operator
- Implement multi-namespace watch logic

### Changed

- Improve rolling upgrade : on ready pods, not just running
- Improve cluster task events filter
- Improve Helm publication removing the cache
- Move secure cluster configuration level (at Spec level)

### Deprecated

### Removed

### Bug Fixes

- [MR #7](https://gitlab.si.francetelecom.fr/kubernetes/nifikop/-/merge_requests/7) - Decommission lifecycle on Coordinator node
- HTTPS port selection

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