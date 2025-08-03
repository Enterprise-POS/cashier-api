package service

import "cashier-api/model"

type TenantService interface {
	/*
		Return all users tenants, nor owner nor a staff
	*/
	GetTenantWithUser(userId int, sub int) ([]*model.Tenant, error)

	/*
		Fresh new tenant, with current user as a owner
		only create 1 tenant, will call new_tenant_user_as_owner function
	*/
	NewTenant(tenant *model.Tenant, sub int) error

	/*
		Remove user from tenant
		- delete from user_mtm_tenant
	*/
	RemoveUserFromTenant(userMtmTenantId *model.UserMtmTenant, performerId, sub int) (string, error)
}
