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
	for _, nodeState := range c.Status.NodesState {
		if nodeState.ConfigurationState != v1.ConfigInSync || nodeState.GracefulActionState.State != v1.GracefulUpscaleSucceeded ||
			!nodeState.PodIsReady {
			return false
		}
	}
	return c.Status.State.IsReady()
}

func (c *InternalCluster) Id() string {
	return c.Name
}
