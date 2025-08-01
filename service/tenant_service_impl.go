package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"

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
		log.Warnf("Forbidden action detected ! userId: %d, sub: %d", userId, sub)
		return nil, errors.New("[TenantService:GetTenantWithUser:1]")
	}

	return service.Repository.GetTenantWithUser(userId)
}

// NewTenant implements TenantService.
func (service *TenantServiceImpl) NewTenant(tenant *model.Tenant, sub int) error {
	if tenant.OwnerUserId != sub {
		log.Warnf("Forbidden action detected ! userId: %d, sub: %d", tenant.OwnerUserId, sub)
		return errors.New("[TenantService:NewTenant:1]")
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
