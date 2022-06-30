package connection

import (
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"

	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("connection-method")

// CreateConnection will deploy the NifiDataflow on NiFi Cluster
func CreateConnection(connection *nigoapi.ConnectionEntity, config *clientconfig.NifiConfig) (*v1alpha1.NifiConnectionStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.CreateConnection(*connection)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Create connection"); err != nil {
		return nil, err
	}

	return &v1alpha1.NifiConnectionStatus{ConnectionId: entity.Id}, nil
}
