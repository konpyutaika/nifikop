package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePortRunStatus(t *testing.T) {
	assert := assert.New(t)

	id := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	mockEntity := MockPortRunStatus("Stopped")

	entity, err := testUpdatePortRunStatus(t, mockEntity, id, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testUpdatePortRunStatus(t, mockEntity, id, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testUpdatePortRunStatus(t, mockEntity, id, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testUpdatePortRunStatus(t *testing.T, entity nigoapi.PortRunStatusEntity, id string, status int) (*nigoapi.ProcessorEntity, error) {

	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/input-ports/%s/run-status", id))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.UpdateInputPortRunStatus(id, entity)
}

func MockPortRunStatus(state string) nigoapi.PortRunStatusEntity {
	var version int64 = 10
	return nigoapi.PortRunStatusEntity{
		Revision: &nigoapi.RevisionDto{Version: &version},
		State:    state,
	}
}
