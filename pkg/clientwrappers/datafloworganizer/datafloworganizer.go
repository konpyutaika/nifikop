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
func ExistDataflowOrganizerGroup(
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
	posX, posY float64,
	groupName string,
	group v1alpha1.OrganizerGroup,
	groupStatus v1alpha1.OrganizerGroupStatus,
	config *clientconfig.NifiConfig) (*v1alpha1.OrganizerGroupStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchTitleEntity := nigoapi.LabelEntity{}
	updateTitleLabelEntity(posX, posY, groupName, group, &scratchTitleEntity)

	titleEntity, err := nClient.CreateLabel(scratchTitleEntity, group.GetParentProcessGroupID(config.RootProcessGroupId))

	if err := clientwrappers.ErrorCreateOperation(log, err, "Create title label"); err != nil {
		return nil, err
	}

	groupStatus.TitleStatus.Id = titleEntity.Id

	scratchContentEntity := nigoapi.LabelEntity{}
	updateContentLabelEntity(posX, posY, groupName, group, &scratchContentEntity)

	contentEntity, err := nClient.CreateLabel(scratchContentEntity, group.GetParentProcessGroupID(config.RootProcessGroupId))

	if err := clientwrappers.ErrorCreateOperation(log, err, "Create content label"); err != nil {
		return nil, err
	}

	groupStatus.ContentStatus.Id = contentEntity.Id

	return &groupStatus, nil
}

func SyncDataflowOrganizerGroup(
	posX, posY float64,
	groupName string,
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

	if !titleLabelIsSync(posX, posY, groupName, group, titleEntity) {
		updateTitleLabelEntity(posX, posY, groupName, group, titleEntity)
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

	if !contentLabelIsSync(posX, posY, groupName, group, contentEntity) {
		updateContentLabelEntity(posX, posY, groupName, group, contentEntity)
		contentEntity, err = nClient.UpdateLabel(*contentEntity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update content label"); err != nil {
			return nil, err
		}
	}
	groupStatus.ContentStatus.Id = contentEntity.Id

	return &groupStatus, nil
}

func RemoveDataflowOrganizerGroup(
	groupStatus v1alpha1.OrganizerGroupStatus,
	config *clientconfig.NifiConfig) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	titleEntity, err := nClient.GetLabel(groupStatus.TitleStatus.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get title label"); err != nil {
		if err != nificlient.ErrNifiClusterReturned404 {
			return err
		}
	} else {
		err = nClient.RemoveLabel(*titleEntity)
		if err := clientwrappers.ErrorRemoveOperation(log, err, "Remove title label"); err != nil {
			return err
		}
	}

	contentEntity, err := nClient.GetLabel(groupStatus.ContentStatus.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get content label"); err != nil {
		if err != nificlient.ErrNifiClusterReturned404 {
			return err
		}
	} else {
		err = nClient.RemoveLabel(*contentEntity)
		if err := clientwrappers.ErrorRemoveOperation(log, err, "Remove content label"); err != nil {
			return err
		}
	}

	return nil
}

func titleLabelIsSync(
	posX, posY float64,
	groupName string,
	group v1alpha1.OrganizerGroup,
	entity *nigoapi.LabelEntity) bool {
	return groupName == entity.Component.Label && group.FontSize == entity.Component.Style["font-size"] &&
		group.Color == entity.Component.Style["background-color"] && posX == entity.Component.Position.X &&
		posY == entity.Component.Position.Y && group.GetTitleWidth(groupName) == entity.Component.Width &&
		group.GetTitleHeight(groupName) == entity.Component.Height
}

func contentLabelIsSync(
	posX, posY float64,
	groupName string,
	group v1alpha1.OrganizerGroup,
	entity *nigoapi.LabelEntity) bool {
	return entity.Component.Label == "" && group.FontSize == entity.Component.Style["font-size"] &&
		group.Color == entity.Component.Style["background-color"] && float64(posX) == entity.Component.Position.X &&
		group.GetTitleHeight(groupName)+float64(posY) == entity.Component.Position.Y && group.GetContentWidth() == entity.Component.Width &&
		group.GetContentHeight() == entity.Component.Height
}

func updateTitleLabelEntity(
	posX, posY float64,
	groupName string,
	group v1alpha1.OrganizerGroup,
	entity *nigoapi.LabelEntity) {
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

	entity.Component.Label = groupName
	entity.Component.Style["background-color"] = group.Color
	entity.Component.Style["font-size"] = group.FontSize
	entity.Component.Width = group.GetTitleWidth(groupName)
	entity.Component.Height = group.GetTitleHeight(groupName)
	entity.Component.Position = &nigoapi.PositionDto{
		X: posX,
		Y: posY,
	}
}

func updateContentLabelEntity(
	posX, posY float64,
	groupName string,
	group v1alpha1.OrganizerGroup,
	entity *nigoapi.LabelEntity) {
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
		X: posX,
		Y: group.GetTitleHeight(groupName) + posY,
	}
}
