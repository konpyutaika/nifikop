package v1alpha1

import (
	v1 "github.com/konpyutaika/nifikop/api/v1"
)

// SecretRef.
func getSecretRef(src v1.SecretReference) SecretReference {
	return SecretReference{
		Name:      src.Name,
		Namespace: src.Namespace,
	}
}

func getV1SecretRef(src SecretReference) v1.SecretReference {
	return v1.SecretReference{
		Name:      src.Name,
		Namespace: src.Namespace,
	}
}

// ReadOnlyConfig.
func getReadOnlyConfig(src v1.ReadOnlyConfig) ReadOnlyConfig {
	dstReadOnlyConfig := ReadOnlyConfig{
		AdditionalSharedEnvs: src.AdditionalSharedEnvs,
	}
	if src.MaximumTimerDrivenThreadCount != nil {
		dstReadOnlyConfig.MaximumTimerDrivenThreadCount = src.MaximumTimerDrivenThreadCount
	}

	if src.MaximumEventDrivenThreadCount != nil {
		dstReadOnlyConfig.MaximumEventDrivenThreadCount = src.MaximumEventDrivenThreadCount
	}

	convertFromNifiProperties(src.NifiProperties, &dstReadOnlyConfig)
	convertFromZookeeperProperties(src.ZookeeperProperties, &dstReadOnlyConfig)
	convertFromBootstrapProperties(src.BootstrapProperties, &dstReadOnlyConfig)
	convertFromLogbackConfig(src.LogbackConfig, &dstReadOnlyConfig)
	convertFromAuthorizerConfig(src.AuthorizerConfig, &dstReadOnlyConfig)
	convertFromBootstrapNotificationServicesReplaceConfig(src.BootstrapNotificationServicesReplaceConfig, &dstReadOnlyConfig)

	return dstReadOnlyConfig
}

func getV1ReadOnlyConfig(src ReadOnlyConfig) v1.ReadOnlyConfig {
	dstReadOnlyConfig := v1.ReadOnlyConfig{
		AdditionalSharedEnvs: src.AdditionalSharedEnvs,
	}
	if src.MaximumTimerDrivenThreadCount != nil {
		dstReadOnlyConfig.MaximumTimerDrivenThreadCount = src.MaximumTimerDrivenThreadCount
	}

	if src.MaximumEventDrivenThreadCount != nil {
		dstReadOnlyConfig.MaximumEventDrivenThreadCount = src.MaximumEventDrivenThreadCount
	}

	convertNifiProperties(src.NifiProperties, &dstReadOnlyConfig)
	convertZookeeperProperties(src.ZookeeperProperties, &dstReadOnlyConfig)
	convertBootstrapProperties(src.BootstrapProperties, &dstReadOnlyConfig)
	convertLogbackConfig(src.LogbackConfig, &dstReadOnlyConfig)
	convertAuthorizerConfig(src.AuthorizerConfig, &dstReadOnlyConfig)
	convertBootstrapNotificationServicesReplaceConfig(src.BootstrapNotificationServicesReplaceConfig, &dstReadOnlyConfig)

	return dstReadOnlyConfig
}

// ClusterRef.
func getClusterReference(src v1.ClusterReference) ClusterReference {
	return ClusterReference{
		Name:      src.Name,
		Namespace: src.Namespace,
	}
}

func getV1ClusterReference(src ClusterReference) v1.ClusterReference {
	return v1.ClusterReference{
		Name:      src.Name,
		Namespace: src.Namespace,
	}
}

// ParameterContextRef.
func getV1ParameterContextRef(src ParameterContextReference) v1.ParameterContextReference {
	return v1.ParameterContextReference{
		Name:      src.Name,
		Namespace: src.Namespace,
	}
}

func getParameterContextRef(src v1.ParameterContextReference) ParameterContextReference {
	return ParameterContextReference{
		Name:      src.Name,
		Namespace: src.Namespace,
	}
}

// AccessPolicy.
func getV1AccessPolicy(src AccessPolicy) v1.AccessPolicy {
	return v1.AccessPolicy{
		Type:          v1.AccessPolicyType(src.Type),
		Action:        v1.AccessPolicyAction(src.Action),
		Resource:      v1.AccessPolicyResource(src.Resource),
		ComponentType: src.ComponentType,
		ComponentId:   src.ComponentId,
	}
}

func getAccessPolicy(src v1.AccessPolicy) AccessPolicy {
	return AccessPolicy{
		Type:          AccessPolicyType(src.Type),
		Action:        AccessPolicyAction(src.Action),
		Resource:      AccessPolicyResource(src.Resource),
		ComponentType: src.ComponentType,
		ComponentId:   src.ComponentId,
	}
}
