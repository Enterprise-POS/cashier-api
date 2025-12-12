package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supabase-community/supabase-go"
)

func TestStoreRepositoryImpl(t *testing.T) {
	var supabaseClient *supabase.Client = client.CreateSupabaseClient()
	storeRepository := NewStoreRepositoryImpl(supabaseClient)

	// Test Setup
	// - Create User
	// - Create Tenant
	userRepository := NewUserRepositoryImpl(supabaseClient)
	tenantRepository := NewTenantRepositoryImpl(supabaseClient)

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
	_, err = supabaseClient.From(TenantTable).
		Select("*", "", false).
		Eq("owner_user_id", fmt.Sprint(testUser.Id)).
		Single().
		ExecuteTo(&testTenant)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		t.Run("NormalCreate", func(t *testing.T) {
			testStoreName := "Create_TestStoreName"
			createdStore, err := storeRepository.Create(testTenant.Id, testStoreName)
			assert.NoError(t, err)
			assert.Equal(t, testStoreName, createdStore.Name)

			// NormalCreate Cleanup
			t.Cleanup(func() {
				_, count, err := supabaseClient.From(StoreTable).
					Delete("", "exact").
					Eq("id", fmt.Sprint(createdStore.Id)).
					Execute()
				require.NoError(t, err)
				require.Equal(t, 1, int(count))
			})
		})

		t.Run("InvalidTenantId", func(t *testing.T) {
			testStoreName := "Create_InvalidTenantId"
			InvalidTenantId := 0
			createdStore, err := storeRepository.Create(InvalidTenantId, testStoreName)
			assert.Error(t, err)
			assert.Nil(t, createdStore)

			// example error return message:
			// (23503) insert or update on table "store" violates foreign key constraint "store_tenant_id_fkey"
			assert.Contains(t, err.Error(), "(23503)")
		})

		t.Run("InvalidByDuplicateStoreName", func(t *testing.T) {
			testStoreName := "Create_InvalidByDuplicateStoreName"
			createdStore, err := storeRepository.Create(testTenant.Id, testStoreName)
			assert.NoError(t, err)
			assert.NotNil(t, createdStore)

			expectedNil, err := storeRepository.Create(testTenant.Id, testStoreName)
			assert.Error(t, err)
			assert.Nil(t, expectedNil)
			assert.Contains(t, err.Error(), "(23505)") // Example error message: (23505) duplicate key value violates unique constraint "unique_store_name"

			t.Cleanup(func() {
				_, count, err := supabaseClient.From(StoreTable).
					Delete("", "exact").
					Eq("id", fmt.Sprint(createdStore.Id)).
					Execute()
				require.NoError(t, err)
				require.Equal(t, 1, int(count))
			})
		})
	})

	t.Run("GetAll", func(t *testing.T) {
		// Create store for this scope only
		// There will 2 store for this scope
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
			// fmt.Println(err)
			assert.NoError(t, err)
			assert.Equal(t, 2, count)
			assert.NotNil(t, stores)

			for _, store := range stores {
				assert.Contains(t, []string{testStoreName1, testStoreName2}, store.Name)
			}
		})

		t.Run("InvalidPaginationRequest", func(t *testing.T) {
			page := 99
			limit := 10
			stores, count, err := storeRepository.GetAll(testTenant.Id, page-1, limit, true)
			assert.ErrorContains(t, err, "(PGRST103)") // example error string (PGRST103) Requested range not satisfiable
			assert.Equal(t, 0, count)
			assert.Nil(t, stores)
		})

		// Will be test by SetActivate method
		// t.Run("GetDeactivateStore", func(t *testing.T) {})

		t.Cleanup(func() {
			_, count, err := supabaseClient.From(StoreTable).
				Delete("", "exact").
				In("id", []string{strconv.Itoa(createdStore1.Id), strconv.Itoa(createdStore2.Id)}).
				Execute()
			require.NoError(t, err)
			require.Equal(t, 2, int(count))
		})
	})

	t.Run("SetActivate", func(t *testing.T) {
		t.Run("NormalSetActivate", func(t *testing.T) {
			// In this test, there will 2 test situation in the same scope
			// - set into non active / store.is_active = false
			// - set into active / store.is_active = true

			// Store for current scope only
			testStoreName := "SetActivate_NormalSetActivate"
			createdStore, err := storeRepository.Create(testTenant.Id, testStoreName)
			require.NoError(t, err)
			require.Equal(t, testStoreName, createdStore.Name)

			// 1 Set into deactivate
			err = storeRepository.SetActivate(testTenant.Id, createdStore.Id, false)
			assert.NoError(t, err)

			page := 1
			limit := 10
			stores, count, err := storeRepository.GetAll(testTenant.Id, page-1, limit, true)
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
			assert.Len(t, stores, 1)

			deactivatedStore := stores[0]
			assert.False(t, deactivatedStore.IsActive)

			// 2 Set into active
			err = storeRepository.SetActivate(testTenant.Id, createdStore.Id, true)
			assert.NoError(t, err)

			stores, count, err = storeRepository.GetAll(testTenant.Id, page-1, limit, false)
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
			assert.Len(t, stores, 1)

			reActivatedStore := stores[0]
			assert.True(t, reActivatedStore.IsActive)

			t.Cleanup(func() {
				_, count, err := supabaseClient.From(StoreTable).
					Delete("", "exact").
					Eq("id", strconv.Itoa(createdStore.Id)).
					Execute()
				require.NoError(t, err)
				require.Equal(t, 1, int(count))
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
			// Store for current scope only
			testStoreName := "StoreRepository_Edit_NormalEdit"
			createdStore, err := storeRepository.Create(testTenant.Id, testStoreName)
			require.NoError(t, err)
			require.Equal(t, testStoreName, createdStore.Name)

			// Edit the properties
			createdStore.Name = "StoreRepository_Edit_NormalEdit(EDITED)"
			updatedStore, err := storeRepository.Edit(createdStore)
			assert.NoError(t, err)
			assert.NotNil(t, updatedStore)
			assert.Equal(t, createdStore.Id, updatedStore.Id)
			assert.Equal(t, createdStore.Name, updatedStore.Name) // Here we edit the createdStore, so we can compare it

			t.Cleanup(func() {
				_, count, err := supabaseClient.From(StoreTable).
					Delete("", "exact").
					Eq("id", strconv.Itoa(updatedStore.Id)).
					Execute()
				require.NoError(t, err)
				require.Equal(t, 1, int(count))
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
		// UserMtmTable
		_, count, err := supabaseClient.From(UserMtmTenantTable).
			Delete("", "exact").
			Eq("user_id", fmt.Sprint(testUser.Id)).
			Eq("tenant_id", fmt.Sprint(testTenant.Id)).
			Execute()

		// Tenant Table
		_, count, err = supabaseClient.From(TenantTable).
			Delete("", "exact").
			Eq("id", fmt.Sprint(testTenant.Id)).
			Execute()
		require.NoError(t, err)
		require.Equal(t, 1, int(count))

		// User Table
		_, count, err = supabaseClient.From(UserTable).
			Delete("", "exact").
			Eq("id", fmt.Sprint(testUser.Id)).
			Execute()
		require.NoError(t, err)
		require.Equal(t, 1, int(count))
	})
}
