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

const (
	// PKIBackendCertManager invokes cert-manager for user certificate management
	PKIBackendCertManager PKIBackend = "cert-manager"
)

// GracefulActionState holds information about GracefulAction State
type GracefulActionState struct {
	// ErrorMessage holds the information what happened with CC
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
	// RackAwarenessState holds info about rack awareness status
	RackAwarenessState RackAwarenessState `json:"rackAwarenessState"`
	// GracefulActionState holds info about cc action status
	GracefulActionState GracefulActionState `json:"gracefulActionState"`
	// ConfigurationState holds info about the config
	ConfigurationState ConfigurationState `json:"configurationState"`
}

const (
	// Configured states the node is running
	Configured RackAwarenessState = "Configured"
	// WaitingForRackAwareness states the node is waiting for the rack awareness config
	WaitingForRackAwareness RackAwarenessState = "WaitingForRackAwareness"
	// GracefulUpscaleSucceeded states the node is updated gracefully
	GracefulUpscaleSucceeded State = "GracefulUpscaleSucceeded"
	// GracefulUpscaleSucceeded states the node is updated gracefully
	GracefulDownscaleSucceeded State = "GracefulDownscaleSucceeded"
	// GracefulUpdateRunning states the node update task is still running in Nifi cluster
	GracefulUpdateRunning State = "GracefulUpdateRunning"
	// GracefulUpdateFailed states the node could not be updated gracefully
	GracefulUpdateFailed State = "GracefulUpdateFailed"
	// GracefulUpdateRequired states the node requires an updated gracefully
	GracefulUpdateRequired State = "GracefulUpdateRequired"
	// GracefulUpdateNotRequired states the node is the part of the initial cluster where Nifi cluster is still in creating stage
	GracefulUpdateNotRequired State = "GracefulUpdateNotRequired"
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
	//
	DisconnectNodeAction ActionStep	= "DISCONNECTING"
	//
	DisconnectStatus ActionStep = "DISCONNECTED"
	//
	OffloadNodeAction ActionStep = "OFFLOADING"
	//
	OffloadStatus ActionStep = "OFFLOADED"
	//
	RemovePodAction ActionStep = "POD_REMOVING"
	//
	RemovePodStatus ActionStep = "POD_REMOVED"
	//
	RemoveNodeAction ActionStep = "REMOVING"
	//
	RemoveStatus ActionStep = "REMOVED"
	//
	ConnectNodeAction ActionStep = "CONNECTING"
	//
	ConnectStatus ActionStep = "CONNECTED"
)
