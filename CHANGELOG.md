## Unreleased

### Added

### Changed

- [PR #340](https://github.com/konpyutaika/nifikop/pull/340) - **[Operator/NifiDataflow]** Updated the logic to stop the entire dataflow instead of just the processors when the parameter context reference is updated.

### Fixed Bugs

### Deprecated

### Removed

## v1.6.0

### Added

- [PR #325](https://github.com/konpyutaika/nifikop/pull/325) - **[Operator/NifiCluster]** Added ability to configure `ReclaimPolicy` for `StorageConfig` persistent volumes.
- [PR #335](https://github.com/konpyutaika/nifikop/pull/335) - **[Operator/NifiCluster]** Added ability to set port protocol and load balancer class for external services via `ExternalServiceConfig`.
- [PR #333](https://github.com/konpyutaika/nifikop/pull/333) - **[Operator]** Replace Update by Patch on K8S resource to avoid update race conditions.

### Changed

- [PR #327](https://github.com/konpyutaika/nifikop/pull/327) - **[Documentation]** Update Nifikop vs Apache NiFi compatibility matrix documentation.
- [PR #330](https://github.com/konpyutaika/nifikop/pull/330) - **[Operator]** Upgrade golang to 1.21.5.
- [PR #337](https://github.com/konpyutaika/nifikop/pull/337) - **[NiGoApi]** Upgrade NiGoApi to v0.0.10.

### Fixed Bugs

- [PR #336](https://github.com/konpyutaika/nifikop/pull/336) - **[Operator/NifiCluster]** Fixed issue where nifikop wouldn't update `storageConfigs.metadata.annotations` if they were changed after initial creation.

## v1.5.0

### Added

- [PR #320](https://github.com/konpyutaika/nifikop/pull/320) - **[Operator/NifiCluster]** Added ability to set NiFi container port protocol via `InternalListenersConfig`.
- [PR #323](https://github.com/konpyutaika/nifikop/pull/323) - **[Operator/NifiCluster]** Added abitility to set `nodePort`.

### Changed

- [PR #307](https://github.com/konpyutaika/nifikop/pull/307) - **[Operator]** Upgrade golang to 1.21.2.
- [PR #314](https://github.com/konpyutaika/nifikop/pull/314) - **[Operator]** Upgrade golang to 1.21.3.
- [PR #314](https://github.com/konpyutaika/nifikop/pull/314) - **[Documentation]** Upgrade node 20.9.0, Docusaurus 3, React 18 and other dependencies.
- [PR #321](https://github.com/konpyutaika/nifikop/pull/321) - **[Operator]** Upgrade golang to 1.21.4.

### Fixed Bugs

- [PR #311](https://github.com/konpyutaika/nifikop/pull/311) - **[Operator/NifiConnection]** Doesn't have its own requeue interval.
- [PR #322](https://github.com/konpyutaika/nifikop/pull/322) - **[Helm Chart]** SingleUserConfiguration block in Helm chart refers to incorrect values and has incorrect indentation.

## v1.4.1

### Fixed Bugs

- [PR #302](https://github.com/konpyutaika/nifikop/pull/302) - **[Operator/NifiNodeGroupAutoscaler]** Set default namespace in NifiNodeGroupAutoscaler's clusterRef.
- [PR #305](https://github.com/konpyutaika/nifikop/pull/305) - **[Operator/NifiRegistryClient]** Registry client URI can't be updated.

## v1.4.0

### Added

- [PR #291](https://github.com/konpyutaika/nifikop/pull/291) - **[Plugin]** Implementation of NiFiKop's plugin.
- [PR #291](https://github.com/konpyutaika/nifikop/pull/291) - **[Operator/NifiConnection]** Implementation of NifiConnection controller.
- [PR #292](https://github.com/konpyutaika/nifikop/pull/292) - **[Operator/NifiCluster]** Modify RBAC kubebuilder annotations so NiFiKop works on OpenShift.
- [PR #292](https://github.com/konpyutaika/nifikop/pull/292) - **[Helm Chart]** Add Parameter for RunAsUser for OpenShift.
- [PR #300](https://github.com/konpyutaika/nifikop/pull/300) - **[Operator/NifiCluster]** Manage no node cluster.
- [PR #300](https://github.com/konpyutaika/nifikop/pull/300) - **[Operator/NifiNodeGroupAutoscaler]** Manage autoscale to 0.

### Changed

- [PR #290](https://github.com/konpyutaika/nifikop/pull/290) - **[Operator/NifiCluster]** Change default sensitive algorithm.
- [PR #295](https://github.com/konpyutaika/nifikop/pull/295) - **[Operator]** Upgrade golang to 1.21.1.
- [PR #300](https://github.com/konpyutaika/nifikop/pull/300) - **[Operator/NifiNodeGroupAutoscaler]** Change new nodes id computation.

### Fixed Bugs

- [PR #300](https://github.com/konpyutaika/nifikop/pull/300) - **[Operator/NifiNodeGroupAutoscaler]** Empty `CreationTime` in node states crashes the operator.

## v1.3.1

### Added

- [PR #286](https://github.com/konpyutaika/nifikop/pull/286) - **[Operator/NifiCluster]** Update resource's status only on change.

### Changed

- [PR #287](https://github.com/konpyutaika/nifikop/pull/287) - **[NiGoApi]** Upgrade nigoapi to v0.0.9.
- [PR #288](https://github.com/konpyutaika/nifikop/pull/288) - **[Operator/NifiCluster]** Block user and user group creation in NiFi with pure single user authentication.

### Fixed Bugs

- [PR #288](https://github.com/konpyutaika/nifikop/pull/288) - **[Operator/NifiCluster]** Fix single user authentication default secret keys.

## v1.3.0

### Added

- [PR #278](https://github.com/konpyutaika/nifikop/pull/278) - **[Operator/NifiCluster]** Added the single-user-authentication method.
- [PR #279](https://github.com/konpyutaika/nifikop/pull/279) - **[Operator/NifDataflow]** Added check to control if a dataflow is unscheduled in order to schedule it.

### Changed

- [PR #281](https://github.com/konpyutaika/nifikop/pull/281) - **[Operator]** Upgrade golang to 1.21.0.

### Fixed Bugs

- [PR #279](https://github.com/konpyutaika/nifikop/pull/279) - **[Operator/NifDataflow]** Trim parameter's description from parameter context.

## v1.2.0

### Added

- [PR #258](https://github.com/konpyutaika/nifikop/pull/141) - **[Helm Chart]** Upgraded helm-deployed HPA to v2 and added flowPosition to NiFiDataflow.
- [PR #269](https://github.com/konpyutaika/nifikop/pull/269) - **[Operator/NifiCluster]** Added ability to attach labels and annotations to PVCs that nifikop creates.

### Changed

- [PR #257](https://github.com/konpyutaika/nifikop/pull/257) - **[Operator]** Updated the operator-sdk to 1.28.0.
- [PR #263](https://github.com/konpyutaika/nifikop/pull/263) - **[NiGoApi]** Upgrade nigoapi to v0.0.8.
- [PR #263](https://github.com/konpyutaika/nifikop/pull/268) - **[Operator]** Upgrade golang to 1.20.5.
- [PR #266](https://github.com/konpyutaika/nifikop/pull/266) - **[Operator]** Add AuthenticationStrategy, ManagerDn, ManagerPassword, IdentityStrategy properties for LDAP integration.
- [PR #276](https://github.com/konpyutaika/nifikop/pull/276) - **[Operator]** Upgrade golang to 1.20.6.

## v1.1.1

### Added
- [PR #244](https://github.com/konpyutaika/nifikop/pull/244) - **[Operator]** Updated the go version in nifikop to 1.20.
- [PR #141](https://github.com/konpyutaika/nifikop/pull/141) - **[Helm Chart]** Added nifi-cluster helm chart.

### Fixed Bugs

- [PR #243](https://github.com/konpyutaika/nifikop/pull/243) - **[Operator]** Re-Fixed bug where an incorrect condition was used to determine whether or not to substitute a custom authorizers template.
- [PR #245](https://github.com/konpyutaika/nifikop/pull/245) - **[Operator]** Added staticcheck linting and go vuln scanning to Makefile. Fixed all linting issues with operator

## v1.1.0

### Added

- [PR #220](https://github.com/konpyutaika/nifikop/pull/220) - **[Operator/NifiCluster]** Made `Pod` readiness and liveness checks configurable.
- [PR #218](https://github.com/konpyutaika/nifikop/pull/218) - **[Operator]** Add cross-platform support to nifikop docker image.

### Changed

- [PR #236](https://github.com/konpyutaika/nifikop/pull/236) - **[Operator]** Fixed issue where operator would infinitely retry requests if it cannot find `Dataflow`/`ParameterContext` update & drop requests.

### Fixed Bugs

- [PR #223](https://github.com/konpyutaika/nifikop/pull/223) - **[Operator]** Fixed bug where an incorrect condition was used to determine whether or not to substitute a custom authorizers template.

## v1.0.0

### Changed

- [PR #190](https://github.com/konpyutaika/nifikop/pull/190) - **[CRDs]** Migrate v1alpha1 to v1

## v0.16.0

### Added

- [PR #202](https://github.com/konpyutaika/nifikop/pull/202) - **[Operator]** Updated the go version in nifikop to 1.19.
- [PR #208](https://github.com/konpyutaika/nifikop/pull/208) - **[Operator]** Updated the cert-manager lib version to v1.10.0.

### Changed

- [PR #205](https://github.com/konpyutaika/nifikop/pull/205) - **[Operator]** Updated operator-sdk to v1.25.2.

### Fixed Bugs

- [PR #195](https://github.com/konpyutaika/nifikop/pull/195) - **[Helm Chart]** Fixed bug where default metrics port collided with default health probe port.
- [PR #210](https://github.com/konpyutaika/nifikop/pull/210) - **[NifiUser]** Fixed issue where `NifiUser` `Certificate` and `Secret` resources get re-created after the `NifiUser` has been marked for deletion and removed. This is most noticeable when deploying NiFi clusters via ArgoCD.

## v0.15.0

### Added

- [PR #165](https://github.com/konpyutaika/nifikop/pull/165) - **[NifiParameterContext]** Add parameter context inheritance.

### Changed

- [PR #165](https://github.com/konpyutaika/nifikop/pull/165) - **[NiGoApi]** Update NiGoApi dependence.

### Fixed Bugs

- [PR #189](https://github.com/konpyutaika/nifikop/pull/189) - **[Operator]** Fixed issue where nifikop's zookeeper init container would not tolerate multiple comma-delimited `host:port` pairs in the `NifiCluster.Spec.ZkAddress` configuration.

## v0.14.1

### Changed

- [PR #160](https://github.com/konpyutaika/nifikop/pull/160) - **[Documentation]** Upgrade documentation dependencies.

### Fixed Bugs

- [PR #174](https://github.com/konpyutaika/nifikop/pull/174) - **[Operator]** Fix K8S version getting.

## v0.14.0

### Added

- [PR #138](https://github.com/konpyutaika/nifikop/pull/138) - **[Operator/NifiCluster]** Add ability to configure the NiFi Load Balance port.
- [PR #144](https://github.com/konpyutaika/nifikop/pull/144) - **[Operator]** Add automatic detection of k8s prior 1.21.
- [PR #153](https://github.com/konpyutaika/nifikop/pull/153) - **[Helm Chart]** Added helm values to set common labels and annotations.

### Changed

- [PR #142](https://github.com/konpyutaika/nifikop/pull/142) - **[Operator]** Fixed issue where operator would modify `NifiCluster` and `NifiDataflow` status on every reconciliation loop unnecessarily.
- [PR #151](https://github.com/konpyutaika/nifikop/pull/151) - **[Operator]** Fixed an issue where the controller logging erroneously appeared to all come from the same controller.

### Fixed Bugs
- [PR #155](https://github.com/konpyutaika/nifikop/pull/155) - **[Operator]** Removed instances where reconcile requeue didn't honor the interval time

## v0.13.1

### Changed

- [PR #146](https://github.com/konpyutaika/nifikop/pull/146) - **[Operator/NifiCluster]** Move from volume prefix to pvc label selection for deletion

## v0.13.0

### Added

- [PR #89](https://github.com/konpyutaika/nifikop/pull/89) - **[Operator/NifiNodeGroupAutoscaler]** Add NifiNodeGroupAutoscaler to automatically horizontally scale a NifiCluster resource via the Kubernetes HorizontalPodAutoscaler.

## v0.12.0

### Added

- [PR #108](https://github.com/konpyutaika/nifikop/pull/108) - **[Operator/Logging]** Migrated from logr library to zap
- [PR #112](https://github.com/konpyutaika/nifikop/pull/112) - **[Documentation]** Add section to explain how upgrade from 0.7.6 to 0.8.0.
- [PR #114](https://github.com/konpyutaika/nifikop/pull/114) - **[Operator/NifiCluster]** Added ability to set the `PodSpec.HostAliases` to provide Pod-level override of hostname resolution when DNS and other options are not applicable.

### Changed

- [PR #136](https://github.com/konpyutaika/nifikop/pull/136) - **[Operator]** Update sync logic of dataflow to stop it fully.
- [PR #115](https://github.com/konpyutaika/nifikop/pull/115) - **[Operator]** Upgrade go version to 1.18.
- [PR #120](https://github.com/konpyutaika/nifikop/pull/120) - **[Operator]** Upgrade operator-sdk to v1.22.1.
- [PR #121](https://github.com/konpyutaika/nifikop/pull/121) - **[Operator]** Refactor much of the nifikop logging to include more context.
- [PR #122](https://github.com/konpyutaika/nifikop/pull/122) - **[Operator/NifiCluster]** Change name of PVCs that nifikop creates to include the name set via `NifiCluster.Spec.node_config_group.StorageConfigs.Name`
- [PR #123](https://github.com/konpyutaika/nifikop/pull/123) - **[Documentation]** Added nifi.sensitive.props.key to config samples


### Fixed Bugs

- [PR #135](https://github.com/konpyutaika/nifikop/pull/135) - **[Operator]** Update log generation to not reference nil variable
- [PR #106](https://github.com/konpyutaika/nifikop/pull/106) - **[Documentation]** Patch documentation version and mixed docs.
- [PR #110](https://github.com/konpyutaika/nifikop/pull/110) - **[Operator]** Handle case where `Certificate` is destroyed before `NifiUser` to avoid Nifi user controller getting stuck on deletion


## v0.11.0

### Added

- [PR #76](https://github.com/konpyutaika/nifikop/pull/76) - **[Operator/NiFiCluster]** Add ability to override default authorizers.xml template.
- [PR #95](https://github.com/konpyutaika/nifikop/pull/95) - **[Operator/NiFiParameterContext]** Allow the operator to take over existing parameter context.
- [PR #96](https://github.com/konpyutaika/nifikop/pull/96) - **[Operator/NifiCluster]** Add ability to specify pod priority class

### Changed

- [PR #75](https://github.com/konpyutaika/nifikop/pull/75) - **[Operator]** Update PodDisruptionBudget version to policy/v1 instead of policy/v1beta1.

### Deprecated

### Removed

- [PR #74](https://github.com/konpyutaika/nifikop/pull/74) - **[Operator]** Removed legacy orange CRDs.

### Fixed Bugs

- [PR #76](https://github.com/konpyutaika/nifikop/pull/88) - **[Operator/NiFiCluster]** Re-ordering config out of sync steps.
- [PR #93](https://github.com/konpyutaika/nifikop/pull/93) - **[Documentation]** Remove serviceAnnotations mentions and fix docs.
- [PR #101](https://github.com/konpyutaika/nifikop/pull/101) - **[Operator]** Handle finalizer removal case where `NifiCluster` is aggressively torn down and pods are no longer available to communicate with.

## v0.10.0

### Changed

- [PR #71](https://github.com/konpyutaika/nifikop/pull/71) - **[Operator]** Update cert-manager dep to v1.7.2 and all Certificate references to v1.
- [PR #29](https://github.com/konpyutaika/nifikop/pull/29) - **[Operator]** Update operator-sdk to v1.18.1.

## v0.9.0

### Added

- [PR #23](https://github.com/konpyutaika/nifikop/pull/23) - **[Operator/NiFiCluster]** Add ability to set services and pods labels
- [PR #21](https://github.com/konpyutaika/nifikop/pull/21) - **[Operator]** Propagate user provided issuerRef Group for custom CertManager Issuer.
- [PR #20](https://github.com/konpyutaika/nifikop/pull/20) - **[Operator]** Configurable log levels
- [PR #19](https://github.com/konpyutaika/nifikop/pull/19) - **[Helm chart]** Support --namespace helm arg
- [PR #18](https://github.com/konpyutaika/nifikop/pull/18) - **[Operator/NiFiCluster]** Support `topologySpreadConstraint`
- [PR #17](https://github.com/konpyutaika/nifikop/pull/17) - **[Operator/NiFiCluster]** Add ability to set max event driven thread count to NiFi Cluster.
- [PR #6](https://github.com/konpyutaika/nifikop/pull/6) - **[Operator/NiFiCluster]** Add ability to attach additional volumes & volumeMounts to NiFi container.

### Changed

- [PR #5](https://github.com/konpyutaika/nifikop/pull/5) - **[Documentation]** Change minikube by k3d.
- [PR #24](https://github.com/konpyutaika/nifikop/pull/24) - **[Operator/NiFiCluster]** Configurable node services and users template

### Deprecated

### Removed

### Fixed Bugs

## v0.8.0

## v0.7.6

### Added

- [PR #191](https://github.com/Orange-OpenSource/nifikop/pull/191) - **[Operator/NiFiDataflow]** Add event on registry client reference error.
- [PR #190](https://github.com/Orange-OpenSource/nifikop/pull/190) - **[Operator/NiFiDataflow]** New parameter: `flowPosition`.

### Changed

- [PR #188](https://github.com/Orange-OpenSource/nifikop/pull/188) - **[Operator/NiFiCluster]** Support all pod status as terminating if the pod phase is `failed`.

### Fixed Bugs

- [PR #167](https://github.com/Orange-OpenSource/nifikop/pull/167) - **[Operator/NiFiDataflow]** Fix nil pointer exception case whe sync Dataflow.
- [PR #189](https://github.com/Orange-OpenSource/nifikop/pull/189) - **[Operator/NiFiParameterContext]** Fix nil pointer exception case on empty description.
- [PR #193](https://github.com/Orange-OpenSource/nifikop/pull/193) - **[Documentation]** Fix some missinformation.
- [PR #196](https://github.com/Orange-OpenSource/nifikop/pull/196) - **[Operator/NiFiParameterContext]** Fix non-update of parameter context.
- [PR #197](https://github.com/Orange-OpenSource/nifikop/pull/197) - **[Operator/NiFiDataflow]** Keep Helm chart CRDs inline with baseline.
- [PR #198](https://github.com/Orange-OpenSource/nifikop/pull/198) - **[Documentation]** Fix versionned doc.

## v0.7.5

### Added

- [PR #162](https://github.com/Orange-OpenSource/nifikop/pull/162) - **[Operator/NiFiParameterContext]** Support declarative sensitive value out of secret.

### Fixed Bugs

- [PR #161](https://github.com/Orange-OpenSource/nifikop/pull/162) - **[Documentation]** NiFiCluster reference.
- [PR #161](https://github.com/Orange-OpenSource/nifikop/pull/162) - **[Operator/NiFiParameterContext]** Fix remove parameter and update set value to "no value set".

## v0.7.4

### Fixed Bugs

- [PR #160](https://github.com/Orange-OpenSource/nifikop/pull/160) - **[Dataflow]** Mandatory Position x and y.

## v0.7.3

### Fixed Bugs

- [PR #156](https://github.com/Orange-OpenSource/nifikop/pull/156) - **[Helm chart]** Operator metrics port configuration.
- [PR #157](https://github.com/Orange-OpenSource/nifikop/pull/157) - **[Operator/NiFiParameterContext]** Support optional parametere context and empty slice.

## v0.7.2

### Added

- [PR #152](https://github.com/Orange-OpenSource/nifikop/pull/152) - **[Operator]** Configurable requeue interval (#124)

### Fixed Bugs

- [PR #152](https://github.com/Orange-OpenSource/nifikop/pull/152) - **[Operator/NiFiParameterContext]** Fix is sync control in nil value case.

## v0.7.1

### Added

- [PR #144](https://github.com/Orange-OpenSource/nifikop/pull/144) - **[Operator/NiFiParameterContext]** Support empty string and no value set.

## v0.7.0

### Added

- [PR #132](https://github.com/Orange-OpenSource/nifikop/pull/132) - **[Operator]** Add the ability to manage dataflow lifecycle on non managed NiFi Cluster.
- [PR #132](https://github.com/Orange-OpenSource/nifikop/pull/132) - **[Operator]** Operator can interact with the NiFi cluster using basic authentication in addition to tls.

### Changed

- [PR #132](https://github.com/Orange-OpenSource/nifikop/pull/132) - **[Operator]** Enabling the ability to move a resource from one cluster to another by just changing the clusterReference.
- [PR #132](https://github.com/Orange-OpenSource/nifikop/pull/132) - **[Operator]** Improves the performances by reducing the amont of errors when interacting with then NiFi cluster API, checking cluster readiness before applying actions.
- [PR #132](https://github.com/Orange-OpenSource/nifikop/pull/132) - **[Operator/NiFiCluster]** Support `evicted` and `shutdown` pod status as terminating.

### Deprecated

### Removed

### Fixed Bugs

- [PR #132](https://github.com/Orange-OpenSource/nifikop/pull/132) - **[Operator/NiFiCluster]** Fix the downscale issue ([PR #131](https://github.com/Orange-OpenSource/nifikop/issues/131)) by removing references to configmap
- [PR #132](https://github.com/Orange-OpenSource/nifikop/pull/132) - **[Helm Chart]** Fix the RBAC definition for configmap and lease generated by operator-sdk with some mistakes.
- [PR #132](https://github.com/Orange-OpenSource/nifikop/pull/132) - **[Helm Chart]** Add corect CRDs in the chart helm.
- [PR #132](https://github.com/Orange-OpenSource/nifikop/pull/132) - **[Operator/NiFiUser]** Fix policy check conflict between user and group scope policy.

## v0.6.4

### Fixed Bugs

- [COMMIT #d98eb15fb3a74a1be17be5d456b02bd6a2d333cd](https://github.com/Orange-OpenSource/nifikop/tree/d98eb15fb3a74a1be17be5d456b02bd6a2d333cd) - **[Fix/NiFiCluster]** Fix external service port configuration being ignore [#133](https://github.com/Orange-OpenSource/nifikop/issues/133)
- [PR #134](https://github.com/Orange-OpenSource/nifikop/pull/134) - **[Operator/NifiCluster]** corrected typo in the nifi configmap for bootstrap-notification-service.
- [PR #119](https://github.com/Orange-OpenSource/nifikop/pull/119) - **[Helm Chart]** bring nificlusters crd in helm chart to spec with rest of repo.

## v0.6.3

### Added

- [PR #114](https://github.com/Orange-OpenSource/nifikop/pull/114) - **[Operator/NiFiCluster]** Additionals environment variables.

### Fixed Bugs

- [PR #113](https://github.com/Orange-OpenSource/nifikop/pull/113) - **[Operator/NiFiDataflow]** Simple work around to avoid null pointer dereferencing on nifi side.

## v0.6.2

### Fixed Bugs

- [PR #107](https://github.com/Orange-OpenSource/nifikop/pull/107) - **[Operator/NiFiCluster]** Correct the way to path PVCs.
- [PR #109](https://github.com/Orange-OpenSource/nifikop/pull/109) - **[Operator/NifiCluster]** Change namespace watch configuration to manage single namespace deletion.

## v0.6.1

### Added

- [PR #97](https://github.com/Orange-OpenSource/nifikop/pull/97) - **[Operator/NifiCluster]** Add ability to o define the maximum number of threads for timer driven processors available to the system.
- [PR #98](https://github.com/Orange-OpenSource/nifikop/pull/98) - **[Operator/NifiCluster]** Add empty_dir volume for `/tmp` dir.
- [PR #93](https://github.com/Orange-OpenSource/nifikop/pull/93) - **[Helm Chart]** Included securityContext and custom service account in helm chart for NiFiKop deployment.
- [PR #100](https://github.com/Orange-OpenSource/nifikop/pull/100) - **[Helm Chart]** Add nodeSelector, affinty and toleration in helm chart for NiFiKop deployment.

## v0.6.0

### Added

- [PR #86](https://github.com/Orange-OpenSource/nifikop/pull/86) - **[Operator/Debugging]** Add events and improve HTTP calls error message
- [PR #87](https://github.com/Orange-OpenSource/nifikop/pull/87) - **[Operator/Configuration]** Allow to override the `.properties` files using a config map and/or a secret.
- [PR #87](https://github.com/Orange-OpenSource/nifikop/pull/87) - **[Operator/Configuration]** Allow to replace the `logback.xml` and `bootstrap_notification_service.xml` files using a config map or a secret.
- [PR #88](https://github.com/Orange-OpenSource/nifikop/pull/88) - **[Operator/Monitoring]** By choosing `prometheus` as type for an internal service in a NiFiCluster resource, the operator automatically creates the associated `reporting task`.

### Changed

- [PR #85](https://github.com/Orange-OpenSource/nifikop/pull/85) - **[Operator/Dependencies]** Upgrade cert-manager & operator sdk dependencies
- [PR #87](https://github.com/Orange-OpenSource/nifikop/pull/87) - **[Operator/Configuration]** The node configuration files are no more stored in a configmap, but in a secret.

### Deprecated

- [PR #85](https://github.com/Orange-OpenSource/nifikop/pull/85) - **[Operator/Finalizers]** The finalizer name format suggested by [Kubernetes docs](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#finalizers) is <qualified-group>/<finalizer-name>, while the format previously documented by Operator SDK docs was <finalizer-name>.<qualified-group>. If your operator uses any finalizers with names matching the incorrect format, change them to match the official format. For example, finalizer.nifiusergroups.nifi.orange.com should be changed to nifiusergroups.nifi.orange.com/finalizer.

### Fixed Bugs

- [PR #87](https://github.com/Orange-OpenSource/nifikop/pull/87) - **[Operator/Configuration]** Since the v0.5.0 the operator doesn't catch certain resource changes, due to bad copy of resource, this issue avoid rolling upgrage trigger when some configuration changes.

## v0.5.3

### Fixed Bugs

- [PR #82](https://github.com/Orange-OpenSource/nifikop/pull/82) - **[Operator/NifiParameterContext]** Enable empty value
- [PR #83](https://github.com/Orange-OpenSource/nifikop/pull/83) - **[Operator/NiFiUser]** Rework the certificate secret creation, to prevent issues with JKS password creation.

## v0.5.2

### Fixed Bugs

- [PR #70](https://github.com/Orange-OpenSource/nifikop/pull/70) - **[Operator/NifiCluster]** Rework DNS names generation to fix non headless mode.

## v0.5.1

### Added

- [PR #61](https://github.com/Orange-OpenSource/nifikop/pull/61) - **[Operator/NifiCluster]** Replace hardcoded FSGroup (1000) with property (Implemented by @made-with-care in [PR #61](https://github.com/Orange-OpenSource/nifikop/pull/61))

### Fixed Bugs

- [PR #61](https://github.com/Orange-OpenSource/nifikop/pull/61) - **[Operator/NifiCluster]** Fix external service annotations (merge maps require a point for destination field)

## v0.5.0

### Added

- [PR #55](https://github.com/Orange-OpenSource/nifikop/pull/55) - **[Operator/NifiCluster]** Add the ability to define additional sidecar to each NiFi nodes.
- [PR #59](https://github.com/Orange-OpenSource/nifikop/pull/59) - **[Operator/NifiCluster]** Give the ability to define kubernetes services to expose the NiFi cluster.
- [PR #55](https://github.com/Orange-OpenSource/nifikop/pull/55) - **[Documentation]** Upgdrade to Docusaurus 2.0.0-alpha70 and enable versioned feature.

### Removed

- [PR #59](https://github.com/Orange-OpenSource/nifikop/pull/59) - **[Operator/NifiCluster]** No more default service to expose the NiFi cluster, you have to explicitly define ExternalService.

### Fixed Bugs

- [PR #62](https://github.com/Orange-OpenSource/nifikop/pull/62) - **[Operator/NifiCluster]** Fix DNS names for services in all-node mode.
- [PR #64](https://github.com/Orange-OpenSource/nifikop/pull/64) - **[Operator/NifiCluster]** Manage several NiFiClusters with the same managed user.

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

- [PR #25](https://github.com/Orange-OpenSource/nifikop/pull/25) - [Helm Chart] Add support for iterating over namespaces
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
- Implement configuration lifecycle management for : nifi.properties, zookeeper.properties, state-management.xml, login-identity-providers.xml, logback.xml, bootstrap.conf, bootstrap-notification-services.xml
- Initiate documentations
- Implementation basic makefile for some actions (debug, build, deploy, run, push, unit-test)
- Create helm chart for operator
- Add Documentation for internal deployment
- Add Gitlab CI Pipeline

### Changed

### Deprecated

### Removed

### Fixed Bugs
