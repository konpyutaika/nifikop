package nifi

import (
	"fmt"
	"github.com/erdrix/nifikop/pkg/resources/templates"
	"strconv"
	"time"
)

const (
	// AllNodeServiceTemplate template for Nifi all nodes service
	AllNodeServiceTemplate = "%s-all-node"
	// HeadlessServiceTemplate template for Nifi headless service
	HeadlessServiceTemplate = "%s-headless"
)

// ParseTimeStampToUnixTime parses the given CC timeStamp to time format
func ParseTimeStampToUnixTime(timestamp string) (time.Time, error) {
	timeStampLayout := "Mon, 2 Jan 2006 15:04:05 GMT"
	t, err := time.Parse(timeStampLayout, timestamp)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func ComputeHostname(headlessServiceEnabled bool, nodeId int32, clusterName, namespace string) string {
	if headlessServiceEnabled {
		return fmt.Sprintf("%s.%s-headless.%s.svc.cluster.local", fmt.Sprintf(templates.NodeNameTemplate,clusterName, nodeId), clusterName, namespace)
	} else {
		return fmt.Sprintf("%s.%s.svc.cluster.local", fmt.Sprintf(templates.NodeNameTemplate,clusterName, nodeId), namespace)
	}
}

func ParseStringToInt32(nodeId string) (int32, error) {
	intNodeId, err := strconv.ParseInt(nodeId, 10, 32)
	int32NodeId := int32(intNodeId)

	return int32NodeId, err
}
