package k8sutil

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func TestErrorIs(t *testing.T) {
	err := &controllerutil.AlreadyOwnedError{
		Object: &v1.ObjectMeta{},
		Owner: v1.OwnerReference{
			Name: "test",
		},
	}

	is := IsAlreadyOwnedError(err)
	if !is {
		t.Errorf("Error is not AlreadyOwnedError")
	}
}
