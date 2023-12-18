package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
)

func TestCreateDropRequest(t *testing.T) {
	assert := assert.New(t)

	connectionId := "16cfd2ec-0174-1000-0000-54654754c"
	mockEntity := MockDropRequest(
		"16cfd2ec-0174-1000-0000-00004b9b35cc", connectionId, "",
		"", "", 50, 10, 15, 5, false)

	entity, err := testCreateDropRequest(t, &mockEntity, connectionId, 202)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testCreateDropRequest(t, &mockEntity, connectionId, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testCreateDropRequest(t, &mockEntity, connectionId, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testCreateDropRequest(t *testing.T, entity *nigoapi.DropRequestEntity, connectionId string, status int) (*nigoapi.DropRequestEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf(
		"/flowfile-queues/%s/drop-requests", connectionId))
	httpmock.RegisterResponder(http.MethodPost, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.CreateDropRequest(connectionId)
}

func TestGetDropRequest(t *testing.T) {
	assert := assert.New(t)

	connectionId := "16cfd2ec-0174-1000-0000-54654754c"
	mockEntity := MockDropRequest(
		"16cfd2ec-0174-1000-0000-00004b9b35cc", connectionId, "",
		"", "", 50, 10, 15, 5, false)

	entity, err := testGetDropRequest(t, &mockEntity, connectionId, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetDropRequest(t, &mockEntity, connectionId, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetDropRequest(t, &mockEntity, connectionId, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetDropRequest(t *testing.T, entity *nigoapi.DropRequestEntity, connectionId string, status int) (*nigoapi.DropRequestEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf(
		"/flowfile-queues/%s/drop-requests/%s", connectionId, entity.DropRequest.Id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.GetDropRequest(connectionId, entity.DropRequest.Id)
}

func MockDropRequest(
	id, connectionId, lastUpdated, failureReason, state string,
	percentCompleted, currentCount, originalCount, droppedCount int32, finished bool) nigoapi.DropRequestEntity {
	return nigoapi.DropRequestEntity{DropRequest: &nigoapi.DropRequestDto{
		Id:               id,
		Uri:              fmt.Sprintf("http://testunit.com:8080/nifi-api/flowfile-queues/%s/drop-requests/%s", connectionId, id),
		SubmissionTime:   "",
		LastUpdated:      lastUpdated,
		PercentCompleted: percentCompleted,
		Finished:         finished,
		FailureReason:    failureReason,
		CurrentCount:     currentCount,
		CurrentSize:      int64(currentCount * 250),
		Current:          "",
		OriginalCount:    originalCount,
		OriginalSize:     int64(originalCount * 250),
		Original:         "",
		DroppedCount:     droppedCount,
		DroppedSize:      int64(droppedCount * 250),
		Dropped:          "",
		State:            state,
	}}
}
