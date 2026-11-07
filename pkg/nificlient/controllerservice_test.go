package nificlient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
)

const mockControllerServiceId = "16cfd2ec-0174-1000-0000-00004b9b35cc"

func TestGetControllerService(t *testing.T) {
	assert := assert.New(t)

	entity, err := testGetControllerService(t, mockControllerServiceId, 200)
	assert.Nil(err)
	if assert.NotNil(entity) {
		assert.Equal(mockControllerServiceId, entity.Id)
		assert.Equal("controllerservice-mock", entity.Component.Name)
		assert.Equal("DISABLED", entity.Component.State)
	}

	entity, err = testGetControllerService(t, mockControllerServiceId, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetControllerService(t, mockControllerServiceId, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetControllerService(t *testing.T, id string, status int) (*nigoapi.ControllerServiceEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller-services/%s", id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockControllerService(id, "", "controllerservice-mock", "DISABLED"))
		})

	return client.GetControllerService(id)
}

func TestCreateControllerService(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockControllerService(mockControllerServiceId, "", "controllerservice-mock", "DISABLED")

	entity, err := testCreateControllerService(t, &mockEntity, 201)
	assert.Nil(err)
	if assert.NotNil(entity) {
		assert.Equal(mockControllerServiceId, entity.Id)
	}

	entity, err = testCreateControllerService(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testCreateControllerService(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned201, err)
	assert.Nil(entity)
}

func testCreateControllerService(t *testing.T, entity *nigoapi.ControllerServiceEntity, status int) (*nigoapi.ControllerServiceEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Controller-level endpoint: the service must be visible to flow registry clients.
	url := nifiAddress(cluster, "/controller/controller-services")
	httpmock.RegisterResponder(http.MethodPost, url,
		func(req *http.Request) (*http.Response, error) {
			var sent nigoapi.ControllerServiceEntity
			if err := json.NewDecoder(req.Body).Decode(&sent); err != nil {
				return httpmock.NewStringResponse(400, err.Error()), nil
			}
			assert.Equal(t, entity.Component.Name, sent.Component.Name)
			assert.Equal(t, entity.Component.Type_, sent.Component.Type_)
			return httpmock.NewJsonResponse(status, entity)
		})

	return client.CreateControllerService(*entity)
}

func TestUpdateControllerService(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockControllerService(mockControllerServiceId, "", "controllerservice-mock", "DISABLED")

	entity, err := testUpdateControllerService(t, &mockEntity, 200)
	assert.Nil(err)
	if assert.NotNil(entity) {
		assert.Equal(mockControllerServiceId, entity.Id)
	}

	entity, err = testUpdateControllerService(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testUpdateControllerService(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testUpdateControllerService(t *testing.T, entity *nigoapi.ControllerServiceEntity, status int) (*nigoapi.ControllerServiceEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller-services/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			var sent nigoapi.ControllerServiceEntity
			if err := json.NewDecoder(req.Body).Decode(&sent); err != nil {
				return httpmock.NewStringResponse(400, err.Error()), nil
			}
			assert.Equal(t, entity.Id, sent.Id)
			assert.Equal(t, *entity.Revision.Version, *sent.Revision.Version)
			return httpmock.NewJsonResponse(status, entity)
		})

	return client.UpdateControllerService(*entity)
}

func TestRemoveControllerService(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockControllerService(mockControllerServiceId, "", "controllerservice-mock", "DISABLED")

	err := testRemoveControllerService(t, &mockEntity, 200)
	assert.Nil(err)

	err = testRemoveControllerService(t, &mockEntity, 404)
	assert.Nil(err)

	err = testRemoveControllerService(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
}

func testRemoveControllerService(t *testing.T, entity *nigoapi.ControllerServiceEntity, status int) error {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller-services/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodDelete, url,
		func(req *http.Request) (*http.Response, error) {
			// NiFi requires the current revision version for optimistic locking.
			assert.Equal(t, strconv.FormatInt(*entity.Revision.Version, 10), req.URL.Query().Get("version"))
			return httpmock.NewJsonResponse(status, entity)
		})

	return client.RemoveControllerService(*entity)
}

func TestUpdateControllerServiceRunStatus(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockControllerService(mockControllerServiceId, "", "controllerservice-mock", "DISABLED")

	entity, err := testUpdateControllerServiceRunStatus(t, &mockEntity, "ENABLED", 200)
	assert.Nil(err)
	if assert.NotNil(entity) {
		assert.Equal("ENABLING", entity.Component.State)
	}

	entity, err = testUpdateControllerServiceRunStatus(t, &mockEntity, "ENABLED", 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testUpdateControllerServiceRunStatus(t, &mockEntity, "ENABLED", 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testUpdateControllerServiceRunStatus(t *testing.T, entity *nigoapi.ControllerServiceEntity, state string, status int) (*nigoapi.ControllerServiceEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller-services/%s/run-status", entity.Id))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			var sent nigoapi.ControllerServiceRunStatusEntity
			if err := json.NewDecoder(req.Body).Decode(&sent); err != nil {
				return httpmock.NewStringResponse(400, err.Error()), nil
			}
			assert.Equal(t, state, sent.State)
			assert.Equal(t, *entity.Revision.Version, *sent.Revision.Version)
			// NiFi answers with the transitional state; the target state is reached later.
			transitioning := MockControllerService(entity.Id, "", entity.Component.Name, "ENABLING")
			return httpmock.NewJsonResponse(status, transitioning)
		})

	return client.UpdateControllerServiceRunStatus(*entity, state)
}

func TestGetControllerServiceReferences(t *testing.T) {
	assert := assert.New(t)

	refs, err := testGetControllerServiceReferences(t, mockControllerServiceId, 200)
	assert.Nil(err)
	if assert.NotNil(refs) && assert.Len(refs.ControllerServiceReferencingComponents, 1) {
		assert.Equal("FlowRegistryClient", refs.ControllerServiceReferencingComponents[0].Component.ReferenceType)
	}

	refs, err = testGetControllerServiceReferences(t, mockControllerServiceId, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(refs)

	refs, err = testGetControllerServiceReferences(t, mockControllerServiceId, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(refs)
}

func testGetControllerServiceReferences(t *testing.T, id string, status int) (*nigoapi.ControllerServiceReferencingComponentsEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/controller-services/%s/references", id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(status, nigoapi.ControllerServiceReferencingComponentsEntity{
				ControllerServiceReferencingComponents: []nigoapi.ControllerServiceReferencingComponentEntity{{
					Id: "a3b4c5d6-0174-1000-0000-00004b9b35cc",
					Component: &nigoapi.ControllerServiceReferencingComponentDto{
						Id:            "a3b4c5d6-0174-1000-0000-00004b9b35cc",
						Name:          "bitbucket-flows",
						ReferenceType: "FlowRegistryClient",
					},
				}},
			})
		})

	return client.GetControllerServiceReferences(id)
}
