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

	if groupStatus.TitleStatus.Id == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetLabel(groupStatus.TitleStatus.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get label"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

// CreateDataflowOrganizer will deploy the Group of NifiDataflowOrganizer on NiFi Cluster
func CreateDataflowOrganizerGroup(
	group v1alpha1.OrganizerGroup,
	groupStatus v1alpha1.OrganizerGroupStatus,
	config *clientconfig.NifiConfig) (*v1alpha1.OrganizerGroupStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchTitleEntity := nigoapi.LabelEntity{}
	updateTitleLabelEntity(group, &scratchTitleEntity)

	titleEntity, err := nClient.CreateLabel(scratchTitleEntity, group.GetParentProcessGroupID(config.RootProcessGroupId))

	if err := clientwrappers.ErrorCreateOperation(log, err, "Create title label"); err != nil {
		return nil, err
	}

	groupStatus.TitleStatus.Id = titleEntity.Id

	scratchContentEntity := nigoapi.LabelEntity{}
	updateContentLabelEntity(group, &scratchContentEntity)

	contentEntity, err := nClient.CreateLabel(scratchContentEntity, group.GetParentProcessGroupID(config.RootProcessGroupId))

	if err := clientwrappers.ErrorCreateOperation(log, err, "Create content label"); err != nil {
		return nil, err
	}

	groupStatus.ContentStatus.Id = contentEntity.Id

	return &groupStatus, nil
}

func SyncDataflowOrganizerGroup(
	group v1alpha1.OrganizerGroup,
	groupStatus v1alpha1.OrganizerGroupStatus,
	config *clientconfig.NifiConfig) (*v1alpha1.OrganizerGroupStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	titleEntity, err := nClient.GetLabel(groupStatus.TitleStatus.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get title label"); err != nil {
		return nil, err
	}

	if !titleLabelIsSync(group, titleEntity) {
		updateTitleLabelEntity(group, titleEntity)
		titleEntity, err = nClient.UpdateLabel(*titleEntity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update title label"); err != nil {
			return nil, err
		}
	}
	groupStatus.TitleStatus.Id = titleEntity.Id

	contentEntity, err := nClient.GetLabel(groupStatus.ContentStatus.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get content label"); err != nil {
		return nil, err
	}

	if !contentLabelIsSync(group, contentEntity) {
		updateContentLabelEntity(group, contentEntity)
		contentEntity, err = nClient.UpdateLabel(*contentEntity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update content label"); err != nil {
			return nil, err
		}
	}
	groupStatus.ContentStatus.Id = contentEntity.Id

	return &groupStatus, nil
}

func titleLabelIsSync(
	group v1alpha1.OrganizerGroup,
	entity *nigoapi.LabelEntity) bool {
	return group.Name == entity.Component.Label && group.FontSize == entity.Component.Style["font-size"] &&
		group.Color == entity.Component.Style["background-color"] && group.GetTitlePosX() == entity.Component.Position.X &&
		group.GetTitlePosY() == entity.Component.Position.Y && group.GetTitleWidth() == entity.Component.Width &&
		group.GetTitleHeight() == entity.Component.Height
}

func contentLabelIsSync(
	group v1alpha1.OrganizerGroup,
	entity *nigoapi.LabelEntity) bool {
	return entity.Component.Label == "" && group.FontSize == entity.Component.Style["font-size"] &&
		group.Color == entity.Component.Style["background-color"] && group.GetContentPosX() == entity.Component.Position.X &&
		group.GetContentPosY() == entity.Component.Position.Y && group.GetContentWidth() == entity.Component.Width &&
		group.GetContentHeight() == entity.Component.Height
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

	if entity.Component.Style == nil {
		entity.Component.Style = make(map[string]string)
	}

	entity.Component.Label = group.Name
	entity.Component.Style["background-color"] = group.Color
	entity.Component.Style["font-size"] = group.FontSize
	entity.Component.Width = group.GetTitleWidth()
	entity.Component.Height = group.GetTitleHeight()
	entity.Component.Position = &nigoapi.PositionDto{
		X: group.GetTitlePosX(),
		Y: group.GetTitlePosY(),
	}
}

func updateContentLabelEntity(group v1alpha1.OrganizerGroup, entity *nigoapi.LabelEntity) {
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

	if entity.Component.Style == nil {
		entity.Component.Style = make(map[string]string)
	}

	entity.Component.Label = ""
	entity.Component.Style["background-color"] = group.Color
	entity.Component.Width = group.GetContentWidth()
	entity.Component.Height = group.GetContentHeight()
	entity.Component.Position = &nigoapi.PositionDto{
		X: group.GetContentPosX(),
		Y: group.GetContentPosY(),
	}
}
