package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supabase-community/supabase-go"
)

func TestStoreStockRepository(t *testing.T) {
	var supabaseClient *supabase.Client = client.CreateSupabaseClient()
	const WarehouseTable = "warehouse"
	const StoreId = 1
	const TenantId = 1

	t.Run("_Get", func(t *testing.T) {
		storeStockRepo := StoreStockRepositoryImpl{Client: supabaseClient}
		storeStocks, count, err := storeStockRepo.Get(TenantId, StoreId, 1, 1)
		assert.Nil(t, err)
		assert.NotNil(t, count)
		assert.Greater(t, count, 0)
		assert.Len(t, storeStocks, 1)

		// Not exist page
		storeStocks, count, err = storeStockRepo.Get(TenantId, StoreId, 999, 999)
		assert.NotNil(t, err)
		assert.Equal(t, "(PGRST103) Requested range not satisfiable", err.Error())
		assert.Equal(t, 0, count)
		assert.Nil(t, storeStocks)
	})

	t.Run("_TransferStockToWarehouse", func(t *testing.T) {
		storeStockRepo := NewStoreStockRepositoryImpl(supabaseClient)
		warehouseRepo := NewWarehouseRepositoryImpl(supabaseClient)

		// Flow: warehouse -> store_stock -> warehouse
		dummyItem := &model.Item{
			ItemName: "Test _TransferStockWarehouse 1",
			Stocks:   100,
			TenantId: 1,
			// IsActive: , -> by default is active when inserting into DB
		}
		_dummyItemsFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
		require.Nil(t, err, "Failed not allowed !")

		dummyItemFromDB := _dummyItemsFromDB[0]

		// stock = 100 - 5 = 95
		err = storeStockRepo.TransferStockToStoreStock(5, dummyItemFromDB.ItemId, StoreId, TenantId)
		require.Nil(t, err)

		// stock = 95 + 5 = 100
		err = storeStockRepo.TransferStockToWarehouse(5, dummyItemFromDB.ItemId, StoreId, TenantId)
		assert.Nil(t, err)

		// Get the updated warehouse 'item'
		rData, _, _ := supabaseClient.From(WarehouseTable).
			Select("*", "", false).
			Eq("item_id", fmt.Sprint(dummyItemFromDB.ItemId)).
			Single().Execute()
		var transferredItemFromDB = new(model.Item)
		err = json.Unmarshal(rData, transferredItemFromDB)
		require.Nil(t, err)

		// Begin test
		assert.Equal(t, 100, transferredItemFromDB.Stocks)
		assert.Equal(t, dummyItemFromDB.ItemId, transferredItemFromDB.ItemId)
		assert.Equal(t, dummyItemFromDB.ItemName, transferredItemFromDB.ItemName)

		// TEST: transfer not enough stock from store_stock
		err = storeStockRepo.TransferStockToWarehouse(100, dummyItemFromDB.ItemId, StoreId, TenantId)
		assert.NotNil(t, err)
		assert.Equal(t, "\"[ERROR] Not enough stock\"", err.Error())

		// Clean up
		// store_stock
		supabaseClient.From(StoreStockTable).
			Delete("", "").
			Eq("item_id", fmt.Sprint(transferredItemFromDB.ItemId)).
			Eq("store_id", fmt.Sprint(StoreId)).
			Execute()

		// warehouse
		supabaseClient.From(WarehouseTable).
			Delete("", "").
			Eq("item_id", fmt.Sprint(transferredItemFromDB.ItemId)).
			Execute()
	})

	t.Run("_TransferStockToStoreStock", func(t *testing.T) {
		storeStockRepo := StoreStockRepositoryImpl{Client: supabaseClient}
		warehouseRepo := WarehouseRepositoryImpl{Client: supabaseClient}

		// Flow: warehouse -transfer-> store_stock
		dummyItem := &model.Item{
			ItemName: "Test _TransferStockToStoreStock 1",
			Stocks:   100,
			TenantId: TenantId,
			// IsActive: , -> by default is active when inserting into DB
		}
		_dummyItemsFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
		require.Nil(t, err, "Failed not allowed !")

		dummyItemFromDB := _dummyItemsFromDB[0]
		require.Equal(t, 100, dummyItemFromDB.Stocks)

		// stocks = 100 - 5
		err = storeStockRepo.TransferStockToStoreStock(5, dummyItemFromDB.ItemId, StoreId, TenantId)
		require.Nil(t, err)

		// Get the transferred item
		rData, _, _ := supabaseClient.From(StoreStockTable).
			Select("*", "", false).
			Eq("item_id", strconv.Itoa(dummyItemFromDB.ItemId)).
			Eq("tenant_id", strconv.Itoa(dummyItem.TenantId)).
			Eq("store_id", strconv.Itoa(StoreId)).
			Single().Execute()
		var storeStockDummyFromDB = new(model.StoreStock)
		err = json.Unmarshal(rData, storeStockDummyFromDB)
		require.Nil(t, err)

		assert.Equal(t, 0, storeStockDummyFromDB.Price) // Current item never exist before, so the price will be 0
		assert.Equal(t, 5, storeStockDummyFromDB.Stocks)
		assert.Equal(t, StoreId, storeStockDummyFromDB.StoreId)
		assert.Equal(t, dummyItemFromDB.ItemId, storeStockDummyFromDB.ItemId)
		assert.Equal(t, dummyItemFromDB.TenantId, storeStockDummyFromDB.TenantId)

		// TEST: not enough stock to store_stock from warehouse
		err = storeStockRepo.TransferStockToStoreStock(999, dummyItemFromDB.ItemId, StoreId, TenantId)
		assert.NotNil(t, err)
		assert.Equal(t, "\"[ERROR] Not enough stock\"", err.Error())

		// Delete the data
		// store_stock
		_, _, err = supabaseClient.From(StoreStockTable).
			Delete("", "").
			Eq("id", fmt.Sprint(storeStockDummyFromDB.Id)).
			Eq("item_id", fmt.Sprint(storeStockDummyFromDB.ItemId)).
			Eq("store_id", fmt.Sprint(storeStockDummyFromDB.StoreId)).
			Execute()
		require.Nil(t, err, "Failed not allowed ! Because test data will persist !")

		// warehouse
		_, _, err = supabaseClient.From(WarehouseTable).
			Delete("", "").
			Eq("item_id", fmt.Sprint(dummyItemFromDB.ItemId)).
			Execute()
		require.Nil(t, err, "Failed not allowed ! Because test data will persist !")
	})
}
