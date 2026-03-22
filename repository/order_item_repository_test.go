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

// seedOrderItemTestDependencies creates a user, tenant, and store within the
// given transaction. All rows are rolled back automatically after each test.
func seedOrderItemTestDependencies(t *testing.T, tx *gorm.DB) (tenantId int, storeId int) {
	t.Helper()

	user := &model.User{
		Name:     "Order Item Test User",
		Email:    "orderitem_test@example.com",
		Password: "password",
	}
	require.NoError(t, tx.Create(user).Error)
	require.NotZero(t, user.Id)

	tenant := &model.Tenant{
		Name:        "Order Item Test Tenant",
		OwnerUserId: user.Id,
		IsActive:    true,
	}
	require.NoError(t, tx.Create(tenant).Error)
	require.NotZero(t, tenant.Id)

	store := &model.Store{
		Name:     "Order Item Test Store",
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

			tenantId, storeId := seedOrderItemTestDependencies(t, tx)

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
			}

			result, err := orderItemRepo.PlaceOrderItem(input)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotZero(t, result.Id)
		})

		t.Run("InvalidTotalQuantity", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, storeId := seedOrderItemTestDependencies(t, tx)
			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  0, // Invalid
				TotalAmount:    10000,
				DiscountAmount: 0,
				Subtotal:       10000,
				TenantId:       tenantId,
				StoreId:        storeId,
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

			tenantId, storeId := seedOrderItemTestDependencies(t, tx)
			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  1,
				TotalAmount:    -1, // Invalid
				DiscountAmount: 0,
				Subtotal:       10000,
				TenantId:       tenantId,
				StoreId:        storeId,
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

			tenantId, storeId := seedOrderItemTestDependencies(t, tx)
			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 10000,
				TotalQuantity:  1,
				TotalAmount:    10000,
				DiscountAmount: -1, // Invalid
				Subtotal:       10000,
				TenantId:       tenantId,
				StoreId:        storeId,
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

			tenantId, _ := seedOrderItemTestDependencies(t, tx)
			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 9999,
				TotalQuantity:  1,
				TotalAmount:    9999,
				DiscountAmount: 0,
				Subtotal:       9999,
				TenantId:       tenantId,
				StoreId:        0, // Invalid: FK violation expected
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

			_, storeId := seedOrderItemTestDependencies(t, tx)
			repo := NewOrderItemRepositoryImpl(tx)

			input := &model.OrderItem{
				PurchasedPrice: 9999,
				TotalQuantity:  1,
				TotalAmount:    9999,
				DiscountAmount: 0,
				Subtotal:       9999,
				TenantId:       0, // Invalid: FK violation expected
				StoreId:        storeId,
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

			tenantId, storeId := seedOrderItemTestDependencies(t, tx)
			orderItemRepo := NewOrderItemRepositoryImpl(tx)

			dummyOrderItems := []*model.OrderItem{
				{PurchasedPrice: 10000, TotalQuantity: 1, TotalAmount: 10000, DiscountAmount: 0, Subtotal: 10000, TenantId: tenantId, StoreId: storeId},
				{PurchasedPrice: 20000, TotalQuantity: 2, TotalAmount: 40000, DiscountAmount: 0, Subtotal: 40000, TenantId: tenantId, StoreId: storeId},
				{PurchasedPrice: 30000, TotalQuantity: 3, TotalAmount: 90000, DiscountAmount: 0, Subtotal: 90000, TenantId: tenantId, StoreId: storeId},
				{PurchasedPrice: 40000, TotalQuantity: 4, TotalAmount: 100000, DiscountAmount: 60000, Subtotal: 160000, TenantId: tenantId, StoreId: storeId},
				{PurchasedPrice: 50000, TotalQuantity: 5, TotalAmount: 250000, DiscountAmount: 0, Subtotal: 250000, TenantId: tenantId, StoreId: storeId},
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

			tenantId, storeId := seedOrderItemTestDependencies(t, tx)
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

			tenantId, storeId := seedOrderItemTestDependencies(t, tx)
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
	t.Run("FindById", func(t *testing.T) {
		t.Skip("DBMS relation too deep")
	})
	t.Run("GetSalesReport", func(t *testing.T) {
		t.Skip("DBMS relation too deep")
	})
}
