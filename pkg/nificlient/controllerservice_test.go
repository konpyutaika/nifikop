package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
)

func TestGetControllerService(t *testing.T) {
	assert := assert.New(t)

	id := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	entity, err := testGetControllerService(t, id, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetControllerService(t, id, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetControllerService(t, id, 500)
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
				MockRootControllerService(id, "controllerservice-mock", "http://uri.com:8888"))
		})

	return client.GetControllerService(id)
}

func MockRootControllerService(id, name, uri string) nigoapi.ControllerServiceEntity {
	var version int64 = 10
	return nigoapi.ControllerServiceEntity{
		Component: &nigoapi.ControllerServiceDto{
			Id:    id,
			State: "DISABLED",
			Name:  name,
		},
		Id:  id,
		Uri: uri,
		Revision: &nigoapi.RevisionDto{
			Version: &version,
		},
	}
}
