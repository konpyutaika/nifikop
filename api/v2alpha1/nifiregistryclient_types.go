package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NifiRegistryClientSpec defines the desired state of NifiRegistryClient.
// +kubebuilder:validation:XValidation:rule="self.type != 'registry' || has(self.registryClientConfig)",message="registryClientConfig is required when type is 'registry'"
// +kubebuilder:validation:XValidation:rule="self.type != 'github' || has(self.githubConfig)",message="githubConfig is required when type is 'github'"
// +kubebuilder:validation:XValidation:rule="self.type != 'gitlab' || has(self.gitlabConfig)",message="gitlabConfig is required when type is 'gitlab'"
type NifiRegistryClientSpec struct {
	// The description of the registry client.
	// +optional
	Description string `json:"description,omitempty"`
	// Reference to the NifiCluster this registry client is linked to.
	// +optional
	ClusterRef ClusterReference `json:"clusterRef,omitempty"`
	// Type of the registry client.
	// +kubebuilder:validation:Enum=registry;github;gitlab
	// +kubebuilder:default=registry
	// +optional
	Type NifiRegistryClientType `json:"type,omitempty"`

	// RegistryClientConfig holds configuration for a NiFi Registry type client.
	// Required when type is "registry".
	// +optional
	RegistryClientConfig *RegistryClientConfig `json:"registryClientConfig,omitempty"`

	// GitHubConfig holds configuration for a GitHub type client.
	// Required when type is "github".
	// +optional
	GitHubConfig *GitHubConfig `json:"githubConfig,omitempty"`

	// GitLabConfig holds configuration for a GitLab type client.
	// Required when type is "gitlab".
	// +optional
	GitLabConfig *GitLabConfig `json:"gitlabConfig,omitempty"`
}

// RegistryClientConfig holds configuration for a NiFi Registry server client.
type RegistryClientConfig struct {
	// The URI of the NiFi Registry server.
	Uri string `json:"uri"`
}

// GitHubConfig holds configuration for a GitHub flow registry client.
// +kubebuilder:validation:XValidation:rule="!has(self.authenticationType) || self.authenticationType != 'PERSONAL_ACCESS_TOKEN' || has(self.personalAccessTokenSecretRef)",message="personalAccessTokenSecretRef is required when authenticationType is PERSONAL_ACCESS_TOKEN"
// +kubebuilder:validation:XValidation:rule="!has(self.authenticationType) || self.authenticationType != 'APP_INSTALLATION' || (has(self.appId) && has(self.appPrivateKeySecretRef))",message="appId and appPrivateKeySecretRef are required when authenticationType is APP_INSTALLATION"
type GitHubConfig struct {
	// URL of the GitHub API. Defaults to https://api.github.com/.
	// +optional
	ApiUrl *string `json:"apiUrl,omitempty"`
	// Owner of the repository (user or organization).
	RepositoryOwner string `json:"repositoryOwner"`
	// Name of the repository.
	RepositoryName string `json:"repositoryName"`
	// Type of authentication to use.
	// +optional
	AuthenticationType *GitHubAuthenticationType `json:"authenticationType,omitempty"`
	// Reference to a Kubernetes Secret containing the personal access token.
	// Required when authenticationType is PERSONAL_ACCESS_TOKEN.
	// +optional
	PersonalAccessTokenSecretRef *SecretConfigReference `json:"personalAccessTokenSecretRef,omitempty"`
	// Identifier of the GitHub App.
	// Required when authenticationType is "App Installation".
	// +optional
	AppId *string `json:"appId,omitempty"`
	// Reference to a Kubernetes Secret containing the RSA private key for the GitHub App.
	// Required when authenticationType is "App Installation".
	// +optional
	AppPrivateKeySecretRef *SecretConfigReference `json:"appPrivateKeySecretRef,omitempty"`
	// Default branch of the repository.
	// +optional
	DefaultBranch *string `json:"defaultBranch,omitempty"`
	// Path within the repository for storing data. Defaults to repository root.
	// +optional
	RepositoryPath *string `json:"repositoryPath,omitempty"`
	// Regex pattern for directories to exclude. Defaults to [.].* (hidden directories).
	// +optional
	DirectoryFilterExclusion *string `json:"directoryFilterExclusion,omitempty"`
	// How to handle parameter context values.
	// +optional
	ParameterContextValues *RegistryClientParameterContextValues `json:"parameterContextValues,omitempty"`
}

// GitLabConfig holds configuration for a GitLab flow registry client.
// +kubebuilder:validation:XValidation:rule="(has(self.authenticationType) && self.authenticationType != 'ACCESS_TOKEN') || has(self.accessTokenSecretRef)",message="accessTokenSecretRef is required when authenticationType is ACCESS_TOKEN or not set"
type GitLabConfig struct {
	// URL of the GitLab API. Defaults to https://gitlab.com/.
	// +optional
	Url *string `json:"url,omitempty"`
	// GitLab API version.
	// +optional
	ApiVersion *GitLabApiVersion `json:"apiVersion,omitempty"`
	// Namespace of the repository (user or group/subgroup path).
	RepositoryNamespace string `json:"repositoryNamespace"`
	// Name of the repository.
	RepositoryName string `json:"repositoryName"`
	// Type of authentication to use.
	// +optional
	AuthenticationType *GitLabAuthenticationType `json:"authenticationType,omitempty"`
	// Reference to a Kubernetes Secret containing the access token.
	// Required when authenticationType is ACCESS_TOKEN.
	// +optional
	AccessTokenSecretRef *SecretConfigReference `json:"accessTokenSecretRef,omitempty"`
	// Connect timeout (e.g. "10 seconds").
	// +optional
	ConnectTimeout *string `json:"connectTimeout,omitempty"`
	// Read timeout (e.g. "10 seconds").
	// +optional
	ReadTimeout *string `json:"readTimeout,omitempty"`
	// Default branch of the repository.
	// +optional
	DefaultBranch *string `json:"defaultBranch,omitempty"`
	// Path within the repository for storing data. Defaults to repository root.
	// +optional
	RepositoryPath *string `json:"repositoryPath,omitempty"`
	// Regex pattern for directories to exclude. Defaults to [.].* (hidden directories).
	// +optional
	DirectoryFilterExclusion *string `json:"directoryFilterExclusion,omitempty"`
	// How to handle parameter context values.
	// +optional
	ParameterContextValues *RegistryClientParameterContextValues `json:"parameterContextValues,omitempty"`
}

// NifiRegistryClientStatus defines the observed state of NifiRegistryClient.
type NifiRegistryClientStatus struct {
	// The nifi registry client's id.
	Id string `json:"id"`
	// The last nifi registry client revision version caught.
	Version int64 `json:"version"`
	// The last observed resource versions of the referenced secrets.
	// +optional
	LatestSecretsResourceVersion []SecretResourceVersion `json:"latestSecretsResourceVersion,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// NifiRegistryClient is the Schema for the nifiregistryclients API.
type NifiRegistryClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NifiRegistryClientSpec   `json:"spec,omitempty"`
	Status NifiRegistryClientStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NifiRegistryClientList contains a list of NifiRegistryClient.
type NifiRegistryClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NifiRegistryClient `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NifiRegistryClient{}, &NifiRegistryClientList{})
}

// GetType returns the NiFi API type string (full class name) for this registry client.
func (s *NifiRegistryClientSpec) GetType() string {
	switch s.Type {
	case GitHubRegistryClientType:
		return "org.apache.nifi.github.GitHubFlowRegistryClient"
	case GitLabRegistryClientType:
		return "org.apache.nifi.gitlab.GitLabFlowRegistryClient"
	default: // RegistryClientType
		return "org.apache.nifi.registry.flow.NifiRegistryFlowRegistryClient"
	}
}
