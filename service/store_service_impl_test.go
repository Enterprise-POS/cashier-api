package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStoreServiceImpl(t *testing.T) {
	storeRepository := repository.NewStoreRepositoryMock(&mock.Mock{}).(*repository.StoreRepositoryMock)
	storeService := NewStoreServiceImpl(storeRepository)

	tenantId := 1
	now := time.Now()

	t.Run("Create", func(t *testing.T) {
		t.Run("NormalCreate", func(t *testing.T) {
			expectedStore := &model.Store{
				Id:        1,
				Name:      "Test_Create_NormalCreate1",
				IsActive:  true,
				CreatedAt: &now,
				TenantId:  tenantId,
			}
			storeRepository.Mock.On("Create", tenantId, expectedStore.Name).Return(expectedStore, nil)
			createdStore, err := storeService.Create(tenantId, expectedStore.Name)
			assert.NoError(t, err)
			assert.NotNil(t, createdStore)
			assert.Equal(t, createdStore.Name, expectedStore.Name)
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			// Store Name more than 50
			invalidStoreName := "This Store Name is more than 50 So error will occurred"
			createdStore, err := storeService.Create(tenantId, invalidStoreName)
			assert.Error(t, err)
			assert.Nil(t, createdStore)

			// , Is not allowed / Not allowed characters
			invalidStoreName = "Invalid Character By ,"
			createdStore, err = storeService.Create(tenantId, invalidStoreName)
			assert.Error(t, err)
			assert.Nil(t, createdStore)

			// Invalid Tenant Id
			createdStore, err = storeService.Create(0, "Valid Name")
			assert.Error(t, err)
			assert.Nil(t, createdStore)
		})

		t.Run("DuplicateNameError", func(t *testing.T) {
			duplicateName := "Duplicate Name" // Just Example

			storeRepository.Mock = &mock.Mock{}
			storeRepository.Mock.On("Create", tenantId, duplicateName).Return(nil, errors.New("Current store name already used / duplicate name"))
			createdStore, err := storeService.Create(tenantId, duplicateName)
			assert.Error(t, err)
			assert.Equal(t, err.Error(), "Current store name already used / duplicate name")
			assert.Nil(t, createdStore)
		})
	})

	t.Run("GetAll", func(t *testing.T) {
		t.Run("NormalGetAll", func(t *testing.T) {
			expectedStores := []*model.Store{
				{
					Id:        1,
					Name:      "Test_GetAll_NormalGetAll1",
					IsActive:  true,
					CreatedAt: &now,
					TenantId:  tenantId,
				},
				{
					Id:        2,
					Name:      "Test_GetAll_NormalGetAll2",
					IsActive:  true,
					CreatedAt: &now,
					TenantId:  tenantId,
				},
			}
			page := 1
			limit := 1
			// page - 1 -> The arguments that expected to put while storeService call storeRepository
			storeRepository.Mock.On("GetAll", tenantId, page-1, limit, true).Return(expectedStores, len(expectedStores), nil)
			stores, count, err := storeService.GetAll(tenantId, page, limit, true)
			assert.Nil(t, err)
			assert.Equal(t, 2, count)
			assert.Len(t, stores, 2)
			for _, store := range stores {
				assert.Contains(t, []string{expectedStores[0].Name, expectedStores[1].Name}, store.Name)
			}
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			page := 1
			limit := 1

			// Invalid tenant id
			stores, count, err := storeService.GetAll(0, page, limit, true)
			assert.Error(t, err)
			assert.Nil(t, stores)
			assert.Equal(t, 0, count)

			// Invalid page
			stores, count, err = storeService.GetAll(tenantId, 0, limit, true)
			assert.Error(t, err)
			assert.Nil(t, stores)
			assert.Equal(t, 0, count)

			// Invalid limit
			stores, count, err = storeService.GetAll(tenantId, page, 0, true)
			assert.Error(t, err)
			assert.Nil(t, stores)
			assert.Equal(t, 0, count)
		})

		t.Run("RequestRangeError", func(t *testing.T) {
			page := 1
			limit := 1
			storeRepository.Mock = &mock.Mock{}
			storeRepository.Mock.On("GetAll", tenantId, page-1, limit, true).Return(nil, 0, errors.New("(PGRST103)"))
			stores, count, err := storeService.GetAll(tenantId, page, limit, true)
			assert.Error(t, err)
			assert.Equal(t, "Requested range not satisfiable", err.Error())
			assert.Nil(t, stores)
			assert.Equal(t, 0, count)
		})
	})

	t.Run("SetActivate", func(t *testing.T) {
		storeId := 1
		t.Run("NormalSetActivate", func(t *testing.T) {
			storeRepository.Mock.On("SetActivate", tenantId, storeId, false).Return(nil)
			err := storeService.SetActivate(tenantId, storeId, false)
			assert.NoError(t, err)
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			// Invalid tenant id
			err := storeService.SetActivate(0, storeId, false)
			assert.Error(t, err)
			// Invalid tenant id
			err = storeService.SetActivate(tenantId, 0, false)
			assert.Error(t, err)
		})

		// Handled by store_repository_test
		// t.Run("NoStoreFound", func(t *testing.T) {})
	})
}
