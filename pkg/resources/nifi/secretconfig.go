// Copyright 2020 Orange SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package apis

package nifi

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	configcommon "github.com/Orange-OpenSource/nifikop/pkg/nificlient/config/common"
	nifiutil "github.com/Orange-OpenSource/nifikop/pkg/util/nifi"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"
	"text/template"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/resources/templates"
	"github.com/Orange-OpenSource/nifikop/pkg/resources/templates/config"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	pkicommon "github.com/Orange-OpenSource/nifikop/pkg/util/pki"
	utilpki "github.com/Orange-OpenSource/nifikop/pkg/util/pki"
	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
)

//func encodeBase64(toEncode string) []byte {
//	return []byte(base64.StdEncoding.EncodeToString([]byte(toEncode)))
//}
func (r *Reconciler) secretConfig(id int32, nodeConfig *v1alpha1.NodeConfig, serverPass, clientPass string, superUsers []string, log logr.Logger) runtimeClient.Object {
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

//
func (r Reconciler) generateNifiPropertiesNodeConfig(id int32, nodeConfig *v1alpha1.NodeConfig, serverPass, clientPass string, superUsers []string, log logr.Logger) string {
	var readOnlyClusterConfig map[string]string
	if &r.NifiCluster.Spec.ReadOnlyConfig != nil && &r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties != nil {
		r.generateReadOnlyConfig(
			&readOnlyClusterConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties.OverrideSecretConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties.OverrideConfigMap,
			r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties.OverrideConfigs, log)
	}

	var readOnlyNodeConfig = map[string]string{}

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.NifiProperties != nil {
			r.generateReadOnlyConfig(
				&readOnlyNodeConfig,
				node.ReadOnlyConfig.NifiProperties.OverrideSecretConfig,
				node.ReadOnlyConfig.NifiProperties.OverrideConfigMap,
				node.ReadOnlyConfig.NifiProperties.OverrideConfigs, log)
			break
		}
	}

	if err := mergo.Merge(&readOnlyNodeConfig, readOnlyClusterConfig); err != nil {
		log.Error(err, "error occurred during merging readonly configs")
	}

	//Generate the Complete Configuration for the Node
	completeConfigMap := map[string]string{}

	if err := mergo.Merge(&completeConfigMap, readOnlyNodeConfig); err != nil {
		log.Error(err, "error occurred during merging readOnly config to complete configs")
	}

	if err := mergo.Merge(&completeConfigMap, util.ParsePropertiesFormat(r.getNifiPropertiesConfigString(nodeConfig, id, serverPass, clientPass, superUsers, log))); err != nil {
		log.Error(err, "error occurred during merging operator generated configs")
	}

	completeConfig := []string{}

	for key, value := range completeConfigMap {
		completeConfig = append(completeConfig, fmt.Sprintf("%s=%s", key, value))
	}

	// We need to sort the config every time to avoid diffs occurred because of ranging through map
	sort.Strings(completeConfig)

	return strings.Join(completeConfig, "\n")
}

//
func (r *Reconciler) getNifiPropertiesConfigString(nConfig *v1alpha1.NodeConfig, id int32, serverPass, clientPass string, superUsers []string, log logr.Logger) string {

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
			r.NifiCluster.Spec.Service.HeadlessEnabled,
			r.NifiCluster.Spec.ListenersConfig.GetClusterDomain(),
			r.NifiCluster.Spec.ListenersConfig.UseExternalDNS,
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
		"KeystoreFile":                       v1alpha1.TLSJKSKeyStore,
		"TrustStoreFile":                     v1alpha1.TLSJKSTrustStore,
		"ServerKeystorePassword":             serverPass,
		"ClientKeystorePassword":             clientPass,
		//
		"LdapConfiguration":      r.NifiCluster.Spec.LdapConfiguration,
		"IsNode":                 nConfig.GetIsNode(),
		"ZookeeperConnectString": r.NifiCluster.Spec.ZKAddress,
		"ZookeeperPath":          r.NifiCluster.Spec.GetZkPath(),
	}); err != nil {
		log.Error(err, "error occurred during parsing the config template")
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

//
func (r Reconciler) generateZookeeperPropertiesNodeConfig(id int32, nodeConfig *v1alpha1.NodeConfig, log logr.Logger) string {
	var readOnlyClusterConfig map[string]string

	if &r.NifiCluster.Spec.ReadOnlyConfig != nil && &r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties != nil {
		r.generateReadOnlyConfig(
			&readOnlyClusterConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties.OverrideSecretConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties.OverrideConfigMap,
			r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties.OverrideConfigs, log)
	}

	var readOnlyNodeConfig = map[string]string{}

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.ZookeeperProperties != nil {
			r.generateReadOnlyConfig(
				&readOnlyNodeConfig,
				node.ReadOnlyConfig.ZookeeperProperties.OverrideSecretConfig,
				node.ReadOnlyConfig.ZookeeperProperties.OverrideConfigMap,
				node.ReadOnlyConfig.ZookeeperProperties.OverrideConfigs, log)
			break
		}
	}

	if err := mergo.Merge(&readOnlyNodeConfig, readOnlyClusterConfig); err != nil {
		log.Error(err, "error occurred during merging readonly configs")
	}

	//Generate the Complete Configuration for the Node
	completeConfigMap := map[string]string{}

	if err := mergo.Merge(&completeConfigMap, readOnlyNodeConfig); err != nil {
		log.Error(err, "error occurred during merging readOnly config to complete configs")
	}

	if err := mergo.Merge(&completeConfigMap, util.ParsePropertiesFormat(r.getZookeeperPropertiesConfigString(nodeConfig, id, log))); err != nil {
		log.Error(err, "error occurred during merging operator generated configs")
	}

	completeConfig := []string{}

	for key, value := range completeConfigMap {
		completeConfig = append(completeConfig, fmt.Sprintf("%s=%s", key, value))
	}

	// We need to sort the config every time to avoid diffs occurred because of ranging through map
	sort.Strings(completeConfig)

	return strings.Join(completeConfig, "\n")
}

//
func (r *Reconciler) getZookeeperPropertiesConfigString(nConfig *v1alpha1.NodeConfig, id int32, log logr.Logger) string {

	base := r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties.DeepCopy()
	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.ZookeeperProperties != nil {
			mergo.Merge(base, node.ReadOnlyConfig.ZookeeperProperties, mergo.WithOverride)
		}
	}

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.ZookeeperPropertiesTemplate))
	if err := t.Execute(&out, map[string]interface{}{}); err != nil {
		log.Error(err, "error occurred during parsing the config template")
	}
	return out.String()
}

/////////////////////////////////////
//  State Management configuration //
/////////////////////////////////////

//
func (r *Reconciler) getStateManagementConfigString(nConfig *v1alpha1.NodeConfig, id int32, log logr.Logger) string {

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.StateManagementTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster":            r.NifiCluster,
		"Id":                     id,
		"ZookeeperConnectString": r.NifiCluster.Spec.ZKAddress,
		"ZookeeperPath":          r.NifiCluster.Spec.GetZkPath(),
	}); err != nil {
		log.Error(err, "error occurred during parsing the config template")
	}
	return out.String()
}

/////////////////////////////////////////////
//  Login identity providers configuration //
/////////////////////////////////////////////

//
func (r *Reconciler) getLoginIdentityProvidersConfigString(nConfig *v1alpha1.NodeConfig, id int32, log logr.Logger) string {

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.LoginIdentityProvidersTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster":       r.NifiCluster,
		"Id":                id,
		"LdapConfiguration": r.NifiCluster.Spec.LdapConfiguration,
	}); err != nil {
		log.Error(err, "error occurred during parsing the config template")
	}
	return out.String()
}

////////////////////////////
//  Logback configuration //
////////////////////////////

//
func (r *Reconciler) getLogbackConfigString(nConfig *v1alpha1.NodeConfig, id int32, log logr.Logger) string {

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.LogbackConfig != nil {
			if node.ReadOnlyConfig.LogbackConfig.ReplaceSecretConfig != nil {
				conf, err := r.getSecrectConfig(context.TODO(), *node.ReadOnlyConfig.LogbackConfig.ReplaceSecretConfig)
				if err == nil {
					return conf
				}
				log.Error(err, "error occurred during getting readonly secret config")
			}

			if node.ReadOnlyConfig.LogbackConfig.ReplaceConfigMap != nil {
				conf, err := r.getConfigMap(context.TODO(), *node.ReadOnlyConfig.LogbackConfig.ReplaceConfigMap)
				if err == nil {
					return conf
				}
				log.Error(err, "error occurred during getting readonly configmap")
			}
			break
		}
	}

	if r.NifiCluster.Spec.ReadOnlyConfig.LogbackConfig.ReplaceSecretConfig != nil {
		conf, err := r.getSecrectConfig(context.TODO(), *r.NifiCluster.Spec.ReadOnlyConfig.LogbackConfig.ReplaceSecretConfig)
		if err == nil {
			return conf
		}
		log.Error(err, "error occurred during getting readonly secret config")
	}

	if r.NifiCluster.Spec.ReadOnlyConfig.LogbackConfig.ReplaceConfigMap != nil {
		conf, err := r.getConfigMap(context.TODO(), *r.NifiCluster.Spec.ReadOnlyConfig.LogbackConfig.ReplaceConfigMap)
		if err == nil {
			return conf
		}
		log.Error(err, "error occurred during getting readonly configmap")
	}

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.LogbackTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster": r.NifiCluster,
		"Id":          id,
	}); err != nil {
		log.Error(err, "error occurred during parsing the config template")
	}
	return out.String()
}

///////////////////////////////////////////////////
//  Bootstrap notification service configuration //
///////////////////////////////////////////////////

//
func (r *Reconciler) getBootstrapNotificationServicesConfigString(nConfig *v1alpha1.NodeConfig, id int32, log logr.Logger) string {

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig != nil {
			if node.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceSecretConfig != nil {
				conf, err := r.getSecrectConfig(context.TODO(), *node.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceSecretConfig)
				if err == nil {
					return conf
				}
				log.Error(err, "error occurred during getting readonly secret config")
			}

			if node.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceConfigMap != nil {
				conf, err := r.getConfigMap(context.TODO(), *node.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceConfigMap)
				if err == nil {
					return conf
				}
				log.Error(err, "error occurred during getting readonly configmap")
			}
			break
		}
	}

	if r.NifiCluster.Spec.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceSecretConfig != nil {
		conf, err := r.getSecrectConfig(context.TODO(), *r.NifiCluster.Spec.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceSecretConfig)
		if err == nil {
			return conf
		}
		log.Error(err, "error occurred during getting readonly secret config")
	}

	if r.NifiCluster.Spec.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceConfigMap != nil {
		conf, err := r.getConfigMap(context.TODO(), *r.NifiCluster.Spec.ReadOnlyConfig.BootstrapNotificationServicesReplaceConfig.ReplaceConfigMap)
		if err == nil {
			return conf
		}
		log.Error(err, "error occurred during getting readonly configmap")
	}

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.BootstrapNotificationServicesTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster": r.NifiCluster,
		"Id":          id,
	}); err != nil {
		log.Error(err, "error occurred during parsing the config template")
	}
	return out.String()
}

////////////////////////////////
//  authorizers configuration //
////////////////////////////////

// TODO: Check if cases where is it necessary before using it (seems to be used for secured use cases)
func (r *Reconciler) getAuthorizersConfigString(nConfig *v1alpha1.NodeConfig, id int32, log logr.Logger) string {

	nodeList := make(map[string]string)

	authorizersTemplate := config.EmptyAuthorizersTemplate
	if r.NifiCluster.Status.NodesState[fmt.Sprint(id)].InitClusterNode {
		authorizersTemplate = config.AuthorizersTemplate
		for nId, nodeState := range r.NifiCluster.Status.NodesState {
			if nodeState.InitClusterNode {
				nodeList[nId] = utilpki.GetNodeUserName(r.NifiCluster, util.ConvertStringToInt32(nId))
			}
		}
	}

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(authorizersTemplate))

	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster": r.NifiCluster,
		"Id":          id,
		"ClusterName": r.NifiCluster.Name,
		"Namespace":   r.NifiCluster.Namespace,
		"NodeList":    nodeList,
		"ControllerUser": fmt.Sprintf(pkicommon.NodeControllerFQDNTemplate,
			fmt.Sprintf(pkicommon.NodeControllerTemplate, r.NifiCluster.Name),
			r.NifiCluster.Namespace,
			r.NifiCluster.Spec.ListenersConfig.GetClusterDomain()),
	}); err != nil {
		log.Error(err, "error occurred during parsing the config template")
	}

	return out.String()
}

/////////////////////////////////////////
//  Bootstrap properties configuration //
/////////////////////////////////////////

//
func (r Reconciler) generateBootstrapPropertiesNodeConfig(id int32, nodeConfig *v1alpha1.NodeConfig, log logr.Logger) string {
	var readOnlyClusterConfig map[string]string

	if &r.NifiCluster.Spec.ReadOnlyConfig != nil && &r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties != nil {
		r.generateReadOnlyConfig(
			&readOnlyClusterConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties.OverrideSecretConfig,
			r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties.OverrideConfigMap,
			r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties.OverrideConfigs, log)
	}

	var readOnlyNodeConfig = map[string]string{}

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.BootstrapProperties != nil {
			r.generateReadOnlyConfig(
				&readOnlyNodeConfig,
				node.ReadOnlyConfig.BootstrapProperties.OverrideSecretConfig,
				node.ReadOnlyConfig.BootstrapProperties.OverrideConfigMap,
				node.ReadOnlyConfig.BootstrapProperties.OverrideConfigs, log)
			break
		}
	}

	if err := mergo.Merge(&readOnlyNodeConfig, readOnlyClusterConfig); err != nil {
		log.Error(err, "error occurred during merging readonly configs")
	}

	//Generate the Complete Configuration for the Node
	completeConfigMap := map[string]string{}

	if err := mergo.Merge(&completeConfigMap, readOnlyNodeConfig); err != nil {
		log.Error(err, "error occurred during merging readOnly config to complete configs")
	}

	if err := mergo.Merge(&completeConfigMap, util.ParsePropertiesFormat(r.getBootstrapPropertiesConfigString(nodeConfig, id, log))); err != nil {
		log.Error(err, "error occurred during merging operator generated configs")
	}

	completeConfig := []string{}

	for key, value := range completeConfigMap {
		completeConfig = append(completeConfig, fmt.Sprintf("%s=%s", key, value))
	}

	// We need to sort the config every time to avoid diffs occurred because of ranging through map
	sort.Strings(completeConfig)

	return strings.Join(completeConfig, "\n")
}

//
func (r *Reconciler) getBootstrapPropertiesConfigString(nConfig *v1alpha1.NodeConfig, id int32, log logr.Logger) string {
	base := r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties.DeepCopy()
	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.BootstrapProperties != nil {
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
		log.Error(err, "error occurred during parsing the config template")
	}
	return out.String()
}

func (r *Reconciler) GetNifiPropertiesBase(id int32) *v1alpha1.NifiProperties {
	base := r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties.DeepCopy()
	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.NifiProperties != nil {
			mergo.Merge(base, node.ReadOnlyConfig.NifiProperties, mergo.WithOverride)
		}
	}

	return base
}

func (r Reconciler) getSecrectConfig(ctx context.Context, ref v1alpha1.SecretConfigReference) (conf string, err error) {
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

func (r Reconciler) getConfigMap(ctx context.Context, ref v1alpha1.ConfigmapReference) (conf string, err error) {
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
	overrideSecretConfig *v1alpha1.SecretConfigReference,
	overrideConfigMap *v1alpha1.ConfigmapReference,
	overrideConfigs string,
	log logr.Logger) {

	var parsedReadOnlySecretClusterConfig map[string]string
	var parsedReadOnlyClusterConfig map[string]string
	var parsedReadOnlyClusterConfigMap map[string]string

	if overrideSecretConfig != nil {
		secretConfig, err := r.getSecrectConfig(context.TODO(), *overrideSecretConfig)
		if err != nil {
			log.Error(err, "error occurred during getting readonly secret config")
		}
		parsedReadOnlySecretClusterConfig = util.ParsePropertiesFormat(secretConfig)
	}

	if overrideConfigMap != nil {
		configMap, err := r.getConfigMap(context.TODO(), *overrideConfigMap)
		if err != nil {
			log.Error(err, "error occurred during getting readonly configmap")
		}
		parsedReadOnlyClusterConfigMap = util.ParsePropertiesFormat(configMap)
	}

	parsedReadOnlyClusterConfig = util.ParsePropertiesFormat(overrideConfigs)

	if err := mergo.Merge(readOnlyClusterConfig, parsedReadOnlySecretClusterConfig); err != nil {
		log.Error(err, "error occurred during merging readonly configs")
	}

	if err := mergo.Merge(readOnlyClusterConfig, parsedReadOnlyClusterConfig); err != nil {
		log.Error(err, "error occurred during merging readonly configs")
	}

	if err := mergo.Merge(readOnlyClusterConfig, parsedReadOnlyClusterConfigMap); err != nil {
		log.Error(err, "error occurred during merging readonly configs")
	}
}
