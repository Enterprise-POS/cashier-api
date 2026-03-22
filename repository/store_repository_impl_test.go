package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestStoreRepositoryImpl(t *testing.T) {
	var gormClient *gorm.DB = client.CreateGormClient()
	storeRepository := NewStoreRepositoryImpl(gormClient)

	// Test Setup
	// - Create User
	// - Create Tenant
	userRepository := NewUserRepositoryImpl(gormClient)
	tenantRepository := NewTenantRepositoryImpl(gormClient)

	// User
	testUserName := "Test User Name"
	testUserIdentity := fmt.Sprintf("%s@gmail.com", strings.ReplaceAll(uuid.NewString(), "-", ""))
	testUserPassword := "12345678"
	testUser, err := userRepository.CreateWithEmailAndPassword(model.User{Name: testUserName, Email: testUserIdentity}, testUserPassword)
	require.NoError(t, err)

	// Tenant
	testTenantName := "Test User Tenant"
	err = tenantRepository.NewTenant(&model.Tenant{Name: testTenantName, OwnerUserId: testUser.Id})
	require.NoError(t, err)
	// Get tenant id
	var testTenant model.Tenant
	err = gormClient.
		Where("owner_user_id", testUser.Id).
		Take(&testTenant).Error
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		t.Run("NormalCreate", func(t *testing.T) {
			testStoreName := "Create_TestStoreName"
			createdStore, err := storeRepository.Create(testTenant.Id, testStoreName)
			assert.NoError(t, err)
			assert.Equal(t, testStoreName, createdStore.Name)

			t.Cleanup(func() {
				result := gormClient.Where("id", createdStore.Id).Delete(&model.Store{})
				require.NoError(t, result.Error)
				require.Equal(t, int64(1), result.RowsAffected)
			})
		})

		t.Run("InvalidTenantId", func(t *testing.T) {
			testStoreName := "Create_InvalidTenantId"
			invalidTenantId := 0
			createdStore, err := storeRepository.Create(invalidTenantId, testStoreName)
			assert.Error(t, err)
			assert.Nil(t, createdStore)

			// Foreign key violation: tenant_id=0 does not exist
			// GORM surfaces Postgres error codes in the error message
			assert.Contains(t, err.Error(), "23503")
		})

		t.Run("InvalidByDuplicateStoreName", func(t *testing.T) {
			testStoreName := "Create_InvalidByDuplicateStoreName"
			createdStore, err := storeRepository.Create(testTenant.Id, testStoreName)
			assert.NoError(t, err)
			assert.NotNil(t, createdStore)

			expectedNil, err := storeRepository.Create(testTenant.Id, testStoreName)
			assert.Error(t, err)
			assert.Nil(t, expectedNil)

			// Unique constraint violation
			// GORM surfaces Postgres error codes in the error message
			assert.Contains(t, err.Error(), "23505")

			t.Cleanup(func() {
				result := gormClient.Where("id = ?", createdStore.Id).Delete(&model.Store{})
				require.NoError(t, result.Error)
				require.Equal(t, int64(1), result.RowsAffected)
			})
		})
	})

	t.Run("GetAll", func(t *testing.T) {
		testStoreName1 := "GetAll_TestStoreName1"
		createdStore1, err := storeRepository.Create(testTenant.Id, testStoreName1)
		require.NoError(t, err)
		require.Equal(t, testStoreName1, createdStore1.Name)

		testStoreName2 := "GetAll_TestStoreName2"
		createdStore2, err := storeRepository.Create(testTenant.Id, testStoreName2)
		require.NoError(t, err)
		require.Equal(t, testStoreName2, createdStore2.Name)

		t.Run("NormalGetAll", func(t *testing.T) {
			page := 1
			limit := 10
			stores, count, err := storeRepository.GetAll(testTenant.Id, page-1, limit, true)
			assert.NoError(t, err)
			assert.Equal(t, 2, count)
			assert.NotEmpty(t, stores)

			for _, store := range stores {
				assert.Contains(t, []string{testStoreName1, testStoreName2}, store.Name)
			}
		})

		t.Run("InvalidPaginationRequest", func(t *testing.T) {
			// GORM does not error on overflow — it simply returns empty results
			// unlike Supabase which returned (PGRST103) for out-of-range pages
			page := 99
			limit := 10
			stores, count, err := storeRepository.GetAll(testTenant.Id, page-1, limit, true)
			assert.NoError(t, err)
			assert.Equal(t, 2, count) // total count is still 2, page just has no results
			assert.Empty(t, stores)
		})

		t.Cleanup(func() {
			result := gormClient.
				Where("id", []int{createdStore1.Id, createdStore2.Id}).
				Delete(&model.Store{})
			require.NoError(t, result.Error)
			require.Equal(t, int64(2), result.RowsAffected)
		})
	})

	t.Run("SetActivate", func(t *testing.T) {
		t.Run("NormalSetActivate", func(t *testing.T) {
			testStoreName := "SetActivate_NormalSetActivate"
			createdStore, err := storeRepository.Create(testTenant.Id, testStoreName)
			require.NoError(t, err)
			require.Equal(t, testStoreName, createdStore.Name)

			// 1. Set into deactivate
			err = storeRepository.SetActivate(testTenant.Id, createdStore.Id, false)
			assert.NoError(t, err)

			page, limit := 1, 10
			stores, count, err := storeRepository.GetAll(testTenant.Id, page-1, limit, true)
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
			assert.Len(t, stores, 1)
			assert.False(t, stores[0].IsActive)

			// 2. Set back into active
			err = storeRepository.SetActivate(testTenant.Id, createdStore.Id, true)
			assert.NoError(t, err)

			stores, count, err = storeRepository.GetAll(testTenant.Id, page-1, limit, false)
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
			assert.Len(t, stores, 1)
			assert.True(t, stores[0].IsActive)

			t.Cleanup(func() {
				result := gormClient.Where("id = ?", createdStore.Id).Delete(&model.Store{})
				require.NoError(t, result.Error)
				require.Equal(t, int64(1), result.RowsAffected)
			})
		})

		t.Run("StoreIdNotAvailable", func(t *testing.T) {
			err := storeRepository.SetActivate(testTenant.Id, 0, true)
			assert.Error(t, err)
			assert.Equal(t, fmt.Sprintf("[ERROR] No store found with tenant_id=%d and id=%d", testTenant.Id, 0), err.Error())
		})
	})

	t.Run("Edit", func(t *testing.T) {
		t.Run("NormalEdit", func(t *testing.T) {
			testStoreName := "StoreRepository_Edit_NormalEdit"
			createdStore, err := storeRepository.Create(testTenant.Id, testStoreName)
			require.NoError(t, err)
			require.Equal(t, testStoreName, createdStore.Name)

			createdStore.Name = "StoreRepository_Edit_NormalEdit(EDITED)"
			updatedStore, err := storeRepository.Edit(createdStore)
			assert.NoError(t, err)
			assert.NotNil(t, updatedStore)
			assert.Equal(t, createdStore.Id, updatedStore.Id)
			assert.Equal(t, createdStore.Name, updatedStore.Name)

			t.Cleanup(func() {
				result := gormClient.Delete(updatedStore)
				require.NoError(t, result.Error)
				require.Equal(t, int64(1), result.RowsAffected)
			})
		})

		t.Run("EditNotFoundInDB", func(t *testing.T) {
			updatedStore, err := storeRepository.Edit(&model.Store{Id: 9999, TenantId: 8888, Name: "Wrong"})
			assert.Error(t, err)
			assert.Nil(t, updatedStore)
		})
	})

	// Cleanup Test Setup
	t.Cleanup(func() {
		err = gormClient.
			Where("user_id = ? AND tenant_id = ?", testUser.Id, testTenant.Id).
			Delete(&model.UserMtmTenant{}).Error
		require.NoError(t, err)

		result := gormClient.Delete(&testTenant)
		require.NoError(t, result.Error)
		require.Equal(t, int64(1), result.RowsAffected)

		result = gormClient.Delete(&testUser)
		require.NoError(t, result.Error)
		require.Equal(t, int64(1), result.RowsAffected)
	})
}
