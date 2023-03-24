package nificlient

import (
	"strconv"

	"github.com/antihax/optional"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

func (n *nifiClient) GetAccessPolicy(action, resource string) (*nigoapi.AccessPolicyEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to get the access policy informations

	for {
		if resource[0:1] == "/" {
			resource = resource[1:]
			continue
		}
		break
	}

	accessPolicyEntity, rsp, body, err := client.PoliciesApi.GetAccessPolicyForResource(context, action, resource)

	if err := errorGetOperation(rsp, body, err, n.log.WithOptions()); err != nil {
		return nil, err
	}

	return &accessPolicyEntity, nil
}

func (n *nifiClient) CreateAccessPolicy(entity nigoapi.AccessPolicyEntity) (*nigoapi.AccessPolicyEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to create the access policy
	accessPolicyEntity, rsp, body, err := client.PoliciesApi.CreateAccessPolicy(context, entity)
	if err := errorCreateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &accessPolicyEntity, nil
}

func (n *nifiClient) UpdateAccessPolicy(entity nigoapi.AccessPolicyEntity) (*nigoapi.AccessPolicyEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to update the access policy
	accessPolicyEntity, rsp, body, err := client.PoliciesApi.UpdateAccessPolicy(context, entity.Id, entity)
	if err := errorUpdateOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &accessPolicyEntity, nil
}

func (n *nifiClient) RemoveAccessPolicy(entity nigoapi.AccessPolicyEntity) error {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return ErrNoNodeClientsAvailable
	}

	// Request on Nifi Rest API to remove the registry client
	_, rsp, body, err := client.PoliciesApi.RemoveAccessPolicy(context, entity.Id,
		&nigoapi.PoliciesApiRemoveAccessPolicyOpts{
			Version: optional.NewString(strconv.FormatInt(*entity.Revision.Version, 10)),
		})

	return errorDeleteOperation(rsp, body, err, n.log)
}
