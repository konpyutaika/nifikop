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
func ExistDataflowOrganizer(
	group v1alpha1.OrganizerGroup,
	groupStatus v1alpha1.OrganizerGroupStatus,
	config *clientconfig.NifiConfig) (bool, error) {

	if groupStatus.TitleLabelStatus.Id == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetLabel(groupStatus.TitleLabelStatus.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get label"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

// CreateDataflowOrganizer will deploy the NifiDataflowOrganizer on NiFi Cluster
func CreateDataflowOrganizer(
	group v1alpha1.OrganizerGroup,
	groupStatus v1alpha1.OrganizerGroupStatus,
	config *clientconfig.NifiConfig) (*v1alpha1.OrganizerGroupStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.LabelEntity{}
	updateTitleLabelEntity(group, &scratchEntity)

	entity, err := nClient.CreateLabel(scratchEntity, group.GetParentProcessGroupID(config.RootProcessGroupId))

	if err := clientwrappers.ErrorCreateOperation(log, err, "Create process-group"); err != nil {
		return nil, err
	}

	groupStatus.TitleLabelStatus.Id = entity.Id
	return &groupStatus, nil
}

func updateTitleLabelEntity(group v1alpha1.OrganizerGroup, entity *nigoapi.LabelEntity) {
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

	entity.Component.Label = group.Name
	if entity.Component.Style == nil {
		entity.Component.Style = make(map[string]string)
	}

	entity.Component.Style["background-color"] = group.Color
	entity.Component.Style["font-size"] = group.FontSize
	entity.Component.Width = group.GetTitleWidth()
	entity.Component.Height = group.GetTitleHeight()
}
