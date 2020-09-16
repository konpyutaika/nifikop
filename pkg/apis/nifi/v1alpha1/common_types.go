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

package v1alpha1

// DataflowState defines the state of a NifiDataflow
type DataflowState string

// DataflowUpdateRequestType defines the type of versioned flow update request
type DataflowUpdateRequestType string

// DataflowUpdateStrategy defines the type of strategy to update a flow
type DataflowUpdateStrategy string

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

//  InitClusterNode holds info about if the node was part of the init cluster setup
type InitClusterNode bool

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

// ClusterReference states a reference to a cluster for dataflow/registryclient/user
// provisioning
type ClusterReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// RegistryClientReference states a reference to a registry client for dataflow
// provisioning
type RegistryClientReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// ParameterContextReference states a reference to a parameter context for dataflow
// provisioning
type ParameterContextReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// SecretReference states a reference to a secret for parameter context
// provisioning
type SecretReference struct {
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
	// DataflowStateCreated describes the status of a NifiDataflow as created
	DataflowStateCreated DataflowState = "Created"
	// DataflowStateStarting describes the status of a NifiDataflow as starting
	DataflowStateStarting DataflowState = "Starting"
	// DataflowStateRunning describes the status of a NifiDataflow as running
	DataflowStateRan DataflowState = "Ran"
	// DataflowStateOutOfSync describes the status of a NifiDataflow as out of sync
	DataflowStateOutOfSync DataflowState = "OutOfSync"
	// DataflowStateInSync describes the status of a NifiDataflow as in sync
	DataflowStateInSync DataflowState = "InSync"

	// RevertRequestType defines a revert changes request.
	RevertRequestType DataflowUpdateRequestType = "Revert"
	// UpdateRequestType defines an update version request.
	UpdateRequestType DataflowUpdateRequestType = "Update"

	// DrainStrategy leads to shutting down only input components (Input processors, remote input process group)
	// and dropping all flowfiles from the flow.
	DrainStrategy DataflowUpdateStrategy = "drain"
	// DropStrategy leads to shutting down all components and dropping all flowfiles from the flow.
	DropStrategy DataflowUpdateStrategy = "drop"

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
	// ClientCertKey stores the client certificate (operator usage)
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
	// InitClusterNode contains if this nodes was part of the initial cluster
	InitClusterNode InitClusterNode `json:"initClusterNode"`
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

	// NifiClusterInitializing states that the cluster is still in initializing stage
	NifiClusterInitializing ClusterState = "ClusterInitializing"
	// NifiClusterInitialized states that the cluster is initialized
	NifiClusterInitialized ClusterState = "ClusterInitialized"
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
	DisconnectNodeAction ActionStep = "DISCONNECTING"
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

	// IsInitClusterNode states the node is part of initial cluster setup
	IsInitClusterNode InitClusterNode = true
	// NotInitClusterNode states the node is not part of initial cluster setup
	NotInitClusterNode InitClusterNode = false
)
