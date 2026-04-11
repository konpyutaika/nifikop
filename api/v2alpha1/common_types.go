package v2alpha1

// NifiRegistryClientType defines the type of registry client.
// +kubebuilder:validation:Enum={"registry","github","gitlab"}
type NifiRegistryClientType string

const (
	// RegistryClientType indicates a NiFi Registry server.
	RegistryClientType NifiRegistryClientType = "registry"
	// GitHubRegistryClientType indicates a GitHub repository.
	GitHubRegistryClientType NifiRegistryClientType = "github"
	// GitLabRegistryClientType indicates a GitLab repository.
	GitLabRegistryClientType NifiRegistryClientType = "gitlab"
)

// ClusterReference states a reference to a cluster for registry client provisioning.
type ClusterReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// ClusterRefsEquals returns true if all ClusterReferences point to the same cluster.
func ClusterRefsEquals(clusterRefs []ClusterReference) bool {
	c1 := clusterRefs[0]
	name := c1.Name
	ns := c1.Namespace
	for _, cluster := range clusterRefs {
		if name != cluster.Name || ns != cluster.Namespace {
			return false
		}
	}
	return true
}

// GitHubAuthenticationType defines the authentication method for GitHub.
// +kubebuilder:validation:Enum=NONE;PERSONAL_ACCESS_TOKEN;APP_INSTALLATION
type GitHubAuthenticationType string

const (
	GitHubAuthNone                GitHubAuthenticationType = "NONE"
	GitHubAuthPersonalAccessToken GitHubAuthenticationType = "PERSONAL_ACCESS_TOKEN"
	GitHubAuthAppInstallation     GitHubAuthenticationType = "APP_INSTALLATION"
)

// RegistryClientParameterContextValues defines how to handle parameter context values.
// +kubebuilder:validation:Enum=RETAIN;REMOVE;IGNORE_CHANGES
type RegistryClientParameterContextValues string

const (
	RegistryClientParamRetain        RegistryClientParameterContextValues = "RETAIN"
	RegistryClientParamRemove        RegistryClientParameterContextValues = "REMOVE"
	RegistryClientParamIgnoreChanges RegistryClientParameterContextValues = "IGNORE_CHANGES"
)

// GitLabApiVersion defines the GitLab API version.
// +kubebuilder:validation:Enum=V4
type GitLabApiVersion string

const (
	GitLabApiVersionV4 GitLabApiVersion = "V4"
)

// GitLabAuthenticationType defines the authentication method for GitLab.
// +kubebuilder:validation:Enum=ACCESS_TOKEN
type GitLabAuthenticationType string

const (
	GitLabAuthAccessToken GitLabAuthenticationType = "ACCESS_TOKEN"
)

// SecretResourceVersion states the resourceVersion of a secret at last sync.
type SecretResourceVersion struct {
	// Name of the secret.
	Name string `json:"name"`
	// Namespace where the secret is located.
	Namespace string `json:"namespace"`
	// ResourceVersion of the secret.
	ResourceVersion string `json:"resourceVersion"`
}

// SecretConfigReference states a reference to a value stored in a Kubernetes Secret.
type SecretConfigReference struct {
	// Name of the secret that we want to refer.
	Name string `json:"name"`
	// Namespace where is located the secret that we want to refer.
	Namespace string `json:"namespace,omitempty"`
	// The key of the value, in data content, that we want to use.
	Data string `json:"data"`
}
