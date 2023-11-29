package templates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/util"
)

// ObjectMeta returns a metav1.ObjectMeta object with labels, ownerReference and name.
func ObjectMeta(name string, labels map[string]string, cluster *v1.NifiCluster) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: cluster.Namespace,
		Labels:    ObjectMetaLabels(cluster, labels),
		OwnerReferences: []metav1.OwnerReference{
			ClusterOwnerReference(cluster),
		},
	}
}

// ObjectMetaWithGeneratedName returns a metav1.ObjectMeta object with labels, ownerReference and generatedname.
func ObjectMetaWithGeneratedName(namePrefix string, labels map[string]string, cluster *v1.NifiCluster) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		GenerateName: namePrefix,
		Namespace:    cluster.Namespace,
		Labels:       ObjectMetaLabels(cluster, labels),
		OwnerReferences: []metav1.OwnerReference{
			ClusterOwnerReference(cluster),
		},
	}
}

// ClusterOwnerReference returns the appropriate metadata to attach to an object to make the provided NifiCluster an owner of some object.
func ClusterOwnerReference(cluster *v1.NifiCluster) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         cluster.APIVersion,
		Kind:               cluster.Kind,
		Name:               cluster.Name,
		UID:                cluster.UID,
		Controller:         util.BoolPointer(true),
		BlockOwnerDeletion: util.BoolPointer(true),
	}
}

func ObjectMetaLabels(cluster *v1.NifiCluster, l map[string]string) map[string]string {
	if cluster.Spec.PropagateLabels {
		return util.MergeLabels(cluster.Labels, l)
	}
	return l
}

// ObjectMetaWithAnnotations returns a metav1.ObjectMeta object with labels, ownerReference, name and annotations.
func ObjectMetaWithAnnotations(name string, labels map[string]string, annotations map[string]string, cluster *v1.NifiCluster) metav1.ObjectMeta {
	o := ObjectMeta(name, labels, cluster)
	o.Annotations = annotations
	return o
}

// ObjectMetaWithGeneratedNameAndAnnotations returns a metav1.ObjectMeta object with labels, ownerReference, generatedname and annotations.
func ObjectMetaWithGeneratedNameAndAnnotations(namePrefix string, labels map[string]string, annotations map[string]string, cluster *v1.NifiCluster) metav1.ObjectMeta {
	o := ObjectMetaWithGeneratedName(namePrefix, labels, cluster)
	o.Annotations = annotations
	return o
}

// ObjectMetaClusterScope returns a metav1.ObjectMeta object with labels, ownerReference, name and annotations.
func ObjectMetaClusterScope(name string, labels map[string]string, cluster *v1.NifiCluster) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   name,
		Labels: ObjectMetaLabels(cluster, labels),
		OwnerReferences: []metav1.OwnerReference{
			ClusterOwnerReference(cluster),
		},
	}
}
