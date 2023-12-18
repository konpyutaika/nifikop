package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
)

func TestGetOutputPort(t *testing.T) {
	assert := assert.New(t)

	id := "41481c3b-a836-37fa-84d1-06e57a6dc2d8"
	mockEntity := MockOutputPort(id, "test-unit", "5eee3064-0183-1000-0000-00004b62d089", 0, 0)

	entity, err := testGetOutputPort(t, id, &mockEntity, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetOutputPort(t, id, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetOutputPort(t, id, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetOutputPort(t *testing.T, id string, entity *nigoapi.PortEntity, status int) (*nigoapi.PortEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/output-ports/%s", id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.GetOutputPort(id)
}

func TestUpdateOutputPortRunStatus(t *testing.T) {
	assert := assert.New(t)

	id := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	mockEntity := MockOutputPortRunStatus("Stopped")

	entity, err := testUpdateOutputPortRunStatus(t, mockEntity, id, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testUpdateOutputPortRunStatus(t, mockEntity, id, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testUpdateOutputPortRunStatus(t, mockEntity, id, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testUpdateOutputPortRunStatus(t *testing.T, entity nigoapi.PortRunStatusEntity, id string, status int) (*nigoapi.ProcessorEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/output-ports/%s/run-status", id))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.UpdateOutputPortRunStatus(id, entity)
}

func MockOutputPortRunStatus(state string) nigoapi.PortRunStatusEntity {
	var version int64 = 10
	return nigoapi.PortRunStatusEntity{
		Revision: &nigoapi.RevisionDto{Version: &version},
		State:    state,
	}
}

func MockOutputPort(id, name, parentId string, posX, posY float64) nigoapi.PortEntity {
	var version int64 = 10
	return nigoapi.PortEntity{
		Id: id,
		Component: &nigoapi.PortDto{
			Id:            id,
			Name:          name,
			ParentGroupId: parentId,
			Type_:         "OUTPUT_PORT",
			Position: &nigoapi.PositionDto{
				X: posX,
				Y: posY,
			},
		},
		Revision: &nigoapi.RevisionDto{Version: &version},
	}
}
