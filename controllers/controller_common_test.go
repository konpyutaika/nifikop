package controllers

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"go.uber.org/zap"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
)

func TestRequeueWithError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	_, err := RequeueWithError(*logger, "test", errors.New("test error"))
	if err == nil {
		t.Error("Expected error to fall through, got nil")
	}
}

func TestReconciled(t *testing.T) {
	res, err := Reconciled()
	if err != nil {
		t.Error("Expected error to be nil, got:", err)
	}
	if res.Requeue {
		t.Error("Expected requeue to be false, got true")
	}
}

func TestGetClusterRefNamespace(t *testing.T) {
	ns := "test-namespace"
	ref := v1.ClusterReference{
		Name: "test-cluster",
	}
	if refNS := GetClusterRefNamespace(ns, ref); refNS != "test-namespace" {
		t.Error("Expected to get 'test-namespace', got:", refNS)
	}
	ref.Namespace = "another-namespace"
	if refNS := GetClusterRefNamespace(ns, ref); refNS != "another-namespace" {
		t.Error("Expected to get 'another-namespace', got:", refNS)
	}
}

func TestGetRegistryClientRefNamespace(t *testing.T) {
	ns := "test-namespace"
	ref := v1.RegistryClientReference{
		Name: "test-cluster",
	}
	if refNS := GetRegistryClientRefNamespace(ns, ref); refNS != "test-namespace" {
		t.Error("Expected to get 'test-namespace', got:", refNS)
	}
	ref.Namespace = "another-namespace"
	if refNS := GetRegistryClientRefNamespace(ns, ref); refNS != "another-namespace" {
		t.Error("Expected to get 'another-namespace', got:", refNS)
	}
}

func TestGetParameterContextRefNamespace(t *testing.T) {
	ns := "test-namespace"
	ref := v1.ParameterContextReference{
		Name: "test-cluster",
	}
	if refNS := GetParameterContextRefNamespace(ns, ref); refNS != "test-namespace" {
		t.Error("Expected to get 'test-namespace', got:", refNS)
	}
	ref.Namespace = "another-namespace"
	if refNS := GetParameterContextRefNamespace(ns, ref); refNS != "another-namespace" {
		t.Error("Expected to get 'another-namespace', got:", refNS)
	}
}

func TestGetSecretRefNamespace(t *testing.T) {
	ns := "test-namespace"
	ref := v1.SecretReference{
		Name: "test-cluster",
	}
	if refNS := GetSecretRefNamespace(ns, ref); refNS != "test-namespace" {
		t.Error("Expected to get 'test-namespace', got:", refNS)
	}
	ref.Namespace = "another-namespace"
	if refNS := GetSecretRefNamespace(ns, ref); refNS != "another-namespace" {
		t.Error("Expected to get 'another-namespace', got:", refNS)
	}
}

func TestGetUserRefNamespace(t *testing.T) {
	ns := "test-namespace"
	ref := v1.UserReference{
		Name: "test-cluster",
	}
	if refNS := GetUserRefNamespace(ns, ref); refNS != "test-namespace" {
		t.Error("Expected to get 'test-namespace', got:", refNS)
	}
	ref.Namespace = "another-namespace"
	if refNS := GetUserRefNamespace(ns, ref); refNS != "another-namespace" {
		t.Error("Expected to get 'another-namespace', got:", refNS)
	}
}

func TestGetComponentRefNamespace(t *testing.T) {
	ns := "test-namespace"
	ref := v1alpha1.ComponentReference{
		Name: "test-cluster",
	}
	if refNS := GetComponentRefNamespace(ns, ref); refNS != "test-namespace" {
		t.Error("Expected to get 'test-namespace', got:", refNS)
	}
	ref.Namespace = "another-namespace"
	if refNS := GetComponentRefNamespace(ns, ref); refNS != "another-namespace" {
		t.Error("Expected to get 'another-namespace', got:", refNS)
	}
}

func TestClusterLabelString(t *testing.T) {
	cluster := &v1.NifiCluster{}
	cluster.Name = "test-cluster"
	cluster.Namespace = "test-namespace"
	if label := ClusterLabelString(cluster); label != "test-cluster.test-namespace" {
		t.Error("Expected label value 'test-cluster.test-namespace', got:", label)
	}
}

/*func TestNewNodeConnection(t *testing.T) {
	cluster := &v1.NifiCluster{}
	cluster.Name = "test-kafka"
	cluster.Namespace = "test-namespace"
	cluster.Spec = v1.NifiClusterSpec{
		ListenersConfig: v1.ListenersConfig{
			InternalListeners: []v1.InternalListenerConfig{
				{ContainerPort: 8080},
			},
		},
	}
	client := fake.NewFakeClient()
	// overwrite the var in controller_common to point kafka connections at mock
	newNifiFromCluster = kafkaclient.NewMockFromCluster

	_, close, err := NewNodeConnection(log, client, cluster)
	if err != nil {
		t.Error("Expected no error got:", err)
	}
	close()

	// reset the newKafkaFromCluster var - will attempt to connect to a cluster
	// that doesn't exist
	newNifiFromCluster = kafkaclient.NewFromCluster
	if _, _, err = NewNodeConnection(log, client, cluster); err == nil {
		t.Error("Expected error got nil")
	} else if !emperrors.As(err, &errorfactory.NodesUnreachable{}) {
		t.Error("Expected brokers unreachable error, got:", err)
	}
}*/

func TestCheckNodeConnectionError(t *testing.T) {
	var err error

	// Test nodes unreachable
	err = errorfactory.New(errorfactory.NodesUnreachable{}, errors.New("test error"), "test message")
	logger, _ := zap.NewDevelopment()

	if res, err := CheckNodeConnectionError(*logger, err); err != nil {
		t.Error("Expected no error in result, got:", err)
	} else {
		if !res.Requeue {
			t.Error("Expected requeue to be true, got false")
		}
		if res.RequeueAfter != time.Duration(15)*time.Second {
			t.Error("Expected 15 second requeue time, got:", res.RequeueAfter)
		}
	}

	// Test nodes not ready
	err = errorfactory.New(errorfactory.NodesNotReady{}, errors.New("test error"), "test message")
	if res, err := CheckNodeConnectionError(*logger, err); err != nil {
		t.Error("Expected no error in result, got:", err)
	} else {
		if !res.Requeue {
			t.Error("Expected requeue to be true, got false")
		}
		if res.RequeueAfter != time.Duration(15)*time.Second {
			t.Error("Expected 15 second requeue time, got:", res.RequeueAfter)
		}
	}

	// test external resource not ready
	err = errorfactory.New(errorfactory.ResourceNotReady{}, errors.New("test error"), "test message")
	if res, err := CheckNodeConnectionError(*logger, err); err != nil {
		t.Error("Expected no error in result, got:", err)
	} else {
		if !res.Requeue {
			t.Error("Expected requeue to be true, got false")
		}
		if res.RequeueAfter != time.Duration(5)*time.Second {
			t.Error("Expected 5 second requeue time, got:", res.RequeueAfter)
		}
	}

	// test default response
	err = errorfactory.New(errorfactory.InternalError{}, errors.New("test error"), "test message")
	if _, err := CheckNodeConnectionError(*logger, err); err == nil {
		t.Error("Expected error to fall through, got nil")
	}
}

func TestApplyClusterRefLabel(t *testing.T) {
	cluster := &v1.NifiCluster{}
	cluster.Name = "test-nifi"
	cluster.Namespace = "test-namespace"

	// nil labels input
	var labels map[string]string
	expected := map[string]string{ClusterRefLabel: "test-nifi.test-namespace"}
	newLabels := ApplyClusterRefLabel(cluster, labels)
	if !reflect.DeepEqual(newLabels, expected) {
		t.Error("Expected:", expected, "Got:", newLabels)
	}

	// existing label but no conflicts
	labels = map[string]string{"otherLabel": "otherValue"}
	expected = map[string]string{
		ClusterRefLabel: "test-nifi.test-namespace",
		"otherLabel":    "otherValue",
	}
	newLabels = ApplyClusterRefLabel(cluster, labels)
	if !reflect.DeepEqual(newLabels, expected) {
		t.Error("Expected:", expected, "Got:", newLabels)
	}

	// existing label with wrong value
	labels = map[string]string{
		ClusterRefLabel: "test-nifi.wrong-namespace",
		"otherLabel":    "otherValue",
	}
	expected = map[string]string{
		ClusterRefLabel: "test-nifi.test-namespace",
		"otherLabel":    "otherValue",
	}
	newLabels = ApplyClusterRefLabel(cluster, labels)
	if !reflect.DeepEqual(newLabels, expected) {
		t.Error("Expected:", expected, "Got:", newLabels)
	}

	// existing labels with correct value - should come back untainted
	labels = map[string]string{
		ClusterRefLabel: "test-nifi.test-namespace",
		"otherLabel":    "otherValue",
	}
	newLabels = ApplyClusterRefLabel(cluster, labels)
	if !reflect.DeepEqual(newLabels, labels) {
		t.Error("Expected:", labels, "Got:", newLabels)
	}
}
