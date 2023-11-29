package v1alpha1

import (
	"reflect"
	"testing"

	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
)

func TestConvertNifiCluster(t *testing.T) {
	alphaNC := createNifiCluster()

	nc := &v1.NifiCluster{}

	// convert v1alpha1 to v1
	alphaNC.ConvertTo(nc)
	assertNifiClustersEqual(alphaNC, nc, t)

	// convert v1 to v1alpha1
	newCluster := &NifiCluster{}
	newCluster.ConvertFrom(nc)
	assertNifiClustersEqual(newCluster, nc, t)
}

func assertNifiClustersEqual(anc *NifiCluster, nc *v1.NifiCluster, t *testing.T) {
	if !reflect.DeepEqual(anc.ObjectMeta, nc.ObjectMeta) {
		t.Error("Object metas not equal")
	}
	if string(anc.Spec.ClientType) != string(nc.Spec.ClientType) {
		t.Error("client types not equal")
	}
	if string(anc.Spec.Type) != string(nc.Spec.Type) {
		t.Error("cluster types not equal")
	}
	if anc.Spec.NodeURITemplate != nc.Spec.NodeURITemplate {
		t.Error("node URI templates not equal")
	}
	if anc.Spec.NifiURI != nc.Spec.NifiURI {
		t.Error("nifi URIs not equal")
	}
	if anc.Spec.RootProcessGroupId != nc.Spec.RootProcessGroupId {
		t.Error("root process group IDs not equal")
	}
	if anc.Spec.ProxyUrl != nc.Spec.ProxyUrl {
		t.Error("root process group IDs not equal")
	}
	if anc.Spec.ZKAddress != nc.Spec.ZKAddress {
		t.Error("ZK Addresses not equal")
	}
	if anc.Spec.ZKPath != nc.Spec.ZKPath {
		t.Error("ZK Paths not equal")
	}
	if anc.Spec.InitContainerImage != nc.Spec.InitContainerImage {
		t.Error("Init container images not equal")
	}
	if anc.Spec.ClusterImage != nc.Spec.ClusterImage {
		t.Error("Cluster images not equal")
	}
	if anc.Spec.OneNifiNodePerNode != nc.Spec.OneNifiNodePerNode {
		t.Error("one nifi node per nodes not equal")
	}
	if anc.Spec.PropagateLabels != nc.Spec.PropagateLabels {
		t.Error("Propagate labels not equal")
	}
	if anc.Spec.SecretRef.Name != nc.Spec.SecretRef.Name ||
		anc.Spec.SecretRef.Namespace != nc.Spec.SecretRef.Namespace {
		t.Error("secret refs not equal")
	}
	if anc.Spec.Service.HeadlessEnabled != nc.Spec.Service.HeadlessEnabled ||
		!reflect.DeepEqual(anc.Spec.Service.Annotations, nc.Spec.Service.Annotations) ||
		!reflect.DeepEqual(anc.Spec.Service.Labels, nc.Spec.Service.Labels) ||
		anc.Spec.Service.ServiceTemplate != nc.Spec.Service.ServiceTemplate {
		t.Error("service policies not equal")
	}
	if !reflect.DeepEqual(anc.Spec.Pod.Annotations, nc.Spec.Pod.Annotations) ||
		!reflect.DeepEqual(anc.Spec.Pod.Labels, nc.Spec.Pod.Labels) ||
		!reflect.DeepEqual(anc.Spec.Pod.HostAliases, nc.Spec.Pod.HostAliases) {
		t.Error("pod policies not equal")
	}
	if !reflect.DeepEqual(anc.Spec.InitContainers, nc.Spec.InitContainers) {
		t.Error("init containers not equal")
	}
	if !managedUsersEqual(anc.Spec.ManagedAdminUsers, nc.Spec.ManagedAdminUsers) {
		t.Errorf("Managed admin users not equal. %+v vs %+v", anc.Spec.ManagedAdminUsers, nc.Spec.ManagedAdminUsers)
	}
	if !managedUsersEqual(anc.Spec.ManagedReaderUsers, nc.Spec.ManagedReaderUsers) {
		t.Errorf("Managed reader users not equal. %+v vs %+v", anc.Spec.ManagedReaderUsers, nc.Spec.ManagedReaderUsers)
	}
	assertReadOnlyConfigsEqual(nc.Spec.ReadOnlyConfig, anc.Spec.ReadOnlyConfig, t)
	assertNodeConfigGroupsEqual(anc.Spec.NodeConfigGroups, nc.Spec.NodeConfigGroups, t)
	if anc.Spec.NodeUserIdentityTemplate != nc.Spec.NodeUserIdentityTemplate {
		t.Error("node user identity templates not equal")
	}
	if !nodesEqual(anc.Spec.Nodes, nc.Spec.Nodes, t) {
		t.Error("nodes are not equal")
	}
	if anc.Spec.DisruptionBudget.Budget != nc.Spec.DisruptionBudget.Budget ||
		anc.Spec.DisruptionBudget.Create != nc.Spec.DisruptionBudget.Create {
		t.Error("disruption budgets are not equal")
	}
	if anc.Spec.LdapConfiguration.Enabled != nc.Spec.LdapConfiguration.Enabled ||
		anc.Spec.LdapConfiguration.SearchBase != nc.Spec.LdapConfiguration.SearchBase ||
		anc.Spec.LdapConfiguration.SearchFilter != nc.Spec.LdapConfiguration.SearchFilter ||
		anc.Spec.LdapConfiguration.Url != nc.Spec.LdapConfiguration.Url {
		t.Error("LDAP configurations are not equal")
	}
	if anc.Spec.NifiClusterTaskSpec.RetryDurationMinutes != nc.Spec.NifiClusterTaskSpec.RetryDurationMinutes {
		t.Error("cluster task specs are not equal")
	}
	if anc.Spec.ListenersConfig.ClusterDomain != nc.Spec.ListenersConfig.ClusterDomain ||
		!internalListenersConfigsEqual(anc.Spec.ListenersConfig.InternalListeners, nc.Spec.ListenersConfig.InternalListeners) ||
		anc.Spec.ListenersConfig.SSLSecrets.ClusterScoped != nc.Spec.ListenersConfig.SSLSecrets.ClusterScoped ||
		anc.Spec.ListenersConfig.SSLSecrets.Create != nc.Spec.ListenersConfig.SSLSecrets.Create ||
		!reflect.DeepEqual(anc.Spec.ListenersConfig.SSLSecrets.IssuerRef, nc.Spec.ListenersConfig.SSLSecrets.IssuerRef) ||
		string(anc.Spec.ListenersConfig.SSLSecrets.PKIBackend) != string(nc.Spec.ListenersConfig.SSLSecrets.PKIBackend) ||
		anc.Spec.ListenersConfig.SSLSecrets.TLSSecretName != nc.Spec.ListenersConfig.SSLSecrets.TLSSecretName ||
		anc.Spec.ListenersConfig.UseExternalDNS != nc.Spec.ListenersConfig.UseExternalDNS {
		t.Error("listeners configs are not equal")
	}
	if !reflect.DeepEqual(anc.Spec.SidecarConfigs, nc.Spec.SidecarConfigs) {
		t.Error("sidecar configs are not equal")
	}
	if !externalServicesEqual(anc.Spec.ExternalServices, nc.Spec.ExternalServices) {
		t.Error("external service configs are not equal")
	}
	if !reflect.DeepEqual(anc.Spec.TopologySpreadConstraints, nc.Spec.TopologySpreadConstraints) {
		t.Error("topology constraints are not equal")
	}
	if anc.Spec.NifiControllerTemplate != nc.Spec.NifiControllerTemplate {
		t.Error("nifi controller templates are not equal")
	}
	if anc.Spec.ControllerUserIdentity != nc.Spec.ControllerUserIdentity {
		t.Error("controller user identities are not equal")
	}
	if !clusterStatesEqual(anc.Status, nc.Status) {
		t.Error("cluster statuses are not equal")
	}
}

func clusterStatesEqual(s1 NifiClusterStatus, s2 v1.NifiClusterStatus) bool {
	if s1.PrometheusReportingTask.Id != s2.PrometheusReportingTask.Id ||
		s1.PrometheusReportingTask.Version != s2.PrometheusReportingTask.Version ||
		s1.RollingUpgrade.ErrorCount != s2.RollingUpgrade.ErrorCount ||
		s1.RollingUpgrade.LastSuccess != s2.RollingUpgrade.LastSuccess ||
		string(s1.State) != string(s2.State) ||
		s1.RootProcessGroupId != s2.RootProcessGroupId {
		return false
	}

	for i, ns := range s1.NodesState {
		if string(ns.ConfigurationState) != string(s2.NodesState[i].ConfigurationState) ||
			!reflect.DeepEqual(ns.CreationTime, s2.NodesState[i].CreationTime) ||
			string(ns.GracefulActionState.ActionStep) != string(s2.NodesState[i].GracefulActionState.ActionStep) ||
			ns.GracefulActionState.ErrorMessage != s2.NodesState[i].GracefulActionState.ErrorMessage ||
			string(ns.GracefulActionState.State) != string(s2.NodesState[i].GracefulActionState.State) ||
			ns.GracefulActionState.TaskStarted != s2.NodesState[i].GracefulActionState.TaskStarted ||
			ns.LastUpdatedTime != s2.NodesState[i].LastUpdatedTime ||
			bool(ns.InitClusterNode) != bool(s2.NodesState[i].InitClusterNode) ||
			ns.PodIsReady != s2.NodesState[i].PodIsReady {
			return false
		}
	}

	return true
}

func externalServicesEqual(es1 []ExternalServiceConfig, es2 []v1.ExternalServiceConfig) bool {
	if len(es1) != len(es2) {
		return false
	}

	for i, es := range es1 {
		if !metadataEqual(es.Metadata, es2[i].Metadata) ||
			es.Name != es2[i].Name ||
			es.Spec.ClusterIP != es2[i].Spec.ClusterIP ||
			es.Spec.ExternalName != es2[i].Spec.ExternalName ||
			es.Spec.LoadBalancerIP != es2[i].Spec.LoadBalancerIP ||
			!reflect.DeepEqual(es.Spec.ExternalIPs, es2[i].Spec.ExternalIPs) ||
			!reflect.DeepEqual(es.Spec.LoadBalancerSourceRanges, es2[i].Spec.LoadBalancerSourceRanges) ||
			es.Spec.Type != es2[i].Spec.Type {
			return false
		}
		for j, pc := range es.Spec.PortConfigs {
			if pc.InternalListenerName != es2[i].Spec.PortConfigs[j].InternalListenerName ||
				pc.Port != es2[i].Spec.PortConfigs[j].Port || es2[i].Spec.PortConfigs[j].Protocol != corev1.ProtocolTCP {
				return false
			}
		}
	}
	return true
}

func internalListenersConfigsEqual(lc1 []InternalListenerConfig, lc2 []v1.InternalListenerConfig) bool {
	if len(lc1) != len(lc2) {
		return false
	}

	for i, lc := range lc1 {
		if lc.ContainerPort != lc2[i].ContainerPort ||
			lc.Name != lc2[i].Name ||
			lc.Type != lc2[i].Type ||
			// this protocol assertion verifies the default gets set properly
			lc2[i].Protocol != corev1.ProtocolTCP {
			return false
		}
	}
	return true
}

func nodesEqual(n1 []Node, n2 []v1.Node, t *testing.T) bool {
	if len(n1) != len(n2) {
		return false
	}

	for i, node := range n1 {
		assertNodeConfigsEqual(string(node.Id), *node.NodeConfig, *n2[i].NodeConfig, t)
		assertReadOnlyConfigsEqual(*n2[i].ReadOnlyConfig, *node.ReadOnlyConfig, t)
		if node.Id != n2[i].Id ||
			!reflect.DeepEqual(node.Labels, n2[i].Labels) ||
			node.NodeConfigGroup != n2[i].NodeConfigGroup {
			return false
		}
	}
	return true
}

func assertNodeConfigGroupsEqual(ncg map[string]NodeConfig, v1ncg map[string]v1.NodeConfig, t *testing.T) {
	if !reflect.DeepEqual(maps.Keys(ncg), maps.Keys(v1ncg)) {
		t.Error("node config group keys are not equal")
	}

	for key, config := range ncg {
		assertNodeConfigsEqual(key, config, v1ncg[key], t)
	}
}

func assertNodeConfigsEqual(group string, nc NodeConfig, v1nc v1.NodeConfig, t *testing.T) {
	if nc.ProvenanceStorage != v1nc.ProvenanceStorage {
		t.Errorf("node config provenance storages not equal for group %s", group)
	}
	if !externalVolumeConfigsEqual(nc.ExternalVolumeConfigs, v1nc.ExternalVolumeConfigs) {
		t.Errorf("node config external volume not equal for group %s", group)
	}
	if !reflect.DeepEqual(nc.HostAliases, v1nc.HostAliases) {
		t.Errorf("node config host aliases not equal for group %s", group)
	}
	if !reflect.DeepEqual(nc.ImagePullPolicy, v1nc.ImagePullPolicy) {
		t.Errorf("node config image pull policies not equal for group %s", group)
	}
	if !reflect.DeepEqual(nc.ImagePullSecrets, v1nc.ImagePullSecrets) {
		t.Errorf("node config image pull secretes not equal for group %s", group)
	}
	if !reflect.DeepEqual(nc.NodeAffinity, v1nc.NodeAffinity) {
		t.Errorf("node config node affinity not equal for group %s", group)
	}
	if !reflect.DeepEqual(nc.NodeSelector, v1nc.NodeSelector) {
		t.Errorf("node config node selector not equal for group %s", group)
	}
	if !reflect.DeepEqual(nc.ResourcesRequirements, v1nc.ResourcesRequirements) {
		t.Errorf("node config resources requirements not equal for group %s", group)
	}
	if !reflect.DeepEqual(nc.Tolerations, v1nc.Tolerations) {
		t.Errorf("node config tolerations not equal for group %s", group)
	}
	if nc.FSGroup != v1nc.FSGroup {
		t.Errorf("node config FS Groups not equal for group %s", group)
	}
	if nc.RunAsUser != v1nc.RunAsUser {
		t.Errorf("node config RunAs User not equal for group %s", group)
	}
	if nc.Image != v1nc.Image {
		t.Errorf("node config FS Groups not equal for group %s", group)
	}
	if nc.IsNode != v1nc.IsNode {
		t.Errorf("node config IsNode not equal for group %s", group)
	}
	if nc.PriorityClassName != v1nc.PriorityClassName {
		t.Errorf("node config priority class name not equal for group %s", group)
	}
	if nc.ServiceAccountName != v1nc.ServiceAccountName {
		t.Errorf("node config service account name not equal for group %s", group)
	}
	if !metadataEqual(nc.PodMetadata, v1nc.PodMetadata) {
		t.Errorf("node config pod metadata not equal for group %s", group)
	}
	if !storageConfigsEqual(nc.StorageConfigs, v1nc.StorageConfigs) {
		t.Errorf("node config storage configs not equal for group %s", group)
	}
}

func storageConfigsEqual(sc1 []StorageConfig, sc2 []v1.StorageConfig) bool {
	if len(sc1) != len(sc2) {
		return false
	}
	for i, sc := range sc1 {
		if sc.MountPath != sc2[i].MountPath ||
			sc.Name != sc2[i].Name ||
			corev1.PersistentVolumeReclaimDelete != sc2[i].ReclaimPolicy ||
			!reflect.DeepEqual(sc.PVCSpec, sc2[i].PVCSpec) {
			return false
		}
	}
	return true
}

func metadataEqual(m1 Metadata, m2 v1.Metadata) bool {
	if !reflect.DeepEqual(m1.Annotations, m2.Annotations) ||
		!reflect.DeepEqual(m1.Labels, m2.Labels) {
		return false
	}
	return true
}

func externalVolumeConfigsEqual(vc1 []VolumeConfig, vc2 []v1.VolumeConfig) bool {
	if len(vc1) != len(vc2) {
		return false
	}
	for i, vc := range vc1 {
		if !reflect.DeepEqual(vc.VolumeMount, vc2[i].VolumeMount) ||
			!reflect.DeepEqual(vc.VolumeSource, vc2[i].VolumeSource) {
			return false
		}
	}
	return true
}

func managedUsersEqual(s1 []ManagedUser, s2 []v1.ManagedUser) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, elem := range s1 {
		if elem.Identity != s2[i].Identity ||
			elem.Name != s2[i].Name {
			return false
		}
	}
	return true
}

func createNifiCluster() *NifiCluster {
	now := metav1.Now()
	return &NifiCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nc",
			Namespace: "namespace",
			Labels: map[string]string{
				"key": "value",
			},
			Annotations: map[string]string{
				"key": "value",
			},
		},
		Spec: NifiClusterSpec{
			ClientType:         ClientConfigTLS,
			Type:               ExternalCluster,
			NodeURITemplate:    "nodeUri.template",
			NifiURI:            "nifi.host.name",
			RootProcessGroupId: "abc-123",
			ZKAddress:          "zk.host",
			ZKPath:             "/path/in/zk",
			InitContainerImage: "image:tag",
			ProxyUrl:           "proxyUrl",
			ClusterImage:       "image:tag",
			OneNifiNodePerNode: true,
			PropagateLabels:    true,
			SecretRef: SecretReference{
				Name:      "secretRef",
				Namespace: "namespace",
			},
			Service: ServicePolicy{
				HeadlessEnabled: true,
				ServiceTemplate: "serviceTemplate",
				Annotations: map[string]string{
					"key": "value",
				},
				Labels: map[string]string{
					"key": "value",
				},
			},
			Pod: PodPolicy{
				Annotations: map[string]string{
					"key": "value",
				},
				Labels: map[string]string{
					"key": "value",
				},
				HostAliases: []corev1.HostAlias{
					{
						IP:        "1.2.3.4",
						Hostnames: []string{"blah.host"},
					},
				},
			},
			InitContainers: []corev1.Container{
				{
					Name: "blah",
				},
			},
			ManagedAdminUsers: []ManagedUser{
				{
					Identity: "identity",
					Name:     "name",
				},
			},
			ManagedReaderUsers: []ManagedUser{
				{
					Identity: "identity",
					Name:     "name",
				},
			},
			ReadOnlyConfig: createReadOnlyConfig(),
			NodeConfigGroups: map[string]NodeConfig{
				"group": createNodeConfig(),
			},
			NodeUserIdentityTemplate: new(string),
			Nodes: []Node{
				{
					Id:              int32(4),
					NodeConfigGroup: "group",
					NodeConfig:      createNodeConfigPtr(),
					ReadOnlyConfig:  createReadOnlyConfigPtr(),
					Labels: map[string]string{
						"key": "value",
					},
				},
			},
			DisruptionBudget: DisruptionBudget{
				Create: true,
				Budget: "50",
			},
			LdapConfiguration: LdapConfiguration{
				Enabled:      true,
				Url:          "url",
				SearchBase:   "searchBase",
				SearchFilter: "searchFilter",
			},
			NifiClusterTaskSpec: NifiClusterTaskSpec{
				RetryDurationMinutes: 5,
			},
			ListenersConfig: &ListenersConfig{
				InternalListeners: []InternalListenerConfig{
					{
						Type:          "type",
						Name:          "name",
						ContainerPort: 44,
					},
				},
				SSLSecrets: &SSLSecrets{
					TLSSecretName: "secret",
					Create:        true,
					ClusterScoped: true,
					IssuerRef:     &cmmeta.ObjectReference{},
					PKIBackend:    PKIBackendCertManager,
				},
				ClusterDomain:  "domain",
				UseExternalDNS: true,
			},
			SidecarConfigs: []corev1.Container{
				{
					Name: "sidecar",
				},
			},
			ExternalServices: []ExternalServiceConfig{
				{
					Name: "externalService",
					Metadata: Metadata{
						Annotations: map[string]string{
							"key": "value",
						},
						Labels: map[string]string{
							"key": "value",
						},
					},
					Spec: ExternalServiceSpec{
						PortConfigs: []PortConfig{
							{
								Port:                 4,
								InternalListenerName: "listener",
							},
						},
						ClusterIP:                "clusterIp",
						Type:                     corev1.ServiceTypeClusterIP,
						ExternalIPs:              []string{"ip1", "ip2"},
						LoadBalancerIP:           "lbip",
						LoadBalancerSourceRanges: []string{"r1", "r2"},
						ExternalName:             "externalName",
					},
				},
			},
			TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
				{
					MaxSkew: 10,
				},
			},
			NifiControllerTemplate: new(string),
			ControllerUserIdentity: new(string),
		},
		Status: NifiClusterStatus{
			NodesState: map[string]NodeState{
				"1": {
					GracefulActionState: GracefulActionState{
						ErrorMessage: "error",
						ActionStep:   ConnectNodeAction,
						TaskStarted:  "started",
						State:        GracefulDownscaleRequired,
					},
					ConfigurationState: ConfigInSync,
					InitClusterNode:    IsInitClusterNode,
					PodIsReady:         true,
					CreationTime:       &now,
					LastUpdatedTime:    now,
				},
			},
			State: NifiClusterInitialized,
			RollingUpgrade: RollingUpgradeStatus{
				LastSuccess: "lastSuccess",
				ErrorCount:  10,
			},
			RootProcessGroupId: "rootProcessGroupId",
			PrometheusReportingTask: PrometheusReportingTaskStatus{
				Id:      "id",
				Version: 77,
			},
		},
	}
}

func createNodeConfigPtr() *NodeConfig {
	nc := createNodeConfig()
	return &nc
}

func createNodeConfig() NodeConfig {
	return NodeConfig{
		ProvenanceStorage: "provStorage",
		RunAsUser:         new(int64),
		FSGroup:           new(int64),
		IsNode:            new(bool),
		Image:             "image:tag",
		ImagePullPolicy:   corev1.PullAlways,
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "key",
								Operator: corev1.NodeSelectorOpExists,
							},
						},
					},
				},
			},
		},
		StorageConfigs: []StorageConfig{
			{
				Name:      "storage",
				MountPath: "/path",
				PVCSpec: &corev1.PersistentVolumeClaimSpec{
					StorageClassName: new(string),
				},
			},
		},
		ServiceAccountName: "serviceAccount",
		NodeSelector: map[string]string{
			"key": "value",
		},
		Tolerations: []corev1.Toleration{
			{
				Key:      "k",
				Operator: corev1.TolerationOpExists,
				Value:    "v",
			},
		},
		PodMetadata: Metadata{
			Annotations: map[string]string{
				"key": "value",
			},
			Labels: map[string]string{
				"key": "value",
			},
		},
		ResourcesRequirements: &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				"memory": *resource.NewDecimalQuantity(*resource.Zero.AsDec(), resource.DecimalSI),
				"cpu":    *resource.NewDecimalQuantity(*resource.Zero.AsDec(), resource.DecimalSI),
			},
			Requests: corev1.ResourceList{
				"memory": *resource.NewDecimalQuantity(*resource.Zero.AsDec(), resource.DecimalSI),
				"cpu":    *resource.NewDecimalQuantity(*resource.Zero.AsDec(), resource.DecimalSI),
			},
		},
		ExternalVolumeConfigs: []VolumeConfig{
			{
				VolumeMount: corev1.VolumeMount{
					Name: "name",
				},
			},
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: "pullSecret",
			},
		},
		HostAliases: []corev1.HostAlias{
			{
				IP:        "1.2.3.4",
				Hostnames: []string{"hostname"},
			},
		},
		PriorityClassName: new(string),
	}
}
