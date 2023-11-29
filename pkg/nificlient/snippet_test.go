package nificlient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
)

func TestCreateSnippet(t *testing.T) {
	assert := assert.New(t)

	entity := MockSnippet("16cfd2ec-0174-1000-0000-00004b9b35cc", "456fd2ec-0174-1000-0340-00004b9b35cc")

	snippetEntity, err := testCreateSnippet(t, entity, 201)
	assert.Nil(err)
	assert.NotNil(snippetEntity)

	snippetEntity, err = testCreateSnippet(t, entity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(snippetEntity)

	snippetEntity, err = testCreateSnippet(t, entity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(snippetEntity)
}

func testCreateSnippet(t *testing.T, entity nigoapi.SnippetEntity, status int) (*nigoapi.SnippetEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, "/snippets")
	httpmock.RegisterResponder(http.MethodPost, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.CreateSnippet(entity)
}

func TestUpdateSnippet(t *testing.T) {
	assert := assert.New(t)

	entity := MockSnippet("16cfd2ec-0174-1000-0000-00004b9b35cc", "456fd2ec-0174-1000-0340-00004b9b35cc")

	snippetEntity, err := testUpdateSnippet(t, entity, 200)
	assert.Nil(err)
	assert.NotNil(snippetEntity)

	snippetEntity, err = testUpdateSnippet(t, entity, 404)
	assert.IsType(ErrNifiClusterReturned404, err)
	assert.Nil(snippetEntity)

	snippetEntity, err = testUpdateSnippet(t, entity, 500)
	assert.IsType(ErrNifiClusterNotReturned200, err)
	assert.Nil(snippetEntity)
}

func testUpdateSnippet(t *testing.T, entity nigoapi.SnippetEntity, status int) (*nigoapi.SnippetEntity, error) {
	cluster := testClusterMock(t)

	client, err := testClientFromCluster(cluster, false)
	if err != nil {
		return nil, err
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := nifiAddress(cluster, fmt.Sprintf("/snippets/%s", entity.Snippet.Id))
	httpmock.RegisterResponder(http.MethodPut, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				status,
				entity)
		})

	return client.UpdateSnippet(entity)
}

func MockSnippet(pgID, parentGroupID string) nigoapi.SnippetEntity {
	return nigoapi.SnippetEntity{
		Snippet: &nigoapi.SnippetDto{
			ParentGroupId: parentGroupID,
			ProcessGroups: map[string]nigoapi.RevisionDto{pgID: {}},
		},
	}
}
