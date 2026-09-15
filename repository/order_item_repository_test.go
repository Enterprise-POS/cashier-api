package repository

import (
	"cashier-api/helper/client"
	"cashier-api/helper/query"
	"cashier-api/model"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// seedOrderItemTestDependencies creates a user, tenant, and store within the
// given transaction. All rows are rolled back automatically after each test.
func seedOrderItemTestDependencies(t *testing.T, tx *gorm.DB, email string, userName string) (tenantId int, storeId int) {
	t.Helper()

	user := &model.User{
		Name:     fmt.Sprintf("%s Test User", userName),
		Email:    email,
		Password: "password",
	}
	require.NoError(t, tx.Create(user).Error)
	require.NotZero(t, user.Id)

	tenant := &model.Tenant{
		Name:        fmt.Sprintf("%s Test Tenant", userName),
		OwnerUserId: user.Id,
		IsActive:    true,
	}
	require.NoError(t, tx.Create(tenant).Error)
	require.NotZero(t, tenant.Id)

	store := &model.Store{
		Name:     fmt.Sprintf("%s Test Store", userName),
		TenantId: tenant.Id,
		IsActive: true,
	}
	require.NoError(t, tx.Create(store).Error)
	require.NotZero(t, store.Id)

	return tenant.Id, store.Id
}

func TestOrderItemRepository(t *testing.T) {
	var gormClient *gorm.DB = client.CreateGormClient()

	t.Run("_PlaceOrderItem", func(t *testing.T) {
		t.Run("SuccessCase", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")

			warehouseRepo := NewWarehouseRepositoryImpl(tx)
			orderItemRepo := NewOrderItemRepositoryImpl(tx)
			storeStockRepo := NewStoreStockRepositoryImpl(tx)

			dummyItem := &model.Item{
				ItemName:  "Test PlaceOrderItem Success",
				Stocks:    100,
				StockType: model.StockTypeTracked,
				TenantId:  tenantId,
			}

			items, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
			assert.NoError(t, err)
			assert.NotEmpty(t, items)

			item := items[0]

			err = storeStockRepo.TransferStockToStoreStock(
				5,
				item.ItemId,
				storeId,
				tenantId,
			)
			assert.NoError(t, err)

			input := &model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  1,
				TotalAmount:    10000,
				DiscountAmount: 0,
				Subtotal:       10000,
				TenantId:       tenantId,
				StoreId:        storeId,
				PaymentType:    model.PaymentTypeCash,
				PaymentStatus:  model.PaymentStatusSuccess,
			}

			result, err := orderItemRepo.PlaceOrderItem(input)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotZero(t, result.Id)
		})

		t.Run("InvalidTotalQuantity", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")
			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  0, // Invalid
				TotalAmount:    10000,
				DiscountAmount: 0,
				Subtotal:       10000,
				TenantId:       tenantId,
				StoreId:        storeId,
				PaymentType:    model.PaymentTypeCash,
				TransactionId:  "",
				PaymentStatus:  model.PaymentStatusSuccess,
			}

			result, err := repo.PlaceOrderItem(input)

			assert.Nil(t, result)
			assert.Error(t, err)

			pgErr, ok := err.(*pgconn.PgError)
			assert.True(t, ok)
			assert.Equal(t, "23514", pgErr.Code)
		})

		t.Run("InvalidTotalAmount", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")
			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  1,
				TotalAmount:    -1, // Invalid
				DiscountAmount: 0,
				Subtotal:       10000,
				TenantId:       tenantId,
				StoreId:        storeId,
				PaymentType:    model.PaymentTypeCash,
				TransactionId:  "",
				PaymentStatus:  model.PaymentStatusSuccess,
			}

			result, err := repo.PlaceOrderItem(input)

			assert.Nil(t, result)
			assert.Error(t, err)

			pgErr, ok := err.(*pgconn.PgError)
			assert.True(t, ok)
			assert.Equal(t, "23514", pgErr.Code)
		})

		t.Run("InvalidDiscountAmount", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")
			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  1,
				TotalAmount:    10000,
				DiscountAmount: -1, // Invalid
				Subtotal:       10000,
				TenantId:       tenantId,
				StoreId:        storeId,
				PaymentType:    model.PaymentTypeCash,
				TransactionId:  "",
				PaymentStatus:  model.PaymentStatusSuccess,
			}

			result, err := repo.PlaceOrderItem(input)

			assert.Nil(t, result)
			assert.Error(t, err)

			pgErr, ok := err.(*pgconn.PgError)
			assert.True(t, ok)
			assert.Equal(t, "23514", pgErr.Code)
		})

		t.Run("InvalidStoreForeignKey", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")
			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 9999,
				TotalQuantity:  1,
				TotalAmount:    9999,
				DiscountAmount: 0,
				Subtotal:       9999,
				TenantId:       tenantId,
				StoreId:        0, // Invalid: FK violation expected
				PaymentType:    model.PaymentTypeCash,
				TransactionId:  "",
				PaymentStatus:  model.PaymentStatusSuccess,
			}

			result, err := repo.PlaceOrderItem(input)

			assert.Nil(t, result)
			assert.Error(t, err)

			pgErr, ok := err.(*pgconn.PgError)
			assert.True(t, ok)
			assert.Equal(t, "23503", pgErr.Code)
		})

		t.Run("InvalidTenantForeignKey", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			_, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")
			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 9999,
				TotalQuantity:  1,
				TotalAmount:    9999,
				DiscountAmount: 0,
				Subtotal:       9999,
				TenantId:       0, // Invalid: FK violation expected
				StoreId:        storeId,
				PaymentType:    model.PaymentTypeCash,
				TransactionId:  "",
				PaymentStatus:  model.PaymentStatusSuccess,
			}

			result, err := repo.PlaceOrderItem(input)

			assert.Nil(t, result)
			assert.Error(t, err)

			pgErr, ok := err.(*pgconn.PgError)
			assert.True(t, ok)
			assert.Equal(t, "23503", pgErr.Code)
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("NormalQuery", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")
			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			transactionId1 := fmt.Sprintf("TEST-%s", uuid.NewString())
			transactionId2 := fmt.Sprintf("TEST-%s", uuid.NewString())
			transactionId3 := fmt.Sprintf("TEST-%s", uuid.NewString())
			transactionId4 := fmt.Sprintf("TEST-%s", uuid.NewString())
			transactionId5 := fmt.Sprintf("TEST-%s", uuid.NewString())
			dummyOrderItems := []*model.OrderItem{
				{PurchasedPrice: 10000, TotalQuantity: 1, TotalAmount: 10000, DiscountAmount: 0, Subtotal: 10000, TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeCash, PaymentStatus: model.PaymentStatusSuccess, TransactionId: transactionId1},
				{PurchasedPrice: 20000, TotalQuantity: 2, TotalAmount: 40000, DiscountAmount: 0, Subtotal: 40000, TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeCash, PaymentStatus: model.PaymentStatusSuccess, TransactionId: transactionId2},
				{PurchasedPrice: 30000, TotalQuantity: 3, TotalAmount: 90000, DiscountAmount: 0, Subtotal: 90000, TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeCash, PaymentStatus: model.PaymentStatusSuccess, TransactionId: transactionId3},
				{PurchasedPrice: 40000, TotalQuantity: 4, TotalAmount: 100000, DiscountAmount: 60000, Subtotal: 160000, TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeCash, PaymentStatus: model.PaymentStatusSuccess, TransactionId: transactionId4},
				{PurchasedPrice: 50000, TotalQuantity: 5, TotalAmount: 250000, DiscountAmount: 0, Subtotal: 250000, TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeCash, PaymentStatus: model.PaymentStatusSuccess, TransactionId: transactionId5},
			}

			for _, item := range dummyOrderItems {
				result, err := orderItemRepo.PlaceOrderItem(item)
				assert.Nil(t, err)
				assert.NotZero(t, result.Id)
			}

			results, count, err := orderItemRepo.Get(tenantId, 0, 5, 0, nil, nil)

			assert.Nil(t, err)
			assert.Equal(t, 5, count)
			assert.Len(t, results, 5)

			for _, r := range results {
				assert.Equal(t, tenantId, r.TenantId)
				assert.NotZero(t, r.TotalAmount)
			}
		})

		t.Run("DateSortByDesc", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")
			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			base := time.Date(2025, 1, 10, 10, 0, 0, 0, time.UTC)

			dates := []time.Time{
				base,
				base.AddDate(0, 0, -1),
				base.AddDate(0, 0, -2),
				base.AddDate(0, 0, -3),
				base.AddDate(0, 0, -4),
			}

			for i := 0; i < 5; i++ {
				_, err := orderItemRepo.PlaceOrderItem(&model.OrderItem{
					PurchasedPrice: 100 * (i + 1),
					TotalQuantity:  1,
					TotalAmount:    100 * (i + 1),
					DiscountAmount: 0,
					Subtotal:       100 * (i + 1),
					TenantId:       tenantId,
					StoreId:        storeId,
					CreatedAt:      dates[i],
					PaymentType:    model.PaymentTypeCash,
					PaymentStatus:  model.PaymentStatusSuccess,
				})
				assert.Nil(t, err)
			}

			results, count, err := orderItemRepo.Get(
				tenantId,
				0,
				5,
				0,
				[]*query.QueryFilter{
					{Column: query.CreatedAtColumn, Ascending: false},
				},
				nil,
			)

			assert.Nil(t, err)
			assert.Equal(t, 5, count)
			assert.Len(t, results, 5)

			for i := 1; i < len(results); i++ {
				assert.False(t, results[i].CreatedAt.After(results[i-1].CreatedAt))
			}
		})

		t.Run("SortByTotalAmountDesc", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")
			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			amounts := []int{10000, 40000, 90000, 100000, 250000}

			for _, amt := range amounts {
				_, err := orderItemRepo.PlaceOrderItem(&model.OrderItem{
					PurchasedPrice: amt,
					TotalQuantity:  1,
					TotalAmount:    amt,
					DiscountAmount: 0,
					Subtotal:       amt,
					TenantId:       tenantId,
					StoreId:        storeId,
					PaymentType:    model.PaymentTypeCash,
					PaymentStatus:  model.PaymentStatusSuccess,
				})
				assert.Nil(t, err)
			}

			results, count, err := orderItemRepo.Get(
				tenantId,
				0,
				5,
				0,
				[]*query.QueryFilter{
					{Column: query.TotalAmountColumn, Ascending: false},
				},
				nil,
			)

			assert.Nil(t, err)
			assert.Equal(t, 5, count)
			assert.Len(t, results, 5)

			for i := 1; i < len(results); i++ {
				assert.LessOrEqual(t, results[i].TotalAmount, results[i-1].TotalAmount)
			}
		})
	})

	t.Run("Transactions", func(t *testing.T) {
		t.Skip("DBMS relation too deep")
	})

	t.Run("SetPaymentStatus", func(t *testing.T) {
		t.Run("Success_FromVariousStartingStatuses", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_setpaymentstatus_success@example.com", "Order Item")
			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			transactionId1 := fmt.Sprintf("TEST-%s", uuid.NewString())
			transactionId2 := fmt.Sprintf("TEST-%s", uuid.NewString())
			transactionId3 := fmt.Sprintf("TEST-%s", uuid.NewString())
			transactionId4 := fmt.Sprintf("TEST-%s", uuid.NewString())
			dummyOrderItems := []*model.OrderItem{
				{PurchasedPrice: 20000, TotalQuantity: 2, TotalAmount: 40000, DiscountAmount: 0, Subtotal: 40000, TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeQRIS, PaymentStatus: model.PaymentStatusPending, TransactionId: transactionId1},
				{PurchasedPrice: 30000, TotalQuantity: 3, TotalAmount: 90000, DiscountAmount: 0, Subtotal: 90000, TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeQRIS, PaymentStatus: model.PaymentStatusCancelled, TransactionId: transactionId2},
				{PurchasedPrice: 40000, TotalQuantity: 4, TotalAmount: 100000, DiscountAmount: 60000, Subtotal: 160000, TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeQRIS, PaymentStatus: model.PaymentStatusExpired, TransactionId: transactionId3},
				{PurchasedPrice: 50000, TotalQuantity: 5, TotalAmount: 250000, DiscountAmount: 0, Subtotal: 250000, TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeQRIS, PaymentStatus: model.PaymentStatusRefunded, TransactionId: transactionId4},
			}

			resultsOrderItems := make([]*model.OrderItem, 0)
			for _, item := range dummyOrderItems {
				result, err := orderItemRepo.PlaceOrderItem(item)
				assert.Nil(t, err)
				assert.NotZero(t, result.Id)
				assert.Equal(t, item.PaymentStatus, result.PaymentStatus)
				assert.Equal(t, item.PaymentType, result.PaymentType)
				resultsOrderItems = append(resultsOrderItems, result)
			}

			for _, item := range resultsOrderItems {
				err := orderItemRepo.SetPaymentStatus(item.Id, item.TransactionId, model.PaymentStatusSuccess)
				assert.NoError(t, err)

				var checkOrderItem model.OrderItem
				err = tx.First(&checkOrderItem, item.Id).Error
				require.NoError(t, err)
				assert.Equal(t, model.PaymentStatusSuccess, checkOrderItem.PaymentStatus)
			}
		})

		t.Run("NotFound_ReturnsErrRecordNotFound", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			err := orderItemRepo.SetPaymentStatus(999999999, "NONEXISTENT-TX-ID", model.PaymentStatusSuccess)

			assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		})

		t.Run("AmbiguousMatch_RollsBackAndErrors", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_setpaymentstatus_ambiguous@example.com", "Order Item")
			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			transactionIdA := fmt.Sprintf("TEST-%s", uuid.NewString())
			transactionIdB := fmt.Sprintf("TEST-%s", uuid.NewString())

			itemA, err := orderItemRepo.PlaceOrderItem(&model.OrderItem{
				PurchasedPrice: 10000, TotalQuantity: 1, TotalAmount: 10000, Subtotal: 10000,
				TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeQRIS,
				PaymentStatus: model.PaymentStatusPending, TransactionId: transactionIdA,
			})
			require.NoError(t, err)

			itemB, err := orderItemRepo.PlaceOrderItem(&model.OrderItem{
				PurchasedPrice: 15000, TotalQuantity: 1, TotalAmount: 15000, Subtotal: 15000,
				TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeQRIS,
				PaymentStatus: model.PaymentStatusPending, TransactionId: transactionIdB,
			})
			require.NoError(t, err)

			// transaction_id = itemA's txId OR id = itemB's id -> matches BOTH rows.
			err = orderItemRepo.SetPaymentStatus(itemB.Id, transactionIdA, model.PaymentStatusSuccess)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "matched")

			// Rolled back: neither row should have changed.
			var checkA, checkB model.OrderItem
			require.NoError(t, tx.First(&checkA, itemA.Id).Error)
			require.NoError(t, tx.First(&checkB, itemB.Id).Error)
			assert.Equal(t, model.PaymentStatusPending, checkA.PaymentStatus)
			assert.Equal(t, model.PaymentStatusPending, checkB.PaymentStatus)
		})
	})

	t.Run("SyncDataStock", func(t *testing.T) {
		t.Run("Success_DecrementsTrackedItemsOnly_AggregatesDuplicateRows", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_syncstock_success@example.com", "Order Item")
			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			trackedItemId := seedWarehouseItemAndStock(t, tx, tenantId, storeId, "TRACKED", 10)
			untrackedItemId := seedWarehouseItemAndStock(t, tx, tenantId, storeId, "UNLIMITED", 10)

			order, err := orderItemRepo.PlaceOrderItem(&model.OrderItem{
				PurchasedPrice: 60000, TotalQuantity: 4, TotalAmount: 65000, Subtotal: 65000,
				TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeQRIS,
				PaymentStatus: model.PaymentStatusSuccess, TransactionId: fmt.Sprintf("TEST-%s", uuid.NewString()),
			})
			require.NoError(t, err)

			// Two rows for the same TRACKED -> proves SUM/GROUP BY aggregation.
			// One row for the UNLIMITED -> proves it's excluded from the decrement.
			require.NoError(t, tx.Exec(`
			INSERT INTO purchased_item_list (order_item_id, item_id, quantity, store_price_snapshot, base_price_snapshot, total_amount, item_name_snapshot, discount_amount)
			VALUES (?, ?, 2, 20000, 10000, 40000, 'TRACKED', 0),
			       (?, ?, 1, 20000, 10000, 20000, 'TRACKED', 0),
			       (?, ?, 1,  5000,  2000,  5000, 'UNLIMITED', 0)
		`, order.Id, trackedItemId, order.Id, trackedItemId, order.Id, untrackedItemId).Error)

			err = orderItemRepo.SyncDataStock(order.Id)
			require.NoError(t, err)

			assert.Equal(t, 7, getStoreStockQty(t, tx, tenantId, storeId, trackedItemId))    // 10 - (2+1)
			assert.Equal(t, 10, getStoreStockQty(t, tx, tenantId, storeId, untrackedItemId)) // untouched

			var reloaded model.OrderItem
			require.NoError(t, tx.First(&reloaded, order.Id).Error)
			assert.True(t, reloaded.IsDataStockSync)
		})

		t.Run("Idempotent_SecondCallDoesNotDecrementAgain", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_syncstock_idempotent@example.com", "Order Item")
			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			itemId := seedWarehouseItemAndStock(t, tx, tenantId, storeId, "TRACKED", 10)

			order, err := orderItemRepo.PlaceOrderItem(&model.OrderItem{
				PurchasedPrice: 20000, TotalQuantity: 3, TotalAmount: 20000, Subtotal: 20000,
				TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeQRIS,
				PaymentStatus: model.PaymentStatusSuccess, TransactionId: fmt.Sprintf("TEST-%s", uuid.NewString()),
			})
			require.NoError(t, err)

			require.NoError(t, tx.Exec(`
			INSERT INTO purchased_item_list (order_item_id, item_id, quantity, store_price_snapshot, base_price_snapshot, total_amount, item_name_snapshot, discount_amount)
			VALUES (?, ?, 3, 20000, 10000, 20000, 'TRACKED', 0)
		`, order.Id, itemId).Error)

			require.NoError(t, orderItemRepo.SyncDataStock(order.Id))
			assert.Equal(t, 7, getStoreStockQty(t, tx, tenantId, storeId, itemId))

			// Second call must be a no-op.
			require.NoError(t, orderItemRepo.SyncDataStock(order.Id))
			assert.Equal(t, 7, getStoreStockQty(t, tx, tenantId, storeId, itemId))
		})

		t.Run("UntrackedItemsOnly_MarksSyncedWithoutDecrementing", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_syncstock_untracked@example.com", "Order Item")
			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			itemId := seedWarehouseItemAndStock(t, tx, tenantId, storeId, "UNLIMITED", 10)

			order, err := orderItemRepo.PlaceOrderItem(&model.OrderItem{
				PurchasedPrice: 5000, TotalQuantity: 1, TotalAmount: 5000, Subtotal: 5000,
				TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeQRIS,
				PaymentStatus: model.PaymentStatusSuccess, TransactionId: fmt.Sprintf("TEST-%s", uuid.NewString()),
			})
			require.NoError(t, err)

			require.NoError(t, tx.Exec(`
			INSERT INTO purchased_item_list (order_item_id, item_id, quantity, store_price_snapshot, base_price_snapshot, total_amount, item_name_snapshot, discount_amount)
			VALUES (?, ?, 1, 5000, 2000, 5000, 'UNLIMITED', 0)
		`, order.Id, itemId).Error)

			err = orderItemRepo.SyncDataStock(order.Id)
			require.NoError(t, err)

			assert.Equal(t, 10, getStoreStockQty(t, tx, tenantId, storeId, itemId))

			var reloaded model.OrderItem
			require.NoError(t, tx.First(&reloaded, order.Id).Error)
			assert.True(t, reloaded.IsDataStockSync)
		})

		t.Run("PaymentNotSuccessful_ReturnsErrorWithoutDecrementing", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_syncstock_unpaid@example.com", "Order Item")
			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			itemId := seedWarehouseItemAndStock(t, tx, tenantId, storeId, "TRACKED", 10)

			order, err := orderItemRepo.PlaceOrderItem(&model.OrderItem{
				PurchasedPrice: 20000, TotalQuantity: 2, TotalAmount: 20000, Subtotal: 20000,
				TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeQRIS,
				PaymentStatus: model.PaymentStatusPending, TransactionId: fmt.Sprintf("TEST-%s", uuid.NewString()),
			})
			require.NoError(t, err)

			require.NoError(t, tx.Exec(`
			INSERT INTO purchased_item_list (order_item_id, item_id, quantity, store_price_snapshot, base_price_snapshot, total_amount, item_name_snapshot, discount_amount)
			VALUES (?, ?, 2, 20000, 10000, 20000, 'TRACKED', 0)
		`, order.Id, itemId).Error)

			err = orderItemRepo.SyncDataStock(order.Id)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "refusing to decrement stock")
			assert.Equal(t, 10, getStoreStockQty(t, tx, tenantId, storeId, itemId))

			var reloaded model.OrderItem
			require.NoError(t, tx.First(&reloaded, order.Id).Error)
			assert.False(t, reloaded.IsDataStockSync)
		})

		t.Run("NoMatchingStoreStockRow_ReturnsError", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_syncstock_missingstock@example.com", "Order Item")
			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			// Warehouse row exists (TRACKED), but no matching store_stock row for this store.
			itemId := seedWarehouseItemOnly(t, tx, tenantId, "TRACKED")

			order, err := orderItemRepo.PlaceOrderItem(&model.OrderItem{
				PurchasedPrice: 20000, TotalQuantity: 1, TotalAmount: 20000, Subtotal: 20000,
				TenantId: tenantId, StoreId: storeId, PaymentType: model.PaymentTypeQRIS,
				PaymentStatus: model.PaymentStatusSuccess, TransactionId: fmt.Sprintf("TEST-%s", uuid.NewString()),
			})
			require.NoError(t, err)

			require.NoError(t, tx.Exec(`
			INSERT INTO purchased_item_list (order_item_id, item_id, quantity, store_price_snapshot, base_price_snapshot, total_amount, item_name_snapshot, discount_amount)
			VALUES (?, ?, 1, 20000, 10000, 20000, 'TRACKED', 0)
		`, order.Id, itemId).Error)

			err = orderItemRepo.SyncDataStock(order.Id)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "no store_stock row")

			var reloaded model.OrderItem
			require.NoError(t, tx.First(&reloaded, order.Id).Error)
			assert.False(t, reloaded.IsDataStockSync) // whole transaction rolled back
		})

		t.Run("NotFound_ReturnsErrRecordNotFound", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			err := orderItemRepo.SyncDataStock(999999999)

			assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		})
	})

	t.Run("FindById", func(t *testing.T) {
		t.Skip("DBMS relation too deep")
	})
	t.Run("GetSalesReport", func(t *testing.T) {
		t.Skip("DBMS relation too deep")
	})

	t.Run("DeleteInvoice", func(t *testing.T) {
		t.Run("SuccessCase", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")
			repo := NewOrderItemRepositoryImpl(tx)

			created, err := repo.PlaceOrderItem(&model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  1,
				TotalAmount:    10000,
				DiscountAmount: 0,
				Subtotal:       10000,
				TenantId:       tenantId,
				StoreId:        storeId,
				PaymentType:    model.PaymentTypeCash,
				PaymentStatus:  model.PaymentStatusSuccess,
				TransactionId:  "",
			})
			require.NoError(t, err)
			require.NotZero(t, created.Id)

			err = repo.DeleteInvoice(created.Id, tenantId)
			assert.NoError(t, err)

			// Verify soft deleted — row still exists but deleted_at is set
			var deleted model.OrderItem
			err = tx.Unscoped().First(&deleted, created.Id).Error
			assert.NoError(t, err)
			assert.True(t, deleted.DeletedAt.Valid)

			// Verify excluded from normal queries
			results, count, err := repo.Get(tenantId, 0, 10, 0, nil, nil)
			assert.NoError(t, err)
			assert.Equal(t, 0, count)
			assert.Len(t, results, 0)
		})

		t.Run("NotFound", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")
			repo := NewOrderItemRepositoryImpl(tx)

			// id: 1 should not exist
			err := repo.DeleteInvoice(1, tenantId)

			// RowsAffected = 0 now returns an error since we added the check
			assert.Error(t, err)
		})

		t.Run("WrongTenant", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx, "orderitem_test@example.com", "Order Item")
			repo := NewOrderItemRepositoryImpl(tx)

			created, err := repo.PlaceOrderItem(&model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  1,
				TotalAmount:    10000,
				DiscountAmount: 0,
				Subtotal:       10000,
				TenantId:       tenantId,
				StoreId:        storeId,
				PaymentType:    model.PaymentTypeCash,
				PaymentStatus:  model.PaymentStatusSuccess,
				TransactionId:  "",
			})
			require.NoError(t, err)
			require.NotZero(t, created.Id)

			err = repo.DeleteInvoice(created.Id, tenantId+1)
			assert.Error(t, err)

			// Verify the record is untouched
			var untouched model.OrderItem
			err = tx.Unscoped().First(&untouched, created.Id).Error
			assert.NoError(t, err)
			assert.False(t, untouched.DeletedAt.Valid)
		})
	})
}

// seedWarehouseItemAndStock inserts a warehouse row and a matching store_stock
// row for the given tenant/store, returning the shared item_id.
//
// ASSUMPTION: item_id is warehouse's own generated primary key, and
// store_stock.item_id references it directly with no separate "item" master
// table — inferred from the PL/pgSQL function's join `w.item_id = ss.item_id`.
// Adjust if your schema has an intermediate items table.
func seedWarehouseItemAndStock(t *testing.T, tx *gorm.DB, tenantId, storeId int, stockType model.StockType, startingStock int) int {
	t.Helper()

	var itemId int
	require.NoError(t, tx.Raw(`
		INSERT INTO warehouse (tenant_id, item_name, base_price, stock_type, stocks)
		VALUES (?, ?, ?, ?, ?)
		RETURNING item_id
	`, tenantId, "Test Item", 10000, stockType, startingStock).Scan(&itemId).Error)

	require.NoError(t, tx.Exec(`
		INSERT INTO store_stock (tenant_id, store_id, item_id, price, stocks)
		VALUES (?, ?, ?, ?, ?)
	`, tenantId, storeId, itemId, 20000, startingStock).Error)

	return itemId
}

// seedWarehouseItemOnly inserts a warehouse row with NO matching store_stock
// row, for testing the "missing stock row" error path.
func seedWarehouseItemOnly(t *testing.T, tx *gorm.DB, tenantId int, stockType model.StockType) int {
	t.Helper()

	var itemId int
	require.NoError(t, tx.Raw(`
		INSERT INTO warehouse (tenant_id, item_name, base_price, stock_type, stocks)
		VALUES (?, ?, ?, ?, ?)
		RETURNING item_id
	`, tenantId, "Test Item No Stock", 10000, stockType, 0).Scan(&itemId).Error)

	return itemId
}

// getStoreStockQty reads the current stocks value directly, bypassing the
// GORM model so the test doesn't depend on exact struct field mapping.
func getStoreStockQty(t *testing.T, tx *gorm.DB, tenantId, storeId, itemId int) int {
	t.Helper()

	var stocks int
	require.NoError(t, tx.Raw(`
		SELECT stocks FROM store_stock
		WHERE tenant_id = ? AND store_id = ? AND item_id = ?
	`, tenantId, storeId, itemId).Scan(&stocks).Error)

	return stocks
}
