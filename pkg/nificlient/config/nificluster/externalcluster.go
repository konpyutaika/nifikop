package nificluster

import (
	"fmt"
	"github.com/Orange-OpenSource/nifikop/pkg/common"
	"github.com/Orange-OpenSource/nifikop/pkg/nificlient"
	"github.com/Orange-OpenSource/nifikop/pkg/util/clientconfig"
	"github.com/go-logr/logr"
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
	return fmt.Sprintf("%s", cluster.Name)
}

func (cluster ExternalCluster) IsReady(log logr.Logger) bool {
	nClient, err := common.NewClusterConnection(log, cluster.NifiConfig)
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
