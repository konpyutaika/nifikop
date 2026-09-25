package registryclient

import (
	"testing"

	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v2alpha1 "github.com/konpyutaika/nifikop/api/v2alpha1"
)

func stringPointer(s string) *string {
	return &s
}

func createAzureDevOpsRegistryClient() *v2alpha1.NifiRegistryClient {
	authStrategy := v2alpha1.AzureDevOpsAuthServicePrincipal
	paramValues := v2alpha1.RegistryClientParamRetain
	return &v2alpha1.NifiRegistryClient{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "registryClient",
			Namespace: "namespace",
		},
		Spec: v2alpha1.NifiRegistryClientSpec{
			Description: "description",
			Type:        v2alpha1.AzureDevOpsRegistryClientType,
			ClusterRef: v2alpha1.ClusterReference{
				Name:      "cluster",
				Namespace: "namespace",
			},
			AzureDevOpsConfig: &v2alpha1.AzureDevOpsConfig{
				ApiUrl:                 stringPointer("https://dev.azure.com"),
				Organization:           "my-org",
				Project:                "my-project",
				RepositoryName:         "nifi-flows",
				AuthenticationStrategy: &authStrategy,
				OAuthTokenProviderId:   stringPointer("oauth-provider-id"),
				WebClientServiceId:     "web-client-service-id",
				DefaultBranch:          stringPointer("main"),
				ParameterContextValues: &paramValues,
			},
		},
	}
}

func TestUpdateRegistryClientEntityAzureDevOps(t *testing.T) {
	rc := createAzureDevOpsRegistryClient()

	entity := &nigoapi.FlowRegistryClientEntity{}
	updateRegistryClientEntity(rc, nil, entity)

	if entity.Component.Type_ != "org.apache.nifi.azure.devops.AzureDevOpsFlowRegistryClient" {
		t.Error("component type not equal to the AzureDevOpsFlowRegistryClient class name")
	}
	if entity.Component.Name != rc.Name {
		t.Error("component name not equal")
	}
	if entity.Component.Description != rc.Spec.Description {
		t.Error("component description not equal")
	}

	expectedProperties := map[string]string{
		"Azure DevOps API URL":         "https://dev.azure.com",
		"Organization":                 "my-org",
		"Project":                      "my-project",
		"Repository Name":              "nifi-flows",
		"Authentication Strategy":      "SERVICE_PRINCIPAL",
		"OAuth2 Access Token Provider": "oauth-provider-id",
		"Web Client Service":           "web-client-service-id",
		"Default Branch":               "main",
		"Parameter Context Values":     "RETAIN",
	}
	for key, expected := range expectedProperties {
		if actual := entity.Component.Properties[key]; actual != expected {
			t.Errorf("property %q not equal: expected %q, got %q", key, expected, actual)
		}
	}
	if len(entity.Component.Properties) != len(expectedProperties) {
		t.Errorf("unexpected number of properties: expected %d, got %d", len(expectedProperties), len(entity.Component.Properties))
	}
}

func TestRegistryClientIsSyncAzureDevOps(t *testing.T) {
	rc := createAzureDevOpsRegistryClient()

	entity := &nigoapi.FlowRegistryClientEntity{}
	updateRegistryClientEntity(rc, nil, entity)

	if !registryClientIsSync(rc, nil, entity) {
		t.Error("registry client should be in sync with the entity built from it")
	}

	rc.Spec.AzureDevOpsConfig.Organization = "other-org"
	if registryClientIsSync(rc, nil, entity) {
		t.Error("registry client should be out of sync after changing the organization")
	}
}
