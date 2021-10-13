package registryclient

import (
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers"
	"github.com/Orange-OpenSource/nifikop/pkg/common"
	"github.com/Orange-OpenSource/nifikop/pkg/nificlient"
	"github.com/Orange-OpenSource/nifikop/pkg/util/clientconfig"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("registryclient-method")

func ExistRegistryClient(registryClient *v1alpha1.NifiRegistryClient, config *clientconfig.NifiConfig) (bool, error) {

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

func CreateRegistryClient(registryClient *v1alpha1.NifiRegistryClient,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiRegistryClientStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.RegistryClientEntity{}
	updateRegistryClientEntity(registryClient, &scratchEntity)

	entity, err := nClient.CreateRegistryClient(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Create registry-client"); err != nil {
		return nil, err
	}

	return &v1alpha1.NifiRegistryClientStatus{
		Id:      entity.Id,
		Version: *entity.Revision.Version,
	}, nil
}

func SyncRegistryClient(registryClient *v1alpha1.NifiRegistryClient,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiRegistryClientStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.GetRegistryClient(registryClient.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get registry-client"); err != nil {
		return nil, err
	}

	if !registryClientIsSync(registryClient, entity) {
		updateRegistryClientEntity(registryClient, entity)
		entity, err = nClient.UpdateRegistryClient(*entity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update registry-client"); err != nil {
			return nil, err
		}
	}

	status := registryClient.Status
	status.Version = *entity.Revision.Version
	status.Id = entity.Id

	return &status, nil
}

func RemoveRegistryClient(registryClient *v1alpha1.NifiRegistryClient,
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

	updateRegistryClientEntity(registryClient, entity)
	err = nClient.RemoveRegistryClient(*entity)

	return clientwrappers.ErrorRemoveOperation(log, err, "Remove registry-client")
}

func registryClientIsSync(registryClient *v1alpha1.NifiRegistryClient, entity *nigoapi.RegistryClientEntity) bool {
	return registryClient.Name == entity.Component.Name &&
		registryClient.Spec.Description == entity.Component.Description &&
		registryClient.Spec.Uri == entity.Component.Uri
}

func updateRegistryClientEntity(registryClient *v1alpha1.NifiRegistryClient, entity *nigoapi.RegistryClientEntity) {

	var defaultVersion int64 = 0

	if entity == nil {
		entity = &nigoapi.RegistryClientEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.RegistryDto{}
	}

	entity.Component.Name = registryClient.Name
	entity.Component.Description = registryClient.Spec.Description
	entity.Component.Uri = registryClient.Spec.Uri
}
