package nificluster

import (
	"go.uber.org/zap"

	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

type ExternalCluster struct {
	NodeURITemplate    string
	NodeIds            []int32
	NifiURI            string
	Name               string
	RootProcessGroupId string

	NifiConfig *clientconfig.NifiConfig
}

func (cluster *ExternalCluster) IsExternal() bool {
	return true
}

func (cluster *ExternalCluster) IsInternal() bool {
	return false
}

func (cluster *ExternalCluster) ClusterLabelString() string {
	return cluster.Name
}

func (cluster ExternalCluster) IsReady(log zap.Logger) bool {
	nClient, err := common.NewClusterConnection(&log, cluster.NifiConfig)
	if err != nil {
		return false
	}

	clusterEntity, err := nClient.DescribeCluster()
	if err != nil {
		return false
	}

	for _, node := range clusterEntity.Cluster.Nodes {
		if node.Status != nificlient.CONNECTED_STATUS {
			return false
		}
	}
	return true
}

func (cluster *ExternalCluster) Id() string {
	return cluster.Name
}
