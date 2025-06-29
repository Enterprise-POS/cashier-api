package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/supabase-community/supabase-go"
)

var supabaseClient *supabase.Client = client.CreateSupabaseClient()

func TestWarehouseRepository_FindById(t *testing.T) {
	warehouseRepo := WarehouseRepositoryImpl{Client: supabaseClient}
	item := warehouseRepo.FindById(1, 1)
	assert.Equal(t, 1, item.ItemId)
	assert.Equal(t, "Apple", item.ItemName)
	assert.Equal(t, 40, item.Stocks)
	assert.Equal(t, 27, item.CreatedAt.Day())
}

func TestWarehouseRepository_CreateItem(t *testing.T) {
	var dummyItem = &model.Item{
		ItemName: "Test Name",
		Stocks:   20,
		TenantId: 1,
	}

	// Create new data
	warehouseRepo := WarehouseRepositoryImpl{Client: supabaseClient}
	dummyItemFromDB, err := warehouseRepo.CreateItem(dummyItem)
	assert.Nil(t, err)

	assert.NotEqual(t, 0, dummyItemFromDB.ItemId)
	assert.Equal(t, dummyItem.ItemName, dummyItemFromDB.ItemName)
	assert.Equal(t, dummyItem.Stocks, dummyItem.Stocks)
	assert.Equal(t, dummyItem.TenantId, dummyItemFromDB.TenantId)
	assert.Equal(t, "Item", reflect.TypeOf(*dummyItemFromDB).Name())

	// What happen if duplicate ?
	// dummyItemFromDB contain item_id, so it's allowed to create own id
	dataNil, err := warehouseRepo.CreateItem(dummyItemFromDB)
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
}

func TestWarehouseRepository_Edit(t *testing.T) {
	warehouseRepo := WarehouseRepositoryImpl{Client: supabaseClient}

	// Create dummy data first
	var dummyItem = &model.Item{
		ItemName: "Test Name2",
		Stocks:   20,
		TenantId: 1,
	}

	dummyItemFromDB, err := warehouseRepo.CreateItem(dummyItem)
	assert.Nil(t, err)

	dummyItemFromDB.ItemName = "Edit Name2"
	dummyItemFromDB.Stocks = 15

	// In the front this code below should be applied
	// Decrement -
	err = warehouseRepo.Edit(-(dummyItem.Stocks - dummyItemFromDB.Stocks), dummyItemFromDB)
	assert.Nil(t, err)

	// Check if edited item is exist
	editedDummyItemFromDB := warehouseRepo.FindById(dummyItemFromDB.ItemId, dummyItemFromDB.TenantId)
	assert.NotNil(t, editedDummyItemFromDB)
	assert.Equal(t, dummyItemFromDB.ItemId, editedDummyItemFromDB.ItemId)
	assert.Equal(t, dummyItemFromDB.ItemName, editedDummyItemFromDB.ItemName)
	assert.Equal(t, dummyItemFromDB.Stocks, editedDummyItemFromDB.Stocks)

	// Increment +
	editedDummyItemFromDB.ItemName = "Edit Name 2 :"
	editedDummyItemFromDB.Stocks = 100
	err = warehouseRepo.Edit(-(dummyItemFromDB.Stocks - editedDummyItemFromDB.Stocks), editedDummyItemFromDB)
	assert.Nil(t, err)

	editIncrementDummyItemFromDB := warehouseRepo.FindById(editedDummyItemFromDB.ItemId, editedDummyItemFromDB.TenantId)
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
}
