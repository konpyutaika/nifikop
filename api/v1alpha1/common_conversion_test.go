package v1alpha1

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

func TestGetSecretRef(t *testing.T) {
	sec := v1.SecretReference{
		Name:      "foo",
		Namespace: "namespace",
	}
	alphaSecRef := getSecretRef(sec)

	if sec.Name != alphaSecRef.Name || sec.Namespace != alphaSecRef.Namespace {
		t.Error("Secret refs are not equivalent")
	}
}

func TestGetV1SecretRef(t *testing.T) {
	sec := SecretReference{
		Name:      "foo",
		Namespace: "namespace",
	}
	alphaSecRef := getV1SecretRef(sec)

	if sec.Name != alphaSecRef.Name || sec.Namespace != alphaSecRef.Namespace {
		t.Error("Secret refs are not equivalent")
	}
}

func TestGetReadOnlyConfig(t *testing.T) {
	var num int32 = 8
	var configMapName string = "configMapRef"
	var secretName string = "secretRef"
	var namespace string = "namespace"
	roc := v1.ReadOnlyConfig{
		MaximumTimerDrivenThreadCount: &num,
		MaximumEventDrivenThreadCount: &num,
		AdditionalSharedEnvs: []corev1.EnvVar{
			{
				Name:  "ENV_VAR",
				Value: "foo",
			},
		},
		NifiProperties: v1.NifiProperties{
			OverrideConfigMap: &v1.ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			OverrideConfigs: "nifi.prop.foo=value",
			OverrideSecretConfig: &v1.SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
			WebProxyHosts: []string{
				"foo.host",
			},
			NeedClientAuth: true,
			Authorizer:     "blah-authorizer",
		},
		ZookeeperProperties: v1.ZookeeperProperties{
			OverrideConfigMap: &v1.ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			OverrideConfigs: "nifi.prop.foo=value",
			OverrideSecretConfig: &v1.SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
		},
		BootstrapProperties: v1.BootstrapProperties{
			OverrideConfigMap: &v1.ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			OverrideConfigs: "nifi.prop.foo=value",
			OverrideSecretConfig: &v1.SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
			NifiJvmMemory: "jvmMem",
		},
		LogbackConfig: v1.LogbackConfig{
			ReplaceConfigMap: &v1.ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			ReplaceSecretConfig: &v1.SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
		},
		AuthorizerConfig: v1.AuthorizerConfig{
			ReplaceTemplateConfigMap: &v1.ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			ReplaceTemplateSecretConfig: &v1.SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
		},
		BootstrapNotificationServicesReplaceConfig: v1.BootstrapNotificationServicesConfig{
			ReplaceConfigMap: &v1.ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			ReplaceSecretConfig: &v1.SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
		},
	}

	alphaRoc := getReadOnlyConfig(roc)
	assertReadOnlyConfigsEqual(roc, alphaRoc, t)
}

func TestGetV1ReadOnlyConfig(t *testing.T) {
	roc := createReadOnlyConfig()

	v1Roc := getV1ReadOnlyConfig(roc)
	assertReadOnlyConfigsEqual(v1Roc, roc, t)
}

func createReadOnlyConfigPtr() *ReadOnlyConfig {
	roc := createReadOnlyConfig()
	return &roc
}

func createReadOnlyConfig() ReadOnlyConfig {
	var num int32 = 8
	var configMapName string = "configMapRef"
	var secretName string = "secretRef"
	var namespace string = "namespace"
	return ReadOnlyConfig{
		MaximumTimerDrivenThreadCount: &num,
		MaximumEventDrivenThreadCount: &num,
		AdditionalSharedEnvs: []corev1.EnvVar{
			{
				Name:  "ENV_VAR",
				Value: "foo",
			},
		},
		NifiProperties: NifiProperties{
			OverrideConfigMap: &ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			OverrideConfigs: "nifi.prop.foo=value",
			OverrideSecretConfig: &SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
			WebProxyHosts: []string{
				"foo.host",
			},
			NeedClientAuth: true,
			Authorizer:     "blah-authorizer",
		},
		ZookeeperProperties: ZookeeperProperties{
			OverrideConfigMap: &ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			OverrideConfigs: "nifi.prop.foo=value",
			OverrideSecretConfig: &SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
		},
		BootstrapProperties: BootstrapProperties{
			OverrideConfigMap: &ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			OverrideConfigs: "nifi.prop.foo=value",
			OverrideSecretConfig: &SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
			NifiJvmMemory: "jvmMem",
		},
		LogbackConfig: LogbackConfig{
			ReplaceConfigMap: &ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			ReplaceSecretConfig: &SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
		},
		AuthorizerConfig: AuthorizerConfig{
			ReplaceTemplateConfigMap: &ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			ReplaceTemplateSecretConfig: &SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
		},
		BootstrapNotificationServicesReplaceConfig: BootstrapNotificationServicesConfig{
			ReplaceConfigMap: &ConfigmapReference{
				Name:      configMapName,
				Namespace: namespace,
			},
			ReplaceSecretConfig: &SecretConfigReference{
				Name:      secretName,
				Namespace: namespace,
			},
		},
	}
}

func assertReadOnlyConfigsEqual(roc v1.ReadOnlyConfig, alphaRoc ReadOnlyConfig, t *testing.T) {
	if roc.MaximumEventDrivenThreadCount != alphaRoc.MaximumEventDrivenThreadCount ||
		roc.MaximumTimerDrivenThreadCount != alphaRoc.MaximumTimerDrivenThreadCount {
		t.Error("Thread counts are not equal.")
	}

	if !reflect.DeepEqual(roc.AdditionalSharedEnvs, alphaRoc.AdditionalSharedEnvs) {
		t.Error("Additional shared env variables are not equal.")
	}

	rocAC := roc.AuthorizerConfig
	alphaRocAC := alphaRoc.AuthorizerConfig
	if rocAC.ReplaceTemplateConfigMap.Name != alphaRocAC.ReplaceTemplateConfigMap.Name ||
		rocAC.ReplaceTemplateConfigMap.Namespace != alphaRocAC.ReplaceTemplateConfigMap.Namespace ||
		rocAC.ReplaceTemplateConfigMap.Data != alphaRocAC.ReplaceTemplateConfigMap.Data ||
		rocAC.ReplaceTemplateSecretConfig.Name != alphaRocAC.ReplaceTemplateSecretConfig.Name ||
		rocAC.ReplaceTemplateSecretConfig.Namespace != alphaRocAC.ReplaceTemplateSecretConfig.Namespace ||
		rocAC.ReplaceTemplateSecretConfig.Data != alphaRocAC.ReplaceTemplateSecretConfig.Data {
		t.Error("Authorizer configs are not equal.")
	}

	rocBNSRC := roc.BootstrapNotificationServicesReplaceConfig
	alphaRocBNSRC := alphaRoc.BootstrapNotificationServicesReplaceConfig
	if rocBNSRC.ReplaceConfigMap.Name != alphaRocBNSRC.ReplaceConfigMap.Name ||
		rocBNSRC.ReplaceConfigMap.Namespace != alphaRocBNSRC.ReplaceConfigMap.Namespace ||
		rocBNSRC.ReplaceConfigMap.Data != alphaRocBNSRC.ReplaceConfigMap.Data ||
		rocBNSRC.ReplaceSecretConfig.Name != alphaRocBNSRC.ReplaceSecretConfig.Name ||
		rocBNSRC.ReplaceSecretConfig.Namespace != alphaRocBNSRC.ReplaceSecretConfig.Namespace ||
		rocBNSRC.ReplaceSecretConfig.Data != alphaRocBNSRC.ReplaceSecretConfig.Data {
		t.Error("Bootstrap notification services replace configs are not equal.")
	}

	rocNP := roc.NifiProperties
	alphaRocNP := alphaRoc.NifiProperties
	if rocNP.NeedClientAuth != alphaRocNP.NeedClientAuth ||
		rocNP.OverrideConfigs != alphaRocNP.OverrideConfigs ||
		!reflect.DeepEqual(rocNP.WebProxyHosts, alphaRocNP.WebProxyHosts) ||
		rocNP.OverrideConfigMap.Name != alphaRocNP.OverrideConfigMap.Name ||
		rocNP.OverrideConfigMap.Namespace != alphaRocNP.OverrideConfigMap.Namespace ||
		rocNP.OverrideConfigMap.Data != alphaRocNP.OverrideConfigMap.Data ||
		rocNP.OverrideSecretConfig.Name != alphaRocNP.OverrideSecretConfig.Name ||
		rocNP.OverrideSecretConfig.Namespace != alphaRocNP.OverrideSecretConfig.Namespace ||
		rocNP.OverrideSecretConfig.Data != alphaRocNP.OverrideSecretConfig.Data {
		t.Error("Nifi properties are not equal.")
	}

	rocBP := roc.BootstrapProperties
	alphaRocBP := alphaRoc.BootstrapProperties
	if rocBP.NifiJvmMemory != alphaRocBP.NifiJvmMemory ||
		rocBP.OverrideConfigMap.Name != alphaRocBP.OverrideConfigMap.Name ||
		rocBP.OverrideConfigMap.Namespace != alphaRocBP.OverrideConfigMap.Namespace ||
		rocBP.OverrideConfigMap.Data != alphaRocBP.OverrideConfigMap.Data ||
		rocBP.OverrideConfigs != alphaRocBP.OverrideConfigs ||
		rocBP.OverrideSecretConfig.Name != alphaRocBP.OverrideSecretConfig.Name ||
		rocBP.OverrideSecretConfig.Namespace != alphaRocBP.OverrideSecretConfig.Namespace ||
		rocBP.OverrideSecretConfig.Data != alphaRocBP.OverrideSecretConfig.Data {
		t.Error("Bootstrap properties are not equal.")
	}

	rocLC := roc.LogbackConfig
	alphaRocLC := alphaRoc.LogbackConfig
	if rocLC.ReplaceConfigMap.Name != alphaRocLC.ReplaceConfigMap.Name ||
		rocLC.ReplaceConfigMap.Namespace != alphaRocLC.ReplaceConfigMap.Namespace ||
		rocLC.ReplaceConfigMap.Data != alphaRocLC.ReplaceConfigMap.Data ||
		rocLC.ReplaceSecretConfig.Name != alphaRocLC.ReplaceSecretConfig.Name ||
		rocLC.ReplaceSecretConfig.Namespace != alphaRocLC.ReplaceSecretConfig.Namespace ||
		rocLC.ReplaceSecretConfig.Data != alphaRocLC.ReplaceSecretConfig.Data {
		t.Error("Logback configs are not equal")
	}

	rocZP := roc.ZookeeperProperties
	alphaRocZP := alphaRoc.ZookeeperProperties
	if rocZP.OverrideConfigs != alphaRocZP.OverrideConfigs ||
		rocZP.OverrideConfigMap.Name != alphaRocZP.OverrideConfigMap.Name ||
		rocZP.OverrideConfigMap.Namespace != alphaRocZP.OverrideConfigMap.Namespace ||
		rocZP.OverrideConfigMap.Data != alphaRocZP.OverrideConfigMap.Data ||
		rocZP.OverrideSecretConfig.Name != alphaRocZP.OverrideSecretConfig.Name ||
		rocZP.OverrideSecretConfig.Namespace != alphaRocZP.OverrideSecretConfig.Namespace ||
		rocZP.OverrideSecretConfig.Data != alphaRocZP.OverrideSecretConfig.Data {
		t.Error("Zookeeper properties are not equal")
	}
}

func TestGetClusterReference(t *testing.T) {
	clusterRef := v1.ClusterReference{
		Name:      "cluster",
		Namespace: "namespace",
	}
	alphaRef := getClusterReference(clusterRef)
	if clusterRef.Name != alphaRef.Name ||
		clusterRef.Namespace != alphaRef.Namespace {
		t.Error("cluster refs are not equal")
	}
}

func TestGetV1ClusterReference(t *testing.T) {
	clusterRef := ClusterReference{
		Name:      "cluster",
		Namespace: "namespace",
	}
	v1Ref := getV1ClusterReference(clusterRef)
	if clusterRef.Name != v1Ref.Name ||
		clusterRef.Namespace != v1Ref.Namespace {
		t.Error("cluster refs are not equal")
	}
}

func TestGetParameterContextRef(t *testing.T) {
	paramConRef := v1.ParameterContextReference{
		Name:      "pcr",
		Namespace: "namespace",
	}

	alphaRef := getParameterContextRef(paramConRef)
	if paramConRef.Name != alphaRef.Name ||
		paramConRef.Namespace != alphaRef.Namespace {
		t.Error("Parameter context references not equal")
	}
}

func TestGetV1ParameterContextRef(t *testing.T) {
	paramConRef := ParameterContextReference{
		Name:      "pcr",
		Namespace: "namespace",
	}

	v1Ref := getV1ParameterContextRef(paramConRef)
	if paramConRef.Name != v1Ref.Name ||
		paramConRef.Namespace != v1Ref.Namespace {
		t.Error("Parameter context references not equal")
	}
}

func TestGetAccessPolicy(t *testing.T) {
	ap := v1.AccessPolicy{
		Type:          v1.ComponentAccessPolicyType,
		Action:        v1.ReadAccessPolicyAction,
		Resource:      v1.ComponentsAccessPolicyResource,
		ComponentType: "component type",
		ComponentId:   "id",
	}

	alphaAp := getAccessPolicy(ap)
	if string(ap.Type) != string(alphaAp.Type) ||
		string(ap.Action) != string(alphaAp.Action) ||
		string(ap.Resource) != string(alphaAp.Resource) ||
		ap.ComponentType != alphaAp.ComponentType ||
		ap.ComponentId != alphaAp.ComponentId {
		t.Error("Access policies not equal")
	}
}

func TestGetV1AccessPolicy(t *testing.T) {
	ap := AccessPolicy{
		Type:          ComponentAccessPolicyType,
		Action:        ReadAccessPolicyAction,
		Resource:      ComponentsAccessPolicyResource,
		ComponentType: "component type",
		ComponentId:   "id",
	}

	v1Ap := getV1AccessPolicy(ap)
	if string(ap.Type) != string(v1Ap.Type) ||
		string(ap.Action) != string(v1Ap.Action) ||
		string(ap.Resource) != string(v1Ap.Resource) ||
		ap.ComponentType != v1Ap.ComponentType ||
		ap.ComponentId != v1Ap.ComponentId {
		t.Error("Access policies not equal")
	}
}
