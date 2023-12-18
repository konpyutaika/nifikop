package connection

import (
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("connection-method")

// CreateConnection will deploy the NifiDataflow on NiFi Cluster.
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
		Component: &nigoapi.ConnectionDto{
			Name:          connection.Name,
			ParentGroupId: source.ParentGroupId,
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
			BackPressureDataSizeThreshold: connection.Spec.Configuration.BackPressureDataSizeThreshold,
			BackPressureObjectThreshold:   connection.Spec.Configuration.BackPressureObjectThreshold,
			LoadBalanceStrategy:           string(connection.Spec.Configuration.LoadBalanceStrategy),
			LoadBalancePartitionAttribute: connection.Spec.Configuration.GetLoadBalancePartitionAttribute(),
			LoadBalanceCompression:        string(connection.Spec.Configuration.LoadBalancePartitionAttribute),
			Prioritizers:                  connection.Spec.Configuration.GetStringPrioritizers(),
			LabelIndex:                    connection.Spec.Configuration.GetLabelIndex(),
			Bends:                         bends,
		},
	}

	entity, err := nClient.CreateConnection(connectionEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Create connection"); err != nil {
		return nil, err
	}

	return &v1alpha1.NifiConnectionStatus{ConnectionId: entity.Id}, nil
}

// GetConnectionInformation retrieve the connection information.
func GetConnectionInformation(connection *v1alpha1.NifiConnection, config *clientconfig.NifiConfig) (*nigoapi.ConnectionEntity, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	connectionEntity, err := nClient.GetConnection(connection.Status.ConnectionId)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get connection"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil, nil
		}
		return nil, err
	}

	return connectionEntity, nil
}

// ConnectionExist check if the NifiConnection exist on NiFi Cluster.
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

// SyncConnectionConfig implements the logic to sync a NifiConnection config with the deployed connection config.
func SyncConnectionConfig(connection *v1alpha1.NifiConnection,
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
		return &connection.Status, errorfactory.NifiConnectionDeleting{}
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
		connectionEntity.Component.BackPressureDataSizeThreshold = connection.Spec.Configuration.BackPressureDataSizeThreshold
		connectionEntity.Component.BackPressureObjectThreshold = connection.Spec.Configuration.BackPressureObjectThreshold
		connectionEntity.Component.LoadBalanceStrategy = string(connection.Spec.Configuration.LoadBalanceStrategy)
		connectionEntity.Component.LoadBalancePartitionAttribute = connection.Spec.Configuration.GetLoadBalancePartitionAttribute()
		connectionEntity.Component.LoadBalanceCompression = string(connection.Spec.Configuration.LoadBalanceCompression)
		connectionEntity.Component.Prioritizers = connection.Spec.Configuration.GetStringPrioritizers()
		connectionEntity.Component.LabelIndex = connection.Spec.Configuration.GetLabelIndex()
		connectionEntity.Component.Bends = bends

		_, err := nClient.UpdateConnection(*connectionEntity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update connection"); err != nil {
			return nil, err
		}
		return &connection.Status, errorfactory.NifiConnectionSyncing{}
	}

	if isDestinationChanged(connectionEntity, destination) {
		_, err := SyncConnectionDestination(connection, destination, config)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update connection"); err != nil {
			return nil, err
		}
		return &connection.Status, errorfactory.NifiConnectionSyncing{}
	}

	return &connection.Status, nil
}

// IsOutOfSyncConnection control if the deployed connection is out of sync with the NifiConnection resource.
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

	return isConfigurationChanged(connectionEntity, connection) || isSourceChanged(connectionEntity, source) ||
		isDestinationChanged(connectionEntity, destination), nil
}

// isConfigurationChanged control if the deployed connection configuration is out of sync.
func isConfigurationChanged(connectionEntity *nigoapi.ConnectionEntity, connection *v1alpha1.NifiConnection) bool {
	var bends []nigoapi.PositionDto
	for _, bend := range connection.Spec.Configuration.GetBends() {
		bends = append(bends, nigoapi.PositionDto{
			X: float64(*bend.X),
			Y: float64(*bend.Y),
		})
	}

	return connectionEntity.Component.FlowFileExpiration != connection.Spec.Configuration.GetFlowFileExpiration() ||
		connectionEntity.Component.BackPressureDataSizeThreshold != connection.Spec.Configuration.BackPressureDataSizeThreshold ||
		connectionEntity.Component.BackPressureObjectThreshold != connection.Spec.Configuration.BackPressureObjectThreshold ||
		connectionEntity.Component.LoadBalanceStrategy != string(connection.Spec.Configuration.LoadBalanceStrategy) ||
		connectionEntity.Component.LoadBalancePartitionAttribute != connection.Spec.Configuration.GetLoadBalancePartitionAttribute() ||
		connectionEntity.Component.LoadBalanceCompression != string(connection.Spec.Configuration.LoadBalanceCompression) ||
		!util.StringSliceStrictCompare(connectionEntity.Component.Prioritizers, connection.Spec.Configuration.GetStringPrioritizers()) ||
		connectionEntity.Component.LabelIndex != connection.Spec.Configuration.GetLabelIndex() ||
		isBendChanged(connectionEntity.Component.Bends, bends)
}

// isBendChanged control if the deployed connection bends are out of sync.
func isBendChanged(current []nigoapi.PositionDto, original []nigoapi.PositionDto) bool {
	if len(current) != len(original) {
		return true
	}

	for i, posC := range current {
		if posC.X != original[i].X || posC.Y != original[i].Y {
			return true
		}
	}

	return false
}

// isSourceChanged control if the deployed connection source is out of sync.
func isSourceChanged(
	connectionEntity *nigoapi.ConnectionEntity,
	source *v1alpha1.ComponentInformation) bool {
	return connectionEntity.Component.Source.Id != source.Id || connectionEntity.Component.Source.GroupId != source.GroupId ||
		connectionEntity.Component.Source.Type_ != source.Type
}

// isSourceChanged control if the deployed connection destination is out of sync.
func isDestinationChanged(
	connectionEntity *nigoapi.ConnectionEntity,
	destination *v1alpha1.ComponentInformation) bool {
	return connectionEntity.Component.Destination.Id != destination.Id || connectionEntity.Component.Destination.GroupId != destination.GroupId ||
		connectionEntity.Component.Destination.Type_ != destination.Type
}

// SyncConnectionDestination implements the logic to sync a NifiConnection with the deployed connection destination.
func SyncConnectionDestination(connection *v1alpha1.NifiConnection, destination *v1alpha1.ComponentInformation,
	config *clientconfig.NifiConfig) (*v1alpha1.NifiConnectionStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	connectionEntity, err := nClient.GetConnection(connection.Status.ConnectionId)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get connection"); err != nil {
		return nil, err
	}

	connectionEntity.Component.Destination.Id = destination.Id
	connectionEntity.Component.Destination.Type_ = destination.Type
	connectionEntity.Component.Destination.GroupId = destination.GroupId

	_, err = nClient.UpdateConnection(*connectionEntity)
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Update connection"); err != nil {
		return nil, err
	}
	return &connection.Status, nil
}

// DeleteConnection implements the logic to delete a connection.
func DeleteConnection(connection *v1alpha1.NifiConnection, config *clientconfig.NifiConfig) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil
	}

	connectionEntity, err := nClient.GetConnection(connection.Status.ConnectionId)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get connection"); err != nil {
		return err
	}

	err = nClient.DeleteConnection(*connectionEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Remove process-group"); err != nil {
		return err
	}

	return nil
}

// DropConnectionFlowFiles implements the logic to drop the flowfiles from a connection.
func DropConnectionFlowFiles(connection *v1alpha1.NifiConnection,
	config *clientconfig.NifiConfig) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil
	}

	_, err = nClient.CreateDropRequest(connection.Status.ConnectionId)
	if err := clientwrappers.ErrorUpdateOperation(log, err, "Create drop-request"); err != nil {
		return err
	}

	return nil
}
