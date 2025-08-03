package repository

import "cashier-api/model"

type TenantRepository interface {
	/*
		Return many by 'user id'
	*/
	GetByUserId(userId int) ([]*model.Tenant, error)

	/*
		WARN: under mandatory, maybe not used for production
		Only create 1 tenant
	*/
	Create(tenant *model.Tenant) (*model.Tenant, error)

	/*
		Delete 1 tenant
	*/
	Delete(tenantId int) error

	/*
		Return all tenants from 1 user tenant
		- will call user_mtm_tenant
	*/
	GetTenantWithUser(userId int) ([]*model.Tenant, error)

	/*
		Fresh new tenant, with current user as a owner
		only create 1 tenant, will call new_tenant_user_as_owner function
	*/
	NewTenant(tenant *model.Tenant) error

	/*
		Will create new data into user_mtm_tenant table
		that will make user have many to many relation with tenant table
	*/
	AddUserToTenant(userId, tenantId int) (*model.UserMtmTenant, error)

	/*
		Remove user from tenant
		- delete from user_mtm_tenant
	*/
	RemoveUserFromTenant(userMtmTenantId *model.UserMtmTenant, userId int) (string, error)
}
