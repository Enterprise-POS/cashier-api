package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantRepositoryImpl(t *testing.T) {
	supabaseClient := client.CreateSupabaseClient()

	// This is test user id; Do not delete any accidentally
	// delete may cause another test case throw panic
	const UserId = 1

	t.Run("GetByUserId", func(t *testing.T) {
		tenantRepo := NewTenantRepositoryImpl(supabaseClient)

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
			_, _, err = supabaseClient.From(TenantTable).Delete("", "").Eq("id", strconv.Itoa(newDummyTenant.Id)).Execute()
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/Create/NormalCreate")
		})
	})

	t.Run("Create", func(t *testing.T) {
		tenantRepo := NewTenantRepositoryImpl(supabaseClient)

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
			_, _, err = supabaseClient.From(TenantTable).Delete("", "").Eq("id", strconv.Itoa(newDummyTenant.Id)).Execute()
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/Create/NormalCreate")
		})
	})

	t.Run("GetTenantWithUser", func(t *testing.T) {
		tenantRepo := NewTenantRepositoryImpl(supabaseClient)

		t.Run("NormalGet", func(t *testing.T) {
			// Create the tenant first
			dummyTenant := &model.Tenant{
				Name:        "Test_TenantRepositoryImpl/GetTenantWithUser/NormalGet 1 Group_" + uuid.NewString(),
				OwnerUserId: UserId,
				IsActive:    true,
			}
			newDummyTenant, err := tenantRepo.Create(dummyTenant)
			require.Nil(t, err)

			// For test purpose, insert into user_mtm_tenant table manually
			supabaseClient.From("user_mtm_tenant").Insert(&model.UserMtmTenant{UserId: UserId, TenantId: newDummyTenant.Id}, false, "", "", "").Execute()

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

			// Clean up
			_, _, err = supabaseClient.From("user_mtm_tenant").Delete("", "").Eq("user_id", strconv.Itoa(UserId)).Eq("tenant_id", strconv.Itoa(newDummyTenant.Id)).Execute()
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/GetTenantWithUser/NormalGet 1")
			_, _, err = supabaseClient.From(TenantTable).Delete("", "").Eq("id", strconv.Itoa(newDummyTenant.Id)).Execute()
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/GetTenantWithUser/NormalGet 2")
		})
	})

	t.Run("NewTenant", func(t *testing.T) {
		tenantRepo := NewTenantRepositoryImpl(supabaseClient)

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
			_, _, err = supabaseClient.From("user_mtm_tenant").Delete("", "").Eq("user_id", strconv.Itoa(UserId)).Eq("tenant_id", strconv.Itoa(createdTenantId)).Execute()
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/NewTenant/NormalInput 1")
			_, _, err = supabaseClient.From(TenantTable).Delete("", "").Eq("id", strconv.Itoa(createdTenantId)).Execute()
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/NewTenant/NormalInput 2")
		})
	})

	t.Run("AddUserToTenant", func(t *testing.T) {
		tenantRepo := NewTenantRepositoryImpl(supabaseClient)
		userRepo := NewUserRepositoryImpl(supabaseClient)

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
			var checkData []*model.UserMtmTenant
			count, err := supabaseClient.From(UserMtmTenantTable).Select("*", "exact", false).Eq("tenant_id", strconv.Itoa(createdTenantId)).ExecuteTo(&checkData)
			assert.Nil(t, err)
			assert.Equal(t, 2, int(count))

			// Clean up
			_, _, err = supabaseClient.From(UserMtmTenantTable).Delete("", "").Eq("user_id", strconv.Itoa(newCreatedDummyUser2.Id)).Eq("tenant_id", strconv.Itoa(createdTenantId)).Execute()
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/NewTenant/NormalInput 1")
			_, _, err = supabaseClient.From(UserTable).Delete("", "").Eq("id", strconv.Itoa(newCreatedDummyUser2.Id)).Execute()
			require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/NewTenant/NormalInput 2")
		})

		t.Run("UserIdNotAvailable", func(t *testing.T) {
			notAvailableUserId := 0
			data, err := tenantRepo.AddUserToTenant(notAvailableUserId, createdTenantId)
			assert.Nil(t, data)
			assert.NotNil(t, err)
			assert.Equal(t, "(23503) insert or update on table \"user_mtm_tenant\" violates foreign key constraint \"user_mtm_tenant_user_id_fkey\"", err.Error())
		})

		t.Run("TenantIdNotAvailable", func(t *testing.T) {
			notAvailableTenantId := 0
			data, err := tenantRepo.AddUserToTenant(newCreatedDummyUser.Id, notAvailableTenantId)
			assert.Nil(t, data)
			assert.NotNil(t, err)
			assert.Equal(t, "(23503) insert or update on table \"user_mtm_tenant\" violates foreign key constraint \"user_mtm_tenant_tenant_id_fkey\"", err.Error())
		})

		// Clean up helper dummy
		_, _, err = supabaseClient.From(UserMtmTenantTable).Delete("", "").Eq("user_id", strconv.Itoa(newCreatedDummyUser.Id)).Eq("tenant_id", strconv.Itoa(createdTenantId)).Execute()
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/Register 1")
		_, _, err = supabaseClient.From(TenantTable).Delete("", "").Eq("id", strconv.Itoa(createdTenantId)).Execute()
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/Register 2")
		_, _, err = supabaseClient.From(UserTable).Delete("", "").Eq("id", strconv.Itoa(newCreatedDummyUser.Id)).Execute()
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/Register 3")
	})

	t.Run("RemoveUserFromTenant", func(t *testing.T) {
		tenantRepo := NewTenantRepositoryImpl(supabaseClient)
		userRepo := NewUserRepositoryImpl(supabaseClient)

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
			assert.Equal(t, "\"[ERROR] Illegal action! Removing user only allowed by the owner\"", err.Error())

			// Clean up
			_, _, err = supabaseClient.From(UserTable).Delete("", "").Eq("id", strconv.Itoa(newCreatedDummyUser2.Id)).Execute()
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
			count, err := supabaseClient.From(UserMtmTenantTable).Select("*", "exact", false).Eq("tenant_id", strconv.Itoa(createdTenantId)).ExecuteTo(&checkData)
			assert.Nil(t, err)
			assert.Equal(t, 2, int(count))

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
			_, _, err = supabaseClient.From(UserTable).Delete("", "").Eq("id", strconv.Itoa(newCreatedDummyUser2.Id)).Execute()
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
		_, _, err = supabaseClient.From(UserMtmTenantTable).Delete("", "").Eq("user_id", strconv.Itoa(newCreatedDummyUser.Id)).Eq("tenant_id", strconv.Itoa(createdTenantId)).Execute()
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/RemoveUserFromTenant 2")
		_, _, err = supabaseClient.From(TenantTable).Delete("", "").Eq("id", strconv.Itoa(createdTenantId)).Execute()
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/RemoveUserFromTenant 3")
		_, _, err = supabaseClient.From(UserTable).Delete("", "").Eq("id", strconv.Itoa(newCreatedDummyUser.Id)).Execute()
		require.Nil(t, err, "If this fail, then delete data immediately TestTenantRepositoryImpl/RemoveUserFromTenant 4")
	})
}
