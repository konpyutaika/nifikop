// Copyright 2020 Orange SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package apis

package nificlient

import (
	"net/http"

	"emperror.dev/errors"
)

var ErrNodeNotConnected = errors.New("The targeted node id disconnected")
var ErrNifiClusterNotReturned200 = errors.New("non 200 response from NiFi cluster")
var ErrNifiClusterNotReturned201 = errors.New("non 201 response from NiFi cluster")
var ErrNifiClusterReturned404 = errors.New("404 response from NiFi cluster")
var ErrNifiClusterNodeNotFound = errors.New("The target node id doesn't exist in the cluster")

var ErrNoNodeClientsAvailable = errors.New("Cannot create a node client to perform actions")

func errorGetOperation(rsp *http.Response, body *string, err error) error {
	if rsp != nil && rsp.StatusCode == 404 {
		log.Info("404 response from nifi node: " + rsp.Status)
		return ErrNifiClusterReturned404
	}

	if rsp != nil && rsp.StatusCode != 200 {
		log.Error(errors.New("Non 200 response from nifi node: "+rsp.Status), *body)
		return ErrNifiClusterNotReturned200
	}

	if err != nil || rsp == nil {
		log.Error(err, "Error during talking to nifi node")
		return err
	}
	return nil
}

func errorCreateOperation(rsp *http.Response, body *string, err error) error {
	if rsp != nil && rsp.StatusCode != 201 {
		log.Error(errors.New("Non 201 response from nifi node: "+rsp.Status), *body)
		return ErrNifiClusterNotReturned201
	}

	if err != nil || rsp == nil {
		log.Error(err, "Error during talking to nifi node")
		return err
	}

	return nil
}

func errorUpdateOperation(rsp *http.Response, body *string, err error) error {
	if rsp != nil && rsp.StatusCode != 200 && rsp.StatusCode != 202 {
		log.Error(errors.New("Non 200 response from nifi node: "+rsp.Status), *body)
		return ErrNifiClusterNotReturned200
	}

	if err != nil || rsp == nil {
		log.Error(err, "Error during talking to nifi node")
		return err
	}

	return nil
}

func errorDeleteOperation(rsp *http.Response, body *string, err error) error {
	if rsp != nil && rsp.StatusCode == 404 {
		log.Error(errors.New("404 response from nifi node: "+rsp.Status), *body)
		return nil
	}

	if rsp != nil && rsp.StatusCode != 200 {
		log.Error(errors.New("Non 200 response from nifi node: "+rsp.Status), *body)
		return ErrNifiClusterNotReturned200
	}

	if err != nil || rsp == nil {
		log.Error(err, "Error during talking to nifi node")
		return err
	}

	return nil
}
