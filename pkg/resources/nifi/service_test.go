package nifi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources"
)

func TestGenerateServicePortForExternalListeners(t *testing.T) {
	r := resources.Reconciler{
		NifiCluster: &v1.NifiCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster",
				Namespace: "namespace",
			},
			Spec: v1.NifiClusterSpec{
				ListenersConfig: &v1.ListenersConfig{
					InternalListeners: []v1.InternalListenerConfig{
						{
							Name:     "foo-listener",
							Protocol: corev1.ProtocolTCP,
						},
					},
				},
			},
		},
	}
	rec := Reconciler{
		Reconciler: r,
	}

	lbClass := "foo-lb-class"
	esconfig := v1.ExternalServiceConfig{
		Name: "foo",
		Spec: v1.ExternalServiceSpec{
			LoadBalancerClass: &lbClass,
			PortConfigs: []v1.PortConfig{
				{
					Port:                 5,
					InternalListenerName: "foo-listener",
					NodePort:             new(int32),
					Protocol:             corev1.ProtocolTCP,
				},
			},
		},
	}

	servicePorts := rec.generateServicePortForExternalListeners(esconfig)

	assert.NotEmpty(t, servicePorts, "servicePorts should not be empty")
	assert.Equal(t, 1, len(servicePorts), "there should only be 1 service port")
	assert.Equal(t, esconfig.Spec.PortConfigs[0].Protocol, servicePorts[0].Protocol, "service port protocol should be TCP")
	assert.Equal(t, esconfig.Spec.PortConfigs[0].InternalListenerName, servicePorts[0].Name, "service port name should be same as listener")
	assert.Equal(t, esconfig.Spec.PortConfigs[0].Port, servicePorts[0].Port, "service port has incorrect port")
}

func TestExternalServices(t *testing.T) {
	lbClass := "foo-lb-class"
	esconfig := v1.ExternalServiceConfig{
		Name: "foo",
		Spec: v1.ExternalServiceSpec{
			LoadBalancerClass: &lbClass,
			PortConfigs: []v1.PortConfig{
				{
					Port:                 5,
					InternalListenerName: "foo-listener",
					NodePort:             new(int32),
					Protocol:             corev1.ProtocolTCP,
				},
			},
		},
	}
	r := resources.Reconciler{
		NifiCluster: &v1.NifiCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster",
				Namespace: "namespace",
			},
			Spec: v1.NifiClusterSpec{
				ListenersConfig: &v1.ListenersConfig{
					InternalListeners: []v1.InternalListenerConfig{
						{
							Name:     "foo-listener",
							Protocol: corev1.ProtocolTCP,
						},
					},
				},
				ExternalServices: []v1.ExternalServiceConfig{
					esconfig,
				},
			},
		},
	}

	rec := Reconciler{
		Reconciler: r,
	}

	externalServices := rec.externalServices(*zap.NewNop())

	assert.NotEmpty(t, externalServices, "external services should not be empty")
	assert.Equal(t, 1, len(externalServices), "there should only be 1 external service")
	assert.Equal(t, esconfig.Spec.LoadBalancerClass, externalServices[0].(*corev1.Service).Spec.LoadBalancerClass, "service load balancer class should be same as external service")
	assert.Equal(t, esconfig.Spec.PortConfigs[0].InternalListenerName, externalServices[0].(*corev1.Service).Spec.Ports[0].Name, "service port name should be same as external service")
	assert.Equal(t, esconfig.Spec.PortConfigs[0].Protocol, externalServices[0].(*corev1.Service).Spec.Ports[0].Protocol, "service protocol should be same as external service")
	assert.Equal(t, esconfig.Spec.PortConfigs[0].Port, externalServices[0].(*corev1.Service).Spec.Ports[0].Port, "service port name should be same as external service")
}
