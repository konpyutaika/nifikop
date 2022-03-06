package k8sutil

import (
	"context"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// LookupNifiCluster returns the running cluster instance based on its name and namespace
func LookupNifiCluster(client runtimeClient.Client, clusterName, clusterNamespace string) (cluster *v1alpha1.NifiCluster, err error) {
	cluster = &v1alpha1.NifiCluster{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: clusterName, Namespace: clusterNamespace}, cluster)
	return
}

// LookupNifiRegistryClient returns the running registry client instance based on its name and namespace
func LookupNifiRegistryClient(client runtimeClient.Client, registryClientName, registryClientNamespace string) (registryClient *v1alpha1.NifiRegistryClient, err error) {
	registryClient = &v1alpha1.NifiRegistryClient{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: registryClientName, Namespace: registryClientNamespace}, registryClient)
	return
}

// LookupNifiParameterContext returns the parameter context instance based on its name and namespace
func LookupNifiParameterContext(client runtimeClient.Client, parameterContextName, parameterContextNamespace string) (parameterContext *v1alpha1.NifiParameterContext, err error) {
	parameterContext = &v1alpha1.NifiParameterContext{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: parameterContextName, Namespace: parameterContextNamespace}, parameterContext)
	return
}

// LookupSecret returns the secret instance based on its name and namespace
func LookupSecret(client runtimeClient.Client, secretName, secretNamespace string) (secret *corev1.Secret, err error) {
	secret = &corev1.Secret{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret)
	return
}

// LookupNifiUser returns the user instance based on its name and namespace
func LookupNifiUser(client runtimeClient.Client, userName, userNamespace string) (user *v1alpha1.NifiUser, err error) {
	user = &v1alpha1.NifiUser{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: userName, Namespace: userNamespace}, user)
	return
}
