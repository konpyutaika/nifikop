package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
)

func TestGetParameterContext(t *testing.T) {
	assert := assert.New(t)

	id := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	entity, err := testGetParameterContext(t, id, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetParameterContext(t, id, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetParameterContext(t, id, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetParameterContext(t *testing.T, id string, status int) (*nigoapi.ParameterContextEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/parameter-contexts/%s", id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockParameterContext(id, "test-unit", "unit test",
					map[string]string{"key1": "value1", "key2": "value2"},
					map[string]string{"secret1": "value1", "secret2": "value2"}))
		})

	return client.GetParameterContext(id)
}

func TestCreateParameterContext(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockParameterContext("16cfd2ec-0174-1000-0000-00004b9b35cc", "test-unit", "unit test",
		map[string]string{"key1": "value1", "key2": "value2"},
		map[string]string{"secret1": "value1", "secret2": "value2"})

	entity, err := testCreateParameterContext(t, &mockEntity, 201)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testCreateParameterContext(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testCreateParameterContext(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testCreateParameterContext(t *testing.T, entity *nigoapi.ParameterContextEntity, status int) (*nigoapi.ParameterContextEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, "/parameter-contexts")
	httpmock.RegisterResponder(http.MethodPost, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.CreateParameterContext(*entity)
}

func TestRemoveParameterContext(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockParameterContext("16cfd2ec-0174-1000-0000-00004b9b35cc", "test-unit", "unit test",
		map[string]string{"key1": "value1", "key2": "value2"},
		map[string]string{"secret1": "value1", "secret2": "value2"})

	err := testRemoveParameterContext(t, &mockEntity, 200)
	assert.Nil(err)

	err = testRemoveParameterContext(t, &mockEntity, 404)
	assert.Nil(err)

	err = testRemoveParameterContext(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
}

func testRemoveParameterContext(t *testing.T, entity *nigoapi.ParameterContextEntity, status int) error {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/parameter-contexts/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodDelete, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.RemoveParameterContext(*entity)
}

func TestCreateParameterContextUpdateRequest(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockParameterContext("16cfd2ec-0174-1000-0000-00004b9b35cc", "test-unit",
		"unit test",
		map[string]string{"key1": "value1", "key2": "value2"},
		map[string]string{"secret1": "value1", "secret2": "value2"})

	entity, err := testCreateParameterContextUpdateRequest(t, &mockEntity, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testCreateParameterContextUpdateRequest(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testCreateParameterContextUpdateRequest(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testCreateParameterContextUpdateRequest(
	t *testing.T,
	entity *nigoapi.ParameterContextEntity,
	status int) (*nigoapi.ParameterContextUpdateRequestEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf(
		"/parameter-contexts/%s/update-requests", entity.Id))
	httpmock.RegisterResponder(http.MethodPost, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.CreateParameterContextUpdateRequest(entity.Id, *entity)
}

func TestGetParameterContextUpdateRequest(t *testing.T) {
	assert := assert.New(t)

	id := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	mockEntity := MockParameterContext("16cfd2ec-0174-1000-0000-00004b9b35cc", "test-unit",
		"unit test",
		map[string]string{"key1": "value1", "key2": "value2"},
		map[string]string{"secret1": "value1", "secret2": "value2"})

	entity, err := testGetParameterContextUpdateRequest(t, &mockEntity, id, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetParameterContextUpdateRequest(t, &mockEntity, id, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetParameterContextUpdateRequest(t, &mockEntity, id, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetParameterContextUpdateRequest(
	t *testing.T,
	entity *nigoapi.ParameterContextEntity,
	id string, status int) (*nigoapi.ParameterContextUpdateRequestEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf(
		"/parameter-contexts/%s/update-requests/%s", entity.Component.Id, id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.GetParameterContextUpdateRequest(entity.Component.Id, id)
}

func MockParameterContext(
	id, name, description string,
	params, sensitivesParameters map[string]string) nigoapi.ParameterContextEntity {
	var version int64 = 10
	parameters := map2Parameters(params, false)
	parameters = append(parameters, map2Parameters(sensitivesParameters, true)...)
	return nigoapi.ParameterContextEntity{
		Id: id,
		Component: &nigoapi.ParameterContextDto{
			Name:        name,
			Description: description,
			Id:          id,
			Parameters:  parameters,
		},
		Revision: &nigoapi.RevisionDto{Version: &version},
	}
}

func map2Parameters(params map[string]string, sensitive bool) []nigoapi.ParameterEntity {
	var parameters []nigoapi.ParameterEntity
	emptyString := ""
	for k, v := range params {
		parameters = append(parameters, nigoapi.ParameterEntity{
			Parameter: &nigoapi.ParameterDto{
				Name:        k,
				Description: &emptyString,
				Sensitive:   sensitive,
				Value:       &v,
			},
		})
	}

	return parameters
}
