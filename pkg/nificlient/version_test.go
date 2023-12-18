package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
)

func TestCreateVersionUpdateRequest(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockVersionUpdateRequest(
		"16cfd2ec-0174-1000-0000-00004b9b35cc",
		"16cfd2ec-0174-1450-0000-00004b9b35cc",
		"16cfd2ec-0174-6580-0000-00004b9b35cc",
		"16cfd2ec-0174-10546-0000-00004b9b35cc",
		5)

	entity, err := testCreateVersionUpdateRequest(t, &mockEntity, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testCreateVersionUpdateRequest(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testCreateVersionUpdateRequest(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testCreateVersionUpdateRequest(t *testing.T, entity *nigoapi.VersionControlInformationEntity, status int) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf(
		"/versions/update-requests/process-groups/%s", entity.VersionControlInformation.GroupId))
	httpmock.RegisterResponder(http.MethodPost, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.CreateVersionUpdateRequest(entity.VersionControlInformation.GroupId, *entity)
}

func TestGetVersionUpdateRequest(t *testing.T) {
	assert := assert.New(t)

	id := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	mockEntity := MockVersionUpdateRequest(
		"16cfd2ec-0174-1000-0000-00004b9b35cc",
		"16cfd2ec-0174-1450-0000-00004b9b35cc",
		"16cfd2ec-0174-6580-0000-00004b9b35cc",
		"16cfd2ec-0174-10546-0000-00004b9b35cc",
		5)

	entity, err := testGetVersionUpdateRequest(t, &mockEntity, id, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetVersionUpdateRequest(t, &mockEntity, id, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetVersionUpdateRequest(t, &mockEntity, id, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetVersionUpdateRequest(t *testing.T, entity *nigoapi.VersionControlInformationEntity, id string, status int) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf(
		"/versions/update-requests/%s", id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.GetVersionUpdateRequest(id)
}

func TestCreateVersionRevertRequest(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockVersionRevertRequest(
		"16cfd2ec-0174-1000-0000-00004b9b35cc",
		"16cfd2ec-0174-1450-0000-00004b9b35cc",
		"16cfd2ec-0174-6580-0000-00004b9b35cc",
		"16cfd2ec-0174-10546-0000-00004b9b35cc",
		5)

	entity, err := testCreateVersionRevertRequest(t, &mockEntity, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testCreateVersionRevertRequest(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testCreateVersionRevertRequest(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testCreateVersionRevertRequest(t *testing.T, entity *nigoapi.VersionControlInformationEntity, status int) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf(
		"/versions/revert-requests/process-groups/%s", entity.VersionControlInformation.GroupId))
	httpmock.RegisterResponder(http.MethodPost, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.CreateVersionRevertRequest(entity.VersionControlInformation.GroupId, *entity)
}

func TestGetVersionRevertRequest(t *testing.T) {
	assert := assert.New(t)

	id := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	mockEntity := MockVersionRevertRequest(
		"16cfd2ec-0174-1000-0000-00004b9b35cc",
		"16cfd2ec-0174-1450-0000-00004b9b35cc",
		"16cfd2ec-0174-6580-0000-00004b9b35cc",
		"16cfd2ec-0174-10546-0000-00004b9b35cc",
		5)

	entity, err := testGetVersionRevertRequest(t, &mockEntity, id, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetVersionRevertRequest(t, &mockEntity, id, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetVersionRevertRequest(t, &mockEntity, id, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetVersionRevertRequest(t *testing.T, entity *nigoapi.VersionControlInformationEntity, id string, status int) (*nigoapi.VersionedFlowUpdateRequestEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf(
		"/versions/revert-requests/%s", id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.GetVersionRevertRequest(id)
}

func MockVersionUpdateRequest(pgId, registryId, bucketId, flowId string, flowVersion int32) nigoapi.VersionControlInformationEntity {
	var version int64 = 10
	return nigoapi.VersionControlInformationEntity{
		ProcessGroupRevision: &nigoapi.RevisionDto{
			Version: &version,
		},
		VersionControlInformation: MockVersionControlInformationDto(pgId, registryId, bucketId, flowId, flowVersion),
	}
}

func MockVersionRevertRequest(pgId, registryId, bucketId, flowId string, flowVersion int32) nigoapi.VersionControlInformationEntity {
	var version int64 = 10
	return nigoapi.VersionControlInformationEntity{
		ProcessGroupRevision: &nigoapi.RevisionDto{
			Version: &version,
		},
		VersionControlInformation: MockVersionControlInformationDto(pgId, registryId, bucketId, flowId, flowVersion),
	}
}

func MockVersionControlInformationDto(pgId, registryId, bucketId, flowId string, flowVersion int32) *nigoapi.VersionControlInformationDto {
	return &nigoapi.VersionControlInformationDto{
		GroupId:    pgId,
		RegistryId: registryId,
		BucketId:   bucketId,
		FlowId:     flowId,
		Version:    flowVersion,
	}
}
