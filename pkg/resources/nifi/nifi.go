package nifi

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	"github.com/orangeopensource/nifi-operator/pkg/apis/nifi/v1alpha1"
	"github.com/orangeopensource/nifi-operator/pkg/errorfactory"
	"github.com/orangeopensource/nifi-operator/pkg/k8sutil"
	"github.com/orangeopensource/nifi-operator/pkg/resources"
	"github.com/orangeopensource/nifi-operator/pkg/resources/templates"
	"github.com/orangeopensource/nifi-operator/pkg/scale"
	"github.com/orangeopensource/nifi-operator/pkg/util"
	nifiutils "github.com/orangeopensource/nifi-operator/pkg/util/nifi"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
	"time"
)

const(
	componentName		= "nifi"
	nodeConfigTemplate	= "%s-config"
	nodeStorageTemplate	= "%s-storage"
//	nodeStorageTemplate	= "%s-%s-storage"
	nodeName            = "%s-%d"

	clusterListenerType = "cluster"
	httpListenerType 	= "http"
	httpsListenerType 	= "https"
	s2sListenerType 	= "s2s"

	ProvenanceStorage   = "8 GB"

	nodeConfigMapVolumeMount	= "node-config"
	nifiDataVolumeMount			= "nifi-data"

	serverKeystoreVolume	= "server-ks-files"
	serverKeystorePath		= "/var/run/secrets/java.io/keystores/server"
	clientKeystoreVolume	= "client-ks-files"
	clientKeystorePath		= "/var/run/secrets/java.io/keystores/client"

	metricsPort 		= 9020
	defaultServicePort 	= 8080
)

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
	Scheme *runtime.Scheme
}

// labelsForNifi returns the labels for selecting the resources
// belonging to the given Nifi CR name.
func labelsForNifi(name string) map[string]string {
	return map[string]string{"app": "nifi", "nifi_cr": name}
}

// New creates a new reconciler for Nifi
func New(client client.Client, scheme *runtime.Scheme, cluster *v1alpha1.NifiCluster) *Reconciler {
	return &Reconciler{
		Scheme: scheme,
		Reconciler: resources.Reconciler{
			Client:       client,
			NifiCluster: cluster,
		},
	}
}

//
func getCreatedPVCForNode(c client.Client, nodeID int32, namespace, crName string) ([]corev1.PersistentVolumeClaim, error) {
	foundPVCList := &corev1.PersistentVolumeClaimList{}
	matchingLabels := client.MatchingLabels{
		"nifi_cr": crName,
		"nodeId": fmt.Sprintf("%d", nodeID),
	}
	err := c.List(context.TODO(), foundPVCList, client.ListOption(client.InNamespace(namespace)), client.ListOption(matchingLabels))
	if err != nil {
		return nil, err
	}
	if len(foundPVCList.Items) == 0 {
		return nil, fmt.Errorf("no persistentvolume found for node %d", nodeID)
	}
	return foundPVCList.Items, nil
}

// Reconcile implements the reconcile logic for nifi
func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName, "clusterName", r.NifiCluster.Name, "clusterNamespace", r.NifiCluster.Namespace)

	log.V(1).Info("Reconciling")

	if r.NifiCluster.Spec.HeadlessServiceEnabled {
		o := r.headlessService()
		err := k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	} else {
		o := r.allNodeService()
		err := k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	o := r.lbService()
	err := k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
	}

	// Handle Pod delete
	podList := &corev1.PodList{}
	matchingLabels := client.MatchingLabels{
		"nifi_cr": r.NifiCluster.Name,
	}

	err = r.Client.List(context.TODO(), podList, client.ListOption(client.InNamespace(r.NifiCluster.Namespace)), client.ListOption(matchingLabels))
	if err != nil {
		return errors.WrapIf(err, "failed to reconcile resource")
	}
	if len(podList.Items) > len(r.NifiCluster.Spec.Nodes) {
		deletedNodes := make([]corev1.Pod, 0)
	OUTERLOOP:
		for _, pod := range podList.Items {
			for _, node := range r.NifiCluster.Spec.Nodes {
				if pod.Labels["nodeId"] == fmt.Sprintf("%d", node.Id) {
					continue OUTERLOOP
				}
			}
			deletedNodes = append(deletedNodes, pod)
		}

		if !arePodsAlreadyDeleted(deletedNodes, log) {
			if r.NifiCluster.Status.NodesState[generateNodeIdsFromPodSlice(deletedNodes)[0]].GracefulActionState.State != v1alpha1.GracefulUpdateRunning &&
				r.NifiCluster.Status.NodesState[generateNodeIdsFromPodSlice(deletedNodes)[0]].GracefulActionState.State != v1alpha1.GracefulDownscaleSucceeded {
				uTaskId, taskStartTime, err := scale.DownsizeCluster(strings.Join(generateNodeIdsFromPodSlice(deletedNodes), ","),
					r.NifiCluster.Namespace, r.NifiCluster.Name)
				if err != nil {
					log.Info(fmt.Sprintf("nifi cluster communication error during downscaling node(s) id(s): %s", strings.Join(generateNodeIdsFromPodSlice(deletedNodes), ",")))
					return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, fmt.Sprintf("node(s) id(s): %s", strings.Join(generateNodeIdsFromPodSlice(deletedNodes), ",")))
				}
				err = k8sutil.UpdateNodeStatus(r.Client, generateNodeIdsFromPodSlice(deletedNodes), r.NifiCluster,
					v1alpha1.GracefulActionState{ActionStep: uTaskId, State: v1alpha1.GracefulUpdateRunning,
						TaskStarted: taskStartTime}, log)
				if err != nil {
					return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", strings.Join(generateNodeIdsFromPodSlice(deletedNodes), ","))
				}
			}
			if r.NifiCluster.Status.NodesState[generateNodeIdsFromPodSlice(deletedNodes)[0]].GracefulActionState.State == v1alpha1.GracefulUpdateRunning {
				err = r.checkCCTaskState(generateNodeIdsFromPodSlice(deletedNodes),
					r.NifiCluster.Status.NodesState[generateNodeIdsFromPodSlice(deletedNodes)[0]], v1alpha1.GracefulDownscaleSucceeded, log)
				if err != nil {
					return err
				}
			}
		}

		for _, node := range deletedNodes {
			if node.ObjectMeta.DeletionTimestamp != nil {
				log.Info(fmt.Sprintf("Nopde %s is already on terminating state", node.Labels["nodeId"]))
				continue
			}
			err = r.Client.Delete(context.TODO(), &node)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not delete node", "id", node.Labels["nodeId"])
			}
			err = r.Client.Delete(context.TODO(), &corev1.ConfigMap{ObjectMeta: templates.ObjectMeta(fmt.Sprintf(nodeConfigTemplate+"-%s", r.NifiCluster.Name, node.Labels["nodeId"]), labelsForNifi(r.NifiCluster.Name), r.NifiCluster)})
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not delete configmap for node", "id", node.Labels["nodeId"])
			}
			if !r.NifiCluster.Spec.HeadlessServiceEnabled {
				err = r.Client.Delete(context.TODO(), &corev1.Service{ObjectMeta: templates.ObjectMeta(fmt.Sprintf("%s-%s", r.NifiCluster.Name, node.Labels["nodeId"]), labelsForNifi(r.NifiCluster.Name), r.NifiCluster)})
				if err != nil {
					return errors.WrapIfWithDetails(err, "could not delete service for node", "id", node.Labels["nodeId"])
				}
			}
			for _, volume := range node.Spec.Volumes {
				if strings.HasPrefix(volume.Name, nifiDataVolumeMount) {
					err = r.Client.Delete(context.TODO(), &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
						Name:      volume.PersistentVolumeClaim.ClaimName,
						Namespace: r.NifiCluster.Namespace,
					}})
					if err != nil {
						return errors.WrapIfWithDetails(err, "could not delete pvc for node", "id", node.Labels["nodeId"])
					}
				}
			}
			err = k8sutil.DeleteStatus(r.Client, node.Labels["nodeId"], r.NifiCluster, log)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not delete status for node", "id", node.Labels["nodeId"])
			}
		}
	}

	for _, node := range r.NifiCluster.Spec.Nodes {
		nodeConfig, err := util.GetNodeConfig(node, r.NifiCluster.Spec)
		if err != nil {
			return errors.WrapIf(err, "failed to reconcile resource")
		}
		for _, storage := range nodeConfig.StorageConfigs {
			o := r.pvc(node.Id, storage, log)
			err := r.reconcileNifiPVC(log, o.(*corev1.PersistentVolumeClaim))
			if err != nil {
				return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
			}

		}
		if r.NifiCluster.Spec.RackAwareness == nil {
			o := r.configMap(node.Id, nodeConfig, log)
			err := k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
			if err != nil {
				return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
			}
		} else {
			if nodeState, ok := r.NifiCluster.Status.NodesState[strconv.Itoa(int(node.Id))]; ok {
				if nodeState.RackAwarenessState == v1alpha1.Configured {
					o := r.configMap(node.Id, nodeConfig, log)
					err := k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
					if err != nil {
						return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
					}
				}
			}
		}

		pvcs, err := getCreatedPVCForNode(r.Client, node.Id, r.NifiCluster.Namespace, r.NifiCluster.Name)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to list PVC's")
		}

		if !r.NifiCluster.Spec.HeadlessServiceEnabled {
			o := r.service(node.Id, log)
			err := k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
			if err != nil {
				return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
			}
		}
		o := r.pod(node.Id, nodeConfig, pvcs, log)
		err = r.reconcileNifiPod(log, o.(*corev1.Pod))
		if err != nil {
			return err
		}
	}


	log.V(1).Info("Reconciled")

	return nil
}

//
func arePodsAlreadyDeleted(pods []corev1.Pod, log logr.Logger) bool {
	for _, node := range pods {
		if node.ObjectMeta.DeletionTimestamp == nil {
			return false
		}
		log.Info(fmt.Sprintf("Node %s is already on terminating state", node.Labels["nodeId"]))
	}
	return true
}

//
func generateNodeIdsFromPodSlice(pods []corev1.Pod) []string {
	ids := make([]string, len(pods))
	for i, node := range pods {
		ids[i] = node.Labels["nodeId"]
	}
	return ids
}

//
func (r *Reconciler) checkCCTaskState(nodeIds []string, nodeState v1alpha1.NodeState, state v1alpha1.State, log logr.Logger) error {
	parsedTime, err := nifiutils.ParseTimeStampToUnixTime(nodeState.GracefulActionState.TaskStarted)
	if err != nil {
		return errors.WrapIf(err, "could not parse timestamp")
	}
	if time.Now().Sub(parsedTime).Minutes() < r.NifiCluster.Spec.NifiClusterTaskSpec.GetDurationMinutes() {
		finished, err := scale.CheckIfNCTaskFinished(nodeState.GracefulActionState.ActionStep,
			r.NifiCluster.Namespace, r.NifiCluster.Name)
		if err != nil {
			log.Info(fmt.Sprintf("Nifi cluster communication error checking running task: %s", nodeState.GracefulActionState.ActionStep))
			return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, "nifi cluster communication error")
		}
		if !finished {
			err = k8sutil.UpdateNodeStatus(r.Client, nodeIds, r.NifiCluster,
				v1alpha1.GracefulActionState{TaskStarted: nodeState.GracefulActionState.TaskStarted,
					ActionStep: nodeState.GracefulActionState.ActionStep,
					State:  nodeState.GracefulActionState.State,
				}, log)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", strings.Join(nodeIds, ","))
			}
			log.Info(fmt.Sprintf("Cruise control task: %s is still running", nodeState.GracefulActionState.ActionStep))
			return errorfactory.New(errorfactory.NifiClusterTaskRunning{}, errors.New("nifi cluster task is still running"), fmt.Sprintf("cc task id: %s", nodeState.GracefulActionState.ActionStep))
		}
		err = k8sutil.UpdateNodeStatus(r.Client, nodeIds, r.NifiCluster,
			v1alpha1.GracefulActionState{State: state,
				TaskStarted:	nodeState.GracefulActionState.TaskStarted,
				ActionStep: 	nodeState.GracefulActionState.ActionStep,
			}, log)
		if err != nil {
			return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", strings.Join(nodeIds, ","))
		}

	} else {
		// TODO: implement logic for each cases (decommission, add node)
		log.Info(fmt.Sprintf("Rollback nifi cluster task: %s", nodeState.GracefulActionState.ActionStep))
		err := scale.KillCCTask(r.NifiCluster.Namespace, r.NifiCluster.Name)
		if err != nil {
			return errorfactory.New(errorfactory.NifiClusterNotReady{}, err, "nifi cluster communication error")
		}
		err = k8sutil.UpdateNodeStatus(r.Client, nodeIds, r.NifiCluster,
			v1alpha1.GracefulActionState{State: v1alpha1.GracefulUpdateFailed,
				ErrorMessage: "Timed out waiting for the task to complete",
				TaskStarted:  nodeState.GracefulActionState.TaskStarted,
			}, log)
		if err != nil {
			return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", strings.Join(nodeIds, ","))
		}
	}
	return nil
}

func (r *Reconciler) reconcileNifiPVC(log logr.Logger, desiredPVC *corev1.PersistentVolumeClaim) error {
	var currentPVC = desiredPVC.DeepCopy()
	desiredType := reflect.TypeOf(desiredPVC)
	log = log.WithValues("kind", desiredType)
	log.V(1).Info("searching with label because name is empty")

	pvcList := &corev1.PersistentVolumeClaimList{}
	matchingLabels := client.MatchingLabels{
		"nifi_cr": r.NifiCluster.Name,
		"nodeId": desiredPVC.Labels["nodeId"],
	}
	err := r.Client.List(context.TODO(), pvcList,
		client.InNamespace(currentPVC.Namespace), matchingLabels)
	if err != nil && len(pvcList.Items) == 0 {
		return errorfactory.New(errorfactory.APIFailure{}, err, "getting resource failed", "kind", desiredType)
	}
	mountPath := currentPVC.Annotations["mountPath"]

	// Creating the first PersistentVolume For Pod
	if len(pvcList.Items) == 0 {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPVC); err != nil {
			return errors.WrapIf(err, "could not apply last state to annotation")
		}
		if err := r.Client.Create(context.TODO(), desiredPVC); err != nil {
			return errorfactory.New(errorfactory.APIFailure{}, err, "creating resource failed", "kind", desiredType)
		}
		log.Info("resource created")
		return nil
	}
	alreadyCreated := false
	for _, pvc := range pvcList.Items {
		if mountPath == pvc.Annotations["mountPath"] {
			currentPVC = pvc.DeepCopy()
			alreadyCreated = true
			break
		}
	}
	if !alreadyCreated {
		// Creating the 2+ PersistentVolumes for Pod
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPVC); err != nil {
			return errors.WrapIf(err, "could not apply last state to annotation")
		}
		if err := r.Client.Create(context.TODO(), desiredPVC); err != nil {
			return errorfactory.New(errorfactory.APIFailure{}, err, "creating resource failed", "kind", desiredType)
		}
		return nil
	}
	if err == nil {
		if k8sutil.CheckIfObjectUpdated(log, desiredType, currentPVC, desiredPVC) {

			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPVC); err != nil {
				return errors.WrapIf(err, "could not apply last state to annotation")
			}
			desiredPVC = currentPVC

			if err := r.Client.Update(context.TODO(), desiredPVC); err != nil {
				return errorfactory.New(errorfactory.APIFailure{}, err, "updating resource failed", "kind", desiredType)
			}
			log.Info("resource updated")
		}
	}
	return nil
}


func (r *Reconciler) reconcileNifiPod(log logr.Logger, desiredPod *corev1.Pod) error {
	currentPod := desiredPod.DeepCopy()
	desiredType := reflect.TypeOf(desiredPod)

	log = log.WithValues("kind", desiredType)
	log.V(1).Info("searching with label because name is empty")

	podList := &corev1.PodList{}
	matchingLabels := client.MatchingLabels{
		"nifi_cr": r.NifiCluster.Name,
		"nodeId": desiredPod.Labels["nodeId"],
	}
	err := r.Client.List(context.TODO(), podList, client.InNamespace(currentPod.Namespace), matchingLabels)
	if err != nil && len(podList.Items) == 0 {
		return errorfactory.New(errorfactory.APIFailure{}, err, "getting resource failed", "kind", desiredType)
	}
	if len(podList.Items) == 0 {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPod); err != nil {
			return errors.WrapIf(err, "could not apply last state to annotation")
		}
		if err := r.Client.Create(context.TODO(), desiredPod); err != nil {
			return errorfactory.New(errorfactory.APIFailure{}, err, "creating resource failed", "kind", desiredType)
		}
		// Update status to Config InSync because node is configured to go
		statusErr := k8sutil.UpdateNodeStatus(r.Client, []string{desiredPod.Labels["nodeId"]}, r.NifiCluster, v1alpha1.ConfigInSync, log)
		if statusErr != nil {
			return errorfactory.New(errorfactory.StatusUpdateError{}, err, "updating status for resource failed", "kind", desiredType)
		}
		if val, ok := r.NifiCluster.Status.NodesState[desiredPod.Labels["nodeId"]]; ok && val.GracefulActionState.State != v1alpha1.GracefulUpdateNotRequired {
			gracefulActionState := v1alpha1.GracefulActionState{ErrorMessage: "", State: v1alpha1.GracefulUpdateNotRequired}

/*			if r.NifiCluster.Status.CruiseControlTopicStatus == v1alpha1.CruiseControlTopicReady {
				gracefulActionState = v1alpha1.GracefulActionState{ErrorMessage: "", CruiseControlState: v1alpha1.GracefulUpdateRequired}
			}*/
			statusErr = k8sutil.UpdateNodeStatus(r.Client, []string{desiredPod.Labels["nodeId"]}, r.NifiCluster, gracefulActionState, log)
			if statusErr != nil {
				return errorfactory.New(errorfactory.StatusUpdateError{}, err, "could not update node graceful action state")
			}
		}
		if r.NifiCluster.Spec.RackAwareness != nil {
			if val, ok := r.NifiCluster.Status.NodesState[desiredPod.Labels["nodeId"]]; ok && val.RackAwarenessState == v1alpha1.Configured {
				return nil
			}
			statusErr := k8sutil.UpdateNodeStatus(r.Client, []string{desiredPod.Labels["nodeId"]}, r.NifiCluster, v1alpha1.WaitingForRackAwareness, log)
			if statusErr != nil {
				return errorfactory.New(errorfactory.StatusUpdateError{}, err, "could not update node rack state")
			}
		}

		log.Info("resource created")
		return nil
	} else if len(podList.Items) == 1 {
		currentPod = podList.Items[0].DeepCopy()
		nodeId := currentPod.Labels["nodeId"]
		if _, ok := r.NifiCluster.Status.NodesState[nodeId]; ok {
			if r.NifiCluster.Spec.RackAwareness != nil && (r.NifiCluster.Status.NodesState[nodeId].RackAwarenessState == v1alpha1.WaitingForRackAwareness || r.NifiCluster.Status.NodesState[nodeId].RackAwarenessState == "") {
				err := k8sutil.UpdateCrWithRackAwarenessConfig(currentPod, r.NifiCluster, r.Client)
				if err != nil {
					return err
				}
				statusErr := k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, r.NifiCluster, v1alpha1.Configured, log)
				if statusErr != nil {
					return errorfactory.New(errorfactory.StatusUpdateError{}, err, "updating status for resource failed", "kind", desiredType)
				}
			}
			if currentPod.Status.Phase == corev1.PodRunning && r.NifiCluster.Status.NodesState[nodeId].GracefulActionState.State == v1alpha1.GracefulUpdateRequired {
				if r.NifiCluster.Status.NodesState[nodeId].GracefulActionState.State != v1alpha1.GracefulUpdateRunning &&
					r.NifiCluster.Status.NodesState[nodeId].GracefulActionState.State != v1alpha1.GracefulUpscaleSucceeded {
					uTaskId, taskStartTime, scaleErr := scale.UpScaleCluster(desiredPod.Labels["nodeId"], desiredPod.Namespace,  r.NifiCluster.Name)
					if scaleErr != nil {
						log.Info(fmt.Sprintf("Nifi cluster communication error during upscaling node id: %s", nodeId))
						return errorfactory.New(errorfactory.NifiClusterNotReady{}, scaleErr, fmt.Sprintf("node id: %s", nodeId))
					}
					statusErr := k8sutil.UpdateNodeStatus(r.Client, []string{nodeId}, r.NifiCluster,
						v1alpha1.GracefulActionState{ActionStep: uTaskId, State: v1alpha1.GracefulUpdateRunning,
							TaskStarted: taskStartTime}, log)
					if statusErr != nil {
						return errors.WrapIfWithDetails(err, "could not update status for node", "id", nodeId)
					}
				}
			}
			if r.NifiCluster.Status.NodesState[nodeId].GracefulActionState.State == v1alpha1.GracefulUpdateRunning {
				err = r.checkCCTaskState([]string{nodeId}, r.NifiCluster.Status.NodesState[nodeId], v1alpha1.GracefulUpscaleSucceeded, log)
				if err != nil {
					return err
				}
			}
		} else {
			return errorfactory.New(errorfactory.InternalError{}, errors.New("reconcile failed"), fmt.Sprintf("could not find status for the given node id, %s", nodeId))
		}
	} else {
		return errorfactory.New(errorfactory.TooManyResources{}, errors.New("reconcile failed"), "more then one matching pod found", "labels", matchingLabels)
	}
	// TODO check if this err == nil check necessary (baluchicken)
	if err == nil {
		//Since toleration does not support patchStrategy:"merge,retainKeys", we need to add all toleration from the current pod if the toleration is set in the CR
		if len(desiredPod.Spec.Tolerations) > 0 {
			desiredPod.Spec.Tolerations = append(desiredPod.Spec.Tolerations, currentPod.Spec.Tolerations...)
			uniqueTolerations := []corev1.Toleration{}
			keys := make(map[corev1.Toleration]bool)
			for _, t := range desiredPod.Spec.Tolerations {
				if _, value := keys[t]; !value {
					keys[t] = true
					uniqueTolerations = append(uniqueTolerations, t)
				}
			}
			desiredPod.Spec.Tolerations = uniqueTolerations
		}
		// Check if the resource actually updated
		patchResult, err := patch.DefaultPatchMaker.Calculate(currentPod, desiredPod)
		if err != nil {
			log.Error(err, "could not match objects", "kind", desiredType)
		} else if patchResult.IsEmpty() {
			if isPodHealthy(currentPod) && r.NifiCluster.Status.NodesState[currentPod.Labels["nodeId"]].ConfigurationState == v1alpha1.ConfigInSync {
				log.V(1).Info("resource is in sync")
				return nil
			}
		} else {
			log.Info("resource diffs",
				"patch", string(patchResult.Patch),
				"current", string(patchResult.Current),
				"modified", string(patchResult.Modified),
				"original", string(patchResult.Original))
		}

		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPod); err != nil {
			return errors.WrapIf(err, "could not apply last state to annotation")
		}

		if isPodHealthy(currentPod) {

			if r.NifiCluster.Status.State != v1alpha1.NifiClusterRollingUpgrading {
				if err := k8sutil.UpdateCRStatus(r.Client, r.NifiCluster, v1alpha1.NifiClusterRollingUpgrading, log); err != nil {
					return errorfactory.New(errorfactory.StatusUpdateError{}, err, "setting state to rolling upgrade failed")
				}
			}

			if r.NifiCluster.Status.State == v1alpha1.NifiClusterRollingUpgrading {
				// Check if any nifi pod is in terminating state
				podList := &corev1.PodList{}
				matchingLabels := client.MatchingLabels{
					"nifi_cr": r.NifiCluster.Name,
					"app":      "nifi",
				}
				err := r.Client.List(context.TODO(), podList, client.ListOption(client.InNamespace(r.NifiCluster.Namespace)), client.ListOption(matchingLabels))
				if err != nil {
					return errors.WrapIf(err, "failed to reconcile resource")
				}
				for _, pod := range podList.Items {
					if k8sutil.IsMarkedForDeletion(pod.ObjectMeta) {
						return errorfactory.New(errorfactory.ReconcileRollingUpgrade{}, errors.New("pod is still terminating"), "rolling upgrade in progress")
					}
				}
				// TODO : replace with offloading check ??
				/*
				errorCount := r.NifiCluster.Status.RollingUpgrade.ErrorCount

				kClient, err := kafkaclient.NewFromCluster(r.Client, r.NifiCluster)
				if err != nil {
					return errorfactory.New(errorfactory.NodesUnreachable{}, err, "could not connect to nifi nodes")
				}
				defer func() {
					if err := kClient.Close(); err != nil {
						log.Error(err, "could not close client")
					}
				}()
				offlineReplicaCount, err := kClient.OfflineReplicaCount()
				if err != nil {
					return errors.WrapIf(err, "health check failed")
				}
				replicasInSync, err := kClient.AllReplicaInSync()
				if err != nil {
					return errors.WrapIf(err, "health check failed")
				}

				if offlineReplicaCount > 0 && !replicasInSync {
					errorCount++
				}
				if errorCount >= r.NifiCluster.Spec.RollingUpgradeConfig.FailureThreshold {
					return errorfactory.New(errorfactory.ReconcileRollingUpgrade{}, errors.New("cluster is not healthy"), "rolling upgrade in progress")
				}*/
			}
		}

		err = r.Client.Delete(context.TODO(), currentPod)
		if err != nil {
			return errorfactory.New(errorfactory.APIFailure{}, err, "deleting resource failed", "kind", desiredType)
		}
	}
	return nil

}

func isPodHealthy(pod *corev1.Pod) bool {
	healthy := true
	for _, containerState := range pod.Status.ContainerStatuses {
		if containerState.State.Terminated != nil {
			healthy = false
			break
		}
	}
	return healthy
}

