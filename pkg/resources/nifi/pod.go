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
	"fmt"
	configcommon "github.com/Orange-OpenSource/nifikop/pkg/nificlient/config/common"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/resources/templates"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	nifiutil "github.com/Orange-OpenSource/nifikop/pkg/util/nifi"
	pkicommon "github.com/Orange-OpenSource/nifikop/pkg/util/pki"
	zk "github.com/Orange-OpenSource/nifikop/pkg/util/zookeeper"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// InitContainer resources
	defaultInitContainerLimitsCPU      = "0.5"
	defaultInitContainerLimitsMemory   = "0.5Gi"
	defaultInitContainerRequestsCPU    = "0.5"
	defaultInitContainerRequestsMemory = "0.5Gi"

	ContainerName string = "nifi"
)

func (r *Reconciler) pod(id int32, nodeConfig *v1alpha1.NodeConfig, pvcs []corev1.PersistentVolumeClaim, log logr.Logger) runtimeClient.Object {

	zkAddress := r.NifiCluster.Spec.ZKAddress
	zkHostname := zk.GetHostnameAddress(zkAddress)
	zkPort := zk.GetPortAddress(zkAddress)

	dataVolume, dataVolumeMount := generateDataVolumeAndVolumeMount(pvcs)

	volume := []corev1.Volume{}
	volumeMount := []corev1.VolumeMount{}
	initContainers := append([]corev1.Container{}, r.NifiCluster.Spec.InitContainers...)

	volume = append(volume, dataVolume...)
	volumeMount = append(volumeMount, dataVolumeMount...)

	if r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil {
		volume = append(volume, generateVolumesForSSL(r.NifiCluster, id)...)
		volumeMount = append(volumeMount, generateVolumeMountForSSL()...)
	}

	podVolumes := append(volume, []corev1.Volume{
		{
			Name: nodeSecretVolumeMount,
			VolumeSource: corev1.VolumeSource{
				//ConfigMap: &corev1.ConfigMapVolumeSource{
				Secret: &corev1.SecretVolumeSource{
					//LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf(templates.NodeConfigTemplate+"-%d", r.NifiCluster.Name, id)},
					SecretName:  fmt.Sprintf(templates.NodeConfigTemplate+"-%d", r.NifiCluster.Name, id),
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
		nodeConfig.GetNodeAnnotations(),
		r.NifiCluster.Spec.Pod.Annotations,
	}

	if r.NifiCluster.Spec.GetMetricPort() != nil {
		anntotationsToMerge = append(anntotationsToMerge, util.MonitoringAnnotations(*r.NifiCluster.Spec.GetMetricPort()))
	}

	// curl -kv --cert /var/run/secrets/java.io/keystores/client/tls.crt --key /var/run/secrets/java.io/keystores/client/tls.key https://nifi.trycatchlearn.fr:8433/nifi
	// curl -kv --cert /var/run/secrets/java.io/keystores/client/tls.crt --key /var/run/secrets/java.io/keystores/client/tls.key https://securenc-headless.external-dns-test.gcp.trycatchlearn.fr:8443/nifi-api/controller/cluster
	// keytool -import -noprompt -keystore /home/nifi/truststore.jks -file /var/run/secrets/java.io/keystores/server/ca.crt -storepass $(cat /var/run/secrets/java.io/keystores/server/password) -alias test1
	pod := &corev1.Pod{
		//ObjectMeta: templates.ObjectMetaWithAnnotations(
		ObjectMeta: templates.ObjectMetaWithGeneratedNameAndAnnotations(
			nifiutil.ComputeNodeName(id, r.NifiCluster.Name),
			util.MergeLabels(
				nifiutil.LabelsForNifi(r.NifiCluster.Name),
				map[string]string{"nodeId": fmt.Sprintf("%d", id)},
			),
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
					Command: []string{"sh", "-c", fmt.Sprintf(`
echo trying to contact Zookeeper: %s
until nc -vzw 1 %s %s; do
	echo "waiting for zookeeper..."
	sleep 2
done`,
						zkAddress, zkHostname, zkPort)},
					Resources: generateInitContainerResources(),
				},
			}...)),
			Affinity: &corev1.Affinity{
				PodAntiAffinity: generatePodAntiAffinity(r.NifiCluster.Name, r.NifiCluster.Spec.OneNifiNodePerNode),
			},
			Containers:                    r.injectAdditionalEnvVars(r.generateContainers(nodeConfig, id, podVolumeMounts, zkAddress)),
			Volumes:                       podVolumes,
			RestartPolicy:                 corev1.RestartPolicyNever,
			TerminationGracePeriodSeconds: util.Int64Pointer(120),
			DNSPolicy:                     corev1.DNSClusterFirst,
			ImagePullSecrets:              nodeConfig.GetImagePullSecrets(),
			ServiceAccountName:            nodeConfig.GetServiceAccount(),
			Priority:                      util.Int32Pointer(0),
			SchedulerName:                 "default-scheduler",
			Tolerations:                   nodeConfig.GetTolerations(),
			NodeSelector:                  nodeConfig.GetNodeSelector(),
		},
	}

	//if r.NifiCluster.Spec.Service.HeadlessEnabled {
	pod.Spec.Hostname = nifiutil.ComputeNodeName(id, r.NifiCluster.Name)
	pod.Spec.Subdomain = nifiutil.ComputeRequestNiFiAllNodeService(r.NifiCluster.Name, r.NifiCluster.Spec.Service.HeadlessEnabled)
	//}

	if nodeConfig.NodeAffinity != nil {
		pod.Spec.Affinity.NodeAffinity = nodeConfig.NodeAffinity
	}
	return pod
}

//
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

//
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

//
func (r *Reconciler) generateContainerPortForInternalListeners() []corev1.ContainerPort {
	var usedPorts []corev1.ContainerPort

	for _, iListeners := range r.NifiCluster.Spec.ListenersConfig.InternalListeners {
		usedPorts = append(usedPorts, corev1.ContainerPort{
			Name:          strings.ReplaceAll(iListeners.Name, "_", ""),
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: iListeners.ContainerPort,
		})
	}

	return usedPorts
}

//
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

//
func (r *Reconciler) generateDefaultContainerPort() []corev1.ContainerPort {

	usedPorts := []corev1.ContainerPort{
		// Prometheus metrics port for monitoring
		/*{
			Name:          "metrics",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: v1alpha1.MetricsPort,
		},*/
	}

	return usedPorts
}

// TODO : manage default port
func GetServerPort(l *v1alpha1.ListenersConfig) int32 {
	var httpsServerPort int32
	var httpServerPort int32
	for _, iListener := range l.InternalListeners {
		if iListener.Type == v1alpha1.HttpsListenerType {
			httpsServerPort = iListener.ContainerPort
		} else if iListener.Type == v1alpha1.HttpListenerType {
			httpServerPort = iListener.ContainerPort
		}
	}
	if httpsServerPort != 0 {
		return httpsServerPort
	}
	return httpServerPort
}

func generateVolumesForSSL(cluster *v1alpha1.NifiCluster, nodeId int32) []corev1.Volume {
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
					SecretName:  fmt.Sprintf(pkicommon.NodeControllerTemplate, cluster.Name),
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

func (r *Reconciler) generateContainers(nodeConfig *v1alpha1.NodeConfig, id int32, podVolumeMounts []corev1.VolumeMount, zkAddress string) []corev1.Container {
	var containers []corev1.Container
	containers = append(containers, r.createNifiNodeContainer(nodeConfig, id, podVolumeMounts, zkAddress))
	containers = append(containers, r.NifiCluster.Spec.SidecarConfigs...)
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})

	return containers
}

func (r *Reconciler) createNifiNodeContainer(nodeConfig *v1alpha1.NodeConfig, id int32, podVolumeMounts []corev1.VolumeMount, zkAddress string) corev1.Container {
	// ContainersPorts initialization
	nifiNodeContainersPorts := r.generateContainerPortForInternalListeners()

	nifiNodeContainersPorts = append(nifiNodeContainersPorts, r.generateContainerPortForExternalListeners()...)
	nifiNodeContainersPorts = append(nifiNodeContainersPorts, r.generateDefaultContainerPort()...)

	readinessCommand := fmt.Sprintf(`curl -kv http://$(hostname -f):%d/nifi-api`,
		GetServerPort(r.NifiCluster.Spec.ListenersConfig))

	if r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil {
		readinessCommand = fmt.Sprintf(`curl -kv --cert  %s/%s --key %s/%s https://$(hostname -f):%d/nifi`,
			serverKeystorePath,
			v1alpha1.TLSCert,
			serverKeystorePath,
			v1alpha1.TLSKey,
			GetServerPort(r.NifiCluster.Spec.ListenersConfig))
	}

	failCondition := ""

	if val, ok := r.NifiCluster.Status.NodesState[fmt.Sprint(id)]; !ok || (val.InitClusterNode != v1alpha1.IsInitClusterNode &&
		(val.GracefulActionState.State == v1alpha1.GracefulUpscaleRequired ||
			val.GracefulActionState.State == v1alpha1.GracefulUpscaleRunning)) {
		failCondition = `else
	echo fail to request cluster
	exit 1
`
	}

	requestClusterStatus := fmt.Sprintf("curl --fail -v http://%s/nifi-api/controller/cluster > $NIFI_BASE_DIR/cluster.state",
		nifiutil.GenerateRequestNiFiAllNodeAddressFromCluster(r.NifiCluster))

	if configcommon.UseSSL(r.NifiCluster) {
		requestClusterStatus = fmt.Sprintf(
			"curl --fail -kv --cert /var/run/secrets/java.io/keystores/client/tls.crt --key /var/run/secrets/java.io/keystores/client/tls.key https://%s/nifi-api/controller/cluster > $NIFI_BASE_DIR/cluster.state",
			nifiutil.GenerateRequestNiFiAllNodeAddressFromCluster(r.NifiCluster))
	}

	removesFileAction := fmt.Sprintf(`if %s; then
	echo "Successfully query NiFi cluster"
	%s
	echo "state $STATUS"
	if [[ -z "$STATUS" ]]; then 
		echo "Removing previous exec setup"
		if [ -f "$NIFI_BASE_DIR/data/users.xml" ]; then rm -f $NIFI_BASE_DIR/data/users.xml; fi
		if [ -f "$NIFI_BASE_DIR/data/authorizations.xml" ]; then rm -f  $NIFI_BASE_DIR/data/authorizations.xml; fi
		if [ -f " $NIFI_BASE_DIR/data/flow.xml.gz" ]; then rm -f  $NIFI_BASE_DIR/data/flow.xml.gz; fi
	fi
%s
fi
rm -f $NIFI_BASE_DIR/cluster.state `,
		requestClusterStatus,
		"STATUS=$(jq -r \".cluster.nodes[] | select(.address==\\\"$(hostname -f)\\\") | .status\" $NIFI_BASE_DIR/cluster.state)",
		failCondition)

	nodeAddress := nifiutil.ComputeHostListenerNodeAddress(
		id, r.NifiCluster.Name, r.NifiCluster.Namespace, r.NifiCluster.Spec.Service.HeadlessEnabled,
		r.NifiCluster.Spec.ListenersConfig.GetClusterDomain(), r.NifiCluster.Spec.ListenersConfig.UseExternalDNS,
		r.NifiCluster.Spec.ListenersConfig.InternalListeners)

	resolveIp := ""

	if r.NifiCluster.Spec.Service.HeadlessEnabled {
		resolveIp = fmt.Sprintf(`echo "Waiting for host to be reachable"
notMatchedIp=true
while $notMatchedIp
do
	echo "failed to reach %s"
	echo "Found: $ipResolved, expecting: $POD_IP"
    sleep 5

	ipResolved=$(wget --tries=1 -T 1 -O /dev/null %s  2>&1 | sed -n 3p| awk '{split($0,a,"|"); print a[2] }')
	echo "Found : $ipResolved"
    if [[ "$ipResolved" == "$POD_IP" ]]; then
		echo Ip match for $POD_IP
		notMatchedIp=false
	fi
done
echo "Hostname is successfully binded withy IP adress"`, nodeAddress, nodeAddress)
	}
	command := []string{"bash", "-ce", fmt.Sprintf(`cp ${NIFI_HOME}/tmp/* ${NIFI_HOME}/conf/
%s
%s
exec bin/nifi.sh run`, resolveIp, removesFileAction)}

	return corev1.Container{
		Name:            ContainerName,
		Image:           util.GetNodeImage(nodeConfig, r.NifiCluster.Spec.ClusterImage),
		ImagePullPolicy: nodeConfig.GetImagePullPolicy(),
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{"bash", "-c", "$NIFI_HOME/bin/nifi.sh stop"},
				},
			},
		},
		// TODO : Manage https setup use cases https://github.com/cetic/helm-nifi/blob/master/templates/statefulset.yaml#L165
		ReadinessProbe: &corev1.Probe{
			InitialDelaySeconds: readinessInitialDelaySeconds,
			TimeoutSeconds:      readinessHealthCheckTimeout,
			PeriodSeconds:       readinessHealthCheckPeriod,
			FailureThreshold:    readinessHealthCheckThreshold,
			Handler: corev1.Handler{
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
		},
		LivenessProbe: &corev1.Probe{
			InitialDelaySeconds: livenessInitialDelaySeconds,
			TimeoutSeconds:      livenessHealthCheckTimeout,
			PeriodSeconds:       livenessHealthCheckPeriod,
			FailureThreshold:    livenessHealthCheckThreshold,
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: *util.IntstrPointer(int(GetServerPort(r.NifiCluster.Spec.ListenersConfig))),
				},
			},
		},
		Env: []corev1.EnvVar{
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
		},
		Command:      command,
		Ports:        nifiNodeContainersPorts,
		VolumeMounts: podVolumeMounts,
		Resources:    *nodeConfig.GetResources(),
	}
}

func (r *Reconciler) injectAdditionalEnvVars(containers []corev1.Container) (injectedContainers []corev1.Container) {

	for _, container := range containers {
		container.Env = append(container.Env, r.NifiCluster.Spec.ReadOnlyConfig.AdditionalSharedEnvs...)
		injectedContainers = append(injectedContainers, container)
	}
	return
}
