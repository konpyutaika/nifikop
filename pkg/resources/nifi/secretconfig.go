package nifi

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/imdario/mergo"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	configcommon "github.com/konpyutaika/nifikop/pkg/nificlient/config/common"
	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	"github.com/konpyutaika/nifikop/pkg/resources/templates/config"
	"github.com/konpyutaika/nifikop/pkg/util"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
	utilpki "github.com/konpyutaika/nifikop/pkg/util/pki"
)

//	func encodeBase64(toEncode string) []byte {
//		return []byte(base64.StdEncoding.EncodeToString([]byte(toEncode)))
//	}
func (r *Reconciler) secretConfig(id int32, nodeConfig *v1.NodeConfig, serverPass, clientPass string, superUsers []string, log zap.Logger) runtimeClient.Object {
	secret := &corev1.Secret{
		ObjectMeta: templates.ObjectMeta(
			fmt.Sprintf(templates.NodeConfigTemplate+"-%d", r.NifiCluster.Name, id),
			util.MergeLabels(
				nifiutil.LabelsForNifi(r.NifiCluster.Name),
				map[string]string{"nodeId": fmt.Sprintf("%d", id)},
			),
			r.NifiCluster,
		),
		Data: map[string][]byte{
			"nifi.properties":                     []byte(r.generateNifiPropertiesNodeConfig(id, nodeConfig, serverPass, clientPass, superUsers, log)),
			"zookeeper.properties":                []byte(r.generateZookeeperPropertiesNodeConfig(id, nodeConfig, log)),
			"state-management.xml":                []byte(r.getStateManagementConfigString(nodeConfig, id, log)),
			"login-identity-providers.xml":        []byte(r.getLoginIdentityProvidersConfigString(nodeConfig, id, log)),
			"logback.xml":                         []byte(r.getLogbackConfigString(nodeConfig, id, log)),
			"bootstrap.conf":                      []byte(r.generateBootstrapPropertiesNodeConfig(id, nodeConfig, log)),
			"bootstrap-notification-services.xml": []byte(r.getBootstrapNotificationServicesConfigString(nodeConfig, id, log)),
		},
	}

	if configcommon.UseSSL(r.NifiCluster) {
		secret.Data["authorizers.xml"] = []byte(r.getAuthorizersConfigString(nodeConfig, id, log))
	}
	return secret
}

////////////////////////////////////
//  Nifi properties configuration //
////////////////////////////////////

func (r Reconciler) generateNifiPropertiesNodeConfig(id int32, nodeConfig *v1.NodeConfig, serverPass, clientPass string, superUsers []string, log zap.Logger) string {
	var readOnlyClusterConfig map[string]string
	if &r.NifiCluster.Spec.ReadOnlyConfig != (&v1.ReadOnlyConfig{}) && &r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties != (&v1.NifiProperties{}) {
		r.generateReadOnlyConfig(
			&readOnlyClusterConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties.OverrideSecretConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties.OverrideConfigMap,
			r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties.OverrideConfigs, log)
	}

	var readOnlyNodeConfig = map[string]string{}

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.NifiProperties != (&v1.NifiProperties{}) {
			r.generateReadOnlyConfig(
				&readOnlyNodeConfig,
				node.ReadOnlyConfig.NifiProperties.OverrideSecretConfig,
				node.ReadOnlyConfig.NifiProperties.OverrideConfigMap,
				node.ReadOnlyConfig.NifiProperties.OverrideConfigs, log)
			break
		}
	}

	if err := mergo.Merge(&readOnlyNodeConfig, readOnlyClusterConfig); err != nil {
		log.Error("error occurred during merging readonly configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	// Generate the Complete Configuration for the Node
	completeConfigMap := map[string]string{}

	if err := mergo.Merge(&completeConfigMap, readOnlyNodeConfig); err != nil {
		log.Error("error occurred during merging readOnly config to complete configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	if err := mergo.Merge(&completeConfigMap, util.ParsePropertiesFormat(r.getNifiPropertiesConfigString(nodeConfig, id, serverPass, clientPass, superUsers, log))); err != nil {
		log.Error("error occurred during merging operator generated configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	completeConfig := []string{}

	for key, value := range completeConfigMap {
		completeConfig = append(completeConfig, fmt.Sprintf("%s=%s", key, value))
	}

	// We need to sort the config every time to avoid diffs occurred because of ranging through map
	sort.Strings(completeConfig)

	return strings.Join(completeConfig, "\n")
}

func (r *Reconciler) getNifiPropertiesConfigString(nConfig *v1.NodeConfig, id int32, serverPass, clientPass string, superUsers []string, log zap.Logger) string {
	base := r.GetNifiPropertiesBase(id)
	var dnsNames []string
	for _, dnsName := range utilpki.ClusterDNSNames(r.NifiCluster, id) {
		dnsNames = append(dnsNames, fmt.Sprintf("%s:%d", dnsName, GetServerPort(r.NifiCluster.Spec.ListenersConfig)))
	}

	webProxyHosts := strings.Join(dnsNames, ",")
	if len(base.WebProxyHosts) > 0 {
		webProxyHosts = strings.Join(append(dnsNames, base.WebProxyHosts...), ",")
	}

	useSSL := configcommon.UseSSL(r.NifiCluster)
	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.NifiPropertiesTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster": r.NifiCluster,
		"Id":          id,
		"ListenerConfig": config.GenerateListenerSpecificConfig(
			r.NifiCluster.Spec.ListenersConfig,
			id,
			r.NifiCluster.Namespace,
			r.NifiCluster.Name,
			r.NifiCluster.Spec.ListenersConfig.GetClusterDomain(),
			r.NifiCluster.Spec.ListenersConfig.UseExternalDNS,
			r.NifiCluster.Spec.Service.GetServiceTemplate(),
			log),
		"ProvenanceStorage":                  nConfig.GetProvenanceStorage(),
		"SiteToSiteSecure":                   useSSL,
		"ClusterSecure":                      useSSL,
		"WebProxyHosts":                      webProxyHosts,
		"NeedClientAuth":                     base.NeedClientAuth,
		"Authorizer":                         base.GetAuthorizer(),
		"SSLEnabledForInternalCommunication": r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil && util.IsSSLEnabledForInternalCommunication(r.NifiCluster.Spec.ListenersConfig.InternalListeners),
		"SuperUsers":                         strings.Join(generateSuperUsers(superUsers), ";"),
		"ServerKeystorePath":                 serverKeystorePath,
		"ClientKeystorePath":                 clientKeystorePath,
		"KeystoreFile":                       v1.TLSJKSKeyStore,
		"TrustStoreFile":                     v1.TLSJKSTrustStore,
		"ServerKeystorePassword":             serverPass,
		"ClientKeystorePassword":             clientPass,
		//
		"LdapConfiguration":       r.NifiCluster.Spec.LdapConfiguration,
		"SingleUserConfiguration": r.NifiCluster.Spec.SingleUserConfiguration,
		"IsNode":                  nConfig.GetIsNode(),
		"ZookeeperConnectString":  r.NifiCluster.Spec.ZKAddress,
		"ZookeeperPath":           r.NifiCluster.Spec.GetZkPath(),
	}); err != nil {
		log.Error("error occurred during parsing the config template",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}
	return out.String()
}
func generateSuperUsers(users []string) (suStrings []string) {
	suStrings = make([]string, 0)
	for _, x := range users {
		suStrings = append(suStrings, fmt.Sprintf("User:%s", x))
	}
	return
}

/////////////////////////////////////////
//  Zookeeper properties configuration //
/////////////////////////////////////////

func (r Reconciler) generateZookeeperPropertiesNodeConfig(id int32, nodeConfig *v1.NodeConfig, log zap.Logger) string {
	var readOnlyClusterConfig map[string]string

	if &r.NifiCluster.Spec.ReadOnlyConfig != (&v1.ReadOnlyConfig{}) && &r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties != (&v1.ZookeeperProperties{}) {
		r.generateReadOnlyConfig(
			&readOnlyClusterConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties.OverrideSecretConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties.OverrideConfigMap,
			r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties.OverrideConfigs, log)
	}

	var readOnlyNodeConfig = map[string]string{}

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.ZookeeperProperties != (&v1.ZookeeperProperties{}) {
			r.generateReadOnlyConfig(
				&readOnlyNodeConfig,
				node.ReadOnlyConfig.ZookeeperProperties.OverrideSecretConfig,
				node.ReadOnlyConfig.ZookeeperProperties.OverrideConfigMap,
				node.ReadOnlyConfig.ZookeeperProperties.OverrideConfigs, log)
			break
		}
	}

	if err := mergo.Merge(&readOnlyNodeConfig, readOnlyClusterConfig); err != nil {
		log.Error("error occurred during merging readonly configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	// Generate the Complete Configuration for the Node
	completeConfigMap := map[string]string{}

	if err := mergo.Merge(&completeConfigMap, readOnlyNodeConfig); err != nil {
		log.Error("error occurred during merging readOnly config to complete configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	if err := mergo.Merge(&completeConfigMap, util.ParsePropertiesFormat(r.getZookeeperPropertiesConfigString(nodeConfig, id, log))); err != nil {
		log.Error("error occurred during merging operator generated configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	completeConfig := []string{}

	for key, value := range completeConfigMap {
		completeConfig = append(completeConfig, fmt.Sprintf("%s=%s", key, value))
	}

	// We need to sort the config every time to avoid diffs occurred because of ranging through map
	sort.Strings(completeConfig)

	return strings.Join(completeConfig, "\n")
}

func (r *Reconciler) getZookeeperPropertiesConfigString(nConfig *v1.NodeConfig, id int32, log zap.Logger) string {
	base := r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties.DeepCopy()
	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.ZookeeperProperties != (&v1.ZookeeperProperties{}) {
			mergo.Merge(base, node.ReadOnlyConfig.ZookeeperProperties, mergo.WithOverride)
		}
	}

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.ZookeeperPropertiesTemplate))
	if err := t.Execute(&out, map[string]interface{}{}); err != nil {
		log.Error("error occurred during parsing the config template",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}
	return out.String()
}

/////////////////////////////////////
//  State Management configuration //
/////////////////////////////////////

func (r *Reconciler) getStateManagementConfigString(nConfig *v1.NodeConfig, id int32, log zap.Logger) string {
	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.StateManagementTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster":            r.NifiCluster,
		"Id":                     id,
		"ZookeeperConnectString": r.NifiCluster.Spec.ZKAddress,
		"ZookeeperPath":          r.NifiCluster.Spec.GetZkPath(),
	}); err != nil {
		log.Error("error occurred during parsing the config template",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}
	return out.String()
}

/////////////////////////////////////////////
//  Login identity providers configuration //
/////////////////////////////////////////////

func (r *Reconciler) getLoginIdentityProvidersConfigString(nConfig *v1.NodeConfig, id int32, log zap.Logger) string {
	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.LoginIdentityProvidersTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster":             r.NifiCluster,
		"Id":                      id,
		"LdapConfiguration":       r.NifiCluster.Spec.LdapConfiguration,
		"SingleUserConfiguration": r.NifiCluster.Spec.SingleUserConfiguration,
	}); err != nil {
		log.Error("error occurred during parsing the config template",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}
	return out.String()
}

////////////////////////////
//  Logback configuration //
////////////////////////////

func (r *Reconciler) getLogbackConfigString(nConfig *v1.NodeConfig, id int32, log zap.Logger) string {
	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.LogbackConfig != (&v1.LogbackConfig{}) {
			if node.ReadOnlyConfig.LogbackConfig.ReplaceSecretConfig != nil {
				conf, err := r.getSecrectConfig(context.TODO(), *node.ReadOnlyConfig.LogbackConfig.ReplaceSecretConfig)
				if err == nil {
					return conf
				}
				log.Error("error occurred during getting readonly secret config",
					zap.String("clusterName", r.NifiCluster.Name),
					zap.Int32("nodeId", id),
					zap.Error(err))
			}

			if node.ReadOnlyConfig.LogbackConfig.ReplaceConfigMap != nil {
				conf, err := r.getConfigMap(context.TODO(), *node.ReadOnlyConfig.LogbackConfig.ReplaceConfigMap)
				if err == nil {
					return conf
				}
				log.Error("error occurred during getting readonly configmap",
					zap.String("clusterName", r.NifiCluster.Name),
					zap.Int32("nodeId", id),
					zap.Error(err))
			}
			break
		}
	}

	if r.NifiCluster.Spec.ReadOnlyConfig.LogbackConfig.ReplaceSecretConfig != nil {
		conf, err := r.getSecrectConfig(context.TODO(), *r.NifiCluster.Spec.ReadOnlyConfig.LogbackConfig.ReplaceSecretConfig)
		if err == nil {
			return conf
		}
		log.Error("error occurred during getting readonly secret config",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	if r.NifiCluster.Spec.ReadOnlyConfig.LogbackConfig.ReplaceConfigMap != nil {
		conf, err := r.getConfigMap(context.TODO(), *r.NifiCluster.Spec.ReadOnlyConfig.LogbackConfig.ReplaceConfigMap)
		if err == nil {
			return conf
		}
		log.Error("error occurred during getting readonly configmap",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.LogbackTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster": r.NifiCluster,
		"Id":          id,
	}); err != nil {
		log.Error("error occurred during parsing the config template",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}
	return out.String()
}

///////////////////////////////////////////////////
//  Bootstrap notification service configuration //
///////////////////////////////////////////////////

func (r *Reconciler) getBootstrapNotificationServicesConfigString(nConfig *v1.NodeConfig, id int32, log zap.Logger) string {
	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig != (&v1.BootstrapNotificationServicesConfig{}) {
			if node.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceSecretConfig != nil {
				conf, err := r.getSecrectConfig(context.TODO(), *node.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceSecretConfig)
				if err == nil {
					return conf
				}
				log.Error("error occurred during getting bootstrap notification readonly secret config",
					zap.String("clusterName", r.NifiCluster.Name),
					zap.Int32("nodeId", id),
					zap.Error(err))
			}

			if node.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceConfigMap != nil {
				conf, err := r.getConfigMap(context.TODO(), *node.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceConfigMap)
				if err == nil {
					return conf
				}
				log.Error("error occurred during getting bootstrap notification readonly configmap",
					zap.String("clusterName", r.NifiCluster.Name),
					zap.Int32("nodeId", id),
					zap.Error(err))
			}
			break
		}
	}

	if r.NifiCluster.Spec.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceSecretConfig != nil {
		conf, err := r.getSecrectConfig(context.TODO(), *r.NifiCluster.Spec.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceSecretConfig)
		if err == nil {
			return conf
		}
		log.Error("error occurred during getting cluster bootstrap notification readonly secret config",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	if r.NifiCluster.Spec.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceConfigMap != nil {
		conf, err := r.getConfigMap(context.TODO(), *r.NifiCluster.Spec.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceConfigMap)
		if err == nil {
			return conf
		}
		log.Error("error occurred during getting cluster bootstrap notification readonly configmap",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.BootstrapNotificationServicesTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster": r.NifiCluster,
		"Id":          id,
	}); err != nil {
		log.Error("error occurred during parsing the bootstrap notification config template",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}
	return out.String()
}

////////////////////////////////
//  authorizers configuration //
////////////////////////////////

// TODO: Check if cases where is it necessary before using it (seems to be used for secured use cases).
func (r *Reconciler) getAuthorizersConfigString(nConfig *v1.NodeConfig, id int32, log zap.Logger) string {
	nodeList := make(map[string]string)

	authorizersTemplate := config.EmptyAuthorizersTemplate
	if r.NifiCluster.Status.NodesState[fmt.Sprint(id)].InitClusterNode {
		authorizersTemplate = config.AuthorizersTemplate

		// Check for secret/configmap overrides. If there aren't any, then use the default template.
		if r.NifiCluster.Spec.ReadOnlyConfig.AuthorizerConfig.ReplaceTemplateConfigMap != nil {
			conf, err := r.getConfigMap(context.TODO(), *r.NifiCluster.Spec.ReadOnlyConfig.AuthorizerConfig.ReplaceTemplateConfigMap)
			if err != nil {
				log.Error("error occurred during getting authorizer readonly configmap",
					zap.String("clusterName", r.NifiCluster.Name),
					zap.String("configMapName", r.NifiCluster.Spec.ReadOnlyConfig.AuthorizerConfig.ReplaceTemplateConfigMap.Name),
					zap.String("configMapNamespace", r.NifiCluster.Spec.ReadOnlyConfig.AuthorizerConfig.ReplaceTemplateConfigMap.Namespace),
					zap.Int32("nodeId", id),
					zap.Error(err))
			} else {
				authorizersTemplate = conf
			}
		}

		// The secret takes precedence over the ConfigMap, if it exists.
		if r.NifiCluster.Spec.ReadOnlyConfig.AuthorizerConfig.ReplaceTemplateSecretConfig != nil {
			conf, err := r.getSecrectConfig(context.TODO(), *r.NifiCluster.Spec.ReadOnlyConfig.AuthorizerConfig.ReplaceTemplateSecretConfig)
			if err != nil {
				log.Error("error occurred during getting authorizer readonly secret config",
					zap.String("clusterName", r.NifiCluster.Name),
					zap.String("secretName", r.NifiCluster.Spec.ReadOnlyConfig.AuthorizerConfig.ReplaceTemplateSecretConfig.Name),
					zap.String("secretNamespace", r.NifiCluster.Spec.ReadOnlyConfig.AuthorizerConfig.ReplaceTemplateSecretConfig.Namespace),
					zap.Int32("nodeId", id),
					zap.Error(err))
			} else {
				authorizersTemplate = conf
			}
		}

		for nId, nodeState := range r.NifiCluster.Status.NodesState {
			if nodeState.InitClusterNode {
				nodeList[nId] = utilpki.GetNodeUserName(r.NifiCluster, util.ConvertStringToInt32(nId))
			}
		}
	}

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(authorizersTemplate))

	/*nifiControllerName := fmt.Sprintf(
		pkicommon.NodeControllerFQDNTemplate,
		r.NifiCluster.GetNifiControllerUserIdentity(),
		r.NifiCluster.Namespace,
		r.NifiCluster.Spec.ListenersConfig.GetClusterDomain(),
	)

	if r.NifiCluster.Spec.ControllerUserIdentity != nil {
		nifiControllerName = *r.NifiCluster.Spec.ControllerUserIdentity
	}*/

	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster":             r.NifiCluster,
		"Id":                      id,
		"ClusterName":             r.NifiCluster.Name,
		"Namespace":               r.NifiCluster.Namespace,
		"NodeList":                nodeList,
		"ControllerUser":          r.NifiCluster.GetNifiControllerUserIdentity(),
		"SingleUserConfiguration": r.NifiCluster.Spec.SingleUserConfiguration,
	}); err != nil {
		log.Error("error occurred during parsing the config template",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	return out.String()
}

/////////////////////////////////////////
//  Bootstrap properties configuration //
/////////////////////////////////////////

func (r Reconciler) generateBootstrapPropertiesNodeConfig(id int32, nodeConfig *v1.NodeConfig, log zap.Logger) string {
	var readOnlyClusterConfig map[string]string

	if &r.NifiCluster.Spec.ReadOnlyConfig != (&v1.ReadOnlyConfig{}) && &r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties != (&v1.BootstrapProperties{}) {
		r.generateReadOnlyConfig(
			&readOnlyClusterConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties.OverrideSecretConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties.OverrideConfigMap,
			r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties.OverrideConfigs, log)
	}

	var readOnlyNodeConfig = map[string]string{}

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.BootstrapProperties != (&v1.BootstrapProperties{}) {
			r.generateReadOnlyConfig(
				&readOnlyNodeConfig,
				node.ReadOnlyConfig.BootstrapProperties.OverrideSecretConfig,
				node.ReadOnlyConfig.BootstrapProperties.OverrideConfigMap,
				node.ReadOnlyConfig.BootstrapProperties.OverrideConfigs, log)
			break
		}
	}

	if err := mergo.Merge(&readOnlyNodeConfig, readOnlyClusterConfig); err != nil {
		log.Error("error occurred during merging readonly configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	// Generate the Complete Configuration for the Node
	completeConfigMap := map[string]string{}

	if err := mergo.Merge(&completeConfigMap, readOnlyNodeConfig); err != nil {
		log.Error("error occurred during merging readOnly config to complete configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	if err := mergo.Merge(&completeConfigMap, util.ParsePropertiesFormat(r.getBootstrapPropertiesConfigString(nodeConfig, id, log))); err != nil {
		log.Error("error occurred during merging operator generated configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}

	completeConfig := []string{}

	for key, value := range completeConfigMap {
		completeConfig = append(completeConfig, fmt.Sprintf("%s=%s", key, value))
	}

	// We need to sort the config every time to avoid diffs occurred because of ranging through map
	sort.Strings(completeConfig)

	return strings.Join(completeConfig, "\n")
}

func (r *Reconciler) getBootstrapPropertiesConfigString(nConfig *v1.NodeConfig, id int32, log zap.Logger) string {
	base := r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties.DeepCopy()
	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.BootstrapProperties != (&v1.BootstrapProperties{}) {
			mergo.Merge(base, node.ReadOnlyConfig.BootstrapProperties, mergo.WithOverride)
		}
	}

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.BootstrapPropertiesTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster": r.NifiCluster,
		"Id":          id,
		"JvmMemory":   base.GetNifiJvmMemory(),
	}); err != nil {
		log.Error("error occurred during parsing the config template",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Int32("nodeId", id),
			zap.Error(err))
	}
	return out.String()
}

func (r *Reconciler) GetNifiPropertiesBase(id int32) *v1.NifiProperties {
	base := r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties.DeepCopy()
	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.NifiProperties != (&v1.NifiProperties{}) {
			mergo.Merge(base, node.ReadOnlyConfig.NifiProperties, mergo.WithOverride)
		}
	}

	return base
}

func (r Reconciler) getSecrectConfig(ctx context.Context, ref v1.SecretConfigReference) (conf string, err error) {
	secret := &corev1.Secret{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: ref.Namespace}, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return conf, errorfactory.New(errorfactory.ResourceNotReady{}, err, "config secret not ready")
		}
		return conf, errorfactory.New(errorfactory.APIFailure{}, err, "failed to get config secret")
	}
	conf = string(secret.Data[ref.Data])

	return conf, nil
}

func (r Reconciler) getConfigMap(ctx context.Context, ref v1.ConfigmapReference) (conf string, err error) {
	configmap := &corev1.ConfigMap{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: ref.Namespace}, configmap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return conf, errorfactory.New(errorfactory.ResourceNotReady{}, err, "configmap not ready")
		}
		return conf, errorfactory.New(errorfactory.APIFailure{}, err, "failed to get configmap")
	}
	conf = configmap.Data[ref.Data]

	return conf, nil
}

func (r Reconciler) generateReadOnlyConfig(
	readOnlyClusterConfig *map[string]string,
	overrideSecretConfig *v1.SecretConfigReference,
	overrideConfigMap *v1.ConfigmapReference,
	overrideConfigs string,
	log zap.Logger) {
	var parsedReadOnlySecretClusterConfig map[string]string
	var parsedReadOnlyClusterConfig map[string]string
	var parsedReadOnlyClusterConfigMap map[string]string

	if overrideSecretConfig != nil {
		secretConfig, err := r.getSecrectConfig(context.TODO(), *overrideSecretConfig)
		if err != nil {
			log.Error("error occurred during getting readonly secret config",
				zap.String("clusterName", r.NifiCluster.Name),
				zap.Error(err))
		}
		parsedReadOnlySecretClusterConfig = util.ParsePropertiesFormat(secretConfig)
	}

	if overrideConfigMap != nil {
		configMap, err := r.getConfigMap(context.TODO(), *overrideConfigMap)
		if err != nil {
			log.Error("error occurred during getting readonly configmap",
				zap.String("clusterName", r.NifiCluster.Name),
				zap.Error(err))
		}
		parsedReadOnlyClusterConfigMap = util.ParsePropertiesFormat(configMap)
	}

	parsedReadOnlyClusterConfig = util.ParsePropertiesFormat(overrideConfigs)

	if err := mergo.Merge(readOnlyClusterConfig, parsedReadOnlySecretClusterConfig); err != nil {
		log.Error("error occurred during merging readonly configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Error(err))
	}

	if err := mergo.Merge(readOnlyClusterConfig, parsedReadOnlyClusterConfig); err != nil {
		log.Error("error occurred during merging readonly configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Error(err))
	}

	if err := mergo.Merge(readOnlyClusterConfig, parsedReadOnlyClusterConfigMap); err != nil {
		log.Error("error occurred during merging readonly configs",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.Error(err))
	}
}
