package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
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

	t.Run("LoadCashierData", func(t *testing.T) {
		t.Run("NormalLoadCashierData", func(t *testing.T) {
			expectedCashierData := []*model.CashierData{
				{
					CategoryId:   1,
					CategoryName: "Some Category",

					ItemId:    1,
					ItemName:  "Some Item Name",
					Stocks:    99,
					StockType: model.StockTypeUnlimited,
					IsActive:  true,

					StoreStockId:     1,
					StoreStockStocks: 10,
					StoreStockPrice:  10000,
				},
			}

			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 1, 1).Return(expectedCashierData, nil)
			cashierData, err := storeStockService.LoadCashierData(1, 1)
			assert.NoError(t, err)
			assert.NotNil(t, cashierData)
			assert.Len(t, cashierData, 1)
		})

		t.Run("InvalidTenantId_Zero", func(t *testing.T) {
			storeStockRepository.Mock = &mock.Mock{}
			cashierData, err := storeStockService.LoadCashierData(0, 1)
			assert.Error(t, err)
			assert.Nil(t, cashierData)
			assert.Equal(t, "Item id could not be empty or fill with 0", err.Error())
		})

		t.Run("InvalidTenantId_Negative", func(t *testing.T) {
			storeStockRepository.Mock = &mock.Mock{}
			cashierData, err := storeStockService.LoadCashierData(-1, 1)
			assert.Error(t, err)
			assert.Nil(t, cashierData)
			assert.Equal(t, "Item id could not be empty or fill with 0", err.Error())
		})

		t.Run("InvalidStoreId_Zero", func(t *testing.T) {
			storeStockRepository.Mock = &mock.Mock{}
			cashierData, err := storeStockService.LoadCashierData(1, 0)
			assert.Error(t, err)
			assert.Nil(t, cashierData)
			assert.Equal(t, "Store id could not be empty or fill with 0", err.Error())
		})

		t.Run("InvalidStoreId_Negative", func(t *testing.T) {
			storeStockRepository.Mock = &mock.Mock{}
			cashierData, err := storeStockService.LoadCashierData(1, -1)
			assert.Error(t, err)
			assert.Nil(t, cashierData)
			assert.Equal(t, "Store id could not be empty or fill with 0", err.Error())
		})

		t.Run("BothIds_Invalid", func(t *testing.T) {
			storeStockRepository.Mock = &mock.Mock{}
			cashierData, err := storeStockService.LoadCashierData(0, 0)
			assert.Error(t, err)
			assert.Nil(t, cashierData)
			assert.Equal(t, "Item id could not be empty or fill with 0", err.Error())
		})

		t.Run("EmptyResult_NoRowsFound", func(t *testing.T) {
			// PostgreSQL returns [] when no rows match
			expectedCashierData := []*model.CashierData{}

			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 1, 1).Return(expectedCashierData, nil)
			cashierData, err := storeStockService.LoadCashierData(1, 1)
			assert.NoError(t, err)
			assert.NotNil(t, cashierData)
			assert.Empty(t, cashierData)
		})

		t.Run("RepositoryError_DatabaseError", func(t *testing.T) {
			expectedError := errors.New("database connection error")

			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 1, 1).Return(nil, expectedError)
			cashierData, err := storeStockService.LoadCashierData(1, 1)
			assert.Error(t, err)
			assert.Nil(t, cashierData)
			assert.Equal(t, "database connection error", err.Error())
		})

		t.Run("RepositoryError_InvalidJSON", func(t *testing.T) {
			expectedError := errors.New("invalid character 'i' looking for beginning of value")

			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 1, 1).Return(nil, expectedError)
			cashierData, err := storeStockService.LoadCashierData(1, 1)
			assert.Error(t, err)
			assert.Nil(t, cashierData)
		})

		t.Run("MultipleItems", func(t *testing.T) {
			expectedCashierData := []*model.CashierData{
				{
					CategoryId:       1,
					CategoryName:     "Category 1",
					ItemId:           1,
					ItemName:         "Item 1",
					Stocks:           50,
					StockType:        model.StockTypeUnlimited,
					IsActive:         true,
					StoreStockId:     1,
					StoreStockStocks: 10,
					StoreStockPrice:  5000,
				},
				{
					CategoryId:       2,
					CategoryName:     "Category 2",
					ItemId:           2,
					ItemName:         "Item 2",
					Stocks:           30,
					StockType:        model.StockTypeTracked,
					IsActive:         true,
					StoreStockId:     2,
					StoreStockStocks: 5,
					StoreStockPrice:  7500,
				},
				{
					CategoryId:       3,
					CategoryName:     "Category 3",
					ItemId:           3,
					ItemName:         "Item 3",
					Stocks:           0,
					StockType:        model.StockTypeTracked,
					IsActive:         false,
					StoreStockId:     3,
					StoreStockStocks: 0,
					StoreStockPrice:  10000,
				},
			}

			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 1, 1).Return(expectedCashierData, nil)
			cashierData, err := storeStockService.LoadCashierData(1, 1)
			assert.NoError(t, err)
			assert.NotNil(t, cashierData)
			assert.Len(t, cashierData, 3)
			assert.Equal(t, "Item 1", cashierData[0].ItemName)
			assert.Equal(t, "Item 2", cashierData[1].ItemName)
			assert.Equal(t, "Item 3", cashierData[2].ItemName)
		})

		t.Run("RepositoryError_NullResponse", func(t *testing.T) {
			// With the fixed repository, this should return an error
			expectedError := errors.New("unexpected null response from database")

			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 1, 1).Return(nil, expectedError)
			cashierData, err := storeStockService.LoadCashierData(1, 1)
			assert.Error(t, err)
			assert.Nil(t, cashierData)
			assert.Equal(t, "unexpected null response from database", err.Error())
		})

		t.Run("ItemWithNullCategory", func(t *testing.T) {
			expectedCashierData := []*model.CashierData{
				{
					CategoryId:   0, // NULL category
					CategoryName: "",

					ItemId:    1,
					ItemName:  "Uncategorized Item",
					Stocks:    25,
					StockType: model.StockTypeUnlimited,
					IsActive:  true,

					StoreStockId:     1,
					StoreStockStocks: 15,
					StoreStockPrice:  8000,
				},
			}

			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 1, 1).Return(expectedCashierData, nil)
			cashierData, err := storeStockService.LoadCashierData(1, 1)
			assert.NoError(t, err)
			assert.NotNil(t, cashierData)
			assert.Len(t, cashierData, 1)
			assert.Equal(t, 0, cashierData[0].CategoryId)
			assert.Equal(t, "", cashierData[0].CategoryName)
		})

		t.Run("ItemWithZeroStocks", func(t *testing.T) {
			expectedCashierData := []*model.CashierData{
				{
					CategoryId:   1,
					CategoryName: "Category",

					ItemId:    1,
					ItemName:  "Out of Stock Item",
					Stocks:    0,
					StockType: model.StockTypeTracked,
					IsActive:  true,

					StoreStockId:     1,
					StoreStockStocks: 0,
					StoreStockPrice:  12000,
				},
			}

			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 1, 1).Return(expectedCashierData, nil)
			cashierData, err := storeStockService.LoadCashierData(1, 1)
			assert.NoError(t, err)
			assert.NotNil(t, cashierData)
			assert.Len(t, cashierData, 1)
			assert.Equal(t, 0, cashierData[0].Stocks)
			assert.Equal(t, 0, cashierData[0].StoreStockStocks)
		})

		t.Run("InactiveItem", func(t *testing.T) {
			expectedCashierData := []*model.CashierData{
				{
					CategoryId:   1,
					CategoryName: "Category",

					ItemId:    1,
					ItemName:  "Inactive Item",
					Stocks:    100,
					StockType: model.StockTypeUnlimited,
					IsActive:  false,

					StoreStockId:     1,
					StoreStockStocks: 50,
					StoreStockPrice:  15000,
				},
			}

			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 1, 1).Return(expectedCashierData, nil)
			cashierData, err := storeStockService.LoadCashierData(1, 1)
			assert.NoError(t, err)
			assert.NotNil(t, cashierData)
			assert.Len(t, cashierData, 1)
			assert.False(t, cashierData[0].IsActive)
		})

		t.Run("DifferentTenantAndStoreIds", func(t *testing.T) {
			expectedCashierData := []*model.CashierData{
				{
					CategoryId:   5,
					CategoryName: "Special Category",

					ItemId:    999,
					ItemName:  "Special Item",
					Stocks:    200,
					StockType: model.StockTypeUnlimited,
					IsActive:  true,

					StoreStockId:     555,
					StoreStockStocks: 100,
					StoreStockPrice:  20000,
				},
			}

			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 999, 555).Return(expectedCashierData, nil)
			cashierData, err := storeStockService.LoadCashierData(999, 555)
			assert.NoError(t, err)
			assert.NotNil(t, cashierData)
			assert.Len(t, cashierData, 1)
			assert.Equal(t, 999, cashierData[0].ItemId)
		})

		t.Run("RepositoryReturnsNil", func(t *testing.T) {
			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 1, 1).Return(nil, nil)
			cashierData, err := storeStockService.LoadCashierData(1, 1)
			assert.NoError(t, err)
			assert.Nil(t, cashierData)
		})

		t.Run("LargeDataSet", func(t *testing.T) {
			expectedCashierData := make([]*model.CashierData, 1000)
			for i := range 1000 {
				expectedCashierData[i] = &model.CashierData{
					CategoryId:       i + 1,
					CategoryName:     fmt.Sprintf("Category %d", i+1),
					ItemId:           i + 1,
					ItemName:         fmt.Sprintf("Item %d", i+1),
					Stocks:           i * 10,
					StockType:        model.StockTypeUnlimited,
					IsActive:         true,
					StoreStockId:     i + 1,
					StoreStockStocks: i * 5,
					StoreStockPrice:  (i + 1) * 1000,
				}
			}

			storeStockRepository.Mock = &mock.Mock{}
			storeStockRepository.Mock.On("LoadCashierData", 1, 1).Return(expectedCashierData, nil)
			cashierData, err := storeStockService.LoadCashierData(1, 1)
			assert.NoError(t, err)
			assert.NotNil(t, cashierData)
			assert.Len(t, cashierData, 1000)
		})
	})
}
