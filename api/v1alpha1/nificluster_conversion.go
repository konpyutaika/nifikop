package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

// ConvertNifiClusterTo converts a v1alpha1 to v1 (Hub).
func (src *NifiCluster) ConvertTo(dst conversion.Hub) error {
	ncV1 := dst.(*v1.NifiCluster)

	if err := ConvertNifiClusterTo(src, ncV1); err != nil {
		return fmt.Errorf("unable to convert NifiCluster %s/%s to version: %v, err: %w", src.Namespace, src.Name, dst.GetObjectKind().GroupVersionKind().Version, err)
	}

	return nil
}

// ConvertFrom converts a v1 (Hub) to v1alpha1 (local).
func (dst *NifiCluster) ConvertFrom(src conversion.Hub) error { //nolint
	ncV1 := src.(*v1.NifiCluster)
	dst.ObjectMeta = ncV1.ObjectMeta
	if err := ConvertNifiClusterFrom(dst, ncV1); err != nil {
		return fmt.Errorf("unable to convert NiFiCluster %s/%s from version: %v, err: %w", dst.Namespace, dst.Name, src.GetObjectKind().GroupVersionKind().Version, err)
	}
	return nil
}

// ---- Convert TO ----

// ConvertNifiClusterTo use to convert v1alpha1.NifiCluster to v1.NifiCluster.
func ConvertNifiClusterTo(src *NifiCluster, dst *v1.NifiCluster) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertNifiClusterSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertNifiClusterStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertNifiClusterSpec(src *NifiClusterSpec, dst *v1.NifiCluster) error {
	if src == nil {
		return nil
	}

	dst.Spec.ClientType = v1.ClientConfigType(src.ClientType)
	dst.Spec.Type = v1.ClusterType(src.Type)
	dst.Spec.NodeURITemplate = src.NodeURITemplate
	dst.Spec.NifiURI = src.NifiURI
	dst.Spec.RootProcessGroupId = src.RootProcessGroupId
	dst.Spec.ProxyUrl = src.ProxyUrl
	dst.Spec.ZKAddress = src.ZKAddress
	dst.Spec.ZKPath = src.ZKPath
	dst.Spec.InitContainerImage = src.InitContainerImage
	dst.Spec.InitContainers = src.InitContainers
	dst.Spec.ClusterImage = src.ClusterImage
	dst.Spec.OneNifiNodePerNode = src.OneNifiNodePerNode
	dst.Spec.PropagateLabels = src.PropagateLabels
	if src.NodeUserIdentityTemplate != nil {
		dst.Spec.NodeUserIdentityTemplate = src.NodeUserIdentityTemplate
	}
	dst.Spec.SidecarConfigs = src.SidecarConfigs
	dst.Spec.TopologySpreadConstraints = src.TopologySpreadConstraints
	if src.NifiControllerTemplate != nil {
		dst.Spec.NifiControllerTemplate = src.NifiControllerTemplate
	}
	if src.ControllerUserIdentity != nil {
		dst.Spec.ControllerUserIdentity = src.ControllerUserIdentity
	}

	convertNifiClusterSecretRef(src.SecretRef, dst)
	convertNifiClusterPodPolicy(src.Pod, dst)
	convertNifiClusterServicePolicy(src.Service, dst)
	convertNifiClusterManagedAdminUsers(src.ManagedAdminUsers, dst)
	convertNifiClusterManagedReaderUsers(src.ManagedReaderUsers, dst)
	convertNifiClusterReadOnlyConfig(src.ReadOnlyConfig, dst)
	convertNifiClusterNodeConfigGroups(src.NodeConfigGroups, dst)
	convertNifiClusterNodes(src.Nodes, dst)
	convertNifiClusterDisruptionBudget(src.DisruptionBudget, dst)
	convertNifiClusterLdapConfiguration(src.LdapConfiguration, dst)
	convertNifiClusterTaskSpec(src.NifiClusterTaskSpec, dst)
	convertNifiClusterListenersConfig(src.ListenersConfig, dst)
	convertNifiClusterExternalServices(src.ExternalServices, dst)
	return nil
}

func convertNifiClusterSecretRef(src SecretReference, dst *v1.NifiCluster) {
	dst.Spec.SecretRef = getV1SecretRef(src)
}

func convertNifiClusterPodPolicy(src PodPolicy, dst *v1.NifiCluster) {
	dst.Spec.Pod = v1.PodPolicy{
		HostAliases:    src.HostAliases,
		Annotations:    src.Annotations,
		Labels:         src.Labels,
		ReadinessProbe: nil,
		LivenessProbe:  nil,
	}
}

func convertNifiClusterServicePolicy(src ServicePolicy, dst *v1.NifiCluster) {
	dst.Spec.Service = v1.ServicePolicy{
		HeadlessEnabled: src.HeadlessEnabled,
		ServiceTemplate: src.ServiceTemplate,
		Annotations:     src.Annotations,
		Labels:          src.Labels,
	}
}

func convertNifiClusterManagedAdminUsers(src []ManagedUser, dst *v1.NifiCluster) {
	dst.Spec.ManagedAdminUsers = []v1.ManagedUser{}
	for _, user := range src {
		dst.Spec.ManagedAdminUsers = append(dst.Spec.ManagedAdminUsers, v1.ManagedUser{
			Identity: user.Identity,
			Name:     user.Name,
		})
	}
}

func convertNifiClusterManagedReaderUsers(src []ManagedUser, dst *v1.NifiCluster) {
	dst.Spec.ManagedReaderUsers = []v1.ManagedUser{}
	for _, user := range src {
		dst.Spec.ManagedReaderUsers = append(dst.Spec.ManagedReaderUsers, v1.ManagedUser{
			Identity: user.Identity,
			Name:     user.Name,
		})
	}
}

func convertNifiClusterReadOnlyConfig(src ReadOnlyConfig, dst *v1.NifiCluster) {
	dst.Spec.ReadOnlyConfig = getV1ReadOnlyConfig(src)
}

func convertNifiProperties(src NifiProperties, dst *v1.ReadOnlyConfig) {
	dst.NifiProperties = v1.NifiProperties{
		OverrideConfigMap:    convertConfigMapReference(src.OverrideConfigMap),
		OverrideConfigs:      src.OverrideConfigs,
		OverrideSecretConfig: convertSecretConfigReference(src.OverrideSecretConfig),
		WebProxyHosts:        src.WebProxyHosts,
		NeedClientAuth:       src.NeedClientAuth,
		Authorizer:           src.Authorizer,
	}
}

func convertBootstrapProperties(src BootstrapProperties, dst *v1.ReadOnlyConfig) {
	dst.BootstrapProperties = v1.BootstrapProperties{
		NifiJvmMemory:        src.NifiJvmMemory,
		OverrideConfigMap:    convertConfigMapReference(src.OverrideConfigMap),
		OverrideConfigs:      src.OverrideConfigs,
		OverrideSecretConfig: convertSecretConfigReference(src.OverrideSecretConfig),
	}
}

func convertZookeeperProperties(src ZookeeperProperties, dst *v1.ReadOnlyConfig) {
	dst.ZookeeperProperties = v1.ZookeeperProperties{
		OverrideConfigMap:    convertConfigMapReference(src.OverrideConfigMap),
		OverrideConfigs:      src.OverrideConfigs,
		OverrideSecretConfig: convertSecretConfigReference(src.OverrideSecretConfig),
	}
}

func convertLogbackConfig(src LogbackConfig, dst *v1.ReadOnlyConfig) {
	dst.LogbackConfig = v1.LogbackConfig{
		ReplaceConfigMap:    convertConfigMapReference(src.ReplaceConfigMap),
		ReplaceSecretConfig: convertSecretConfigReference(src.ReplaceSecretConfig),
	}
}

func convertAuthorizerConfig(src AuthorizerConfig, dst *v1.ReadOnlyConfig) {
	dst.AuthorizerConfig = v1.AuthorizerConfig{
		ReplaceTemplateConfigMap:    convertConfigMapReference(src.ReplaceTemplateConfigMap),
		ReplaceTemplateSecretConfig: convertSecretConfigReference(src.ReplaceTemplateSecretConfig),
	}
}

func convertBootstrapNotificationServicesReplaceConfig(src BootstrapNotificationServicesConfig, dst *v1.ReadOnlyConfig) {
	dst.BootstrapNotificationServicesReplaceConfig = v1.BootstrapNotificationServicesConfig{
		ReplaceConfigMap:    convertConfigMapReference(src.ReplaceConfigMap),
		ReplaceSecretConfig: convertSecretConfigReference(src.ReplaceSecretConfig),
	}
}

func convertConfigMapReference(src *ConfigmapReference) *v1.ConfigmapReference {
	if src == nil {
		return nil
	}
	return &v1.ConfigmapReference{
		Name:      src.Name,
		Namespace: src.Namespace,
		Data:      src.Data,
	}
}

func convertSecretConfigReference(src *SecretConfigReference) *v1.SecretConfigReference {
	if src == nil {
		return nil
	}
	return &v1.SecretConfigReference{
		Name:      src.Name,
		Namespace: src.Namespace,
		Data:      src.Data,
	}
}

func convertNifiClusterNodeConfigGroups(src map[string]NodeConfig, dst *v1.NifiCluster) {
	dst.Spec.NodeConfigGroups = map[string]v1.NodeConfig{}
	for key, val := range src {
		dst.Spec.NodeConfigGroups[key] = convertNodeConfig(val)
	}
}

func convertNodeConfig(src NodeConfig) v1.NodeConfig {
	nConfig := v1.NodeConfig{
		ProvenanceStorage:  src.ProvenanceStorage,
		Image:              src.Image,
		ImagePullPolicy:    src.ImagePullPolicy,
		ServiceAccountName: src.ServiceAccountName,
		ImagePullSecrets:   src.ImagePullSecrets,
		NodeSelector:       src.NodeSelector,
		Tolerations:        src.Tolerations,
		HostAliases:        src.HostAliases,
		NifiContainerSpec:  src.NifiContainerSpec,
	}
	if src.RunAsUser != nil {
		nConfig.RunAsUser = src.RunAsUser
	}
	if src.FSGroup != nil {
		nConfig.FSGroup = src.FSGroup
	}
	if src.IsNode != nil {
		nConfig.IsNode = src.IsNode
	}
	if src.NodeAffinity != nil {
		nConfig.NodeAffinity = src.NodeAffinity
	}
	if src.ResourcesRequirements != nil {
		nConfig.ResourcesRequirements = src.ResourcesRequirements
	}
	if src.PriorityClassName != nil {
		nConfig.PriorityClassName = src.PriorityClassName
	}
	convertStorageConfigs(src.StorageConfigs, &nConfig)
	convertExternalVolumeConfigs(src.ExternalVolumeConfigs, &nConfig)
	nConfig.PodMetadata = convertMetadata(src.PodMetadata)

	return nConfig
}

func convertStorageConfigs(src []StorageConfig, dst *v1.NodeConfig) {
	dst.StorageConfigs = []v1.StorageConfig{}
	for _, srcConfig := range src {
		dstConfig := v1.StorageConfig{
			Name:          srcConfig.Name,
			MountPath:     srcConfig.MountPath,
			ReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			Metadata: v1.Metadata{
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
		}
		if srcConfig.PVCSpec != nil {
			dstConfig.PVCSpec = srcConfig.PVCSpec
		}
		dst.StorageConfigs = append(dst.StorageConfigs, dstConfig)
	}
}

func convertExternalVolumeConfigs(src []VolumeConfig, dst *v1.NodeConfig) {
	dst.ExternalVolumeConfigs = []v1.VolumeConfig{}
	for _, srcConfig := range src {
		dstConfig := v1.VolumeConfig{
			VolumeMount:  srcConfig.VolumeMount,
			VolumeSource: srcConfig.VolumeSource,
		}
		dst.ExternalVolumeConfigs = append(dst.ExternalVolumeConfigs, dstConfig)
	}
}

func convertMetadata(src Metadata) v1.Metadata {
	return v1.Metadata{
		Annotations: src.Annotations,
		Labels:      src.Labels,
	}
}

func convertNifiClusterNodes(src []Node, dst *v1.NifiCluster) {
	dst.Spec.Nodes = []v1.Node{}
	for _, srcNode := range src {
		dstNode := v1.Node{
			Id:              srcNode.Id,
			NodeConfigGroup: srcNode.NodeConfigGroup,
			Labels:          srcNode.Labels,
		}
		if srcNode.ReadOnlyConfig != nil {
			dstReadOnlyConfig := getV1ReadOnlyConfig(*srcNode.ReadOnlyConfig)
			dstNode.ReadOnlyConfig = &dstReadOnlyConfig
		}

		if srcNode.NodeConfig != nil {
			dstNodeConfig := convertNodeConfig(*srcNode.NodeConfig)
			dstNode.NodeConfig = &dstNodeConfig
		}
		dst.Spec.Nodes = append(dst.Spec.Nodes, dstNode)
	}
}

func convertNifiClusterDisruptionBudget(src DisruptionBudget, dst *v1.NifiCluster) {
	dst.Spec.DisruptionBudget = v1.DisruptionBudget{
		Create: src.Create,
		Budget: src.Budget,
	}
}

func convertNifiClusterLdapConfiguration(src LdapConfiguration, dst *v1.NifiCluster) {
	dst.Spec.LdapConfiguration = v1.LdapConfiguration{
		Enabled:      src.Enabled,
		Url:          src.Url,
		SearchBase:   src.SearchBase,
		SearchFilter: src.SearchFilter,
	}
}

func convertNifiClusterTaskSpec(src NifiClusterTaskSpec, dst *v1.NifiCluster) {
	dst.Spec.NifiClusterTaskSpec = v1.NifiClusterTaskSpec{
		RetryDurationMinutes: src.RetryDurationMinutes,
	}
}

func convertNifiClusterListenersConfig(src *ListenersConfig, dst *v1.NifiCluster) {
	if src == nil {
		return
	}
	dst.Spec.ListenersConfig = &v1.ListenersConfig{
		ClusterDomain:     src.ClusterDomain,
		UseExternalDNS:    src.UseExternalDNS,
		InternalListeners: convertInternalListeners(src.InternalListeners),
	}
	convertSSLSecrets(src.SSLSecrets, dst.Spec.ListenersConfig)
}

func convertInternalListeners(src []InternalListenerConfig) []v1.InternalListenerConfig {
	var dstInternalListenerConfig []v1.InternalListenerConfig
	for _, srcInternalListenerConfig := range src {
		dstInternalListenerConfig = append(dstInternalListenerConfig, v1.InternalListenerConfig{
			Type:          srcInternalListenerConfig.Type,
			Name:          srcInternalListenerConfig.Name,
			ContainerPort: srcInternalListenerConfig.ContainerPort,
			// default to TCP when converting from v1alpha1 to v1
			Protocol: corev1.ProtocolTCP,
		})
	}

	return dstInternalListenerConfig
}

func convertSSLSecrets(src *SSLSecrets, dst *v1.ListenersConfig) {
	if src == nil {
		dst.SSLSecrets = nil
		return
	}
	dst.SSLSecrets = &v1.SSLSecrets{
		TLSSecretName: src.TLSSecretName,
		Create:        src.Create,
		ClusterScoped: src.ClusterScoped,
		PKIBackend:    convertPKIBackend(src.PKIBackend),
	}

	if src.IssuerRef != nil {
		dst.SSLSecrets.IssuerRef = src.IssuerRef
	}
}

func convertPKIBackend(src PKIBackend) v1.PKIBackend {
	return v1.PKIBackend(src)
}

func convertNifiClusterExternalServices(src []ExternalServiceConfig, dst *v1.NifiCluster) {
	dst.Spec.ExternalServices = []v1.ExternalServiceConfig{}
	for _, srcExternalServiceConfig := range src {
		dst.Spec.ExternalServices = append(dst.Spec.ExternalServices, v1.ExternalServiceConfig{
			Name:     srcExternalServiceConfig.Name,
			Metadata: convertMetadata(srcExternalServiceConfig.Metadata),
			Spec:     convertExternalServiceSpec(srcExternalServiceConfig.Spec),
		})
	}
}

func convertExternalServiceSpec(src ExternalServiceSpec) v1.ExternalServiceSpec {
	return v1.ExternalServiceSpec{
		PortConfigs:              convertPortConfigs(src.PortConfigs),
		ClusterIP:                src.ClusterIP,
		Type:                     src.Type,
		ExternalIPs:              src.ExternalIPs,
		LoadBalancerIP:           src.LoadBalancerIP,
		LoadBalancerSourceRanges: src.LoadBalancerSourceRanges,
		ExternalName:             src.ExternalName,
	}
}

func convertPortConfigs(src []PortConfig) []v1.PortConfig {
	var dstPortConfigs []v1.PortConfig
	for _, srcPortConfig := range src {
		dstPortConfigs = append(dstPortConfigs, v1.PortConfig{
			Port:                 srcPortConfig.Port,
			InternalListenerName: srcPortConfig.InternalListenerName,
			Protocol:             corev1.ProtocolTCP,
		})
	}

	return dstPortConfigs
}

func convertNifiClusterStatus(src *NifiClusterStatus, dst *v1.NifiCluster) error {
	if src == nil {
		return nil
	}

	dst.Status.RootProcessGroupId = src.RootProcessGroupId
	dst.Status.State = v1.ClusterState(src.State)
	convertNifiClusterNodesState(src.NodesState, dst)
	convertNifiClusterRollingUpgrade(src.RollingUpgrade, dst)
	convertNifiClusterPrometheusReportingTask(src.PrometheusReportingTask, dst)

	return nil
}

func convertNifiClusterNodesState(src map[string]NodeState, dst *v1.NifiCluster) {
	dst.Status.NodesState = map[string]v1.NodeState{}
	for srcNodeId, srcNodeState := range src {
		dstNodeState := v1.NodeState{
			GracefulActionState: convertGracefulActionState(srcNodeState.GracefulActionState),
			ConfigurationState:  v1.ConfigurationState(srcNodeState.ConfigurationState),
			InitClusterNode:     v1.InitClusterNode(srcNodeState.InitClusterNode),
			PodIsReady:          srcNodeState.PodIsReady,
			LastUpdatedTime:     srcNodeState.LastUpdatedTime,
		}

		if srcNodeState.CreationTime != nil {
			srcCreationTime := *srcNodeState.CreationTime
			dstNodeState.CreationTime = &srcCreationTime
		}

		dst.Status.NodesState[srcNodeId] = dstNodeState
	}
}

func convertGracefulActionState(src GracefulActionState) v1.GracefulActionState {
	return v1.GracefulActionState{
		ErrorMessage: src.ErrorMessage,
		ActionStep:   v1.ActionStep(src.ActionStep),
		TaskStarted:  src.TaskStarted,
		State:        v1.State(src.State),
	}
}

func convertNifiClusterRollingUpgrade(src RollingUpgradeStatus, dst *v1.NifiCluster) {
	dst.Status.RollingUpgrade = v1.RollingUpgradeStatus{
		LastSuccess: src.LastSuccess,
		ErrorCount:  src.ErrorCount,
	}
}

func convertNifiClusterPrometheusReportingTask(src PrometheusReportingTaskStatus, dst *v1.NifiCluster) {
	dst.Status.PrometheusReportingTask = v1.PrometheusReportingTaskStatus{
		Id:      src.Id,
		Version: src.Version,
	}
}

// ---- Convert FROM ----

// ConvertFrom use to convert v1alpha1.NifiCluster from v1.NifiCluster.
func ConvertNifiClusterFrom(dst *NifiCluster, src *v1.NifiCluster) error {
	// Copying ObjectMeta as a whole
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	if err := convertNifiClusterFromSpec(&src.Spec, dst); err != nil {
		return err
	}

	// Convert status
	if err := convertNifiClusterFromStatus(&src.Status, dst); err != nil {
		return err
	}
	return nil
}

// Convert the top level structs.
func convertNifiClusterFromSpec(src *v1.NifiClusterSpec, dst *NifiCluster) error {
	if src == nil {
		return nil
	}

	dst.Spec.ClientType = ClientConfigType(src.ClientType)
	dst.Spec.Type = ClusterType(src.Type)
	dst.Spec.NodeURITemplate = src.NodeURITemplate
	dst.Spec.NifiURI = src.NifiURI
	dst.Spec.RootProcessGroupId = src.RootProcessGroupId
	dst.Spec.ProxyUrl = src.ProxyUrl
	dst.Spec.ZKAddress = src.ZKAddress
	dst.Spec.ZKPath = src.ZKPath
	dst.Spec.InitContainerImage = src.InitContainerImage
	dst.Spec.InitContainers = src.InitContainers
	dst.Spec.ClusterImage = src.ClusterImage
	dst.Spec.OneNifiNodePerNode = src.OneNifiNodePerNode
	dst.Spec.PropagateLabels = src.PropagateLabels
	if src.NodeUserIdentityTemplate != nil {
		dst.Spec.NodeUserIdentityTemplate = src.NodeUserIdentityTemplate
	}
	dst.Spec.SidecarConfigs = src.SidecarConfigs
	dst.Spec.TopologySpreadConstraints = src.TopologySpreadConstraints
	if src.NifiControllerTemplate != nil {
		dst.Spec.NifiControllerTemplate = src.NifiControllerTemplate
	}
	if src.ControllerUserIdentity != nil {
		dst.Spec.ControllerUserIdentity = src.ControllerUserIdentity
	}

	convertNifiClusterFromSecretRef(src.SecretRef, dst)
	convertNifiClusterFromPodPolicy(src.Pod, dst)
	convertNifiClusterFromServicePolicy(src.Service, dst)
	convertNifiClusterFromManagedAdminUsers(src.ManagedAdminUsers, dst)
	convertNifiClusterFromManagedReaderUsers(src.ManagedReaderUsers, dst)
	convertNifiClusterFromReadOnlyConfig(src.ReadOnlyConfig, dst)
	convertNifiClusterFromNodeConfigGroups(src.NodeConfigGroups, dst)
	convertNifiClusterFromNodes(src.Nodes, dst)
	convertNifiClusterFromDisruptionBudget(src.DisruptionBudget, dst)
	convertNifiClusterFromLdapConfiguration(src.LdapConfiguration, dst)
	convertFromNifiClusterTaskSpec(src.NifiClusterTaskSpec, dst)
	convertNifiClusterFromListenersConfig(src.ListenersConfig, dst)
	convertNifiClusterFromExternalServices(src.ExternalServices, dst)
	return nil
}

func convertNifiClusterFromSecretRef(src v1.SecretReference, dst *NifiCluster) {
	dst.Spec.SecretRef = getSecretRef(src)
}

func convertNifiClusterFromPodPolicy(src v1.PodPolicy, dst *NifiCluster) {
	dst.Spec.Pod = PodPolicy{
		HostAliases: src.HostAliases,
		Annotations: src.Annotations,
		Labels:      src.Labels,
	}
}

func convertNifiClusterFromServicePolicy(src v1.ServicePolicy, dst *NifiCluster) {
	dst.Spec.Service = ServicePolicy{
		HeadlessEnabled: src.HeadlessEnabled,
		ServiceTemplate: src.ServiceTemplate,
		Annotations:     src.Annotations,
		Labels:          src.Labels,
	}
}

func convertNifiClusterFromManagedAdminUsers(src []v1.ManagedUser, dst *NifiCluster) {
	dst.Spec.ManagedAdminUsers = []ManagedUser{}
	for _, user := range src {
		dst.Spec.ManagedAdminUsers = append(dst.Spec.ManagedAdminUsers, ManagedUser{
			Identity: user.Identity,
			Name:     user.Name,
		})
	}
}

func convertNifiClusterFromManagedReaderUsers(src []v1.ManagedUser, dst *NifiCluster) {
	dst.Spec.ManagedReaderUsers = []ManagedUser{}
	for _, user := range src {
		dst.Spec.ManagedReaderUsers = append(dst.Spec.ManagedReaderUsers, ManagedUser{
			Identity: user.Identity,
			Name:     user.Name,
		})
	}
}

func convertNifiClusterFromReadOnlyConfig(src v1.ReadOnlyConfig, dst *NifiCluster) {
	dst.Spec.ReadOnlyConfig = getReadOnlyConfig(src)
}

func convertFromNifiProperties(src v1.NifiProperties, dst *ReadOnlyConfig) {
	dst.NifiProperties = NifiProperties{
		OverrideConfigMap:    convertFromConfigMapReference(src.OverrideConfigMap),
		OverrideConfigs:      src.OverrideConfigs,
		OverrideSecretConfig: convertFromSecretConfigReference(src.OverrideSecretConfig),
		WebProxyHosts:        src.WebProxyHosts,
		NeedClientAuth:       src.NeedClientAuth,
		Authorizer:           src.Authorizer,
	}
}

func convertFromBootstrapProperties(src v1.BootstrapProperties, dst *ReadOnlyConfig) {
	dst.BootstrapProperties = BootstrapProperties{
		NifiJvmMemory:        src.NifiJvmMemory,
		OverrideConfigMap:    convertFromConfigMapReference(src.OverrideConfigMap),
		OverrideConfigs:      src.OverrideConfigs,
		OverrideSecretConfig: convertFromSecretConfigReference(src.OverrideSecretConfig),
	}
}

func convertFromZookeeperProperties(src v1.ZookeeperProperties, dst *ReadOnlyConfig) {
	dst.ZookeeperProperties = ZookeeperProperties{
		OverrideConfigMap:    convertFromConfigMapReference(src.OverrideConfigMap),
		OverrideConfigs:      src.OverrideConfigs,
		OverrideSecretConfig: convertFromSecretConfigReference(src.OverrideSecretConfig),
	}
}

func convertFromLogbackConfig(src v1.LogbackConfig, dst *ReadOnlyConfig) {
	dst.LogbackConfig = LogbackConfig{
		ReplaceConfigMap:    convertFromConfigMapReference(src.ReplaceConfigMap),
		ReplaceSecretConfig: convertFromSecretConfigReference(src.ReplaceSecretConfig),
	}
}

func convertFromAuthorizerConfig(src v1.AuthorizerConfig, dst *ReadOnlyConfig) {
	dst.AuthorizerConfig = AuthorizerConfig{
		ReplaceTemplateConfigMap:    convertFromConfigMapReference(src.ReplaceTemplateConfigMap),
		ReplaceTemplateSecretConfig: convertFromSecretConfigReference(src.ReplaceTemplateSecretConfig),
	}
}

func convertFromBootstrapNotificationServicesReplaceConfig(src v1.BootstrapNotificationServicesConfig, dst *ReadOnlyConfig) {
	dst.BootstrapNotificationServicesReplaceConfig = BootstrapNotificationServicesConfig{
		ReplaceConfigMap:    convertFromConfigMapReference(src.ReplaceConfigMap),
		ReplaceSecretConfig: convertFromSecretConfigReference(src.ReplaceSecretConfig),
	}
}

func convertFromConfigMapReference(src *v1.ConfigmapReference) *ConfigmapReference {
	if src == nil {
		return nil
	}
	return &ConfigmapReference{
		Name:      src.Name,
		Namespace: src.Namespace,
		Data:      src.Data,
	}
}

func convertFromSecretConfigReference(src *v1.SecretConfigReference) *SecretConfigReference {
	if src == nil {
		return nil
	}
	return &SecretConfigReference{
		Name:      src.Name,
		Namespace: src.Namespace,
		Data:      src.Data,
	}
}

func convertNifiClusterFromNodeConfigGroups(src map[string]v1.NodeConfig, dst *NifiCluster) {
	dst.Spec.NodeConfigGroups = map[string]NodeConfig{}
	for key, val := range src {
		dst.Spec.NodeConfigGroups[key] = convertFromNodeConfig(val)
	}
}

func convertFromNodeConfig(src v1.NodeConfig) NodeConfig {
	nConfig := NodeConfig{
		ProvenanceStorage:  src.ProvenanceStorage,
		Image:              src.Image,
		ImagePullPolicy:    src.ImagePullPolicy,
		ServiceAccountName: src.ServiceAccountName,
		ImagePullSecrets:   src.ImagePullSecrets,
		NodeSelector:       src.NodeSelector,
		Tolerations:        src.Tolerations,
		HostAliases:        src.HostAliases,
		NifiContainerSpec:  src.NifiContainerSpec,
	}
	if src.RunAsUser != nil {
		nConfig.RunAsUser = src.RunAsUser
	}
	if src.FSGroup != nil {
		nConfig.FSGroup = src.FSGroup
	}
	if src.IsNode != nil {
		nConfig.IsNode = src.IsNode
	}
	if src.NodeAffinity != nil {
		nConfig.NodeAffinity = src.NodeAffinity
	}
	if src.ResourcesRequirements != nil {
		nConfig.ResourcesRequirements = src.ResourcesRequirements
	}
	if src.PriorityClassName != nil {
		nConfig.PriorityClassName = src.PriorityClassName
	}
	convertFromStorageConfigs(src.StorageConfigs, &nConfig)
	convertFromExternalVolumeConfigs(src.ExternalVolumeConfigs, &nConfig)
	nConfig.PodMetadata = convertFromMetadata(src.PodMetadata)

	return nConfig
}

func convertFromStorageConfigs(src []v1.StorageConfig, dst *NodeConfig) {
	dst.StorageConfigs = []StorageConfig{}
	for _, srcConfig := range src {
		dstConfig := StorageConfig{
			Name:      srcConfig.Name,
			MountPath: srcConfig.MountPath,
		}
		if srcConfig.PVCSpec != nil {
			dstConfig.PVCSpec = srcConfig.PVCSpec
		}
		dst.StorageConfigs = append(dst.StorageConfigs, dstConfig)
	}
}

func convertFromExternalVolumeConfigs(src []v1.VolumeConfig, dst *NodeConfig) {
	dst.ExternalVolumeConfigs = []VolumeConfig{}
	for _, srcConfig := range src {
		dstConfig := VolumeConfig{
			VolumeMount:  srcConfig.VolumeMount,
			VolumeSource: srcConfig.VolumeSource,
		}
		dst.ExternalVolumeConfigs = append(dst.ExternalVolumeConfigs, dstConfig)
	}
}

func convertFromMetadata(src v1.Metadata) Metadata {
	return Metadata{
		Annotations: src.Annotations,
		Labels:      src.Labels,
	}
}

func convertNifiClusterFromNodes(src []v1.Node, dst *NifiCluster) {
	dst.Spec.Nodes = []Node{}
	for _, srcNode := range src {
		dstNode := Node{
			Id:              srcNode.Id,
			NodeConfigGroup: srcNode.NodeConfigGroup,
			Labels:          srcNode.Labels,
		}
		if srcNode.ReadOnlyConfig != nil {
			dstReadOnlyConfig := getReadOnlyConfig(*srcNode.ReadOnlyConfig)
			dstNode.ReadOnlyConfig = &dstReadOnlyConfig
		}

		if srcNode.NodeConfig != nil {
			dstNodeConfig := convertFromNodeConfig(*srcNode.NodeConfig)
			dstNode.NodeConfig = &dstNodeConfig
		}
		dst.Spec.Nodes = append(dst.Spec.Nodes, dstNode)
	}
}

func convertNifiClusterFromDisruptionBudget(src v1.DisruptionBudget, dst *NifiCluster) {
	dst.Spec.DisruptionBudget = DisruptionBudget{
		Create: src.Create,
		Budget: src.Budget,
	}
}

func convertNifiClusterFromLdapConfiguration(src v1.LdapConfiguration, dst *NifiCluster) {
	dst.Spec.LdapConfiguration = LdapConfiguration{
		Enabled:      src.Enabled,
		Url:          src.Url,
		SearchBase:   src.SearchBase,
		SearchFilter: src.SearchFilter,
	}
}

func convertFromNifiClusterTaskSpec(src v1.NifiClusterTaskSpec, dst *NifiCluster) {
	dst.Spec.NifiClusterTaskSpec = NifiClusterTaskSpec{
		RetryDurationMinutes: src.RetryDurationMinutes,
	}
}

func convertNifiClusterFromListenersConfig(src *v1.ListenersConfig, dst *NifiCluster) {
	if src == nil {
		return
	}
	dst.Spec.ListenersConfig = &ListenersConfig{
		ClusterDomain:     src.ClusterDomain,
		UseExternalDNS:    src.UseExternalDNS,
		InternalListeners: convertFromInternalListeners(src.InternalListeners),
	}
	convertFromSSLSecrets(src.SSLSecrets, dst.Spec.ListenersConfig)
}

func convertFromInternalListeners(src []v1.InternalListenerConfig) []InternalListenerConfig {
	var dstInternalListenerConfig []InternalListenerConfig
	for _, srcInternalListenerConfig := range src {
		dstInternalListenerConfig = append(dstInternalListenerConfig, InternalListenerConfig{
			Type:          srcInternalListenerConfig.Type,
			Name:          srcInternalListenerConfig.Name,
			ContainerPort: srcInternalListenerConfig.ContainerPort,
		})
	}

	return dstInternalListenerConfig
}

func convertFromSSLSecrets(src *v1.SSLSecrets, dst *ListenersConfig) {
	if src == nil {
		dst.SSLSecrets = nil
		return
	}
	dst.SSLSecrets = &SSLSecrets{
		TLSSecretName: src.TLSSecretName,
		Create:        src.Create,
		ClusterScoped: src.ClusterScoped,
		PKIBackend:    convertFromPKIBackend(src.PKIBackend),
	}

	if src.IssuerRef != nil {
		dst.SSLSecrets.IssuerRef = src.IssuerRef
	}
}

func convertFromPKIBackend(src v1.PKIBackend) PKIBackend {
	return PKIBackend(src)
}

func convertNifiClusterFromExternalServices(src []v1.ExternalServiceConfig, dst *NifiCluster) {
	dst.Spec.ExternalServices = []ExternalServiceConfig{}
	for _, srcExternalServiceConfig := range src {
		dst.Spec.ExternalServices = append(dst.Spec.ExternalServices, ExternalServiceConfig{
			Name:     srcExternalServiceConfig.Name,
			Metadata: convertFromMetadata(srcExternalServiceConfig.Metadata),
			Spec:     convertFromExternalServiceSpec(srcExternalServiceConfig.Spec),
		})
	}
}

func convertFromExternalServiceSpec(src v1.ExternalServiceSpec) ExternalServiceSpec {
	return ExternalServiceSpec{
		PortConfigs:              convertFromPortConfigs(src.PortConfigs),
		ClusterIP:                src.ClusterIP,
		Type:                     src.Type,
		ExternalIPs:              src.ExternalIPs,
		LoadBalancerIP:           src.LoadBalancerIP,
		LoadBalancerSourceRanges: src.LoadBalancerSourceRanges,
		ExternalName:             src.ExternalName,
	}
}

func convertFromPortConfigs(src []v1.PortConfig) []PortConfig {
	var dstPortConfigs []PortConfig
	for _, srcPortConfig := range src {
		dstPortConfigs = append(dstPortConfigs, PortConfig{
			Port:                 srcPortConfig.Port,
			InternalListenerName: srcPortConfig.InternalListenerName,
		})
	}

	return dstPortConfigs
}

func convertNifiClusterFromStatus(src *v1.NifiClusterStatus, dst *NifiCluster) error {
	if src == nil {
		return nil
	}

	dst.Status.RootProcessGroupId = src.RootProcessGroupId
	dst.Status.State = ClusterState(src.State)
	convertNifiClusterFromNodesState(src.NodesState, dst)
	convertNifiClusterFromRollingUpgrade(src.RollingUpgrade, dst)
	convertNifiClusterFromPrometheusReportingTask(src.PrometheusReportingTask, dst)

	return nil
}

func convertNifiClusterFromNodesState(src map[string]v1.NodeState, dst *NifiCluster) {
	dst.Status.NodesState = map[string]NodeState{}
	for srcNodeId, srcNodeState := range src {
		dstNodeState := NodeState{
			GracefulActionState: convertFromGracefulActionState(srcNodeState.GracefulActionState),
			ConfigurationState:  ConfigurationState(srcNodeState.ConfigurationState),
			InitClusterNode:     InitClusterNode(srcNodeState.InitClusterNode),
			PodIsReady:          srcNodeState.PodIsReady,
			LastUpdatedTime:     srcNodeState.LastUpdatedTime,
		}

		if srcNodeState.CreationTime != nil {
			srcCreationTime := *srcNodeState.CreationTime
			dstNodeState.CreationTime = &srcCreationTime
		}

		dst.Status.NodesState[srcNodeId] = dstNodeState
	}
}

func convertFromGracefulActionState(src v1.GracefulActionState) GracefulActionState {
	return GracefulActionState{
		ErrorMessage: src.ErrorMessage,
		ActionStep:   ActionStep(src.ActionStep),
		TaskStarted:  src.TaskStarted,
		State:        State(src.State),
	}
}

func convertNifiClusterFromRollingUpgrade(src v1.RollingUpgradeStatus, dst *NifiCluster) {
	dst.Status.RollingUpgrade = RollingUpgradeStatus{
		LastSuccess: src.LastSuccess,
		ErrorCount:  src.ErrorCount,
	}
}

func convertNifiClusterFromPrometheusReportingTask(src v1.PrometheusReportingTaskStatus, dst *NifiCluster) {
	dst.Status.PrometheusReportingTask = PrometheusReportingTaskStatus{
		Id:      src.Id,
		Version: src.Version,
	}
}
