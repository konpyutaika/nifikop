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

package zookeeper

import (
	"reflect"
	"testing"
)

func TestPrepareConnectionAddress(t *testing.T) {
	zkAddresse := "zkhostname.subdomain.com:8081"
	zkPath := "/zkpath"
	expectedZkConnectionAddress := "zkhostname.subdomain.com:8081/zkpath"
	zkConnectionAddress := PrepareConnectionAddress(zkAddresse, zkPath)

	if !reflect.DeepEqual(zkConnectionAddress, expectedZkConnectionAddress) {
		t.Errorf("Expected %+v\nGot %+v", expectedZkConnectionAddress, zkConnectionAddress)
	}
}

func TestGetHostnameAddress(t *testing.T) {
	zkAddresse := "zkhostname.subdomain.com:8081"
	expectedZkHostname := "zkhostname.subdomain.com"
	zkHostname := GetHostnameAddress(zkAddresse)

	if !reflect.DeepEqual(zkHostname, expectedZkHostname) {
		t.Errorf("Expected %+v\nGot %+v", expectedZkHostname, zkHostname)
	}
}

func TestGetHostnamePort(t *testing.T) {
	zkAddresse := "zkhostname.subdomain.com:8081"
	expectedZkPort := "8081"
	zkPort := GetPortAddress(zkAddresse)

	if !reflect.DeepEqual(zkPort, expectedZkPort) {
		t.Errorf("Expected %+v\nGot %+v", expectedZkPort, zkPort)
	}
}
