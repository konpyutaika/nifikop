package nificlient

import (
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	"go.uber.org/zap"
)

func (n *nifiClient) GetControllerService(id string) (*nigoapi.ControllerServiceEntity, error) {
	// Get nigoapi client, favoring the one associated to the coordinator node.
	client, context := n.privilegeCoordinatorClient()
	if client == nil {
		n.log.Error("Error during creating node client", zap.Error(ErrNoNodeClientsAvailable))
		return nil, ErrNoNodeClientsAvailable
	}
	var opts *nigoapi.ControllerServicesApiGetControllerServiceOpts = nil
	// Request on Nifi Rest API to get the controller service informations
	out, rsp, body, err := client.ControllerServicesApi.GetControllerService(context, id, opts)

	if err := errorGetOperation(rsp, body, err, n.log); err != nil {
		return nil, err
	}

	return &out, nil
}
