package controllerservice

import (
	"errors"
	"testing"

	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
)

func TestPlanSync(t *testing.T) {
	tests := []struct {
		state  string
		inSync bool
		want   action
	}{
		{"ENABLED", true, actionNone},
		{"ENABLED", false, actionDisable},
		{"DISABLED", true, actionEnable},
		{"DISABLED", false, actionUpdateAndEnable},
		{"ENABLING", true, actionNone},
		{"ENABLING", false, actionNone},
		{"DISABLING", true, actionNone},
		{"DISABLING", false, actionNone},
		{"", false, actionUpdateAndEnable},
	}
	for _, tt := range tests {
		assert.Equalf(t, tt.want, planSync(tt.state, tt.inSync), "state=%q inSync=%v", tt.state, tt.inSync)
	}
}

func TestPlanRemove(t *testing.T) {
	tests := []struct {
		state      string
		references int
		want       action
	}{
		{"ENABLED", 1, actionRefuse},
		{"DISABLED", 2, actionRefuse},
		{"ENABLED", 0, actionDisable},
		{"ENABLING", 0, actionWait},
		{"DISABLING", 0, actionWait},
		{"DISABLED", 0, actionRemove},
		{"", 0, actionRemove},
	}
	for _, tt := range tests {
		assert.Equalf(t, tt.want, planRemove(tt.state, tt.references), "state=%q references=%d", tt.state, tt.references)
	}
}

func webClientService(name, description string, spec *v1alpha1.StandardWebClientServiceProviderSpec) *v1alpha1.NifiControllerService {
	return &v1alpha1.NifiControllerService{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1alpha1.NifiControllerServiceSpec{
			Description:                          description,
			Type:                                 v1alpha1.StandardWebClientServiceProviderType,
			StandardWebClientServiceProviderSpec: spec,
		},
	}
}

func webClientEntity(name, description, connect, read, write string) *nigoapi.ControllerServiceEntity {
	var version int64 = 3
	return &nigoapi.ControllerServiceEntity{
		Id:       "16cfd2ec-0174-1000-0000-00004b9b35cc",
		Revision: &nigoapi.RevisionDto{Version: &version},
		Component: &nigoapi.ControllerServiceDto{
			Name:     name,
			Type_:    string(v1alpha1.StandardWebClientServiceProviderType),
			Comments: description,
			State:    "ENABLED",
			Properties: map[string]string{
				"Connect Timeout":     connect,
				"Read Timeout":        read,
				"Write Timeout":       write,
				"SSL Context Service": "ssl-context-id",
			},
		},
	}
}

func TestControllerServiceIsSync(t *testing.T) {
	spec := &v1alpha1.StandardWebClientServiceProviderSpec{ConnectTimeout: "10 secs", ReadTimeout: "20 secs", WriteTimeout: "30 secs"}

	tests := []struct {
		name   string
		cs     *v1alpha1.NifiControllerService
		entity *nigoapi.ControllerServiceEntity
		want   bool
	}{
		{"identical", webClientService("web-client", "shared http client", spec), webClientEntity("web-client", "shared http client", "10 secs", "20 secs", "30 secs"), true},
		{"name differs", webClientService("other", "shared http client", spec), webClientEntity("web-client", "shared http client", "10 secs", "20 secs", "30 secs"), false},
		{"description differs", webClientService("web-client", "changed", spec), webClientEntity("web-client", "shared http client", "10 secs", "20 secs", "30 secs"), false},
		{"timeout differs", webClientService("web-client", "shared http client", spec), webClientEntity("web-client", "shared http client", "10 secs", "20 secs", "5 secs"), false},
		{"type differs", webClientService("web-client", "shared http client", spec), func() *nigoapi.ControllerServiceEntity {
			e := webClientEntity("web-client", "shared http client", "10 secs", "20 secs", "30 secs")
			e.Component.Type_ = "org.apache.nifi.ssl.StandardSSLContextService"
			return e
		}(), false},
		{"nil provider spec ignores properties", webClientService("web-client", "shared http client", nil), webClientEntity("web-client", "shared http client", "1 secs", "1 secs", "1 secs"), true},
		{"nil component", webClientService("web-client", "shared http client", spec), &nigoapi.ControllerServiceEntity{}, false},
	}
	for _, tt := range tests {
		assert.Equalf(t, tt.want, controllerServiceIsSync(tt.cs, tt.entity), tt.name)
	}
}

func TestUpdateControllerServiceEntity(t *testing.T) {
	assert := assert.New(t)
	spec := &v1alpha1.StandardWebClientServiceProviderSpec{ConnectTimeout: "10 secs", ReadTimeout: "20 secs", WriteTimeout: "30 secs"}
	cs := webClientService("web-client", "shared http client", spec)

	t.Run("fresh entity for creation", func(t *testing.T) {
		entity := &nigoapi.ControllerServiceEntity{}
		updateControllerServiceEntity(cs, entity)

		assert.Equal(int64(0), *entity.Revision.Version)
		assert.Equal(string(v1alpha1.StandardWebClientServiceProviderType), entity.Component.Type_)
		assert.Equal("web-client", entity.Component.Name)
		assert.Equal("shared http client", entity.Component.Comments)
		assert.Equal(map[string]string{
			"Connect Timeout": "10 secs",
			"Read Timeout":    "20 secs",
			"Write Timeout":   "30 secs",
		}, entity.Component.Properties)
	})

	t.Run("existing entity keeps revision and only carries managed properties", func(t *testing.T) {
		entity := webClientEntity("web-client", "old", "1 secs", "1 secs", "1 secs")
		updateControllerServiceEntity(cs, entity)

		assert.Equal(int64(3), *entity.Revision.Version)
		assert.Equal("shared http client", entity.Component.Comments)
		assert.Equal("30 secs", entity.Component.Properties["Write Timeout"])
		// Unmanaged properties are not sent back: NiFi merges the map and keeps them unchanged.
		_, hasSSL := entity.Component.Properties["SSL Context Service"]
		assert.False(hasSSL)
	})

	t.Run("nil provider spec sends no properties", func(t *testing.T) {
		entity := &nigoapi.ControllerServiceEntity{}
		updateControllerServiceEntity(webClientService("web-client", "", nil), entity)
		assert.Empty(entity.Component.Properties)
	})
}

func TestExtractStatus(t *testing.T) {
	assert := assert.New(t)

	entity := webClientEntity("web-client", "", "10 secs", "10 secs", "10 secs")
	entity.Component.State = "ENABLING"
	status := extractStatus(entity)
	assert.Equal("16cfd2ec-0174-1000-0000-00004b9b35cc", status.Id)
	assert.Equal(int64(3), status.Version)
	assert.Equal(v1alpha1.ControllerServiceStateEnabling, status.State)

	status = extractStatus(&nigoapi.ControllerServiceEntity{Id: "x"})
	assert.Equal("x", status.Id)
	assert.Equal(int64(0), status.Version)
	assert.Equal(v1alpha1.NifiControllerServiceState(""), status.State)
}

type stubLookup struct {
	registryClients map[string]bool
	reportingTasks  map[string]bool
	err             error
}

func (s stubLookup) GetRegistryClient(id string) (*nigoapi.FlowRegistryClientEntity, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.registryClients[id] {
		return &nigoapi.FlowRegistryClientEntity{Id: id}, nil
	}
	return nil, nificlient.ErrNifiClusterReturned404
}

func (s stubLookup) GetReportingTask(id string) (*nigoapi.ReportingTaskEntity, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.reportingTasks[id] {
		return &nigoapi.ReportingTaskEntity{Id: id}, nil
	}
	return nil, nificlient.ErrNifiClusterReturned404
}

func reference(id, name, refType string) nigoapi.ControllerServiceReferencingComponentEntity {
	return nigoapi.ControllerServiceReferencingComponentEntity{
		Id:        id,
		Component: &nigoapi.ControllerServiceReferencingComponentDto{Id: id, Name: name, ReferenceType: refType},
	}
}

func TestLiveReferences(t *testing.T) {
	lookup := stubLookup{
		registryClients: map[string]bool{"rc-live": true},
		reportingTasks:  map[string]bool{"rt-live": true},
	}
	refs := []nigoapi.ControllerServiceReferencingComponentEntity{
		reference("rc-live", "bitbucket-flows", "FlowRegistryClient"),
		reference("rc-stale", "deleted-client", "FlowRegistryClient"),
		reference("rt-live", "site-to-site", "ReportingTask"),
		reference("rt-stale", "deleted-task", "ReportingTask"),
		reference("rule-1", "analysis-rule", "FlowAnalysisRule"),
		{Id: "no-component"},
	}

	live, err := liveReferences(lookup, refs)
	require.NoError(t, err)

	var ids []string
	for _, r := range live {
		ids = append(ids, r.Id)
	}
	assert.Equal(t, []string{"rc-live", "rt-live", "rule-1", "no-component"}, ids,
		"stale lookups are dropped, unknown types are kept conservatively")
	assert.Equal(t, "bitbucket-flows (FlowRegistryClient), site-to-site (ReportingTask), analysis-rule (FlowAnalysisRule), no-component", referenceNames(live))
}

func TestLiveReferencesLookupError(t *testing.T) {
	_, err := liveReferences(stubLookup{err: errors.New("nifi unavailable")},
		[]nigoapi.ControllerServiceReferencingComponentEntity{reference("rc", "x", "FlowRegistryClient")})
	assert.Error(t, err, "a failed existence check must not be mistaken for a stale reference")
}
