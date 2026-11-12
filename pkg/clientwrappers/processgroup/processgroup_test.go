package processgroup

import (
	"testing"

	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/api/v1alpha1"
)

func testResource(parameterContextRef *v1.ParameterContextReference) *v1alpha1.NifiResource {
	resource := &v1alpha1.NifiResource{
		Spec: v1alpha1.NifiResourceSpec{
			Type:                v1alpha1.ResourceProcessGroup,
			DisplayName:         "tenant-a",
			Configuration:       runtime.RawExtension{Raw: []byte(`{"comments":"tenant"}`)},
			ParameterContextRef: parameterContextRef,
		},
	}
	resource.Name = "tenant-a"
	return resource
}

func testEntity(boundParameterContext *nigoapi.ParameterContextReferenceEntity) *nigoapi.ProcessGroupEntity {
	return &nigoapi.ProcessGroupEntity{
		Component: &nigoapi.ProcessGroupDto{
			Name:             "tenant-a",
			Comments:         "tenant",
			ParameterContext: boundParameterContext,
		},
	}
}

func TestUpdateProcessGroupEntityParameterContext(t *testing.T) {
	ref := &v1.ParameterContextReference{Name: "tenant-a"}

	t.Run("binds the parameter context id", func(t *testing.T) {
		entity := &nigoapi.ProcessGroupEntity{}
		require.NoError(t, updateProcessGroupEntity(testResource(ref), entity, "pc-1"))

		require.NotNil(t, entity.Component)
		require.NotNil(t, entity.Component.ParameterContext)
		assert.Equal(t, "pc-1", entity.Component.ParameterContext.Id)
		assert.Equal(t, "tenant-a", entity.Component.Name)
		assert.Equal(t, "tenant", entity.Component.Comments)
	})

	t.Run("replaces an existing binding", func(t *testing.T) {
		entity := testEntity(&nigoapi.ParameterContextReferenceEntity{Id: "pc-old"})
		require.NoError(t, updateProcessGroupEntity(testResource(ref), entity, "pc-new"))

		assert.Equal(t, "pc-new", entity.Component.ParameterContext.Id)
	})

	t.Run("unbinds with an empty reference when no id is given", func(t *testing.T) {
		entity := testEntity(&nigoapi.ParameterContextReferenceEntity{Id: "pc-1"})
		require.NoError(t, updateProcessGroupEntity(testResource(nil), entity, ""))

		require.NotNil(t, entity.Component.ParameterContext)
		assert.Equal(t, "", entity.Component.ParameterContext.Id)
	})
}

func TestProcessGroupIsSyncParameterContext(t *testing.T) {
	ref := &v1.ParameterContextReference{Name: "tenant-a"}

	tests := []struct {
		name               string
		resource           *v1alpha1.NifiResource
		bound              *nigoapi.ParameterContextReferenceEntity
		parameterContextId string
		want               bool
	}{
		{name: "same id", resource: testResource(ref), bound: &nigoapi.ParameterContextReferenceEntity{Id: "pc-1"}, parameterContextId: "pc-1", want: true},
		{name: "different id", resource: testResource(ref), bound: &nigoapi.ParameterContextReferenceEntity{Id: "pc-1"}, parameterContextId: "pc-2", want: false},
		{name: "unbound but expected", resource: testResource(ref), bound: nil, parameterContextId: "pc-1", want: false},
		{name: "empty binding but expected", resource: testResource(ref), bound: &nigoapi.ParameterContextReferenceEntity{}, parameterContextId: "pc-1", want: false},
		{name: "bound but ref removed", resource: testResource(nil), bound: &nigoapi.ParameterContextReferenceEntity{Id: "pc-1"}, parameterContextId: "", want: false},
		{name: "unbound and no ref", resource: testResource(nil), bound: nil, parameterContextId: "", want: true},
		{name: "empty binding and no ref", resource: testResource(nil), bound: &nigoapi.ParameterContextReferenceEntity{}, parameterContextId: "", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processGroupIsSync(tt.resource, testEntity(tt.bound), tt.parameterContextId)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEmptyConfiguration(t *testing.T) {
	resource := testResource(&v1.ParameterContextReference{Name: "tenant-a"})
	resource.Spec.Configuration = runtime.RawExtension{}

	entity := &nigoapi.ProcessGroupEntity{}
	require.NoError(t, updateProcessGroupEntity(resource, entity, "pc-1"))
	assert.Equal(t, "pc-1", entity.Component.ParameterContext.Id)

	isSync, err := processGroupIsSync(resource, entity, "pc-1")
	require.NoError(t, err)
	assert.True(t, isSync)
}

func TestBoundParameterContextId(t *testing.T) {
	assert.Equal(t, "", boundParameterContextId(nil))
	assert.Equal(t, "", boundParameterContextId(&nigoapi.ProcessGroupEntity{}))
	assert.Equal(t, "", boundParameterContextId(testEntity(nil)))
	assert.Equal(t, "pc-1", boundParameterContextId(testEntity(&nigoapi.ParameterContextReferenceEntity{Id: "pc-1"})))
}
