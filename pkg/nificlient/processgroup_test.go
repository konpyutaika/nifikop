package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestGetProcessGroup(t *testing.T) {
	assert := assert.New(t)

	id := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	entity, err := testGetProcessGroup(t, id, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetProcessGroup(t, id, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetProcessGroup(t, id, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetProcessGroup(t *testing.T, id string, status int) (*nigoapi.ProcessGroupEntity, error) {

	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/process-groups/%s", id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockProcessGroup(
					id,
					"test-unit",
					"16cfd2ec-0174-1050-0000-00004b9b35cc",
					"16cfd2ec-0174-5445-0000-00004b9b35cc",
					"16cfd2ec-0174-2000-0000-00004b9b35cc",
					"16cfd2ec-0174-3000-0000-00004b9b35cc",
					20))
		})

	return client.GetProcessGroup(id)
}

func TestCreateProcessGroup(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockProcessGroup(
		"16cfd2ec-0174-1000-0000-00004b9b35cc",
		"test-unit",
		"16cfd2ec-0174-1050-0000-00004b9b35cc",
		"16cfd2ec-0174-5445-0000-00004b9b35cc",
		"16cfd2ec-0174-2000-0000-00004b9b35cc",
		"16cfd2ec-0174-3000-0000-00004b9b35cc",
		20)

	entity, err := testCreateProcessGroup(t, &mockEntity, 201)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testCreateProcessGroup(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testCreateProcessGroup(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testCreateProcessGroup(t *testing.T, entity *nigoapi.ProcessGroupEntity, status int) (*nigoapi.ProcessGroupEntity, error) {

	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/process-groups/%s/process-groups", entity.Component.ParentGroupId))
	httpmock.RegisterResponder(http.MethodPost, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.CreateProcessGroup(*entity, entity.Component.ParentGroupId)
}

func TestUpdateProcessGroup(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockProcessGroup(
		"16cfd2ec-0174-1000-0000-00004b9b35cc",
		"test-unit",
		"16cfd2ec-0174-1050-0000-00004b9b35cc",
		"16cfd2ec-0174-5445-0000-00004b9b35cc",
		"16cfd2ec-0174-2000-0000-00004b9b35cc",
		"16cfd2ec-0174-3000-0000-00004b9b35cc",
		20)

	entity, err := testUpdateProcessGroup(t, &mockEntity, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testUpdateProcessGroup(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testUpdateProcessGroup(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testUpdateProcessGroup(t *testing.T, entity *nigoapi.ProcessGroupEntity, status int) (*nigoapi.ProcessGroupEntity, error) {

	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/process-groups/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.UpdateProcessGroup(*entity)
}

func TestRemoveProcessGroup(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockProcessGroup(
		"16cfd2ec-0174-1000-0000-00004b9b35cc",
		"test-unit",
		"16cfd2ec-0174-1050-0000-00004b9b35cc",
		"16cfd2ec-0174-5445-0000-00004b9b35cc",
		"16cfd2ec-0174-2000-0000-00004b9b35cc",
		"16cfd2ec-0174-3000-0000-00004b9b35cc",
		20)

	err := testRemoveProcessGroup(t, &mockEntity, 200)
	assert.Nil(err)

	err = testRemoveProcessGroup(t, &mockEntity, 404)
	assert.Nil(err)

	err = testRemoveProcessGroup(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
}

func testRemoveProcessGroup(t *testing.T, entity *nigoapi.ProcessGroupEntity, status int) error {

	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/process-groups/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodDelete, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.RemoveProcessGroup(*entity)
}

func MockProcessGroup(id, name, parentPGId, registryId, bucketId, flowId string, flowVersion int32) nigoapi.ProcessGroupEntity {
	var version int64 = 10
	return nigoapi.ProcessGroupEntity{
		Id: id,
		Component: &nigoapi.ProcessGroupDto{
			Name:                      name,
			ParentGroupId:             parentPGId,
			VersionControlInformation: MockVersionControlInformationDto(id, registryId, bucketId, flowId, flowVersion),
		},
		Revision: &nigoapi.RevisionDto{Version: &version},
	}
}
