package nifi

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/controllersettings"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/dataflow"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/reportingtask"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/scale"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config"
	"github.com/konpyutaika/nifikop/pkg/pki"
	"github.com/konpyutaika/nifikop/pkg/resources"
	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	"github.com/konpyutaika/nifikop/pkg/util"
	certutil "github.com/konpyutaika/nifikop/pkg/util/cert"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
	pkicommon "github.com/konpyutaika/nifikop/pkg/util/pki"
)

const (
	componentName = "nifi"

	podTLSAutoReloadEnabledAnnotation  = "nifikop.konpyutaika.com/tls-autoreload-enabled"
	podTLSAutoReloadIntervalAnnotation = "nifikop.konpyutaika.com/tls-autoreload-interval"

	nodeSecretVolumeMount = "node-config"
	nodeTmp               = "node-tmp"

	serverKeystoreVolume = "server-ks-files"
	serverKeystorePath   = "/var/run/secrets/java.io/keystores/server"
	clientKeystoreVolume = "client-ks-files"
	clientKeystorePath   = "/var/run/secrets/java.io/keystores/client"
)

var tlsAutoReloadIntervalRegexp = regexp.MustCompile(`^[[:space:]]*[1-9][0-9]*[[:space:]]+(millis|milliseconds?|ms|secs?|seconds?|mins?|minutes?|hours?|days?)[[:space:]]*$`)

func isLinkerdInjected(p *corev1.Pod) bool {
	if p == nil {
		return false
	}
	if _, ok := p.Annotations["linkerd.io/proxy-version"]; ok {
		return true
	}
	if v, ok := p.Annotations["linkerd.io/inject"]; ok && (v == "enabled" || v == "true") {
		return true
	}
	for _, c := range p.Spec.Containers {
		if c.Name == "linkerd-proxy" {
			return true
		}
	}
	return false
}

func unionLinkerdAnnotations(desired, current *corev1.Pod) {
	if desired.Annotations == nil {
		desired.Annotations = map[string]string{}
	}
	for k, v := range current.Annotations {
		if strings.HasPrefix(k, "linkerd.io/") || strings.HasPrefix(k, "config.linkerd.io/") {
			desired.Annotations[k] = v
		}
	}
}

func unionLinkerdVolumes(desired, current *corev1.Pod) {
	have := make(map[string]struct{}, len(desired.Spec.Volumes))
	for _, v := range desired.Spec.Volumes {
		have[v.Name] = struct{}{}
	}
	for _, v := range current.Spec.Volumes {
		if strings.HasPrefix(v.Name, "linkerd-") || strings.HasPrefix(v.Name, "kube-api-access-") {
			if _, ok := have[v.Name]; !ok {
				desired.Spec.Volumes = append(desired.Spec.Volumes, v)
				have[v.Name] = struct{}{}
			}
		}
	}
}

func applyTLSAutoReloadAnnotations(pod *corev1.Pod, properties *v1.NifiProperties) {
	if pod == nil {
		return
	}

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	delete(pod.Annotations, podTLSAutoReloadEnabledAnnotation)
	delete(pod.Annotations, podTLSAutoReloadIntervalAnnotation)

	if properties == nil || !properties.IsTLSAutoReloadEnabled() {
		return
	}

	pod.Annotations[podTLSAutoReloadEnabledAnnotation] = "true"
	pod.Annotations[podTLSAutoReloadIntervalAnnotation] = properties.GetTLSAutoReloadInterval()
}

func validateTLSAutoReloadProperties(properties *v1.NifiProperties) error {
	if properties == nil || !properties.IsTLSAutoReloadEnabled() {
		return nil
	}

	interval := properties.GetTLSAutoReloadInterval()
	if !tlsAutoReloadIntervalRegexp.MatchString(interval) {
		return errorfactory.New(
			errorfactory.FatalReconcileError{},
			errors.New("invalid tlsAutoReload interval"),
			"tlsAutoReload.interval must look like '10 secs', '500 ms', '30 mins' or '12 hours'",
			"interval", interval,
		)
	}

	return nil
}

func unionLinkerdVolumeMounts(dst, src []corev1.Container) []corev1.Container {
	idx := make(map[string]int, len(dst))
	for i := range dst {
		idx[dst[i].Name] = i
	}
	for _, sc := range src {
		di, ok := idx[sc.Name]
		if !ok {
			continue
		}
		have := make(map[string]struct{}, len(dst[di].VolumeMounts))
		for _, m := range dst[di].VolumeMounts {
			have[m.Name] = struct{}{}
		}
		for _, m := range sc.VolumeMounts {
			if strings.HasPrefix(m.Name, "linkerd-") || strings.HasPrefix(m.Name, "kube-api-access-") {
				if _, ok := have[m.Name]; !ok {
					dst[di].VolumeMounts = append(dst[di].VolumeMounts, m)
					have[m.Name] = struct{}{}
				}
			}
		}
	}
	return dst
}

func meshAwareMerge(desired, current *corev1.Pod) {
	if !isLinkerdInjected(current) {
		return
	}
	unionLinkerdAnnotations(desired, current)
	unionLinkerdVolumes(desired, current)
	desired.Spec.Containers = unionLinkerdVolumeMounts(desired.Spec.Containers, current.Spec.Containers)
	desired.Spec.InitContainers = unionLinkerdVolumeMounts(desired.Spec.InitContainers, current.Spec.InitContainers)
}

// Reconciler implements the Component Reconciler.
type Reconciler struct {
	resources.Reconciler
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// New creates a new reconciler for Nifi.
func New(client client.Client, directClient client.Reader, scheme *runtime.Scheme, cluster *v1.NifiCluster, currentStatus v1.NifiClusterStatus) *Reconciler {
	return &Reconciler{
		Scheme: scheme,
		Reconciler: resources.Reconciler{
			Client:                   client,
			DirectClient:             directClient,
			NifiCluster:              cluster,
			NifiClusterCurrentStatus: currentStatus,
		},
	}
}

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
	PVCList := make([]corev1.PersistentVolumeClaim, 0)
	for _, pvc := range foundPVCList.Items {
		if !k8sutil.IsMarkedForDeletion(pvc.ObjectMeta) {
			PVCList = append(PVCList, pvc)
		}
	}
	return PVCList, nil
}

// Reconcile implements the reconcile logic for nifi.
func (r *Reconciler) Reconcile(log zap.Logger) error {
	log.Debug("reconciling",
		zap.String("component", componentName),
		zap.String("clusterName", r.NifiCluster.Name),
		zap.String("clusterNamespace", r.NifiCluster.Namespace),
	)

	if r.NifiCluster.IsExternal() || r.NifiCluster.Status.State == v1.NifiClusterNoNodes {
		log.Debug("reconciled",
			zap.String("component", componentName),
			zap.String("clusterName", r.NifiCluster.Name),
			zap.String("clusterNamespace", r.NifiCluster.Namespace))
		return nil
	}
	// TODO: manage external LB
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
	// Preserving order
	sort.Strings(uniqueHostnames)

	// Setup the PKI if using SSL
	if r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil {
		// reconcile the PKI
		if err := pki.GetPKIManager(r.Client, r.NifiCluster).ReconcilePKI(context.TODO(), log, r.Scheme, uniqueHostnames); err != nil {
			return err
		}
	}

	if r.NifiCluster.Spec.Service.HeadlessEnabled {
		o := r.headlessService()
		err := k8sutil.Reconcile(log, r.Client, o, r.NifiCluster, &r.NifiClusterCurrentStatus)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	} else {
		o := r.allNodeService()
		err := k8sutil.Reconcile(log, r.Client, o, r.NifiCluster, &r.NifiClusterCurrentStatus)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	// Drain removed nodes even while another node is waiting on its configuration.
	if err := r.reconcileNifiPodDelete(log); err != nil {
		return errors.WrapIf(err, "failed to reconcile removed nodes")
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
		// look up any existing PVCs for this node
		pvcs, err := getCreatedPVCForNode(r.Client, node.Id, r.NifiCluster.Namespace, r.NifiCluster.Name)
		// if pvcs is nil, then an error occurred. otherwise, it's just an empty list.
		if err != nil && pvcs == nil {
			return errors.WrapIfWithDetails(err, "failed to list PVCs")
		}

		// copy list of pvcs to detect those to delete
		pvcsToDelete := make([]corev1.PersistentVolumeClaim, len(pvcs))
		copy(pvcsToDelete, pvcs)

		for _, storage := range nodeConfig.StorageConfigs {
			var pvc *corev1.PersistentVolumeClaim
			pvcExists, existingPvc := r.storageConfigPVCExists(pvcs, storage.Name)
			// if the (volume reclaim policy is Retain and the PVC doesn't exist) OR the reclaim policy is Delete, then create it.
			if (storage.ReclaimPolicy == corev1.PersistentVolumeReclaimRetain && !pvcExists) ||
				storage.ReclaimPolicy == corev1.PersistentVolumeReclaimDelete {
				o := r.pvc(node.Id, storage, log)
				pvc = o.(*corev1.PersistentVolumeClaim)
			} else {
				// volume reclaim policy is Retain and the PVC exists
				log.Info("Volume reclaim policy is Retain. Re-using existing PVC",
					zap.String("clusterName", r.NifiCluster.Name),
					zap.Int32("nodeId", node.Id),
					zap.String("pvcName", existingPvc.Name),
					zap.String("storageName", storage.Name))
				// ensure we apply the reclaim policy label to handle PVC deletion properly for pre-existing PVCs
				existingPvc.Labels[nifiutil.NifiVolumeReclaimPolicyKey] = string(storage.ReclaimPolicy)
				pvc = existingPvc
			}
			err := r.reconcileNifiPVC(node.Id, log, pvc)
			if err != nil {
				return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", pvc.GetObjectKind().GroupVersionKind())
			}

			// remove pvc from the list of those to deleted
			for i, pvc := range pvcsToDelete {
				if pvcExists && pvc.Name == existingPvc.Name {
					pvcsToDelete = append(pvcsToDelete[:i], pvcsToDelete[i+1:]...)
					break
				}
			}
		}

		for _, pvc := range pvcsToDelete {
			// If this is a nifi data volume AND it has a Delete reclaim policy, then delete it. Otherwise, it is configured to be retained.
			if pvc.Labels[nifiutil.NifiDataVolumeMountKey] == "true" && pvc.Labels[nifiutil.NifiVolumeReclaimPolicyKey] == string(corev1.PersistentVolumeReclaimDelete) {
				err = r.Client.Delete(context.TODO(), &pvc)
				if err != nil {
					if apierrors.IsNotFound(err) {
						// can happen when node was not fully initialized and now is deleted
						log.Info(fmt.Sprintf("PVC for Node %s not found. Continue", node.Labels["nodeId"]))
					}

					return errors.WrapIfWithDetails(err, "could not delete pvc for node", "id", node.Labels["nodeId"])
				}
			} else {
				log.Debug("Not deleting PVC because it should be retained.", zap.String("nodeId", node.Labels["nodeId"]), zap.String("pvcName", pvc.Name))
			}
		}

		// re-lookup the PVCs after we've created any we need to create.
		pvcs, err = getCreatedPVCForNode(r.Client, node.Id, r.NifiCluster.Namespace, r.NifiCluster.Name)
		// remove pvcs that were delete
		for _, pvcToDelete := range pvcsToDelete {
			for i, pvc := range pvcs {
				if pvcToDelete.Name == pvc.Name {
					pvcs = append(pvcs[:i], pvcs[i+1:]...)
				}
			}
		}

		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to list PVCs")
		}

		nifiProperties := r.GetNifiPropertiesBase(node.Id)
		if err := validateTLSAutoReloadProperties(nifiProperties); err != nil {
			return err
		}
		if nifiProperties.IsTLSAutoReloadEnabled() && r.NifiCluster.Spec.ListenersConfig.SSLSecrets == nil {
			log.Warn("tlsAutoReload is enabled but TLS is not configured; no autoreload properties will be rendered",
				zap.String("clusterName", r.NifiCluster.Name),
				zap.Int32("nodeId", node.Id),
			)
		}

		o := r.secretConfig(node.Id, nodeConfig, serverPass, clientPass, superUsers, log)
		err = k8sutil.Reconcile(log, r.Client, o, r.NifiCluster, &r.NifiClusterCurrentStatus)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}

		if !r.NifiCluster.Spec.Service.HeadlessEnabled {
			o := r.service(node.Id, log)
			err := k8sutil.Reconcile(log, r.Client, o, r.NifiCluster, &r.NifiClusterCurrentStatus)
			if err != nil {
				return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
			}
		}
		o, err = r.pod(node, nodeConfig, pvcs, log)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to generate pod", "nodeId", node.Id)
		}
		pod := o.(*corev1.Pod)
		applyTLSAutoReloadAnnotations(pod, nifiProperties)

		err, isReady := r.reconcileNifiPod(log, pod)
		if err != nil {
			return err
		}
		if nodeState, ok := r.NifiCluster.Status.NodesState[pod.Labels["nodeId"]]; ok &&
			nodeState.PodIsReady != isReady {
			if err = k8sutil.UpdateNodeStatus(r.Client, []string{pod.Labels["nodeId"]}, r.NifiCluster, r.NifiClusterCurrentStatus, isReady, log); err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)",
					"id(s)", pod.Labels["nodeId"])
			}
		}
	}

	var err error
	// Reconcile external services
	services := r.externalServices(log)
	for _, o := range services {
		err = k8sutil.Reconcile(log, r.Client, o, r.NifiCluster, &r.NifiClusterCurrentStatus)
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
		err = k8sutil.Reconcile(log, r.Client, o, r.NifiCluster, &r.NifiClusterCurrentStatus)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	configManager := config.GetClientConfigManager(r.Client, v1.ClusterReference{
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

	if err := k8sutil.UpdateRootProcessGroupIdStatus(r.Client, r.NifiCluster, r.NifiClusterCurrentStatus, pgRootId, log); err != nil {
		return err
	}

	if clientConfig.UseSSL {
		if err := r.reconcileNifiUsersAndGroups(log); err != nil {
			return errors.WrapIf(err, "failed to reconcile resource")
		}
	}

	if r.NifiCluster.Spec.ReadOnlyConfig.MaximumTimerDrivenThreadCount != nil ||
		r.NifiCluster.Spec.ReadOnlyConfig.MaximumEventDrivenThreadCount != nil {
		if err := r.reconcileMaximumThreadCounts(); err != nil {
			return errors.WrapIf(err, "failed to reconcile ressource")
		}
	}

	if r.NifiCluster.Spec.GetMetricPort() != nil {
		if err := r.reconcilePrometheusReportingTask(); err != nil {
			return errors.WrapIf(err, "failed to reconcile ressource")
		}
	}

	log.Info("Successfully reconciled cluster",
		zap.String("component", componentName),
		zap.String("clusterName", r.NifiCluster.Name),
		zap.String("clusterNamespace", r.NifiCluster.Namespace))

	return nil
}

func (r *Reconciler) reconcileNifiPodDelete(log zap.Logger) error {
	podList := &corev1.PodList{}
	matchingLabels := client.MatchingLabels(nifiutil.LabelsForNifi(r.NifiCluster.Name))

	err := r.DirectClient.List(context.TODO(), podList,
		client.ListOption(client.InNamespace(r.NifiCluster.Namespace)), client.ListOption(matchingLabels))
	if err != nil {
		return errors.WrapIf(err, "failed to reconcile resource")
	}

	specIDs := v1.NodesToIdMap(r.NifiCluster.Spec.Nodes)
	present := make(map[string]bool)
	for _, pod := range podList.Items {
		id := pod.Labels["nodeId"]
		if id == "" {
			continue
		}
		present[id] = true
		nodeID, err := nifiutil.ParseStringToInt32(id)
		if err != nil {
			return errors.WrapIf(err, "invalid NiFi pod nodeId")
		}
		if _, desired := specIDs[nodeID]; desired {
			continue
		}
		if err := r.reconcileRemovedNode(log, &pod, nodeID); err != nil {
			return err
		}
	}

	// Wait for actual pod removal, including after a restart between DELETE and the status update.
	for id, state := range r.NifiCluster.Status.NodesState {
		action := state.GracefulActionState
		if present[id] || action.State != v1.GracefulDownscaleRunning || action.ActionStep != v1.RemovePodAction {
			continue
		}
		if state.PodIsReady {
			if err := k8sutil.UpdateNodeStatus(r.Client, []string{id}, r.NifiCluster, r.NifiClusterCurrentStatus, false, log); err != nil {
				return errors.WrapIfWithDetails(err, "could not clear removed pod readiness", "id", id)
			}
		}
		action.ActionStep = v1.RemovePodStatus
		if len(specIDs) == 0 {
			action.State = v1.GracefulDownscaleSucceeded
		}
		if err := k8sutil.UpdateNodeStatus(r.Client, []string{id}, r.NifiCluster, r.NifiClusterCurrentStatus, action, log); err != nil {
			return errors.WrapIfWithDetails(err, "could not record pod removal", "id", id)
		}
	}
	return nil
}

func (r *Reconciler) reconcileRemovedNode(log zap.Logger, node *corev1.Pod, nodeID int32) error {
	id := fmt.Sprint(nodeID)
	state := r.NifiCluster.Status.NodesState[id]
	ready := node.DeletionTimestamp == nil && k8sutil.PodReady(node)
	if state.PodIsReady != ready {
		if err := k8sutil.UpdateNodeStatus(r.Client, []string{id}, r.NifiCluster, r.NifiClusterCurrentStatus, ready, log); err != nil {
			return errors.WrapIfWithDetails(err, "could not update removed node readiness", "id", id)
		}
	}
	if node.DeletionTimestamp != nil {
		return nil
	}

	action := state.GracefulActionState
	if len(r.NifiCluster.Spec.Nodes) > 0 {
		if !action.State.IsDownscale() {
			if action.State == v1.GracefulUpscaleRunning || action.State == "" {
				log.Warn("Starting graceful downscale for orphaned NiFi pod", zap.String("pod", node.Name),
					zap.String("nodeId", id), zap.String("actionStep", string(action.ActionStep)))
				if r.Recorder != nil {
					r.Recorder.Eventf(r.NifiCluster, corev1.EventTypeWarning, "OrphanedNode",
						"Starting graceful downscale for pod %s (node %s, previous state %s)", node.Name, id, action.State)
				}
			}
			return k8sutil.UpdateNodeStatus(r.Client, []string{id}, r.NifiCluster, r.NifiClusterCurrentStatus,
				v1.GracefulActionState{State: v1.GracefulDownscaleRequired}, log)
		}
		if action.State != v1.GracefulDownscaleRunning || (action.ActionStep != v1.OffloadStatus && action.ActionStep != v1.RemovePodAction) {
			return nil
		}
	}
	if action.ActionStep != v1.RemovePodAction {
		action.State = v1.GracefulDownscaleRunning
		action.ActionStep = v1.RemovePodAction
		if err := k8sutil.UpdateNodeStatus(r.Client, []string{id}, r.NifiCluster, r.NifiClusterCurrentStatus, action, log); err != nil {
			return errors.WrapIfWithDetails(err, "could not record pod deletion", "id", id)
		}
	}

	for _, volume := range node.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		pvcFound := &corev1.PersistentVolumeClaim{}
		if err := r.Client.Get(context.TODO(),
			types.NamespacedName{
				Name:      volume.PersistentVolumeClaim.ClaimName,
				Namespace: r.NifiCluster.Namespace,
			},
			pvcFound,
		); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return errors.WrapIfWithDetails(err, "could not get pvc for node", "id", id)
		}

		// If this is a nifi data volume AND it has a Delete reclaim policy, then delete it. Otherwise, it is configured to be retained.
		if pvcFound.Labels[nifiutil.NifiDataVolumeMountKey] == "true" && pvcFound.Labels[nifiutil.NifiVolumeReclaimPolicyKey] == string(corev1.PersistentVolumeReclaimDelete) {
			err := r.Client.Delete(context.TODO(), &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
				Name:      volume.PersistentVolumeClaim.ClaimName,
				Namespace: r.NifiCluster.Namespace,
			}})
			if client.IgnoreNotFound(err) != nil {
				return errors.WrapIfWithDetails(err, "could not delete pvc for node", "id", id)
			}
		} else {
			log.Debug("Not deleting PVC because it should be retained.", zap.String("nodeId", id), zap.String("pvcName", pvcFound.Name))
		}
	}
	meta := metav1.ObjectMeta{Name: fmt.Sprintf(templates.NodeConfigTemplate+"-%s", r.NifiCluster.Name, id), Namespace: node.Namespace}
	if err := r.Client.Delete(context.TODO(), &corev1.Secret{ObjectMeta: meta}); client.IgnoreNotFound(err) != nil {
		return errors.WrapIfWithDetails(err, "could not delete config secret for node", "id", id)
	}
	if !r.NifiCluster.Spec.Service.HeadlessEnabled {
		meta.Name = nifiutil.ComputeNodeName(nodeID, r.NifiCluster.Name)
		if err := r.Client.Delete(context.TODO(), &corev1.Service{ObjectMeta: meta}); client.IgnoreNotFound(err) != nil {
			return errors.WrapIfWithDetails(err, "could not delete service for node", "id", id)
		}
	}
	log.Debug("Deleting pod", zap.String("pod", node.Name))
	return client.IgnoreNotFound(r.Client.Delete(context.TODO(), node))
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
	serverPass := string(serverSecret.Data[v1.PasswordKey])

	clientName := types.NamespacedName{Name: r.NifiCluster.GetNifiControllerUserIdentity(), Namespace: r.NifiCluster.Namespace}
	clientSecret := &corev1.Secret{}
	if err := r.Client.Get(context.TODO(), clientName, clientSecret); err != nil {
		if apierrors.IsNotFound(err) {
			return "", "", nil, errorfactory.New(errorfactory.ResourceNotReady{}, err, "client secret not ready")
		}
		return "", "", nil, errors.WrapIfWithDetails(err, "failed to get client secret")
	}
	clientPass := string(clientSecret.Data[v1.PasswordKey])

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

func (r *Reconciler) reconcileNifiPVC(nodeId int32, log zap.Logger, desiredPVC *corev1.PersistentVolumeClaim) error {
	var currentPVC = desiredPVC.DeepCopy()
	desiredType := reflect.TypeOf(desiredPVC)
	log.Debug("searching for pvc with label because name is empty",
		zap.String("nifiCluster", r.NifiCluster.Name),
		zap.String("nodeId", desiredPVC.Labels["nodeId"]),
		zap.String("kind", desiredType.String()))

	pvcList, err := getCreatedPVCForNode(r.Client, nodeId, currentPVC.Namespace, r.NifiCluster.Name)
	if err != nil && len(pvcList) == 0 {
		return errorfactory.New(errorfactory.APIFailure{}, err, "getting resource failed", "kind", desiredType)
	}
	mountPath := currentPVC.Annotations["mountPath"]
	storageName := currentPVC.Annotations["storageName"]

	// Creating the first PersistentVolume For Pod
	if len(pvcList) == 0 {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPVC); err != nil {
			return errors.WrapIf(err, "could not apply last state to annotation")
		}
		if err := r.Client.Create(context.TODO(), desiredPVC); err != nil {
			return errorfactory.New(errorfactory.APIFailure{}, err, "creating resource failed", "kind", desiredType)
		}
		log.Info("Persistent volume created",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.String("pvcName", desiredPVC.Name),
			zap.String("pvcNamespace", desiredPVC.Namespace))
		return nil
	}
	alreadyCreated := false
	for _, pvc := range pvcList {
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
			annotations := desiredPVC.Annotations
			desiredPVC = currentPVC.DeepCopy()
			desiredPVC.Spec.Resources.Requests = resReq
			desiredPVC.Labels = labels
			desiredPVC.Annotations = annotations
			desiredPVC.SetOwnerReferences([]metav1.OwnerReference{templates.ClusterOwnerReference(r.NifiCluster)})

			if err := r.Client.Update(context.TODO(), desiredPVC); err != nil {
				return errorfactory.New(errorfactory.APIFailure{}, err, "updating resource failed", "kind", desiredType)
			}
			log.Debug("persistent volume updated",
				zap.String("clusterName", r.NifiCluster.Name),
				zap.String("pvcName", desiredPVC.Name),
				zap.String("pvcNamespace", desiredPVC.Namespace))
		}
	}
	return nil
}

func isDesiredStorageValueInvalid(desired, current *corev1.PersistentVolumeClaim) bool {
	return desired.Spec.Resources.Requests.Storage().Value() < current.Spec.Resources.Requests.Storage().Value()
}

func (r *Reconciler) reconcileNifiPod(log zap.Logger, desiredPod *corev1.Pod) (error, bool) {
	currentPod := desiredPod.DeepCopy()
	desiredType := reflect.TypeOf(desiredPod)

	log.Debug("searching for pod with label because name is empty",
		zap.String("clusterName", r.NifiCluster.Name),
		zap.String("nodeId", desiredPod.Labels["nodeId"]),
		zap.String("kind", desiredType.String()))

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

	created := len(podList.Items) == 0
	switch len(podList.Items) {
	case 0:
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPod); err != nil {
			return errors.WrapIf(err, "could not apply last state to annotation"), false
		}
		if err := r.Client.Create(context.TODO(), desiredPod); err != nil {
			return errorfactory.New(errorfactory.APIFailure{}, err, "creating resource failed", "kind", desiredType), false
		}
		currentPod = desiredPod
		log.Info("Pod created", zap.String("clusterName", r.NifiCluster.Name), zap.String("podName", currentPod.Name))
	case 1:
		currentPod = podList.Items[0].DeepCopy()
	default:
		return errorfactory.New(errorfactory.TooManyResources{}, errors.New("reconcile failed"),
			"more than one matching pod found", "labels", matchingLabels), false
	}

	// Pod creation and status updates are separate requests. Resume initialization after a conflict or restart.
	id := currentPod.Labels["nodeId"]
	state := r.NifiCluster.Status.NodesState[id]
	if created || state.GracefulActionState.State == "" {
		if created || state.ConfigurationState == "" {
			if err := k8sutil.UpdateNodeStatus(r.Client, []string{id}, r.NifiCluster, r.NifiClusterCurrentStatus, v1.ConfigInSync, log); err != nil {
				return errors.WrapIf(err, "could not initialize node configuration state"), false
			}
		}
		if created || state.CreationTime == nil {
			if err := k8sutil.UpdateNodeStatus(r.Client, []string{id}, r.NifiCluster, r.NifiClusterCurrentStatus, currentPod.CreationTimestamp, log); err != nil {
				return errors.WrapIf(err, "could not record pod creation time"), false
			}
		}
		if state.GracefulActionState.State != v1.GracefulUpscaleSucceeded {
			if err := k8sutil.UpdateNodeStatus(r.Client, []string{id}, r.NifiCluster, r.NifiClusterCurrentStatus,
				v1.GracefulActionState{State: v1.GracefulUpscaleRequired}, log); err != nil {
				return errors.WrapIf(err, "could not request node upscale"), false
			}
		}
		return nil, k8sutil.PodReady(currentPod)
	}

	if err == nil {
		// k8s-objectmatcher options
		opts := []patch.CalculateOption{
			patch.IgnoreStatusFields(),
			patch.IgnoreVolumeClaimTemplateTypeMetaAndStatus(),
			patch.IgnorePDBSelector(),
		}
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
		// If there are extra initContainers from webhook injections we need to add them
		if len(currentPod.Spec.InitContainers) > len(desiredPod.Spec.InitContainers) {
			desiredPod.Spec.InitContainers = append(desiredPod.Spec.InitContainers, currentPod.Spec.InitContainers...)
			uniqueContainers := []corev1.Container{}
			keys := make(map[string]bool)
			for _, c := range desiredPod.Spec.InitContainers {
				if _, value := keys[c.Name]; !value {
					keys[c.Name] = true
					uniqueContainers = append(uniqueContainers, c)
				}
			}
			desiredPod.Spec.InitContainers = uniqueContainers
		}
		// If there are extra containers from webhook injections we need to add them
		if len(currentPod.Spec.Containers) > len(desiredPod.Spec.Containers) {
			desiredPod.Spec.Containers = append(desiredPod.Spec.Containers, currentPod.Spec.Containers...)
			uniqueContainers := []corev1.Container{}
			keys := make(map[string]bool)
			for _, c := range desiredPod.Spec.Containers {
				if _, value := keys[c.Name]; !value {
					keys[c.Name] = true
					uniqueContainers = append(uniqueContainers, c)
				}
			}
			desiredPod.Spec.Containers = uniqueContainers
		}
		// Remove problematic fields if istio
		if _, ok := currentPod.Annotations["istio.io/rev"]; ok {
			// Prometheus scrape port is overridden by istio injection
			delete(currentPod.Annotations, "prometheus.io/port")
			delete(desiredPod.Annotations, "prometheus.io/port")
			// Liveness probe port is overridden by istio injection
			desiredContainer := corev1.Container{}
			for _, c := range desiredPod.Spec.Containers {
				if c.Name == "nifi" {
					desiredContainer = c
				}
			}
			currentContainers := []corev1.Container{}
			for _, c := range currentPod.Spec.Containers {
				if c.Name == "nifi" {
					c.LivenessProbe = desiredContainer.LivenessProbe
				}
				currentContainers = append(currentContainers, c)
			}
			currentPod.Spec.Containers = currentContainers
		}
		// Patch image name if Zarf has modified the pod spec
		if _, ok := currentPod.Labels["zarf-agent"]; ok {
			var oldImage, currentImage string
			for _, c := range currentPod.Spec.Containers {
				if c.Name == "nifi" {
					imageChunks := strings.Split(c.Image, "/")
					oldTag := strings.Split(imageChunks[len(imageChunks)-1], "-zarf")[0]
					oldRepoChunks := imageChunks[1 : len(imageChunks)-1]
					oldImage = fmt.Sprintf("%s/%s", strings.Join(oldRepoChunks, "/"), oldTag)
					currentImage = c.Image
				}
			}
			log.Debug("Patching Nifi container image for Zarf",
				zap.String("current", currentImage),
				zap.String("original", oldImage))
			desiredContainers := []corev1.Container{}
			for _, c := range desiredPod.Spec.Containers {
				if c.Name == "nifi" {
					// If the incoming image matches the spec from before the zarf patch then the pod is in sync
					if c.Image == oldImage {
						// We want to prevent a reconcile loop by setting the incoming image to the zarf image spec
						c.Image = currentImage
					}
				}
				desiredContainers = append(desiredContainers, c)
			}
			desiredPod.Spec.Containers = desiredContainers
		}
		// Linkerd – normalize desired against current before diffing
		meshAwareMerge(desiredPod, currentPod)
		patchResult, err := patch.DefaultPatchMaker.Calculate(currentPod, desiredPod, opts...)
		if err != nil {
			log.Error("could not match pod objects",
				zap.String("clusterName", r.NifiCluster.Name),
				zap.String("kind", desiredType.String()),
				zap.Error(err))
		} else if patchResult.IsEmpty() {
			if !k8sutil.IsPodTerminatedOrShutdown(currentPod) &&
				r.NifiCluster.Status.NodesState[currentPod.Labels["nodeId"]].ConfigurationState == v1.ConfigInSync {
				log.Debug("pod resource is in sync",
					zap.String("clusterName", r.NifiCluster.Name),
					zap.String("podName", currentPod.Name))

				return nil, k8sutil.PodReady(currentPod)
			}
		} else {
			var pr map[string]interface{}
			if err := json.Unmarshal(patchResult.Patch, &pr); err == nil {
				if spec, ok := pr["spec"].(map[string]interface{}); ok && len(spec) > 0 {
					onlyOrder := true
					for k := range spec {
						if !strings.HasPrefix(k, "$setElementOrder/") {
							onlyOrder = false
							break
						}
					}
					if onlyOrder {
						if !k8sutil.IsPodTerminatedOrShutdown(currentPod) &&
							r.NifiCluster.Status.NodesState[currentPod.Labels["nodeId"]].ConfigurationState == v1.ConfigInSync {
							log.Debug("pod resource is in sync (order-only diff ignored)",
								zap.String("clusterName", r.NifiCluster.Name),
								zap.String("podName", currentPod.Name))
							return nil, k8sutil.PodReady(currentPod)
						}
					}
				}
			}
			log.Debug("resource diffs",
				zap.String("patch", string(patchResult.Patch)),
				zap.String("current", string(patchResult.Current)),
				zap.String("modified", string(patchResult.Modified)),
				zap.String("original", string(patchResult.Original)))
		}

		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPod); err != nil {
			return errors.WrapIf(err, "could not apply last state to annotation"), false
		}

		if !k8sutil.IsPodTerminatedOrShutdown(currentPod) {
			if r.NifiCluster.Status.State != v1.NifiClusterRollingUpgrading {
				if err := k8sutil.UpdateCRStatus(r.Client, r.NifiCluster, r.NifiClusterCurrentStatus, v1.NifiClusterRollingUpgrading, log); err != nil {
					return errorfactory.New(errorfactory.StatusUpdateError{},
						err, "setting state to rolling upgrade failed"), false
				}
			}

			if r.NifiCluster.Status.State == v1.NifiClusterRollingUpgrading {
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

		log.Info(fmt.Sprintf("Deleting pod %s", currentPod.Name))
		err = r.Client.Delete(context.TODO(), currentPod)
		if err != nil {
			return errorfactory.New(errorfactory.APIFailure{},
				err, "deleting resource failed", "kind", desiredType), false
		}
	}

	return nil, k8sutil.PodReady(currentPod)
}

func (r *Reconciler) reconcileNifiUsersAndGroups(log zap.Logger) error {
	controllerNamespacedName := types.NamespacedName{
		Name: r.NifiCluster.GetNifiControllerUserIdentity(), Namespace: r.NifiCluster.Namespace}

	managedUsers := append(r.NifiCluster.Spec.ManagedAdminUsers, r.NifiCluster.Spec.ManagedReaderUsers...)
	var users []*v1.NifiUser
	pFalse := false
	for _, managedUser := range managedUsers {
		users = append(users, &v1.NifiUser{
			ObjectMeta: templates.ObjectMeta(
				fmt.Sprintf("%s.%s", r.NifiCluster.Name, managedUser.Name),
				pkicommon.LabelsForNifiPKI(r.NifiCluster.Name), r.NifiCluster,
			),
			Spec: v1.NifiUserSpec{
				Identity:   managedUser.GetIdentity(),
				CreateCert: &pFalse,
				ClusterRef: v1.ClusterReference{
					Name:      r.NifiCluster.Name,
					Namespace: r.NifiCluster.Namespace,
				},
			},
		})
	}

	var managedAdminUserRef []v1.UserReference
	for _, user := range r.NifiCluster.Spec.ManagedAdminUsers {
		managedAdminUserRef = append(managedAdminUserRef, v1.UserReference{Name: fmt.Sprintf("%s.%s", r.NifiCluster.Name, user.Name)})
	}

	var managedReaderUserRef []v1.UserReference
	for _, user := range r.NifiCluster.Spec.ManagedReaderUsers {
		managedReaderUserRef = append(managedReaderUserRef, v1.UserReference{Name: fmt.Sprintf("%s.%s", r.NifiCluster.Name, user.Name)})
	}

	var managedNodeUserRef []v1.UserReference
	for _, node := range r.NifiCluster.Spec.Nodes {
		managedNodeUserRef = append(managedNodeUserRef, v1.UserReference{Name: pkicommon.GetNodeUserName(r.NifiCluster, node.Id)})
	}

	groups := []*v1.NifiUserGroup{
		// Managed admins
		{
			ObjectMeta: templates.ObjectMeta(
				fmt.Sprintf("%s.managed-admins", r.NifiCluster.Name),
				pkicommon.LabelsForNifiPKI(r.NifiCluster.Name), r.NifiCluster,
			),
			Spec: v1.NifiUserGroupSpec{
				ClusterRef: v1.ClusterReference{
					Name:      r.NifiCluster.Name,
					Namespace: r.NifiCluster.Namespace,
				},
				UsersRef: append(managedAdminUserRef, v1.UserReference{
					Name:      controllerNamespacedName.Name,
					Namespace: controllerNamespacedName.Namespace,
				},
				),
				AccessPolicies: []v1.AccessPolicy{
					// Global
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.FlowAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.FlowAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ControllerAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.ControllerAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ParameterContextAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.ParameterContextAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ProvenanceAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.ProvenanceAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.RestrictedComponentsAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.RestrictedComponentsAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.PoliciesAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.PoliciesAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.TenantsAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.TenantsAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.SiteToSiteAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.SystemAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.SiteToSiteAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ProxyAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.ProxyAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.CountersAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.CountersAccessPolicyResource},
					// Root process group
					{Type: v1.ComponentAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ComponentsAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					{Type: v1.ComponentAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.ComponentsAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					{Type: v1.ComponentAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.OperationAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					{Type: v1.ComponentAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ProvenanceDataAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					{Type: v1.ComponentAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.DataAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					{Type: v1.ComponentAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.DataAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					// {Type: v1.ComponentAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.PoliciesComponentAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					// {Type: v1.ComponentAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.PoliciesComponentAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					// {Type: v1.ComponentAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.DataTransferAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					// {Type: v1.ComponentAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.DataTransferAccessPolicyResource, ComponentType: v1.ProcessGroupType},
				},
			},
		},
		// Managed Readers
		{
			ObjectMeta: templates.ObjectMeta(
				fmt.Sprintf("%s.managed-readers", r.NifiCluster.Name),
				pkicommon.LabelsForNifiPKI(r.NifiCluster.Name), r.NifiCluster,
			),
			Spec: v1.NifiUserGroupSpec{
				ClusterRef: v1.ClusterReference{
					Name:      r.NifiCluster.Name,
					Namespace: r.NifiCluster.Namespace,
				},
				UsersRef: managedReaderUserRef,
				AccessPolicies: []v1.AccessPolicy{
					// Global
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.FlowAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ControllerAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ParameterContextAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ProvenanceAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.RestrictedComponentsAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.PoliciesAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.TenantsAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.SiteToSiteAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.SystemAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ProxyAccessPolicyResource},
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.CountersAccessPolicyResource},
					// Root process group
					{Type: v1.ComponentAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ComponentsAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					{Type: v1.ComponentAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.OperationAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					{Type: v1.ComponentAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ProvenanceDataAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					{Type: v1.ComponentAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.DataAccessPolicyResource, ComponentType: v1.ProcessGroupType},
				},
			},
		},
		// Managed Nodes
		{
			ObjectMeta: templates.ObjectMeta(
				fmt.Sprintf("%s.managed-nodes", r.NifiCluster.Name),
				pkicommon.LabelsForNifiPKI(r.NifiCluster.Name), r.NifiCluster,
			),
			Spec: v1.NifiUserGroupSpec{
				ClusterRef: v1.ClusterReference{
					Name:      r.NifiCluster.Name,
					Namespace: r.NifiCluster.Namespace,
				},
				UsersRef: managedNodeUserRef,
				AccessPolicies: []v1.AccessPolicy{
					// Global
					{Type: v1.GlobalAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ProxyAccessPolicyResource},
					// Root process group
					{Type: v1.ComponentAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.ProvenanceDataAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					{Type: v1.ComponentAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.ProvenanceDataAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					{Type: v1.ComponentAccessPolicyType, Action: v1.ReadAccessPolicyAction, Resource: v1.DataAccessPolicyResource, ComponentType: v1.ProcessGroupType},
					{Type: v1.ComponentAccessPolicyType, Action: v1.WriteAccessPolicyAction, Resource: v1.DataAccessPolicyResource, ComponentType: v1.ProcessGroupType},
				},
			},
		},
	}

	for _, user := range users {
		if err := k8sutil.Reconcile(log, r.Client, user, r.NifiCluster, &r.NifiClusterCurrentStatus); err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", user.GetObjectKind().GroupVersionKind())
		}
	}

	for _, group := range groups {
		if err := k8sutil.Reconcile(log, r.Client, group, r.NifiCluster, &r.NifiClusterCurrentStatus); err != nil {
			return errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", group.GetObjectKind().GroupVersionKind())
		}
	}

	return nil
}

func (r *Reconciler) reconcilePrometheusReportingTask() error {
	var err error

	patchNifiCluster := client.MergeFrom(r.NifiCluster.DeepCopy())

	configManager := config.GetClientConfigManager(r.Client, v1.ClusterReference{
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
		if !reflect.DeepEqual(r.NifiCluster.Status, r.NifiClusterCurrentStatus) {
			if err := r.Client.Status().Patch(context.TODO(), r.NifiCluster, patchNifiCluster); err != nil {
				return errors.WrapIfWithDetails(err, "failed to update PrometheusReportingTask status")
			}
		}
	}

	// Sync prometheus reporting task resource with NiFi side component
	status, err := reportingtask.SyncReportingTask(clientConfig, r.NifiCluster)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to sync PrometheusReportingTask")
	}

	r.NifiCluster.Status.PrometheusReportingTask = *status
	if !reflect.DeepEqual(r.NifiCluster.Status, r.NifiClusterCurrentStatus) {
		if err := r.Client.Status().Patch(context.TODO(), r.NifiCluster, patchNifiCluster); err != nil {
			return errors.WrapIfWithDetails(err, "failed to update PrometheusReportingTask status")
		}
	}
	return nil
}

func (r *Reconciler) reconcileMaximumThreadCounts() error {
	configManager := config.GetClientConfigManager(r.Client, v1.ClusterReference{
		Namespace: r.NifiCluster.Namespace,
		Name:      r.NifiCluster.Name,
	})
	clientConfig, err := configManager.BuildConfig()
	if err != nil {
		return err
	}

	// Sync Maximum Timer Driven Thread Count and Maximum Event Driven Thread Count with NiFi side component
	err = controllersettings.SyncConfiguration(clientConfig, r.NifiCluster)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to sync MaximumThreadCount configuration")
	}

	return nil
}
