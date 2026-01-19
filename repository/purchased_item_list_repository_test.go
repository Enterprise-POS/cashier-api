package repository

import (
	"cashier-api/helper/client"
	"cashier-api/helper/query"
	"cashier-api/model"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supabase-community/supabase-go"
)

func TestPurchasedItem(t *testing.T) {
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

	t.Run("CreateList", func(t *testing.T) {
		orderItemRepo := NewOrderItemRepositoryImpl(supabaseClient)
		purchasedItemListRepo := NewPurchasedItemRepositoryImpl(supabaseClient)

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

		dummyPurchasedItem1 := &model.PurchasedItem{
			ItemId:           APPLE_ID,
			OrderItemId:      dummyOrderItemFromDB.Id,
			Quantity:         2,
			PurchasedPrice:   APPLE_PRICE,
			DiscountAmount:   0,
			TotalAmount:      2 * APPLE_PRICE,
			ItemNameSnapshot: "Item Name Snapshot 1",
		}
		dummyPurchasedItem2 := &model.PurchasedItem{
			ItemId:           PEACH_ID,
			OrderItemId:      dummyOrderItemFromDB.Id,
			Quantity:         1,
			PurchasedPrice:   PEACH_PRICE,
			DiscountAmount:   0,
			TotalAmount:      1 * PEACH_PRICE,
			ItemNameSnapshot: "Item Name Snapshot 2",
		}

		// TEST: No error and no return data
		returnedData, err := purchasedItemListRepo.CreateList([]*model.PurchasedItem{dummyPurchasedItem1, dummyPurchasedItem2}, false)
		assert.Nil(t, returnedData)
		assert.Nil(t, err)

		// TEST: No error with return inserted data
		returnedData, err = purchasedItemListRepo.CreateList([]*model.PurchasedItem{dummyPurchasedItem1, dummyPurchasedItem2}, true)
		assert.NotNil(t, returnedData)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(returnedData))

		// Will clean up first 2 unreturned data and clean 2 returned data
		// TODO: use return data to for deleting
		supabaseClient.
			From(query.PurchasedItemTable).
			Delete("", "").
			Filter("item_id", "in", fmt.Sprintf("(%d, %d)", APPLE_ID, PEACH_ID)).
			Execute()

		// TEST: Error with un-exist foreign key
		// item_id
		// order_item_id
		dummyPurchasedItem3 := &model.PurchasedItem{
			ItemId:           -1,
			OrderItemId:      dummyOrderItemFromDB.Id,
			Quantity:         1,
			PurchasedPrice:   PEACH_PRICE,
			DiscountAmount:   0,
			TotalAmount:      1 * PEACH_PRICE,
			ItemNameSnapshot: "Item Name Snapshot",
		}
		_, err = purchasedItemListRepo.CreateList([]*model.PurchasedItem{dummyPurchasedItem3}, false)
		assert.NotNil(t, err)
		assert.Equal(t, "(23503) insert or update on table \"purchased_item_list\" violates foreign key constraint \"purchased_item_list_item_id_fkey\"", err.Error())
		dummyPurchasedItem4 := &model.PurchasedItem{
			ItemId:           PEACH_ID,
			OrderItemId:      -1,
			Quantity:         1,
			PurchasedPrice:   PEACH_PRICE,
			DiscountAmount:   0,
			TotalAmount:      1 * PEACH_PRICE,
			ItemNameSnapshot: "Item Name Snapshot",
		}
		_, err = purchasedItemListRepo.CreateList([]*model.PurchasedItem{dummyPurchasedItem4}, false)
		assert.NotNil(t, err)
		assert.Equal(t, "(23503) insert or update on table \"purchased_item_list\" violates foreign key constraint \"purchased_item_list_order_item_id_fkey\"", err.Error())

		// TEST: 1 of the rows is invalid
		_, err = purchasedItemListRepo.CreateList([]*model.PurchasedItem{dummyPurchasedItem1, dummyPurchasedItem2, dummyPurchasedItem4}, false)
		assert.NotNil(t, err)

		// Clean up order_item
		supabaseClient.From("order_item").Delete("", "").Eq("tenant_id", "1").Eq("store_id", "1").Execute()
	})

	t.Run("GetByOrderItemId", func(t *testing.T) {
		orderItemRepo := NewOrderItemRepositoryImpl(supabaseClient)
		purchasedItemListRepo := NewPurchasedItemRepositoryImpl(supabaseClient)

		t.Run("NormalGet", func(t *testing.T) {
			// Create the dummy item first
			dummyOrderItem := &model.OrderItem{
				PurchasedPrice: 20_000,
				TotalQuantity:  3,
				DiscountAmount: 10_000,
				TotalAmount:    (APPLE_PRICE * 2) + (PEACH_PRICE * 1) - 10_000,
				Subtotal:       (APPLE_PRICE * 2) + (PEACH_PRICE * 1),
				StoreId:        STORE_ID,
				TenantId:       TENANT_ID,
			}
			newDummyOrderItem, err := orderItemRepo.PlaceOrderItem(dummyOrderItem)
			require.Nil(t, err)
			require.NotNil(t, newDummyOrderItem)

			dummyPurchasedItem1 := &model.PurchasedItem{
				ItemId:           APPLE_ID,
				OrderItemId:      newDummyOrderItem.Id,
				Quantity:         2,
				PurchasedPrice:   APPLE_PRICE,
				DiscountAmount:   0,
				TotalAmount:      2 * APPLE_PRICE,
				ItemNameSnapshot: "Item Name Snapshot",
			}
			dummyPurchasedItem2 := &model.PurchasedItem{
				ItemId:           PEACH_ID,
				OrderItemId:      newDummyOrderItem.Id,
				Quantity:         1,
				PurchasedPrice:   PEACH_PRICE,
				DiscountAmount:   0,
				TotalAmount:      1 * PEACH_PRICE,
				ItemNameSnapshot: "Item Name Snapshot",
			}
			returnedData, err := purchasedItemListRepo.CreateList([]*model.PurchasedItem{dummyPurchasedItem1, dummyPurchasedItem2}, false)
			assert.Nil(t, returnedData)
			assert.Nil(t, err)

			// Begin test here
			purchasedItemsList, err := purchasedItemListRepo.GetByOrderItemId(newDummyOrderItem.Id)
			assert.Nil(t, err)
			assert.Equal(t, 2, len(purchasedItemsList))
			for _, purchasedItem := range purchasedItemsList {
				assert.Greater(t, purchasedItem.Id, 0)
				// Searching if correct item inputted
				check1 := false
				for _, testI := range []*model.PurchasedItem{dummyPurchasedItem1, dummyPurchasedItem2} {
					if testI.ItemId == purchasedItem.ItemId {
						check1 = true
						break
					}
				}
				assert.True(t, check1)
			}

			// Clear up
			_, _, err = supabaseClient.
				From(query.PurchasedItemTable).
				Delete("", "").
				Filter("id", "in", fmt.Sprintf("(%d, %d)", purchasedItemsList[0].Id, purchasedItemsList[1].Id)).
				Execute()
			require.Nil(t, err, "If this error shown, then the TestPurchasedItem/GeByOrderItemId/NormalGet error while checking the item id")

			_, _, err = supabaseClient.From("order_item").Delete("", "").Eq("tenant_id", "1").Eq("store_id", "1").Eq("id", strconv.Itoa(newDummyOrderItem.Id)).Execute()
			require.Nil(t, err, "If this error shown, then the TestPurchasedItem/GeByOrderItemId/NormalGet error while checking the item id")
		})
	})
}
