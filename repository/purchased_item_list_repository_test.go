package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPurchasedItem(t *testing.T) {
	var gormClient *gorm.DB = client.CreateGormClient()

	const (
		APPLE_PRICE int = 10000
		PEACH_PRICE int = 20000
	)

	// seedPurchasedItemDependencies creates user→tenant→store→item(apple)→item(peach)
	// and an order_item within the given transaction.
	// All rows roll back automatically after each test.
	seedPurchasedItemDependencies := func(t *testing.T, tx *gorm.DB) (appleId int, peachId int, orderItemId int) {
		t.Helper()

		tenantId, storeId := seedOrderItemTestDependencies(t, tx)

		apple := &model.Item{
			ItemName:  "Apple",
			Stocks:    100,
			StockType: model.StockTypeTracked,
			BasePrice: APPLE_PRICE,
			TenantId:  tenantId,
			IsActive:  true,
		}
		require.NoError(t, tx.Create(apple).Error)
		require.NotZero(t, apple.ItemId)

		peach := &model.Item{
			ItemName:  "Peach",
			Stocks:    100,
			StockType: model.StockTypeTracked,
			BasePrice: PEACH_PRICE,
			TenantId:  tenantId,
			IsActive:  true,
		}
		require.NoError(t, tx.Create(peach).Error)
		require.NotZero(t, peach.ItemId)

		orderItem := &model.OrderItem{
			PurchasedPrice: 40000,
			TotalQuantity:  4,
			DiscountAmount: 0,
			TotalAmount:    40000,
			Subtotal:       40000,
			StoreId:        storeId,
			TenantId:       tenantId,
		}
		require.NoError(t, tx.Create(orderItem).Error)
		require.NotZero(t, orderItem.Id)

		return apple.ItemId, peach.ItemId, orderItem.Id
	}

	t.Run("CreateList", func(t *testing.T) {
		t.Run("NoReturnData", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			appleId, peachId, orderItemId := seedPurchasedItemDependencies(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			data := []*model.PurchasedItem{
				{
					ItemId:             appleId,
					OrderItemId:        orderItemId,
					Quantity:           2,
					StorePriceSnapshot: APPLE_PRICE,
					DiscountAmount:     0,
					TotalAmount:        2 * APPLE_PRICE,
					ItemNameSnapshot:   "Apple Snapshot",
					BasePriceSnapshot:  100,
				},
				{
					ItemId:             peachId,
					OrderItemId:        orderItemId,
					Quantity:           1,
					StorePriceSnapshot: PEACH_PRICE,
					DiscountAmount:     0,
					TotalAmount:        1 * PEACH_PRICE,
					ItemNameSnapshot:   "Peach Snapshot",
					BasePriceSnapshot:  100,
				},
			}

			returnedData, err := repo.CreateList(data, false)
			assert.Nil(t, returnedData)
			assert.Nil(t, err)
		})

		t.Run("WithReturnData", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			appleId, peachId, orderItemId := seedPurchasedItemDependencies(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			data := []*model.PurchasedItem{
				{
					ItemId:             appleId,
					OrderItemId:        orderItemId,
					Quantity:           2,
					StorePriceSnapshot: APPLE_PRICE,
					DiscountAmount:     0,
					TotalAmount:        2 * APPLE_PRICE,
					ItemNameSnapshot:   "Apple Snapshot",
					BasePriceSnapshot:  100,
				},
				{
					ItemId:             peachId,
					OrderItemId:        orderItemId,
					Quantity:           1,
					StorePriceSnapshot: PEACH_PRICE,
					DiscountAmount:     0,
					TotalAmount:        1 * PEACH_PRICE,
					ItemNameSnapshot:   "Peach Snapshot",
					BasePriceSnapshot:  100,
				},
			}

			returnedData, err := repo.CreateList(data, true)
			assert.Nil(t, err)
			assert.NotNil(t, returnedData)
			assert.Equal(t, 2, len(returnedData))
			for _, item := range returnedData {
				assert.NotZero(t, item.Id)
			}
		})

		t.Run("InvalidItemIdForeignKey", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			_, _, orderItemId := seedPurchasedItemDependencies(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			data := []*model.PurchasedItem{
				{
					ItemId:             -1, // Invalid: FK violation expected
					OrderItemId:        orderItemId,
					Quantity:           1,
					StorePriceSnapshot: PEACH_PRICE,
					DiscountAmount:     0,
					TotalAmount:        1 * PEACH_PRICE,
					ItemNameSnapshot:   "Snapshot",
					BasePriceSnapshot:  100,
				},
			}

			_, err := repo.CreateList(data, false)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "23503")
		})

		t.Run("InvalidOrderItemIdForeignKey", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			_, peachId, _ := seedPurchasedItemDependencies(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			data := []*model.PurchasedItem{
				{
					ItemId:             peachId,
					OrderItemId:        -1, // Invalid: FK violation expected
					Quantity:           1,
					StorePriceSnapshot: PEACH_PRICE,
					DiscountAmount:     0,
					TotalAmount:        1 * PEACH_PRICE,
					ItemNameSnapshot:   "Snapshot",
					BasePriceSnapshot:  100,
				},
			}

			_, err := repo.CreateList(data, false)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "23503")
		})

		t.Run("OneInvalidRowFailsAll", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			appleId, peachId, orderItemId := seedPurchasedItemDependencies(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			data := []*model.PurchasedItem{
				{
					ItemId:             appleId,
					OrderItemId:        orderItemId,
					Quantity:           2,
					StorePriceSnapshot: APPLE_PRICE,
					DiscountAmount:     0,
					TotalAmount:        2 * APPLE_PRICE,
					ItemNameSnapshot:   "Apple Snapshot",
					BasePriceSnapshot:  100,
				},
				{
					ItemId:             peachId,
					OrderItemId:        orderItemId,
					Quantity:           1,
					StorePriceSnapshot: PEACH_PRICE,
					DiscountAmount:     0,
					TotalAmount:        1 * PEACH_PRICE,
					ItemNameSnapshot:   "Peach Snapshot",
					BasePriceSnapshot:  100,
				},
				{
					ItemId:             peachId,
					OrderItemId:        -1, // Invalid: this row fails the whole batch
					Quantity:           1,
					StorePriceSnapshot: PEACH_PRICE,
					DiscountAmount:     0,
					TotalAmount:        1 * PEACH_PRICE,
					ItemNameSnapshot:   "Snapshot",
					BasePriceSnapshot:  100,
				},
			}

			_, err := repo.CreateList(data, false)
			assert.NotNil(t, err)
		})
	})

	t.Run("GetByOrderItemId", func(t *testing.T) {
		t.Run("NormalGet", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			appleId, peachId, orderItemId := seedPurchasedItemDependencies(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			data := []*model.PurchasedItem{
				{
					ItemId:             appleId,
					OrderItemId:        orderItemId,
					Quantity:           2,
					StorePriceSnapshot: APPLE_PRICE,
					DiscountAmount:     0,
					TotalAmount:        2 * APPLE_PRICE,
					ItemNameSnapshot:   "Apple Snapshot",
					BasePriceSnapshot:  100,
				},
				{
					ItemId:             peachId,
					OrderItemId:        orderItemId,
					Quantity:           1,
					StorePriceSnapshot: PEACH_PRICE,
					DiscountAmount:     0,
					TotalAmount:        1 * PEACH_PRICE,
					ItemNameSnapshot:   "Peach Snapshot",
					BasePriceSnapshot:  100,
				},
			}

			_, err := repo.CreateList(data, false)
			require.Nil(t, err)

			purchasedItemsList, err := repo.GetByOrderItemId(orderItemId)
			assert.Nil(t, err)
			assert.Equal(t, 2, len(purchasedItemsList))

			// Collect inserted item IDs for validation
			insertedItemIds := map[int]bool{appleId: true, peachId: true}
			for _, purchasedItem := range purchasedItemsList {
				assert.Greater(t, purchasedItem.Id, 0)
				assert.True(t, insertedItemIds[purchasedItem.ItemId],
					"unexpected item_id %d in result", purchasedItem.ItemId)
			}
		})

		t.Run("NotFound", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			repo := NewPurchasedItemRepositoryImpl(tx)

			_, err := repo.GetByOrderItemId(-1)
			assert.NotNil(t, err)
			assert.Equal(t, "fatal error list of purchased item not available", err.Error())
		})
	})
}
