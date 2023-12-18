package usergroup

import (
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"

	"github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers"
	"github.com/konpyutaika/nifikop/pkg/clientwrappers/accesspolicies"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

var log = common.CustomLogger().Named("usergroup-method")

func ExistUserGroup(userGroup *v1.NifiUserGroup, config *clientconfig.NifiConfig) (bool, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entities, err := nClient.GetUserGroups()
	if err := clientwrappers.ErrorGetOperation(log, err, "Get user-groups"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	for _, entity := range entities {
		if entity.Component.Identity == userGroup.GetIdentity() {
			return true, nil
		}
	}

	return false, nil
}

func CreateUserGroup(userGroup *v1.NifiUserGroup,
	users []*v1.NifiUser, config *clientconfig.NifiConfig) (*v1.NifiUserGroupStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	scratchEntity := nigoapi.UserGroupEntity{}
	updateUserGroupEntity(userGroup, users, &scratchEntity)

	entity, err := nClient.CreateUserGroup(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Create user-group"); err != nil {
		return nil, err
	}

	return &v1.NifiUserGroupStatus{
		Id:      entity.Id,
		Version: *entity.Revision.Version,
	}, nil
}

func SyncUserGroup(userGroup *v1.NifiUserGroup, users []*v1.NifiUser,
	config *clientconfig.NifiConfig) (*v1.NifiUserGroupStatus, error) {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return nil, err
	}

	var entity *nigoapi.UserGroupEntity
	if userGroup.Status.Id == "" {
		entities, err := nClient.GetUserGroups()
		if err := clientwrappers.ErrorGetOperation(log, err, "Get user-groups"); err != nil {
			if err == nificlient.ErrNifiClusterReturned404 {
				return nil, nil
			}
			return nil, err
		}

		for _, e := range entities {
			if e.Component.Identity == userGroup.GetIdentity() {
				entity = &e
				userGroup.Status.Id = e.Component.Id
				break
			}
		}
	} else {
		entity, err = nClient.GetUserGroup(userGroup.Status.Id)
		if err := clientwrappers.ErrorGetOperation(log, err, "Get user-group"); err != nil {
			return nil, err
		}
	}

	if !userGroupIsSync(userGroup, users, entity) {
		updateUserGroupEntity(userGroup, users, entity)
		entity, err = nClient.UpdateUserGroup(*entity)
		if err := clientwrappers.ErrorUpdateOperation(log, err, "Update user-group"); err != nil {
			return nil, err
		}
	}

	status := userGroup.Status
	status.Version = *entity.Revision.Version
	status.Id = entity.Id

	// Remove from access policy
	for _, entity := range entity.Component.AccessPolicies {
		contains := userGroupContainsAccessPolicy(userGroup, entity, config.RootProcessGroupId)
		if !contains {
			if err := accesspolicies.UpdateAccessPolicyEntity(&entity,
				[]*v1.NifiUser{}, []*v1.NifiUser{},
				[]*v1.NifiUserGroup{}, []*v1.NifiUserGroup{userGroup}, config); err != nil {
				return &status, err
			}
		}
	}

	// add
	for _, accessPolicy := range userGroup.Spec.AccessPolicies {
		contains := UserGroupEntityContainsAccessPolicy(entity, accessPolicy, config.RootProcessGroupId)
		if !contains {
			if err := accesspolicies.UpdateAccessPolicy(&accessPolicy,
				[]*v1.NifiUser{}, []*v1.NifiUser{},
				[]*v1.NifiUserGroup{userGroup}, []*v1.NifiUserGroup{}, config); err != nil {
				return &status, err
			}
		}
	}

	return &status, nil
}

func RemoveUserGroup(userGroup *v1.NifiUserGroup, users []*v1.NifiUser, config *clientconfig.NifiConfig) error {
	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	entity, err := nClient.GetUserGroup(userGroup.Status.Id)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get user-group"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return nil
		}
		return err
	}

	updateUserGroupEntity(userGroup, users, entity)
	err = nClient.RemoveUserGroup(*entity)

	return clientwrappers.ErrorRemoveOperation(log, err, "Remove user-group")
}

func userGroupIsSync(
	userGroup *v1.NifiUserGroup,
	users []*v1.NifiUser,
	entity *nigoapi.UserGroupEntity) bool {
	if userGroup.GetIdentity() != entity.Component.Identity {
		return false
	}

	for _, expected := range users {
		notFound := true
		for _, tenant := range entity.Component.Users {
			if expected.Status.Id == tenant.Id {
				notFound = false
				break
			}
		}
		if notFound {
			return false
		}
	}
	return true
}

func updateUserGroupEntity(userGroup *v1.NifiUserGroup, users []*v1.NifiUser, entity *nigoapi.UserGroupEntity) {
	var defaultVersion int64 = 0

	if entity == nil {
		entity = &nigoapi.UserGroupEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.UserGroupDto{}
	}

	entity.Component.Identity = userGroup.GetIdentity()

	for _, user := range users {
		entity.Component.Users = append(entity.Component.Users, nigoapi.TenantEntity{Id: user.Status.Id})
	}
}

func userGroupContainsAccessPolicy(userGroup *v1.NifiUserGroup, entity nigoapi.AccessPolicyEntity, rootPGId string) bool {
	for _, accessPolicy := range userGroup.Spec.AccessPolicies {
		if entity.Component.Action == string(accessPolicy.Action) &&
			entity.Component.Resource == accessPolicy.GetResource(rootPGId) {
			return true
		}
	}
	return false
}

func UserGroupEntityContainsAccessPolicy(entity *nigoapi.UserGroupEntity, accessPolicy v1.AccessPolicy, rootPGId string) bool {
	for _, entity := range entity.Component.AccessPolicies {
		if entity.Component.Action == string(accessPolicy.Action) &&
			entity.Component.Resource == accessPolicy.GetResource(rootPGId) {
			return true
		}
	}
	return false
}
