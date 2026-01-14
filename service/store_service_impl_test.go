package service

import (
	"cashier-api/helper/query"
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

	t.Run("Edit", func(t *testing.T) {
		now := time.Now()
		t.Run("NormalEdit", func(t *testing.T) {
			editedStore := &model.Store{
				Id:       1,
				Name:     "Edited Store",
				TenantId: 1,
			}
			expectedEditedStoreReturn := &model.Store{
				Id:        1,
				Name:      "Edited Store",
				TenantId:  1,
				CreatedAt: &now,
				IsActive:  true,
			}
			storeRepository.Mock.On("Edit", editedStore).Return(expectedEditedStoreReturn, nil)
			editedTestStore, err := storeService.Edit(editedStore)
			assert.NoError(t, err)
			assert.NotNil(t, editedTestStore)
			assert.Equal(t, expectedEditedStoreReturn.Id, editedTestStore.Id)
			assert.Equal(t, expectedEditedStoreReturn.Name, editedTestStore.Name)
		})

		t.Run("WrongInput", func(t *testing.T) {
			editedStore := &model.Store{
				// Id:       1,
				Name:     "Edited Store",
				TenantId: 1,
			}
			editedTestStore, err := storeService.Edit(editedStore)
			assert.Error(t, err)
			assert.Nil(t, editedTestStore)

			editedStore = &model.Store{
				Id:       1,
				Name:     "@WeirdName",
				TenantId: 1,
			}
			editedTestStore, err = storeService.Edit(editedStore)
			assert.Error(t, err)
			assert.Nil(t, editedTestStore)

			editedStore = &model.Store{
				Id:   1,
				Name: "Edited Store",
				// TenantId: 1,
			}
			editedTestStore, err = storeService.Edit(editedStore)
			assert.Error(t, err)
			assert.Nil(t, editedTestStore)
		})
	})

	t.Run("GetSalesReport", func(t *testing.T) {
		storeId := 1
		orderItemRepository := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
		orderItemService := NewOrderItemServiceImpl(orderItemRepository)

		t.Run("NormalGetSalesReport", func(t *testing.T) {
			startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
			endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC).Unix()

			dateFilter := &query.DateFilter{
				Column:    "created_at",
				StartDate: &startDate,
				EndDate:   &endDate,
			}

			expectedReport := &repository.SalesReport{
				SumPurchasedPrice: 50000,
				SumTotalQuantity:  150,
				SumTotalAmount:    100000,
				SumDiscountAmount: 10000,
				SumSubtotal:       90000,
				SumTransactions:   50,
			}

			orderItemRepository.Mock = &mock.Mock{}
			orderItemRepository.Mock.On("GetSalesReport", tenantId, storeId, dateFilter).Return(expectedReport, nil)

			report, err := orderItemService.GetSalesReport(tenantId, storeId, dateFilter)
			assert.NoError(t, err)
			assert.NotNil(t, report)
			assert.Equal(t, expectedReport.SumPurchasedPrice, report.SumPurchasedPrice)
			assert.Equal(t, expectedReport.SumTotalQuantity, report.SumTotalQuantity)
			assert.Equal(t, expectedReport.SumTotalAmount, report.SumTotalAmount)
			assert.Equal(t, expectedReport.SumDiscountAmount, report.SumDiscountAmount)
			assert.Equal(t, expectedReport.SumSubtotal, report.SumSubtotal)
			assert.Equal(t, expectedReport.SumTransactions, report.SumTransactions)
		})

		t.Run("WithOnlyStartDate", func(t *testing.T) {
			startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

			dateFilter := &query.DateFilter{
				Column:    "created_at",
				StartDate: &startDate,
				EndDate:   nil,
			}

			expectedReport := &repository.SalesReport{
				SumPurchasedPrice: 25000,
				SumTotalQuantity:  75,
				SumTotalAmount:    50000,
				SumDiscountAmount: 5000,
				SumSubtotal:       45000,
				SumTransactions:   25,
			}

			orderItemRepository.Mock = &mock.Mock{}
			orderItemRepository.Mock.On("GetSalesReport", tenantId, storeId, dateFilter).Return(expectedReport, nil)

			report, err := orderItemService.GetSalesReport(tenantId, storeId, dateFilter)
			assert.NoError(t, err)
			assert.NotNil(t, report)
			assert.Equal(t, expectedReport.SumSubtotal, report.SumSubtotal)
		})

		t.Run("WithOnlyEndDate", func(t *testing.T) {
			endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC).Unix()

			dateFilter := &query.DateFilter{
				Column:    "created_at",
				StartDate: nil,
				EndDate:   &endDate,
			}

			expectedReport := &repository.SalesReport{
				SumPurchasedPrice: 40000,
				SumTotalQuantity:  120,
				SumTotalAmount:    80000,
				SumDiscountAmount: 8000,
				SumSubtotal:       72000,
				SumTransactions:   40,
			}

			orderItemRepository.Mock = &mock.Mock{}
			orderItemRepository.Mock.On("GetSalesReport", tenantId, storeId, dateFilter).Return(expectedReport, nil)

			report, err := orderItemService.GetSalesReport(tenantId, storeId, dateFilter)
			assert.NoError(t, err)
			assert.NotNil(t, report)
			assert.Equal(t, expectedReport.SumSubtotal, report.SumSubtotal)
		})

		t.Run("WithoutDateFilter", func(t *testing.T) {
			dateFilter := &query.DateFilter{
				Column:    "created_at",
				StartDate: nil,
				EndDate:   nil,
			}

			expectedReport := &repository.SalesReport{
				SumPurchasedPrice: 75000,
				SumTotalQuantity:  225,
				SumTotalAmount:    150000,
				SumDiscountAmount: 15000,
				SumSubtotal:       135000,
				SumTransactions:   75,
			}

			orderItemRepository.Mock = &mock.Mock{}
			orderItemRepository.Mock.On("GetSalesReport", tenantId, storeId, dateFilter).Return(expectedReport, nil)

			report, err := orderItemService.GetSalesReport(tenantId, storeId, dateFilter)
			assert.NoError(t, err)
			assert.NotNil(t, report)
			assert.Equal(t, expectedReport.SumSubtotal, report.SumSubtotal)
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
			endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC).Unix()

			dateFilter := &query.DateFilter{
				Column:    "created_at",
				StartDate: &startDate,
				EndDate:   &endDate,
			}

			// Invalid tenant id
			report, err := orderItemService.GetSalesReport(0, storeId, dateFilter)
			assert.Error(t, err)
			assert.Nil(t, report)

			// Invalid store id
			report, err = orderItemService.GetSalesReport(tenantId, -1, dateFilter)
			assert.Error(t, err)
			assert.Nil(t, report)
		})

		t.Run("InvalidDateRange", func(t *testing.T) {
			// End date before start date
			startDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC).Unix()
			endDate := time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC).Unix()

			dateFilter := &query.DateFilter{
				Column:    "created_at",
				StartDate: &startDate,
				EndDate:   &endDate,
			}

			report, err := orderItemService.GetSalesReport(tenantId, storeId, dateFilter)
			assert.Error(t, err)
			assert.Nil(t, report)
		})

		t.Run("NoDataFound", func(t *testing.T) {
			startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
			endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC).Unix()

			dateFilter := &query.DateFilter{
				Column:    "created_at",
				StartDate: &startDate,
				EndDate:   &endDate,
			}

			emptyReport := &repository.SalesReport{
				SumPurchasedPrice: 0,
				SumTotalQuantity:  0,
				SumTotalAmount:    0,
				SumDiscountAmount: 0,
				SumSubtotal:       0,
				SumTransactions:   0,
			}

			orderItemRepository.Mock = &mock.Mock{}
			orderItemRepository.Mock.On("GetSalesReport", tenantId, storeId, dateFilter).Return(emptyReport, nil)

			report, err := orderItemService.GetSalesReport(tenantId, storeId, dateFilter)
			assert.NoError(t, err)
			assert.NotNil(t, report)
			assert.Equal(t, 0, report.SumPurchasedPrice)
			assert.Equal(t, 0, report.SumTotalQuantity)
			assert.Equal(t, 0, report.SumTransactions)
		})
	})
}
