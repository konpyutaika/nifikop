package controllerservice

import (
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	corev1 "k8s.io/api/core/v1"

	v2alpha1 "github.com/konpyutaika/nifikop/api/v2alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("controllerservice-method")

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

func ExistControllerService(controllerService *v2alpha1.NifiControllerService, config *clientconfig.NifiConfig) (bool, error) {
	if controllerService.Status.Id == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetControllerService(controllerService.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get controller-service"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

func CreateControllerService(controllerService *v2alpha1.NifiControllerService,
	secrets map[string]*corev1.Secret,
	config *clientconfig.NifiConfig) (*v2alpha1.NifiControllerServiceStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.ControllerServiceEntity{}
	updateControllerServiceEntity(controllerService, secrets, &scratchEntity)

	entity, err := nClient.CreateControllerService(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Failed to create controller-service "+controllerService.Name); err != nil {
		return nil, err
	}

	return &v2alpha1.NifiControllerServiceStatus{
		Id:                           entity.Id,
		Version:                      *entity.Revision.Version,
		LatestSecretsResourceVersion: extractSecretsResourceVersion(secrets),
	}, nil
}

func SyncControllerService(controllerService *v2alpha1.NifiControllerService,
	secrets map[string]*corev1.Secret,
	config *clientconfig.NifiConfig) (*v2alpha1.NifiControllerServiceStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.GetControllerService(controllerService.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get controller-service"); err != nil {
		return nil, err
	}

	if !controllerServiceIsSync(controllerService, secrets, entity) {
		updateControllerServiceEntity(controllerService, secrets, entity)
		entity, err = nClient.UpdateControllerService(*entity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update controller-service"); err != nil {
			return nil, err
		}
	}

	status := controllerService.Status
	status.Version = *entity.Revision.Version
	status.Id = entity.Id
	status.LatestSecretsResourceVersion = extractSecretsResourceVersion(secrets)

	return &status, nil
}

func RemoveControllerService(controllerService *v2alpha1.NifiControllerService,
	config *clientconfig.NifiConfig) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	entity, err := nClient.GetControllerService(controllerService.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get controller-service"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil
		}
		return err
	}

	updateControllerServiceEntity(controllerService, nil, entity)
	err = nClient.RemoveControllerService(*entity)

	return clientwrappers.ErrorRemoveOperation(log, err, "Remove controller-service")
}

func controllerServiceIsSync(controllerService *v2alpha1.NifiControllerService, secrets map[string]*corev1.Secret, entity *nigoapi.ControllerServiceEntity) bool {
	if controllerService.Name != entity.Component.Name ||
		controllerService.Spec.GetType() != entity.Component.Type_ {
		return false
	}

	if isSecretResourceVersionUpdated(secrets, controllerService.Status.LatestSecretsResourceVersion) {
		return false
	}

	switch controllerService.Spec.Type {
	case v2alpha1.StandardWebClientServiceProviderType:
		return controllerServiceIsSync_StandardWebClientServiceProvider(controllerService.Spec.StandardWebClientServiceProviderSpec, entity)
	}
	return true
}

var StandardWebClientServiceProperties = struct {
	ConnectTimeout string
	ReadTimeout    string
	WriteTimeout   string
}{
	ConnectTimeout: "Connect Timeout",
	ReadTimeout:    "Read Timeout",
	WriteTimeout:   "Write Timeout",
}

func controllerServiceIsSync_StandardWebClientServiceProvider(cfg *v2alpha1.StandardWebClientServiceProviderSpec, entity *nigoapi.ControllerServiceEntity) bool {
	if cfg == nil {
		return true
	}
	return cfg.ConnectTimeout == entity.Component.Properties[StandardWebClientServiceProperties.ConnectTimeout] &&
		cfg.ReadTimeout == entity.Component.Properties[StandardWebClientServiceProperties.ReadTimeout] &&
		cfg.WriteTimeout == entity.Component.Properties[StandardWebClientServiceProperties.WriteTimeout]
}

func updateControllerServiceEntity(controllerService *v2alpha1.NifiControllerService, secrets map[string]*corev1.Secret, entity *nigoapi.ControllerServiceEntity) {
	var defaultVersion int64 = 0

	if entity == nil {
		entity = &nigoapi.ControllerServiceEntity{}
	}

	if entity.Revision == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.ControllerServiceDto{
			Type_: controllerService.Spec.GetType(),
		}
	}

	entity.Component.Name = controllerService.Name

	entity.Component.Properties = make(map[string]string)

	switch controllerService.Spec.Type {
	case v2alpha1.StandardWebClientServiceProviderType:
		updateEntity_StandardWebClientServiceProvider(controllerService.Spec.StandardWebClientServiceProviderSpec, entity)
	}
}

func updateEntity_StandardWebClientServiceProvider(spec *v2alpha1.StandardWebClientServiceProviderSpec, entity *nigoapi.ControllerServiceEntity) {
	if spec == nil {
		return
	}
	entity.Component.Properties[StandardWebClientServiceProperties.ConnectTimeout] = spec.ConnectTimeout
	entity.Component.Properties[StandardWebClientServiceProperties.ReadTimeout] = spec.ReadTimeout
	entity.Component.Properties[StandardWebClientServiceProperties.WriteTimeout] = spec.WriteTimeout
}
