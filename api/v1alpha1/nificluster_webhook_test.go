package v1alpha1

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	//+kubebuilder:scaffold:imports
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NifiCluster webhook", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		ClusterName      = "test-cluster"
		ClusterNamespace = "default"
	)

	Context("When creating a NifiCluster resource", func() {
		It("Should validate that the CR spec satisfies the validation rules", func() {
			By("By creating a new NifiCluster")
			ctx := context.Background()
			template := "my-%d-template"
			// A nifi cluster with nodes only is acceptable
			nifiCluster := &NifiCluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       "NifiCluster",
					APIVersion: "nifi.konpyutaika.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ClusterName,
					Namespace: ClusterNamespace,
				},
				Spec: NifiClusterSpec{
					ZKPath:                   "/path",
					ZKAddress:                "localhost:2181",
					PropagateLabels:          true,
					NodeUserIdentityTemplate: &template,
					NifiControllerTemplate:   &template,
					ControllerUserIdentity:   &template,
					Nodes: []Node{
						{
							Id: 1,
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, nifiCluster)).Should(Succeed())

			// a nifi cluster with autoscaling only is acceptable
			nifiCluster = &NifiCluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       "NifiCluster",
					APIVersion: "nifi.konpyutaika.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "correct-config",
					Namespace: "default",
				},
				Spec: NifiClusterSpec{
					ZKPath:                   "/path",
					ZKAddress:                "localhost:2181",
					PropagateLabels:          true,
					NodeUserIdentityTemplate: &template,
					NifiControllerTemplate:   &template,
					ControllerUserIdentity:   &template,
					AutoScalingConfig: AutoScalingConfig{
						Enabled: true,
					},
				},
			}
			Expect(k8sClient.Create(ctx, nifiCluster)).Should(Succeed())

			// a nifi cluster with autoscaling and nodes configs is not acceptable
			nifiCluster = &NifiCluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       "NifiCluster",
					APIVersion: "nifi.konpyutaika.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "correct-config",
					Namespace: "default",
				},
				Spec: NifiClusterSpec{
					ZKPath:                   "/path",
					ZKAddress:                "localhost:2181",
					PropagateLabels:          true,
					NodeUserIdentityTemplate: &template,
					NifiControllerTemplate:   &template,
					ControllerUserIdentity:   &template,
					Nodes: []Node{
						{
							Id: 1,
						},
					},
					AutoScalingConfig: AutoScalingConfig{
						Enabled: true,
					},
				},
			}
			Expect(k8sClient.Create(ctx, nifiCluster)).ShouldNot(Succeed())
		})
	})
})
