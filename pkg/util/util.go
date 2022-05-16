package util

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/konpyutaika/nifikop/api/v1alpha1"

	"emperror.dev/errors"
	"github.com/imdario/mergo"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IntstrPointer generate IntOrString pointer from int
func IntstrPointer(i int) *intstr.IntOrString {
	is := intstr.FromInt(i)
	return &is
}

// Int64Pointer generates int64 pointer from int64
func Int64Pointer(i int64) *int64 {
	return &i
}

// Int32Pointer generates int32 pointer from int32
func Int32Pointer(i int32) *int32 {
	return &i
}

// BoolPointer generates bool pointer from bool
func BoolPointer(b bool) *bool {
	return &b
}

// IntPointer generates int pointer from int
func IntPointer(i int) *int {
	return &i
}

// StringPointer generates string pointer from string
func StringPointer(s string) *string {
	return &s
}

// MapStringStringPointer generates a map[string]*string
func MapStringStringPointer(in map[string]string) (out map[string]*string) {
	out = make(map[string]*string, 0)
	for k, v := range in {
		out[k] = StringPointer(v)
	}
	return
}

// MergeLabels merges two given labels
func MergeLabels(l ...map[string]string) map[string]string {
	res := make(map[string]string)

	for _, v := range l {
		for lKey, lValue := range v {
			res[lKey] = lValue
		}
	}
	return res
}

// MonitoringAnnotations returns specific prometheus annotations
func MonitoringAnnotations(port int) map[string]string {
	return map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   strconv.Itoa(port),
	}
}

func MergeAnnotations(annotations ...map[string]string) map[string]string {
	rtn := make(map[string]string)
	for _, a := range annotations {
		for k, v := range a {
			rtn[k] = v
		}
	}

	return rtn
}

// ConvertStringToInt32 converts the given string to int32
func ConvertStringToInt32(s string) int32 {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return -1
	}
	return int32(i)
}

// IsSSLEnabledForInternalCommunication checks if ssl is enabled for internal communication
func IsSSLEnabledForInternalCommunication(l []v1alpha1.InternalListenerConfig) (enabled bool) {

	for _, listener := range l {
		if strings.ToLower(listener.Type) == "ssl" {
			enabled = true
			break
		}
	}
	return enabled
}

// ConvertMapStringToMapStringPointer converts a simple map[string]string to map[string]*string
func ConvertMapStringToMapStringPointer(inputMap map[string]string) map[string]*string {

	result := map[string]*string{}
	for key, value := range inputMap {
		result[key] = StringPointer(value)
	}
	return result
}

// StringSliceContains returns true if list contains s
func StringSliceContains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// StringSliceRemove will remove s from list
func StringSliceRemove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}

// ParsePropertiesFormat parses the properties format configuration into map[string]string
func ParsePropertiesFormat(properties string) map[string]string {
	config := map[string]string{}

	splitProps := strings.Split(properties, "\n")

	for _, line := range splitProps {
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				config[key] = value
			}
		}
	}

	return config
}

// GetNodeConfig compose the nodeConfig for a given nifi node
func GetNodeConfig(node v1alpha1.Node, clusterSpec v1alpha1.NifiClusterSpec) (*v1alpha1.NodeConfig, error) {

	nConfig := &v1alpha1.NodeConfig{}
	if node.NodeConfigGroup == "" {
		return node.NodeConfig, nil
	} else if node.NodeConfig != nil {
		nConfig = node.NodeConfig.DeepCopy()
	}

	err := mergo.Merge(nConfig, clusterSpec.NodeConfigGroups[node.NodeConfigGroup], mergo.WithAppendSlice)
	if err != nil {
		return nil, errors.WrapIf(err, "could not merge nodeConfig with ConfigGroup")
	}
	return nConfig, nil
}

// GetNodeImage returns the used node image
func GetNodeImage(nodeConfig *v1alpha1.NodeConfig, clusterImage string) string {
	if nodeConfig.Image != "" {
		return nodeConfig.Image
	}
	return clusterImage
}

// NifiUserSliceContains returns true if list contains s
func NifiUserSliceContains(list []*v1alpha1.NifiUser, u *v1alpha1.NifiUser) bool {
	for _, v := range list {
		if reflect.DeepEqual(&v, &u) {
			return true
		}
	}
	return false
}

func NodesToIdList(nodes []v1alpha1.Node) (ids []int32) {
	for _, node := range nodes {
		ids = append(ids, node.Id)
	}
	return
}

func NodesToIdMap(nodes []v1alpha1.Node) (nodeMap map[int32]v1alpha1.Node) {
	nodeMap = make(map[int32]v1alpha1.Node)
	for _, node := range nodes {
		nodeMap[node.Id] = node
	}
	return
}

// New nodes are assigned an Id in the following manner:
//
// - Assigned node Ids will always be a non-negative integer starting with zero
//
// - extract and sort the node Ids in the provided node list
//
// - iterate through the node Id list starting with zero. For any unassigned node Id, assign it
//
// - return the list of assigned node Ids
func ComputeNewNodeIds(nodes []v1alpha1.Node, numNewNodes int32) []int32 {
	nodeIdList := NodesToIdList(nodes)
	sort.Slice(nodeIdList, func(i, j int) bool {
		return nodeIdList[i] < nodeIdList[j]
	})

	newNodeIds := []int32{}
	index := int32(0)

	// assign IDs in any gaps in the existing node list, starting with zero
	var i int32
	for i = int32(0); i < nodeIdList[len(nodeIdList)-1] && int32(len(newNodeIds)) < numNewNodes; i++ {
		if nodeIdList[index] == i {
			index++
		} else {
			newNodeIds = append(newNodeIds, i)
		}
	}

	// add any remaining nodes needed
	remainder := numNewNodes - int32(len(newNodeIds))
	for j := int32(1); j <= remainder; j++ {
		newNodeIds = append(newNodeIds, i+j)
	}

	return newNodeIds
}

// SubtractNodes removes nodesToRemove from the originalNodes list by the node's Ids and returns the result
func SubtractNodes(originalNodes []v1alpha1.Node, nodesToRemove []v1alpha1.Node) (results []v1alpha1.Node) {
	if len(originalNodes) == 0 || len(nodesToRemove) == 0 {
		return originalNodes
	}
	nodesToRemoveMap := NodesToIdMap(nodesToRemove)
	results = []v1alpha1.Node{}

	for _, node := range originalNodes {
		if _, found := nodesToRemoveMap[node.Id]; !found {
			// results are those which are _not_ in the nodesToRemove map
			results = append(results, node)
		}
	}
	return results
}

// computes the max between 2 ints
func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func Hash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return string(h.Sum(nil))
}

func GetEnvWithDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func MustConvertToInt(str string, name string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		fmt.Printf("%s", fmt.Errorf("%w problem converting string to integer (%s)", err, name))
		os.Exit(1)
	}
	return i
}

func GetRequeueInterval(interval int, offset int) time.Duration {
	// @TODO : check what is the expected behaviour with offset
	duration := interval + rand.Intn(offset+1) - (offset / 2)
	duration = Max(duration, rand.Intn(5)+1) // make sure duration does not go zero for very large offsets
	return time.Duration(duration) * time.Second
}
