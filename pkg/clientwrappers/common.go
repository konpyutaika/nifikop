package clientwrappers

import (
	"fmt"

	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"go.uber.org/zap"
)

func ErrorUpdateOperation(log *zap.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error(fmt.Sprintf("%s failed since Nifi node returned non 200", action), zap.Error(err))
	}

	if err != nil {
		log.Error("could not communicate with nifi node", zap.Error(err))
	}
	return err
}

func ErrorGetOperation(log *zap.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error(fmt.Sprintf("%s failed since Nifi node returned non 200", action), zap.Error(err))
	}

	if err != nil {
		log.Error("could not communicate with nifi node", zap.Error(err))
	}

	return err
}

func ErrorCreateOperation(log *zap.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned201 {
		log.Error(fmt.Sprintf("%s request failed since Nifi node returned non 201", action), zap.Error(err))
	}

	if err != nil {
		log.Error("could not communicate with nifi node", zap.Error(err))
	}

	return err
}

func ErrorRemoveOperation(log *zap.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error(fmt.Sprintf("%s failed since Nifi node returned non 200", action), zap.Error(err))
	}

	if err != nil {
		log.Error("could not communicate with nifi node", zap.Error(err))
	}

	return err
}
