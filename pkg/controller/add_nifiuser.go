package controller

import (
	"gitlab.si.francetelecom.fr/kubernetes/nifikop/pkg/controller/nifiuser"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, nifiuser.Add)
}
