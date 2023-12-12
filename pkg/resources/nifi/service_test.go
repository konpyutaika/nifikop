package nifi

import (
	"testing"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	esconfig := v1.ExternalServiceConfig{
		Name: "foo",
		Spec: v1.ExternalServiceSpec{
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
