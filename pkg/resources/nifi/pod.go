package nifi

import (
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/imdario/mergo"
	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	"github.com/konpyutaika/nifikop/pkg/util"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
	pkicommon "github.com/konpyutaika/nifikop/pkg/util/pki"
)

const (
	livenessInitialDelaySeconds  int32 = 90
	livenessHealthCheckTimeout   int32 = 20
	livenessHealthCheckPeriod    int32 = 60
	livenessHealthCheckThreshold int32 = 5

	readinessInitialDelaySeconds  int32 = 60
	readinessHealthCheckTimeout   int32 = 10
	readinessHealthCheckPeriod    int32 = 20
	readinessHealthCheckThreshold int32 = 5

	// InitContainer resources.
	defaultInitContainerLimitsCPU      = "0.5"
	defaultInitContainerLimitsMemory   = "0.5Gi"
	defaultInitContainerRequestsCPU    = "0.5"
	defaultInitContainerRequestsMemory = "0.5Gi"

	ContainerName string = "nifi"
)

func (r *Reconciler) pod(node v1.Node, nodeConfig *v1.NodeConfig, pvcs []corev1.PersistentVolumeClaim, log zap.Logger) (runtimeClient.Object, error) {
	zkAddress := r.NifiCluster.Spec.ZKAddress
	singleUserConfiguration := r.NifiCluster.Spec.SingleUserConfiguration
	dataVolume, dataVolumeMount := generateDataVolumeAndVolumeMount(pvcs)

	volume := []corev1.Volume{}
	volumeMount := []corev1.VolumeMount{}
	initContainers := append([]corev1.Container{}, r.NifiCluster.Spec.InitContainers...)

	volume = append(volume, dataVolume...)
	volumeMount = append(volumeMount, dataVolumeMount...)

	if len(nodeConfig.ExternalVolumeConfigs) > 0 {
		for _, volumeConfig := range nodeConfig.ExternalVolumeConfigs {
			v, vM := volumeConfig.GenerateVolumeAndVolumeMount()
			volume = append(volume, v)
			volumeMount = append(volumeMount, vM)
		}
	}

	if r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil {
		volume = append(volume, generateVolumesForSSL(r.NifiCluster, node.Id)...)
		volumeMount = append(volumeMount, generateVolumeMountForSSL()...)
	}

	podVolumes := append(volume, []corev1.Volume{
		{
			Name: nodeSecretVolumeMount,
			VolumeSource: corev1.VolumeSource{
				//ConfigMap: &corev1.ConfigMapVolumeSource{
				Secret: &corev1.SecretVolumeSource{
					//LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf(templates.NodeConfigTemplate+"-%d", r.NifiCluster.Name, id)},
					SecretName:  fmt.Sprintf(templates.NodeConfigTemplate+"-%d", r.NifiCluster.Name, node.Id),
					DefaultMode: util.Int32Pointer(0644),
				},
			},
		},
		{
			Name: nodeTmp,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}...)

	podVolumeMounts := append(volumeMount, []corev1.VolumeMount{
		{
			Name:      nodeSecretVolumeMount,
			MountPath: "/opt/nifi/nifi-current/tmp",
		},
		{
			Name:      nodeTmp,
			MountPath: "/tmp",
		},
	}...)

	sort.Slice(podVolumes, func(i, j int) bool {
		return podVolumes[i].Name < podVolumes[j].Name
	})

	sort.Slice(podVolumeMounts, func(i, j int) bool {
		return podVolumeMounts[i].Name < podVolumeMounts[j].Name
	})

	sort.Slice(initContainers, func(i, j int) bool {
		return initContainers[i].Name < initContainers[j].Name
	})

	anntotationsToMerge := []map[string]string{
		r.NifiCluster.Spec.Pod.Annotations,
		nodeConfig.GetPodAnnotations(),
	}

	labelsToMerge := []map[string]string{
		r.NifiCluster.Spec.Pod.Labels,
		nodeConfig.GetPodLabels(),
		nifiutil.LabelsForNifi(r.NifiCluster.Name),
		node.Labels,
		{"nodeId": fmt.Sprintf("%d", node.Id)},
	}

	// merge host aliases together, preferring the aliases in the nodeConfig
	allHostAliases := util.MergeHostAliases(r.NifiCluster.Spec.Pod.HostAliases, nodeConfig.HostAliases)

	if r.NifiCluster.Spec.GetMetricPort() != nil {
		anntotationsToMerge = append(anntotationsToMerge, util.MonitoringAnnotations(*r.NifiCluster.Spec.GetMetricPort()))
	}

	containers := r.generateContainers(nodeConfig, node.Id, podVolumeMounts, zkAddress, singleUserConfiguration)

	// merge provided NifiContainerSpec into the Nifi Container
	for x, container := range containers {
		if container.Name == ContainerName {
			if err := mergo.Merge(&containers[x], nodeConfig.NifiContainerSpec, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
				return nil, err
			}
		}
	}

	// curl -kv --cert /var/run/secrets/java.io/keystores/client/tls.crt --key /var/run/secrets/java.io/keystores/client/tls.key https://nifi.trycatchlearn.fr:8433/nifi
	// curl -kv --cert /var/run/secrets/java.io/keystores/client/tls.crt --key /var/run/secrets/java.io/keystores/client/tls.key https://securenc-headless.external-dns-test.gcp.trycatchlearn.fr:8443/nifi-api/controller/cluster
	// keytool -import -noprompt -keystore /home/nifi/truststore.jks -file /var/run/secrets/java.io/keystores/server/ca.crt -storepass $(cat /var/run/secrets/java.io/keystores/server/password) -alias test1
	pod := &corev1.Pod{
		//ObjectMeta: templates.ObjectMetaWithAnnotations(
		ObjectMeta: templates.ObjectMetaWithGeneratedNameAndAnnotations(
			nifiutil.ComputeNodeName(node.Id, r.NifiCluster.Name),
			util.MergeLabels(labelsToMerge...),
			util.MergeAnnotations(anntotationsToMerge...), r.NifiCluster,
		),
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser:    nodeConfig.GetRunAsUser(),
				RunAsNonRoot: func(b bool) *bool { return &b }(true),
				FSGroup:      nodeConfig.GetFSGroup(),
			},
			InitContainers: r.injectAdditionalEnvVars(append(initContainers, []corev1.Container{
				{
					Name:            "zookeeper",
					Image:           r.NifiCluster.Spec.GetInitContainerImage(),
					ImagePullPolicy: nodeConfig.GetImagePullPolicy(),
					Env: []corev1.EnvVar{
						{
							Name:  "ZK_ADDRESS",
							Value: zkAddress,
						},
					},
					// The zookeeper init check here just ensures that at least one configured zookeeper host is alive
					Command: []string{"bash", "-c", `
set -e
echo "Trying to contact Zookeeper using connection string: ${ZK_ADDRESS}"

connected=0
IFS=',' read -r -a zk_hosts <<< "${ZK_ADDRESS}"
until [ $connected -eq 1 ]
do
	for zk_host in "${zk_hosts[@]}"
	do
		IFS=':' read -r -a zk_host_port <<< "${zk_host}"
		
		echo "Checking Zookeeper Host: [${zk_host_port[0]}] Port: [${zk_host_port[1]}]"
		nc -vzw 1 ${zk_host_port[0]} ${zk_host_port[1]}
		if [ $? -eq 0 ]; then
			echo "Connected to ${zk_host_port}"
			connected=1
		fi
	done

	sleep 1
done
`},
					Resources: generateInitContainerResources(),
				},
			}...)),
			Affinity: &corev1.Affinity{
				PodAntiAffinity: generatePodAntiAffinity(r.NifiCluster.Name, r.NifiCluster.Spec.OneNifiNodePerNode),
			},
			TopologySpreadConstraints:     r.NifiCluster.Spec.TopologySpreadConstraints,
			Containers:                    r.injectAdditionalEnvVars(containers),
			HostAliases:                   allHostAliases,
			Volumes:                       podVolumes,
			RestartPolicy:                 corev1.RestartPolicyNever,
			TerminationGracePeriodSeconds: util.Int64Pointer(120),
			DNSPolicy:                     corev1.DNSClusterFirst,
			ImagePullSecrets:              nodeConfig.GetImagePullSecrets(),
			ServiceAccountName:            nodeConfig.GetServiceAccount(),
			PriorityClassName:             nodeConfig.GetPriorityClass(),
			Tolerations:                   nodeConfig.GetTolerations(),
			NodeSelector:                  nodeConfig.GetNodeSelector(),
		},
	}

	// if r.NifiCluster.Spec.Service.HeadlessEnabled {
	pod.Spec.Hostname = nifiutil.ComputeNodeName(node.Id, r.NifiCluster.Name)
	pod.Spec.Subdomain = nifiutil.ComputeRequestNiFiAllNodeService(r.NifiCluster.Name,
		r.NifiCluster.Spec.Service.GetServiceTemplate())
	//}

	if nodeConfig.NodeAffinity != nil {
		pod.Spec.Affinity.NodeAffinity = nodeConfig.NodeAffinity
	}
	return pod, nil
}

func generateDataVolumeAndVolumeMount(pvcs []corev1.PersistentVolumeClaim) (volume []corev1.Volume, volumeMount []corev1.VolumeMount) {
	for _, pvc := range pvcs {
		volume = append(volume, corev1.Volume{
			//Name: fmt.Sprintf(nifiDataVolumeMount+"-%d", i),
			//Name: fmt.Sprintf(nifiDataVolumeMount+"-%s", pvc.Name),
			Name: pvc.Annotations["storageName"],
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.Name,
				},
			},
		})
		volumeMount = append(volumeMount, corev1.VolumeMount{
			//Name:      fmt.Sprintf(nifiDataVolumeMount+"-%d", i),
			//Name: fmt.Sprintf(nifiDataVolumeMount+"-%s", pvc.Name),
			Name:      pvc.Annotations["storageName"],
			MountPath: pvc.Annotations["mountPath"],
		})
	}
	return
}

func generatePodAntiAffinity(clusterName string, hardRuleEnabled bool) *corev1.PodAntiAffinity {
	podAntiAffinity := corev1.PodAntiAffinity{}
	if hardRuleEnabled {
		podAntiAffinity = corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: nifiutil.LabelsForNifi(clusterName),
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		}
	} else {
		podAntiAffinity = corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: int32(100),
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: nifiutil.LabelsForNifi(clusterName),
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}
	}
	return &podAntiAffinity
}

func (r *Reconciler) generateContainerPortForInternalListeners() []corev1.ContainerPort {
	var usedPorts []corev1.ContainerPort
	for _, iListeners := range r.NifiCluster.Spec.ListenersConfig.InternalListeners {
		protocol := iListeners.Protocol
		if protocol == "" {
			protocol = corev1.ProtocolTCP
		}
		usedPorts = append(usedPorts, corev1.ContainerPort{
			Name:          strings.ReplaceAll(iListeners.Name, "_", ""),
			Protocol:      protocol,
			ContainerPort: iListeners.ContainerPort,
		})
	}

	return usedPorts
}

func (r *Reconciler) generateContainerPortForExternalListeners() []corev1.ContainerPort {
	var usedPorts []corev1.ContainerPort

	/*for _, eListener := range r.NifiCluster.Spec.ListenersConfig.ExternalListeners {
		usedPorts = append(usedPorts, corev1.ContainerPort{
			Name:       	eListener.Name,
			Protocol:   	corev1.ProtocolTCP,
			ContainerPort: 	eListener.ContainerPort,
		})
	}*/

	return usedPorts
}

func (r *Reconciler) generateDefaultContainerPort() []corev1.ContainerPort {
	usedPorts := []corev1.ContainerPort{
		// Prometheus metrics port for monitoring
		/*{
			Name:          "metrics",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: v1.MetricsPort,
		},*/
	}

	return usedPorts
}

// TODO: manage default port.
func GetServerPort(l *v1.ListenersConfig) int32 {
	var httpsServerPort int32
	var httpServerPort int32
	for _, iListener := range l.InternalListeners {
		if iListener.Type == v1.HttpsListenerType {
			httpsServerPort = iListener.ContainerPort
		} else if iListener.Type == v1.HttpListenerType {
			httpServerPort = iListener.ContainerPort
		}
	}
	if httpsServerPort != 0 {
		return httpsServerPort
	}
	return httpServerPort
}

func generateVolumesForSSL(cluster *v1.NifiCluster, nodeId int32) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: serverKeystoreVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  fmt.Sprintf(pkicommon.NodeServerCertTemplate, cluster.Name, nodeId),
					DefaultMode: util.Int32Pointer(0644),
				},
			},
		},
		{
			Name: clientKeystoreVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  cluster.GetNifiControllerUserIdentity(),
					DefaultMode: util.Int32Pointer(0644),
				},
			},
		},
	}
}

func generateVolumeMountForSSL() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      serverKeystoreVolume,
			MountPath: serverKeystorePath,
		},
		{
			Name:      clientKeystoreVolume,
			MountPath: clientKeystorePath,
		},
	}
}

func generateInitContainerResources() corev1.ResourceRequirements {
	resourcesLimits := corev1.ResourceList{}
	resourcesLimits[corev1.ResourceCPU], _ = resource.ParseQuantity(defaultInitContainerLimitsCPU)
	resourcesLimits[corev1.ResourceMemory], _ = resource.ParseQuantity(defaultInitContainerLimitsMemory)

	resourcesReqs := corev1.ResourceList{}
	resourcesReqs[corev1.ResourceCPU], _ = resource.ParseQuantity(defaultInitContainerRequestsCPU)
	resourcesReqs[corev1.ResourceMemory], _ = resource.ParseQuantity(defaultInitContainerRequestsMemory)

	return corev1.ResourceRequirements{
		Limits:   resourcesLimits,
		Requests: resourcesReqs,
	}
}

func (r *Reconciler) generateContainers(nodeConfig *v1.NodeConfig, id int32, podVolumeMounts []corev1.VolumeMount, zkAddress string, singleUserConfiguration v1.SingleUserConfiguration) []corev1.Container {
	var containers []corev1.Container
	containers = append(containers, r.createNifiNodeContainer(nodeConfig, id, podVolumeMounts, zkAddress, singleUserConfiguration))
	containers = append(containers, r.NifiCluster.Spec.SidecarConfigs...)
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})

	return containers
}

func (r *Reconciler) createNifiNodeContainer(nodeConfig *v1.NodeConfig, id int32, podVolumeMounts []corev1.VolumeMount, zkAddress string, singleUserConfiguration v1.SingleUserConfiguration) corev1.Container {
	// ContainersPorts initialization
	nifiNodeContainersPorts := r.generateContainerPortForInternalListeners()

	nifiNodeContainersPorts = append(nifiNodeContainersPorts, r.generateContainerPortForExternalListeners()...)
	nifiNodeContainersPorts = append(nifiNodeContainersPorts, r.generateDefaultContainerPort()...)

	readinessCommand := fmt.Sprintf(`curl -kv http://$(hostname -f):%d/nifi-api`,
		GetServerPort(r.NifiCluster.Spec.ListenersConfig))

	if r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil {
		readinessCommand = fmt.Sprintf(`curl -kv --cert  %s/%s --key %s/%s https://$(hostname -f):%d/nifi`,
			serverKeystorePath,
			v1.TLSCert,
			serverKeystorePath,
			v1.TLSKey,
			GetServerPort(r.NifiCluster.Spec.ListenersConfig))
	}
	// TODO: Manage https setup use cases https://github.com/cetic/helm-nifi/blob/master/templates/statefulset.yaml#L165
	readinessProbe := &corev1.Probe{
		InitialDelaySeconds: readinessInitialDelaySeconds,
		TimeoutSeconds:      readinessHealthCheckTimeout,
		PeriodSeconds:       readinessHealthCheckPeriod,
		FailureThreshold:    readinessHealthCheckThreshold,
		ProbeHandler: corev1.ProbeHandler{
			/*HTTPGet: &corev1.HTTPGetAction{
				Path: "/nifi-api",
				Port: intstr.FromInt(int(GetServerPort(&r.NifiCluster.Spec.ListenersConfig))),
				Scheme: corev1.URISchemeHTTPS,
				//Host: nodeHostname,
			},*/
			Exec: &corev1.ExecAction{
				Command: []string{
					"bash",
					"-c",
					readinessCommand,
				},
			},
		},
	}
	// if the readiness probe has been overridden, then use that
	if r.NifiCluster.Spec.Pod.ReadinessProbe != nil {
		readinessProbe = r.NifiCluster.Spec.Pod.ReadinessProbe
	}

	livenessProbe := &corev1.Probe{
		InitialDelaySeconds: livenessInitialDelaySeconds,
		TimeoutSeconds:      livenessHealthCheckTimeout,
		PeriodSeconds:       livenessHealthCheckPeriod,
		FailureThreshold:    livenessHealthCheckThreshold,
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: *util.IntstrPointer(int(GetServerPort(r.NifiCluster.Spec.ListenersConfig))),
			},
		},
	}
	// if the liveness probe has been overridden, then use that
	if r.NifiCluster.Spec.Pod.LivenessProbe != nil {
		livenessProbe = r.NifiCluster.Spec.Pod.LivenessProbe
	}

	nodeAddress := nifiutil.ComputeHostListenerNodeAddress(
		id, r.NifiCluster.Name, r.NifiCluster.Namespace, r.NifiCluster.Spec.ListenersConfig.GetClusterDomain(),
		r.NifiCluster.Spec.ListenersConfig.UseExternalDNS, r.NifiCluster.Spec.ListenersConfig.InternalListeners,
		r.NifiCluster.Spec.Service.GetServiceTemplate())

	envVar := []corev1.EnvVar{
		{
			Name:  "NIFI_ZOOKEEPER_CONNECT_STRING",
			Value: zkAddress,
		},
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
	}

	singleUser := ""

	if singleUserConfiguration.Enabled && singleUserConfiguration.SecretRef != nil {
		singleUser = "./bin/nifi.sh set-single-user-credentials ${SINGLE_USER_CREDENTIALS_USERNAME} ${SINGLE_USER_CREDENTIALS_PASSWORD}"
		single_user_username := corev1.EnvVar{
			Name: "SINGLE_USER_CREDENTIALS_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: singleUserConfiguration.SecretRef.Name,
					},
					Key: singleUserConfiguration.SecretKeys.Username,
				},
			}}

		single_user_password := corev1.EnvVar{
			Name: "SINGLE_USER_CREDENTIALS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: singleUserConfiguration.SecretRef.Name,
					},
					Key: singleUserConfiguration.SecretKeys.Password,
				},
			},
		}
		envVar = append(envVar, single_user_username, single_user_password)
	}

	resolveIp := ""

	if r.NifiCluster.Spec.Service.HeadlessEnabled {
		resolveIp = fmt.Sprintf(`echo "Waiting for host to be reachable"
notMatchedIp=true
while $notMatchedIp
do
	echo "failed to reach %s"
	echo "Found: $ipResolved, expecting: $POD_IP"
    sleep 5

	ipResolved=$(curl -v -4 -m 1 --connect-timeout 1 %s 2>&1 | grep -o 'Trying [0-9.]*' | awk '{gsub(/\.\.\./, ""); print $2}' | head -n 1)
	echo "Found: $ipResolved"
    if [[ "$ipResolved" == "$POD_IP" ]]; then
		echo Ip match for $POD_IP
		notMatchedIp=false
	fi
done
echo "Hostname is successfully binded withy IP address"`, nodeAddress, nodeAddress)
	}
	command := []string{"bash", "-ce", fmt.Sprintf(`cp ${NIFI_HOME}/tmp/* ${NIFI_HOME}/conf/
%s
%s
exec bin/nifi.sh run`, resolveIp, singleUser)}

	return corev1.Container{
		Name:            ContainerName,
		Image:           util.GetNodeImage(nodeConfig, r.NifiCluster.Spec.ClusterImage),
		ImagePullPolicy: nodeConfig.GetImagePullPolicy(),
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"bash", "-c", "$NIFI_HOME/bin/nifi.sh stop"},
				},
			},
		},
		ReadinessProbe: readinessProbe,
		LivenessProbe:  livenessProbe,
		Env:            envVar,
		Command:        command,
		Ports:          nifiNodeContainersPorts,
		VolumeMounts:   podVolumeMounts,
		Resources:      *nodeConfig.GetResources(),
	}
}

func (r *Reconciler) injectAdditionalEnvVars(containers []corev1.Container) (injectedContainers []corev1.Container) {
	for _, container := range containers {
		container.Env = append(container.Env, r.NifiCluster.Spec.ReadOnlyConfig.AdditionalSharedEnvs...)
		injectedContainers = append(injectedContainers, container)
	}
	return
}
