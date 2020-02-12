package zookeeper

import "strings"

// PrepareConnectionAddress prepares the proper address for Nifi and CC
// The required path for Nifi and CC looks 'example-1:2181/nifi'
func PrepareConnectionAddress(zkAddresse string, zkPath string) string {
	return zkAddresse + zkPath
}

//
func GetHostnameAddress(zkAddresse string) string {
	return strings.Split(zkAddresse, ":")[0]
}

//
func GetPortAddress(zkAddresse string) string {
	return strings.Split(zkAddresse, ":")[1]
}

