package errorfactory

import (
	"errors"
	"reflect"
	"testing"

	emperrors "emperror.dev/errors"
)

var errorTypes = []error{
	ResourceNotReady{},
	APIFailure{},
	VaultAPIFailure{},
	StatusUpdateError{},
	NodesUnreachable{},
	NodesNotReady{},
	NodesRequestError{},
	GracefulUpscaleFailed{},
	TooManyResources{},
	InternalError{},
	FatalReconcileError{},
	NifiClusterNotReady{},
	NifiClusterTaskRunning{},
}

func TestNew(t *testing.T) {
	for _, errType := range errorTypes {
		err := New(errType, errors.New("test-error"), "test-message")
		expected := "test-message: test-error"
		got := err.Error()
		if got != expected {
			t.Error("Expected:", expected, "got:", got)
		}
		if !emperrors.As(err, &errType) {
			t.Error("Expected:", reflect.TypeOf(errType), "got:", reflect.TypeOf(err))
		}
	}
}

func TestNil(t *testing.T) {
	for _, errType := range errorTypes {
		err := New(errType, nil, "no-wrapped-error")
		expected := "no-wrapped-error"
		got := err.Error()
		if got != expected {
			t.Error("Expected:", expected, "got:", got)
		}
		if !emperrors.As(err, &errType) {
			t.Error("Expected:", reflect.TypeOf(errType), "got:", reflect.TypeOf(err))
		}
	}
}
