package zookeeper

import "strings"

// PrepareConnectionAddress prepares the proper address for Nifi and CC
// The required path for Nifi and CC looks 'example-1:2181/nifi'
func PrepareConnectionAddress(zkAddress string, zkPath string) string {
	return zkAddress + zkPath
}

func GetHostnameAddress(zkAddress string) string {
	return strings.Split(zkAddress, ":")[0]
}

func GetPortAddress(zkAddress string) string {
	return strings.Split(zkAddress, ":")[1]
}
