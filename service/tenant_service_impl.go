package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
)

type TenantServiceImpl struct {
	Repository repository.TenantRepository
}

func NewTenantServiceImpl(repository repository.TenantRepository) TenantService {
	return &TenantServiceImpl{
		Repository: repository,
	}
}

// GetTenantWithUser implements TenantService.
func (service *TenantServiceImpl) GetTenantWithUser(userId int, sub int) ([]*model.Tenant, error) {
	if userId != sub {
		log.Warnf("Forbidden action detected ! userId: %d, sub: %d; Performing GetTenantWithUser", userId, sub)
		return nil, errors.New("[TenantService:GetTenantWithUser:1]")
	}

	return service.Repository.GetTenantWithUser(userId)
}

// NewTenant implements TenantService.
func (service *TenantServiceImpl) NewTenant(tenant *model.Tenant, sub int) error {
	if tenant.OwnerUserId != sub {
		log.Warnf("Forbidden action detected ! userId: %d, sub: %d; Performing NewTenant", tenant.OwnerUserId, sub)
		return errors.New("[TenantService:NewTenant:1]")
	}

	var tenantNameValidator = regexp.MustCompile(`.*[^ ].*`)
	if !tenantNameValidator.MatchString(tenant.Name) {
		return errors.New("Tenant name is not allowed. Please try another.")
	}

	if tenant.Id != 0 {
		log.Errorf("Data type error. tenant Id should not be inserted. Specified tenant id: %d", tenant.Id)
		return fmt.Errorf("Data type error. tenant Id should not be inserted. Specified tenant id: %d", tenant.Id)
	}
	if tenant.CreatedAt != nil {
		log.Errorf("Data type error. tenant created_at should not be inserted. Specified tenant created at: %s", tenant.CreatedAt.String())
		return fmt.Errorf("Data type error. tenant  should not be inserted. Specified tenant : %s", tenant.CreatedAt.String())
	}

	return service.Repository.NewTenant(tenant)
}

// RemoveUserFromTenant implements TenantService.
func (service *TenantServiceImpl) RemoveUserFromTenant(userMtmTenant *model.UserMtmTenant, performerId, sub int) (string, error) {
	if performerId != sub {
		log.Warnf("Forbidden action detected ! userId: %d, sub: %d; Performing RemoveUserFromTenant", userMtmTenant.UserId, sub)
		return "", errors.New("[TenantService:RemoveUserFromTenant]")
	}

	if userMtmTenant.UserId == 0 {
		log.Errorf("Data type error. User Id should be inserted. Specified sub id: %d", userMtmTenant.Id)
		return "", fmt.Errorf("Data type error. User Id should be inserted. Specified sub id: %d", userMtmTenant.Id)
	}

	if userMtmTenant.TenantId == 0 {
		log.Errorf("Data type error. Tenant Id should be inserted. Specified sub id: %d", userMtmTenant.Id)
		return "", fmt.Errorf("Data type error. Tenant Id should be inserted. Specified sub id: %d", userMtmTenant.Id)
	}

	// Extra security if user actually the owner of the tenant
	// performerId is an identity whether the requester is a owner or not
	// currently only owner could remove normal user from tenant
	// FYI: user_mtm_tenant doesn't store owner user id
	return service.Repository.RemoveUserFromTenant(userMtmTenant, performerId)
}

// Future admin, staff, owner role should be implemented.
// AddUserToTenant implements TenantService.
func (service *TenantServiceImpl) AddUserToTenant(userId, tenantId, performerId, sub int) (*model.UserMtmTenant, error) {
	if performerId != sub {
		log.Warnf("Forbidden action detected ! userId: %d, sub: %d; Performing RemoveUserFromTenant", performerId, sub)
		return nil, errors.New("[TenantService:AddUserToTenant]")
	}

	return service.Repository.AddUserToTenant(userId, tenantId)
}

// GetTenantMembers implements TenantService.
func (service *TenantServiceImpl) GetTenantMembers(tenantId int, sub int) ([]*model.User, error) {
	users, err := service.Repository.GetTenantMembers(tenantId)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Id == sub {
			return users, nil
		}
	}

	return nil, errors.New("[TenantService:GetTenantMembers]")
}
