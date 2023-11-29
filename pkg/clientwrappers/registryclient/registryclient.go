package registryclient

import (
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("registryclient-method")

func ExistRegistryClient(registryClient *v1.NifiRegistryClient, config *clientconfig.NifiConfig) (bool, error) {
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

func CreateRegistryClient(registryClient *v1.NifiRegistryClient,
	config *clientconfig.NifiConfig) (*v1.NifiRegistryClientStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.FlowRegistryClientEntity{}
	updateRegistryClientEntity(registryClient, &scratchEntity)

	entity, err := nClient.CreateRegistryClient(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Failed to create registry-client "+registryClient.Name); err != nil {
		return nil, err
	}

	return &v1.NifiRegistryClientStatus{
		Id:      entity.Id,
		Version: *entity.Revision.Version,
	}, nil
}

func SyncRegistryClient(registryClient *v1.NifiRegistryClient,
	config *clientconfig.NifiConfig) (*v1.NifiRegistryClientStatus, error) {
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

func RemoveRegistryClient(registryClient *v1.NifiRegistryClient,
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

func registryClientIsSync(registryClient *v1.NifiRegistryClient, entity *nigoapi.FlowRegistryClientEntity) bool {
	return registryClient.Name == entity.Component.Name &&
		registryClient.Spec.Description == entity.Component.Description &&
		registryClient.Spec.Uri == entity.Component.Uri
}

func updateRegistryClientEntity(registryClient *v1.NifiRegistryClient, entity *nigoapi.FlowRegistryClientEntity) {
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
		entity.Component = &nigoapi.FlowRegistryClientDto{}
	}

	entity.Component.Properties = make(map[string]string)

	entity.Component.Name = registryClient.Name
	entity.Component.Description = registryClient.Spec.Description
	entity.Component.Uri = registryClient.Spec.Uri
	entity.Component.Properties["url"] = registryClient.Spec.Uri
}
