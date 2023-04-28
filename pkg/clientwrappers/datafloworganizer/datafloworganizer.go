package datafloworganizer

import (
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
)

var log = common.CustomLogger().Named("datafloworganizer-method")

// ExistDataflowOrganizer check if the NifiDataflowOrganizer exist on NiFi Cluster
func ExistDataflowOrganizer(dataflowOrganizer *v1alpha1.NifiDataflowOrganizer,
	config *clientconfig.NifiConfig) (bool, error) {

	if dataflowOrganizer.Status.TitleLabelStatus.Id == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetLabel(dataflowOrganizer.Status.TitleLabelStatus.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get label"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

// CreateDataflowOrganizer will deploy the NifiDataflowOrganizer on NiFi Cluster
func CreateDataflowOrganizer(dataflowOrganizer *v1alpha1.NifiDataflowOrganizer,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiDataflowOrganizerStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.LabelEntity{}
	updateLabelEntity(dataflowOrganizer, &scratchEntity)

	entity, err := nClient.CreateLabel(scratchEntity, dataflowOrganizer.Spec.GetParentProcessGroupID(config.RootProcessGroupId))

	if err := clientwrappers.ErrorCreateOperation(log, err, "Create process-group"); err != nil {
		return nil, err
	}

	dataflowOrganizer.Status.TitleLabelStatus.Id = entity.Id
	return &dataflowOrganizer.Status, nil
}

func updateLabelEntity(dataflowOrganizer *v1alpha1.NifiDataflowOrganizer, entity *nigoapi.LabelEntity) {

	var defaultVersion int64 = 0

	if entity == nil {
		entity = &nigoapi.LabelEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.LabelDto{}
	}

	entity.Component.Label = dataflowOrganizer.Name
	entity.Component.Style["background-color"] = dataflowOrganizer.Spec.Color
}
