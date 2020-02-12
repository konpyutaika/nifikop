package nifi

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/orangeopensource/nifi-operator/pkg/apis/nifi/v1alpha1"
	"github.com/orangeopensource/nifi-operator/pkg/resources/templates"
	"github.com/orangeopensource/nifi-operator/pkg/util"
	nifiutils "github.com/orangeopensource/nifi-operator/pkg/util/nifi"
	zk "github.com/orangeopensource/nifi-operator/pkg/util/zookeeper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sort"
	"strings"
)

const(
	livenessInitialDelaySeconds int32 = 90
	livenessHealthCheckTimeout  int32 = 20
	livenessHealthCheckPeriod   int32 = 60

	readinessInitialDelaySeconds int32 = 60
	readinessHealthCheckTimeout  int32 = 10
	readinessHealthCheckPeriod   int32 = 20
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
	initContainers 	:= []corev1.Container{}

	volume 		= append(volume, dataVolume...)
	volumeMount	= append(volumeMount, dataVolumeMount...)

	podVolumes   := append(volume, []corev1.Volume{
		{
			Name: nodeConfigMapVolumeMount,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf(nodeConfigTemplate+"-%d", r.NifiCluster.Name, id)},
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

	// TODO : manage properties based on configmap
	command := []string{"bash", "-ce", `
prop_replace () {
	target_file=${NIFI_HOME}/conf/${3:-nifi.properties}
	echo "updating ${1} in ${target_file}"
	if egrep "^${1}=" ${target_file} &> /dev/null; then
		sed -i -e "s|^$1=.*$|$1=$2|"  ${target_file}
	else
		echo ${1}=${2} >> ${target_file}
	fi
}
FQDN=$(hostname -f)
cp ${NIFI_HOME}/tmp/* ${NIFI_HOME}/conf/ 
#cat "${NIFI_HOME}/conf/nifi.temp" > "${NIFI_HOME}/conf/nifi.properties"

exec bin/nifi.sh run
`}


	pod := &corev1.Pod{
		//ObjectMeta: templates.ObjectMetaWithAnnotations(
		ObjectMeta: templates.ObjectMetaWithGeneratedNameAndAnnotations(
			//fmt.Sprintf(nodeName, r.NifiCluster.Name, id),
			r.NifiCluster.Name,
			util.MergeLabels(
				labelsForNifi(r.NifiCluster.Name),
				map[string]string{"nodeId": fmt.Sprintf("%d", id)},
			),
			util.MergeAnnotations(
				nodeConfig.GetNodeAnnotations(),
				util.MonitoringAnnotations(metricsPort),
			), r.NifiCluster,
		),
		Spec: corev1.PodSpec{
			InitContainers: append(initContainers, []corev1.Container{
				{
					Name: 		"zookeeper",
					// @TODO: replace with dynamic value
					Image:		"busybox",
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
				{
					Name:	"nifi",
					Image: 	util.GetNodeImage(nodeConfig, r.NifiCluster.Spec.ClusterImage),
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.Handler{
							Exec: &corev1.ExecAction{
								Command: []string{"bash", "-c", "$NIFI_HOME/bin/nifi.sh stop"},
							},
						},
						// TODO: add dynamic PostStart for additional lib https://github.com/cetic/helm-nifi/blob/master/values.yaml#L58

					},
					// TODO : Manage https setup use cases https://github.com/cetic/helm-nifi/blob/master/templates/statefulset.yaml#L165
					ReadinessProbe: &corev1.Probe{
						InitialDelaySeconds: readinessInitialDelaySeconds,
						TimeoutSeconds:      readinessHealthCheckTimeout,
						PeriodSeconds:       readinessHealthCheckPeriod,
						Handler: corev1.Handler{
							Exec: &corev1.ExecAction{
								Command: []string{
									"bash",
									"-c",
									fmt.Sprintf(`curl -kv \
http://$(hostname -f):%d/nifi-api/controller/cluster > $NIFI_BASE_DIR/data/cluster.state
STATUS=$(jq -r ".cluster.nodes[] | select((.address==\"$(hostname -f)\") or .address==\"localhost\") | .status" $NIFI_BASE_DIR/data/cluster.state)
if [[ ! $STATUS = "CONNECTED" ]]; then
	echo "Node not found with CONNECTED state. Full cluster state:"
	jq . $NIFI_BASE_DIR/data/cluster.state
	exit 1
fi`,
										getServerPort(&r.NifiCluster.Spec.ListenersConfig)),
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
								Port: *util.IntstrPointer(int(getServerPort(&r.NifiCluster.Spec.ListenersConfig))),
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
			SecurityContext:               	&corev1.PodSecurityContext{},
			Priority:                      	util.Int32Pointer(0),
			SchedulerName:                 	"default-scheduler",
			Tolerations:                   	nodeConfig.GetTolerations(),
			NodeSelector:                  	nodeConfig.GetNodeSelector(),
		},
	}

	if r.NifiCluster.Spec.HeadlessServiceEnabled {
		pod.Spec.Hostname	= fmt.Sprintf(nodeName, r.NifiCluster.Name, id)
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
			Name: fmt.Sprintf(nifiDataVolumeMount+"-%s", pvc.Name),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.Name,
				},
			},
		})
		volumeMount = append(volumeMount, corev1.VolumeMount{
			//Name:      fmt.Sprintf(nifiDataVolumeMount+"-%d", i),
			Name: fmt.Sprintf(nifiDataVolumeMount+"-%s", pvc.Name),
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
						MatchLabels: labelsForNifi(clusterName),
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
							MatchLabels: labelsForNifi(clusterName),
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

	for _, eListener := range r.NifiCluster.Spec.ListenersConfig.ExternalListeners {
		usedPorts = append(usedPorts, corev1.ContainerPort{
			Name:       	eListener.Name,
			Protocol:   	corev1.ProtocolTCP,
			ContainerPort: 	eListener.ContainerPort,
		})
	}

	return usedPorts
}

//
func (r *Reconciler) generateDefaultContainerPort() []corev1.ContainerPort{

	usedPorts := []corev1.ContainerPort{
		// Prometheus metrics port for monitoring
		{
			Name:          "metrics",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: metricsPort,
		},
	}


	return usedPorts
}

// TODO : manage default port
func getServerPort(l *v1alpha1.ListenersConfig) int32 {
	var httpsServerPort int32
	var httpServerPort int32
	for _, iListener := range l.InternalListeners {
		if iListener.Type == httpsListenerType {
			httpsServerPort = iListener.ContainerPort
		} else if iListener.Type == httpListenerType {
			httpsServerPort = iListener.ContainerPort
		}
	}
	if &httpsServerPort != nil {
		return httpsServerPort
	}
	return httpServerPort
}