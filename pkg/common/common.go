package common

import (
	"fmt"
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/nificlient"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// newNifiFromCluster points to the function for retrieving nifi clients,
// use as var so it can be overwritten from unit tests
var newNifiFromCluster = nificlient.NewFromCluster

// newNodeConnection is a convenience wrapper for creating a node connection
// and creating a safer close function
func NewNodeConnection(log logr.Logger, client client.Client, cluster *v1alpha1.NifiCluster) (node nificlient.NifiClient, err error) {

	// Get a nifi connection
	log.Info(fmt.Sprintf("Retrieving Nifi client for %s/%s", cluster.Namespace, cluster.Name))
	node, err = newNifiFromCluster(client, cluster)
	if err != nil {
		return
	}
	return
}
