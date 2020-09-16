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
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	"github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/resources/templates"
	"github.com/Orange-OpenSource/nifikop/pkg/resources/templates/config"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	utilpki "github.com/Orange-OpenSource/nifikop/pkg/util/pki"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) configMap(id int32, nodeConfig *v1alpha1.NodeConfig, serverPass, clientPass string, superUsers []string, log logr.Logger) runtime.Object {
	configMap := &corev1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(
			fmt.Sprintf(templates.NodeConfigTemplate+"-%d", r.NifiCluster.Name, id),
			util.MergeLabels(
				LabelsForNifi(r.NifiCluster.Name),
				map[string]string{"nodeId": fmt.Sprintf("%d", id)},
			),
			r.NifiCluster,
		),
		Data: map[string]string{
			"nifi.properties":                    r.generateNifiPropertiesNodeConfig(id, nodeConfig, serverPass, clientPass, superUsers, log),
			"zookeeper.properties":               r.generateZookeeperPropertiesNodeConfig(id, nodeConfig, log),
			"state-management.xml":               r.getStateManagementConfigString(nodeConfig, id, log),
			"login-identity-providers.xml":       r.getLoginIdentityProvidersConfigString(nodeConfig, id, log),
			"logback.xml":                        r.getLogbackConfigString(nodeConfig, id, log),
			"bootstrap.conf":                     r.generateBootstrapPropertiesNodeConfig(id, nodeConfig, log),
			"bootstrap-notification-servces.xml": r.getBootstrapNotificationServicesConfigString(nodeConfig, id, log),
		},
	}

	if r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil && r.NifiCluster.Spec.ClusterSecure && r.NifiCluster.Spec.SiteToSiteSecure {
		configMap.Data["authorizers.xml"] = r.getAuthorizersConfigString(nodeConfig, id, log)
	}
	return configMap
}

////////////////////////////////////
//  Nifi properties configuration //
////////////////////////////////////

//
func (r Reconciler) generateNifiPropertiesNodeConfig(id int32, nodeConfig *v1alpha1.NodeConfig, serverPass, clientPass string, superUsers []string, log logr.Logger) string {
	var parsedReadOnlyClusterConfig map[string]string

	if &r.NifiCluster.Spec.ReadOnlyConfig != nil && &r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties != nil {
		parsedReadOnlyClusterConfig = util.ParsePropertiesFormat(r.NifiCluster.Spec.ReadOnlyConfig.NifiProperties.OverrideConfigs)
	}

	var parsedReadOnlyNodeConfig = map[string]string{}

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.NifiProperties != nil {
			parsedReadOnlyNodeConfig = util.ParsePropertiesFormat(node.ReadOnlyConfig.NifiProperties.OverrideConfigs)
			break
		}
	}

	if err := mergo.Merge(&parsedReadOnlyNodeConfig, parsedReadOnlyClusterConfig); err != nil {
		log.Error(err, "error occurred during merging readonly configs")
	}

	//Generate the Complete Configuration for the Node
	completeConfigMap := map[string]string{}

	if err := mergo.Merge(&completeConfigMap, util.ParsePropertiesFormat(r.getNifiPropertiesConfigString(nodeConfig, id, serverPass, clientPass, superUsers, log))); err != nil {
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

//
func (r *Reconciler) getNifiPropertiesConfigString(nConfig *v1alpha1.NodeConfig, id int32, serverPass, clientPass string, superUsers []string, log logr.Logger) string {

	base := r.GetNifiPropertiesBase(id)
	var dnsNames []string
	for _, dnsName := range utilpki.ClusterDNSNames(r.NifiCluster, id) {
		dnsNames = append(dnsNames, fmt.Sprintf("%s:%d", dnsName, GetServerPort(&r.NifiCluster.Spec.ListenersConfig)))
	}

	webProxyHosts := strings.Join(dnsNames, ",")
	if len(base.WebProxyHosts) > 0 {
		webProxyHosts = strings.Join(append(dnsNames, base.WebProxyHosts...), ",")
	}

	var out bytes.Buffer
	t := template.Must(template.New("nConfig-config").Parse(config.NifiPropertiesTemplate))
	if err := t.Execute(&out, map[string]interface{}{
		"NifiCluster": r.NifiCluster,
		"Id":          id,
		"ListenerConfig": config.GenerateListenerSpecificConfig(
			&r.NifiCluster.Spec.ListenersConfig,
			id,
			r.NifiCluster.Namespace,
			r.NifiCluster.Name,
			r.NifiCluster.Spec.Service.HeadlessEnabled,
			r.NifiCluster.Spec.ListenersConfig.GetClusterDomain(),
			r.NifiCluster.Spec.ListenersConfig.UseExternalDNS,
			log),
		"ProvenanceStorage":                  nConfig.GetProvenanceStorage(),
		"SiteToSiteSecure":                   r.NifiCluster.Spec.SiteToSiteSecure,
		"ClusterSecure":                      r.NifiCluster.Spec.ClusterSecure,
		"WebProxyHosts":                      webProxyHosts,
		"NeedClientAuth":                     base.NeedClientAuth,
		"Authorizer":                         base.GetAuthorizer(),
		"SSLEnabledForInternalCommunication": r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil && util.IsSSLEnabledForInternalCommunication(r.NifiCluster.Spec.ListenersConfig.InternalListeners),
		"SuperUsers":                         strings.Join(generateSuperUsers(superUsers), ";"),
		"ServerKeystorePath":                 serverKeystorePath,
		"ClientKeystorePath":                 clientKeystorePath,
		"KeystoreFile":                       v1alpha1.TLSJKSKey,
		"ServerKeystorePassword":             serverPass,
		"ClientKeystorePassword":             clientPass,
		//
		"LdapConfiguration":      r.NifiCluster.Spec.LdapConfiguration,
		"IsNode":                 nConfig.GetIsNode(),
		"ZookeeperConnectString": r.NifiCluster.Spec.ZKAddresse,
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
	var parsedReadOnlyClusterConfig map[string]string

	if &r.NifiCluster.Spec.ReadOnlyConfig != nil && &r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties != nil {
		parsedReadOnlyClusterConfig = util.ParsePropertiesFormat(r.NifiCluster.Spec.ReadOnlyConfig.ZookeeperProperties.OverrideConfigs)
	}

	var parsedReadOnlyNodeConfig = map[string]string{}

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.ZookeeperProperties != nil {
			parsedReadOnlyNodeConfig = util.ParsePropertiesFormat(node.ReadOnlyConfig.ZookeeperProperties.OverrideConfigs)
			break
		}
	}

	if err := mergo.Merge(&parsedReadOnlyNodeConfig, parsedReadOnlyClusterConfig); err != nil {
		log.Error(err, "error occurred during merging readonly configs")
	}

	//Generate the Complete Configuration for the Node
	completeConfigMap := map[string]string{}

	if err := mergo.Merge(&completeConfigMap, util.ParsePropertiesFormat(r.getZookeeperPropertiesConfigString(nodeConfig, id, log))); err != nil {
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
	if err := t.Execute(&out, map[string]interface{}{
	}); err != nil {
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
		"ZookeeperConnectString": r.NifiCluster.Spec.ZKAddresse,
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
		"NifiCluster":      r.NifiCluster,
		"Id":               id,
		"ClusterName":      r.NifiCluster.Name,
		"Namespace":        r.NifiCluster.Namespace,
		"NodeList":         nodeList,
		"InitialAdminUser": r.NifiCluster.Spec.InitialAdminUser,
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
	var parsedReadOnlyClusterConfig map[string]string

	if &r.NifiCluster.Spec.ReadOnlyConfig != nil && &r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties != nil {
		parsedReadOnlyClusterConfig = util.ParsePropertiesFormat(r.NifiCluster.Spec.ReadOnlyConfig.BootstrapProperties.OverrideConfigs)
	}

	var parsedReadOnlyNodeConfig = map[string]string{}

	for _, node := range r.NifiCluster.Spec.Nodes {
		if node.Id == id && node.ReadOnlyConfig != nil && &node.ReadOnlyConfig.BootstrapProperties != nil {
			parsedReadOnlyNodeConfig = util.ParsePropertiesFormat(node.ReadOnlyConfig.BootstrapProperties.OverrideConfigs)
			break
		}
	}

	if err := mergo.Merge(&parsedReadOnlyNodeConfig, parsedReadOnlyClusterConfig); err != nil {
		log.Error(err, "error occurred during merging readonly configs")
	}

	//Generate the Complete Configuration for the Node
	completeConfigMap := map[string]string{}

	if err := mergo.Merge(&completeConfigMap, util.ParsePropertiesFormat(r.getBootstrapPropertiesConfigString(nodeConfig, id, log))); err != nil {
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
