package repository

import (
	"cashier-api/model"
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TenantRepositoryImpl struct {
	Client *gorm.DB
}

const TenantTable string = "tenant"
const UserMtmTenantTable string = "user_mtm_tenant"

func NewTenantRepositoryImpl(client *gorm.DB) TenantRepository {
	return &TenantRepositoryImpl{Client: client}
}

func (repository *TenantRepositoryImpl) GetByUserId(ownerUserId int) ([]*model.Tenant, error) {
	var results = make([]*model.Tenant, 0)
	err := repository.Client.Table(TenantTable).
		Where("owner_user_id = ?", ownerUserId).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (repository *TenantRepositoryImpl) Create(tenant *model.Tenant) (*model.Tenant, error) {
	err := repository.Client.Table(TenantTable).
		Create(tenant).Error
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

func (repository *TenantRepositoryImpl) Delete(tenantId int) error {
	panic("not implemented") // TODO: Implement
}

func (repository *TenantRepositoryImpl) GetTenantWithUser(userId int) ([]*model.Tenant, error) {
	var user model.User

	err := repository.Client.
		Preload("Tenants").
		First(&user, userId).Error
	if err != nil {
		return nil, err
	}

	results := make([]*model.Tenant, len(user.Tenants))
	for i := range user.Tenants {
		results[i] = &user.Tenants[i]
	}

	return results, nil
}

/*
Fresh new tenant, with current user as a owner
only create 1 tenant, will call new_tenant_user_as_owner function
? when fresh new tenant created, automatically also insert into new_tenant_user_as_owner table
*/
func (repository *TenantRepositoryImpl) NewTenant(tenant *model.Tenant) error {
	var response string
	err := repository.Client.
		Raw("SELECT new_tenant_user_as_owner(?, ?)", tenant.OwnerUserId, tenant.Name).
		Scan(&response).Error
	if err != nil {
		return err
	}

	if strings.Contains(response, "[ERROR]") {
		return errors.New(response)
	}
	if response == "" {
		return errors.New("[ERROR] Fatal error response return nothing")
	}

	return nil
}

/*
Will create new data into user_mtm_tenant table
that will make user have many to many relation with tenant table
*/
func (repository *TenantRepositoryImpl) AddUserToTenant(userId, tenantId int) (*model.UserMtmTenant, error) {
	newUserMtmTenant := &model.UserMtmTenant{UserId: userId, TenantId: tenantId}
	err := repository.Client.Table(UserMtmTenantTable).
		Create(newUserMtmTenant).Error
	if err != nil {
		return nil, err
	}
	return newUserMtmTenant, nil
}

func (repository *TenantRepositoryImpl) RemoveUserFromTenant(userMtmTenant *model.UserMtmTenant, userId int) (string, error) {
	var response string
	err := repository.Client.
		Raw("SELECT remove_user_from_tenant(?, ?, ?)", userId, userMtmTenant.UserId, userMtmTenant.TenantId).
		Scan(&response).Error
	if err != nil {
		return "", err
	}

	if strings.Contains(response, "[ERROR] ") {
		return "", errors.New(response)
	}

	switch {
	case strings.Contains(response, "Current tenant will be archived"),
		strings.Contains(response, "Removed from tenant"):
		log.Infof("%s, tenantId: %d, userId: %d", response, userMtmTenant.TenantId, userMtmTenant.UserId)
	default:
		log.Warnf("Unknown success response executed; tenantId: %d, userId: %d", userMtmTenant.TenantId, userMtmTenant.UserId)
		log.Warnf("Response as %s", response)
	}

	return response, nil
}

// GetTenantMembers implements TenantRepository.
func (repository *TenantRepositoryImpl) GetTenantMembers(tenantId int) ([]*model.User, error) {
	var tenant model.Tenant

	err := repository.Client.
		Preload("Users").
		First(&tenant, tenantId).Error
	if err != nil {
		return nil, err
	}

	results := make([]*model.User, len(tenant.Users))
	for i := range tenant.Users {
		results[i] = &tenant.Users[i]
	}

	return results, nil
}
