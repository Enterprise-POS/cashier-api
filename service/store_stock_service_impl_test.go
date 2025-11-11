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

func TestStoreStockServiceImpl(t *testing.T) {
	storeStockRepository := repository.NewStoreStockRepositoryMock(&mock.Mock{}).(*repository.StoreStockRepositoryMock)
	storeStockService := NewStoreStockServiceImpl(storeStockRepository)

	testTenantId := 1
	testStoreId := 1
	t.Run("Get", func(t *testing.T) {
		now := time.Now()

		limit := 2
		page := 1
		t.Run("NormalGet", func(t *testing.T) {

			expectedStoreStocks := []*model.StoreStock{
				{
					Id:        1,
					Stocks:    99,
					Price:     10000,
					ItemId:    1,
					CreatedAt: &now,
					TenantId:  1,
					StoreId:   1,
				},
				{
					Id:        2,
					Stocks:    10,
					Price:     15000,
					ItemId:    2,
					CreatedAt: &now,
					TenantId:  1,
					StoreId:   1,
				},
			}

			storeStockRepository.Mock.
				On("Get", testTenantId, testStoreId, limit, page-1).
				Return(expectedStoreStocks, len(expectedStoreStocks), nil)
			storeStocks, count, err := storeStockService.Get(testTenantId, testStoreId, limit, page)
			assert.NoError(t, err)
			assert.Equal(t, len(expectedStoreStocks), count)
			assert.Len(t, storeStocks, len(expectedStoreStocks))
			for _, storeStock := range storeStocks {
				assert.Contains(t, []int{
					expectedStoreStocks[0].Id,
					expectedStoreStocks[1].Id},
					storeStock.Id)
			}
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			invalidLimit := -1
			storeStocks, count, err := storeStockService.Get(testTenantId, testStoreId, invalidLimit, page)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidLimit = 0
			storeStocks, count, err = storeStockService.Get(testTenantId, testStoreId, invalidLimit, page)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidPage := -1
			storeStocks, count, err = storeStockService.Get(testTenantId, testStoreId, limit, invalidPage)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidPage = 0
			storeStocks, count, err = storeStockService.Get(testTenantId, testStoreId, limit, invalidPage)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidTenantId := -1
			storeStocks, count, err = storeStockService.Get(invalidTenantId, testStoreId, limit, page)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidTenantId = 0
			storeStocks, count, err = storeStockService.Get(invalidTenantId, testStoreId, limit, page)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidStoreId := -1
			storeStocks, count, err = storeStockService.Get(testTenantId, invalidStoreId, limit, page)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidStoreId = 0
			storeStocks, count, err = storeStockService.Get(testTenantId, invalidStoreId, limit, page)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)
		})
	})

	t.Run("GetV2", func(t *testing.T) {
		now := time.Now()

		limit := 2
		page := 1
		t.Run("NormalGet", func(t *testing.T) {

			expectedStoreStocks := []*model.StoreStockV2{
				{
					Id:        1,
					Stocks:    99,
					Price:     10000,
					ItemId:    1,
					CreatedAt: &now,
				},
				{
					Id:        2,
					Stocks:    10,
					Price:     15000,
					ItemId:    2,
					CreatedAt: &now,
				},
			}

			storeStockRepository.Mock.
				On("GetV2", testTenantId, testStoreId, limit, page-1).
				Return(expectedStoreStocks, len(expectedStoreStocks), nil)
			storeStocks, count, err := storeStockService.GetV2(testTenantId, testStoreId, limit, page, "")
			assert.NoError(t, err)
			assert.Equal(t, len(expectedStoreStocks), count)
			assert.Len(t, storeStocks, len(expectedStoreStocks))
			for _, storeStock := range storeStocks {
				assert.Contains(t, []int{
					expectedStoreStocks[0].Id,
					expectedStoreStocks[1].Id},
					storeStock.Id)
			}
		})

		t.Run("InvalidParam", func(t *testing.T) {
			invalidLimit := -1
			storeStocks, count, err := storeStockService.GetV2(testTenantId, testStoreId, invalidLimit, page, "")
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidLimit = 0
			storeStocks, count, err = storeStockService.GetV2(testTenantId, testStoreId, invalidLimit, page, "")
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidPage := -1
			storeStocks, count, err = storeStockService.GetV2(testTenantId, testStoreId, limit, invalidPage, "")
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidPage = 0
			storeStocks, count, err = storeStockService.GetV2(testTenantId, testStoreId, limit, invalidPage, "")
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidTenantId := -1
			storeStocks, count, err = storeStockService.GetV2(invalidTenantId, testStoreId, limit, page, "")
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidTenantId = 0
			storeStocks, count, err = storeStockService.GetV2(invalidTenantId, testStoreId, limit, page, "")
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidStoreId := -1
			storeStocks, count, err = storeStockService.GetV2(testTenantId, invalidStoreId, limit, page, "")
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidStoreId = 0
			storeStocks, count, err = storeStockService.GetV2(testTenantId, invalidStoreId, limit, page, "")
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			invalidNameQuery := "SELECT * FROM store_stock"
			storeStocks, count, err = storeStockService.GetV2(testTenantId, invalidStoreId, limit, page, invalidNameQuery)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)
		})
	})

	t.Run("Edit", func(t *testing.T) {
		t.Run("NormalEdit", func(t *testing.T) {
			storeStockRepository.Mock = &mock.Mock{}

			// Non exist data at DB, this is what user expected to send
			mockedStoreStock := model.StoreStock{
				Id:       1,
				Price:    10000,
				ItemId:   1,
				StoreId:  1,
				TenantId: 1,
			}
			storeStockRepository.Mock.On("Edit", &mockedStoreStock).Return(nil)
			err := storeStockService.Edit(&mockedStoreStock)
			assert.NoError(t, err)
		})

		t.Run("InvalidParams", func(t *testing.T) {
			// Invalid Price
			mockedStoreStock := &model.StoreStock{
				Id:       1,
				Price:    100_000_001,
				ItemId:   1,
				StoreId:  1,
				TenantId: 1,
			}
			err := storeStockService.Edit(mockedStoreStock)
			assert.Error(t, err)

			mockedStoreStock = &model.StoreStock{
				Id:       1,
				Price:    -1,
				ItemId:   1,
				StoreId:  1,
				TenantId: 1,
			}
			err = storeStockService.Edit(mockedStoreStock)
			assert.Error(t, err)

			// Invalid store id
			mockedStoreStock = &model.StoreStock{
				Id:       1,
				Price:    10_000,
				ItemId:   1,
				StoreId:  -1,
				TenantId: 1,
			}
			err = storeStockService.Edit(mockedStoreStock)
			assert.Error(t, err)

			// Invalid tenant id
			mockedStoreStock = &model.StoreStock{
				Id:       1,
				Price:    10_000,
				ItemId:   1,
				StoreId:  1,
				TenantId: -1,
			}
			err = storeStockService.Edit(mockedStoreStock)
			assert.Error(t, err)
		})
	})

	t.Run("TransferStockToStoreStock", func(t *testing.T) {
		quantity := 10
		testItemId := 1
		t.Run("NormalTransferStockToStoreStock", func(t *testing.T) {
			storeStockRepository.Mock.
				On("TransferStockToStoreStock", quantity, testItemId, testStoreId, testTenantId).
				Return(nil)
			err := storeStockService.TransferStockToStoreStock(quantity, testItemId, testStoreId, testTenantId)
			assert.NoError(t, err)
			assert.Nil(t, err)
		})

		t.Run("ErrorResponse", func(t *testing.T) {
			storeStockRepository.Mock.
				On("TransferStockToStoreStock", quantity, testItemId, testStoreId, 3).
				Return(errors.New("[ERROR]"))
			err := storeStockService.TransferStockToStoreStock(quantity, testItemId, testStoreId, 3)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "[ERROR]")
		})
	})

	t.Run("TransferStockToWarehouse", func(t *testing.T) {
		quantity := 10
		testItemId := 1
		t.Run("NormalTransferStockToWarehouse", func(t *testing.T) {
			storeStockRepository.Mock.
				On("TransferStockToWarehouse", quantity, testItemId, testStoreId, testTenantId).
				Return(nil)
			err := storeStockService.TransferStockToWarehouse(quantity, testItemId, testStoreId, testTenantId)
			assert.NoError(t, err)
			assert.Nil(t, err)
		})

		t.Run("ErrorResponse", func(t *testing.T) {
			storeStockRepository.Mock.
				On("TransferStockToWarehouse", quantity, testItemId, testStoreId, 2).
				Return(errors.New("[ERROR]"))
			err := storeStockService.TransferStockToWarehouse(quantity, testItemId, testStoreId, 2)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "[ERROR]")
		})
	})
}
