package nifi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/resources"
)

func TestGetNifiPropertiesConfigStringIncludesTLSAutoReloadWhenEnabled(t *testing.T) {
	r := resources.Reconciler{
		NifiCluster: &v1.NifiCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster",
				Namespace: "namespace",
			},
			Spec: v1.NifiClusterSpec{
				Service: v1.ServicePolicy{},
				ListenersConfig: &v1.ListenersConfig{
					InternalListeners: []v1.InternalListenerConfig{
						{
							Type:          v1.HttpsListenerType,
							ContainerPort: 8443,
						},
					},
					SSLSecrets: &v1.SSLSecrets{},
				},
				ReadOnlyConfig: v1.ReadOnlyConfig{
					NifiProperties: v1.NifiProperties{
						TLSAutoReload: &v1.TLSAutoReloadConfig{
							Enabled:  true,
							Interval: "15 secs",
						},
					},
				},
			},
		},
	}

	rec := Reconciler{Reconciler: r}
	props := rec.getNifiPropertiesConfigString(&v1.NodeConfig{}, 0, "server-pass", "client-pass", nil, *zap.NewNop())

	assert.Contains(t, props, "nifi.security.autoreload.enabled=true")
	assert.Contains(t, props, "nifi.security.autoreload.interval=15 secs")
}

func TestGetNifiPropertiesConfigStringUsesDefaultTLSAutoReloadInterval(t *testing.T) {
	r := resources.Reconciler{
		NifiCluster: &v1.NifiCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster",
				Namespace: "namespace",
			},
			Spec: v1.NifiClusterSpec{
				Service: v1.ServicePolicy{},
				ListenersConfig: &v1.ListenersConfig{
					InternalListeners: []v1.InternalListenerConfig{
						{
							Type:          v1.HttpsListenerType,
							ContainerPort: 8443,
						},
					},
					SSLSecrets: &v1.SSLSecrets{},
				},
				ReadOnlyConfig: v1.ReadOnlyConfig{
					NifiProperties: v1.NifiProperties{
						TLSAutoReload: &v1.TLSAutoReloadConfig{
							Enabled: true,
						},
					},
				},
			},
		},
	}

	rec := Reconciler{Reconciler: r}
	props := rec.getNifiPropertiesConfigString(&v1.NodeConfig{}, 0, "server-pass", "client-pass", nil, *zap.NewNop())

	assert.Contains(t, props, "nifi.security.autoreload.enabled=true")
	assert.Contains(t, props, "nifi.security.autoreload.interval=10 secs")
}

func TestGetNifiPropertiesConfigStringSkipsTLSAutoReloadWhenDisabled(t *testing.T) {
	r := resources.Reconciler{
		NifiCluster: &v1.NifiCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster",
				Namespace: "namespace",
			},
			Spec: v1.NifiClusterSpec{
				Service: v1.ServicePolicy{},
				ListenersConfig: &v1.ListenersConfig{
					InternalListeners: []v1.InternalListenerConfig{
						{
							Type:          v1.HttpsListenerType,
							ContainerPort: 8443,
						},
					},
					SSLSecrets: &v1.SSLSecrets{},
				},
			},
		},
	}

	rec := Reconciler{Reconciler: r}
	props := rec.getNifiPropertiesConfigString(&v1.NodeConfig{}, 0, "server-pass", "client-pass", nil, *zap.NewNop())

	assert.NotContains(t, props, "nifi.security.autoreload.enabled=true")
	assert.NotContains(t, props, "nifi.security.autoreload.interval=")
}

func TestApplyTLSAutoReloadAnnotations(t *testing.T) {
	t.Run("adds annotations when enabled", func(t *testing.T) {
		pod := &corev1.Pod{}

		applyTLSAutoReloadAnnotations(pod, &v1.NifiProperties{
			TLSAutoReload: &v1.TLSAutoReloadConfig{
				Enabled:  true,
				Interval: "15 secs",
			},
		})

		assert.Equal(t, "true", pod.Annotations[podTLSAutoReloadEnabledAnnotation])
		assert.Equal(t, "15 secs", pod.Annotations[podTLSAutoReloadIntervalAnnotation])
	})

	t.Run("removes annotations when disabled", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					podTLSAutoReloadEnabledAnnotation:  "true",
					podTLSAutoReloadIntervalAnnotation: "15 secs",
					"other":                            "keep",
				},
			},
		}

		applyTLSAutoReloadAnnotations(pod, &v1.NifiProperties{})

		assert.NotContains(t, pod.Annotations, podTLSAutoReloadEnabledAnnotation)
		assert.NotContains(t, pod.Annotations, podTLSAutoReloadIntervalAnnotation)
		assert.Equal(t, "keep", pod.Annotations["other"])
	})
}

func TestValidateTLSAutoReloadProperties(t *testing.T) {
	t.Run("accepts valid interval", func(t *testing.T) {
		err := validateTLSAutoReloadProperties(&v1.NifiProperties{
			TLSAutoReload: &v1.TLSAutoReloadConfig{
				Enabled:  true,
				Interval: "30 mins",
			},
		})

		require.NoError(t, err)
	})

	t.Run("rejects invalid interval", func(t *testing.T) {
		err := validateTLSAutoReloadProperties(&v1.NifiProperties{
			TLSAutoReload: &v1.TLSAutoReloadConfig{
				Enabled:  true,
				Interval: "whenever",
			},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "tlsAutoReload.interval")
	})
}
