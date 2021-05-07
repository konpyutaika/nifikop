package controllersettings

import (
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers"
	"github.com/Orange-OpenSource/nifikop/pkg/common"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log = ctrl.Log.WithName("controllersettings-method")

func controllerConfigIsSync(cluster *v1alpha1.NifiCluster, entity *nigoapi.ControllerConfigurationEntity) bool {
	return cluster.Spec.ReadOnlyConfig.GetMaximumTimerDrivenThreadCount() == entity.Component.MaxTimerDrivenThreadCount
}

func SyncConfiguration(client client.Client, cluster *v1alpha1.NifiCluster) error {

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return err
	}

	entity, err := nClient.GetControllerConfig()
	if err := clientwrappers.ErrorGetOperation(log, err, "Get controller config"); err != nil {
		return err
	}

	if !controllerConfigIsSync(cluster, entity) {
		updateControllerConfigEntity(cluster, entity)
		entity, err = nClient.UpdateControllerConfig(*entity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update controller conif"); err != nil {
			return err
		}
	}
	return nil
}

func updateControllerConfigEntity(cluster *v1alpha1.NifiCluster, entity *nigoapi.ControllerConfigurationEntity) {
	if entity == nil {
		entity = &nigoapi.ControllerConfigurationEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.ControllerConfigurationDto{}
	}
	entity.Component.MaxTimerDrivenThreadCount = cluster.Spec.ReadOnlyConfig.GetMaximumTimerDrivenThreadCount()
}
