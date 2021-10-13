package accesspolicies

import (
	"github.com/Orange-OpenSource/nifikop/api/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers"
	"github.com/Orange-OpenSource/nifikop/pkg/common"
	"github.com/Orange-OpenSource/nifikop/pkg/nificlient"
	"github.com/Orange-OpenSource/nifikop/pkg/util/clientconfig"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("accesspolicies-method")

func ExistAccessPolicies(accessPolicy *v1alpha1.AccessPolicy, config *clientconfig.NifiConfig) (bool, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetAccessPolicy(string(accessPolicy.Action), accessPolicy.GetResource(config.RootProcessGroupId))
	if err := clientwrappers.ErrorGetOperation(log, err, "Get access policy"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}

func CreateAccessPolicy(accessPolicy *v1alpha1.AccessPolicy, config *clientconfig.NifiConfig) (string, error) {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return "", err
	}

	scratchEntity := nigoapi.AccessPolicyEntity{}
	updateAccessPolicyEntity(
		accessPolicy,
		[]*v1alpha1.NifiUser{}, []*v1alpha1.NifiUser{},
		[]*v1alpha1.NifiUserGroup{}, []*v1alpha1.NifiUserGroup{},
		config,
		&scratchEntity)

	entity, err := nClient.CreateAccessPolicy(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Access policy user"); err != nil {
		return "", err
	}

	return entity.Id, nil
}

func UpdateAccessPolicy(
	accessPolicy *v1alpha1.AccessPolicy,
	addUsers []*v1alpha1.NifiUser,
	removeUsers []*v1alpha1.NifiUser,
	addUserGroups []*v1alpha1.NifiUserGroup,
	removeUserGroups []*v1alpha1.NifiUserGroup,
	config *clientconfig.NifiConfig) error {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	// Check if the access policy  exist
	exist, err := ExistAccessPolicies(accessPolicy, config)
	if err != nil {
		return err
	}

	if !exist {
		_, err := CreateAccessPolicy(accessPolicy, config)
		if err != nil {
			return err
		}
	}

	entity, err := nClient.GetAccessPolicy(string(accessPolicy.Action), accessPolicy.GetResource(config.RootProcessGroupId))
	if err := clientwrappers.ErrorGetOperation(log, err, "Get access policy"); err != nil {
		return err
	}

	updateAccessPolicyEntity(accessPolicy, addUsers, removeUsers, addUserGroups, removeUserGroups, config, entity)
	entity, err = nClient.UpdateAccessPolicy(*entity)
	return clientwrappers.ErrorUpdateOperation(log, err, "Update user")
}

func UpdateAccessPolicyEntity(
	entity *nigoapi.AccessPolicyEntity,
	addUsers []*v1alpha1.NifiUser,
	removeUsers []*v1alpha1.NifiUser,
	addUserGroups []*v1alpha1.NifiUserGroup,
	removeUserGroups []*v1alpha1.NifiUserGroup,
	config *clientconfig.NifiConfig) error {

	nClient, err := common.NewClusterConnection(log, config)
	if err != nil {
		return err
	}

	entity, err = nClient.GetAccessPolicy(entity.Component.Action, entity.Component.Resource)
	if err := clientwrappers.ErrorGetOperation(log, err, "Get access policy"); err != nil {
		return err
	}

	addRemoveUsersFromAccessPolicyEntity(addUsers, removeUsers, entity)
	addRemoveUserGroupsFromAccessPolicyEntity(addUserGroups, removeUserGroups, entity)

	entity, err = nClient.UpdateAccessPolicy(*entity)
	return clientwrappers.ErrorUpdateOperation(log, err, "Update user")
}

func updateAccessPolicyEntity(
	accessPolicy *v1alpha1.AccessPolicy,
	addUsers []*v1alpha1.NifiUser,
	removeUsers []*v1alpha1.NifiUser,
	addUserGroups []*v1alpha1.NifiUserGroup,
	removeUserGroups []*v1alpha1.NifiUserGroup,
	config *clientconfig.NifiConfig,
	entity *nigoapi.AccessPolicyEntity) {

	var defaultVersion int64 = 0

	if entity == nil {
		entity = &nigoapi.AccessPolicyEntity{}
	}

	if entity.Component == nil {
		entity.Revision = &nigoapi.RevisionDto{
			Version: &defaultVersion,
		}
	}

	if entity.Component == nil {
		entity.Component = &nigoapi.AccessPolicyDto{}
	}

	entity.Component.Action = string(accessPolicy.Action)
	entity.Component.Resource = accessPolicy.GetResource(config.RootProcessGroupId)

	addRemoveUsersFromAccessPolicyEntity(addUsers, removeUsers, entity)
	addRemoveUserGroupsFromAccessPolicyEntity(addUserGroups, removeUserGroups, entity)
}

func addRemoveUserGroupsFromAccessPolicyEntity(
	addUserGroups []*v1alpha1.NifiUserGroup,
	removeUserGroups []*v1alpha1.NifiUserGroup,
	entity *nigoapi.AccessPolicyEntity) {

	// Add new userGroup from the access policy
	for _, userGroup := range addUserGroups {
		entity.Component.UserGroups = append(entity.Component.UserGroups, nigoapi.TenantEntity{Id: userGroup.Status.Id})
	}

	// Remove user from the access policy
	var userGroupsAccessPolicy []nigoapi.TenantEntity
	for _, userGroup := range entity.Component.UserGroups {
		contains := false

		for _, toRemove := range removeUserGroups {
			if userGroup.Id == toRemove.Status.Id {
				contains = true
				break
			}
		}

		if !contains {
			userGroupsAccessPolicy = append(userGroupsAccessPolicy, userGroup)
		}
	}
	entity.Component.UserGroups = userGroupsAccessPolicy
}

func addRemoveUsersFromAccessPolicyEntity(
	addUsers []*v1alpha1.NifiUser,
	removeUsers []*v1alpha1.NifiUser,
	entity *nigoapi.AccessPolicyEntity) {

	// Add new user from the access policy
	for _, user := range addUsers {
		entity.Component.Users = append(entity.Component.Users, nigoapi.TenantEntity{Id: user.Status.Id})
	}

	// Remove user from the access policy
	var usersAccessPolicy []nigoapi.TenantEntity
	for _, user := range entity.Component.Users {
		contains := false

		for _, toRemove := range removeUsers {
			if user.Id == toRemove.Status.Id {
				contains = true
				break
			}
		}

		if !contains {
			usersAccessPolicy = append(usersAccessPolicy, user)
		}
	}
	entity.Component.Users = usersAccessPolicy
}
