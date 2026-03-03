package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supabase-community/supabase-go"
	"gorm.io/gorm"
)

func TestWarehouseRepository(t *testing.T) {
	var gormClient *gorm.DB = client.CreateGormClient()
	var supabaseClient *supabase.Client = client.CreateSupabaseClient()

	warehouseRepo := NewWarehouseRepositoryImpl(gormClient)

	t.Run("TestWarehouseRepository_FindById", func(t *testing.T) {

		// TEST: normal search
		// Create item first
		var dummyItem = &model.Item{
			ItemName:  "Test TestWarehouseRepository_FindById 1",
			Stocks:    40,
			TenantId:  1,
			StockType: model.StockTypeTracked,
		}
		_dummiesFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
		require.Nil(t, err)
		require.Greater(t, len(_dummiesFromDB), 0)

		dummyItemFromDB := _dummiesFromDB[0]

		item, err := warehouseRepo.FindById(dummyItemFromDB.ItemId, 1)
		assert.Nil(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, dummyItemFromDB.ItemId, item.ItemId)
		assert.Equal(t, dummyItemFromDB.ItemName, item.ItemName)
		assert.Equal(t, dummyItemFromDB.Stocks, item.Stocks)

		// may cause error because the day not sync
		now := time.Now()
		assert.Equal(t, now.UTC().Day(), item.CreatedAt.UTC().Day())
		// fmt.Println(now)

		// TEST: id not found
		itemNotFound, err := warehouseRepo.FindById(0, 1)
		assert.Nil(t, itemNotFound)
		assert.NotNil(t, err)
		//assert.Equal(t, "(PGRST116) JSON object requested, multiple (or no) rows returned", err.Error())
		assert.Equal(t, "record not found", err.Error())

		// Clean up
		// Delete the dummy data
		_, _, err = supabaseClient.From("warehouse").
			Delete("", "").
			Eq("tenant_id", strconv.Itoa(dummyItemFromDB.TenantId)).
			Eq("item_name", dummyItemFromDB.ItemName).Execute()
		if err != nil {
			t.Fatal("unexpected error while testing to delete dummy data _CreateItem")
		}
	})

	t.Run("TestWarehouseRepository_CreateItem", func(t *testing.T) {
		var dummyItem = &model.Item{
			ItemName:  "Test Name",
			Stocks:    20,
			TenantId:  1,
			StockType: model.StockTypeTracked,
		}

		// Create new data
		_dummyItemFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(_dummyItemFromDB))

		dummyItemFromDB := _dummyItemFromDB[0]

		assert.NotEqual(t, 0, dummyItemFromDB.ItemId)
		assert.Equal(t, dummyItem.ItemName, dummyItemFromDB.ItemName)
		assert.Equal(t, dummyItem.Stocks, dummyItem.Stocks)
		assert.Equal(t, dummyItem.TenantId, dummyItemFromDB.TenantId)
		assert.Equal(t, "Item", reflect.TypeOf(*dummyItemFromDB).Name())

		// What happen if duplicate ?
		// dummyItemFromDB contain item_id, so it's allowed to create own id
		dataNil, err := warehouseRepo.CreateItem([]*model.Item{dummyItemFromDB})
		assert.NotNil(t, err)
		//assert.Equal(t, "(23505) duplicate key value violates unique constraint \"stock_pkey\"", err.Error())
		assert.Contains(t, err.Error(), "23505")
		assert.Nil(t, dataNil)

		// Delete the dummy data
		_, _, err = supabaseClient.From("warehouse").
			Delete("", "").
			Eq("item_id", strconv.Itoa(dummyItemFromDB.ItemId)).
			Eq("tenant_id", strconv.Itoa(dummyItemFromDB.TenantId)).Execute()
		if err != nil {
			t.Fatal("unexpected error while testing to delete dummy data _CreateItem")
		}
	})

	t.Run("TestWarehouseRepository_Edit", func(t *testing.T) {
		tx := gormClient.Begin()
		defer tx.Rollback()

		// Initialize repo with the transaction instead of the main client
		warehouseRepo := NewWarehouseRepositoryImpl(tx)

		// Create dummy data with initial stock of 20
		var dummyItem = &model.Item{
			ItemName:  "Test Item Original",
			Stocks:    20,
			TenantId:  1,
			IsActive:  true,
			StockType: model.StockTypeTracked,
			BasePrice: 1000,
		}

		createdItems, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(createdItems))

		// This is our source of truth from the DB
		itemInDB := createdItems[0]

		// Decrement stock (-5) ---
		// Goal: 20 -> 15
		deltaDecrement := -5
		itemInDB.ItemName = "Name After Decrement"

		err = warehouseRepo.Edit(deltaDecrement, itemInDB)
		assert.NoError(t, err)

		// Verify change
		afterDec, err := warehouseRepo.FindById(itemInDB.ItemId, itemInDB.TenantId)
		assert.NoError(t, err)
		assert.Equal(t, 15, afterDec.Stocks, "Stock should be 20 - 5 = 15")
		assert.Equal(t, "Name After Decrement", afterDec.ItemName)

		// Increment stock
		// Goal: 15 -> 100
		deltaIncrement := 85
		afterDec.ItemName = "Name After Increment"

		err = warehouseRepo.Edit(deltaIncrement, afterDec)
		assert.NoError(t, err)

		// Verify change
		afterInc, err := warehouseRepo.FindById(itemInDB.ItemId, itemInDB.TenantId)
		assert.NoError(t, err)
		assert.Equal(t, 100, afterInc.Stocks, "Stock should be 15 + 85 = 100")
		assert.Equal(t, "Name After Increment", afterInc.ItemName)

		//Negative stock protection
		// Attempting to subtract 101 from 100 should fail based on your SQL logic
		err = warehouseRepo.Edit(-101, afterInc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid quantities")

		// Non existent item
		var notExistItem = &model.Item{
			ItemId:    999999, // Random ID
			ItemName:  "Ghost Item",
			TenantId:  1,
			StockType: model.StockTypeTracked,
		}
		err = warehouseRepo.Edit(-1, notExistItem)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "never exist at warehouse")
	})

	t.Run("TestWarehouseRepository_Get", func(t *testing.T) {
		dummy1 := &model.Item{
			ItemName:  "Test WarehouseRepository_Get 1",
			Stocks:    20,
			TenantId:  1,
			StockType: model.StockTypeTracked,
		}
		dummy2 := &model.Item{
			ItemName:  "Test WarehouseRepository_Get 2",
			Stocks:    99,
			TenantId:  1,
			StockType: model.StockTypeTracked,
		}
		dummy3 := &model.Item{
			ItemName:  "Test WarehouseRepository_Get 3",
			Stocks:    99,
			TenantId:  1,
			StockType: model.StockTypeTracked,
		}
		dummy4 := &model.Item{
			ItemName:  "Test WarehouseRepository_Get 4",
			Stocks:    0,
			TenantId:  1,
			StockType: model.StockTypeTracked,
		}

		dummies := []*model.Item{dummy1, dummy2, dummy3, dummy4}

		_dummiesFromDB, err := warehouseRepo.CreateItem(dummies)
		assert.Nil(t, err)
		assert.Equal(t, 4, len(_dummiesFromDB))

		// First page
		currentPage := 1
		itemPerPage := 2
		items, count, err := warehouseRepo.Get(1, itemPerPage, currentPage, "")
		assert.NotEqual(t, 0, count)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(items))

		// Go to next page
		currentPage += 1
		items, _, err = warehouseRepo.Get(1, itemPerPage, currentPage, "")
		assert.Nil(t, err)
		assert.Equal(t, 2, len(items))

		// Check page that not even exist
		currentPage += 999
		items, _, err = warehouseRepo.Get(1, itemPerPage, currentPage, "")
		assert.NoError(t, err)
		assert.NotNil(t, items)
		assert.Equal(t, 0, len(items))

		for _, dummy := range dummies {
			gormClient.Where("item_name = ?", dummy.ItemName).Delete(&model.Item{})
		}
	})

	t.Run("Get", func(t *testing.T) {
		warehouseRepo := NewWarehouseRepositoryImpl(gormClient)

		t.Run("TenantIdNotExist", func(t *testing.T) {
			items, count, err := warehouseRepo.Get(0, 5, 1, "")
			// Because the nature how gorm handle the range, it will not return error
			//assert.Equal(t, "(PGRST103) Requested range not satisfiable", err.Error())
			assert.NoError(t, err)
			assert.Equal(t, 0, len(items))
			assert.NotNil(t, items)
			assert.Equal(t, 0, count)
		})
	})

	t.Run("FindCompleteById", func(t *testing.T) {
		t.Run("NormalFindCompleteById", func(t *testing.T) {

			var dummyItem = &model.Item{
				ItemName:  "Test TestWarehouseRepository_FindCompleteById 1",
				Stocks:    40,
				TenantId:  1,
				StockType: model.StockTypeTracked,
			}
			_dummiesFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
			require.Nil(t, err)
			require.Greater(t, len(_dummiesFromDB), 0)

			dummyItemFromDB := _dummiesFromDB[0]

			item, err := warehouseRepo.FindCompleteById(dummyItemFromDB.ItemId, 1)
			assert.Nil(t, err)
			assert.NotNil(t, item)
			assert.Equal(t, dummyItemFromDB.ItemId, item.ItemId)
			assert.Equal(t, dummyItemFromDB.ItemName, item.ItemName)
			assert.Equal(t, dummyItemFromDB.Stocks, item.Stocks)

			t.Cleanup(func() {
				// Clean up
				// Delete the dummy data
				err := gormClient.
					Where("tenant_id", dummyItemFromDB.TenantId).
					Where("item_name", dummyItemFromDB.ItemName).
					Delete(dummyItemFromDB).Error
				require.NoError(t, err, "Unexpected error while testing to delete dummy data _CreateItem")
			})
		})

		t.Run("NoDataFoundError", func(t *testing.T) {
			// TEST: id not found
			itemNotFound, err := warehouseRepo.FindCompleteById(0, 1)
			assert.Nil(t, itemNotFound)
			assert.NotNil(t, err)
			assert.Equal(t, "NO_DATA_FOUND", err.Error())
		})
	})

	t.Run("SetActivate", func(t *testing.T) {

		t.Run("NormalDeactivate", func(t *testing.T) {

			tx := gormClient.Begin()
			defer tx.Rollback()

			repo := NewWarehouseRepositoryImpl(tx)

			dummy := &model.Item{
				ItemName:  "Test_SetActivate_Deactivate",
				TenantId:  1,
				IsActive:  true,
				Stocks:    10,
				StockType: model.StockTypeTracked,
			}

			created, err := repo.CreateItem([]*model.Item{dummy})
			require.Nil(t, err)

			item := created[0]

			// deactivate
			err = repo.SetActivate(item.TenantId, item.ItemId, false)
			require.Nil(t, err)

			// verify
			var dbItem model.Item
			err = tx.First(&dbItem, "item_id = ? AND tenant_id = ?", item.ItemId, item.TenantId).Error
			require.Nil(t, err)

			assert.False(t, dbItem.IsActive)
		})

		t.Run("NormalActivate", func(t *testing.T) {

			tx := gormClient.Begin()
			defer tx.Rollback()

			repo := NewWarehouseRepositoryImpl(tx)

			dummy := &model.Item{
				ItemName:  "Test_SetActivate_Activate",
				TenantId:  1,
				IsActive:  false,
				Stocks:    10,
				StockType: model.StockTypeTracked,
			}

			created, err := repo.CreateItem([]*model.Item{dummy})
			require.Nil(t, err)

			item := created[0]

			// activate
			err = repo.SetActivate(item.TenantId, item.ItemId, true)
			require.Nil(t, err)

			var dbItem model.Item
			err = tx.First(&dbItem, "item_id = ? AND tenant_id = ?", item.ItemId, item.TenantId).Error
			require.Nil(t, err)

			assert.True(t, dbItem.IsActive)
		})

		t.Run("DeactivateThatNotExist", func(t *testing.T) {
			err := warehouseRepo.SetActivate(9999, 9999, false)

			require.NotNil(t, err)
			assert.Equal(t, "ITEM_NOT_FOUND", err.Error())
		})
	})

	t.Run("GetActiveItem", func(t *testing.T) {
		dummy1 := &model.Item{
			ItemName:  "Test WarehouseRepository_GetActiveItem 1",
			Stocks:    20,
			TenantId:  1,
			IsActive:  true,
			StockType: model.StockTypeTracked,
		}
		dummy2 := &model.Item{
			ItemName:  "Test WarehouseRepository_GetActiveItem 2",
			Stocks:    99,
			TenantId:  1,
			IsActive:  true,
			StockType: model.StockTypeTracked,
		}
		dummy3 := &model.Item{
			ItemName:  "Test WarehouseRepository_GetActiveItem 3",
			Stocks:    99,
			TenantId:  1,
			IsActive:  true,
			StockType: model.StockTypeTracked,
		}
		dummy4 := &model.Item{
			ItemName:  "Test WarehouseRepository_GetActiveItem 4",
			Stocks:    0,
			TenantId:  1,
			IsActive:  true,
			StockType: model.StockTypeTracked,
		}

		dummies := []*model.Item{dummy1, dummy2, dummy3, dummy4}

		_dummiesFromDB, err := warehouseRepo.CreateItem(dummies)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(_dummiesFromDB))

		// // First page
		currentPage := 1
		itemPerPage := 2
		items, count, err := warehouseRepo.GetActiveItem(1, itemPerPage, currentPage, "Test")
		assert.NotEqual(t, 0, count)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(items))

		// Test without nameQuery
		items, count, err = warehouseRepo.GetActiveItem(1, itemPerPage, currentPage, "")
		assert.NotEqual(t, 0, count)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(items))

		// Go to next page
		currentPage += 1
		items, _, err = warehouseRepo.GetActiveItem(1, itemPerPage, currentPage, "")
		assert.Nil(t, err)
		assert.Equal(t, 2, len(items))

		// Check page that not even exist
		currentPage += 999
		items, _, err = warehouseRepo.GetActiveItem(1, itemPerPage, currentPage, "")
		assert.NoError(t, err)
		assert.NotNil(t, items)
		assert.Equal(t, 0, len(items))

		for _, dummy := range dummies {
			supabaseClient.From("warehouse").Delete("", "").Eq("item_name", dummy.ItemName).Execute()
		}
	})
}
