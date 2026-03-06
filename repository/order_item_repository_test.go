package repository

import (
	"cashier-api/helper/client"
	"cashier-api/helper/query"
	"cashier-api/model"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOrderItemRepository(t *testing.T) {
	var gormClient *gorm.DB = client.CreateGormClient()

	const STORE_ID = 1
	const TENANT_ID = 1
	t.Run("_PlaceOrderItem", func(t *testing.T) {
		t.Run("SuccessCase", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			warehouseRepo := NewWarehouseRepositoryImpl(tx)
			orderItemRepo := NewOrderItemRepositoryImpl(tx)
			storeStockRepo := NewStoreStockRepositoryImpl(tx)

			dummyItem := &model.Item{
				ItemName:  "Test PlaceOrderItem Success",
				Stocks:    100,
				StockType: model.StockTypeTracked,
				TenantId:  TENANT_ID,
			}

			items, err := warehouseRepo.CreateItem([]*model.Item{dummyItem})
			require.NoError(t, err)
			require.NotEmpty(t, items)

			item := items[0]

			err = storeStockRepo.TransferStockToStoreStock(
				5,
				item.ItemId,
				STORE_ID,
				TENANT_ID,
			)
			require.NoError(t, err)

			input := &model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  1,
				TotalAmount:    10000,
				DiscountAmount: 0,
				Subtotal:       10000,
				TenantId:       TENANT_ID,
				StoreId:        STORE_ID,
			}

			result, err := orderItemRepo.PlaceOrderItem(input)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotZero(t, result.Id)
		})

		t.Run("InvalidTotalQuantity", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  0,
				TotalAmount:    10000,
				DiscountAmount: 0,
				Subtotal:       10000,
				TenantId:       TENANT_ID,
				StoreId:        STORE_ID,
			}

			result, err := repo.PlaceOrderItem(input)

			require.Nil(t, result)
			require.Error(t, err)

			pgErr, ok := err.(*pgconn.PgError)
			require.True(t, ok)
			assert.Equal(t, "23514", pgErr.Code)
		})

		t.Run("InvalidTotalAmount", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  1,
				TotalAmount:    -1,
				DiscountAmount: 0,
				Subtotal:       10000,
				TenantId:       TENANT_ID,
				StoreId:        STORE_ID,
			}

			result, err := repo.PlaceOrderItem(input)

			require.Nil(t, result)
			require.Error(t, err)

			pgErr, ok := err.(*pgconn.PgError)
			require.True(t, ok)
			assert.Equal(t, "23514", pgErr.Code)
		})

		t.Run("InvalidDiscountAmount", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  1,
				TotalAmount:    10000,
				DiscountAmount: -1,
				Subtotal:       10000,
				TenantId:       TENANT_ID,
				StoreId:        STORE_ID,
			}

			result, err := repo.PlaceOrderItem(input)

			require.Nil(t, result)
			require.Error(t, err)

			pgErr, ok := err.(*pgconn.PgError)
			require.True(t, ok)
			assert.Equal(t, "23514", pgErr.Code)
		})

		t.Run("InvalidStoreForeignKey", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 9999,
				TotalQuantity:  1,
				TotalAmount:    9999,
				DiscountAmount: 0,
				Subtotal:       9999,
				TenantId:       TENANT_ID,
				StoreId:        0,
			}

			result, err := repo.PlaceOrderItem(input)

			require.Nil(t, result)
			require.Error(t, err)

			pgErr, ok := err.(*pgconn.PgError)
			require.True(t, ok)
			assert.Equal(t, "23503", pgErr.Code)
		})

		t.Run("InvalidTenantForeignKey", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 9999,
				TotalQuantity:  1,
				TotalAmount:    9999,
				DiscountAmount: 0,
				Subtotal:       9999,
				TenantId:       0,
				StoreId:        STORE_ID,
			}

			result, err := repo.PlaceOrderItem(input)

			require.Nil(t, result)
			require.Error(t, err)

			pgErr, ok := err.(*pgconn.PgError)
			require.True(t, ok)
			assert.Equal(t, "23503", pgErr.Code)
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("NormalQuery", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			dummyOrderItems := []*model.OrderItem{
				{PurchasedPrice: 10000, TotalQuantity: 1, TotalAmount: 10000, DiscountAmount: 0, Subtotal: 10000, TenantId: TENANT_ID, StoreId: STORE_ID},
				{PurchasedPrice: 20000, TotalQuantity: 2, TotalAmount: 40000, DiscountAmount: 0, Subtotal: 40000, TenantId: TENANT_ID, StoreId: STORE_ID},
				{PurchasedPrice: 30000, TotalQuantity: 3, TotalAmount: 90000, DiscountAmount: 0, Subtotal: 90000, TenantId: TENANT_ID, StoreId: STORE_ID},
				{PurchasedPrice: 40000, TotalQuantity: 4, TotalAmount: 100000, DiscountAmount: 60000, Subtotal: 160000, TenantId: TENANT_ID, StoreId: STORE_ID},
				{PurchasedPrice: 50000, TotalQuantity: 5, TotalAmount: 250000, DiscountAmount: 0, Subtotal: 250000, TenantId: TENANT_ID, StoreId: STORE_ID},
			}

			for _, item := range dummyOrderItems {
				result, err := orderItemRepo.PlaceOrderItem(item)
				require.Nil(t, err)
				require.NotZero(t, result.Id)
			}

			results, count, err := orderItemRepo.Get(TENANT_ID, 0, 5, 0, nil, nil)

			require.Nil(t, err)
			require.Equal(t, 5, count)
			require.Len(t, results, 5)

			for _, r := range results {
				assert.Equal(t, TENANT_ID, r.TenantId)
				assert.NotZero(t, r.TotalAmount)
			}
		})

		t.Run("DateSortByDesc", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

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
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
					CreatedAt:      dates[i],
				})
				require.Nil(t, err)
			}

			results, count, err := orderItemRepo.Get(
				TENANT_ID,
				0,
				5,
				0,
				[]*query.QueryFilter{
					{Column: query.CreatedAtColumn, Ascending: false},
				},
				nil,
			)

			require.Nil(t, err)
			require.Equal(t, 5, count)
			require.Len(t, results, 5)

			for i := 1; i < len(results); i++ {
				require.False(t, results[i].CreatedAt.After(results[i-1].CreatedAt))
			}
		})

		t.Run("SortByTotalAmountDesc", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			amounts := []int{10000, 40000, 90000, 100000, 250000}

			for _, amt := range amounts {
				_, err := orderItemRepo.PlaceOrderItem(&model.OrderItem{
					PurchasedPrice: amt,
					TotalQuantity:  1,
					TotalAmount:    amt,
					DiscountAmount: 0,
					Subtotal:       amt,
					TenantId:       TENANT_ID,
					StoreId:        STORE_ID,
				})
				require.Nil(t, err)
			}

			results, count, err := orderItemRepo.Get(
				TENANT_ID,
				0,
				5,
				0,
				[]*query.QueryFilter{
					{Column: query.TotalAmountColumn, Ascending: false},
				},
				nil,
			)

			require.Nil(t, err)
			require.Equal(t, 5, count)
			require.Len(t, results, 5)

			for i := 1; i < len(results); i++ {
				require.LessOrEqual(t, results[i].TotalAmount, results[i-1].TotalAmount)
			}
		})
	})

	t.Run("Transactions", func(t *testing.T) {
		t.Skip("DBMS relation too deep")
	})
	t.Run("FindById", func(t *testing.T) {
		t.Skip("DBMS relation too deep")
	})
	t.Run("GetSalesReport", func(t *testing.T) {
		t.Skip("DBMS relation too deep")
	})
}
