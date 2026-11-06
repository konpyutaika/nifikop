package nificlient

import (
	"fmt"
	"net/http"
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
