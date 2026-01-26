package nifi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	componentName                   = "nifi"
	podServerCertHashAnnotation     = "nifikop.konpyutaika.com/server-cert-hash"
	podClientCertHashAnnotation     = "nifikop.konpyutaika.com/client-cert-hash"
	podServerCertNotAfterAnnotation = "nifikop.konpyutaika.com/server-cert-not-after"
	podClientCertNotAfterAnnotation = "nifikop.konpyutaika.com/client-cert-not-after"

	nodeSecretVolumeMount = "node-config"
	nodeTmp               = "node-tmp"

	serverKeystoreVolume = "server-ks-files"
	serverKeystorePath   = "/var/run/secrets/java.io/keystores/server"
	clientKeystoreVolume = "client-ks-files"
	clientKeystorePath   = "/var/run/secrets/java.io/keystores/client"
)

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

func secretCertMaterialHash(s *corev1.Secret) string {
	if s == nil || s.Data == nil {
		return ""
	}

	keys := []string{
		corev1.TLSCertKey,
		corev1.TLSPrivateKeyKey,
		"ca.crt",
		v1.PasswordKey,
		"keystore.jks",
		"truststore.jks",
	}

	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte{0})

		if b, ok := s.Data[k]; ok {
			h.Write(b)
		}

		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func podAnn(p *corev1.Pod, key string) string {
	if p == nil || p.Annotations == nil {
		return ""
	}
	return p.Annotations[key]
}

func podCertHashesDiffer(current, desired *corev1.Pod) bool {
	return podAnn(current, podServerCertHashAnnotation) != podAnn(desired, podServerCertHashAnnotation) ||
		podAnn(current, podClientCertHashAnnotation) != podAnn(desired, podClientCertHashAnnotation)
}

func parseRFC3339(s string) (time.Time, bool) {
	if strings.TrimSpace(s) == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func minNonZeroTime(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}
	if a.Before(b) {
		return a
	}
	return b
}

func parseDay(s string) (time.Weekday, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "mon", "monday":
		return time.Monday, true
	case "tue", "tues", "tuesday":
		return time.Tuesday, true
	case "wed", "wednesday":
		return time.Wednesday, true
	case "thu", "thur", "thurs", "thursday":
		return time.Thursday, true
	case "fri", "friday":
		return time.Friday, true
	case "sat", "saturday":
		return time.Saturday, true
	case "sun", "sunday":
		return time.Sunday, true
	default:
		return 0, false
	}
}

func prevWeekday(w time.Weekday) time.Weekday {
	if w == time.Sunday {
		return time.Saturday
	}
	return w - 1
}

func parseClockMinutes(s string) (int, error) {
	// expects "HH:MM"
	t, err := time.Parse("15:04", strings.TrimSpace(s))
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}

func weekdayInList(days []v1.Weekday, wd time.Weekday) bool {
	if len(days) == 0 {
		return true
	}
	for _, d := range days {
		w, ok := parseDay(string(d))
		if ok && w == wd {
			return true
		}
	}
	return false
}

func nowInAnyCertWindow(now time.Time, tz string, windows []v1.CertRotationWindow) (bool, error) {
	if tz == "" || len(windows) == 0 {
		return false, nil
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return false, err
	}

	local := now.In(loc)
	wd := local.Weekday()
	mins := local.Hour()*60 + local.Minute()

	for _, w := range windows {
		startM, err := parseClockMinutes(w.Start)
		if err != nil {
			return false, err
		}
		endM, err := parseClockMinutes(w.End)
		if err != nil {
			return false, err
		}

		if startM == endM {
			return false, fmt.Errorf("invalid certRotation window: start == end (start=%q end=%q days=%v)", w.Start, w.End, w.Days)
		}

		if startM < endM {
			if weekdayInList(w.Days, wd) && mins >= startM && mins < endM {
				return true, nil
			}
			continue
		}

		// spans midnight (e.g. 22:00 -> 06:00)
		okToday := weekdayInList(w.Days, wd)
		okPrev := weekdayInList(w.Days, prevWeekday(wd))

		if okToday && mins >= startM {
			return true, nil
		}
		if okPrev && mins < endM {
			return true, nil
		}
	}

	return false, nil
}

func isCertOnlyPatch(patchBytes []byte) bool {
	if len(patchBytes) == 0 {
		return false
	}

	var pr map[string]interface{}
	if err := json.Unmarshal(patchBytes, &pr); err != nil {
		return false
	}

	if len(pr) == 0 {
		return false
	}

	for k := range pr {
		if k != "metadata" && k != "spec" {
			return false
		}
	}

	if specAny, ok := pr["spec"]; ok {
		spec, ok := specAny.(map[string]interface{})
		if !ok {
			return false
		}
		for k := range spec {
			if !strings.HasPrefix(k, "$setElementOrder/") {
				return false
			}
		}
	}

	// Allow metadata only if it contains ONLY annotations
	if metaAny, ok := pr["metadata"]; ok {
		meta, ok := metaAny.(map[string]interface{})
		if !ok {
			return false
		}
		for k := range meta {
			if k != "annotations" {
				return false
			}
		}

		annAny, ok := meta["annotations"]
		if !ok {
			return false
		}
		ann, ok := annAny.(map[string]interface{})
		if !ok {
			return false
		}

		allowed := map[string]struct{}{
			podServerCertHashAnnotation:                        {},
			podClientCertHashAnnotation:                        {},
			podServerCertNotAfterAnnotation:                    {},
			podClientCertNotAfterAnnotation:                    {},
			"banzaicloud.com/last-applied":                     {},
			"kubectl.kubernetes.io/last-applied-configuration": {},
		}
		for k := range ann {
			if _, ok := allowed[k]; !ok {
				return false
			}
		}
	}

	return true
}

// returns (delay, reason, error). If delay=true => do NOT restart now.
func (r *Reconciler) shouldDelayCertRotationRestart(
	log zap.Logger,
	currentPod, desiredPod *corev1.Pod,
	patchBytes []byte,
) (bool, string, error) {
	p := r.NifiCluster.Spec.CertRotation
	if p == nil || p.Strategy != v1.CertRotationWindowed {
		return false, "", nil
	}

	tz := strings.TrimSpace(p.Timezone)
	if tz == "" || len(p.Windows) == 0 {
		log.Warn("certRotation strategy is Windowed but timezone/windows are not set; falling back to Immediate behavior",
			zap.String("timezone", p.Timezone),
			zap.Int("windows", len(p.Windows)),
		)
		return false, "", nil
	}

	if !podCertHashesDiffer(currentPod, desiredPod) {
		return false, "", nil
	}

	if !isCertOnlyPatch(patchBytes) {
		return false, "", nil
	}

	now := time.Now().UTC()
	inWin, err := nowInAnyCertWindow(now, tz, p.Windows)
	if err != nil {
		log.Warn("invalid certRotation window config; falling back to immediate restart", zap.Error(err))
		return false, "", nil
	}
	if inWin {
		return false, "", nil
	}

	urgentBefore := 24 * time.Hour
	if p.UrgentBefore != nil {
		urgentBefore = p.UrgentBefore.Duration
	}

	srvExp, _ := parseRFC3339(podAnn(currentPod, podServerCertNotAfterAnnotation))
	cliExp, _ := parseRFC3339(podAnn(currentPod, podClientCertNotAfterAnnotation))
	exp := minNonZeroTime(srvExp, cliExp)

	if exp.IsZero() {
		dsrv, _ := parseRFC3339(podAnn(desiredPod, podServerCertNotAfterAnnotation))
		dcli, _ := parseRFC3339(podAnn(desiredPod, podClientCertNotAfterAnnotation))
		dexp := minNonZeroTime(dsrv, dcli)

		if dexp.IsZero() {
			log.Warn("missing cert expiry annotations; delaying until maintenance window",
				zap.String("podName", currentPod.Name),
			)
			return true, "missing pod expiry annotations (pre-feature pod); delaying until window", nil
		}

		ttl := dexp.Sub(now)
		if ttl <= urgentBefore {
			log.Warn("cert rotation urgent (desired expiry); restarting outside maintenance window",
				zap.String("podName", currentPod.Name),
				zap.Duration("timeToExpiry", ttl),
				zap.Duration("urgentBefore", urgentBefore),
			)
			return false, "", nil
		}

		log.Warn("missing pod expiry annotations; delaying until maintenance window",
			zap.String("podName", currentPod.Name),
			zap.Duration("timeToExpiryFromDesired", ttl),
			zap.Duration("urgentBefore", urgentBefore),
		)
		return true, "missing pod expiry annotations (pre-feature pod); delaying until window", nil
	}

	ttl := exp.Sub(now)
	if ttl <= urgentBefore {
		log.Warn("cert rotation urgent (loaded expiry); restarting outside maintenance window",
			zap.String("podName", currentPod.Name),
			zap.Duration("timeToExpiry", ttl),
			zap.Duration("urgentBefore", urgentBefore),
		)
		return false, "", nil
	}

	return true, fmt.Sprintf("outside maintenance window; timeToExpiry=%s urgentBefore=%s", ttl, urgentBefore), nil
}

// Reconciler implements the Component Reconciler.
type Reconciler struct {
	resources.Reconciler
	Scheme *runtime.Scheme
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

	var pendingRolling error
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
	if r.NifiCluster.Spec.ListenersConfig != nil && r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil {
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

	for _, node := range r.NifiCluster.Spec.Nodes {
		// We need to grab names for servers and client in case user is enabling ACLs
		// That way we can continue to manage dataflows and users
		serverPass, clientPass, superUsers, serverHash, clientHash, serverNotAfter, clientNotAfter, err := r.getServerAndClientDetails(node.Id)
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
		o = r.pod(node, nodeConfig, pvcs, log)
		pod := o.(*corev1.Pod)

		if r.NifiCluster.Spec.ListenersConfig != nil && r.NifiCluster.Spec.ListenersConfig.SSLSecrets != nil {
			if pod.Annotations == nil {
				pod.Annotations = map[string]string{}
			}

			pod.Annotations[podServerCertHashAnnotation] = serverHash
			pod.Annotations[podClientCertHashAnnotation] = clientHash

			pod.Annotations[podServerCertNotAfterAnnotation] = serverNotAfter.UTC().Format(time.RFC3339)
			pod.Annotations[podClientCertNotAfterAnnotation] = clientNotAfter.UTC().Format(time.RFC3339)
		}

		err, isReady := r.reconcileNifiPod(log, pod)
		if err != nil {
			if isReconcileRollingUpgradeErr(err) {
				pendingRolling = err
				break
			}
			return err
		}

		if nodeState, ok := r.NifiCluster.Status.NodesState[pod.Labels["nodeId"]]; ok &&
			nodeState.PodIsReady != isReady {
			if err = k8sutil.UpdateNodeStatus(
				r.Client,
				[]string{pod.Labels["nodeId"]},
				r.NifiCluster,
				r.NifiClusterCurrentStatus,
				isReady,
				log,
			); err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", pod.Labels["nodeId"])
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

	// Handle Pod delete
	err = r.reconcileNifiPodDelete(log)
	if err != nil {
		return errors.WrapIf(err, "failed to reconcile resource")
	}

	if pendingRolling != nil {
		log.Info("Requeueing due to pending rolling upgrade", zap.Error(pendingRolling))
		return pendingRolling
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
		zap.String("clusterNamespace", r.NifiCluster.Namespace),
	)

	return nil
}

func (r *Reconciler) reconcileNifiPodDelete(log zap.Logger) error {
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
					if nState != v1.GracefulDownscaleRunning && (nState == v1.GracefulUpscaleSucceeded ||
						nState == v1.GracefulUpscaleRequired) {
						nodesPendingGracefulDownscale = append(nodesPendingGracefulDownscale, deletedNodesId[i])
					}
				}
			}

			if len(nodesPendingGracefulDownscale) > 0 {
				err = k8sutil.UpdateNodeStatus(r.Client, nodesPendingGracefulDownscale, r.NifiCluster, r.NifiClusterCurrentStatus,
					v1.GracefulActionState{
						State: v1.GracefulDownscaleRequired,
					}, log)
				if err != nil {
					return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)",
						strings.Join(nodesPendingGracefulDownscale, ","))
				}
			}
		}

		for _, node := range deletedNodes {
			if node.ObjectMeta.DeletionTimestamp != nil {
				log.Info("Node is already on terminating state",
					zap.String("nodeId", node.Labels["nodeId"]))
				continue
			}

			if len(r.NifiCluster.Spec.Nodes) > 0 {
				if nodeState, ok := r.NifiCluster.Status.NodesState[node.Labels["nodeId"]]; ok &&
					nodeState.GracefulActionState.ActionStep != v1.OffloadStatus && nodeState.GracefulActionState.ActionStep != v1.RemovePodAction {
					if nodeState.GracefulActionState.State == v1.GracefulDownscaleRunning {
						log.Info("Nifi task is still running for node",
							zap.String("nodeId", node.Labels["nodeId"]),
							zap.String("ActionStep", string(nodeState.GracefulActionState.ActionStep)))
					}
					continue
				}
			}

			err = k8sutil.UpdateNodeStatus(r.Client, []string{node.Labels["nodeId"]}, r.NifiCluster, r.NifiClusterCurrentStatus,
				v1.GracefulActionState{ActionStep: v1.RemovePodAction, State: v1.GracefulDownscaleRunning,
					TaskStarted: r.NifiCluster.Status.NodesState[node.Labels["nodeId"]].GracefulActionState.TaskStarted}, log)

			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", node.Labels["nodeId"])
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
					return errors.WrapIfWithDetails(err, "could not get pvc for node", "id", node.Labels["nodeId"])
				}

				// If this is a nifi data volume AND it has a Delete reclaim policy, then delete it. Otherwise, it is configured to be retained.
				if pvcFound.Labels[nifiutil.NifiDataVolumeMountKey] == "true" && pvcFound.Labels[nifiutil.NifiVolumeReclaimPolicyKey] == string(corev1.PersistentVolumeReclaimDelete) {
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
				} else {
					log.Debug("Not deleting PVC because it should be retained.", zap.String("nodeId", node.Labels["nodeId"]), zap.String("pvcName", pvcFound.Name))
				}
			}

			log.Debug("Deleting pod.", zap.String("pod", node.Name))
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

			err = k8sutil.UpdateNodeStatus(r.Client, []string{node.Labels["nodeId"]}, r.NifiCluster, r.NifiClusterCurrentStatus,
				v1.GracefulActionState{
					ActionStep:  v1.RemovePodStatus,
					State:       v1.GracefulDownscaleSucceeded,
					TaskStarted: r.NifiCluster.Status.NodesState[node.Labels["nodeId"]].GracefulActionState.TaskStarted},
				log)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update status for node(s)", "id(s)", node.Labels["nodeId"])
			}
		}
	}

	return nil
}

func arePodsAlreadyDeleted(pods []corev1.Pod, log zap.Logger) bool {
	for _, node := range pods {
		if node.ObjectMeta.DeletionTimestamp == nil {
			return false
		}
		log.Info("Node is already on terminating state",
			zap.String("nodeId", node.Labels["nodeId"]))
	}
	return true
}

func requireTLSSecretReady(secret *corev1.Secret, which string) error {
	if secret == nil || secret.Data == nil {
		return errorfactory.New(
			errorfactory.ResourceNotReady{},
			errors.New("secret data not populated yet"),
			fmt.Sprintf("%s secret not ready", which),
		)
	}

	if len(secret.Data[corev1.TLSCertKey]) == 0 {
		return errorfactory.New(
			errorfactory.ResourceNotReady{},
			errors.New("tls.crt missing/empty"),
			fmt.Sprintf("%s secret not ready", which),
		)
	}

	if secret.Type == corev1.SecretTypeTLS {
		if len(secret.Data[corev1.TLSPrivateKeyKey]) == 0 {
			return errorfactory.New(
				errorfactory.ResourceNotReady{},
				errors.New("tls.key missing/empty"),
				fmt.Sprintf("%s secret not ready", which),
			)
		}
	}

	if len(secret.Data[v1.PasswordKey]) == 0 {
		return errorfactory.New(
			errorfactory.ResourceNotReady{},
			errors.New("password missing/empty"),
			fmt.Sprintf("%s secret not ready", which),
		)
	}

	return nil
}

func (r *Reconciler) getServerAndClientDetails(nodeId int32) (string, string, []string, string, string, time.Time, time.Time, error) {
	if r.NifiCluster.Spec.ListenersConfig == nil || r.NifiCluster.Spec.ListenersConfig.SSLSecrets == nil {
		return "", "", []string{}, "", "", time.Time{}, time.Time{}, nil
	}

	serverName := types.NamespacedName{
		Name:      fmt.Sprintf(pkicommon.NodeServerCertTemplate, r.NifiCluster.Name, nodeId),
		Namespace: r.NifiCluster.Namespace,
	}
	serverSecret := &corev1.Secret{}
	if err := r.DirectClient.Get(context.TODO(), serverName, serverSecret); err != nil {
		if apierrors.IsNotFound(err) {
			return "", "", []string{}, "", "", time.Time{}, time.Time{},
				errorfactory.New(errorfactory.ResourceNotReady{}, err, "server secret not ready")
		}
		return "", "", []string{}, "", "", time.Time{}, time.Time{},
			errors.WrapIfWithDetails(err, "failed to get server secret")
	}

	if err := requireTLSSecretReady(serverSecret, "server"); err != nil {
		return "", "", []string{}, "", "", time.Time{}, time.Time{}, err
	}

	serverPass := string(serverSecret.Data[v1.PasswordKey])
	serverHash := secretCertMaterialHash(serverSecret)

	serverCert, err := certutil.DecodeCertificate(serverSecret.Data[corev1.TLSCertKey])
	if err != nil {
		return "", "", []string{}, "", "", time.Time{}, time.Time{},
			errorfactory.New(errorfactory.ResourceNotReady{}, err, "server certificate not ready/decodable yet")
	}
	serverNotAfter := serverCert.NotAfter

	clientName := types.NamespacedName{
		Name:      r.NifiCluster.GetNifiControllerUserIdentity(),
		Namespace: r.NifiCluster.Namespace,
	}
	clientSecret := &corev1.Secret{}
	if err := r.DirectClient.Get(context.TODO(), clientName, clientSecret); err != nil {
		if apierrors.IsNotFound(err) {
			return "", "", []string{}, "", "", time.Time{}, time.Time{},
				errorfactory.New(errorfactory.ResourceNotReady{}, err, "client secret not ready")
		}
		return "", "", []string{}, "", "", time.Time{}, time.Time{},
			errors.WrapIfWithDetails(err, "failed to get client secret")
	}

	if err := requireTLSSecretReady(clientSecret, "client"); err != nil {
		return "", "", []string{}, "", "", time.Time{}, time.Time{}, err
	}

	clientPass := string(clientSecret.Data[v1.PasswordKey])
	clientHash := secretCertMaterialHash(clientSecret)

	clientCert, err := certutil.DecodeCertificate(clientSecret.Data[corev1.TLSCertKey])
	if err != nil {
		return "", "", []string{}, "", "", time.Time{}, time.Time{},
			errorfactory.New(errorfactory.ResourceNotReady{}, err, "client certificate not ready/decodable yet")
	}
	clientNotAfter := clientCert.NotAfter

	superUsers := []string{
		serverCert.Subject.String(),
		clientCert.Subject.String(),
	}

	return serverPass, clientPass, superUsers, serverHash, clientHash, serverNotAfter, clientNotAfter, nil
}
func generateNodeIdsFromPodSlice(pods []corev1.Pod) []string {
	ids := make([]string, len(pods))
	for i, node := range pods {
		ids[i] = node.Labels["nodeId"]
	}
	return ids
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

	if len(podList.Items) == 0 {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desiredPod); err != nil {
			return errors.WrapIf(err, "could not apply last state to annotation"), false
		}

		if err := r.Client.Create(context.TODO(), desiredPod); err != nil {
			return errorfactory.New(errorfactory.APIFailure{},
				err, "creating resource failed", "kind", desiredType), false
		}

		// Update status to Config InSync because node is configured to go
		statusErr := k8sutil.UpdateNodeStatus(r.Client, []string{desiredPod.Labels["nodeId"]}, r.NifiCluster, r.NifiClusterCurrentStatus, v1.ConfigInSync, log)
		if statusErr != nil {
			return errorfactory.New(errorfactory.StatusUpdateError{},
				statusErr, "updating status for resource failed", "kind", desiredType), false
		}

		// set node creation time
		statusErr = k8sutil.UpdateNodeStatus(r.Client, []string{desiredPod.Labels["nodeId"]}, r.NifiCluster, r.NifiClusterCurrentStatus, metav1.NewTime(time.Now().UTC()), log)
		if statusErr != nil {
			return errorfactory.New(errorfactory.StatusUpdateError{},
				statusErr, "failed to update node status creation time", "kind", desiredType), false
		}

		if val, ok := r.NifiCluster.Status.NodesState[desiredPod.Labels["nodeId"]]; ok &&
			val.GracefulActionState.State != v1.GracefulUpscaleSucceeded {
			gracefulActionState := v1.GracefulActionState{ErrorMessage: "", State: v1.GracefulUpscaleSucceeded}

			if !k8sutil.PodReady(currentPod) {
				gracefulActionState = v1.GracefulActionState{ErrorMessage: "", State: v1.GracefulUpscaleRequired}
			}

			statusErr = k8sutil.UpdateNodeStatus(r.Client, []string{desiredPod.Labels["nodeId"]}, r.NifiCluster, r.NifiClusterCurrentStatus, gracefulActionState, log)
			if statusErr != nil {
				return errorfactory.New(errorfactory.StatusUpdateError{},
					statusErr, "could not update node graceful action state"), false
			}
		}
		log.Info("Pod created",
			zap.String("clusterName", r.NifiCluster.Name),
			zap.String("nodeId", desiredPod.Labels["nodeId"]),
			zap.String("podName", desiredPod.Name))

		return nil, false
	} else if len(podList.Items) == 1 {
		currentPod = podList.Items[0].DeepCopy()
		nodeId := currentPod.Labels["nodeId"]
		if _, ok := r.NifiCluster.Status.NodesState[nodeId]; ok {
			if currentPod.Spec.NodeName == "" {
				log.Debug("pod for NodeId is not scheduled to node yet",
					zap.String("clusterName", r.NifiCluster.Name),
					zap.String("nodeId", nodeId))
			}
		} else {
			return errorfactory.New(errorfactory.InternalError{}, errors.New("reconcile failed"),
				fmt.Sprintf("could not find status for the given node id, %s", nodeId)), false
		}
	} else {
		return errorfactory.New(errorfactory.TooManyResources{}, errors.New("reconcile failed"),
			"more than one matching pod found", "labels", matchingLabels), false
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
				zap.String("podName", currentPod.Name),
				zap.Error(err),
			)

			return errorfactory.New(errorfactory.APIFailure{}, err, "could not calculate pod patch"), false
		}

		if patchResult.IsEmpty() {
			if !k8sutil.IsPodTerminatedOrShutdown(currentPod) &&
				r.NifiCluster.Status.NodesState[currentPod.Labels["nodeId"]].ConfigurationState == v1.ConfigInSync {

				if val, found := r.NifiCluster.Status.NodesState[desiredPod.Labels["nodeId"]]; found &&
					val.GracefulActionState.State == v1.GracefulUpscaleRunning &&
					val.GracefulActionState.ActionStep == v1.ConnectStatus &&
					k8sutil.PodReady(currentPod) {

					if err := k8sutil.UpdateNodeStatus(
						r.Client,
						[]string{desiredPod.Labels["nodeId"]},
						r.NifiCluster,
						r.NifiClusterCurrentStatus,
						v1.GracefulActionState{ErrorMessage: "", State: v1.GracefulUpscaleSucceeded},
						log,
					); err != nil {
						return errorfactory.New(errorfactory.StatusUpdateError{},
							err, "could not update node graceful action state"), false
					}
				}

				log.Debug("pod resource is in sync",
					zap.String("clusterName", r.NifiCluster.Name),
					zap.String("podName", currentPod.Name))

				return nil, k8sutil.PodReady(currentPod)
			}
		} else {
			var pr map[string]interface{}
			if err := json.Unmarshal(patchResult.Patch, &pr); err == nil {
				if len(pr) == 1 {
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

								if val, found := r.NifiCluster.Status.NodesState[desiredPod.Labels["nodeId"]]; found &&
									val.GracefulActionState.State == v1.GracefulUpscaleRunning &&
									val.GracefulActionState.ActionStep == v1.ConnectStatus &&
									k8sutil.PodReady(currentPod) {

									if err := k8sutil.UpdateNodeStatus(
										r.Client,
										[]string{desiredPod.Labels["nodeId"]},
										r.NifiCluster,
										r.NifiClusterCurrentStatus,
										v1.GracefulActionState{ErrorMessage: "", State: v1.GracefulUpscaleSucceeded},
										log,
									); err != nil {
										return errorfactory.New(errorfactory.StatusUpdateError{},
											err, "could not update node graceful action state"), false
									}
								}

								log.Debug("pod resource is in sync (order-only diff ignored)",
									zap.String("clusterName", r.NifiCluster.Name),
									zap.String("podName", currentPod.Name))

								return nil, k8sutil.PodReady(currentPod)
							}
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

		if delay, reason, derr := r.shouldDelayCertRotationRestart(log, currentPod, desiredPod, patchResult.Patch); derr != nil {
			log.Warn("cert rotation gate errored; falling back to immediate restart", zap.Error(derr))
		} else if delay {
			log.Info("Delaying pod restart for cert rotation", zap.String("podName", currentPod.Name), zap.String("reason", reason))

			return errorfactory.New(
				errorfactory.ReconcileRollingUpgrade{},
				errors.New("cert rotation pending (outside maintenance window)"),
				"waiting for cert rotation maintenance window",
				"pod", currentPod.Name,
				"nodeId", currentPod.Labels["nodeId"],
				"details", reason,
			), k8sutil.PodReady(currentPod)
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

func isReconcileRollingUpgradeErr(err error) bool {
	if err == nil {
		return false
	}
	_, ok := errors.Cause(err).(errorfactory.ReconcileRollingUpgrade)
	return ok
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
