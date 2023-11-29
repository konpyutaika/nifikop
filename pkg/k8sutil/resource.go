package k8sutil

import (
	"context"
	"reflect"

	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	nifikopv1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
)

// Reconcile reconciles K8S resources.
func Reconcile(log zap.Logger, client runtimeClient.Client, desired runtimeClient.Object, cr *nifikopv1.NifiCluster, currentStatus *nifikopv1.NifiClusterStatus) error {
	desiredType := reflect.TypeOf(desired)
	current := desired.DeepCopyObject().(runtimeClient.Object)

	var err error
	switch desired.(type) {
	default:
		var key runtimeClient.ObjectKey
		key = runtimeClient.ObjectKeyFromObject(current)
		log.Debug("reconciling", zap.String("kind", desiredType.String()), zap.String("name", key.Name))

		err = client.Get(context.TODO(), key, current)
		if err != nil && !apierrors.IsNotFound(err) {
			return errorfactory.New(
				errorfactory.APIFailure{},
				err,
				"getting resource failed",
				"kind", desiredType, "name", key.Name,
			)
		}
		if apierrors.IsNotFound(err) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				return errors.WrapIf(err, "could not apply last state to annotation")
			}
			if err := client.Create(context.TODO(), desired); err != nil {
				return errorfactory.New(
					errorfactory.APIFailure{},
					err,
					"creating resource failed",
					"kind", desiredType, "name", key.Name,
				)
			}
			log.Info("resource created",
				zap.String("name", desired.GetName()),
				zap.String("namespace", desired.GetNamespace()),
				zap.String("kind", desired.GetObjectKind().GroupVersionKind().Kind))
			return nil
		}
	}
	if err == nil {
		switch desired.(type) {
		case *nifikopv1.NifiUser:
			user := desired.(*nifikopv1.NifiUser)
			user.Status = current.(*nifikopv1.NifiUser).Status
			desired = user
		case *nifikopv1.NifiUserGroup:
			group := desired.(*nifikopv1.NifiUserGroup)
			group.Status = current.(*nifikopv1.NifiUserGroup).Status
			desired = group
		}

		if CheckIfObjectUpdated(log, desiredType, current, desired) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				return errors.WrapIf(err, "could not apply last state to annotation")
			}

			switch d := desired.(type) {
			default:
				d.(metav1.ObjectMetaAccessor).GetObjectMeta().SetResourceVersion(current.(metav1.ObjectMetaAccessor).GetObjectMeta().GetResourceVersion())
			case *corev1.Service:
				svc := desired.(*corev1.Service)
				svc.ResourceVersion = current.(*corev1.Service).ResourceVersion
				svc.Spec.ClusterIP = current.(*corev1.Service).Spec.ClusterIP
				desired = svc
			}

			if cr != nil {
				// switch desired.(type) {
				switch desired := desired.(type) {
				case *corev1.ConfigMap:
					// Only update status when configmap belongs to node
					if id, ok := desired.Labels["nodeId"]; ok {
						statusErr := UpdateNodeStatus(client, []string{id}, cr, *currentStatus, nifikopv1.ConfigOutOfSync, log)
						if statusErr != nil {
							return errors.WrapIfWithDetails(err, "updating status for resource failed", "kind", desiredType)
						}
					}
				case *corev1.Secret:
					// Only update status when secret belongs to node
					if id, ok := desired.Labels["nodeId"]; ok {
						statusErr := UpdateNodeStatus(client, []string{id}, cr, *currentStatus, nifikopv1.ConfigOutOfSync, log)
						if statusErr != nil {
							return errors.WrapIfWithDetails(err, "updating status for resource failed", "kind", desiredType)
						}
					}
				}
			}

			if err := client.Update(context.TODO(), desired); err != nil {
				return errorfactory.New(errorfactory.APIFailure{}, err, "updating resource failed", "kind", desiredType)
			}
		}

		log.Debug("resource updated",
			zap.String("name", desired.GetName()),
			zap.String("namespace", desired.GetNamespace()),
			zap.String("kind", desired.GetObjectKind().GroupVersionKind().Kind))
	}
	return nil
}

// CheckIfObjectUpdated checks if the given object is updated using K8sObjectMatcher.
func CheckIfObjectUpdated(log zap.Logger, desiredType reflect.Type, current, desired runtime.Object) bool {
	patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired)
	if err != nil {
		log.Error("could not match objects", zap.Error(err), zap.String("kind", desiredType.String()))
		return true
	} else if patchResult.IsEmpty() {
		log.Debug("resource is in sync", zap.String("kind", desiredType.String()))
		return false
	} else {
		log.Debug("resource diffs",
			zap.String("patch", string(patchResult.Patch)),
			zap.String("current", string(patchResult.Current)),
			zap.String("modified", string(patchResult.Modified)),
			zap.String("original", string(patchResult.Original)))
		return true
	}
}

func IsPodTerminatedOrShutdown(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodFailed || IsPodContainsTerminatedContainer(pod)
}

func IsPodContainsTerminatedContainer(pod *corev1.Pod) bool {
	for _, containerState := range pod.Status.ContainerStatuses {
		if containerState.State.Terminated != nil {
			return true
		}
	}
	return false
}

func IsPodContainsPendingContainer(pod *corev1.Pod) bool {
	for _, containerState := range pod.Status.ContainerStatuses {
		if containerState.State.Waiting != nil {
			return true
		}
	}
	return false
}

func PodReady(pod *corev1.Pod) bool {
	for i := range pod.Status.Conditions {
		if pod.Status.Conditions[i].Type == corev1.PodReady && pod.Status.Conditions[i].Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
