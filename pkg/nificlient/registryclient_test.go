package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
)

func TestGetRegistryClient(t *testing.T) {
	assert := assert.New(t)

	id := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	entity, err := testGetRegistryClient(t, id, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetRegistryClient(t, id, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetRegistryClient(t, id, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetRegistryClient(t *testing.T, id string, status int) (*nigoapi.FlowRegistryClientEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller/registry-clients/%s", id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockRegistryClient(id, "registry-mock", "description", "http://uri.com:8888"))
		})

	return client.GetRegistryClient(id)
}

func TestCreateRegistryClient(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockRegistryClient("16cfd2ec-0174-1000-0000-00004b9b35cc", "mock", "description", "http://uri:8888")

	entity, err := testCreateRegistryClient(t, &mockEntity, 201)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testCreateRegistryClient(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testCreateRegistryClient(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testCreateRegistryClient(t *testing.T, entity *nigoapi.FlowRegistryClientEntity, status int) (*nigoapi.FlowRegistryClientEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, "/controller/registry-clients")
	httpmock.RegisterResponder(http.MethodPost, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.CreateRegistryClient(*entity)
}

func TestUpdateRegistryClient(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockRegistryClient("16cfd2ec-0174-1000-0000-00004b9b35cc", "mock", "description", "http://uri:8888")

	entity, err := testUpdateRegistryClient(t, &mockEntity, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testUpdateRegistryClient(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testUpdateRegistryClient(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testUpdateRegistryClient(t *testing.T, entity *nigoapi.FlowRegistryClientEntity, status int) (*nigoapi.FlowRegistryClientEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller/registry-clients/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.UpdateRegistryClient(*entity)
}

func TestRemoveRegistryClient(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockRegistryClient("16cfd2ec-0174-1000-0000-00004b9b35cc", "mock", "description", "http://uri:8888")

	err := testRemoveRegistryClient(t, &mockEntity, 200)
	assert.Nil(err)

	err = testRemoveRegistryClient(t, &mockEntity, 404)
	assert.Nil(err)

	err = testRemoveRegistryClient(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
}

func testRemoveRegistryClient(t *testing.T, entity *nigoapi.FlowRegistryClientEntity, status int) error {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller/registry-clients/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodDelete, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.RemoveRegistryClient(*entity)
}

func MockRegistryClient(id, name, description, uri string) nigoapi.FlowRegistryClientEntity {
	var version int64 = 10
	return nigoapi.FlowRegistryClientEntity{
		Id: id,
		Component: &nigoapi.FlowRegistryClientDto{
			Id:          id,
			Name:        name,
			Description: description,
			Uri:         uri,
		},
		Revision: &nigoapi.RevisionDto{Version: &version},
	}
}
