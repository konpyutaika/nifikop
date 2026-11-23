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

	generation := cluster.Generation
	previous := make(map[string]v1.NodeState, len(nodeIds))
	for _, id := range nodeIds {
		if node, ok := cluster.Status.NodesState[id]; ok {
			previous[id] = node
		}
	}

	setNodeStatus(cluster, nodeIds, state)

	if !reflect.DeepEqual(cluster.Status, currentStatus) {
		err := c.Status().Update(context.Background(), cluster)
		if apierrors.IsNotFound(err) {
			err = c.Update(context.Background(), cluster)
		}
		if err != nil {
			if !apierrors.IsConflict(err) {
				return errors.WrapIff(err, "could not update Nifi node(s) %s state", strings.Join(nodeIds, ","))
			}
			conflict := err
			err := c.Get(context.TODO(), types.NamespacedName{
				Namespace: cluster.Namespace,
				Name:      cluster.Name,
			}, cluster)
			if err != nil {
				return errors.WrapIf(err, "could not get config for updating status")
			}

			// Re-evaluate the transition if another reconciler changed membership or advanced the task.
			if generation != cluster.Generation {
				return conflict
			}
			for _, id := range nodeIds {
				before, existed := previous[id]
				latest, exists := cluster.Status.NodesState[id]
				if existed != exists || before.GracefulActionState != latest.GracefulActionState {
					return conflict
				}
			}

			currentStatus = *cluster.Status.DeepCopy()
			setNodeStatus(cluster, nodeIds, state)
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

func setNodeStatus(cluster *v1.NifiCluster, ids []string, value interface{}) {
	if cluster.Status.NodesState == nil {
		cluster.Status.NodesState = make(map[string]v1.NodeState)
	}
	for _, id := range ids {
		node := cluster.Status.NodesState[id]
		switch state := value.(type) {
		case v1.GracefulActionState:
			node.GracefulActionState = state
		case v1.ConfigurationState:
			node.ConfigurationState = state
		case v1.InitClusterNode:
			node.InitClusterNode = state
		case bool:
			node.PodIsReady = state
		case metav1.Time:
			if node.CreationTime == nil {
				node.CreationTime = &state
			} else {
				node.LastUpdatedTime = state
			}
		}
		cluster.Status.NodesState[id] = node
	}
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
			// A conflicting update may have reintroduced the node. Reconcile the latest snapshot.
			return errors.WrapIff(err, "could not delete Nifi cluster node %s state", nodeId)
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
		return err
	}
	return nil
}
