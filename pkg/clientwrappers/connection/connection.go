package connection

import (
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"

	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("connection-method")

// CreateConnection will deploy the NifiDataflow on NiFi Cluster
func CreateConnection(source *v1alpha1.ComponentInformation, destination *v1alpha1.ComponentInformation,
	configuration *v1alpha1.ConnectionConfiguration, name string, config *clientconfig.NifiConfig) (*v1alpha1.NifiConnectionStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	var bends []nigoapi.PositionDto
	for _, bend := range configuration.GetBends() {
		bends = append(bends, nigoapi.PositionDto{
			X: float64(*bend.X),
			Y: float64(*bend.Y),
		})
	}

	var defaultVersion int64 = 0
	connection := &nigoapi.ConnectionEntity{
		Revision: &nigoapi.RevisionDto{
			Version: &defaultVersion,
		},
		Id: source.ParentGroupId,
		Component: &nigoapi.ConnectionDto{
			Name: name,
			Source: &nigoapi.ConnectableDto{
				Id:      source.Id,
				Type_:   source.Type,
				GroupId: source.GroupId,
			},
			Destination: &nigoapi.ConnectableDto{
				Id:      destination.Id,
				Type_:   destination.Type,
				GroupId: destination.GroupId,
			},
			FlowFileExpiration:            configuration.GetFlowFileExpiration(),
			BackPressureDataSizeThreshold: configuration.GetBackPressureDataSizeThreshold(),
			BackPressureObjectThreshold:   configuration.GetBackPressureObjectThreshold(),
			LoadBalanceStrategy:           string(configuration.GetLoadBalanceStrategy()),
			LoadBalancePartitionAttribute: configuration.GetLoadBalancePartitionAttribute(),
			LoadBalanceCompression:        string(configuration.GetLoadBalanceCompression()),
			Prioritizers:                  configuration.GetStringPrioritizers(),
			Bends:                         bends,
		},
	}

	entity, err := nClient.CreateConnection(*connection)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Create connection"); err != nil {
		return nil, err
	}

	return &v1alpha1.NifiConnectionStatus{ConnectionId: entity.Id}, nil
}

// ConnectionExist check if the NifiConnection exist on NiFi Cluster
func ConnectionExist(connection *v1alpha1.NifiConnection, config *clientconfig.NifiConfig) (bool, error) {

	if connection.Status.ConnectionId == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	connectionEntity, err := nClient.GetConnection(connection.Status.ConnectionId)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get connection"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return connectionEntity != nil, nil
}
