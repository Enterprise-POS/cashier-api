package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTenantServiceImpl(t *testing.T) {
	tenantRepo := repository.NewTenantRepositoryMock(&mock.Mock{}).(*repository.TenantRepositoryMock)
	tenantService := NewTenantServiceImpl(tenantRepo)

	t.Run("GetTenantWithUser", func(t *testing.T) {
		t.Run("NormalGet", func(t *testing.T) {
		})
	})

	t.Run("NewTenant", func(t *testing.T) {
		t.Run("NormalCreate", func(t *testing.T) {
			now := time.Now()
			userId := 1
			sub := userId // typically get from JWT payload
			dummyTenant := &model.Tenant{
				Name:        "Dummy Tenant",
				OwnerUserId: userId,
				IsActive:    true,
			}

			expectedDummyTenant := &model.Tenant{
				Id:          1,
				Name:        "Dummy Tenant",
				OwnerUserId: userId,
				IsActive:    true,
				CreatedAt:   &now,
			}
			tenantRepo.Mock.On("NewTenant", dummyTenant).Return(expectedDummyTenant, nil)
			err := tenantService.NewTenant(dummyTenant, sub)
			assert.Nil(t, err)
		})

		t.Run("SubAndTenantIdNotMatch", func(t *testing.T) {
			userId := 1
			sub := 99 // typically get from JWT payload
			dummyTenant := &model.Tenant{
				Name:        "Dummy Tenant",
				OwnerUserId: userId,
				IsActive:    true,
			}

			err := tenantService.NewTenant(dummyTenant, sub)
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), "[TenantService:NewTenant:1]")
		})

		t.Run("InvalidExpectedFieldForId", func(t *testing.T) {
			userId := 1
			sub := 1
			dummyTenant := &model.Tenant{
				Id:          99, // specified id is not allowed
				Name:        "Dummy Tenant",
				OwnerUserId: userId,
				IsActive:    true,
			}

			err := tenantService.NewTenant(dummyTenant, sub)
			assert.NotNil(t, err)
		})

		t.Run("InvalidExpectedFieldForCreatedAt", func(t *testing.T) {
			now := time.Now()
			userId := 1
			sub := 1
			dummyTenant := &model.Tenant{
				Name:        "Dummy Tenant",
				OwnerUserId: userId,
				IsActive:    true,
				CreatedAt:   &now, // specified created_at is not allowed
			}

			err := tenantService.NewTenant(dummyTenant, sub)
			assert.NotNil(t, err)
		})
	})
}
