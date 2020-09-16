package clientwrappers

import (
	"fmt"

	"github.com/Orange-OpenSource/nifikop/pkg/nificlient"
	"github.com/go-logr/logr"
)

func ErrorUpdateOperation(log logr.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error(err, fmt.Sprintf("%s failed since Nifi node returned non 200", action))
	}

	if err != nil {
		log.Error(err, "could not communicate with nifi node")
	}
	return err
}

func ErrorGetOperation(log logr.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error(err, fmt.Sprintf("%s failed since Nifi node returned non 200", action))
	}

	if err != nil {
		log.Error(err, "could not communicate with nifi node")
	}

	return err
}

func ErrorCreateOperation(log logr.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned201 {
		log.Error(err, fmt.Sprintf("%s request failed since Nifi node returned non 201", action))
	}

	if err != nil {
		log.Error(err, "could not communicate with nifi node")
	}

	return err
}

func ErrorRemoveOperation(log logr.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error(err, fmt.Sprintf("%s failed since Nifi node returned non 200", action))
	}

	if err != nil {
		log.Error(err, "could not communicate with nifi node")
	}

	return err
}
