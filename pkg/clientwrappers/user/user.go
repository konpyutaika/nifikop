package user

import (
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/accesspolicies"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers/usergroup"
	"github.com/Orange-OpenSource/nifikop/pkg/common"
	"github.com/Orange-OpenSource/nifikop/pkg/nificlient"
	"github.com/Orange-OpenSource/nifikop/pkg/util/clientconfig"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("user-method")

func ExistUser(user *v1alpha1.NifiUser, config *clientconfig.NifiConfig) (bool, error) {

	if user.Status.Id == "" {
		return false, nil
	}

	nClient, err := common.NewClusterConnection(log, config)
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

func FindUserByIdentity(user *v1alpha1.NifiUser, config *clientconfig.NifiConfig) (*v1alpha1.NifiUserStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
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

func CreateUser(user *v1alpha1.NifiUser, config *clientconfig.NifiConfig) (*v1alpha1.NifiUserStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
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

func SyncUser(user *v1alpha1.NifiUser, config *clientconfig.NifiConfig) (*v1alpha1.NifiUserStatus, error) {

	nClient, err := common.NewClusterConnection(log, config)
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
	for _, ent := range entity.Component.AccessPolicies {
		contains := false
		for _, group := range entity.Component.UserGroups {
			userGroupEntity, err := nClient.GetUserGroup(group.Id)
			if err := clientwrappers.ErrorGetOperation(log, err, "Get user-group"); err != nil {
				return nil, err
			}

			if userGroupEntityContainsAccessPolicyEntity(userGroupEntity, ent) {
				contains = true
				break
			}
		}
		if !contains && !userContainsAccessPolicy(user, ent, config.RootProcessGroupId) {
			if err := accesspolicies.UpdateAccessPolicyEntity(
				&nigoapi.AccessPolicyEntity{
					Component: &nigoapi.AccessPolicyDto{
						Id:       ent.Component.Id,
						Resource: ent.Component.Resource,
						Action:   ent.Component.Action,
					},
				},
				[]*v1alpha1.NifiUser{}, []*v1alpha1.NifiUser{user},
				[]*v1alpha1.NifiUserGroup{}, []*v1alpha1.NifiUserGroup{}, config); err != nil {
				return &status, err
			}
		}
	}

	// add
	for _, accessPolicy := range user.Spec.AccessPolicies {
		contains := false
		for _, group := range entity.Component.UserGroups {
			userGroupEntity, err := nClient.GetUserGroup(group.Id)
			if err := clientwrappers.ErrorGetOperation(log, err, "Get user-group"); err != nil {
				return nil, err
			}

			if usergroup.UserGroupEntityContainsAccessPolicy(userGroupEntity, accessPolicy, config.RootProcessGroupId) {
				contains = true
				break
			}
		}
		if !contains && !userEntityContainsAccessPolicy(entity, accessPolicy, config.RootProcessGroupId) {
			if err := accesspolicies.UpdateAccessPolicy(&accessPolicy,
				[]*v1alpha1.NifiUser{user}, []*v1alpha1.NifiUser{},
				[]*v1alpha1.NifiUserGroup{}, []*v1alpha1.NifiUserGroup{}, config); err != nil {
				return &status, err
			}
		}
	}

	return &status, nil
}

func RemoveUser(user *v1alpha1.NifiUser, config *clientconfig.NifiConfig) error {
	nClient, err := common.NewClusterConnection(log, config)
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

func userContainsAccessPolicy(user *v1alpha1.NifiUser, entity nigoapi.AccessPolicySummaryEntity, rootPGId string) bool {
	for _, accessPolicy := range user.Spec.AccessPolicies {
		if entity.Component.Action == string(accessPolicy.Action) &&
			entity.Component.Resource == accessPolicy.GetResource(rootPGId) {
			return true
		}
	}
	return false
}

func userEntityContainsAccessPolicy(entity *nigoapi.UserEntity, accessPolicy v1alpha1.AccessPolicy, rootPGId string) bool {
	for _, entity := range entity.Component.AccessPolicies {
		if entity.Component.Action == string(accessPolicy.Action) &&
			entity.Component.Resource == accessPolicy.GetResource(rootPGId) {
			return true
		}
	}
	return false
}

func userGroupEntityContainsAccessPolicyEntity(entity *nigoapi.UserGroupEntity, accessPolicy nigoapi.AccessPolicySummaryEntity) bool {
	for _, entity := range entity.Component.AccessPolicies {
		if entity.Component.Action == accessPolicy.Component.Action &&
			entity.Component.Resource == accessPolicy.Component.Resource {
			return true
		}
	}
	return false
}
