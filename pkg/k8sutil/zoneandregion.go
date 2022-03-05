package k8sutil

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func specificNodeLabels(nodeName string, client runtimeClient.Client, filter []string) (map[string]string, error) {
	node := &corev1.Node{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: nodeName, Namespace: ""}, node)
	if err != nil {
		return nil, err
	}
	requestedLabels := map[string]string{}

	for _, label := range filter {
		if val, ok := node.Labels[label]; ok {
			requestedLabels[label] = val
		}
	}
	return requestedLabels, nil
}
