package util

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

// IntstrPointer generate IntOrString pointer from int.
func IntstrPointer(i int) *intstr.IntOrString {
	is := intstr.FromInt(i)
	return &is
}

// Int64Pointer generates int64 pointer from int64.
func Int64Pointer(i int64) *int64 {
	return &i
}

// Int32Pointer generates int32 pointer from int32.
func Int32Pointer(i int32) *int32 {
	return &i
}

// BoolPointer generates bool pointer from bool.
func BoolPointer(b bool) *bool {
	return &b
}

// IntPointer generates int pointer from int.
func IntPointer(i int) *int {
	return &i
}

// StringPointer generates string pointer from string.
func StringPointer(s string) *string {
	return &s
}

// MapStringStringPointer generates a map[string]*string.
func MapStringStringPointer(in map[string]string) (out map[string]*string) {
	out = make(map[string]*string, 0)
	for k, v := range in {
		out[k] = StringPointer(v)
	}
	return
}

// MergeHostAliases takes two host alias lists and merges them. For IP conflicts, the latter/second list takes precedence.
func MergeHostAliases(globalAliases []corev1.HostAlias, overrideAliases []corev1.HostAlias) []corev1.HostAlias {
	aliasesMap := map[string]corev1.HostAlias{}
	aliases := []corev1.HostAlias{}

	for _, alias := range globalAliases {
		aliasesMap[alias.IP] = alias
	}
	// the below will override any existing IPs
	for _, alias := range overrideAliases {
		aliasesMap[alias.IP] = alias
	}
	for _, alias := range aliasesMap {
		aliases = append(aliases, alias)
	}

	return aliases
}

// MergeLabels merges two given labels.
func MergeLabels(l ...map[string]string) map[string]string {
	res := make(map[string]string)

	for _, v := range l {
		for lKey, lValue := range v {
			res[lKey] = lValue
		}
	}
	return res
}

// MonitoringAnnotations returns specific prometheus annotations.
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

// ConvertStringToInt32 converts the given string to int32.
func ConvertStringToInt32(s string) int32 {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return -1
	}
	return int32(i)
}

// IsSSLEnabledForInternalCommunication checks if ssl is enabled for internal communication.
func IsSSLEnabledForInternalCommunication(l []v1.InternalListenerConfig) (enabled bool) {
	for _, listener := range l {
		if strings.EqualFold(listener.Type, "ssl") {
			enabled = true
			break
		}
	}
	return enabled
}

// ConvertMapStringToMapStringPointer converts a simple map[string]string to map[string]*string.
func ConvertMapStringToMapStringPointer(inputMap map[string]string) map[string]*string {
	result := map[string]*string{}
	for key, value := range inputMap {
		result[key] = StringPointer(value)
	}
	return result
}

// StringSliceCompare returns true if the two lists of string are the same.
func StringSliceCompare(list1 []string, list2 []string) bool {
	if len(list1) != len(list2) {
		return false
	}
	for _, v := range list1 {
		if !StringSliceContains(list2, v) {
			return false
		}
	}
	return true
}

// StringSliceStrictCompare returns true if the two lists of string are the same with the order taking into account.
func StringSliceStrictCompare(list1 []string, list2 []string) bool {
	if len(list1) != len(list2) {
		return false
	}
	for i, v := range list1 {
		if list2[i] != v {
			return false
		}
	}
	return true
}

// StringSliceContains returns true if list contains s.
func StringSliceContains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// StringSliceRemove will remove s from list.
func StringSliceRemove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}

// ParsePropertiesFormat parses the properties format configuration into map[string]string.
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

// GetNodeConfig compose the nodeConfig for a given nifi node.
func GetNodeConfig(node v1.Node, clusterSpec v1.NifiClusterSpec) (*v1.NodeConfig, error) {
	nConfig := &v1.NodeConfig{}
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

// GetNodeImage returns the used node image.
func GetNodeImage(nodeConfig *v1.NodeConfig, clusterImage string) string {
	if nodeConfig.Image != "" {
		return nodeConfig.Image
	}
	return clusterImage
}

// NifiUserSliceContains returns true if list contains s.
func NifiUserSliceContains(list []*v1.NifiUser, u *v1.NifiUser) bool {
	for _, v := range list {
		if reflect.DeepEqual(&v, &u) {
			return true
		}
	}
	return false
}

func NodesToIdList(nodes []v1.Node) (ids []int32) {
	for _, node := range nodes {
		ids = append(ids, node.Id)
	}
	return
}

func NodesToIdMap(nodes []v1.Node) (nodeMap map[int32]v1.Node) {
	nodeMap = make(map[int32]v1.Node)
	for _, node := range nodes {
		nodeMap[node.Id] = node
	}
	return
}

// SubtractNodes removes nodesToRemove from the originalNodes list by the node's Ids and returns the result.
func SubtractNodes(originalNodes []v1.Node, nodesToRemove []v1.Node) (results []v1.Node) {
	if len(originalNodes) == 0 || len(nodesToRemove) == 0 {
		return originalNodes
	}
	nodesToRemoveMap := NodesToIdMap(nodesToRemove)
	results = []v1.Node{}

	for _, node := range originalNodes {
		if _, found := nodesToRemoveMap[node.Id]; !found {
			// results are those which are _not_ in the nodesToRemove map
			results = append(results, node)
		}
	}
	return results
}

// computes the max between 2 ints.
func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// computes the max in a int32 slice.
func MaxSlice32(s []int32) (int32, error) {
	if len(s) == 0 {
		return 0, errors.New("Cannot detect a maximum value in an empty slice")
	}

	max := s[0]
	for _, v := range s {
		if v > max {
			max = v
		}
	}

	return max, nil
}

// computes the max in a int32 slice.
func MinSlice32(s []int32) (int32, error) {
	if len(s) == 0 {
		return 0, errors.New("Cannot detect a minimum value in an empty slice")
	}

	min := s[0]
	for _, v := range s {
		if v < min {
			min = v
		}
	}

	return min, nil
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
	// @TODO : check what is the expected behavior with offset
	duration := interval + rand.Intn(offset+1) - (offset / 2)
	duration = Max(duration, rand.Intn(5)+1) // make sure duration does not go zero for very large offsets
	return time.Duration(duration) * time.Second
}

func IsK8sPrior1_21() bool {
	major, minor, err := GetK8sVersion()
	return err == nil && *major == 1 && *minor < 21
}

func GetK8sVersion() (major *int, minor *int, err error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	info, err := discoveryClient.ServerVersion()
	if err != nil {
		return nil, nil, err
	}
	maj, err := strconv.Atoi(info.Major)
	if err != nil {
		return nil, nil, err
	}
	major = &maj
	min, err := strconv.Atoi(info.Minor)
	if err != nil {
		return nil, nil, err
	}
	minor = &min
	return major, minor, nil
}
