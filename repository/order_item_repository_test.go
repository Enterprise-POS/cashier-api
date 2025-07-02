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

func TestOrderItemRepository(t *testing.T) {
	var supabaseClient *supabase.Client = client.CreateSupabaseClient()
	const STORE_ID = 1
	const TENANT_ID = 1

	t.Run("_PlaceOrderItem", func(t *testing.T) {
		// Normal insert
		warehouseRepo := WarehouseRepositoryImpl{Client: supabaseClient}
		orderItemRepo := OrderItemRepositoryImpl{Client: supabaseClient}
		storeStockRepo := StoreStockRepositoryImpl{Client: supabaseClient}

		dummyItem := &model.Item{
			ItemName: "Test TestOrderItemRepository_PlaceOrderItem 1",
			Stocks:   100,
			TenantId: TENANT_ID,
		}

		_dummyItemFromDB, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
		require.Nil(t, err)
		require.NotNil(t, _dummyItemFromDB)
		require.Greater(t, len(_dummyItemFromDB), 0)

		var dummyItemFromDB *model.Item = _dummyItemFromDB[0]
		dummyItemFromDB.Stocks -= 10 // current: 90 | created store_stock: 10

		err = storeStockRepo.TransferStockToStoreStock(5, dummyItemFromDB.ItemId, STORE_ID, TENANT_ID)
		require.Nil(t, err)

		rStoreStockData, _, err := supabaseClient.From("store_stock").
			Select("*", "", false).
			Eq("item_id", fmt.Sprint(dummyItemFromDB.ItemId)).
			Eq("tenant_id", strconv.Itoa(TENANT_ID)).
			Eq("store_id", strconv.Itoa(STORE_ID)).
			Single().Execute()
		require.Nil(t, err)

		var transferredStoreStockFromDB = new(model.StoreStock)
		err = json.Unmarshal(rStoreStockData, transferredStoreStockFromDB)
		require.Nil(t, err)

		// This is only test so any purchased amount will not synchronized
		// dummy store_stock data price is 0, but here 10000
		dummyOrderItem := &model.OrderItem{
			PurchasedPrice: 10000,
			TotalQuantity:  1,
			TotalAmount:    10000,
			DiscountAmount: 0,
			Subtotal:       10000,
			TenantId:       TENANT_ID,
			StoreId:        STORE_ID,
		}

		dummyOrderItemFromDB, err := orderItemRepo.PlaceOrderItem(dummyOrderItem)
		assert.Nil(t, err)
		assert.NotNil(t, dummyOrderItemFromDB)
		assert.NotNil(t, dummyOrderItemFromDB.Id)
		assert.NotEqual(t, 0, dummyOrderItemFromDB.Id)
		assert.Equal(t, dummyOrderItem.PurchasedPrice, dummyOrderItemFromDB.PurchasedPrice)
		assert.Equal(t, dummyOrderItem.TotalQuantity, dummyOrderItemFromDB.TotalQuantity)
		assert.Equal(t, dummyOrderItem.TotalAmount, dummyOrderItemFromDB.TotalAmount)
		assert.Equal(t, dummyOrderItem.DiscountAmount, dummyOrderItemFromDB.DiscountAmount)
		assert.Equal(t, dummyOrderItem.Subtotal, dummyOrderItemFromDB.Subtotal)
		assert.Equal(t, TENANT_ID, dummyOrderItemFromDB.TenantId)
		assert.Equal(t, STORE_ID, dummyOrderItemFromDB.StoreId)

		// Clean up; order_item -> store_stock (also act as shop) -> warehouse
		// supabaseClient.From("order_item").Delete("", "").Eq("id", strconv.Itoa(dummyOrderItemFromDB.Id)).Execute()
		// supabaseClient.From("store_stock").Delete("", "").Eq("id", strconv.Itoa(transferredStoreStockFromDB.Id)).Execute()
		// supabaseClient.From("warehouse").Delete("", "").Eq("item_id", strconv.Itoa(dummyItemFromDB.ItemId)).Execute()
	})

}
