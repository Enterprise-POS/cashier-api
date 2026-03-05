package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantRepositoryImpl(t *testing.T) {
	gormClient := client.CreateGormClient()

	// This is test user id; Do not delete any accidentally
	// delete may cause another test case throw panic
	const UserId = 1

	t.Run("GetByUserId", func(t *testing.T) {
		tenantRepo := NewTenantRepositoryImpl(gormClient)

		t.Run("NormalGetByUserId", func(t *testing.T) {
			// Create the tenant first
			dummyTenant := &model.Tenant{
				Name:        "Test_TenantRepositoryImpl/GetByUserId/NormalGetByUserId 1 Group_" + uuid.NewString(),
				OwnerUserId: UserId,
				IsActive:    true,
			}
			newDummyTenant, err := tenantRepo.Create(dummyTenant)
			require.Nil(t, err)

			dummyTenants, err := tenantRepo.GetByUserId(UserId)
			assert.Nil(t, err)
			assert.NotNil(t, dummyTenant)
			assert.GreaterOrEqual(t, len(dummyTenants), 1)

			// Clean up
			err = gormClient.Delete(&newDummyTenant).Error
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/Create/NormalCreate")
		})
	})

	t.Run("Create", func(t *testing.T) {
		tenantRepo := NewTenantRepositoryImpl(gormClient)

		t.Run("NormalCreate", func(t *testing.T) {
			// Create the tenant first
			dummyTenant := &model.Tenant{
				Name:        "Test_TenantRepositoryImpl/Create/NormalCreate 1 Group_" + uuid.NewString(),
				OwnerUserId: UserId,
				IsActive:    true,
			}
			newDummyTenant, err := tenantRepo.Create(dummyTenant)
			assert.Nil(t, err)
			assert.NotNil(t, newDummyTenant)
			assert.Equal(t, dummyTenant.Name, newDummyTenant.Name)
			assert.Equal(t, dummyTenant.OwnerUserId, newDummyTenant.OwnerUserId)

			// Clean up
			err = gormClient.Delete(&newDummyTenant).Error
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/Create/NormalCreate")
		})
	})

	t.Run("GetTenantWithUser", func(t *testing.T) {
		tenantRepo := NewTenantRepositoryImpl(gormClient)

		t.Run("NormalGet", func(t *testing.T) {
			dummyTenant := &model.Tenant{
				Name:        "Test_TenantRepositoryImpl/GetTenantWithUser/NormalGet 1 Group_" + uuid.NewString(),
				OwnerUserId: UserId,
				IsActive:    true,
			}

			newDummyTenant, err := tenantRepo.Create(dummyTenant)
			require.Nil(t, err)

			// Insert into user_mtm_tenant using GORM
			err = gormClient.Create(&model.UserMtmTenant{
				UserId:   UserId,
				TenantId: newDummyTenant.Id,
			}).Error
			require.Nil(t, err)

			createdDummyTenants, err := tenantRepo.GetTenantWithUser(UserId)
			assert.Nil(t, err)
			assert.NotNil(t, createdDummyTenants)
			assert.GreaterOrEqual(t, len(createdDummyTenants), 1)

			isAvailable := false
			for _, tenant := range createdDummyTenants {
				if tenant.Name == dummyTenant.Name {
					isAvailable = true
					break
				}
			}
			assert.True(t, isAvailable)

			// Clean up user_mtm_tenant
			err = gormClient.
				Where("user_id = ? AND tenant_id = ?", UserId, newDummyTenant.Id).
				Delete(&model.UserMtmTenant{}).Error
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/GetTenantWithUser/NormalGet 1")

			// Clean up tenant
			err = gormClient.Delete(&model.Tenant{}, newDummyTenant.Id).Error
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/GetTenantWithUser/NormalGet 2")
		})
	})

	t.Run("NewTenant", func(t *testing.T) {
		tenantRepo := NewTenantRepositoryImpl(gormClient)

		t.Run("NormalInput", func(t *testing.T) {
			dummyTenant := &model.Tenant{
				Name:        "Test_TenantRepositoryImpl/NewTenant/NormalInput 1 Group_" + uuid.NewString(),
				OwnerUserId: UserId,
				IsActive:    true,
			}
			err := tenantRepo.NewTenant(dummyTenant)
			assert.Nil(t, err)

			// check
			createdDummyTenants, err := tenantRepo.GetTenantWithUser(UserId)
			assert.Nil(t, err)
			assert.NotNil(t, createdDummyTenants)
			assert.GreaterOrEqual(t, len(createdDummyTenants), 1)

			isAvailable := false
			createdTenantId := 0
			for _, tenant := range createdDummyTenants {
				if tenant.Name == dummyTenant.Name {
					isAvailable = true
					createdTenantId = tenant.Id
					break
				}
			}
			require.True(t, isAvailable)
			require.NotEqual(t, 0, createdTenantId)

			// Clean up
			err = gormClient.
				Where("user_id", UserId).
				Where("tenant_id", createdTenantId).
				Delete(&model.UserMtmTenant{}).Error
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/NewTenant/NormalInput 1")
			err = gormClient.
				Where("id", createdTenantId).
				Delete(&model.Tenant{}).Error
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/NewTenant/NormalInput 2")
		})
	})

	t.Run("AddUserToTenant", func(t *testing.T) {
		userRepo := NewUserRepositoryImpl(gormClient)
		tenantRepo := NewTenantRepositoryImpl(gormClient)

		// create user
		dummyUser := model.User{
			Name:  "Test_TenantRepositoryImpl/AddUserToTenant user 1" + string(uuid.NewString()),
			Email: "testtenantrepositoryimpl" + uuid.NewString() + "@gmail.com",
		}
		password := "12345678"
		newCreatedDummyUser, err := userRepo.CreateWithEmailAndPassword(dummyUser, password)
		require.Nil(t, err)
		require.NotNil(t, newCreatedDummyUser)

		// Create the tenant first
		dummyTenant := &model.Tenant{
			Name:        "Test_TenantRepositoryImpl/AddUserToTenant 1 Group_" + uuid.NewString(),
			OwnerUserId: newCreatedDummyUser.Id,
			IsActive:    true,
		}
		err = tenantRepo.NewTenant(dummyTenant)
		require.Nil(t, err)

		// check
		createdDummyTenants, err := tenantRepo.GetTenantWithUser(newCreatedDummyUser.Id)
		assert.Nil(t, err)
		assert.NotNil(t, createdDummyTenants)
		assert.GreaterOrEqual(t, len(createdDummyTenants), 1)

		isAvailable := false
		createdTenantId := 0
		for _, tenant := range createdDummyTenants {
			if tenant.Name == dummyTenant.Name {
				isAvailable = true
				createdTenantId = tenant.Id
				break
			}
		}
		require.True(t, isAvailable)
		require.NotEqual(t, 0, createdTenantId)

		// ===============================
		// Tenant and user already created
		// -	newCreatedDummyUser *model.User
		// -	createdTenantId int

		t.Run("NormalAdd", func(t *testing.T) {
			// create user
			dummyUser2 := model.User{
				Name:  "Test_TenantRepositoryImpl/AddUserToTenant/NormalAdd user 2" + string(uuid.NewString()),
				Email: "testtenantrepositoryimpl" + uuid.NewString() + "@gmail.com",
			}
			password = "12345678"
			newCreatedDummyUser2, err := userRepo.CreateWithEmailAndPassword(dummyUser2, password)
			require.Nil(t, err)
			require.NotNil(t, newCreatedDummyUser2)

			data, err := tenantRepo.AddUserToTenant(newCreatedDummyUser2.Id, createdTenantId)
			assert.Nil(t, err)
			assert.NotNil(t, data)
			assert.Equal(t, createdTenantId, data.TenantId)

			// Check manually if current tenant actually added
			var count int64
			err = gormClient.
				Model(&model.UserMtmTenant{}).
				Where("tenant_id = ?", createdTenantId).
				Count(&count).Error
			assert.Nil(t, err)
			assert.Equal(t, int64(2), count)

			// Clean up
			err = gormClient.
				Where("user_id = ? AND tenant_id = ?", newCreatedDummyUser2.Id, createdTenantId).
				Delete(&model.UserMtmTenant{}).Error
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/NewTenant/NormalInput 1")

			err = gormClient.
				Delete(&model.User{}, newCreatedDummyUser2.Id).Error
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/NewTenant/NormalInput 2")
		})

		t.Run("UserIdNotAvailable", func(t *testing.T) {
			notAvailableUserId := 0
			data, err := tenantRepo.AddUserToTenant(notAvailableUserId, createdTenantId)
			assert.Nil(t, data)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "23503")
		})

		t.Run("TenantIdNotAvailable", func(t *testing.T) {
			notAvailableTenantId := 0
			data, err := tenantRepo.AddUserToTenant(newCreatedDummyUser.Id, notAvailableTenantId)
			assert.Nil(t, data)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "23503")
		})

		// Clean up helper dummy
		err = gormClient.
			Where("user_id = ? AND tenant_id = ?", newCreatedDummyUser.Id, createdTenantId).
			Delete(&model.UserMtmTenant{}).Error
		require.Nil(t, err)

		err = gormClient.Delete(&model.Tenant{}, createdTenantId).Error
		require.Nil(t, err)

		err = gormClient.Delete(&newCreatedDummyUser).Error
		require.Nil(t, err)
	})

	t.Run("RemoveUserFromTenant", func(t *testing.T) {
		userRepo := NewUserRepositoryImpl(gormClient)
		tenantRepo := NewTenantRepositoryImpl(gormClient)

		// create user
		dummyUser := model.User{
			Name:  "Test_TenantRepositoryImpl/RemoveUserFromTenant user 1" + string(uuid.NewString()),
			Email: "testtenantrepositoryimpl" + uuid.NewString() + "@gmail.com",
		}
		password := "12345678"
		newCreatedDummyUser, err := userRepo.CreateWithEmailAndPassword(dummyUser, password)
		require.Nil(t, err)
		require.NotNil(t, newCreatedDummyUser)

		// Create the tenant first
		dummyTenant := &model.Tenant{
			Name:        "Test_TenantRepositoryImpl/RemoveUserFromTenant 1 Group_" + uuid.NewString(),
			OwnerUserId: newCreatedDummyUser.Id,
			IsActive:    true,
		}
		err = tenantRepo.NewTenant(dummyTenant)
		require.Nil(t, err)

		// check
		createdDummyTenants, err := tenantRepo.GetTenantWithUser(newCreatedDummyUser.Id)
		assert.Nil(t, err)
		assert.NotNil(t, createdDummyTenants)
		assert.GreaterOrEqual(t, len(createdDummyTenants), 1)

		isAvailable := false
		createdTenantId := 0
		for _, tenant := range createdDummyTenants {
			if tenant.Name == dummyTenant.Name {
				isAvailable = true
				createdTenantId = tenant.Id
				break
			}
		}
		require.True(t, isAvailable)
		require.NotEqual(t, 0, createdTenantId)

		// ===============================
		// Tenant and user already created
		// -	newCreatedDummyUser *model.User
		// -	createdTenantId int

		t.Run("UserIdNotAvailable", func(t *testing.T) {
			response, err := tenantRepo.RemoveUserFromTenant(&model.UserMtmTenant{UserId: 0, TenantId: createdTenantId}, newCreatedDummyUser.Id)
			assert.NotNil(t, err)
			assert.Equal(t, "", response)
			assert.Contains(t, err.Error(), "[ERROR] Fatal error, user id not existed")
		})

		t.Run("TenantIdNotAvailable", func(t *testing.T) {
			response, err := tenantRepo.RemoveUserFromTenant(&model.UserMtmTenant{UserId: newCreatedDummyUser.Id, TenantId: 0}, newCreatedDummyUser.Id)
			assert.NotNil(t, err)
			assert.Equal(t, "", response)
			assert.Contains(t, err.Error(), "[ERROR] Fatal error, tenant id not existed")
		})

		t.Run("IllegalActionRemoveUserFromNonOwner", func(t *testing.T) {
			// create user
			dummyUser2 := model.User{
				Name:  "Test_TenantRepositoryImpl/RemoveUserFromTenant/NormalRemove user 2" + string(uuid.NewString()),
				Email: "testtenantrepositoryimpl" + uuid.NewString() + "@gmail.com",
			}
			password = "12345678"
			newCreatedDummyUser2, err := userRepo.CreateWithEmailAndPassword(dummyUser2, password)
			require.Nil(t, err)
			require.NotNil(t, newCreatedDummyUser2)

			// Begin test
			// Here we insert performerId is newCreatedDummyUser 2, which is not the owner
			// - Tenant will be there
			// - User available
			// // User delete not from the owner as well
			response, err := tenantRepo.RemoveUserFromTenant(&model.UserMtmTenant{UserId: newCreatedDummyUser2.Id, TenantId: createdTenantId}, newCreatedDummyUser2.Id)
			assert.Error(t, err)
			assert.Equal(t, "", response)
			assert.Equal(t, "[ERROR] Illegal action! Removing user only allowed by the owner", err.Error())

			// Clean up
			err = gormClient.Delete(&newCreatedDummyUser2).Error
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/RemoveUserFromTenant/IllegalActionRemoveUserFromNonOwners 1")
		})

		t.Run("NormalRemove", func(t *testing.T) {
			// create user
			dummyUser2 := model.User{
				Name:  "Test_TenantRepositoryImpl/RemoveUserFromTenant/NormalRemove user 2" + string(uuid.NewString()),
				Email: "testtenantrepositoryimpl" + uuid.NewString() + "@gmail.com",
			}
			password = "12345678"
			newCreatedDummyUser2, err := userRepo.CreateWithEmailAndPassword(dummyUser2, password)
			require.Nil(t, err)
			require.NotNil(t, newCreatedDummyUser2)

			data, err := tenantRepo.AddUserToTenant(newCreatedDummyUser2.Id, createdTenantId)
			assert.Nil(t, err)
			assert.NotNil(t, data)
			assert.Equal(t, createdTenantId, data.TenantId)

			// Check manually if current tenant actually added
			var checkData []*model.UserMtmTenant
			err = gormClient.Where("tenant_id", createdTenantId).Find(&checkData).Error
			assert.Nil(t, err)

			var count int64
			err = gormClient.Model(&model.UserMtmTenant{}).Where("tenant_id", createdTenantId).Count(&count).Error
			assert.Nil(t, err)

			assert.Equal(t, int64(2), count)

			// Begin test
			response, err := tenantRepo.RemoveUserFromTenant(&model.UserMtmTenant{UserId: newCreatedDummyUser2.Id, TenantId: createdTenantId}, newCreatedDummyUser.Id)
			assert.Nil(t, err)
			assert.NotEqual(t, "", response)
			assert.Contains(t, response, "[SUCCESS] Removed from tenant")

			// Database check, should be 0 tenant
			tenants, err := tenantRepo.GetTenantWithUser(newCreatedDummyUser2.Id)
			assert.Nil(t, err)
			assert.NotNil(t, tenants)
			assert.Equal(t, 0, len(tenants))

			// Clean up helper dummy
			err = gormClient.Delete(&newCreatedDummyUser2).Error
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/RemoveUserFromTenant 1")
		})

		t.Run("RemoveOwner", func(t *testing.T) {
			// From parent test block, delete the owner
			response, err := tenantRepo.RemoveUserFromTenant(&model.UserMtmTenant{UserId: newCreatedDummyUser.Id, TenantId: createdTenantId}, newCreatedDummyUser.Id)
			assert.Nil(t, err)
			assert.NotEqual(t, "", response)
			assert.Contains(t, response, "[SUCCESS] Current tenant will be archived")

			// // Database check, should be 0 tenant, will call user_mtm_tenant
			// tenants, err := tenantRepo.GetTenantWithUser(newCreatedDummyUser.Id)
			// assert.Nil(t, err)
			// assert.NotNil(t, tenants)
			// assert.Equal(t, 0, len(tenants))
		})

		// Clean up helper dummy
		err = gormClient.
			Where("user_id", newCreatedDummyUser.Id).
			Where("tenant_id", createdTenantId).
			Delete(&model.UserMtmTenant{}).Error
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/RemoveUserFromTenant 2")
		err = gormClient.
			Where("id", createdTenantId).
			Delete(&model.Tenant{}).Error
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/RemoveUserFromTenant 3")
		err = gormClient.
			Where("id", newCreatedDummyUser.Id).
			Delete(&model.User{}).Error
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/RemoveUserFromTenant 4")
	})

	t.Run("GetTenantMembers", func(t *testing.T) {
		userRepo := NewUserRepositoryImpl(gormClient)
		tenantRepo := NewTenantRepositoryImpl(gormClient)

		// create user
		dummyUser := model.User{
			Name:  "Test_TenantRepositoryImpl/GetTenantMembers user 1" + uuid.NewString(),
			Email: "testtenantrepositoryimpl" + uuid.NewString() + "@gmail.com",
		}
		password := "12345678"
		newCreatedDummyUser, err := userRepo.CreateWithEmailAndPassword(dummyUser, password)
		require.Nil(t, err)
		require.NotNil(t, newCreatedDummyUser)

		dummyUser2 := model.User{
			Name:  "Test_TenantRepositoryImpl/GetTenantMembers user 2" + uuid.NewString(),
			Email: "testtenantrepositoryimpl" + uuid.NewString() + "@gmail.com",
		}
		password = "12345678"
		newCreatedDummyUser2, err := userRepo.CreateWithEmailAndPassword(dummyUser2, password)
		require.Nil(t, err)
		require.NotNil(t, newCreatedDummyUser2)

		// Create the tenant first
		dummyTenant := &model.Tenant{
			Name:        "Test_TenantRepositoryImpl/GetTenantMembers 1 Group_" + uuid.NewString(),
			OwnerUserId: newCreatedDummyUser.Id,
			IsActive:    true,
		}
		err = tenantRepo.NewTenant(dummyTenant)
		require.Nil(t, err)

		createdDummyTenants, err := tenantRepo.GetTenantWithUser(newCreatedDummyUser.Id)
		require.Nil(t, err)
		require.NotNil(t, createdDummyTenants)
		require.GreaterOrEqual(t, len(createdDummyTenants), 1)

		// Take the id reference
		createdDummyTenantId := createdDummyTenants[0].Id

		// Add the second members
		_, err = tenantRepo.AddUserToTenant(newCreatedDummyUser2.Id, createdDummyTenantId)
		require.NoError(t, err)

		// The test itself
		users, err := tenantRepo.GetTenantMembers(createdDummyTenantId)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))

		// Clean up helper dummy
		err = gormClient.
			Where("user_id", []int{newCreatedDummyUser.Id, newCreatedDummyUser2.Id}).
			Where("tenant_id", createdDummyTenantId).
			Delete(&model.UserMtmTenant{}).Error
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/Register 1")
		err = gormClient.
			Delete(createdDummyTenants[0]).Error
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/Register 2")
		err = gormClient.
			Where("id", []int{newCreatedDummyUser.Id, newCreatedDummyUser2.Id}).
			Delete(&model.User{}).Error
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/Register 3")
	})
}
