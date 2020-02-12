package nifi

import (
	"bytes"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	"github.com/orangeopensource/nifi-operator/pkg/apis/nifi/v1alpha1"
	"github.com/orangeopensource/nifi-operator/pkg/resources/templates"
	config "github.com/orangeopensource/nifi-operator/pkg/resources/templates/config"
	"github.com/orangeopensource/nifi-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sort"
	"strings"
	"text/template"
)

func (r *Reconciler) configMap(id int32, nodeConfig *v1alpha1.NodeConfig, log logr.Logger) runtime.Object {
	return &corev1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(
			fmt.Sprintf(nodeConfigTemplate+"-%d", r.NifiCluster.Name, id),
			util.MergeLabels(
				labelsForNifi(r.NifiCluster.Name),
				map[string]string{"nodeId": fmt.Sprintf("%d", id)},
			),
			r.NifiCluster,
		),
		Data: map[string]string{"nifi.properties": r.generateNodeConfig(id, nodeConfig, log)},
	}
}

func (r Reconciler) generateNodeConfig(id int32, nodeConfig *v1alpha1.NodeConfig, log logr.Logger) string {
	parsedReadOnlyClusterConfig := util.ParsePropertiesFormat(r.NifiCluster.Spec.ReadOnlyConfig)
	var parsedReadOnlyNodeConfig = map[string]string{}

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id {
			parsedReadOnlyNodeConfig = util.ParsePropertiesFormat(node.ReadOnlyConfig)
			break
		}
	}

	if err := mergo.Merge(&parsedReadOnlyNodeConfig, parsedReadOnlyClusterConfig); err != nil {
		log.Error(err, "error occurred during merging readonly configs")
	}

	//Generate the Complete Configuration for the Node
	completeConfigMap := map[string]string{}

	if err := mergo.Merge(&completeConfigMap, util.ParsePropertiesFormat(r.getConfigString(nodeConfig, id, log))); err != nil {
		log.Error(err, "error occurred during merging operator generated configs")
	}

	if err := mergo.Merge(&completeConfigMap, parsedReadOnlyNodeConfig); err != nil {
		log.Error(err, "error occurred during merging readOnly config to complete configs")
	}

	completeConfig := []string{}

	for key, value := range completeConfigMap {
		completeConfig = append(completeConfig, fmt.Sprintf("%s=%s", key, value))
	}

	// We need to sort the config every time to avoid diffs occurred because of ranging through map
	sort.Strings(completeConfig)

	return strings.Join(completeConfig, "\n")
}

func (r *Reconciler) getConfigString(nConfig *v1alpha1.NodeConfig, id int32, log logr.Logger) string {
	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.NifiPropertiesTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster":				r.NifiCluster,
		"Id": 						id,
		"ListenerConfig":			generateListenerSpecificConfig(&r.NifiCluster.Spec.ListenersConfig, id, r.NifiCluster.Namespace, r.NifiCluster.Name, r.NifiCluster.Spec.HeadlessServiceEnabled, log),
		"ProvenanceStorage":		generateProvenanceStorageConfig(nConfig.StorageConfigs),
		"SiteToSiteSecure": 		false, 					// TODO: replace by dynamic field
		"ClusterSecure":			false,					// TODO: replace by dynamic field
		"WebProxyHost": 			"",						// TODO: replace by dynamic field
		"NeedClientAuth": 			"",
		"Authorizer": 				"managed-authorizer",	// TODO: replace by dynamic field
		"LdapEnabled": 				false,					// TODO: replace by dynamic field
		"IsNode": 					true,					// TODO: replace by dynamic field
		"ZookeeperConnectString":	r.NifiCluster.Spec.ZKAddresse,
		"ZookeeperPath": 			r.NifiCluster.Spec.GetZkPath(),
	}); err != nil {
		log.Error(err, "error occurred during parsing the config template")
	}
	return out.String()
}

//		"CruiseControlBootstrapServers":      	getInternalListener(r.NifiCluster.Spec.ListenersConfig.InternalListeners, id, r.NifiCluster.Namespace, r.NifiCluster.Name, r.NifiCluster.Spec.HeadlessServiceEnabled),
//		"AdvertisedListenersConfig":          	generateAdvertisedListenerConfig(id, r.NifiCluster.Spec.ListenersConfig, r.NifiCluster.Namespace, r.NifiCluster.Name, r.NifiCluster.Spec.HeadlessServiceEnabled),
//		"ServerKeystorePath":                 	serverKeystorePath,
//		"ClientKeystorePath":                 	clientKeystorePath,
//		"KeystoreFile":                       	v1alpha1.TLSJKSKey,
//		"ServerKeystorePassword":             	serverPass,
//		"ClientKeystorePassword":             	clientPass,
// 		"ControlPlaneListener":               	generateControlPlaneListener(r.NifiCluster.Spec.ListenersConfig.InternalListeners),
// 		"ZookeeperConnectString":             	zookeeperutils.PrepareConnectionAddress(r.NifiCluster.Spec.ZKAddresse, r.NifiCluster.Spec.GetZkPath()),
//		"ListenerConfig":                     	generateListenerSpecificConfig(&r.NifiCluster.Spec.ListenersConfig, log),
//		"SSLEnabledForInternalCommunication":	r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil && util.IsSSLEnabledForInternalCommunication(r.NifiCluster.Spec.ListenersConfig.InternalListeners),
//		"StorageConfig":                      	generateStorageConfig(nConfig.StorageConfigs),

func generateProvenanceStorageConfig(sConfig []v1alpha1.StorageConfig) string {
	// TODO :to enabel
	/*for _, storage := range sConfig {
		if storage.IsProvenanceStorage {
			return storage.PVCSpec.Resources.Requests.Memory().String()
		}
	}*/
	return ProvenanceStorage
}

func generateListenerSpecificConfig(l *v1alpha1.ListenersConfig, id int32, namespace, crName string, headlessServiceEnabled bool, log logr.Logger) string {
	var nifiConfig string

	var hostListener string

	if headlessServiceEnabled {
		hostListener = fmt.Sprintf("%s-%d.%s-headless.%s.svc.cluster.local", crName, id, crName, namespace)
	} else {
		hostListener = fmt.Sprintf("%s-%d.%s.svc.cluster.local", crName, id, namespace)
	}

	clusterPortConfig := "nifi.cluster.node.protocol.port=\n"
	httpPortConfig := "nifi.web.http.port=\n"
	httpHostConfig := "nifi.web.http.host=\n"
	httpsPortConfig := "nifi.web.https.port=\n"
	httpsHostConfig := "nifi.web.https.host=\n"
	s2sPortConfig := "nifi.remote.input.socket.port=\n"

	for _, iListener := range l.InternalListeners {
		switch iListener.Type {
		case clusterListenerType:
			clusterPortConfig = fmt.Sprintf("nifi.cluster.node.protocol.port=%d", iListener.ContainerPort) + "\n"
		case httpListenerType:
			httpPortConfig = fmt.Sprintf("nifi.web.http.port=%d", iListener.ContainerPort) + "\n"
			httpHostConfig = fmt.Sprintf("nifi.web.http.host=%s", hostListener) + "\n"
		case httpsListenerType:
			httpsPortConfig = fmt.Sprintf("nifi.web.https.port=%d", iListener.ContainerPort) + "\n"
			httpsHostConfig = fmt.Sprintf("nifi.web.https.host=%s", hostListener) + "\n"
		case s2sListenerType:
			s2sPortConfig = fmt.Sprintf("nifi.remote.input.socket.port=%d", iListener.ContainerPort) + "\n"
		}
	}
	nifiConfig = nifiConfig +
		clusterPortConfig +
		httpPortConfig +
		httpHostConfig +
		httpsPortConfig +
		httpsHostConfig +
		s2sPortConfig

	nifiConfig = nifiConfig + fmt.Sprintf("nifi.remote.input.host=%s", hostListener) + "\n"
	nifiConfig = nifiConfig + fmt.Sprintf("nifi.cluster.node.address=%s", hostListener) + "\n"
	return nifiConfig
}

// TODO: Change replace
func getInternalListener(iListeners []v1alpha1.InternalListenerConfig, id int32, namespace, crName string, headlessServiceEnabled bool) string {

	internalListener := ""

	/*for _, iListener := range iListeners {
		if iListener.UsedForInnerNodeCommunication {
			if headlessServiceEnabled {
				internalListener = fmt.Sprintf("%s://%s-%d.%s-headless.%s.svc.cluster.local:%d", strings.ToUpper(iListener.Name), crName, id, crName, namespace, iListener.ContainerPort)
			} else {
				internalListener = fmt.Sprintf("%s://%s-%d.%s.svc.cluster.local:%d", strings.ToUpper(iListener.Name), crName, id, namespace, iListener.ContainerPort)
			}
		}
	}*/

	return internalListener
}

func generateStorageConfig(sConfig []v1alpha1.StorageConfig) string {
	mountPaths := []string{}
	for _, storage := range sConfig {
		mountPaths = append(mountPaths, storage.MountPath+`/nifi`)
	}
	return strings.Join(mountPaths, ",")
}

func generateAdvertisedListenerConfig(id int32, l v1alpha1.ListenersConfig, loadBalancerIPs []string, namespace, crName string, headlessServiceEnabled bool) string {
	advertisedListenerConfig := []string{}
	for _, eListener := range l.ExternalListeners {
		// use first element of loadBalancerIPs slice for external listener name
		advertisedListenerConfig = append(advertisedListenerConfig,
			fmt.Sprintf("%s://%s:%d", strings.ToUpper(eListener.Name), loadBalancerIPs[0], eListener.ExternalStartingPort+id))
	}
	for _, iListener := range l.InternalListeners {
		if headlessServiceEnabled {
			advertisedListenerConfig = append(advertisedListenerConfig,
				fmt.Sprintf("%s://%s-%d.%s-headless.%s.svc.cluster.local:%d", strings.ToUpper(iListener.Name), crName, id, crName, namespace, iListener.ContainerPort))
		} else {
			advertisedListenerConfig = append(advertisedListenerConfig,
				fmt.Sprintf("%s://%s-%d.%s.svc.cluster.local:%d", strings.ToUpper(iListener.Name), crName, id, namespace, iListener.ContainerPort))
		}
	}
	return fmt.Sprintf("advertised.listeners=%s\n", strings.Join(advertisedListenerConfig, ","))
}

func generateControlPlaneListener(iListeners []v1alpha1.InternalListenerConfig) string {
	controlPlaneListener := ""

	/*for _, iListener := range iListeners {
		if iListener.UsedForControllerCommunication {
			controlPlaneListener = strings.ToUpper(iListener.Name)
		}
	}*/

	return controlPlaneListener
}
