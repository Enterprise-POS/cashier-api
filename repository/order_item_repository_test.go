package repository

import (
	"cashier-api/helper/client"
	"cashier-api/helper/query"
	"cashier-api/model"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supabase-community/supabase-go"
)

func TestOrderItemRepository(t *testing.T) {
	var supabaseClient *supabase.Client = client.CreateSupabaseClient()
	const STORE_ID = 1
	const TENANT_ID = 1
	t.Run("_PlaceOrderItem", func(t *testing.T) {
		// TEST: Normal insert
		warehouseRepo := NewWarehouseRepositoryImpl(supabaseClient)
		orderItemRepo := NewOrderItemRepositoryImpl(supabaseClient)
		storeStockRepo := StoreStockRepositoryImpl{Client: supabaseClient}

		dummyItem := &model.Item{
			ItemName:  "Test TestOrderItemRepository_PlaceOrderItem 1",
			Stocks:    100,
			StockType: model.StockTypeTracked,
			TenantId:  TENANT_ID,
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
		supabaseClient.From("order_item").Delete("", "").Eq("id", strconv.Itoa(dummyOrderItemFromDB.Id)).Execute()
		supabaseClient.From("store_stock").Delete("", "").Eq("id", strconv.Itoa(transferredStoreStockFromDB.Id)).Execute()
		supabaseClient.From("warehouse").Delete("", "").Eq("item_id", strconv.Itoa(dummyItemFromDB.ItemId)).Execute()

		/*
			TEST: invalid value
				total_quantity > 0
				total_amount >= 0
				discount_amount >= 0
				subtotal >= 0
		*/
		dummyOrderItemInvalidTotalQuantity := &model.OrderItem{
			PurchasedPrice: 10000,
			TotalQuantity:  0,
			TotalAmount:    10000,
			DiscountAmount: 0,
			Subtotal:       10000,
			TenantId:       TENANT_ID,
			StoreId:        STORE_ID,
		}
		dummyOrderItemFromDB, err = orderItemRepo.PlaceOrderItem(dummyOrderItemInvalidTotalQuantity)
		assert.Nil(t, dummyOrderItemFromDB)
		assert.NotNil(t, err)
		assert.Equal(t, "(23514) new row for relation \"order_item\" violates check constraint \"order_item_quantity_check\"", err.Error())
		dummyOrderItemInvalidTotalAmount := &model.OrderItem{
			PurchasedPrice: 10000,
			TotalQuantity:  10000,
			TotalAmount:    -1,
			DiscountAmount: 0,
			Subtotal:       10000,
			TenantId:       TENANT_ID,
			StoreId:        STORE_ID,
		}
		dummyOrderItemFromDB, err = orderItemRepo.PlaceOrderItem(dummyOrderItemInvalidTotalAmount)
		assert.Nil(t, dummyOrderItemFromDB)
		assert.NotNil(t, err)
		assert.Equal(t, "(23514) new row for relation \"order_item\" violates check constraint \"order_item_total_amount_check\"", err.Error())
		dummyOrderItemInvalidDiscountAmount := &model.OrderItem{
			PurchasedPrice: 10000,
			TotalQuantity:  1,
			TotalAmount:    10000,
			DiscountAmount: -1,
			Subtotal:       10000,
			TenantId:       TENANT_ID,
			StoreId:        STORE_ID,
		}
		dummyOrderItemFromDB, err = orderItemRepo.PlaceOrderItem(dummyOrderItemInvalidDiscountAmount)
		assert.Nil(t, dummyOrderItemFromDB)
		assert.NotNil(t, err)
		assert.Equal(t, "(23514) new row for relation \"order_item\" violates check constraint \"order_item_discount_amount_check\"", err.Error())

		// TEST: unavailable STORE_ID
		dummyOrderItem = &model.OrderItem{
			PurchasedPrice: 9999,
			TotalQuantity:  90,
			TotalAmount:    9999,
			DiscountAmount: 0,
			Subtotal:       9999,
			TenantId:       TENANT_ID,
			StoreId:        0, // -> will not valid
		}
		dummyOrderItemFromDB, err = orderItemRepo.PlaceOrderItem(dummyOrderItem)
		assert.Nil(t, dummyOrderItemFromDB)
		assert.NotNil(t, err)
		assert.Equal(t, "(23503) insert or update on table \"order_item\" violates foreign key constraint \"order_item_store_id_fkey\"", err.Error())

		// TEST: unavailable TENANT_ID
		dummyOrderItem = &model.OrderItem{
			PurchasedPrice: 9999,
			TotalQuantity:  90,
			TotalAmount:    9999,
			DiscountAmount: 0,
			Subtotal:       9999,
			TenantId:       0, // -> will not valid
			StoreId:        STORE_ID,
		}
		dummyOrderItemFromDB, err = orderItemRepo.PlaceOrderItem(dummyOrderItem)
		assert.Nil(t, dummyOrderItemFromDB)
		assert.NotNil(t, err)
		assert.Equal(t, "(23503) insert or update on table \"order_item\" violates foreign key constraint \"order_item_tenant_id_fkey\"", err.Error())
	})

	t.Run("Get", func(t *testing.T) {
		orderItemRepo := NewOrderItemRepositoryImpl(supabaseClient)

		// This is test purpose !
		// Mock user purchased items,
		// in this test we don't need the purchased_item_list
		t.Run("NormalQuery", func(t *testing.T) {
			dummyOrderItems := []*model.OrderItem{
				{
					PurchasedPrice: 10000,
					TotalQuantity:  1,
					TotalAmount:    10000,
					DiscountAmount: 0,
					Subtotal:       10000,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
				},
				{
					PurchasedPrice: 20000,
					TotalQuantity:  2,
					TotalAmount:    40000,
					DiscountAmount: 0,
					Subtotal:       40000,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
				},
				{
					PurchasedPrice: 30000,
					TotalQuantity:  3,
					TotalAmount:    90000,
					DiscountAmount: 0,
					Subtotal:       90000,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
				},
				{
					PurchasedPrice: 40000,
					TotalQuantity:  4,
					TotalAmount:    100_000,
					DiscountAmount: 60000,
					Subtotal:       160_000,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
				},
				{
					PurchasedPrice: 50000,
					TotalQuantity:  5,
					TotalAmount:    250_000,
					DiscountAmount: 0,
					Subtotal:       250_000,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
				},
			}
			for i, dummyOrderItem := range dummyOrderItems {
				dummyOrderItemFromDB, err := orderItemRepo.PlaceOrderItem(dummyOrderItem)

				require.Nil(t, err)
				require.NotNil(t, dummyOrderItemFromDB)
				require.NotNil(t, dummyOrderItemFromDB.Id)
				require.NotEqual(t, 0, dummyOrderItemFromDB.Id)
				require.Equal(t, dummyOrderItems[i].PurchasedPrice, dummyOrderItemFromDB.PurchasedPrice)
				require.Equal(t, dummyOrderItems[i].TotalQuantity, dummyOrderItemFromDB.TotalQuantity)
				require.Equal(t, dummyOrderItems[i].TotalAmount, dummyOrderItemFromDB.TotalAmount)
				require.Equal(t, dummyOrderItems[i].DiscountAmount, dummyOrderItemFromDB.DiscountAmount)
				require.Equal(t, dummyOrderItems[i].Subtotal, dummyOrderItemFromDB.Subtotal)
				require.Equal(t, TENANT_ID, dummyOrderItemFromDB.TenantId)
				require.Equal(t, STORE_ID, dummyOrderItemFromDB.StoreId)
			}

			page := 1
			limit := 5
			dummyCreatedOrderItems, count, err := orderItemRepo.Get(TENANT_ID, limit, page-1, nil)
			assert.Nil(t, err)
			assert.Greater(t, count, 4)
			assert.Equal(t, limit, len(dummyCreatedOrderItems))
			for _, dummyCreatedItem := range dummyCreatedOrderItems {
				assert.NotEqual(t, 0, dummyCreatedItem.PurchasedPrice)
				assert.NotEqual(t, 0, dummyCreatedItem.TotalQuantity)
				assert.NotEqual(t, 0, dummyCreatedItem.TotalAmount)
				assert.GreaterOrEqual(t, dummyCreatedItem.DiscountAmount, 0)
				assert.NotEqual(t, 0, dummyCreatedItem.Subtotal)
				assert.Equal(t, TENANT_ID, dummyCreatedItem.TenantId)

				// Clean up;
				_, _, err := supabaseClient.From(OrderItemTable).Delete("", "").Eq("id", strconv.Itoa(dummyCreatedItem.Id)).Execute()
				require.Nilf(t, err, "If this error; immediately delete the test data. tenantId: %d, id: %d; TestOrderItemRepository/Get/NormalQuery 1", TENANT_ID, dummyCreatedItem.Id)
			}
		})

		t.Run("DateSortByDesc", func(t *testing.T) {
			now := time.Now()
			min1Day := now.AddDate(0, 0, -1)
			min2Day := now.AddDate(0, 0, -2)
			min3Day := now.AddDate(0, 0, -3)
			min4Day := now.AddDate(0, 0, -4)
			dummyOrderItems := []*model.OrderItem{
				{
					PurchasedPrice: 100,
					TotalQuantity:  1,
					TotalAmount:    100,
					DiscountAmount: 0,
					Subtotal:       100,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
					CreatedAt:      &now,
				},
				{
					PurchasedPrice: 200,
					TotalQuantity:  2,
					TotalAmount:    400,
					DiscountAmount: 0,
					Subtotal:       400,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
					CreatedAt:      &min1Day,
				},
				{
					PurchasedPrice: 300,
					TotalQuantity:  3,
					TotalAmount:    900,
					DiscountAmount: 0,
					Subtotal:       900,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
					CreatedAt:      &min2Day,
				},
				{
					PurchasedPrice: 400,
					TotalQuantity:  4,
					TotalAmount:    1000,
					DiscountAmount: 600,
					Subtotal:       1600,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
					CreatedAt:      &min3Day,
				},
				{
					PurchasedPrice: 500,
					TotalQuantity:  5,
					TotalAmount:    2500,
					DiscountAmount: 0,
					Subtotal:       2500,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
					CreatedAt:      &min4Day,
				},
			}

			record := []*model.OrderItem{}
			for i, dummyOrderItem := range dummyOrderItems {
				dummyOrderItemFromDB, err := orderItemRepo.PlaceOrderItem(dummyOrderItem)

				require.Nil(t, err)
				require.NotNil(t, dummyOrderItemFromDB)
				require.NotNil(t, dummyOrderItemFromDB.Id)
				require.NotEqual(t, 0, dummyOrderItemFromDB.Id)
				require.Equal(t, dummyOrderItems[i].PurchasedPrice, dummyOrderItemFromDB.PurchasedPrice)
				require.Equal(t, dummyOrderItems[i].TotalQuantity, dummyOrderItemFromDB.TotalQuantity)
				require.Equal(t, dummyOrderItems[i].TotalAmount, dummyOrderItemFromDB.TotalAmount)
				require.Equal(t, dummyOrderItems[i].DiscountAmount, dummyOrderItemFromDB.DiscountAmount)
				require.Equal(t, dummyOrderItems[i].Subtotal, dummyOrderItemFromDB.Subtotal)
				require.Equal(t, TENANT_ID, dummyOrderItemFromDB.TenantId)
				require.Equal(t, STORE_ID, dummyOrderItemFromDB.StoreId)

				record = append(record, dummyOrderItemFromDB)
			}

			page := 1
			limit := 5
			dummyCreatedOrderItems, count, err := orderItemRepo.Get(TENANT_ID, limit, page-1, []*query.QueryFilter{
				{
					Column:    query.CreatedAtColumn,
					Ascending: false,
				},
			})
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, count, 5)
			assert.Equal(t, limit, len(dummyCreatedOrderItems))

			var previous *model.OrderItem
			for _, current := range dummyCreatedOrderItems {
				if previous != nil {
					assert.True(t, previous.CreatedAt.After(*current.CreatedAt),
						"Expected descending order, but %v is before %v", previous.CreatedAt, current.CreatedAt)
				}
				previous = current
			}

			// Clean up;
			for _, dummyItem := range record {
				_, _, err := supabaseClient.From(OrderItemTable).Delete("", "").Eq("id", strconv.Itoa(dummyItem.Id)).Execute()
				require.Nilf(t, err, "If this error; immediately delete the test data. tenantId: %d, id: %d; TestOrderItemRepository/Get/NormalQuery 1", TENANT_ID, dummyItem.Id)
			}
		})

		t.Run("SortByTotalAmountDesc", func(t *testing.T) {
			dummyOrderItems := []*model.OrderItem{
				{
					PurchasedPrice: 10000,
					TotalQuantity:  1,
					TotalAmount:    10000,
					DiscountAmount: 0,
					Subtotal:       10000,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
				},
				{
					PurchasedPrice: 20000,
					TotalQuantity:  2,
					TotalAmount:    40000,
					DiscountAmount: 0,
					Subtotal:       40000,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
				},
				{
					PurchasedPrice: 30000,
					TotalQuantity:  3,
					TotalAmount:    90000,
					DiscountAmount: 0,
					Subtotal:       90000,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
				},
				{
					PurchasedPrice: 40000,
					TotalQuantity:  4,
					TotalAmount:    100_000,
					DiscountAmount: 60000,
					Subtotal:       160_000,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
				},
				{
					PurchasedPrice: 50000,
					TotalQuantity:  5,
					TotalAmount:    250_000,
					DiscountAmount: 0,
					Subtotal:       250_000,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
				},
			}

			record := []*model.OrderItem{}
			for i, dummyOrderItem := range dummyOrderItems {
				dummyOrderItemFromDB, err := orderItemRepo.PlaceOrderItem(dummyOrderItem)

				require.Nil(t, err)
				require.NotNil(t, dummyOrderItemFromDB)
				require.NotNil(t, dummyOrderItemFromDB.Id)
				require.NotEqual(t, 0, dummyOrderItemFromDB.Id)
				require.Equal(t, dummyOrderItems[i].PurchasedPrice, dummyOrderItemFromDB.PurchasedPrice)
				require.Equal(t, dummyOrderItems[i].TotalQuantity, dummyOrderItemFromDB.TotalQuantity)
				require.Equal(t, dummyOrderItems[i].TotalAmount, dummyOrderItemFromDB.TotalAmount)
				require.Equal(t, dummyOrderItems[i].DiscountAmount, dummyOrderItemFromDB.DiscountAmount)
				require.Equal(t, dummyOrderItems[i].Subtotal, dummyOrderItemFromDB.Subtotal)
				require.Equal(t, TENANT_ID, dummyOrderItemFromDB.TenantId)
				require.Equal(t, STORE_ID, dummyOrderItemFromDB.StoreId)

				record = append(record, dummyOrderItemFromDB)
			}

			page := 1
			limit := 5
			dummyCreatedOrderItems, count, err := orderItemRepo.Get(TENANT_ID, limit, page-1, []*query.QueryFilter{
				{
					Column:    query.TotalAmountColumn,
					Ascending: false,
				},
			})
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, count, 5)
			assert.Equal(t, limit, len(dummyCreatedOrderItems))
			var previous *model.OrderItem
			for _, current := range dummyCreatedOrderItems {
				if previous != nil {
					assert.Less(t, current.TotalAmount, previous.TotalAmount)
				}
				previous = current
			}

			// Clean up;
			for _, dummyItem := range record {
				_, _, err := supabaseClient.From(OrderItemTable).Delete("", "").Eq("id", strconv.Itoa(dummyItem.Id)).Execute()
				require.Nilf(t, err, "If this error; immediately delete the test data. tenantId: %d, id: %d; TestOrderItemRepository/Get/NormalQuery 1", TENANT_ID, dummyItem.Id)
			}
		})
	})

	t.Run("Transactions", func(t *testing.T) {
		t.Skip("DBMS relation too deep")
	})
}
