package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTenantServiceImpl(t *testing.T) {
	tenantRepo := repository.NewTenantRepositoryMock(&mock.Mock{}).(*repository.TenantRepositoryMock)
	tenantService := NewTenantServiceImpl(tenantRepo)

	t.Run("GetTenantWithUser", func(t *testing.T) {
		t.Run("NormalGet", func(t *testing.T) {
			userId := 1
			sub := 1
			now := time.Now()

			expectedTenants := []*model.Tenant{
				{
					Id:          1,
					Name:        "Dummy Tenant",
					OwnerUserId: userId,
					IsActive:    true,
					CreatedAt:   &now,
				},
				{
					Id:          2,
					Name:        "Dummy Tenant",
					OwnerUserId: 3, // other user id
					IsActive:    true,
					CreatedAt:   &now,
				},
			}
			tenantRepo.Mock.On("GetTenantWithUser", userId).Return(expectedTenants, nil)
			tenants, err := tenantService.GetTenantWithUser(userId, sub)
			assert.Nil(t, err)
			assert.NotNil(t, tenants)
			assert.Equal(t, 2, len(tenants))
		})

		t.Run("SubAndTenantIdNotMatch", func(t *testing.T) {
			userId := 1
			sub := 2

			tenants, err := tenantService.GetTenantWithUser(userId, sub)
			assert.NotNil(t, err)
			assert.Nil(t, tenants)
			assert.Equal(t, "[TenantService:GetTenantWithUser:1]", err.Error())
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

	t.Run("AddUserToTenant", func(t *testing.T) {
		t.Run("NormalAdd", func(t *testing.T) {
			toBeAddedUserId := 99
			tenantId := 1
			performerId := 1 // user who request to add some user
			sub := performerId

			now := time.Now()
			expectedAddedUserFromTenant := &model.UserMtmTenant{
				Id:        1,
				UserId:    toBeAddedUserId,
				TenantId:  tenantId,
				CreatedAt: &now,
			}
			tenantRepo.Mock.On("AddUserToTenant", toBeAddedUserId, tenantId).Return(expectedAddedUserFromTenant, nil)
			testUserMtmTenant, err := tenantService.AddUserToTenant(toBeAddedUserId, tenantId, performerId, sub)
			assert.NoError(t, err)
			assert.NotNil(t, testUserMtmTenant)
			assert.Equal(t, expectedAddedUserFromTenant.TenantId, testUserMtmTenant.TenantId)
			assert.Equal(t, expectedAddedUserFromTenant.UserId, testUserMtmTenant.UserId)
			assert.Equal(t, expectedAddedUserFromTenant.CreatedAt, testUserMtmTenant.CreatedAt)
		})

		t.Run("ForbiddenActionForAddingUserWithOtherUserId", func(t *testing.T) {
			toBeAddedUserId := 99
			tenantId := 1
			performerId := 1
			sub := 2 // JWT id different
			tenantRepo.Mock.On("AddUserToTenant", toBeAddedUserId, tenantId).Return(nil, errors.New("[TenantService:AddUserToTenant]"))
			testUserMtmTenant, err := tenantService.AddUserToTenant(toBeAddedUserId, tenantId, performerId, sub)
			assert.Error(t, err)
			assert.Nil(t, testUserMtmTenant)
			assert.Equal(t, "[TenantService:AddUserToTenant]", err.Error())
		})
	})

	t.Run("RemoveUserFromTenant", func(t *testing.T) {
		t.Run("NormalRemove", func(t *testing.T) {
			now := time.Now()
			userId := 1
			sub := userId // typically get from JWT payload
			userMtmTenant := &model.UserMtmTenant{
				Id:        1,
				UserId:    1,
				TenantId:  1,
				CreatedAt: &now,
			}
			tenantRepo.Mock.On("RemoveUserFromTenant", userMtmTenant, sub).Return("[SUCCESS] Removed from tenant", nil)
			response, err := tenantService.RemoveUserFromTenant(userMtmTenant, userId, sub)
			assert.NoError(t, err)
			assert.Equal(t, "[SUCCESS] Removed from tenant", response)
		})

		t.Run("ForbiddenActionRemovingNotByPerformer", func(t *testing.T) {
			now := time.Now()
			userId := 1
			sub := userId // typically get from JWT payload
			userMtmTenant := &model.UserMtmTenant{
				Id:        1,
				UserId:    1,
				CreatedAt: &now,
			}

			wrongPerformerId := 99
			response, err := tenantService.RemoveUserFromTenant(userMtmTenant, wrongPerformerId, sub)
			assert.Error(t, err)
			assert.Equal(t, "", response)
			assert.Equal(t, "[TenantService:RemoveUserFromTenant]", err.Error())
		})

		t.Run("WrongDataTypeForSpecifyingTenantId", func(t *testing.T) {
			userId := 1
			sub := userId // typically get from JWT payload
			userMtmTenant := &model.UserMtmTenant{
				Id:       1,
				TenantId: 1,
				UserId:   0, // When not specifying UserId, it should be looked like 0
			}

			response, err := tenantService.RemoveUserFromTenant(userMtmTenant, userId, sub)
			assert.Error(t, err)
			assert.Equal(t, "", response)
			assert.Contains(t, err.Error(), "Data type error. User Id should be inserted.")
		})
	})

	t.Run("GetTenantMembers", func(t *testing.T) {
		t.Run("NormalGet", func(t *testing.T) {
			now := time.Now()
			tenantId := 1
			currentRequestedUserId := 999
			tenantMembers := []*model.User{
				{
					Id:        1,
					Name:      "Test 1",
					Email:     "test@gmail.com",
					CreatedAt: &now,
				},
				{
					Id:        currentRequestedUserId,
					Name:      "Test 2",
					Email:     "test2@gmail.com",
					CreatedAt: &now,
				},
			}
			tenantRepo.Mock.On("GetTenantMembers", tenantId).Return(tenantMembers, nil)
			users, err := tenantService.GetTenantMembers(tenantId, currentRequestedUserId)
			require.NoError(t, err)
			require.Equal(t, 2, len(users))
		})

		t.Run("ForbiddenActionForRequestingTenantDataWhileNotRegistered", func(t *testing.T) {
			now := time.Now()
			tenantId := 1
			currentRequestedUserId := 999
			tenantMembers := []*model.User{
				{
					Id:        1,
					Name:      "Test 1",
					Email:     "test@gmail.com",
					CreatedAt: &now,
				},
				{
					Id:        2,
					Name:      "Test 2",
					Email:     "test2@gmail.com",
					CreatedAt: &now,
				},
			}

			// Here mock is required, we must restart the mock otherwise, the mock will be overlap
			tenantRepo.Mock = &mock.Mock{}
			tenantRepo.Mock.On("GetTenantMembers", tenantId).Return(tenantMembers, nil)
			users, err := tenantService.GetTenantMembers(tenantId, currentRequestedUserId)
			require.Error(t, err)
			require.Nil(t, users)
		})
	})
}
