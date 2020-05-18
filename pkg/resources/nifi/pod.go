package nifi

import (
	"fmt"
	"sort"
	"strings"

	"github.com/erdrix/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/erdrix/nifikop/pkg/resources/templates"
	"github.com/erdrix/nifikop/pkg/util"
	nifiutils "github.com/erdrix/nifikop/pkg/util/nifi"
	pkicommon "github.com/erdrix/nifikop/pkg/util/pki"
	zk "github.com/erdrix/nifikop/pkg/util/zookeeper"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const(
	livenessInitialDelaySeconds int32 = 90
	livenessHealthCheckTimeout  int32 = 20
	livenessHealthCheckPeriod   int32 = 60

	readinessInitialDelaySeconds int32 = 60
	readinessHealthCheckTimeout  int32 = 10
	readinessHealthCheckPeriod   int32 = 20

	ContainerName string = "nifi"
)

func (r *Reconciler) pod(id int32, nodeConfig *v1alpha1.NodeConfig, pvcs []corev1.PersistentVolumeClaim, log logr.Logger) runtime.Object {

	zkAddresse := r.NifiCluster.Spec.ZKAddresse
	zkHostname := zk.GetHostnameAddress(zkAddresse)
	zkPort := zk.GetPortAddress(zkAddresse)


	// ContainersPorts initialization
	nifiNodeContainersPorts := r.generateContainerPortForInternalListeners()

	nifiNodeContainersPorts = append(nifiNodeContainersPorts, r.generateContainerPortForExternalListeners()...)
	nifiNodeContainersPorts = append(nifiNodeContainersPorts, r.generateDefaultContainerPort()...)

	dataVolume, dataVolumeMount := generateDataVolumeAndVolumeMount(pvcs)

	volume 			:= []corev1.Volume{}
	volumeMount 	:= []corev1.VolumeMount{}
	//initContainers 	:= []corev1.Container{}
	initContainers 	:= append([]corev1.Container{}, r.NifiCluster.Spec.InitContainers...)

	volume 		= append(volume, dataVolume...)
	volumeMount	= append(volumeMount, dataVolumeMount...)

	readinessCommand := fmt.Sprintf(`curl -kv http://$(hostname -f):%d/nifi-api`,
		GetServerPort(&r.NifiCluster.Spec.ListenersConfig))

	if r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil {
		volume = append(volume, generateVolumesForSSL(r.NifiCluster, id)...)
		volumeMount = append(volumeMount, generateVolumeMountForSSL()...)

		readinessCommand = fmt.Sprintf(`curl -kv --cert  %s/%s --key %s/%s https://$(hostname -f):%d/nifi`,
			serverKeystorePath,
			v1alpha1.TLSCert,
			serverKeystorePath,
			v1alpha1.TLSKey,
			GetServerPort(&r.NifiCluster.Spec.ListenersConfig))
	}

	podVolumes   := append(volume, []corev1.Volume{
		{
			Name: nodeConfigMapVolumeMount,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf(templates.NodeConfigTemplate+"-%d", r.NifiCluster.Name, id)},
					DefaultMode:          util.Int32Pointer(0644),
				},
			},
		},
	}...)

	podVolumeMounts := append(volumeMount, []corev1.VolumeMount{
		{
			Name:      nodeConfigMapVolumeMount,
			MountPath: "/opt/nifi/nifi-current/tmp",

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

	command := []string{"bash", "-ce", `
cp ${NIFI_HOME}/tmp/* ${NIFI_HOME}/conf/
exec bin/nifi.sh run
`}

// curl -kv --cert /var/run/secrets/java.io/keystores/client/tls.crt --key /var/run/secrets/java.io/keystores/client/tls.key https://nifi.trycatchlearn.fr:8433/nifi
// keytool -import -noprompt -keystore /home/nifi/truststore.jks -file /var/run/secrets/java.io/keystores/server/ca.crt -storepass $(cat /var/run/secrets/java.io/keystores/server/password) -alias test1
	pod := &corev1.Pod{
		//ObjectMeta: templates.ObjectMetaWithAnnotations(
		ObjectMeta: templates.ObjectMetaWithGeneratedNameAndAnnotations(
			fmt.Sprintf(templates.NodeNameTemplate, r.NifiCluster.Name, id),
			util.MergeLabels(
				LabelsForNifi(r.NifiCluster.Name),
				map[string]string{"nodeId": fmt.Sprintf("%d", id)},
			),
			util.MergeAnnotations(
				nodeConfig.GetNodeAnnotations(),
				util.MonitoringAnnotations(v1alpha1.MetricsPort),
			), r.NifiCluster,
		),
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser:    nodeConfig.GetRunAsUser(),
				RunAsNonRoot: func(b bool) *bool { return &b }(true),
				FSGroup:      func(i int64) *int64 { return &i }(1000),
			},
			InitContainers: append(initContainers, []corev1.Container{
				{
					Name: 		"zookeeper",
					Image:		r.NifiCluster.Spec.GetInitContainerImage(),
					Command: 	[]string{"sh", "-c",fmt.Sprintf(`
						echo trying to contact %s
						until nc -vzw 1 %s %s; do
						echo "waiting for zookeeper..."
						sleep 2
						done`, zkAddresse, zkHostname, zkPort)},
				},
			}...),
			Affinity: &corev1.Affinity{
				PodAntiAffinity: generatePodAntiAffinity(r.NifiCluster.Name, r.NifiCluster.Spec.OneNifiNodePerNode),
			},
			Containers: []corev1.Container{
					/*{
						Name: "app-log",
						Image: "ez123/alpine-tini",
						Args: []string{"tail", "-n+1", "-F", "/var/log/nifi-app.log"},
						VolumeMounts:[]corev1.VolumeMount{
							{
								Name:     "logs",
								MountPath: "/var/log",

							},
						},
					},*/
				{
					Name:	ContainerName,
					Image: 	util.GetNodeImage(nodeConfig, r.NifiCluster.Spec.ClusterImage),
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
						Handler: corev1.Handler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: *util.IntstrPointer(int(GetServerPort(&r.NifiCluster.Spec.ListenersConfig))),
							},
						},
					},
					Env: []corev1.EnvVar{
						{
							Name: "NIFI_ZOOKEEPER_CONNECT_STRING",
							Value: zkAddresse,
						},
					},
					Command: command,
					Ports: nifiNodeContainersPorts,
					VolumeMounts: podVolumeMounts,
					Resources: *nodeConfig.GetResources(),
				},
			},
			Volumes: podVolumes,
			RestartPolicy: 					corev1.RestartPolicyNever,
			TerminationGracePeriodSeconds:	util.Int64Pointer(120),
			DNSPolicy:                     	corev1.DNSClusterFirst,
			ImagePullSecrets:              	nodeConfig.GetImagePullSecrets(),
			ServiceAccountName:            	nodeConfig.GetServiceAccount(),
			Priority:                      	util.Int32Pointer(0),
			SchedulerName:                 	"default-scheduler",
			Tolerations:                   	nodeConfig.GetTolerations(),
			NodeSelector:                  	nodeConfig.GetNodeSelector(),
		},
	}

	if r.NifiCluster.Spec.HeadlessServiceEnabled {
		pod.Spec.Hostname	= fmt.Sprintf(templates.NodeNameTemplate, r.NifiCluster.Name, id)
		pod.Spec.Subdomain	= fmt.Sprintf(nifiutils.HeadlessServiceTemplate, r.NifiCluster.Name)
	}

	if nodeConfig.NodeAffinity != nil {
		pod.Spec.Affinity.NodeAffinity = nodeConfig.NodeAffinity
	}
	return pod
}

//
func generateDataVolumeAndVolumeMount(pvcs []corev1.PersistentVolumeClaim) (volume []corev1.Volume, volumeMount []corev1.VolumeMount) {
	//for i, pvc := range pvcs {

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
			Name: pvc.Annotations["storageName"],
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
						MatchLabels: LabelsForNifi(clusterName),
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
							MatchLabels: LabelsForNifi(clusterName),
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
func (r *Reconciler) generateContainerPortForInternalListeners() []corev1.ContainerPort{
	var usedPorts []corev1.ContainerPort

	for _, iListeners := range r.NifiCluster.Spec.ListenersConfig.InternalListeners {
		usedPorts = append(usedPorts, corev1.ContainerPort{
			Name: 			strings.ReplaceAll(iListeners.Name, "_", ""),
			Protocol: 		corev1.ProtocolTCP,
			ContainerPort:	iListeners.ContainerPort,
		})
	}

	return usedPorts
}

//
func (r *Reconciler) generateContainerPortForExternalListeners() []corev1.ContainerPort{
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
func (r *Reconciler) generateDefaultContainerPort() []corev1.ContainerPort{

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