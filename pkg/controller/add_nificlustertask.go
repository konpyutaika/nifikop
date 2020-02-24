package controller

import (
	"github.com/orangeopensource/nifi-operator/pkg/controller/nificlustertask"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, nificlustertask.Add)
}
