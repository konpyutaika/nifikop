package nificlient

import "emperror.dev/errors"

var ErrNodeNotConnected          = errors.New("The targeted node id disconnected")
var ErrNifiClusterNotReturned200 = errors.New("non 200 response from NiFi cluster")
var ErrNifiClusterReturned404    = errors.New("404 response from NiFi cluster")
var ErrNifiClusterNodeNotFound   = errors.New("The target node id doesn't exist in the cluster")

var ErrNoNodeClientsAvailable    = errors.New("Cannot create a node client to perform actions")
