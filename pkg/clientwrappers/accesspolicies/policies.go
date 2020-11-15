package accesspolicies

import (
	"github.com/Orange-OpenSource/nifikop/pkg/apis/nifi/v1alpha1"
	"github.com/Orange-OpenSource/nifikop/pkg/clientwrappers"
	"github.com/Orange-OpenSource/nifikop/pkg/controller/common"
	"github.com/Orange-OpenSource/nifikop/pkg/nificlient"
	nigoapi "github.com/erdrix/nigoapi/pkg/nifi"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("accesspolicies-method")

func ExistAccessPolicies(client client.Client, accessPolicy *v1alpha1.AccessPolicy,
	cluster *v1alpha1.NifiCluster) (bool, error) {

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return false, err
	}

	entity, err := nClient.GetAccessPolicy(string(accessPolicy.Action), accessPolicy.GetResource(cluster))
	if err := clientwrappers.ErrorGetOperation(log, err, "Get access policy"); err != nil {
		if err == nificlient.ErrNifiClusterReturned404 {
			return false, nil
		}
		return false, err
	}

	return entity != nil, nil
}


func CreateAccessPolicy(client client.Client, accessPolicy *v1alpha1.AccessPolicy,
	cluster *v1alpha1.NifiCluster) (string, error) {

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return "", err
	}

	scratchEntity := nigoapi.AccessPolicyEntity{}
	updateAccessPolicyEntity(
		accessPolicy,
		[]*v1alpha1.NifiUser{}, []*v1alpha1.NifiUser{},
		[]*v1alpha1.NifiUserGroup{}, []*v1alpha1.NifiUserGroup{},
		cluster,
		&scratchEntity)

	entity, err := nClient.CreateAccessPolicy(scratchEntity)
	if err := clientwrappers.ErrorCreateOperation(log, err, "Access policy user"); err != nil {
		return "", err
	}

	return entity.Id, nil
}

func UpdateAccessPolicy(
	client client.Client,
	accessPolicy *v1alpha1.AccessPolicy,
	addUsers []*v1alpha1.NifiUser,
	removeUsers []*v1alpha1.NifiUser,
	addUserGroups []*v1alpha1.NifiUserGroup,
	removeUserGroups []*v1alpha1.NifiUserGroup,
	cluster *v1alpha1.NifiCluster) error {

	nClient, err := common.NewNodeConnection(log, client, cluster)
	if err != nil {
		return err
	}

	// Check if the access policy  exist
	exist, err := ExistAccessPolicies(client, accessPolicy, cluster)
	if err != nil {
		return err
	}

	if !exist {
		_, err := CreateAccessPolicy(client, accessPolicy, cluster)
		if err != nil {
			return err
		}
	}

	entity, err := nClient.GetAccessPolicy(string(accessPolicy.Action), accessPolicy.GetResource(cluster))
	if err := clientwrappers.ErrorGetOperation(log, err, "Get access policy"); err != nil {
		return err
	}

	updateAccessPolicyEntity(accessPolicy, addUsers, removeUsers, addUserGroups, removeUserGroups, cluster, entity)
	entity, err = nClient.UpdateAccessPolicy(*entity)
	return clientwrappers.ErrorUpdateOperation(log, err, "Update user")
}


func UpdateAccessPolicyEntity(
	client client.Client,
	entity *nigoapi.AccessPolicyEntity,
	addUsers []*v1alpha1.NifiUser,
	removeUsers []*v1alpha1.NifiUser,
	addUserGroups []*v1alpha1.NifiUserGroup,
	removeUserGroups []*v1alpha1.NifiUserGroup,
	cluster *v1alpha1.NifiCluster) error {

	nClient, err := common.NewNodeConnection(log, client, cluster)
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
	cluster *v1alpha1.NifiCluster,
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
		entity.Component = &nigoapi.AccessPolicyDto{
		}
	}

	entity.Component.Action = string(accessPolicy.Action)
	entity.Component.Resource = accessPolicy.GetResource(cluster)

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
