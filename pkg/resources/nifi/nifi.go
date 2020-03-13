package nifi

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/erdrix/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/erdrix/nifikop/pkg/errorfactory"
	"github.com/erdrix/nifikop/pkg/k8sutil"
	"github.com/erdrix/nifikop/pkg/resources"
	"github.com/erdrix/nifikop/pkg/resources/templates"
	"github.com/erdrix/nifikop/pkg/scale"
	"github.com/erdrix/nifikop/pkg/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const(
	componentName		= "nifi"

	nodeConfigMapVolumeMount	= "node-config"
	nifiDataVolumeMount			= "nifi-data"

	serverKeystoreVolume	= "server-ks-files"
	serverKeystorePath		= "/var/run/secrets/java.io/keystores/server"
	clientKeystoreVolume	= "client-ks-files"
	clientKeystorePath		= "/var/run/secrets/java.io/keystores/client"
)

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
	Scheme *runtime.Scheme
}

// LabelsForNifi returns the labels for selecting the resources
// belonging to the given Nifi CR name.
func LabelsForNifi(name string) map[string]string {
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
	err = r.reconcileNifiPodDelete(log)
	if err != nil {
		return errors.WrapIf(err, "failed to reconcile resource")
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
		o := r.configMap(node.Id, nodeConfig, log)
		err = k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
		if err != nil {
				return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
			}
		/*} else {
			if nodeState, ok := r.NifiCluster.Status.NodesState[strconv.Itoa(int(node.Id))]; ok {
				if nodeState.RackAwarenessState == v1alpha1.Configured {
					o := r.configMap(node.Id, nodeConfig, log)
					err := k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
					if err != nil {
						return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
					}
				}
			}
		}*/

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
		o = r.pod(node.Id, nodeConfig, pvcs, log)
		err = r.reconcileNifiPod(log, o.(*corev1.Pod))
		if err != nil {
			return err
		}
	}


	log.V(1).Info("Reconciled")

	return nil
}

func (r *Reconciler) reconcileNifiPodDelete(log logr.Logger) error {

	podList := &corev1.PodList{}
	matchingLabels := client.MatchingLabels(LabelsForNifi(r.NifiCluster.Name))

	err := r.Client.List(context.TODO(), podList,
		client.ListOption(client.InNamespace(r.NifiCluster.Namespace)), client.ListOption(matchingLabels))
	if err != nil {
		return errors.WrapIf(err, "failed to reconcile resource")
	}


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

	if len(deletedNodes) > 0 {
		// If pods is still running
		if !arePodsAlreadyDeleted(deletedNodes, log) {
			var nodesPendingGracefulDownscale []string

			deletedNodesId := generateNodeIdsFromPodSlice(deletedNodes)

			for i := range deletedNodes {
				if nodeState, ok := r.NifiCluster.Status.NodesState[deletedNodesId[i]]; ok {
					nState := nodeState.GracefulActionState.State
					if nState != v1alpha1.GracefulDownscaleRunning && (nState == v1alpha1.GracefulUpscaleSucceeded ||
						nState == v1alpha1.GracefulUpscaleRequired) {
						nodesPendingGracefulDownscale = append(nodesPendingGracefulDownscale, deletedNodesId[i])
					}
				}
			}

			if len(nodesPendingGracefulDownscale) > 0 {
				err = k8sutil.UpdateNodeStatus(r.Client, nodesPendingGracefulDownscale, r.NifiCluster,
					v1alpha1.GracefulActionState{
						State: v1alpha1.GracefulDownscaleRequired,
					}, log)
				if err != nil {
					return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)",
						strings.Join(nodesPendingGracefulDownscale, ","))
				}
			}
		}

		for _, node := range deletedNodes {

			if node.ObjectMeta.DeletionTimestamp != nil {
				log.Info(fmt.Sprintf("Nopde %s is already on terminating state", node.Labels["nodeId"]))
				continue
			}

			if nodeState, ok :=  r.NifiCluster.Status.NodesState[node.Labels["nodeId"]]; ok &&
				nodeState.GracefulActionState.ActionStep != v1alpha1.OffloadStatus && nodeState.GracefulActionState.ActionStep != v1alpha1.RemovePodAction  {

				if nodeState.GracefulActionState.State == v1alpha1.GracefulDownscaleRunning {
					log.Info("Nifi task is still running for node", "nodeId", node.Labels["nodeId"], "ActionStep", nodeState.GracefulActionState.ActionStep)
				}
				continue
			}

			err = k8sutil.UpdateNodeStatus(r.Client, []string{node.Labels["nodeId"]}, r.NifiCluster,
				v1alpha1.GracefulActionState{ActionStep: v1alpha1.RemovePodAction, State: v1alpha1.GracefulDownscaleRunning,
					TaskStarted: r.NifiCluster.Status.NodesState[node.Labels["nodeId"]].GracefulActionState.TaskStarted }, log)

			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", node.Labels["nodeId"])
			}

			err = r.Client.Delete(context.TODO(), &node)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not delete node", "id", node.Labels["nodeId"])
			}

			err = r.Client.Delete(context.TODO(), &corev1.ConfigMap{ObjectMeta: templates.ObjectMeta(fmt.Sprintf(templates.NodeConfigTemplate+"-%s", r.NifiCluster.Name, node.Labels["nodeId"]), LabelsForNifi(r.NifiCluster.Name), r.NifiCluster)})
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not delete configmap for node", "id", node.Labels["nodeId"])
			}

			if !r.NifiCluster.Spec.HeadlessServiceEnabled {
				err = r.Client.Delete(context.TODO(), &corev1.Service{ObjectMeta: templates.ObjectMeta(fmt.Sprintf("%s-%s", r.NifiCluster.Name, node.Labels["nodeId"]), LabelsForNifi(r.NifiCluster.Name), r.NifiCluster)})
				if err != nil {
					if apierrors.IsNotFound(err) {
						// can happen when broker was not fully initialized and now is deleted
						log.Info(fmt.Sprintf("Service for Node %s not found. Continue", node.Labels["nodeId"]))
					}
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
						if apierrors.IsNotFound(err) {
							// can happen when broker was not fully initialized and now is deleted
							log.Info(fmt.Sprintf("PVC for Node %s not found. Continue", node.Labels["nodeId"]))
						}

						return errors.WrapIfWithDetails(err, "could not delete pvc for node", "id", node.Labels["nodeId"])
					}
				}
			}
			err = k8sutil.UpdateNodeStatus(r.Client, []string{node.Labels["nodeId"]}, r.NifiCluster,
			v1alpha1.GracefulActionState{
				ActionStep: v1alpha1.RemovePodStatus,
				State: v1alpha1.GracefulDownscaleRunning,
				TaskStarted: r.NifiCluster.Status.NodesState[node.Labels["nodeId"]].GracefulActionState.TaskStarted },
				log)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", node.Labels["nodeId"])
			}
		}
	}

	// TODO: Ensure usage and needing
	err = scale.EnsureRemovedNodes(r.NifiCluster.Spec.HeadlessServiceEnabled, r.NifiCluster.Spec.Nodes, r.NifiCluster.Status.NodesState,
		GetServerPort(&r.NifiCluster.Spec.ListenersConfig), r.NifiCluster.Namespace, r.NifiCluster.Name)
	if err != nil {
		return err
	}

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
	storageName := currentPVC.Annotations["storageName"]

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
		if mountPath == pvc.Annotations["mountPath"] && storageName == pvc.Annotations["storageName"] {
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

		if val, ok := r.NifiCluster.Status.NodesState[desiredPod.Labels["nodeId"]]; ok && val.GracefulActionState.State != v1alpha1.GracefulUpscaleSucceeded {
			gracefulActionState := v1alpha1.GracefulActionState{ErrorMessage: "", State: v1alpha1.GracefulUpscaleSucceeded}

			statusErr = k8sutil.UpdateNodeStatus(r.Client, []string{desiredPod.Labels["nodeId"]}, r.NifiCluster, gracefulActionState, log)
			if statusErr != nil {
				return errorfactory.New(errorfactory.StatusUpdateError{}, err, "could not update node graceful action state")
			}
		}
		log.Info("resource created")
		return nil
	} else if len(podList.Items) == 1 {
		currentPod = podList.Items[0].DeepCopy()
		nodeId := currentPod.Labels["nodeId"]
		if _, ok := r.NifiCluster.Status.NodesState[nodeId]; ok {
			if currentPod.Spec.NodeName == "" {
				log.Info(fmt.Sprintf("pod for NodeId %s does not scheduled to node yet", nodeId))
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
			//if isPodHealthy(currentPod) && r.NifiCluster.Status.NodesState[currentPod.Labels["nodeId"]].ConfigurationState == v1alpha1.ConfigInSync {
			if !k8sutil.IsPodContainsTerminatedContainer(currentPod) && r.NifiCluster.Status.NodesState[currentPod.Labels["nodeId"]].ConfigurationState == v1alpha1.ConfigInSync {
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

		if !k8sutil.IsPodContainsTerminatedContainer(currentPod) {

			if r.NifiCluster.Status.State != v1alpha1.NifiClusterRollingUpgrading {
				if err := k8sutil.UpdateCRStatus(r.Client, r.NifiCluster, v1alpha1.NifiClusterRollingUpgrading, log); err != nil {
					return errorfactory.New(errorfactory.StatusUpdateError{}, err, "setting state to rolling upgrade failed")
				}
			}

			if r.NifiCluster.Status.State == v1alpha1.NifiClusterRollingUpgrading {
				// Check if any nifi pod is in terminating or pending state
				podList := &corev1.PodList{}
				matchingLabels := client.MatchingLabels(LabelsForNifi(r.NifiCluster.Name))
				err := r.Client.List(context.TODO(), podList, client.ListOption(client.InNamespace(r.NifiCluster.Namespace)), client.ListOption(matchingLabels))
				if err != nil {
					return errors.WrapIf(err, "failed to reconcile resource")
				}
				for _, pod := range podList.Items {
					if k8sutil.IsMarkedForDeletion(pod.ObjectMeta) {
						return errorfactory.New(errorfactory.ReconcileRollingUpgrade{}, errors.New("pod is still terminating"), "rolling upgrade in progress")
					}
					if k8sutil.IsPodContainsPendingContainer(&pod) {
						return errorfactory.New(errorfactory.ReconcileRollingUpgrade{}, errors.New("pod is still creating"), "rolling upgrade in progress")
					}
				}
			}
		}

		err = r.Client.Delete(context.TODO(), currentPod)
		if err != nil {
			return errorfactory.New(errorfactory.APIFailure{}, err, "deleting resource failed", "kind", desiredType)
		}
	}
	return nil
}

