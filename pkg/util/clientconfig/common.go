package clientconfig

import (
	"context"
	"crypto/tls"

	"go.uber.org/zap"
)

const (
	NifiDefaultTimeout = int64(5)
)

type Manager interface {
	BuildConfig() (*NifiConfig, error)
	BuildConnect() (ClusterConnect, error)
}

type ClusterConnect interface {
	// NodeConnection(log zap.Logger, client client.Client) (node nificlient.NifiClient, err error)
	IsInternal() bool
	IsExternal() bool
	ClusterLabelString() string
	IsReady(log zap.Logger) bool
	Id() string
}

// NifiConfig are the options to creating a new ClusterAdmin client.
type NifiConfig struct {
	NodeURITemplate string
	NodesURI        map[int32]NodeUri
	NifiURI         string
	UseSSL          bool
	TLSConfig       *tls.Config
	ProxyUrl        string

	OperationTimeout   int64
	RootProcessGroupId string
	NodesContext       map[int32]context.Context

	SkipDescribeCluster bool
}

type NodeUri struct {
	HostListener string
	RequestHost  string
}
