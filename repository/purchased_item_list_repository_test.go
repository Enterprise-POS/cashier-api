package repository

import (
	"cashier-api/helper/client"
	"cashier-api/helper/query"
	"cashier-api/model"
	"testing"
	"time"

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

	testEmail := "purchased_item_list@gmail.com"
	testUserName := "purchased_item_list"

	// seedPurchasedItemDependencies creates user→tenant→store→item(apple)→item(peach)
	// and an order_item within the given transaction.
	// All rows roll back automatically after each test.
	seedPurchasedItemDependencies := func(t *testing.T, tx *gorm.DB) (appleId int, peachId int, orderItemId int) {
		t.Helper()

		tenantId, storeId := seedOrderItemTestDependencies(t, tx, testEmail, testUserName)

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
			PaymentType:    model.PaymentTypeCash,
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

	t.Run("PurchasedItemListLogs", func(t *testing.T) {
		// seedListLogsScenario seeds tenant/store/apple/peach/order_item and inserts
		// 3 purchased_item rows (2 apple, 1 peach) tied to that order_item.
		seedListLogsScenario := func(t *testing.T, tx *gorm.DB) (tenantId, storeId, appleId, peachId, orderItemId int) {
			t.Helper()

			appleId, peachId, orderItemId = seedPurchasedItemDependencies(t, tx)

			var orderItem model.OrderItem
			require.NoError(t, tx.First(&orderItem, orderItemId).Error)
			tenantId = orderItem.TenantId
			storeId = orderItem.StoreId

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
					ItemId:             appleId,
					OrderItemId:        orderItemId,
					Quantity:           1,
					StorePriceSnapshot: APPLE_PRICE,
					DiscountAmount:     0,
					TotalAmount:        1 * APPLE_PRICE,
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
			require.NoError(t, err)

			return tenantId, storeId, appleId, peachId, orderItemId
		}

		t.Run("NoFilters", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _, _, _, _ := seedListLogsScenario(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			rows, total, err := repo.PurchasedItemListLogs(
				tenantId, 0, nil, 10, 0, nil, nil,
			)
			assert.NoError(t, err)
			assert.Equal(t, 3, total)
			assert.Equal(t, 3, len(rows))
		})

		t.Run("FilterByStoreId", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId, _, _, _ := seedListLogsScenario(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			rows, total, err := repo.PurchasedItemListLogs(
				tenantId, storeId, nil, 10, 0, nil, nil,
			)
			assert.NoError(t, err)
			assert.Equal(t, 3, total)
			assert.Equal(t, 3, len(rows))

			// A non-existent store should return nothing.
			rows, total, err = repo.PurchasedItemListLogs(
				tenantId, storeId+999999, nil, 10, 0, nil, nil,
			)
			assert.NoError(t, err)
			assert.Equal(t, 0, total)
			assert.Equal(t, 0, len(rows))
		})

		t.Run("FilterByItemIds", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _, appleId, _, _ := seedListLogsScenario(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			rows, total, err := repo.PurchasedItemListLogs(
				tenantId, 0, []int{appleId}, 10, 0, nil, nil,
			)
			assert.NoError(t, err)
			assert.Equal(t, 2, total)
			assert.Equal(t, 2, len(rows))
			for _, row := range rows {
				assert.Equal(t, appleId, row.ItemId)
			}
		})

		t.Run("Pagination", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _, _, _, _ := seedListLogsScenario(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			// Page 0, limit 2 -> 2 rows, total still 3.
			rows, total, err := repo.PurchasedItemListLogs(
				tenantId, 0, nil, 2, 0, nil, nil,
			)
			assert.NoError(t, err)
			assert.Equal(t, 3, total)
			assert.Equal(t, 2, len(rows))

			// Page 1, limit 2 -> remaining 1 row.
			rows, total, err = repo.PurchasedItemListLogs(
				tenantId, 0, nil, 2, 1, nil, nil,
			)
			assert.NoError(t, err)
			assert.Equal(t, 3, total)
			assert.Equal(t, 1, len(rows))
		})

		t.Run("SortByTotalAmountAscending", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _, _, _, _ := seedListLogsScenario(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			filters := []query.QueryFilter{
				{Column: query.TotalAmountColumn, Ascending: true},
			}

			rows, total, err := repo.PurchasedItemListLogs(
				tenantId, 0, nil, 10, 0, nil, filters,
			)
			assert.NoError(t, err)
			assert.Equal(t, 3, total)
			require.Equal(t, 3, len(rows))

			for i := 1; i < len(rows); i++ {
				assert.LessOrEqual(t, rows[i-1].TotalAmount, rows[i].TotalAmount)
			}
		})

		t.Run("SortByQuantityDescending", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _, _, _, _ := seedListLogsScenario(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			filters := []query.QueryFilter{
				{Column: query.Quantity, Ascending: false},
			}

			rows, total, err := repo.PurchasedItemListLogs(
				tenantId, 0, nil, 10, 0, nil, filters,
			)
			assert.NoError(t, err)
			assert.Equal(t, 3, total)
			require.Equal(t, 3, len(rows))

			for i := 1; i < len(rows); i++ {
				assert.GreaterOrEqual(t, rows[i-1].Quantity, rows[i].Quantity)
			}
		})

		t.Run("InvalidFilterColumn", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _, _, _, _ := seedListLogsScenario(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			filters := []query.QueryFilter{
				{Column: "not_a_real_column", Ascending: true},
			}

			rows, total, err := repo.PurchasedItemListLogs(
				tenantId, 0, nil, 10, 0, nil, filters,
			)
			assert.Nil(t, rows)
			assert.Zero(t, total)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "[FATAL ERROR]")
		})

		t.Run("DateFilterRange", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _, _, _, orderItemId := seedListLogsScenario(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			var orderItem model.OrderItem
			require.NoError(t, tx.First(&orderItem, orderItemId).Error)

			start := orderItem.CreatedAt.Add(-time.Hour).Unix()
			end := orderItem.CreatedAt.Add(time.Hour).Unix()

			dateFilter := &query.DateFilter{
				StartDate: &start,
				EndDate:   &end,
			}

			rows, total, err := repo.PurchasedItemListLogs(
				tenantId, 0, nil, 10, 0, dateFilter, nil,
			)
			assert.NoError(t, err)
			assert.Equal(t, 3, total)
			assert.Equal(t, 3, len(rows))

			// A window entirely in the past should return nothing.
			farPastStart := orderItem.CreatedAt.Add(-48 * time.Hour).Unix()
			farPastEnd := orderItem.CreatedAt.Add(-24 * time.Hour).Unix()
			pastFilter := &query.DateFilter{
				StartDate: &farPastStart,
				EndDate:   &farPastEnd,
			}

			rows, total, err = repo.PurchasedItemListLogs(
				tenantId, 0, nil, 10, 0, pastFilter, nil,
			)
			assert.NoError(t, err)
			assert.Equal(t, 0, total)
			assert.Equal(t, 0, len(rows))
		})

		t.Run("NoMatchingTenant", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			seedListLogsScenario(t, tx)
			repo := NewPurchasedItemRepositoryImpl(tx)

			rows, total, err := repo.PurchasedItemListLogs(
				-1, 0, nil, 10, 0, nil, nil,
			)
			assert.NoError(t, err)
			assert.Equal(t, 0, total)
			assert.Equal(t, 0, len(rows))
		})
	})
}
