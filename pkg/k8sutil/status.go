package k8sutil

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

// IsAlreadyOwnedError checks if a controller already own the instance.
func IsAlreadyOwnedError(err error) bool {
	errString := err.Error()
	// check if "Object */* is already owned by another * controller" is in the error message
	return strings.Contains(errString, "Object") && strings.Contains(errString, "is already owned by another") && strings.Contains(errString, "controller")
}

// IsMarkedForDeletion determines if the object is marked for deletion.
func IsMarkedForDeletion(m metav1.ObjectMeta) bool {
	return m.GetDeletionTimestamp() != nil
}

// UpdateNodeStatus updates the node status with rack and configuration infos.
func UpdateNodeStatus(c client.Client, nodeIds []string, cluster *v1.NifiCluster, currentStatus v1.NifiClusterStatus, state interface{}, logger zap.Logger) error {
	typeMeta := cluster.TypeMeta

	for _, nodeId := range nodeIds {
		if cluster.Status.NodesState == nil {
			switch s := state.(type) {
			case v1.GracefulActionState:
				cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {GracefulActionState: s}}
			case v1.ConfigurationState:
				cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {ConfigurationState: s}}
			case v1.InitClusterNode:
				cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {InitClusterNode: s}}
			case bool:
				cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {PodIsReady: s}}
			case metav1.Time:
				if cluster.Status.NodesState[nodeId].CreationTime == nil {
					cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {CreationTime: &s}}
				} else {
					cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {LastUpdatedTime: s}}
				}
			}
		} else if val, ok := cluster.Status.NodesState[nodeId]; ok {
			switch s := state.(type) {
			case v1.GracefulActionState:
				val.GracefulActionState = s
			case v1.ConfigurationState:
				val.ConfigurationState = s
			case v1.InitClusterNode:
				val.InitClusterNode = s
			case bool:
				val.PodIsReady = s
			case metav1.Time:
				if cluster.Status.NodesState[nodeId].CreationTime == nil {
					val.CreationTime = &s
				} else {
					val.LastUpdatedTime = s
				}
			}
			cluster.Status.NodesState[nodeId] = val
		} else {
			switch s := state.(type) {
			case v1.GracefulActionState:
				cluster.Status.NodesState[nodeId] = v1.NodeState{GracefulActionState: s}
			case v1.ConfigurationState:
				cluster.Status.NodesState[nodeId] = v1.NodeState{ConfigurationState: s}
			case v1.InitClusterNode:
				cluster.Status.NodesState[nodeId] = v1.NodeState{InitClusterNode: s}
			case bool:
				cluster.Status.NodesState[nodeId] = v1.NodeState{PodIsReady: s}
			case metav1.Time:
				if cluster.Status.NodesState[nodeId].CreationTime == nil {
					cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {CreationTime: &s}}
				} else {
					cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {LastUpdatedTime: s}}
				}
			}
		}
	}

	if !reflect.DeepEqual(cluster.Status, currentStatus) {
		err := c.Status().Update(context.Background(), cluster)
		if apierrors.IsNotFound(err) {
			err = c.Update(context.Background(), cluster)
		}
		if err != nil {
			if !apierrors.IsConflict(err) {
				return errors.WrapIff(err, "could not update Nifi node(s) %s state", strings.Join(nodeIds, ","))
			}
			err := c.Get(context.TODO(), types.NamespacedName{
				Namespace: cluster.Namespace,
				Name:      cluster.Name,
			}, cluster)
			if err != nil {
				return errors.WrapIf(err, "could not get config for updating status")
			}

			for _, nodeId := range nodeIds {
				if cluster.Status.NodesState == nil {
					switch s := state.(type) {
					case v1.GracefulActionState:
						cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {GracefulActionState: s}}
					case v1.ConfigurationState:
						cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {ConfigurationState: s}}
					case v1.InitClusterNode:
						cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {InitClusterNode: s}}
					case bool:
						cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {PodIsReady: s}}
					case metav1.Time:
						if cluster.Status.NodesState[nodeId].CreationTime == nil {
							cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {CreationTime: &s}}
						} else {
							cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {LastUpdatedTime: s}}
						}
					}
				} else if val, ok := cluster.Status.NodesState[nodeId]; ok {
					switch s := state.(type) {
					case v1.GracefulActionState:
						val.GracefulActionState = s
					case v1.ConfigurationState:
						val.ConfigurationState = s
					case v1.InitClusterNode:
						val.InitClusterNode = s
					case bool:
						val.PodIsReady = s
					case metav1.Time:
						if cluster.Status.NodesState[nodeId].CreationTime == nil {
							val.CreationTime = &s
						} else {
							val.LastUpdatedTime = s
						}
					}
					cluster.Status.NodesState[nodeId] = val
				} else {
					switch s := state.(type) {
					case v1.GracefulActionState:
						cluster.Status.NodesState[nodeId] = v1.NodeState{GracefulActionState: s}
					case v1.ConfigurationState:
						cluster.Status.NodesState[nodeId] = v1.NodeState{ConfigurationState: s}
					case v1.InitClusterNode:
						cluster.Status.NodesState[nodeId] = v1.NodeState{InitClusterNode: s}
					case bool:
						cluster.Status.NodesState[nodeId] = v1.NodeState{PodIsReady: s}
					case metav1.Time:
						if cluster.Status.NodesState[nodeId].CreationTime == nil {
							cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {CreationTime: &s}}
						} else {
							cluster.Status.NodesState = map[string]v1.NodeState{nodeId: {LastUpdatedTime: s}}
						}
					}
				}
			}

			err = updateClusterStatus(c, cluster, currentStatus)
			if err != nil {
				return errors.WrapIff(err, "could not update Nifi clusters node(s) %s state", strings.Join(nodeIds, ","))
			}
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cluster.TypeMeta = typeMeta
	logger.Debug("Nifi cluster state updated",
		zap.String("clusterName", cluster.Name),
		zap.Strings("nodeIds", nodeIds))
	return nil
}

// DeleteStatus deletes the given node state from the CR.
func DeleteStatus(c client.Client, nodeId string, cluster *v1.NifiCluster, currentStatus v1.NifiClusterStatus, logger zap.Logger) error {
	typeMeta := cluster.TypeMeta

	nodeStatus := cluster.Status.NodesState

	delete(nodeStatus, nodeId)

	cluster.Status.NodesState = nodeStatus

	if !reflect.DeepEqual(cluster.Status, currentStatus) {
		err := c.Status().Update(context.Background(), cluster)
		if apierrors.IsNotFound(err) {
			err = c.Update(context.Background(), cluster)
		}
		if err != nil {
			if !apierrors.IsConflict(err) {
				return errors.WrapIff(err, "could not delete Nifi cluster node %s state ", nodeId)
			}
			err := c.Get(context.TODO(), types.NamespacedName{
				Namespace: cluster.Namespace,
				Name:      cluster.Name,
			}, cluster)
			if err != nil {
				return errors.WrapIf(err, "could not get config for updating status")
			}
			nodeStatus = cluster.Status.NodesState

			delete(nodeStatus, nodeId)

			cluster.Status.NodesState = nodeStatus
			err = updateClusterStatus(c, cluster, currentStatus)
			if err != nil {
				return errors.WrapIff(err, "could not delete Nifi clusters node %s state ", nodeId)
			}
		}
	}

	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cluster.TypeMeta = typeMeta
	logger.Info(fmt.Sprintf("Nifi node %s state deleted", nodeId))
	return nil
}

// UpdateCRStatus updates the cluster state.
func UpdateCRStatus(c client.Client, cluster *v1.NifiCluster, currentStatus v1.NifiClusterStatus, state interface{}, logger zap.Logger) error {
	typeMeta := cluster.TypeMeta

	switch s := state.(type) {
	case v1.ClusterState:
		cluster.Status.State = s
	}

	if !reflect.DeepEqual(cluster.Status, currentStatus) {
		err := c.Status().Update(context.Background(), cluster)
		if apierrors.IsNotFound(err) {
			err = c.Update(context.Background(), cluster)
		}
		if err != nil {
			if !apierrors.IsConflict(err) {
				return errors.WrapIf(err, "could not update CR state")
			}
			err := c.Get(context.TODO(), types.NamespacedName{
				Namespace: cluster.Namespace,
				Name:      cluster.Name,
			}, cluster)
			if err != nil {
				return errors.WrapIf(err, "could not get config for updating status")
			}
			switch s := state.(type) {
			case v1.ClusterState:
				cluster.Status.State = s
			}

			err = updateClusterStatus(c, cluster, currentStatus)
			if err != nil {
				return errors.WrapIf(err, "could not update CR state")
			}
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cluster.TypeMeta = typeMeta
	return nil
}

// UpdateRootProcessGroupIdStatus updates the cluster root process group id.
func UpdateRootProcessGroupIdStatus(c client.Client, cluster *v1.NifiCluster, currentStatus v1.NifiClusterStatus, id string, logger zap.Logger) error {
	typeMeta := cluster.TypeMeta

	cluster.Status.RootProcessGroupId = id

	if !reflect.DeepEqual(cluster.Status, currentStatus) {
		err := c.Status().Update(context.Background(), cluster)
		if apierrors.IsNotFound(err) {
			err = c.Update(context.Background(), cluster)
		}
		if err != nil {
			if !apierrors.IsConflict(err) {
				return errors.WrapIf(err, "could not update CR state")
			}
			err := c.Get(context.TODO(), types.NamespacedName{
				Namespace: cluster.Namespace,
				Name:      cluster.Name,
			}, cluster)
			if err != nil {
				return errors.WrapIf(err, "could not get config for updating status")
			}
			cluster.Status.RootProcessGroupId = id

			err = updateClusterStatus(c, cluster, currentStatus)
			if err != nil {
				return errors.WrapIf(err, "could not update CR state")
			}
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cluster.TypeMeta = typeMeta
	logger.Debug("Root process group id updated",
		zap.String("clusterName", cluster.Name),
		zap.String("id", id))
	return nil
}

// UpdateRollingUpgradeState updates the state of the cluster with rolling upgrade info.
func UpdateRollingUpgradeState(c client.Client, cluster *v1.NifiCluster, currentStatus v1.NifiClusterStatus, time time.Time, logger zap.Logger) error {
	typeMeta := cluster.TypeMeta

	timeStamp := time.Format("Mon, 2 Jan 2006 15:04:05 GMT")
	cluster.Status.RollingUpgrade.LastSuccess = timeStamp

	if !reflect.DeepEqual(cluster.Status, currentStatus) {
		err := c.Status().Update(context.Background(), cluster)
		if apierrors.IsNotFound(err) {
			err = c.Update(context.Background(), cluster)
		}
		if err != nil {
			if !apierrors.IsConflict(err) {
				return errors.WrapIf(err, "could not update rolling upgrade state")
			}
			err := c.Get(context.TODO(), types.NamespacedName{
				Namespace: cluster.Namespace,
				Name:      cluster.Name,
			}, cluster)
			if err != nil {
				return errors.WrapIf(err, "could not get config for updating status")
			}

			cluster.Status.RollingUpgrade.LastSuccess = timeStamp

			if !reflect.DeepEqual(cluster.Status, currentStatus) {
				err = c.Status().Update(context.Background(), cluster)
				if apierrors.IsNotFound(err) {
					err = c.Update(context.Background(), cluster)
				}
				if err != nil {
					return errors.WrapIf(err, "could not update rolling upgrade state")
				}
			}
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cluster.TypeMeta = typeMeta
	logger.Info("Rolling upgrade status updated", zap.String("status", timeStamp))
	return nil
}

func updateClusterStatus(c client.Client, cluster *v1.NifiCluster, currentStatus v1.NifiClusterStatus) error {
	if !reflect.DeepEqual(cluster.Status, currentStatus) {
		err := c.Status().Update(context.Background(), cluster)
		if apierrors.IsNotFound(err) {
			return c.Update(context.Background(), cluster)
		}
	}
	return nil
}
