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
	"context"
	"fmt"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/dataflow"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/scale"
	"github.com/Orange-OpenSource/nifikop/pkg/nificlient/config"
	"github.com/Orange-OpenSource/nifikop/pkg/pki"
	nifiutil "github.com/Orange-OpenSource/nifikop/pkg/util/nifi"
	"reflect"
	"strings"

	"emperror.dev/errors"
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/controllersettings"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/reportingtask"

	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	"github.com/Orange-OpenSource/nifikop/pkg/k8sutil"
	"github.com/Orange-OpenSource/nifikop/pkg/resources"
	"github.com/Orange-OpenSource/nifikop/pkg/resources/templates"
	"github.com/Orange-OpenSource/nifikop/pkg/util"
	certutil "github.com/Orange-OpenSource/nifikop/pkg/util/cert"
	pkicommon "github.com/Orange-OpenSource/nifikop/pkg/util/pki"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	componentName = "nifi"

	nodeSecretVolumeMount = "node-config"
	nodeTmp               = "node-tmp"
	nifiDataVolumeMount   = "nifi-data"

	serverKeystoreVolume = "server-ks-files"
	serverKeystorePath   = "/var/run/secrets/java.io/keystores/server"
	clientKeystoreVolume = "client-ks-files"
	clientKeystorePath   = "/var/run/secrets/java.io/keystores/client"
)

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
	Scheme *runtime.Scheme
}

// New creates a new reconciler for Nifi
func New(client client.Client, directClient client.Reader, scheme *runtime.Scheme, cluster *v1alpha1.NifiCluster) *Reconciler {
	return &Reconciler{
		Scheme: scheme,
		Reconciler: resources.Reconciler{
			Client:       client,
			DirectClient: directClient,
			NifiCluster:  cluster,
		},
	}
}

//
func getCreatedPVCForNode(c client.Client, nodeID int32, namespace, crName string) ([]corev1.PersistentVolumeClaim, error) {
	foundPVCList := &corev1.PersistentVolumeClaimList{}
	matchingLabels := client.MatchingLabels{
		"nifi_cr": crName,
		"nodeId":  fmt.Sprintf("%d", nodeID),
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

	if r.NifiCluster.IsExternal() {
		log.V(1).Info("Reconciled")
		return nil
	}
	// TODO : manage external LB
	uniqueHostnamesMap := make(map[string]struct{})

	// TODO: review design
	for _, node := range r.NifiCluster.Spec.Nodes {
		for _, webProxyHost := range r.GetNifiPropertiesBase(node.Id).WebProxyHosts {
			uniqueHostnamesMap[strings.Split(webProxyHost, ":")[0]] = struct{}{}
		}
	}

	uniqueHostnames := make([]string, 0)
	for k := range uniqueHostnamesMap {
		uniqueHostnames = append(uniqueHostnames, k)
	}

	// Setup the PKI if using SSL
	if r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil {
		// reconcile the PKI
		if err := pki.GetPKIManager(r.Client, r.NifiCluster).ReconcilePKI(context.TODO(), log, r.Scheme, uniqueHostnames); err != nil {
			return err
		}
	}

	if r.NifiCluster.Spec.Service.HeadlessEnabled {
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

	for _, node := range r.NifiCluster.Spec.Nodes {
		// We need to grab names for servers and client in case user is enabling ACLs
		// That way we can continue to manage dataflows and users
		serverPass, clientPass, superUsers, err := r.getServerAndClientDetails(node.Id)
		if err != nil {
			return err
		}

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

		o := r.secretConfig(node.Id, nodeConfig, serverPass, clientPass, superUsers, log)
		err = k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}

		pvcs, err := getCreatedPVCForNode(r.Client, node.Id, r.NifiCluster.Namespace, r.NifiCluster.Name)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to list PVC's")
		}

		if !r.NifiCluster.Spec.Service.HeadlessEnabled {
			o := r.service(node.Id, log)
			err := k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
			if err != nil {
				return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
			}
		}
		o = r.pod(node.Id, nodeConfig, pvcs, log)
		err, isReady := r.reconcileNifiPod(log, o.(*corev1.Pod))
		if err != nil {
			return err
		}
		if nodeState, ok := r.NifiCluster.Status.NodesState[o.(*corev1.Pod).Labels["nodeId"]]; ok &&
			nodeState.PodIsReady != isReady {
			if err = k8sutil.UpdateNodeStatus(r.Client, []string{o.(*corev1.Pod).Labels["nodeId"]}, r.NifiCluster, isReady, log); err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)",
					"id(s)", o.(*corev1.Pod).Labels["nodeId"])
			}
		}
	}

	var err error
	// Reconcile external services
	services := r.externalServices(log)
	for _, o := range services {
		err = k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	// Handle PDB
	if r.NifiCluster.Spec.DisruptionBudget.Create {
		o, err := r.podDisruptionBudget(log)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to compute podDisruptionBudget")
		}
		err = k8sutil.Reconcile(log, r.Client, o, r.NifiCluster)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	// Handle Pod delete
	err = r.reconcileNifiPodDelete(log)
	if err != nil {
		return errors.WrapIf(err, "failed to reconcile resource")
	}

	configManager := config.GetClientConfigManager(r.Client, v1alpha1.ClusterReference{
		Namespace: r.NifiCluster.Namespace,
		Name:      r.NifiCluster.Name,
	})
	clientConfig, err := configManager.BuildConfig()
	if err != nil {
		// the cluster does not exist - should have been caught pre-flight
		return errors.WrapIf(err, "Failed to create HTTP client the for referenced cluster")
	}

	// TODO: Ensure usage and needing
	err = scale.EnsureRemovedNodes(clientConfig, r.NifiCluster)
	if err != nil && len(r.NifiCluster.Status.NodesState) > 0 {
		return err
	}

	pgRootId, err := dataflow.RootProcessGroup(clientConfig)
	if err != nil {
		return err
	}

	if err := k8sutil.UpdateRootProcessGroupIdStatus(r.Client, r.NifiCluster, pgRootId, log); err != nil {
		return err
	}

	if clientConfig.UseSSL {
		if err := r.reconcileNifiUsersAndGroups(log); err != nil {
			return errors.WrapIf(err, "failed to reconcile resource")
		}
	}

	if r.NifiCluster.Spec.ReadOnlyConfig.MaximumTimerDrivenThreadCount != nil {
		if err := r.reconcileMaximumTimerDrivenThreadCount(log); err != nil {
			return errors.WrapIf(err, "failed to reconcile ressource")
		}
	}

	if r.NifiCluster.Spec.GetMetricPort() != nil {
		if err := r.reconcilePrometheusReportingTask(log); err != nil {
			return errors.WrapIf(err, "failed to reconcile ressource")
		}
	}

	log.V(1).Info("Reconciled")

	return nil
}

func (r *Reconciler) reconcileNifiPodDelete(log logr.Logger) error {

	podList := &corev1.PodList{}
	matchingLabels := client.MatchingLabels(nifiutil.LabelsForNifi(r.NifiCluster.Name))

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

			if nodeState, ok := r.NifiCluster.Status.NodesState[node.Labels["nodeId"]]; ok &&
				nodeState.GracefulActionState.ActionStep != v1alpha1.OffloadStatus && nodeState.GracefulActionState.ActionStep != v1alpha1.RemovePodAction {

				if nodeState.GracefulActionState.State == v1alpha1.GracefulDownscaleRunning {
					log.Info("Nifi task is still running for node", "nodeId", node.Labels["nodeId"], "ActionStep", nodeState.GracefulActionState.ActionStep)
				}
				continue
			}

			err = k8sutil.UpdateNodeStatus(r.Client, []string{node.Labels["nodeId"]}, r.NifiCluster,
				v1alpha1.GracefulActionState{ActionStep: v1alpha1.RemovePodAction, State: v1alpha1.GracefulDownscaleRunning,
					TaskStarted: r.NifiCluster.Status.NodesState[node.Labels["nodeId"]].GracefulActionState.TaskStarted}, log)

			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", node.Labels["nodeId"])
			}

			err = r.Client.Delete(context.TODO(), &node)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not delete node", "id", node.Labels["nodeId"])
			}

			err = r.Client.Delete(context.TODO(), &corev1.Secret{ObjectMeta: templates.ObjectMeta(fmt.Sprintf(templates.NodeConfigTemplate+"-%s", r.NifiCluster.Name, node.Labels["nodeId"]), nifiutil.LabelsForNifi(r.NifiCluster.Name), r.NifiCluster)})
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not delete secret config for node", "id", node.Labels["nodeId"])
			}

			if !r.NifiCluster.Spec.Service.HeadlessEnabled {
				err = r.Client.Delete(context.TODO(), &corev1.Service{ObjectMeta: templates.ObjectMeta(fmt.Sprintf("%s-%s", r.NifiCluster.Name, node.Labels["nodeId"]), nifiutil.LabelsForNifi(r.NifiCluster.Name), r.NifiCluster)})
				if err != nil {
					if apierrors.IsNotFound(err) {
						// can happen when node was not fully initialized and now is deleted
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
							// can happen when node was not fully initialized and now is deleted
							log.Info(fmt.Sprintf("PVC for Node %s not found. Continue", node.Labels["nodeId"]))
						}

						return errors.WrapIfWithDetails(err, "could not delete pvc for node", "id", node.Labels["nodeId"])
					}
				}
			}

			err = k8sutil.UpdateNodeStatus(r.Client, []string{node.Labels["nodeId"]}, r.NifiCluster,
				v1alpha1.GracefulActionState{
					ActionStep:  v1alpha1.RemovePodStatus,
					State:       v1alpha1.GracefulDownscaleRunning,
					TaskStarted: r.NifiCluster.Status.NodesState[node.Labels["nodeId"]].GracefulActionState.TaskStarted},
				log)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", node.Labels["nodeId"])
			}
		}
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

func (r *Reconciler) getServerAndClientDetails(nodeId int32) (string, string, []string, error) {
	if r.NifiCluster.Spec.ListenersConfig.SSLSecrets == nil {
		return "", "", []string{}, nil
	}
	serverName := types.NamespacedName{Name: fmt.Sprintf(pkicommon.NodeServerCertTemplate, r.NifiCluster.Name, nodeId), Namespace: r.NifiCluster.Namespace}
	serverSecret := &corev1.Secret{}
	if err := r.Client.Get(context.TODO(), serverName, serverSecret); err != nil {
		if apierrors.IsNotFound(err) {
			return "", "", nil, errorfactory.New(errorfactory.ResourceNotReady{}, err, "server secret not ready")
		}
		return "", "", nil, errors.WrapIfWithDetails(err, "failed to get server secret")
	}
	serverPass := string(serverSecret.Data[v1alpha1.PasswordKey])

	clientName := types.NamespacedName{Name: fmt.Sprintf(pkicommon.NodeControllerTemplate, r.NifiCluster.Name), Namespace: r.NifiCluster.Namespace}
	clientSecret := &corev1.Secret{}
	if err := r.Client.Get(context.TODO(), clientName, clientSecret); err != nil {
		if apierrors.IsNotFound(err) {
			return "", "", nil, errorfactory.New(errorfactory.ResourceNotReady{}, err, "client secret not ready")
		}
		return "", "", nil, errors.WrapIfWithDetails(err, "failed to get client secret")
	}
	clientPass := string(clientSecret.Data[v1alpha1.PasswordKey])

	superUsers := make([]string, 0)
	for _, secret := range []*corev1.Secret{serverSecret, clientSecret} {
		cert, err := certutil.DecodeCertificate(secret.Data[corev1.TLSCertKey])
		if err != nil {
			return "", "", nil, errors.WrapIfWithDetails(err, "failed to decode certificate")
		}
		superUsers = append(superUsers, cert.Subject.String())
	}

	return serverPass, clientPass, superUsers, nil
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
		"nodeId":  desiredPVC.Labels["nodeId"],
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

			if isDesiredStorageValueInvalid(desiredPVC, currentPVC) {
				return errorfactory.New(errorfactory.InternalError{}, errors.New("could not modify pvc size"),
					"one can not reduce the size of a PVC", "kind", desiredType)
			}
			resReq := desiredPVC.Spec.Resources.Requests
			labels := desiredPVC.Labels
			desiredPVC = currentPVC.DeepCopy()
			desiredPVC.Spec.Resources.Requests = resReq
			desiredPVC.Labels = labels

			if err := r.Client.Update(context.TODO(), desiredPVC); err != nil {
				return errorfactory.New(errorfactory.APIFailure{}, err, "updating resource failed", "kind", desiredType)
			}
			log.Info("resource updated")
		}
	}
	return nil
}

func isDesiredStorageValueInvalid(desired, current *corev1.PersistentVolumeClaim) bool {
	return desired.Spec.Resources.Requests.Storage().Value() < current.Spec.Resources.Requests.Storage().Value()
}

func (r *Reconciler) reconcileNifiPod(log logr.Logger, desiredPod *corev1.Pod) (error, bool) {
	currentPod := desiredPod.DeepCopy()
	desiredType := reflect.TypeOf(desiredPod)

	log = log.WithValues("kind", desiredType)
	log.V(1).Info("searching with label because name is empty")

	podList := &corev1.PodList{}
	matchingLabels := client.MatchingLabels{
		"nifi_cr": r.NifiCluster.Name,
		"nodeId":  desiredPod.Labels["nodeId"],
	}
	err := r.Client.List(context.TODO(), podList, client.InNamespace(currentPod.Namespace), matchingLabels)
	if err != nil && len(podList.Items) == 0 {
		return errorfactory.New(errorfactory.APIFailure{},
			err, "getting resource failed", "kind", desiredType), false
	}

	if len(podList.Items) == 0 {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPod); err != nil {
			return errors.WrapIf(err, "could not apply last state to annotation"), false
		}

		if err := r.Client.Create(context.TODO(), desiredPod); err != nil {
			return errorfactory.New(errorfactory.APIFailure{},
				err, "creating resource failed", "kind", desiredType), false
		}

		// Update status to Config InSync because node is configured to go
		statusErr := k8sutil.UpdateNodeStatus(r.Client, []string{desiredPod.Labels["nodeId"]}, r.NifiCluster, v1alpha1.ConfigInSync, log)
		if statusErr != nil {
			return errorfactory.New(errorfactory.StatusUpdateError{},
				statusErr, "updating status for resource failed", "kind", desiredType), false
		}

		if val, ok := r.NifiCluster.Status.NodesState[desiredPod.Labels["nodeId"]]; ok &&
			val.GracefulActionState.State != v1alpha1.GracefulUpscaleSucceeded {
			gracefulActionState := v1alpha1.GracefulActionState{ErrorMessage: "", State: v1alpha1.GracefulUpscaleSucceeded}

			if !k8sutil.PodReady(currentPod) {
				gracefulActionState = v1alpha1.GracefulActionState{ErrorMessage: "", State: v1alpha1.GracefulUpscaleRequired}
			}

			statusErr = k8sutil.UpdateNodeStatus(r.Client, []string{desiredPod.Labels["nodeId"]}, r.NifiCluster, gracefulActionState, log)
			if statusErr != nil {
				return errorfactory.New(errorfactory.StatusUpdateError{},
					statusErr, "could not update node graceful action state"), false
			}
		}
		log.Info("resource created")
		return nil, false
	} else if len(podList.Items) == 1 {
		currentPod = podList.Items[0].DeepCopy()
		nodeId := currentPod.Labels["nodeId"]
		if _, ok := r.NifiCluster.Status.NodesState[nodeId]; ok {
			if currentPod.Spec.NodeName == "" {
				log.Info(fmt.Sprintf("pod for NodeId %s does not scheduled to node yet", nodeId))
			}
		} else {
			return errorfactory.New(errorfactory.InternalError{}, errors.New("reconcile failed"),
				fmt.Sprintf("could not find status for the given node id, %s", nodeId)), false
		}
	} else {
		return errorfactory.New(errorfactory.TooManyResources{}, errors.New("reconcile failed"),
			"more then one matching pod found", "labels", matchingLabels), false
	}

	// TODO check if this err == nil check necessary (baluchicken)
	if err == nil {

		// Since toleration does not support patchStrategy:"merge,retainKeys", we need to add all toleration from the current pod if the toleration is set in the CR
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
			if !k8sutil.IsPodTerminatedOrShutdown(currentPod) &&
				r.NifiCluster.Status.NodesState[currentPod.Labels["nodeId"]].ConfigurationState == v1alpha1.ConfigInSync {

				if val, found := r.NifiCluster.Status.NodesState[desiredPod.Labels["nodeId"]]; found &&
					val.GracefulActionState.State == v1alpha1.GracefulUpscaleRunning &&
					val.GracefulActionState.ActionStep == v1alpha1.ConnectStatus &&
					k8sutil.PodReady(currentPod) {

					if err := k8sutil.UpdateNodeStatus(r.Client, []string{desiredPod.Labels["nodeId"]}, r.NifiCluster,
						v1alpha1.GracefulActionState{ErrorMessage: "", State: v1alpha1.GracefulUpscaleSucceeded}, log); err != nil {
						return errorfactory.New(errorfactory.StatusUpdateError{},
							err, "could not update node graceful action state"), false
					}
				}

				log.V(1).Info("resource is in sync")
				return nil, k8sutil.PodReady(currentPod)
			}
		} else {
			log.Info("resource diffs",
				"patch", string(patchResult.Patch),
				"current", string(patchResult.Current),
				"modified", string(patchResult.Modified),
				"original", string(patchResult.Original))
		}

		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPod); err != nil {
			return errors.WrapIf(err, "could not apply last state to annotation"), false
		}

		if !k8sutil.IsPodTerminatedOrShutdown(currentPod) {

			if r.NifiCluster.Status.State != v1alpha1.NifiClusterRollingUpgrading {
				if err := k8sutil.UpdateCRStatus(r.Client, r.NifiCluster, v1alpha1.NifiClusterRollingUpgrading, log); err != nil {
					return errorfactory.New(errorfactory.StatusUpdateError{},
						err, "setting state to rolling upgrade failed"), false
				}
			}

			if r.NifiCluster.Status.State == v1alpha1.NifiClusterRollingUpgrading {
				// Check if any nifi pod is in terminating, pending or not ready state
				podList := &corev1.PodList{}
				matchingLabels := client.MatchingLabels(nifiutil.LabelsForNifi(r.NifiCluster.Name))
				err := r.Client.List(context.TODO(), podList, client.ListOption(client.InNamespace(r.NifiCluster.Namespace)), client.ListOption(matchingLabels))
				if err != nil {
					return errors.WrapIf(err, "failed to reconcile resource"), false
				}
				for _, pod := range podList.Items {
					if k8sutil.IsMarkedForDeletion(pod.ObjectMeta) {
						return errorfactory.New(errorfactory.ReconcileRollingUpgrade{},
							errors.New("pod is still terminating"), "rolling upgrade in progress"), false
					}
					if k8sutil.IsPodContainsPendingContainer(&pod) {
						return errorfactory.New(errorfactory.ReconcileRollingUpgrade{},
							errors.New("pod is still creating"), "rolling upgrade in progress"), false
					}

					if !k8sutil.PodReady(&pod) {
						return errorfactory.New(errorfactory.ReconcileRollingUpgrade{},
							errors.New("pod is still not ready"), "rolling upgrade in progress"), false
					}
				}
			}
		}

		err = r.Client.Delete(context.TODO(), currentPod)
		if err != nil {
			return errorfactory.New(errorfactory.APIFailure{},
				err, "deleting resource failed", "kind", desiredType), false
		}
	}

	return nil, k8sutil.PodReady(currentPod)
}

func (r *Reconciler) reconcileNifiUsersAndGroups(log logr.Logger) error {
	controllerName := types.NamespacedName{Name: fmt.Sprintf(pkicommon.NodeControllerFQDNTemplate,
		fmt.Sprintf(pkicommon.NodeControllerTemplate, r.NifiCluster.Name),
		r.NifiCluster.Namespace,
		r.NifiCluster.Spec.ListenersConfig.GetClusterDomain()), Namespace: r.NifiCluster.Namespace}

	managedUsers := append(r.NifiCluster.Spec.ManagedAdminUsers, r.NifiCluster.Spec.ManagedReaderUsers...)
	var users []*v1alpha1.NifiUser
	pFalse := false
	for _, managedUser := range managedUsers {
		users = append(users, &v1alpha1.NifiUser{
			ObjectMeta: templates.ObjectMeta(
				fmt.Sprintf("%s.%s", r.NifiCluster.Name, managedUser.Name),
				pkicommon.LabelsForNifiPKI(r.NifiCluster.Name), r.NifiCluster,
			),
			Spec: v1alpha1.NifiUserSpec{
				Identity:   managedUser.GetIdentity(),
				CreateCert: &pFalse,
				ClusterRef: v1alpha1.ClusterReference{
					Name:      r.NifiCluster.Name,
					Namespace: r.NifiCluster.Namespace,
				},
			},
		})
	}

	var managedAdminUserRef []v1alpha1.UserReference
	for _, user := range r.NifiCluster.Spec.ManagedAdminUsers {
		managedAdminUserRef = append(managedAdminUserRef, v1alpha1.UserReference{Name: fmt.Sprintf("%s.%s", r.NifiCluster.Name, user.Name)})
	}

	var managedReaderUserRef []v1alpha1.UserReference
	for _, user := range r.NifiCluster.Spec.ManagedReaderUsers {
		managedReaderUserRef = append(managedReaderUserRef, v1alpha1.UserReference{Name: fmt.Sprintf("%s.%s", r.NifiCluster.Name, user.Name)})
	}

	var managedNodeUserRef []v1alpha1.UserReference
	for _, node := range r.NifiCluster.Spec.Nodes {
		managedNodeUserRef = append(managedNodeUserRef, v1alpha1.UserReference{Name: pkicommon.GetNodeUserName(r.NifiCluster, node.Id)})
	}

	groups := []*v1alpha1.NifiUserGroup{
		// Managed admins
		{
			ObjectMeta: templates.ObjectMeta(
				fmt.Sprintf("%s.managed-admins", r.NifiCluster.Name),
				pkicommon.LabelsForNifiPKI(r.NifiCluster.Name), r.NifiCluster,
			),
			Spec: v1alpha1.NifiUserGroupSpec{
				ClusterRef: v1alpha1.ClusterReference{
					Name:      r.NifiCluster.Name,
					Namespace: r.NifiCluster.Namespace,
				},
				UsersRef: append(managedAdminUserRef, v1alpha1.UserReference{
					Name:      controllerName.Name,
					Namespace: controllerName.Namespace,
				},
				),
				AccessPolicies: []v1alpha1.AccessPolicy{
					// Global
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.FlowAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.FlowAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ControllerAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.ControllerAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ParameterContextAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.ParameterContextAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ProvenanceAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.ProvenanceAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.RestrictedComponentsAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.RestrictedComponentsAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.PoliciesAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.PoliciesAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.TenantsAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.TenantsAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.SiteToSiteAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.SystemAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.SiteToSiteAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ProxyAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.ProxyAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.CountersAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.CountersAccessPolicyResource},
					// Root process group
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ComponentsAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.ComponentsAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.OperationAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ProvenanceDataAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.DataAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.DataAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					//{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.PoliciesComponentAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					//{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.PoliciesComponentAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					//{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.DataTransferAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					//{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.DataTransferAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
				},
			},
		},
		// Managed Readers
		{
			ObjectMeta: templates.ObjectMeta(
				fmt.Sprintf("%s.managed-readers", r.NifiCluster.Name),
				pkicommon.LabelsForNifiPKI(r.NifiCluster.Name), r.NifiCluster,
			),
			Spec: v1alpha1.NifiUserGroupSpec{
				ClusterRef: v1alpha1.ClusterReference{
					Name:      r.NifiCluster.Name,
					Namespace: r.NifiCluster.Namespace,
				},
				UsersRef: managedReaderUserRef,
				AccessPolicies: []v1alpha1.AccessPolicy{
					// Global
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.FlowAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ControllerAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ParameterContextAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ProvenanceAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.RestrictedComponentsAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.PoliciesAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.TenantsAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.SiteToSiteAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.SystemAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ProxyAccessPolicyResource},
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.CountersAccessPolicyResource},
					// Root process group
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ComponentsAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.OperationAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ProvenanceDataAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.DataAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
				},
			},
		},
		// Managed Nodes
		{
			ObjectMeta: templates.ObjectMeta(
				fmt.Sprintf("%s.managed-nodes", r.NifiCluster.Name),
				pkicommon.LabelsForNifiPKI(r.NifiCluster.Name), r.NifiCluster,
			),
			Spec: v1alpha1.NifiUserGroupSpec{
				ClusterRef: v1alpha1.ClusterReference{
					Name:      r.NifiCluster.Name,
					Namespace: r.NifiCluster.Namespace,
				},
				UsersRef: managedNodeUserRef,
				AccessPolicies: []v1alpha1.AccessPolicy{
					// Global
					{Type: v1alpha1.GlobalAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ProxyAccessPolicyResource},
					// Root process group
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.ProvenanceDataAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.ProvenanceDataAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.ReadAccessPolicyAction, Resource: v1alpha1.DataAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
					{Type: v1alpha1.ComponentAccessPolicyType, Action: v1alpha1.WriteAccessPolicyAction, Resource: v1alpha1.DataAccessPolicyResource, ComponentType: v1alpha1.ProcessGroupType},
				},
			},
		},
	}

	for _, user := range users {
		if err := k8sutil.Reconcile(log, r.Client, user, r.NifiCluster); err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", user.GetObjectKind().GroupVersionKind())
		}
	}

	for _, group := range groups {
		if err := k8sutil.Reconcile(log, r.Client, group, r.NifiCluster); err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", group.GetObjectKind().GroupVersionKind())
		}
	}

	return nil
}

func (r *Reconciler) reconcilePrometheusReportingTask(log logr.Logger) error {

	var err error

	configManager := config.GetClientConfigManager(r.Client, v1alpha1.ClusterReference{
		Namespace: r.NifiCluster.Namespace,
		Name:      r.NifiCluster.Name,
	})
	clientConfig, err := configManager.BuildConfig()
	if err != nil {
		return err
	}

	// Check if the NiFi reporting task already exist
	exist, err := reportingtask.ExistReportingTaks(clientConfig, r.NifiCluster)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failure checking for existing prometheus reporting task")
	}

	if !exist {
		// Create reporting task
		status, err := reportingtask.CreateReportingTask(clientConfig, r.NifiCluster)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failure creating prometheus reporting task")
		}

		r.NifiCluster.Status.PrometheusReportingTask = *status
		if err := r.Client.Status().Update(context.TODO(), r.NifiCluster); err != nil {
			return errors.WrapIfWithDetails(err, "failed to update PrometheusReportingTask status")
		}
	}

	// Sync prometheus reporting task resource with NiFi side component
	status, err := reportingtask.SyncReportingTask(clientConfig, r.NifiCluster)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to sync PrometheusReportingTask")
	}

	r.NifiCluster.Status.PrometheusReportingTask = *status
	if err := r.Client.Status().Update(context.TODO(), r.NifiCluster); err != nil {
		return errors.WrapIfWithDetails(err, "failed to update PrometheusReportingTask status")
	}
	return nil
}

func (r *Reconciler) reconcileMaximumTimerDrivenThreadCount(log logr.Logger) error {
	configManager := config.GetClientConfigManager(r.Client, v1alpha1.ClusterReference{
		Namespace: r.NifiCluster.Namespace,
		Name:      r.NifiCluster.Name,
	})
	clientConfig, err := configManager.BuildConfig()
	if err != nil {
		return err
	}

	// Sync Maximum Timer Driven Thread Count with NiFi side component
	err = controllersettings.SyncConfiguration(clientConfig, r.NifiCluster)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to sync MaximumTimerDrivenThreadCount")
	}

	return nil
}
