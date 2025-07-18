package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supabase-community/supabase-go"
)

func TestWarehouseRepository(t *testing.T) {
	var supabaseClient *supabase.Client = client.CreateSupabaseClient()

	t.Run("TestWarehouseRepository_FindById", func(t *testing.T) {
		warehouseRepo := NewWarehouseRepositoryImpl(supabaseClient)

		// TEST: normal search
		// Create item first
		var dummyItem = &model.Item{
			ItemName: "Test TestWarehouseRepository_FindById 1",
			Stocks:   40,
			TenantId: 1,
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
		fmt.Println(now)

		// TEST: id not found
		itemNotFound, err := warehouseRepo.FindById(0, 1)
		assert.Nil(t, itemNotFound)
		assert.NotNil(t, err)
		assert.Equal(t, "(PGRST116) JSON object requested, multiple (or no) rows returned", err.Error())

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
			ItemName: "Test Name",
			Stocks:   20,
			TenantId: 1,
		}

		// Create new data
		warehouseRepo := NewWarehouseRepositoryImpl(supabaseClient)
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
		assert.Equal(t, "(23505) duplicate key value violates unique constraint \"stock_pkey\"", err.Error())
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
		warehouseRepo := NewWarehouseRepositoryImpl(supabaseClient)

		// Create dummy data first
		var dummyItem = &model.Item{
			ItemName: "Test Name2",
			Stocks:   20,
			TenantId: 1,
			IsActive: true,
		}

		_dummyItemFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
		assert.Equal(t, 1, len(_dummyItemFromDB))
		assert.Nil(t, err)

		dummyItemFromDB := _dummyItemFromDB[0]
		dummyItemFromDB.ItemName = "Edit Name2"
		dummyItemFromDB.Stocks = 15

		// In the front this code below should be applied
		// Decrement -
		err = warehouseRepo.Edit(-(dummyItem.Stocks - dummyItemFromDB.Stocks), dummyItemFromDB)
		assert.Nil(t, err)

		// Check if edited item is exist
		editedDummyItemFromDB, err := warehouseRepo.FindById(dummyItemFromDB.ItemId, dummyItemFromDB.TenantId)
		assert.Nil(t, err)
		assert.NotNil(t, editedDummyItemFromDB)
		assert.Equal(t, dummyItemFromDB.ItemId, editedDummyItemFromDB.ItemId)
		assert.Equal(t, dummyItemFromDB.ItemName, editedDummyItemFromDB.ItemName)
		assert.Equal(t, dummyItemFromDB.Stocks, editedDummyItemFromDB.Stocks)

		// Increment +
		editedDummyItemFromDB.ItemName = "Edit Name 2"
		editedDummyItemFromDB.Stocks = 100
		err = warehouseRepo.Edit(-(dummyItemFromDB.Stocks - editedDummyItemFromDB.Stocks), editedDummyItemFromDB)
		assert.Nil(t, err)

		editIncrementDummyItemFromDB, err := warehouseRepo.FindById(editedDummyItemFromDB.ItemId, editedDummyItemFromDB.TenantId)
		assert.Nil(t, err)
		assert.NotNil(t, editIncrementDummyItemFromDB)
		assert.Equal(t, editedDummyItemFromDB.ItemId, editIncrementDummyItemFromDB.ItemId)
		assert.Equal(t, editedDummyItemFromDB.ItemName, editIncrementDummyItemFromDB.ItemName)
		assert.Equal(t, editedDummyItemFromDB.Stocks, editIncrementDummyItemFromDB.Stocks)

		// Delete the dummy data
		_, _, err = supabaseClient.From("warehouse").
			Delete("", "").
			Eq("item_id", strconv.Itoa(editIncrementDummyItemFromDB.ItemId)).
			Eq("tenant_id", strconv.Itoa(editIncrementDummyItemFromDB.TenantId)).Execute()
		if err != nil {
			t.Fatal("unexpected error while testing to delete dummy data _CreateItem")
		}

		// If the items never even exist handle
		var notExistItem = &model.Item{
			ItemId:   0,
			ItemName: "Test TestWarehouseRepository_Edit Not exist item",
			Stocks:   99,
			TenantId: 1,
		}
		err = warehouseRepo.Edit(-1, notExistItem)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "[ERROR]")
		assert.Equal(t, "\"[ERROR] Fatal error, current item from store never exist at warehouse\"", err.Error())
	})

	t.Run("TestWarehouseRepository_Get", func(t *testing.T) {
		dummy1 := &model.Item{
			ItemName: "Test WarehouseRepository_Get 1",
			Stocks:   20,
			TenantId: 1,
		}
		dummy2 := &model.Item{
			ItemName: "Test WarehouseRepository_Get 2",
			Stocks:   99,
			TenantId: 1,
		}
		dummy3 := &model.Item{
			ItemName: "Test WarehouseRepository_Get 3",
			Stocks:   99,
			TenantId: 1,
		}
		dummy4 := &model.Item{
			ItemName: "Test WarehouseRepository_Get 4",
			Stocks:   0,
			TenantId: 1,
		}

		dummies := []*model.Item{dummy1, dummy2, dummy3, dummy4}

		warehouseRepo := NewWarehouseRepositoryImpl(supabaseClient)
		_dummiesFromDB, err := warehouseRepo.CreateItem(dummies)
		assert.Nil(t, err)
		assert.Equal(t, 4, len(_dummiesFromDB))

		// First page
		currentPage := 1
		itemPerPage := 2
		items, count, err := warehouseRepo.Get(1, itemPerPage, currentPage)
		assert.NotEqual(t, 0, count)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(items))

		// Go to next page
		currentPage += 1
		items, _, err = warehouseRepo.Get(1, itemPerPage, currentPage)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(items))

		// Check page that not even exist
		currentPage += 999
		items, _, err = warehouseRepo.Get(1, itemPerPage, currentPage)
		assert.NotNil(t, err)
		assert.Equal(t, "(PGRST103) Requested range not satisfiable", err.Error())
		assert.Equal(t, 0, len(items))

		for _, dummy := range dummies {
			supabaseClient.From("warehouse").Delete("", "").Eq("item_name", dummy.ItemName).Execute()
		}
	})

	t.Run("Get", func(t *testing.T) {
		warehouseRepo := NewWarehouseRepositoryImpl(supabaseClient)

		t.Run("TenantIdNotExist", func(t *testing.T) {
			items, count, err := warehouseRepo.Get(0, 5, 1)
			assert.NotNil(t, err)
			assert.Equal(t, "(PGRST103) Requested range not satisfiable", err.Error())
			assert.Equal(t, 0, len(items))
			assert.Nil(t, items)
			assert.Equal(t, 0, count)
		})
	})

	t.Run("SetActivate", func(t *testing.T) {
		warehouseRepo := NewWarehouseRepositoryImpl(supabaseClient)

		t.Run("NormalDeactivate", func(t *testing.T) {
			dummyItem := &model.Item{
				ItemName: "Test_TestWarehouseRepository_SetActivate_NormalActivate 1",
				TenantId: 1,
				IsActive: true,
				Stocks:   10,
			}
			_createdItemsFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
			require.Nil(t, err)

			createdItemFromDB := _createdItemsFromDB[0]

			// Set into false
			createdItemFromDB.IsActive = false

			// This what we test
			err = warehouseRepo.SetActivate(createdItemFromDB.TenantId, createdItemFromDB.ItemId, createdItemFromDB.IsActive)
			assert.Nil(t, err)

			// Check if the data at database correct
			var testItem *model.Item
			_, err = supabaseClient.From(WarehouseTable).Select("*", "", false).Eq("item_id", strconv.Itoa(createdItemFromDB.ItemId)).Single().ExecuteTo(&testItem)
			require.Nil(t, err, "If this error, data persist at warehouse table, delete immediately")

			assert.NotNil(t, testItem)
			assert.Equal(t, createdItemFromDB.ItemId, testItem.ItemId)
			assert.Equal(t, createdItemFromDB.ItemName, testItem.ItemName)
			assert.Equal(t, createdItemFromDB.Stocks, testItem.Stocks)
			assert.Equal(t, createdItemFromDB.TenantId, testItem.TenantId)
			assert.NotEqual(t, dummyItem.IsActive, testItem.IsActive)

			_, _, err = supabaseClient.From(WarehouseTable).Delete("", "").Eq("item_id", strconv.Itoa(testItem.ItemId)).Eq("tenant_id", strconv.Itoa(testItem.TenantId)).Execute()
			require.Nilf(t, err, "If this error, data persist at warehouse table itemId: %d, tenantId: %d; TestWarehouse/SetActivate/NormalDeactivate", testItem.ItemId, testItem.TenantId)
		})

		t.Run("NormalActivate", func(t *testing.T) {
			dummyItem := &model.Item{
				ItemName: "Test_TestWarehouseRepository_SetActivate_NormalActivate 1",
				TenantId: 1,
				IsActive: true,
				Stocks:   10,
			}
			_createdItemsFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
			require.Nil(t, err)

			createdItemFromDB := _createdItemsFromDB[0]

			// Set into false
			createdItemFromDB.IsActive = false

			// This what we test
			err = warehouseRepo.SetActivate(createdItemFromDB.TenantId, createdItemFromDB.ItemId, createdItemFromDB.IsActive)
			assert.Nil(t, err)

			// Check if the data at database correct
			var testItem *model.Item
			_, err = supabaseClient.From(WarehouseTable).Select("*", "", false).Eq("item_id", strconv.Itoa(createdItemFromDB.ItemId)).Single().ExecuteTo(&testItem)
			require.Nil(t, err, "If this error, data persist at warehouse table, delete immediately")

			assert.NotNil(t, testItem)
			assert.Equal(t, createdItemFromDB.ItemId, testItem.ItemId)
			assert.Equal(t, createdItemFromDB.ItemName, testItem.ItemName)
			assert.Equal(t, createdItemFromDB.Stocks, testItem.Stocks)
			assert.Equal(t, createdItemFromDB.TenantId, testItem.TenantId)
			assert.NotEqual(t, dummyItem.IsActive, testItem.IsActive)

			_, _, err = supabaseClient.From(WarehouseTable).Delete("", "").Eq("item_id", strconv.Itoa(testItem.ItemId)).Eq("tenant_id", strconv.Itoa(testItem.TenantId)).Execute()
			require.Nilf(t, err, "If this error, data persist at warehouse table itemId: %d, tenantId: %d; TestWarehouse/SetActivate/NormalDeactivate", testItem.ItemId, testItem.TenantId)
		})

		t.Run("DeactivateThatNotExist", func(t *testing.T) {
			var (
				itemId   = 0
				tenantId = 0
			)
			err := warehouseRepo.SetActivate(tenantId, itemId, false)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "(PGRST116)")
			assert.Equal(t, "(PGRST116) JSON object requested, multiple (or no) rows returned", err.Error())
		})
	})
}
