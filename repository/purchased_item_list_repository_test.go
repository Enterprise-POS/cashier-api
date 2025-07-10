package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supabase-community/supabase-go"
)

func TestPurchasedItemList(t *testing.T) {
	var supabaseClient *supabase.Client = client.CreateSupabaseClient()

	const (
		TENANT_ID int = 1
		STORE_ID  int = 1

		// just for faster testing
		APPLE_ID    int = 1
		PEACH_ID    int = 2
		APPLE_PRICE int = 10000
		PEACH_PRICE int = 20000
	)

	t.Run("_CreateList", func(t *testing.T) {
		orderItemRepo := OrderItemRepositoryImpl{Client: supabaseClient}
		purchasedItemListRepo := PurchasedItemListRepositoryImpl{Client: supabaseClient}

		// The dummy data
		dummyOrderItem := &model.OrderItem{
			PurchasedPrice: 40000,
			TotalQuantity:  4,
			DiscountAmount: 0,
			TotalAmount:    40000,
			Subtotal:       40000,
			StoreId:        STORE_ID,
			TenantId:       TENANT_ID,
		}
		dummyOrderItemFromDB, err := orderItemRepo.PlaceOrderItem(dummyOrderItem)
		require.Nil(t, err)
		require.NotNil(t, dummyOrderItemFromDB)

		dummyPurchasedItemList1 := &model.PurchasedItemList{
			ItemId:         APPLE_ID,
			OrderItemId:    dummyOrderItemFromDB.Id,
			Quantity:       2,
			PurchasedPrice: APPLE_PRICE,
			DiscountAmount: 0,
			TotalAmount:    2 * APPLE_PRICE,
		}
		dummyPurchasedItemList2 := &model.PurchasedItemList{
			ItemId:         PEACH_ID,
			OrderItemId:    dummyOrderItemFromDB.Id,
			Quantity:       1,
			PurchasedPrice: PEACH_PRICE,
			DiscountAmount: 0,
			TotalAmount:    1 * PEACH_PRICE,
		}

		// TEST: No error and no return data
		returnedData, err := purchasedItemListRepo.CreateList([]*model.PurchasedItemList{dummyPurchasedItemList1, dummyPurchasedItemList2}, false)
		assert.Nil(t, returnedData)
		assert.Nil(t, err)

		// TEST: No error with return inserted data
		returnedData, err = purchasedItemListRepo.CreateList([]*model.PurchasedItemList{dummyPurchasedItemList1, dummyPurchasedItemList2}, true)
		assert.NotNil(t, returnedData)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(returnedData))

		// Will clean up first 2 unreturned data and clean 2 returned data
		supabaseClient.
			From(PurchasedItemListTable).
			Delete("", "").
			Filter("item_id", "in", fmt.Sprintf("(%d, %d)", APPLE_ID, PEACH_ID)).
			Execute()

		// TEST: Error with un-exist foreign key
		// item_id
		// order_item_id
		dummyPurchasedItemList3 := &model.PurchasedItemList{
			ItemId:         -1,
			OrderItemId:    dummyOrderItemFromDB.Id,
			Quantity:       1,
			PurchasedPrice: PEACH_PRICE,
			DiscountAmount: 0,
			TotalAmount:    1 * PEACH_PRICE,
		}
		_, err = purchasedItemListRepo.CreateList([]*model.PurchasedItemList{dummyPurchasedItemList3}, false)
		assert.NotNil(t, err)
		assert.Equal(t, "(23503) insert or update on table \"purchased_item_list\" violates foreign key constraint \"purchased_item_list_item_id_fkey\"", err.Error())
		dummyPurchasedItemList4 := &model.PurchasedItemList{
			ItemId:         PEACH_ID,
			OrderItemId:    -1,
			Quantity:       1,
			PurchasedPrice: PEACH_PRICE,
			DiscountAmount: 0,
			TotalAmount:    1 * PEACH_PRICE,
		}
		_, err = purchasedItemListRepo.CreateList([]*model.PurchasedItemList{dummyPurchasedItemList4}, false)
		assert.NotNil(t, err)
		assert.Equal(t, "(23503) insert or update on table \"purchased_item_list\" violates foreign key constraint \"purchased_item_list_order_item_id_fkey\"", err.Error())

		// TEST: 1 of the rows is invalid
		_, err = purchasedItemListRepo.CreateList([]*model.PurchasedItemList{dummyPurchasedItemList1, dummyPurchasedItemList2, dummyPurchasedItemList4}, false)
		assert.NotNil(t, err)

		// Clean up order_item
		supabaseClient.From("order_item").Delete("", "").Eq("tenant_id", "1").Eq("store_id", "1").Execute()
	})
}
