package connection

import (
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"

	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
)

var log = common.CustomLogger().Named("connection-method")

// CreateConnection will deploy the NifiDataflow on NiFi Cluster
func CreateConnection(connection *v1alpha1.NifiConnection, source *v1alpha1.ComponentInformation, destination *v1alpha1.ComponentInformation,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiConnectionStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	var bends []nigoapi.PositionDto
	for _, bend := range connection.Spec.Configuration.GetBends() {
		bends = append(bends, nigoapi.PositionDto{
			X: float64(*bend.X),
			Y: float64(*bend.Y),
		})
	}

	var defaultVersion int64 = 0
	connectionEntity := nigoapi.ConnectionEntity{
		Revision: &nigoapi.RevisionDto{
			Version: &defaultVersion,
		},
		Id: source.ParentGroupId,
		Component: &nigoapi.ConnectionDto{
			Name: connection.Name,
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
			FlowFileExpiration:            connection.Spec.Configuration.GetFlowFileExpiration(),
			BackPressureDataSizeThreshold: connection.Spec.Configuration.GetBackPressureDataSizeThreshold(),
			BackPressureObjectThreshold:   connection.Spec.Configuration.GetBackPressureObjectThreshold(),
			LoadBalanceStrategy:           string(connection.Spec.Configuration.GetLoadBalanceStrategy()),
			LoadBalancePartitionAttribute: connection.Spec.Configuration.GetLoadBalancePartitionAttribute(),
			LoadBalanceCompression:        string(connection.Spec.Configuration.GetLoadBalanceCompression()),
			Prioritizers:                  connection.Spec.Configuration.GetStringPrioritizers(),
			Bends:                         bends,
		},
	}

	entity, err := nClient.CreateConnection(connectionEntity)
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

// SyncConnection implements the logic to sync a SyncConnection with the deployed connection.
func SyncConnection(connection *v1alpha1.NifiConnection,
	source *v1alpha1.ComponentInformation, destination *v1alpha1.ComponentInformation,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiConnectionStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	connectionEntity, err := nClient.GetConnection(connection.Status.ConnectionId)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get connection"); err != nil {
		return nil, err
	}

	if isSourceChanged(connectionEntity, source) {
		connectionEntity.SourceId = source.Id

		_, err := nClient.UpdateConnection(*connectionEntity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update connection"); err != nil {
			return nil, err
		}
		return &connection.Status, errorfactory.NifiConnectionSyncing{}
	}

	if isConfigurationChanged(connectionEntity, connection) {
		connectionEntity.Component.Name = connection.Name

		var bends []nigoapi.PositionDto
		for _, bend := range connection.Spec.Configuration.GetBends() {
			bends = append(bends, nigoapi.PositionDto{
				X: float64(*bend.X),
				Y: float64(*bend.Y),
			})
		}

		connectionEntity.Component.FlowFileExpiration = connection.Spec.Configuration.GetFlowFileExpiration()
		connectionEntity.Component.BackPressureDataSizeThreshold = connection.Spec.Configuration.GetBackPressureDataSizeThreshold()
		connectionEntity.Component.BackPressureObjectThreshold = connection.Spec.Configuration.GetBackPressureObjectThreshold()
		connectionEntity.Component.LoadBalanceStrategy = string(connection.Spec.Configuration.GetLoadBalanceStrategy())
		connectionEntity.Component.LoadBalancePartitionAttribute = connection.Spec.Configuration.GetLoadBalancePartitionAttribute()
		connectionEntity.Component.LoadBalanceCompression = string(connection.Spec.Configuration.GetLoadBalanceCompression())
		connectionEntity.Component.Prioritizers = connection.Spec.Configuration.GetStringPrioritizers()
		connectionEntity.Component.Bends = bends

		_, err := nClient.UpdateConnection(*connectionEntity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update connection"); err != nil {
			return nil, err
		}
		return &connection.Status, errorfactory.NifiConnectionSyncing{}
	}

	return &connection.Status, nil
}

// IsOutOfSyncConnection control if the deployed dataflow is out of sync with the NifiDataflow resource
func IsOutOfSyncConnection(connection *v1alpha1.NifiConnection,
	source *v1alpha1.ComponentInformation, destination *v1alpha1.ComponentInformation,
	config *clientconfig.NifiConfig) (bool, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	connectionEntity, err := nClient.GetConnection(connection.Status.ConnectionId)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get connection"); err != nil {
		return false, err
	}

	return isConfigurationChanged(connectionEntity, connection) || isSourceChanged(connectionEntity, source), nil
}

func isConfigurationChanged(connectionEntity *nigoapi.ConnectionEntity, connection *v1alpha1.NifiConnection) bool {
	var bends []nigoapi.PositionDto
	for _, bend := range connection.Spec.Configuration.GetBends() {
		bends = append(bends, nigoapi.PositionDto{
			X: float64(*bend.X),
			Y: float64(*bend.Y),
		})
	}

	return connectionEntity.Component.FlowFileExpiration != connection.Spec.Configuration.GetFlowFileExpiration() ||
		connectionEntity.Component.BackPressureDataSizeThreshold != connection.Spec.Configuration.GetBackPressureDataSizeThreshold() ||
		connectionEntity.Component.BackPressureObjectThreshold != connection.Spec.Configuration.GetBackPressureObjectThreshold() ||
		connectionEntity.Component.LoadBalanceStrategy != string(connection.Spec.Configuration.GetLoadBalanceStrategy()) ||
		connectionEntity.Component.LoadBalancePartitionAttribute != connection.Spec.Configuration.GetLoadBalancePartitionAttribute() ||
		connectionEntity.Component.LoadBalanceCompression != string(connection.Spec.Configuration.GetLoadBalanceCompression()) ||
		!util.StringSliceStrictCompare(connectionEntity.Component.Prioritizers, connection.Spec.Configuration.GetStringPrioritizers()) ||
		isBendChanged(connectionEntity.Component.Bends, bends)
}

func isBendChanged(current []nigoapi.PositionDto, original []nigoapi.PositionDto) bool {
	if len(current) != len(original) {
		return true
	}

	for _, posC := range current {
		var found bool = false
		for _, posO := range original {
			if posC.X == posO.X && posC.Y == posO.Y {
				found = true
			}
		}

		if found {
			return true
		}
	}

	return false
}

func isSourceChanged(
	connectionEntity *nigoapi.ConnectionEntity,
	source *v1alpha1.ComponentInformation) bool {

	return connectionEntity.Component.Source.Id != source.Id || connectionEntity.Component.Source.GroupId != source.GroupId ||
		connectionEntity.Component.Source.Type_ != source.Type
}
