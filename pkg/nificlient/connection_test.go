package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
)

func TestGetConnection(t *testing.T) {
	assert := assert.New(t)

	id := "5f1f9f7e-0183-1000-ffff-ffffa1b9c8d5"
	mockEntity := MockConnection(id, "test-unit", "5a859dfd-0183-1000-b22b-680e3b6fb507",
		"41481c3b-a836-37fa-84d1-06e57a6dc2d8", "OUTPUT_PORT", "5eee3064-0183-1000-0000-00004b62d089",
		"b760a6ed-1421-37d6-813d-94cf7cb03524", "INPUT_PORT", "5eee26c7-0183-1000-ffff-fffffc99fdef",
		"1 hour", "10 GB", "DO_NOT_LOAD_BALANCE", "", "DO_NOT_COMPRESS",
		1000, []string{}, 0, []nigoapi.PositionDto{{X: 0, Y: 0}})

	entity, err := testGetConnection(t, id, &mockEntity, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetConnection(t, id, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetConnection(t, id, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetConnection(t *testing.T, id string, entity *nigoapi.ConnectionEntity, status int) (*nigoapi.ConnectionEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/connections/%s", id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.GetConnection(id)
}

func TestUpdateConnection(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockConnection("5f1f9f7e-0183-1000-ffff-ffffa1b9c8d5", "test-unit", "5a859dfd-0183-1000-b22b-680e3b6fb507",
		"41481c3b-a836-37fa-84d1-06e57a6dc2d8", "OUTPUT_PORT", "5eee3064-0183-1000-0000-00004b62d089",
		"b760a6ed-1421-37d6-813d-94cf7cb03524", "INPUT_PORT", "5eee26c7-0183-1000-ffff-fffffc99fdef",
		"1 hour", "10 GB", "DO_NOT_LOAD_BALANCE", "", "DO_NOT_COMPRESS",
		1000, []string{}, 0, []nigoapi.PositionDto{{X: 0, Y: 0}})

	entity, err := testUpdateConnection(t, &mockEntity, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testUpdateConnection(t, &mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testUpdateConnection(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testUpdateConnection(t *testing.T, entity *nigoapi.ConnectionEntity, status int) (*nigoapi.ConnectionEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/connections/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.UpdateConnection(*entity)
}

func TestDeleteConnection(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockConnection("5f1f9f7e-0183-1000-ffff-ffffa1b9c8d5", "test-unit", "5a859dfd-0183-1000-b22b-680e3b6fb507",
		"41481c3b-a836-37fa-84d1-06e57a6dc2d8", "OUTPUT_PORT", "5eee3064-0183-1000-0000-00004b62d089",
		"b760a6ed-1421-37d6-813d-94cf7cb03524", "INPUT_PORT", "5eee26c7-0183-1000-ffff-fffffc99fdef",
		"1 hour", "10 GB", "DO_NOT_LOAD_BALANCE", "", "DO_NOT_COMPRESS",
		1000, []string{}, 0, []nigoapi.PositionDto{{X: 0, Y: 0}})

	err := testDeleteConnection(t, &mockEntity, 200)
	assert.Nil(err)

	err = testDeleteConnection(t, &mockEntity, 404)
	assert.Nil(err)

	err = testDeleteConnection(t, &mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
}

func testDeleteConnection(t *testing.T, entity *nigoapi.ConnectionEntity, status int) error {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/connections/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodDelete, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.DeleteConnection(*entity)
}

func MockConnection(
	id, name, parentId, srcId, srcType, srcGroupId, dstId, dstType, dstGroupId,
	flowfileExpiration, backPressureDataSizeThreshold, loadBalanceStrategy, loadBalancePartitionAttribute, loadBalanceCompression string,
	backPressureObjectThreshold int64,
	prioritizers []string,
	labelIndex int32, bends []nigoapi.PositionDto) nigoapi.ConnectionEntity {
	var version int64 = 10
	return nigoapi.ConnectionEntity{
		Id: id,
		Component: &nigoapi.ConnectionDto{
			Name:          name,
			Id:            id,
			ParentGroupId: parentId,
			Source: &nigoapi.ConnectableDto{
				Id:      srcId,
				Type_:   srcType,
				GroupId: srcGroupId,
			},
			Destination: &nigoapi.ConnectableDto{
				Id:      dstId,
				Type_:   dstType,
				GroupId: dstGroupId,
			},
			FlowFileExpiration:            flowfileExpiration,
			BackPressureDataSizeThreshold: backPressureDataSizeThreshold,
			BackPressureObjectThreshold:   backPressureObjectThreshold,
			LoadBalanceStrategy:           loadBalanceStrategy,
			LoadBalancePartitionAttribute: loadBalancePartitionAttribute,
			LoadBalanceCompression:        loadBalanceCompression,
			Prioritizers:                  prioritizers,
			LabelIndex:                    labelIndex,
			Bends:                         bends,
		},
		Revision: &nigoapi.RevisionDto{Version: &version},
	}
}
