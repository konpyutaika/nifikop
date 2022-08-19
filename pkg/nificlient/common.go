package nificlient

import (
	"net/http"

	"emperror.dev/errors"
	"go.uber.org/zap"
)

var ErrNodeNotConnected = errors.New("The targeted node id disconnected")
var ErrNifiClusterNotReturned200 = errors.New("non 200 response from NiFi cluster")
var ErrNifiClusterNotReturned201 = errors.New("non 201 response from NiFi cluster")
var ErrNifiClusterReturned404 = errors.New("404 response from NiFi cluster")
var ErrNifiClusterNodeNotFound = errors.New("The target node id doesn't exist in the cluster")

var ErrNoNodeClientsAvailable = errors.New("Cannot create a node client to perform actions")

func errorGetOperation(rsp *http.Response, body *string, err error, log *zap.Logger) error {
	if rsp != nil && rsp.StatusCode == 404 {
		log.Error("404 response from nifi node: ",
			zap.String("statusCode", rsp.Status))
		return ErrNifiClusterReturned404
	}

	if rsp != nil && rsp.StatusCode != 200 {
		log.Error("Non 200 response from nifi node",
			zap.String("statusCode", rsp.Status),
			zap.Stringp("body", body))
		return ErrNifiClusterNotReturned200
	}

	if err != nil || rsp == nil {
		log.Error("Error during talking to nifi node",
			zap.Error(err))
		return err
	}
	return nil
}

func errorCreateOperation(rsp *http.Response, body *string, err error, log *zap.Logger) error {
	if rsp != nil && rsp.StatusCode != 201 {
		log.Error("Non 201 response from nifi node",
			zap.String("statusCode", rsp.Status),
			zap.Stringp("body", body))
		return ErrNifiClusterNotReturned201
	}

	if err != nil || rsp == nil {
		log.Error("Error during talking to nifi node",
			zap.Error(err))
		return err
	}

	return nil
}

func errorUpdateOperation(rsp *http.Response, body *string, err error, log *zap.Logger) error {
	if rsp != nil && rsp.StatusCode != 200 && rsp.StatusCode != 202 {
		log.Error("Non 200 or 202 response from nifi node",
			zap.String("statusCode", rsp.Status),
			zap.Stringp("body", body))
		return ErrNifiClusterNotReturned200
	}

	if err != nil || rsp == nil {
		log.Error("Error during talking to nifi node",
			zap.Error(err))
		return err
	}

	return nil
}

func errorDeleteOperation(rsp *http.Response, body *string, err error, log *zap.Logger) error {
	if rsp != nil && rsp.StatusCode == 404 {
		log.Error("404 response from nifi node: ",
			zap.String("statusCode", rsp.Status))
		return nil
	}

	if rsp != nil && rsp.StatusCode != 200 {
		log.Error("Non 200 response from nifi node",
			zap.String("statusCode", rsp.Status),
			zap.Stringp("body", body))
		return ErrNifiClusterNotReturned200
	}

	if err != nil || rsp == nil {
		log.Error("Error during talking to nifi node",
			zap.Error(err))
		return err
	}

	return nil
}
