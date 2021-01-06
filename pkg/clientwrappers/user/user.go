package user

import (
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/accesspolicies"
	"github.com/Orange-OpenSource/nifikop/pkg/common"
	"github.com/Orange-OpenSource/nifikop/pkg/nificlient"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log = ctrl.Log.WithName("user-method")

func ExistUser(client client.Client, user *v1alpha1.NifiUser,
	cluster *v1alpha1.NifiCluster) (bool, error) {

	if user.Status.Id == "" {
		return false, nil
	}

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetUser(user.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get user"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

func FindUserByIdentity(client client.Client, user *v1alpha1.NifiUser,
	cluster *v1alpha1.NifiCluster) (*v1alpha1.NifiUserStatus, error) {

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return nil, err
	}

	entities, err := nClient.GetUsers()
	if err := clientwrappers.ErrorGetOperation(log, err, "Get users"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil, nil
		}
		return nil, err
	}

	for _, entity := range entities {
		if user.GetIdentity() == entity.Component.Identity {
			return &v1alpha1.NifiUserStatus{
				Id:      entity.Id,
				Version: *entity.Revision.Version,
			}, nil
		}
	}

	return nil, nil
}

func CreateUser(client client.Client, user *v1alpha1.NifiUser,
	cluster *v1alpha1.NifiCluster) (*v1alpha1.NifiUserStatus, error) {

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.UserEntity{}
	updateUserEntity(user, &scratchEntity)

	entity, err := nClient.CreateUser(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Create user"); err != nil {
		return nil, err
	}

	return &v1alpha1.NifiUserStatus{
		Id:      entity.Id,
		Version: *entity.Revision.Version,
	}, nil
}

func SyncUser(client client.Client, user *v1alpha1.NifiUser,
	cluster *v1alpha1.NifiCluster) (*v1alpha1.NifiUserStatus, error) {

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return nil, err
	}

	entity, err := nClient.GetUser(user.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get user"); err != nil {
		return nil, err
	}

	if !userIsSync(user, entity) {
		updateUserEntity(user, entity)
		entity, err = nClient.UpdateUser(*entity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update user"); err != nil {
			return nil, err
		}
	}

	status := user.Status
	status.Version = *entity.Revision.Version
	status.Id = entity.Id

	// Remove from access policy
	for _, entity := range entity.Component.AccessPolicies {
		contains := false
		for _, accessPolicy := range user.Spec.AccessPolicies {
			if entity.Component.Action == string(accessPolicy.Action) &&
				entity.Component.Resource == accessPolicy.GetResource(cluster) {
				contains = true
				break
			}
		}
		if !contains {
			if err := accesspolicies.UpdateAccessPolicyEntity(client,
				&nigoapi.AccessPolicyEntity{
					Component: &nigoapi.AccessPolicyDto{
						Id:       entity.Component.Id,
						Resource: entity.Component.Resource,
						Action:   entity.Component.Action,
					},
				},
				[]*v1alpha1.NifiUser{}, []*v1alpha1.NifiUser{user},
				[]*v1alpha1.NifiUserGroup{}, []*v1alpha1.NifiUserGroup{}, cluster); err != nil {
				return &status, err
			}
		}
	}

	// add
	for _, accessPolicy := range user.Spec.AccessPolicies {
		contains := false
		for _, entity := range entity.Component.AccessPolicies {
			if entity.Component.Action == string(accessPolicy.Action) &&
				entity.Component.Resource == accessPolicy.GetResource(cluster) {
				contains = true
				break
			}
		}
		if !contains {
			if err := accesspolicies.UpdateAccessPolicy(client, &accessPolicy,
				[]*v1alpha1.NifiUser{user}, []*v1alpha1.NifiUser{},
				[]*v1alpha1.NifiUserGroup{}, []*v1alpha1.NifiUserGroup{}, cluster); err != nil {
				return &status, err
			}
		}
	}

	return &status, nil
}

func RemoveUser(client client.Client, user *v1alpha1.NifiUser, cluster *v1alpha1.NifiCluster) error {
	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return err
	}

	entity, err := nClient.GetUser(user.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get user"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil
		}
		return err
	}

	updateUserEntity(user, entity)
	err = nClient.RemoveUser(*entity)

	return clientwrappers.ErrorRemoveOperation(log, err, "Remove user")
}

func userIsSync(user *v1alpha1.NifiUser, entity *nigoapi.UserEntity) bool {
	return user.GetIdentity() == entity.Component.Identity
}

func updateUserEntity(user *v1alpha1.NifiUser, entity *nigoapi.UserEntity) {

	var defaultVersion int64 = 0

	if entity == nil {
		entity = &nigoapi.UserEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.UserDto{}
	}

	entity.Component.Identity = user.GetIdentity()
}
