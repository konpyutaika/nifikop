package nificlient

import (
	"time"

	"github.com/erdrix/nifikop/pkg/apis/nifi/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NifiClient is the exported interface for Nifi operations
type NifiClient interface {
	NumNodes() int
	// ListDataflows()
	// CreateDataflow()
	// DeleteDataflow()
	// GetDataflow()
	// DescribeDataflow()
	CreateUserACLs(v1alpha1.NifiAccessType, string, string) error
	DeleteUserACLs(string) error

	Nodes() map[int32]string
	//DescribeCluster() ()

	Open() error
	Close() error
}

type nifiClient struct {
	NifiClient
	opts *NifiConfig
	// admin
	// client
	timeout time.Duration
	//nodes
}

func New(opts *NifiConfig) NifiClient {
	nclient := &nifiClient{
		opts:       opts,
		timeout:    time.Duration(opts.OperationTimeout) * time.Second,
	}
	return nclient
}

func (n *nifiClient) Open() error {
	var err error
	return err
}


func (n *nifiClient) Close() error {
	var err error
	return err
}

// NewFromCluster is a convenience wrapper around New() and ClusterConfig()
func NewFromCluster(k8sclient client.Client, cluster *v1alpha1.NifiCluster) (NifiClient, error) {
	var client NifiClient
	var err error
	opts, err := ClusterConfig(k8sclient, cluster)
	if err != nil {
		return nil, err
	}
	client = New(opts)
	err = client.Open()
	return client, err
}