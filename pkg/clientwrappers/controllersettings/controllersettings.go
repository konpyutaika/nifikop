package controllersettings

import (
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("controllersettings-method")

func controllerConfigIsSync(cluster *v1.NifiCluster, entity *nigoapi.ControllerConfigurationEntity) bool {
	return cluster.Spec.ReadOnlyConfig.GetMaximumTimerDrivenThreadCount() == entity.Component.MaxTimerDrivenThreadCount &&
		cluster.Spec.ReadOnlyConfig.GetMaximumEventDrivenThreadCount() == entity.Component.MaxEventDrivenThreadCount
}

func SyncConfiguration(config *clientconfig.NifiConfig, cluster *v1.NifiCluster) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	entity, err := nClient.GetControllerConfig()
	if err := clientwrappers.ErrorGetOperation(log, err, "Get controller config"); err != nil {
		return err
	}

	if !controllerConfigIsSync(cluster, entity) {
		updateControllerConfigEntity(cluster, entity)
		_, _ = nClient.UpdateControllerConfig(*entity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update controller conif"); err != nil {
			return err
		}
	}
	return nil
}

func updateControllerConfigEntity(cluster *v1.NifiCluster, entity *nigoapi.ControllerConfigurationEntity) {
	if entity == nil {
		entity = &nigoapi.ControllerConfigurationEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.ControllerConfigurationDto{}
	}
	entity.Component.MaxTimerDrivenThreadCount = cluster.Spec.ReadOnlyConfig.GetMaximumTimerDrivenThreadCount()
	entity.Component.MaxEventDrivenThreadCount = cluster.Spec.ReadOnlyConfig.GetMaximumEventDrivenThreadCount()
}
