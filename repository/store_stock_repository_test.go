package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreStockRepository(t *testing.T) {
	gormClient := client.CreateGormClient()
	const WarehouseTable = "warehouse"
	const StoreId = 1
	const TenantId = 1

	t.Run("_Get", func(t *testing.T) {
		storeStockRepo := NewStoreStockRepositoryImpl(gormClient)
		storeStocks, count, err := storeStockRepo.Get(TenantId, StoreId, 1, 1)
		assert.Nil(t, err)
		assert.NotNil(t, count)
		assert.Greater(t, count, 0)
		assert.Len(t, storeStocks, 1)

		// Not exist page
		storeStocks, count, err = storeStockRepo.Get(TenantId, StoreId, 999, 999)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.NotNil(t, storeStocks)
	})

	t.Run("GetV2", func(t *testing.T) {
		t.Run("NormalGetV2", func(t *testing.T) {
			storeStockRepo := NewStoreStockRepositoryImpl(gormClient)
			storeStocks, count, err := storeStockRepo.GetV2(TenantId, StoreId, 1, 1, "")
			assert.NoError(t, err)
			assert.NotNil(t, storeStocks)
			assert.Greater(t, count, 0)
			assert.Len(t, storeStocks, 1)
		})

		t.Run("NotExistItemAtStoreStock", func(t *testing.T) {
			storeStockRepo := NewStoreStockRepositoryImpl(gormClient)
			storeStocks, count, err := storeStockRepo.GetV2(TenantId, 99, 1, 1, "")
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, storeStocks)

			// _, ok := err.(*exception.PostgreSQLException)
			// assert.True(t, ok)
		})
	})

	t.Run("Edit", func(t *testing.T) {
		storeStockRepo := NewStoreStockRepositoryImpl(gormClient)
		warehouseRepo := NewWarehouseRepositoryImpl(gormClient)

		// Flow: warehouse -transfer-> store_stock
		dummyItem := &model.Item{
			ItemName:  "Test StoreStockRepository_Edit 1",
			Stocks:    100,
			TenantId:  TenantId,
			StockType: model.StockTypeTracked,
		}

		_dummyItemsFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
		require.Nil(t, err, "Failed not allowed !")

		dummyItemFromDB := _dummyItemsFromDB[0]
		require.Equal(t, 100, dummyItemFromDB.Stocks)

		// stocks = 100 - 5
		err = storeStockRepo.TransferStockToStoreStock(
			5,
			dummyItemFromDB.ItemId,
			StoreId,
			TenantId,
		)
		require.Nil(t, err)

		// Get the transferred item using GORM
		var storeStockDummyFromDB model.StoreStock
		err = gormClient.
			Where("item_id", dummyItemFromDB.ItemId).
			Where("tenant_id", dummyItemFromDB.TenantId).
			Where("store_id", StoreId).
			Take(&storeStockDummyFromDB).Error
		require.Nil(t, err)

		require.Equal(t, 0, storeStockDummyFromDB.Price)
		require.Equal(t, 5, storeStockDummyFromDB.Stocks)
		require.Equal(t, StoreId, storeStockDummyFromDB.StoreId)
		require.Equal(t, dummyItemFromDB.ItemId, storeStockDummyFromDB.ItemId)
		require.Equal(t, dummyItemFromDB.TenantId, storeStockDummyFromDB.TenantId)

		t.Run("NormalEdit", func(t *testing.T) {
			expectedPrice := 10000

			err = storeStockRepo.Edit(&model.StoreStock{
				Id:       storeStockDummyFromDB.Id,
				ItemId:   storeStockDummyFromDB.ItemId,
				TenantId: storeStockDummyFromDB.TenantId,
				StoreId:  storeStockDummyFromDB.StoreId,
				Price:    expectedPrice,
			})
			assert.NoError(t, err)

			var testStoreStock model.StoreStock
			err = gormClient.
				Where("id", storeStockDummyFromDB.Id).
				First(&testStoreStock).Error
			assert.NoError(t, err)

			assert.Equal(t, expectedPrice, testStoreStock.Price)
		})

		t.Run("InvalidPriceValue", func(t *testing.T) {
			invalidPriceValue := 100_000_001

			err = storeStockRepo.Edit(&model.StoreStock{
				Id:       storeStockDummyFromDB.Id,
				ItemId:   storeStockDummyFromDB.ItemId,
				TenantId: storeStockDummyFromDB.TenantId,
				StoreId:  storeStockDummyFromDB.StoreId,
				Price:    invalidPriceValue,
			})
			assert.Error(t, err)
		})

		t.Cleanup(func() {
			// Delete store_stock
			err = gormClient.
				Where("id", storeStockDummyFromDB.Id).
				Delete(&model.StoreStock{}).Error
			require.Nil(t, err, "Failed not allowed ! Because test data will persist !")

			// Delete warehouse item
			err = gormClient.
				Where("item_id", dummyItemFromDB.ItemId).
				Delete(&model.Item{}).Error
			require.Nil(t, err, "Failed not allowed ! Because test data will persist !")
		})
	})

	t.Run("_TransferStockToWarehouse", func(t *testing.T) {
		storeStockRepo := NewStoreStockRepositoryImpl(gormClient)
		warehouseRepo := NewWarehouseRepositoryImpl(gormClient)

		// Flow: warehouse -> store_stock -> warehouse
		dummyItem := &model.Item{
			ItemName:  "Test _TransferStockWarehouse 1",
			Stocks:    100,
			TenantId:  TenantId,
			StockType: model.StockTypeTracked,
		}

		_dummyItemsFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
		require.Nil(t, err, "Failed not allowed !")

		dummyItemFromDB := _dummyItemsFromDB[0]

		// stock = 100 - 5 = 95
		err = storeStockRepo.TransferStockToStoreStock(
			5,
			dummyItemFromDB.ItemId,
			StoreId,
			TenantId,
		)
		require.Nil(t, err)

		// stock = 95 + 5 = 100
		err = storeStockRepo.TransferStockToWarehouse(
			5,
			dummyItemFromDB.ItemId,
			StoreId,
			TenantId,
		)
		assert.Nil(t, err)

		// Get the updated warehouse item using GORM
		var transferredItemFromDB model.Item
		err = gormClient.
			Where("item_id = ?", dummyItemFromDB.ItemId).
			First(&transferredItemFromDB).Error
		require.Nil(t, err)

		// Begin test
		assert.Equal(t, 100, transferredItemFromDB.Stocks)
		assert.Equal(t, dummyItemFromDB.ItemId, transferredItemFromDB.ItemId)
		assert.Equal(t, dummyItemFromDB.ItemName, transferredItemFromDB.ItemName)

		// TEST: transfer not enough stock from store_stock
		err = storeStockRepo.TransferStockToWarehouse(
			100,
			dummyItemFromDB.ItemId,
			StoreId,
			TenantId,
		)
		assert.NotNil(t, err)
		assert.Equal(t, "[ERROR] Not enough stock", err.Error())

		// Clean up

		// store_stock
		err = gormClient.
			Where("item_id", transferredItemFromDB.ItemId).
			Where("store_id", StoreId).
			Delete(&model.StoreStock{}).Error
		require.Nil(t, err)

		// warehouse (item)
		err = gormClient.
			Where("item_id = ?", transferredItemFromDB.ItemId).
			Delete(&model.Item{}).Error
		require.Nil(t, err)
	})

	t.Run("_TransferStockToStoreStock", func(t *testing.T) {
		storeStockRepo := NewStoreStockRepositoryImpl(gormClient)
		warehouseRepo := NewWarehouseRepositoryImpl(gormClient)

		// Flow: warehouse -transfer-> store_stock
		dummyItem := &model.Item{
			ItemName:  "Test _TransferStockToStoreStock 1",
			Stocks:    100,
			TenantId:  TenantId,
			StockType: model.StockTypeTracked,
		}

		_dummyItemsFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
		require.Nil(t, err, "Failed not allowed !")

		dummyItemFromDB := _dummyItemsFromDB[0]
		require.Equal(t, 100, dummyItemFromDB.Stocks)

		// stocks = 100 - 5
		err = storeStockRepo.TransferStockToStoreStock(
			5,
			dummyItemFromDB.ItemId,
			StoreId,
			TenantId,
		)
		require.Nil(t, err)

		// Get the transferred item using GORM
		var storeStockDummyFromDB model.StoreStock
		err = gormClient.
			Where("item_id = ?", dummyItemFromDB.ItemId).
			Where("tenant_id = ?", dummyItemFromDB.TenantId).
			Where("store_id = ?", StoreId).
			First(&storeStockDummyFromDB).Error
		require.Nil(t, err)

		assert.Equal(t, 0, storeStockDummyFromDB.Price) // item never existed before
		assert.Equal(t, 5, storeStockDummyFromDB.Stocks)
		assert.Equal(t, StoreId, storeStockDummyFromDB.StoreId)
		assert.Equal(t, dummyItemFromDB.ItemId, storeStockDummyFromDB.ItemId)
		assert.Equal(t, dummyItemFromDB.TenantId, storeStockDummyFromDB.TenantId)

		// TEST: not enough stock to store_stock from warehouse
		err = storeStockRepo.TransferStockToStoreStock(
			999,
			dummyItemFromDB.ItemId,
			StoreId,
			TenantId,
		)
		assert.NotNil(t, err)
		assert.Equal(t, "[ERROR] Not enough stock", err.Error())

		// Clean up

		// Delete store_stock
		err = gormClient.
			Where("id = ?", storeStockDummyFromDB.Id).
			Delete(&model.StoreStock{}).Error
		require.Nil(t, err, "Failed not allowed ! Because test data will persist !")

		// Delete warehouse item
		err = gormClient.
			Where("item_id = ?", dummyItemFromDB.ItemId).
			Delete(&model.Item{}).Error
		require.Nil(t, err, "Failed not allowed ! Because test data will persist !")
	})
}
