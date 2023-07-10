package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
)

func TestGetLabel(t *testing.T) {
	assert := assert.New(t)

	id := "41481c3b-a836-37fa-84d1-06e57a6dc2d8"
	mockEntity := MockLabel(id, "test-unit", "5eee3064-0183-1000-0000-00004b62d089", 100, 100,
		0, 0, map[string]string{"font-size": "18px", "background-color": "#FFF"})

	entity, err := testGetLabel(t, id, &mockEntity, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetLabel(t, id, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetLabel(t, id, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetLabel(t *testing.T, id string, entity *nigoapi.LabelEntity, status int) (*nigoapi.LabelEntity, error) {

	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/labels/%s", id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.GetLabel(id)
}

func TestUpdateLabel(t *testing.T) {
	assert := assert.New(t)

	id := "41481c3b-a836-37fa-84d1-06e57a6dc2d8"
	mockEntity := MockLabel(id, "test-unit", "5eee3064-0183-1000-0000-00004b62d089", 100, 100,
		0, 0, map[string]string{"font-size": "18px", "background-color": "#FFF"})

	entity, err := testUpdateLabel(t, &mockEntity, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testUpdateLabel(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testUpdateLabel(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testUpdateLabel(t *testing.T, entity *nigoapi.LabelEntity, status int) (*nigoapi.LabelEntity, error) {

	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/labels/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.UpdateLabel(*entity)
}

func TestRemoveLabel(t *testing.T) {
	assert := assert.New(t)

	id := "41481c3b-a836-37fa-84d1-06e57a6dc2d8"
	mockEntity := MockLabel(id, "test-unit", "5eee3064-0183-1000-0000-00004b62d089", 100, 100,
		0, 0, map[string]string{"font-size": "18px", "background-color": "#FFF"})

	err := testRemoveLabel(t, &mockEntity, 200)
	assert.Nil(err)

	err = testRemoveLabel(t, &mockEntity, 404)
	assert.Nil(err)

	err = testRemoveLabel(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
}

func testRemoveLabel(t *testing.T, entity *nigoapi.LabelEntity, status int) error {

	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/labels/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodDelete, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.RemoveLabel(*entity)
}

func MockLabel(id, label, parentId string, width, height, posX, posY float64, style map[string]string) nigoapi.LabelEntity {
	var version int64 = 10
	return nigoapi.LabelEntity{
		Id: id,
		Component: &nigoapi.LabelDto{
			Id:            id,
			Label:         label,
			ParentGroupId: parentId,
			Position: &nigoapi.PositionDto{
				X: posX,
				Y: posY,
			},
			Width:  width,
			Height: height,
			Style:  style,
		},
		Revision: &nigoapi.RevisionDto{Version: &version},
	}
}
