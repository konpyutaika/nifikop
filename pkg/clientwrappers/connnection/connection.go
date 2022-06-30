package dataflow

import (
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("connection-method")

// CreateConnection will deploy the NifiDataflow on NiFi Cluster
func CreateDataflow(flow *v1alpha1.NifiConnection, config *clientconfig.NifiConfig) (*v1alpha1.NifiDataflowStatus, error) {

	// nClient, err := common.NewClusterConnection(log, config)
	// if err != nil {
	// 	return nil, err
	// }

	// nClient.CreateC
	// scratchEntity := nigoapi.ProcessGroupEntity{}
	// updateProcessGroupEntity(flow, registry, config, &scratchEntity)

	// entity, err := nClient.CreateProcessGroup(scratchEntity, flow.Spec.GetParentProcessGroupID(config.RootProcessGroupId))

	// if err := clientwrappers.ErrorCreateOperation(log, err, "Create process-group"); err != nil {
	// 	return nil, err
	// }

	// flow.Status.ProcessGroupID = entity.Id
	// return &flow.Status, nil
}
