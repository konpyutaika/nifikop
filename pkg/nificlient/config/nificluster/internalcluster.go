package nificluster

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/konpyutaika/nifikop/api/v1"
)

type InternalCluster struct {
	Status    v1.NifiClusterStatus
	Name      string
	Namespace string
}

func (cluster *InternalCluster) ClusterLabelString() string {
	return fmt.Sprintf("%s.%s", cluster.Name, cluster.Namespace)
}

func (c *InternalCluster) IsInternal() bool {
	return true
}

func (c InternalCluster) IsExternal() bool {
	return false
}

func (c InternalCluster) IsReady(log zap.Logger) bool {
	var blocked map[string]v1.NodeState
	for id, nodeState := range c.Status.NodesState {
		if nodeState.ConfigurationState != v1.ConfigInSync || nodeState.GracefulActionState.State != v1.GracefulUpscaleSucceeded ||
			!nodeState.PodIsReady {
			if blocked == nil {
				blocked = make(map[string]v1.NodeState)
			}
			blocked[id] = nodeState
		}
	}
	ready := c.Status.State.IsReady() && len(blocked) == 0
	if !ready {
		log.Info("Waiting for NiFi cluster readiness", zap.String("cluster", c.ClusterLabelString()),
			zap.String("state", string(c.Status.State)), zap.Any("unreadyNodes", blocked))
	}
	return ready
}

func (c *InternalCluster) Id() string {
	return c.Name
}
