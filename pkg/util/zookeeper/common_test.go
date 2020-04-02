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