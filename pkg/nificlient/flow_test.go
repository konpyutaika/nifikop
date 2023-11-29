package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
)

func TestGetFlow(t *testing.T) {
	assert := assert.New(t)

	id := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	entity, err := testGetFlow(t, id, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetFlow(t, id, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetFlow(t, id, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetFlow(t *testing.T, id string, status int) (*nigoapi.ProcessGroupFlowEntity, error) {
	pgId := "16cfd2ec-2174-1065-0650-10004b9b35cc"
	parameterContext := MockParameterContext("16cfd2ec-0174-1000-0000-00004b9b35cc", "test-unit",
		"unit test",
		map[string]string{"key1": "value1", "key2": "value2"},
		map[string]string{"secret1": "value1", "secret2": "value2"})

	parameterContextRef := nigoapi.ParameterContextReferenceEntity{
		Id: parameterContext.Id,
		Component: &nigoapi.ParameterContextReferenceDto{
			Id:   parameterContext.Id,
			Name: parameterContext.Component.Name,
		},
	}

	processGroups := []nigoapi.ProcessGroupEntity{MockProcessGroup(
		"16cfd2ec-0174-1000-0000-00004b9b35cc",
		"test-unit",
		"16cfd2ec-0174-1050-0000-00004b9b35cc",
		"16cfd2ec-0174-5445-0000-00004b9b35cc",
		"16cfd2ec-0174-2000-0000-00004b9b35cc",
		"16cfd2ec-0174-3000-0000-00004b9b35cc",
		20)}

	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/flow/process-groups/%s", id))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockFlow(id, pgId, &parameterContextRef, processGroups))
		})

	return client.GetFlow(id)
}

func TestGetFlowControllerServices(t *testing.T) {
	assert := assert.New(t)

	id := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	entity, err := testGetFlowControllerServices(t, id, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testGetFlowControllerServices(t, id, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testGetFlowControllerServices(t, id, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testGetFlowControllerServices(t *testing.T, pgId string, status int) (*nigoapi.ControllerServicesEntity, error) {
	cs := []nigoapi.ControllerServiceEntity{
		MockControllerService(
			"16cfd2ec-2174-1065-0650-10004b9b35cc", pgId,
			"unit-test controller 1", "DISABLED"),
		MockControllerService(
			"18cfd2ec-2174-1065-0650-10004b9b35cc", pgId,
			"unit-test controller 2", "ENABLED"),
	}

	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/flow/process-groups/%s/controller-services", pgId))
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				MockFlowControllerServices(cs))
		})

	return client.GetFlowControllerServices(pgId)
}

func TestUpdateFlowControllerServices(t *testing.T) {
	assert := assert.New(t)

	pgId := "16cfd2ec-0174-1000-0000-00004b9b35cc"

	mockEntity := MockFlowControllerServices([]nigoapi.ControllerServiceEntity{
		MockControllerService(
			"16cfd2ec-2174-1065-0650-10004b9b35cc", pgId,
			"unit-test controller 1", "DISABLED"),
		MockControllerService(
			"18cfd2ec-2174-1065-0650-10004b9b35cc", pgId,
			"unit-test controller 2", "ENABLED"),
	})

	entity, err := testUpdateFlowControllerServices(t, &mockEntity, pgId, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testUpdateFlowControllerServices(t, &mockEntity, pgId, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testUpdateFlowControllerServices(t, &mockEntity, pgId, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testUpdateFlowControllerServices(t *testing.T, entity *nigoapi.ControllerServicesEntity, pgId string, status int) (*nigoapi.ActivateControllerServicesEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	acse := MockActivateControllerServicesEntity(pgId, *entity)
	url := nifiAddress(cluster, fmt.Sprintf("/flow/process-groups/%s/controller-services", pgId))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				acse)
		})

	return client.UpdateFlowControllerServices(acse)
}

func TestUpdateFlowProcessGroup(t *testing.T) {
	assert := assert.New(t)

	mockEntity := MockScheduleComponentsEntity("16cfd2ec-2174-1065-0650-10004b9b35cc", "STOPPED")

	entity, err := testUpdateFlowProcessGroup(t, mockEntity, 200)
	assert.Nil(err)
	assert.NotNil(entity)

	entity, err = testUpdateFlowProcessGroup(t, mockEntity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(entity)

	entity, err = testUpdateFlowProcessGroup(t, mockEntity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(entity)
}

func testUpdateFlowProcessGroup(t *testing.T, entity nigoapi.ScheduleComponentsEntity, status int) (*nigoapi.ScheduleComponentsEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/flow/process-groups/%s", entity.Id))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.UpdateFlowProcessGroup(entity)
}

func MockFlowControllerServices(cs []nigoapi.ControllerServiceEntity) nigoapi.ControllerServicesEntity {
	return nigoapi.ControllerServicesEntity{
		ControllerServices: cs,
	}
}

func MockActivateControllerServicesEntity(pgId string, cs nigoapi.ControllerServicesEntity) nigoapi.ActivateControllerServicesEntity {
	components := make(map[string]nigoapi.RevisionDto)
	var version int64 = 10

	state := "ENABLED"
	for _, c := range cs.ControllerServices {
		components[c.Id] = nigoapi.RevisionDto{Version: &version}
		if c.Component.State == "DISABLED" {
			state = c.Component.State
		}
	}
	return nigoapi.ActivateControllerServicesEntity{
		Id:                           pgId,
		State:                        state,
		Components:                   components,
		DisconnectedNodeAcknowledged: false,
	}
}

func MockControllerService(id, pgId, name, state string) nigoapi.ControllerServiceEntity {
	var version int64 = 10
	return nigoapi.ControllerServiceEntity{
		Revision: &nigoapi.RevisionDto{
			Version: &version,
		},
		Id:            id,
		ParentGroupId: pgId,
		Component: &nigoapi.ControllerServiceDto{
			Id:            id,
			ParentGroupId: pgId,
			Name:          name,
			State:         state,
		},
	}
}

func MockFlow(
	id, pgID string,
	parameterContext *nigoapi.ParameterContextReferenceEntity,
	processGroups []nigoapi.ProcessGroupEntity) nigoapi.ProcessGroupFlowEntity {
	return nigoapi.ProcessGroupFlowEntity{
		ProcessGroupFlow: &nigoapi.ProcessGroupFlowDto{
			Id:               id,
			Uri:              "",
			ParentGroupId:    pgID,
			ParameterContext: parameterContext,
			Flow: &nigoapi.FlowDto{
				ProcessGroups: processGroups,
			},
			LastRefreshed: "",
		},
	}
}

func MockScheduleComponentsEntity(id, state string) nigoapi.ScheduleComponentsEntity {
	return nigoapi.ScheduleComponentsEntity{Id: id, State: state}
}
