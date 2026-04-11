package registryclient

import (
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	corev1 "k8s.io/api/core/v1"

	v2alpha1 "github.com/konpyutaika/nifikop/api/v2alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("registryclient-method")

func extractSecretsResourceVersion(secrets map[string]*corev1.Secret) []v2alpha1.SecretResourceVersion {
	result := make([]v2alpha1.SecretResourceVersion, 0, len(secrets))
	for _, secret := range secrets {
		result = append(result, v2alpha1.SecretResourceVersion{
			Name:            secret.Name,
			Namespace:       secret.Namespace,
			ResourceVersion: secret.ResourceVersion,
		})
	}
	return result
}

func isSecretResourceVersionUpdated(secrets map[string]*corev1.Secret, latest []v2alpha1.SecretResourceVersion) bool {
	if len(secrets) != len(latest) {
		return true
	}
	for _, srv := range latest {
		secret, ok := secrets[srv.Name]
		if !ok || secret.ResourceVersion != srv.ResourceVersion {
			return true
		}
	}
	return false
}

func secretValue(ref *v2alpha1.SecretConfigReference, secrets map[string]*corev1.Secret) string {
	if ref == nil || secrets == nil {
		return ""
	}
	secret := secrets[ref.Name]
	if secret == nil {
		return ""
	}
	return string(secret.Data[ref.Data])
}

func ExistRegistryClient(registryClient *v2alpha1.NifiRegistryClient, config *clientconfig.NifiConfig) (bool, error) {
	if registryClient.Status.Id == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetRegistryClient(registryClient.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get registry-client"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

func CreateRegistryClient(registryClient *v2alpha1.NifiRegistryClient,
	secrets map[string]*corev1.Secret,
	config *clientconfig.NifiConfig) (*v2alpha1.NifiRegistryClientStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.FlowRegistryClientEntity{}
	updateRegistryClientEntity(registryClient, secrets, &scratchEntity)

	entity, err := nClient.CreateRegistryClient(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Failed to create registry-client "+registryClient.Name); err != nil {
		return nil, err
	}

	return &v2alpha1.NifiRegistryClientStatus{
		Id:                           entity.Id,
		Version:                      *entity.Revision.Version,
		LatestSecretsResourceVersion: extractSecretsResourceVersion(secrets),
	}, nil
}

func SyncRegistryClient(registryClient *v2alpha1.NifiRegistryClient,
	secrets map[string]*corev1.Secret,
	config *clientconfig.NifiConfig) (*v2alpha1.NifiRegistryClientStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.GetRegistryClient(registryClient.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get registry-client"); err != nil {
		return nil, err
	}

	if !registryClientIsSync(registryClient, secrets, entity) {
		updateRegistryClientEntity(registryClient, secrets, entity)
		entity, err = nClient.UpdateRegistryClient(*entity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update registry-client"); err != nil {
			return nil, err
		}
	}

	status := registryClient.Status
	status.Version = *entity.Revision.Version
	status.Id = entity.Id
	status.LatestSecretsResourceVersion = extractSecretsResourceVersion(secrets)

	return &status, nil
}

func RemoveRegistryClient(registryClient *v2alpha1.NifiRegistryClient,
	config *clientconfig.NifiConfig) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	entity, err := nClient.GetRegistryClient(registryClient.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get registry-client"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil
		}
		return err
	}

	updateRegistryClientEntity(registryClient, nil, entity)
	err = nClient.RemoveRegistryClient(*entity)

	return clientwrappers.ErrorRemoveOperation(log, err, "Remove registry-client")
}

func registryClientIsSync(registryClient *v2alpha1.NifiRegistryClient, secrets map[string]*corev1.Secret, entity *nigoapi.FlowRegistryClientEntity) bool {
	if registryClient.Name != entity.Component.Name ||
		registryClient.Spec.Description != entity.Component.Description ||
		registryClient.Spec.GetType() != entity.Component.Type_ {
		return false
	}

	if isSecretResourceVersionUpdated(secrets, registryClient.Status.LatestSecretsResourceVersion) {
		return false
	}

	switch registryClient.Spec.Type {
	case v2alpha1.RegistryClientType:
		return registryClientIsSync_Registry(registryClient.Spec.RegistryClientConfig, entity)
	case v2alpha1.GitHubRegistryClientType:
		return registryClientIsSync_GitHub(registryClient.Spec.GitHubConfig, entity)
	case v2alpha1.GitLabRegistryClientType:
		return registryClientIsSync_GitLab(registryClient.Spec.GitLabConfig, entity)
	}
	return true
}

func registryClientIsSync_Registry(cfg *v2alpha1.RegistryClientConfig, entity *nigoapi.FlowRegistryClientEntity) bool {
	if cfg == nil {
		return true
	}
	return cfg.Uri == entity.Component.Properties["url"]
}

func registryClientIsSync_GitHub(cfg *v2alpha1.GitHubConfig, entity *nigoapi.FlowRegistryClientEntity) bool {
	if cfg == nil {
		return true
	}
	return (cfg.ApiUrl == nil || *cfg.ApiUrl == entity.Component.Properties["GitHub API URL"]) &&
		cfg.RepositoryOwner == entity.Component.Properties["Repository Owner"] &&
		cfg.RepositoryName == entity.Component.Properties["Repository Name"] &&
		(cfg.AuthenticationType == nil || string(*cfg.AuthenticationType) == entity.Component.Properties["Authentication Type"]) &&
		(cfg.AppId == nil || *cfg.AppId == entity.Component.Properties["App ID"]) &&
		(cfg.DefaultBranch == nil || *cfg.DefaultBranch == entity.Component.Properties["Default Branch"]) &&
		(cfg.RepositoryPath == nil || *cfg.RepositoryPath == entity.Component.Properties["Repository Path"]) &&
		(cfg.DirectoryFilterExclusion == nil || *cfg.DirectoryFilterExclusion == entity.Component.Properties["Directory Filter Exclusion"]) &&
		(cfg.ParameterContextValues == nil || string(*cfg.ParameterContextValues) == entity.Component.Properties["Parameter Context Values"])
}

func registryClientIsSync_GitLab(cfg *v2alpha1.GitLabConfig, entity *nigoapi.FlowRegistryClientEntity) bool {
	if cfg == nil {
		return true
	}
	return (cfg.Url == nil || *cfg.Url == entity.Component.Properties["GitLab API URL"]) &&
		(cfg.ApiVersion == nil || string(*cfg.ApiVersion) == entity.Component.Properties["GitLab API Version"]) &&
		cfg.RepositoryNamespace == entity.Component.Properties["Repository Namespace"] &&
		cfg.RepositoryName == entity.Component.Properties["Repository Name"] &&
		(cfg.AuthenticationType == nil || string(*cfg.AuthenticationType) == entity.Component.Properties["Authentication Type"]) &&
		(cfg.ConnectTimeout == nil || *cfg.ConnectTimeout == entity.Component.Properties["Connect Timeout"]) &&
		(cfg.ReadTimeout == nil || *cfg.ReadTimeout == entity.Component.Properties["Read Timeout"]) &&
		(cfg.DefaultBranch == nil || *cfg.DefaultBranch == entity.Component.Properties["Default Branch"]) &&
		(cfg.RepositoryPath == nil || *cfg.RepositoryPath == entity.Component.Properties["Repository Path"]) &&
		(cfg.DirectoryFilterExclusion == nil || *cfg.DirectoryFilterExclusion == entity.Component.Properties["Directory Filter Exclusion"]) &&
		(cfg.ParameterContextValues == nil || string(*cfg.ParameterContextValues) == entity.Component.Properties["Parameter Context Values"])
}

func updateRegistryClientEntity(registryClient *v2alpha1.NifiRegistryClient, secrets map[string]*corev1.Secret, entity *nigoapi.FlowRegistryClientEntity) {
	var defaultVersion int64 = 0

	if entity == nil {
		entity = &nigoapi.FlowRegistryClientEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.FlowRegistryClientDto{
			Type_: registryClient.Spec.GetType(),
		}
	}

	entity.Component.Properties = make(map[string]string)
	entity.Component.Name = registryClient.Name
	entity.Component.Description = registryClient.Spec.Description

	switch registryClient.Spec.Type {
	case v2alpha1.RegistryClientType:
		updateEntity_Registry(registryClient.Spec.RegistryClientConfig, entity)
	case v2alpha1.GitHubRegistryClientType:
		updateEntity_GitHub(registryClient.Spec.GitHubConfig, secrets, entity)
	case v2alpha1.GitLabRegistryClientType:
		updateEntity_GitLab(registryClient.Spec.GitLabConfig, secrets, entity)
	}
}

func updateEntity_Registry(cfg *v2alpha1.RegistryClientConfig, entity *nigoapi.FlowRegistryClientEntity) {
	if cfg == nil {
		return
	}
	entity.Component.Properties["url"] = cfg.Uri
}

func updateEntity_GitHub(cfg *v2alpha1.GitHubConfig, secrets map[string]*corev1.Secret, entity *nigoapi.FlowRegistryClientEntity) {
	if cfg == nil {
		return
	}
	entity.Component.Properties["Repository Owner"] = cfg.RepositoryOwner
	entity.Component.Properties["Repository Name"] = cfg.RepositoryName
	if cfg.AuthenticationType != nil {
		entity.Component.Properties["Authentication Type"] = string(*cfg.AuthenticationType)
	}
	if cfg.DefaultBranch != nil {
		entity.Component.Properties["Default Branch"] = *cfg.DefaultBranch
	}
	if cfg.ApiUrl != nil {
		entity.Component.Properties["GitHub API URL"] = *cfg.ApiUrl
	}
	if cfg.AppId != nil {
		entity.Component.Properties["App ID"] = *cfg.AppId
	}
	if pat := secretValue(cfg.PersonalAccessTokenSecretRef, secrets); pat != "" {
		entity.Component.Properties["Personal Access Token"] = pat
	}
	if key := secretValue(cfg.AppPrivateKeySecretRef, secrets); key != "" {
		entity.Component.Properties["App Private Key"] = key
	}
	if cfg.RepositoryPath != nil {
		entity.Component.Properties["Repository Path"] = *cfg.RepositoryPath
	}
	if cfg.DirectoryFilterExclusion != nil {
		entity.Component.Properties["Directory Filter Exclusion"] = *cfg.DirectoryFilterExclusion
	}
	if cfg.ParameterContextValues != nil {
		entity.Component.Properties["Parameter Context Values"] = string(*cfg.ParameterContextValues)
	}
}

func updateEntity_GitLab(cfg *v2alpha1.GitLabConfig, secrets map[string]*corev1.Secret, entity *nigoapi.FlowRegistryClientEntity) {
	if cfg == nil {
		return
	}
	if cfg.Url != nil {
		entity.Component.Properties["GitLab API URL"] = *cfg.Url
	}
	if cfg.ApiVersion != nil {
		entity.Component.Properties["GitLab API Version"] = string(*cfg.ApiVersion)
	}
	entity.Component.Properties["Repository Namespace"] = cfg.RepositoryNamespace
	entity.Component.Properties["Repository Name"] = cfg.RepositoryName
	if cfg.AuthenticationType != nil {
		entity.Component.Properties["Authentication Type"] = string(*cfg.AuthenticationType)
	}
	if token := secretValue(cfg.AccessTokenSecretRef, secrets); token != "" {
		entity.Component.Properties["Access Token"] = token
	}
	if cfg.ConnectTimeout != nil {
		entity.Component.Properties["Connect Timeout"] = *cfg.ConnectTimeout
	}
	if cfg.ReadTimeout != nil {
		entity.Component.Properties["Read Timeout"] = *cfg.ReadTimeout
	}
	if cfg.DefaultBranch != nil {
		entity.Component.Properties["Default Branch"] = *cfg.DefaultBranch
	}
	if cfg.RepositoryPath != nil {
		entity.Component.Properties["Repository Path"] = *cfg.RepositoryPath
	}
	if cfg.DirectoryFilterExclusion != nil {
		entity.Component.Properties["Directory Filter Exclusion"] = *cfg.DirectoryFilterExclusion
	}
	if cfg.ParameterContextValues != nil {
		entity.Component.Properties["Parameter Context Values"] = string(*cfg.ParameterContextValues)
	}
}

