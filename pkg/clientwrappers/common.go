package clientwrappers

import (
	"go.uber.org/zap"

	"github.com/konpyutaika/nifikop/pkg/nificlient"
)

func ErrorUpdateOperation(log *zap.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error("failed since Nifi node returned non 200",
			zap.String("action", action),
			zap.Error(err))
	}

	if err != nil {
		log.Error("could not communicate with nifi node",
			zap.String("action", action),
			zap.Error(err))
	}
	return err
}

func ErrorGetOperation(log *zap.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error("failed since Nifi node returned non 200",
			zap.String("action", action),
			zap.Error(err))
	}

	if err != nil {
		log.Error("could not communicate with nifi node",
			zap.String("action", action),
			zap.Error(err))
	}

	return err
}

func ErrorCreateOperation(log *zap.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned201 {
		log.Error("failed since Nifi node returned non 201",
			zap.String("action", action),
			zap.Error(err))
	}

	if err != nil {
		log.Error("could not communicate with nifi node",
			zap.String("action", action),
			zap.Error(err))
	}

	return err
}

func ErrorRemoveOperation(log *zap.Logger, err error, action string) error {
	if err == nificlient.ErrNifiClusterNotReturned200 {
		log.Error("failed since Nifi node returned non 200",
			zap.String("action", action),
			zap.Error(err))
	}

	if err != nil {
		log.Error("could not communicate with nifi node",
			zap.String("action", action),
			zap.Error(err))
	}

	return err
}
