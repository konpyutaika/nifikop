package nifi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func basePodForOverrides() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "cluster-0-",
			Namespace:    "namespace",
			Labels:       map[string]string{"app": "nifi", "nodeId": "0", "team": "platform"},
			Annotations:  map[string]string{"existing": "value"},
		},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser:    ptrTo[int64](1000),
				RunAsNonRoot: ptrTo(true),
				FSGroup:      ptrTo[int64](1000),
			},
			RestartPolicy:      corev1.RestartPolicyNever,
			Hostname:           "cluster-0-node",
			Subdomain:          "cluster-all-node",
			PriorityClassName:  "",
			ServiceAccountName: "default",
			InitContainers: []corev1.Container{
				{Name: "zookeeper", Image: "busybox"},
				{Name: "fetch-certs", Image: "busybox"},
			},
			Containers: []corev1.Container{
				{
					Name:    ContainerName,
					Image:   "apache/nifi:2.0.0",
					Command: []string{"bash", "-ce", "run.sh"},
					Ports:   []corev1.ContainerPort{{Name: "http", ContainerPort: 8080}},
					Env:     []corev1.EnvVar{{Name: "NIFI_ZOOKEEPER_CONNECT_STRING", Value: "zk:2181"}},
					VolumeMounts: []corev1.VolumeMount{
						{Name: "conf", MountPath: "/opt/nifi/nifi-current/conf"},
					},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2Gi")},
					},
				},
			},
			Volumes: []corev1.Volume{
				{Name: "conf", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
		},
	}
}

func TestApplyPodOverridesNilIsNoop(t *testing.T) {
	base := basePodForOverrides()
	expected := base.DeepCopy()

	got, err := applyPodOverrides(base, nil)
	require.NoError(t, err)
	assert.Same(t, base, got, "nil overrides must return the very same pod")
	assert.Equal(t, expected, got)
}

func TestApplyPodOverridesScalarSpecFields(t *testing.T) {
	got, err := applyPodOverrides(basePodForOverrides(), &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			HostUsers:          ptrTo(false),
			PriorityClassName:  "high",
			ServiceAccountName: "nifi-sa",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, got.Spec.HostUsers)
	assert.False(t, *got.Spec.HostUsers)
	assert.Equal(t, "high", got.Spec.PriorityClassName)
	assert.Equal(t, "nifi-sa", got.Spec.ServiceAccountName)
	// untouched fields survive
	assert.Len(t, got.Spec.Containers, 1)
	assert.Equal(t, []string{"bash", "-ce", "run.sh"}, got.Spec.Containers[0].Command)
}

func TestApplyPodOverridesPodSecurityContextWins(t *testing.T) {
	got, err := applyPodOverrides(basePodForOverrides(), &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser: ptrTo[int64](1000650000),
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1000650000), *got.Spec.SecurityContext.RunAsUser)
	// sibling fields of the same struct are kept
	assert.Equal(t, int64(1000), *got.Spec.SecurityContext.FSGroup)
	assert.True(t, *got.Spec.SecurityContext.RunAsNonRoot)
}

func TestApplyPodOverridesManagedContainerMergesByName(t *testing.T) {
	got, err := applyPodOverrides(basePodForOverrides(), &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: ContainerName,
					SecurityContext: &corev1.SecurityContext{
						ReadOnlyRootFilesystem: ptrTo(true),
					},
					Env: []corev1.EnvVar{{Name: "EXTRA", Value: "1"}},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, got.Spec.Containers, 1, "patching the nifi container must not append a second one")
	c := got.Spec.Containers[0]
	assert.Equal(t, ContainerName, c.Name)
	assert.Equal(t, "apache/nifi:2.0.0", c.Image)
	assert.Equal(t, []string{"bash", "-ce", "run.sh"}, c.Command)
	assert.Len(t, c.Ports, 1)
	assert.Len(t, c.VolumeMounts, 1)
	require.NotNil(t, c.SecurityContext)
	assert.True(t, *c.SecurityContext.ReadOnlyRootFilesystem)
	// env merges by name: existing kept, new appended
	assert.ElementsMatch(t, []string{"NIFI_ZOOKEEPER_CONNECT_STRING", "EXTRA"},
		[]string{c.Env[0].Name, c.Env[1].Name})
	// resources map merges key-wise
	assert.Equal(t, "2Gi", c.Resources.Limits.Memory().String())
	assert.Equal(t, "2", c.Resources.Limits.Cpu().String())
}

func TestApplyPodOverridesUnknownContainerAppends(t *testing.T) {
	got, err := applyPodOverrides(basePodForOverrides(), &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "sidecar", Image: "example/sidecar:1"},
			},
			InitContainers: []corev1.Container{
				{Name: "prep", Image: "busybox"},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, got.Spec.Containers, 2)
	assert.Equal(t, ContainerName, got.Spec.Containers[0].Name)
	assert.Equal(t, "sidecar", got.Spec.Containers[1].Name)
	require.Len(t, got.Spec.InitContainers, 3)
	assert.Equal(t, []string{"zookeeper", "fetch-certs", "prep"},
		[]string{got.Spec.InitContainers[0].Name, got.Spec.InitContainers[1].Name, got.Spec.InitContainers[2].Name})
}

func TestApplyPodOverridesVolumesMergeByName(t *testing.T) {
	got, err := applyPodOverrides(basePodForOverrides(), &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				// existing volume: patched, not duplicated
				{Name: "conf", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				}}},
				// new volume: appended
				{Name: "scratch", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, got.Spec.Volumes, 2)
	assert.Equal(t, "conf", got.Spec.Volumes[0].Name)
	assert.Equal(t, corev1.StorageMediumMemory, got.Spec.Volumes[0].EmptyDir.Medium)
	assert.Equal(t, "scratch", got.Spec.Volumes[1].Name)
}

func TestApplyPodOverridesMetadataLabelsAnnotations(t *testing.T) {
	got, err := applyPodOverrides(basePodForOverrides(), &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "must-be-ignored",
			Namespace: "must-be-ignored",
			Labels:    map[string]string{"team": "overridden", "extra": "label"},
			Annotations: map[string]string{
				"existing": "overridden",
				"new":      "annotation",
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "cluster-0-", got.GenerateName)
	assert.Equal(t, "", got.Name)
	assert.Equal(t, "namespace", got.Namespace)
	assert.Equal(t, map[string]string{"app": "nifi", "nodeId": "0", "team": "overridden", "extra": "label"}, got.Labels)
	assert.Equal(t, map[string]string{"existing": "overridden", "new": "annotation"}, got.Annotations)
}

func TestApplyPodOverridesOperatorOwnedLabelsAreKept(t *testing.T) {
	base := basePodForOverrides()
	base.Labels["nifi_cr"] = "cluster"
	got, err := applyPodOverrides(base, &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "hijack", "nifi_cr": "hijack", "nodeId": "7", "team": "data"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"app": "nifi", "nifi_cr": "cluster", "nodeId": "0", "team": "data"}, got.Labels)
}

func TestApplyPodOverridesEnvKeepsOperatorOrder(t *testing.T) {
	base := basePodForOverrides()
	base.Spec.Containers[0].Env = []corev1.EnvVar{{Name: "A", Value: "a"}, {Name: "B", Value: "$(A)"}}
	got, err := applyPodOverrides(base, &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{Containers: []corev1.Container{{
			Name: ContainerName,
			Env:  []corev1.EnvVar{{Name: "C", Value: "$(B)"}, {Name: "A", Value: "override"}},
		}}},
	})
	require.NoError(t, err)
	env := got.Spec.Containers[0].Env
	require.Len(t, env, 3)
	assert.Equal(t, []string{"A", "B", "C"}, []string{env[0].Name, env[1].Name, env[2].Name})
	assert.Equal(t, "override", env[0].Value)
}

func TestApplyPodOverridesRejectsDuplicateNames(t *testing.T) {
	_, err := applyPodOverrides(basePodForOverrides(), &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "x", Image: "1"}, {Name: "x", Image: "2"}}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already used")
}

func TestApplyPodOverridesOperatorOwnedFieldsAreKept(t *testing.T) {
	got, err := applyPodOverrides(basePodForOverrides(), &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Hostname:      "evil",
			Subdomain:     "evil",
			RestartPolicy: corev1.RestartPolicyAlways,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "cluster-0-node", got.Spec.Hostname)
	assert.Equal(t, "cluster-all-node", got.Spec.Subdomain)
	assert.Equal(t, corev1.RestartPolicyNever, got.Spec.RestartPolicy)
}

func TestApplyPodOverridesRejectsUnnamedContainers(t *testing.T) {
	for _, tc := range []struct {
		name      string
		overrides *corev1.PodTemplateSpec
	}{
		{"container", &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Image: "x"}}}}},
		{"initContainer", &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{{Image: "x"}}}}},
		{"volume", &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := applyPodOverrides(basePodForOverrides(), tc.overrides)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "name")
		})
	}
}

func TestApplyPodOverridesDoesNotMutateInput(t *testing.T) {
	base := basePodForOverrides()
	snapshot := base.DeepCopy()
	overrides := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{PriorityClassName: "high"}}
	overridesSnapshot := overrides.DeepCopy()

	_, err := applyPodOverrides(base, overrides)
	require.NoError(t, err)
	assert.Equal(t, snapshot, base)
	assert.Equal(t, overridesSnapshot, overrides)
}
