package processgroup

import (
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"

	v1alpha1 "github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("processgroup-method")

func ExistProcessGroup(resource *v1alpha1.NifiResource, config *clientconfig.NifiConfig) (bool, error) {
	if resource.Status.Id == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetProcessGroup(resource.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process-group"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

func CreateProcessGroup(resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.ProcessGroupEntity{}
	updateProcessGroupEntity(resource, &scratchEntity)

	entity, err := nClient.CreateProcessGroup(scratchEntity, resource.Spec.GetParentProcessGroupID(config.RootProcessGroupId))
	if err := clientwrappers.ErrorCreateOperation(log, err, "Failed to create resource "+resource.Name); err != nil {
		return nil, err
	}

	return &v1alpha1.NifiResourceStatus{
		Id:      entity.Id,
		Version: *entity.Revision.Version,
	}, nil
}

func SyncProcessGroup(resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiResourceStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.GetProcessGroup(resource.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get process-group"); err != nil {
		return nil, err
	}

	if !processGroupIsSync(resource, entity) {
		updateProcessGroupEntity(resource, entity)
		entity, err = nClient.UpdateProcessGroup(*entity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update process-group"); err != nil {
			return nil, err
		}
	}

	status := resource.Status
	status.Version = *entity.Revision.Version
	status.Id = entity.Id

	return &status, nil
}

func RemoveProcessGroup(resource *v1alpha1.NifiResource,
	config *clientconfig.NifiConfig) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	entity, err := nClient.GetProcessGroup(resource.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get resource"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil
		}
		return err
	}

	updateProcessGroupEntity(resource, entity)
	err = nClient.RemoveProcessGroup(*entity)

	return clientwrappers.ErrorRemoveOperation(log, err, "Remove resource")
}

func processGroupIsSync(resource *v1alpha1.NifiResource, entity *nigoapi.ProcessGroupEntity) bool {
	return resource.GetName() == entity.Component.Name
}

func updateProcessGroupEntity(resource *v1alpha1.NifiResource, entity *nigoapi.ProcessGroupEntity) {
	var defaultVersion int64 = 0

	if entity == nil {
		entity = &nigoapi.ProcessGroupEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	// entity.Component.Properties = make(map[string]string)

	entity.Component.Name = resource.GetName()
	// entity.Component.Description = registryClient.Spec.Description
	// entity.Uri = registryClient.Spec.Uri
	// entity.Component.Properties["url"] = registryClient.Spec.Uri
}
