package nificlient

import (
	"strconv"

	"github.com/antihax/optional"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

func (n *nifiClient) GetParameterContexts() ([]nigoapi.ParameterContextEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the parameter contexts informations
	pcEntity, rsp, body, err := client.FlowApi.GetParameterContexts(context)
	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return pcEntity.ParameterContexts, nil
}

func (n *nifiClient) GetParameterContext(id string) (*nigoapi.ParameterContextEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the parameter context informations
	pcEntity, rsp, body, err := client.ParameterContextsApi.GetParameterContext(
		context,
		id,
		&nigoapi.ParameterContextsApiGetParameterContextOpts{IncludeInheritedParameters: optional.NewBool(false)})
	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &pcEntity, nil
}

func (n *nifiClient) CreateParameterContext(entity nigoapi.ParameterContextEntity) (*nigoapi.ParameterContextEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the parameter context
	pcEntity, rsp, body, err := client.ParameterContextsApi.CreateParameterContext(context, entity)
	if err := errorCreateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &pcEntity, nil
}

func (n *nifiClient) RemoveParameterContext(entity nigoapi.ParameterContextEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the parameter context
	_, rsp, body, err := client.ParameterContextsApi.DeleteParameterContext(context, entity.Id,
		&nigoapi.ParameterContextsApiDeleteParameterContextOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, body, err, n.log)
}

func (n *nifiClient) CreateParameterContextUpdateRequest(contextId string, entity nigoapi.ParameterContextEntity) (*nigoapi.ParameterContextUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the parameter context update request
	request, rsp, body, err := client.ParameterContextsApi.SubmitParameterContextUpdate(context, contextId, entity)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &request, nil
}

func (n *nifiClient) GetParameterContextUpdateRequest(contextId, id string) (*nigoapi.ParameterContextUpdateRequestEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the parameter context update request information
	request, rsp, body, err := client.ParameterContextsApi.GetParameterContextUpdate(context, contextId, id)
	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &request, nil
}
