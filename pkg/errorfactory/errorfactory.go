package errorfactory

import "emperror.dev/errors"

// ResourceNotReady states that resource is not ready.
type ResourceNotReady struct{ error }

// APIFailure states that something went wrong with the api.
type APIFailure struct{ error }

// VaultAPIFailure states an error communicating with the configured vault server.
type VaultAPIFailure struct{ error }

// StatusUpdateError states that the operator failed to update the Status.
type StatusUpdateError struct{ error }

// NodesUnreachable states that the given node is unreachable.
type NodesUnreachable struct{ error }

// NodesNotReady states that the node is not ready.
type NodesNotReady struct{ error }

// NodesRequestError states that the node could not understand the request.
type NodesRequestError struct{ error }

// GracefulUpscaleFailed states that the operator failed to update the cluster gracefully.
type GracefulUpscaleFailed struct{ error }

// TooManyResources states that too many resource found.
type TooManyResources struct{ error }

// InternalError states that internal error happened.
type InternalError struct{ error }

// FatalReconcileError states that a fatal error happened.
type FatalReconcileError struct{ error }

// ReconcileRollingUpgrade states that rolling upgrade is reconciling.
type ReconcileRollingUpgrade struct{ error }

// NilClientConfig states that the client config is nil.
type NilClientConfig struct{ error }

// NifiClusterNotReady states that NC is not ready to receive actions.
type NifiClusterNotReady struct{ error }

// NifiClusterTaskRunning states that NC task is still running.
type NifiClusterTaskRunning struct{ error }

// NifiClusterTaskTimeout states that NC task timed out.
type NifiClusterTaskTimeout struct{ error }

// NifiClusterTaskFailure states that NC task was not found (CC restart?) or failed.
type NifiClusterTaskFailure struct{ error }

// NifiConnectionDropping states that flowfile drop is still running.
type NifiConnectionDropping struct{ error }

// NifiConnectionDropRequestNotFound states that flowfile drop request was not found.
type NifiConnectionDropRequestNotFound struct{ error }

// NifiFlowDraining states that flowfile drop is still draining.
type NifiFlowDraining struct{ error }

// NifiParameterContextUpdateRequestRunning states that the parameter context update request is still running.
type NifiParameterContextUpdateRequestRunning struct{ error }

// NifiParameterContextUpdateRequestNotFound states that the parameter context update request was not found.
type NifiParameterContextUpdateRequestNotFound struct{ error }

// NifiFlowUpdateRequestRunning states that the flow update request is still running.
type NifiFlowUpdateRequestRunning struct{ error }

// NifiFlowUpdateRequestNotFound states that the flow update request was not found.
type NifiFlowUpdateRequestNotFound struct{ error }

// NifiFlowControllerServiceScheduling states that the flow's controller service are still scheduling.
type NifiFlowControllerServiceScheduling struct{ error }

// NifiFlowSyncing states that the flow is still syncing.
type NifiFlowSyncing struct{ error }

// NifiFlowScheduling states that the flow is still scheduling.
type NifiFlowScheduling struct{ error }

// NifiReportingTasksValidating states that the reporting task is still validating.
type NifiReportingTasksValidating struct{ error }

// NifiReportingTasksInvalid states that the reporting task is invalid.
type NifiReportingTasksInvalid struct{ error }

// NifiConnectionSyncing states that the connection is still syncing.
type NifiConnectionSyncing struct{ error }

// NifiConnectionDeleting states that the connection is still deleting.
type NifiConnectionDeleting struct{ error }

// New creates a new error factory error.
func New(t interface{}, err error, msg string, wrapArgs ...interface{}) error {
	wrapped := errors.WrapIfWithDetails(err, msg, wrapArgs...)
	switch t.(type) {
	case ResourceNotReady:
		return ResourceNotReady{wrapped}
	case APIFailure:
		return APIFailure{wrapped}
	case VaultAPIFailure:
		return VaultAPIFailure{wrapped}
	case StatusUpdateError:
		return StatusUpdateError{wrapped}
	case NodesUnreachable:
		return NodesUnreachable{wrapped}
	case NodesNotReady:
		return NodesNotReady{wrapped}
	case NodesRequestError:
		return NodesRequestError{wrapped}
	case GracefulUpscaleFailed:
		return GracefulUpscaleFailed{wrapped}
	case TooManyResources:
		return TooManyResources{wrapped}
	case InternalError:
		return InternalError{wrapped}
	case FatalReconcileError:
		return FatalReconcileError{wrapped}
	case ReconcileRollingUpgrade:
		return ReconcileRollingUpgrade{wrapped}
	}
	return wrapped
}
