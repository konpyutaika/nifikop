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

package k8sutil

import (
	"context"
	"reflect"

	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/errorfactory"
	v1 "k8s.io/api/core/v1"

	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconcile reconciles K8S resources
func Reconcile(log logr.Logger, client runtimeClient.Client, desired runtimeClient.Object, cr *v1alpha1.NifiCluster) error {
	desiredType := reflect.TypeOf(desired)
	current := desired.DeepCopyObject().(runtimeClient.Object)

	var err error
	switch desired.(type) {
	default:
		var key runtimeClient.ObjectKey
		key = runtimeClient.ObjectKeyFromObject(current)
		log = log.WithValues("kind", desiredType, "name", key.Name)

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
			log.Info("resource created")
			return nil
		}
	}
	if err == nil {
		switch desired.(type) {
		case *v1alpha1.NifiUser:
			user := desired.(*v1alpha1.NifiUser)
			user.Status = current.(*v1alpha1.NifiUser).Status
			desired = user
		case *v1alpha1.NifiUserGroup:
			group := desired.(*v1alpha1.NifiUserGroup)
			group.Status = current.(*v1alpha1.NifiUserGroup).Status
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

			if err := client.Update(context.TODO(), desired); err != nil {
				return errorfactory.New(errorfactory.APIFailure{}, err, "updating resource failed", "kind", desiredType)
			}
			if cr != nil {
				switch desired.(type) {
				case *corev1.ConfigMap:
					// Only update status when configmap belongs to node
					if id, ok := desired.(*corev1.ConfigMap).Labels["nodeId"]; ok {
						statusErr := UpdateNodeStatus(client, []string{id}, cr, v1alpha1.ConfigOutOfSync, log)
						if statusErr != nil {
							return errors.WrapIfWithDetails(err, "updating status for resource failed", "kind", desiredType)
						}
					}
				case *corev1.Secret:
					// Only update status when secret belongs to node
					if id, ok := desired.(*corev1.Secret).Labels["nodeId"]; ok {
						statusErr := UpdateNodeStatus(client, []string{id}, cr, v1alpha1.ConfigOutOfSync, log)
						if statusErr != nil {
							return errors.WrapIfWithDetails(err, "updating status for resource failed", "kind", desiredType)
						}
					}
				}
			}
		}

		log.Info("resource updated")
	}
	return nil
}

// CheckIfObjectUpdated checks if the given object is updated using K8sObjectMatcher
func CheckIfObjectUpdated(log logr.Logger, desiredType reflect.Type, current, desired runtime.Object) bool {
	patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired)
	if err != nil {
		log.Error(err, "could not match objects", "kind", desiredType)
		return true
	} else if patchResult.IsEmpty() {
		log.V(1).Info("resource is in sync")
		return false
	} else {
		log.Info("resource diffs",
			"patch", string(patchResult.Patch),
			"current", string(patchResult.Current),
			"modified", string(patchResult.Modified),
			"original", string(patchResult.Original))
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
	if &pod.Status != nil && len(pod.Status.Conditions) > 0 {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodReady &&
				condition.Status == v1.ConditionTrue {
				return true
			}
		}
	}
	return false
}
