package repository

import (
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type TenantRepositoryMock struct {
	Mock *mock.Mock
}

func NewTenantRepositoryMock(mock *mock.Mock) TenantRepository {
	return &TenantRepositoryMock{
		Mock: mock,
	}
}

// AddUserToTenant implements TenantRepository.
func (t *TenantRepositoryMock) AddUserToTenant(userId int, tenantId int) (*model.UserMtmTenant, error) {
	panic("unimplemented")
}

// Create implements TenantRepository.
func (t *TenantRepositoryMock) Create(tenant *model.Tenant) (*model.Tenant, error) {
	panic("unimplemented")
}

// Delete implements TenantRepository.
func (t *TenantRepositoryMock) Delete(tenantId int) error {
	panic("unimplemented")
}

// GetByUserId implements TenantRepository.
func (t *TenantRepositoryMock) GetByUserId(userId int) ([]*model.Tenant, error) {
	panic("unimplemented")
}

// GetTenantWithUser implements TenantRepository.
func (repository *TenantRepositoryMock) GetTenantWithUser(userId int) ([]*model.Tenant, error) {
	args := repository.Mock.Called(userId)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	// Normal condition
	return args.Get(0).([]*model.Tenant), nil
}

// NewTenant implements TenantRepository.
func (repository *TenantRepositoryMock) NewTenant(tenant *model.Tenant) error {
	args := repository.Mock.Called(tenant)

	if args.Get(0) != nil {
		return args.Error(1)
	}

	// Normal condition
	return nil
}

// RemoveUserFromTenant implements TenantRepository.
func (repository *TenantRepositoryMock) RemoveUserFromTenant(userMtmTenantId *model.UserMtmTenant, userId int) (string, error) {
	args := repository.Mock.Called(userMtmTenantId, userId)

	if args.String(0) == "" {
		return "", args.Error(1)
	}

	return args.String(0), nil
}
