package repository

import (
	"cashier-api/model"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/supabase-community/supabase-go"
)

type TenantRepositoryImpl struct {
	Client *supabase.Client
}

const TenantTable string = "tenant"
const UserMtmTenantTable string = "user_mtm_tenant"

func NewTenantRepositoryImpl(client *supabase.Client) *TenantRepositoryImpl {
	return &TenantRepositoryImpl{Client: client}
}

func (repository *TenantRepositoryImpl) GetByUserId(ownerUserId int) ([]*model.Tenant, error) {
	// For now we don't limit how many should return
	var results []*model.Tenant
	_, err := repository.Client.From(TenantTable).
		Select("*", "", false).
		Eq("owner_user_id", strconv.Itoa(ownerUserId)).
		ExecuteTo(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (repository *TenantRepositoryImpl) Create(tenant *model.Tenant) (*model.Tenant, error) {
	var newTenant *model.Tenant
	_, err := repository.Client.From(TenantTable).
		Insert(tenant, false, "", "", "").
		Single().
		ExecuteTo(&newTenant)
	if err != nil {
		return nil, err
	}

	return newTenant, nil
}

func (repository *TenantRepositoryImpl) Delete(tenantId int) (_ error) {
	panic("not implemented") // TODO: Implement
}

/*
Return all tenants from 1 user tenant
- will call user_mtm_tenant
*/
func (repository *TenantRepositoryImpl) GetTenantWithUser(userId int) ([]*model.Tenant, error) {
	data := repository.Client.Rpc("get_tenant_base_on_user_id", "", map[string]int{
		"p_user_id": userId,
	})

	var results []*model.Tenant
	err := json.Unmarshal([]byte(data), &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

/*
Fresh new tenant, with current user as a owner
only create 1 tenant, will call new_tenant_user_as_owner function
? by default field Tenant.category is TRUE, so don't need to update the file for now
*/
func (repository *TenantRepositoryImpl) NewTenant(tenant *model.Tenant) error {
	var response string = repository.Client.Rpc("new_tenant_user_as_owner", "", map[string]interface{}{
		"p_user_id":     tenant.OwnerUserId,
		"p_tenant_name": tenant.Name,
	})

	// space after [ERROR] are required
	if strings.HasPrefix(response, "[ERROR] ") {
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
	var newUserMtmTenant *model.UserMtmTenant
	_, err := repository.Client.From(UserMtmTenantTable).
		Insert(&model.UserMtmTenant{UserId: userId, TenantId: tenantId}, false, "", "", "").
		Single().
		ExecuteTo(&newUserMtmTenant)
	if err != nil {
		return nil, err
	}

	return newUserMtmTenant, nil
}

func (repository *TenantRepositoryImpl) RemoveUserFromTenant(userMtmTenant *model.UserMtmTenant) (string, error) {
	response := repository.Client.Rpc("remove_user_from_tenant", "", map[string]any{
		"p_user_id":   userMtmTenant.UserId,
		"p_tenant_id": userMtmTenant.TenantId,
	})

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
