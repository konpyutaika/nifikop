// Copyright 2020 Orange SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package apis

package k8sutil

import (
	"context"
	"fmt"
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// IsAlreadyOwnedError checks if a controller already own the instance
func IsAlreadyOwnedError(err error) bool {
	return errors.Is(err, &controllerutil.AlreadyOwnedError{})
}

// IsMarkedForDeletion determines if the object is marked for deletion
func IsMarkedForDeletion(m metav1.ObjectMeta) bool {
	return m.GetDeletionTimestamp() != nil
}

// UpdateNodeStatus updates the node status with rack and configuration infos
func UpdateNodeStatus(c client.Client, nodeIds []string, cluster *v1alpha1.NifiCluster, state interface{}, logger logr.Logger) error {
	typeMeta := cluster.TypeMeta

	for _, nodeId := range nodeIds {

		if cluster.Status.NodesState == nil {
			switch s := state.(type) {
			case v1alpha1.GracefulActionState:
				cluster.Status.NodesState = map[string]v1alpha1.NodeState{nodeId: {GracefulActionState: s}}
			case v1alpha1.ConfigurationState:
				cluster.Status.NodesState = map[string]v1alpha1.NodeState{nodeId: {ConfigurationState: s}}
			case v1alpha1.InitClusterNode:
				cluster.Status.NodesState = map[string]v1alpha1.NodeState{nodeId: {InitClusterNode: s}}
			case bool:
				cluster.Status.NodesState = map[string]v1alpha1.NodeState{nodeId: {PodIsReady: s}}
			}
		} else if val, ok := cluster.Status.NodesState[nodeId]; ok {
			switch s := state.(type) {
			case v1alpha1.GracefulActionState:
				val.GracefulActionState = s
			case v1alpha1.ConfigurationState:
				val.ConfigurationState = s
			case v1alpha1.InitClusterNode:
				val.InitClusterNode = s
			case bool:
				val.PodIsReady = s
			}
			cluster.Status.NodesState[nodeId] = val
		} else {
			switch s := state.(type) {
			case v1alpha1.GracefulActionState:
				cluster.Status.NodesState[nodeId] = v1alpha1.NodeState{GracefulActionState: s}
			case v1alpha1.ConfigurationState:
				cluster.Status.NodesState[nodeId] = v1alpha1.NodeState{ConfigurationState: s}
			case v1alpha1.InitClusterNode:
				cluster.Status.NodesState[nodeId] = v1alpha1.NodeState{InitClusterNode: s}
			case bool:
				cluster.Status.NodesState[nodeId] = v1alpha1.NodeState{PodIsReady: s}
			}
		}
	}

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
				case v1alpha1.GracefulActionState:
					cluster.Status.NodesState = map[string]v1alpha1.NodeState{nodeId: {GracefulActionState: s}}
				case v1alpha1.ConfigurationState:
					cluster.Status.NodesState = map[string]v1alpha1.NodeState{nodeId: {ConfigurationState: s}}
				case v1alpha1.InitClusterNode:
					cluster.Status.NodesState = map[string]v1alpha1.NodeState{nodeId: {InitClusterNode: s}}
				case bool:
					cluster.Status.NodesState = map[string]v1alpha1.NodeState{nodeId: {PodIsReady: s}}
				}
			} else if val, ok := cluster.Status.NodesState[nodeId]; ok {
				switch s := state.(type) {
				case v1alpha1.GracefulActionState:
					val.GracefulActionState = s
				case v1alpha1.ConfigurationState:
					val.ConfigurationState = s
				case v1alpha1.InitClusterNode:
					val.InitClusterNode = s
				case bool:
					val.PodIsReady = s
				}
				cluster.Status.NodesState[nodeId] = val
			} else {
				switch s := state.(type) {
				case v1alpha1.GracefulActionState:
					cluster.Status.NodesState[nodeId] = v1alpha1.NodeState{GracefulActionState: s}
				case v1alpha1.ConfigurationState:
					cluster.Status.NodesState[nodeId] = v1alpha1.NodeState{ConfigurationState: s}
				case v1alpha1.InitClusterNode:
					cluster.Status.NodesState[nodeId] = v1alpha1.NodeState{InitClusterNode: s}
				case bool:
					cluster.Status.NodesState[nodeId] = v1alpha1.NodeState{PodIsReady: s}
				}
			}
		}

		err = updateClusterStatus(c, cluster)
		if err != nil {
			return errors.WrapIff(err, "could not update Nifi clusters node(s) %s state", strings.Join(nodeIds, ","))
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cluster.TypeMeta = typeMeta
	logger.Info("Nifi cluster state updated")
	return nil
}

// DeleteStatus deletes the given node state from the CR
func DeleteStatus(c client.Client, nodeId string, cluster *v1alpha1.NifiCluster, logger logr.Logger) error {
	typeMeta := cluster.TypeMeta

	nodeStatus := cluster.Status.NodesState

	delete(nodeStatus, nodeId)

	cluster.Status.NodesState = nodeStatus

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
		err = updateClusterStatus(c, cluster)
		if err != nil {
			return errors.WrapIff(err, "could not delete Nifi clusters node %s state ", nodeId)
		}
	}

	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cluster.TypeMeta = typeMeta
	logger.Info(fmt.Sprintf("Nifi node %s state deleted", nodeId))
	return nil
}

// UpdateCRStatus updates the cluster state
func UpdateCRStatus(c client.Client, cluster *v1alpha1.NifiCluster, state interface{}, logger logr.Logger) error {
	typeMeta := cluster.TypeMeta

	switch s := state.(type) {
	case v1alpha1.ClusterState:
		cluster.Status.State = s
	}

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
		case v1alpha1.ClusterState:
			cluster.Status.State = s
		}

		err = updateClusterStatus(c, cluster)
		if err != nil {
			return errors.WrapIf(err, "could not update CR state")
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cluster.TypeMeta = typeMeta
	logger.Info("CR status updated", "status", state)
	return nil
}

// UpdateRootProcessGroupIdStatus updates the cluster root process group id
func UpdateRootProcessGroupIdStatus(c client.Client, cluster *v1alpha1.NifiCluster, id string, logger logr.Logger) error {
	typeMeta := cluster.TypeMeta

	cluster.Status.RootProcessGroupId = id

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

		err = updateClusterStatus(c, cluster)
		if err != nil {
			return errors.WrapIf(err, "could not update CR state")
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cluster.TypeMeta = typeMeta
	logger.Info("Root process grout id updated", "id", id)
	return nil
}

// UpdateRollingUpgradeState updates the state of the cluster with rolling upgrade info
func UpdateRollingUpgradeState(c client.Client, cluster *v1alpha1.NifiCluster, time time.Time, logger logr.Logger) error {
	typeMeta := cluster.TypeMeta

	timeStamp := time.Format("Mon, 2 Jan 2006 15:04:05 GMT")
	cluster.Status.RollingUpgrade.LastSuccess = timeStamp

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

		err = c.Status().Update(context.Background(), cluster)
		if apierrors.IsNotFound(err) {
			err = c.Update(context.Background(), cluster)
		}
		if err != nil {
			return errors.WrapIf(err, "could not update rolling upgrade state")
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	cluster.TypeMeta = typeMeta
	logger.Info("Rolling upgrade status updated", "status", timeStamp)
	return nil
}

func updateClusterStatus(c client.Client, cluster *v1alpha1.NifiCluster) error {
	err := c.Status().Update(context.Background(), cluster)
	if apierrors.IsNotFound(err) {
		return c.Update(context.Background(), cluster)
	}
	return nil
}
