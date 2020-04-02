// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//Ò
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

// RackAwarenessState stores info about rack awareness status
type RackAwarenessState string

// State holds info about the state of action
type State string

// Action step holds info about the action step
type ActionStep string

// ClusterState holds info about the cluster state
type ClusterState string

// ConfigurationState holds info about the configuration state
type ConfigurationState string

// PKIBackend represents an interface implementing the PKIManager
type PKIBackend string

func (r State) IsUpscale() bool {
	return r == GracefulUpscaleRequired || r == GracefulUpscaleSucceeded || r == GracefulUpscaleRunning
}

func (r State) IsDownscale() bool {
	return r == GracefulDownscaleRequired || r == GracefulDownscaleSucceeded || r == GracefulDownscaleRunning
}

func (r State) IsRunningState() bool {
	return r == GracefulDownscaleRunning || r == GracefulUpscaleRunning
}

func (r State) IsRequiredState() bool {
	return r == GracefulDownscaleRequired || r == GracefulUpscaleRequired
}

func (r State) Complete() State {
	switch r {
	case GracefulUpscaleRequired, GracefulUpscaleRunning:
		return GracefulUpscaleSucceeded
	case GracefulDownscaleRequired, GracefulDownscaleRunning:
		return GracefulDownscaleSucceeded
	default:
		return r
	}
}

// NifiAccessType hold info about Nifi ACL
type NifiAccessType string

// UserState defines the state of a NifiUser
type UserState string

// ClusterReference states a reference to a cluster for topic/user
// provisioning
type ClusterReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

const (
	// PKIBackendCertManager invokes cert-manager for user certificate management
	PKIBackendCertManager PKIBackend = "cert-manager"
	// TODO : Add vault
	//PKIBackendVault invokes vault PKI for user certificate management
	//PKIBackendVault PKIBackend = "vault"
)

const (
	// UserStateCreated describes the status of a NifiUser as created
	UserStateCreated UserState = "created"
	// TLSCert is where a cert is stored in a user secret when requested
	TLSCert string = "tls.crt"
	// TLSCert is where a private key is stored in a user secret when requested
	TLSKey string = "tls.key"
	// TLSJKSKey is where a JKS is stored in a user secret when requested
	TLSJKSKey string = "tls.jks"
	// CoreCACertKey is where ca ceritificates are stored in user certificates
	CoreCACertKey string = "ca.crt"
	// CACertKey is the key where the CA certificate is stored in the operator secrets
	CACertKey string = "caCert"
	// CAPrivateKeyKey stores the private key for the CA
	CAPrivateKeyKey string = "caKey"
	// ClientCertKey stores the client certificate (cruisecontrol/operator usage)
	ClientCertKey string = "clientCert"
	// ClientPrivateKeyKey stores the client private key
	ClientPrivateKeyKey string = "clientKey"
	// PeerCertKey stores the peer certificate (node certificates)
	PeerCertKey string = "peerCert"
	// PeerPrivateKeyKey stores the peer private key
	PeerPrivateKeyKey string = "peerKey"
	// PasswordKey stores the JKS password
	PasswordKey string = "password"
)

// GracefulActionState holds information about GracefulAction State
type GracefulActionState struct {
	// ErrorMessage holds the information what happened with Nifi Cluster
	ErrorMessage string `json:"errorMessage"`
	// ActionStep holds info about the action step ran
	ActionStep ActionStep `json:"actionStep,omitempty"`
	// TaskStarted hold the time when the execution started
	TaskStarted string `json:"TaskStarted,omitempty"`
	// ActionState holds the information about Action state
	State State `json:"actionState"`
}

// NifiState holds information about nifi state
type NodeState struct {
	// GracefulActionState holds info about nifi cluster action status
	GracefulActionState GracefulActionState `json:"gracefulActionState"`
	// ConfigurationState holds info about the config
	ConfigurationState ConfigurationState `json:"configurationState"`
}

// RackAwarenessState holds info about rack awareness status
//RackAwarenessState RackAwarenessState `json:"rackAwarenessState"`

const (
	// Configured states the node is running
	Configured RackAwarenessState = "Configured"

	// GracefulUpscaleRequired states that a node upscale is required
	GracefulUpscaleRequired State = "GracefulUpscaleRequired"
	// GracefulUpscaleRunning states that the node upscale task is still running
	GracefulUpscaleRunning State = "GracefulUpscaleRunning"
	// GracefulUpscaleSucceeded states the node is updated gracefully
	GracefulUpscaleSucceeded State = "GracefulUpscaleSucceeded"

	// Downscale nifi cluster states
	// GracefulDownscaleRequired states that a node downscale is required
	GracefulDownscaleRequired State = "GracefulDownscaleRequired"
	// GracefulDownscaleRunning states that the node downscale is still running in
	GracefulDownscaleRunning State = "GracefulDownscaleRunning"
	// GracefulUpscaleSucceeded states the node is updated gracefully
	GracefulDownscaleSucceeded State = "GracefulDownscaleSucceeded"

	// NifiClusterReconciling states that the cluster is still in reconciling stage
	NifiClusterReconciling ClusterState = "ClusterReconciling"
	// NifiClusterRollingUpgrading states that the cluster is rolling upgrading
	NifiClusterRollingUpgrading ClusterState = "ClusterRollingUpgrading"
	// NifiClusterRunning states that the cluster is in running state
	NifiClusterRunning ClusterState = "ClusterRunning"

	// ConfigInSync states that the generated nodeConfig is in sync with the Node
	ConfigInSync ConfigurationState = "ConfigInSync"
	// ConfigOutOfSync states that the generated nodeConfig is out of sync with the Node
	ConfigOutOfSync ConfigurationState = "ConfigOutOfSync"

	// DisconnectNodeAction states that the NiFi node is disconnecting from NiFi Cluster
	DisconnectNodeAction ActionStep	= "DISCONNECTING"
	// DisconnectStatus states that the NiFi node is disconnected from NiFi Cluster
	DisconnectStatus ActionStep = "DISCONNECTED"
	// OffloadNodeAction states that the NiFi node is offloading data to NiFi Cluster
	OffloadNodeAction ActionStep = "OFFLOADING"
	// OffloadStatus states that the NiFi node offloaded data to NiFi Cluster
	OffloadStatus ActionStep = "OFFLOADED"
	// RemovePodAction states that the NiFi node pod and object related are removing by operator.
	RemovePodAction ActionStep = "POD_REMOVING"
	// RemovePodAction states that the NiFi node pod and object related have been removed by operator.
	RemovePodStatus ActionStep = "POD_REMOVED"
	// RemoveNodeAction states that the NiFi node is removing from NiFi Cluster
	RemoveNodeAction ActionStep = "REMOVING"
	// RemoveStatus states that the NiFi node is removed from NiFi Cluster
	RemoveStatus ActionStep = "REMOVED"
	// ConnectNodeAction states that the NiFi node is connecting to the NiFi Cluster
	ConnectNodeAction ActionStep = "CONNECTING"
	// ConnectStatus states that the NiFi node is connected to the NiFi Cluster
	ConnectStatus ActionStep = "CONNECTED"
)
