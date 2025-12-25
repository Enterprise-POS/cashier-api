package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderItemServiceImpl(t *testing.T) {
	now := time.Now()
	const USER_ID = 1
	const TENANT_ID = 1
	const STORE_ID = 1
	const LIMIT = 10
	const PAGE = 1

	t.Run("Get", func(t *testing.T) {
		t.Run("NormalGet", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)
			// filters := []*query.QueryFilter{}
			// // dateFilters := nil
			// orderItemRepo.Mock = &mock.Mock{}
			// orderItemRepo.Mock.On("Get", TENANT_ID, STORE_ID, LIMIT, PAGE, filters, nil)
			filters := []*query.QueryFilter{}
			expectedItems := []*model.OrderItem{
				{
					Id:             1,
					PurchasedPrice: 1000,
					TotalQuantity:  1,
					TotalAmount:    1000,
					DiscountAmount: 0,
					Subtotal:       1000,
					CreatedAt:      &now,
					StoreId:        STORE_ID,
					TenantId:       TENANT_ID,
				},
			}
			expectedCount := 1

			// Mock expects page-1 (0-based indexing)
			orderItemRepo.Mock.On("Get", TENANT_ID, STORE_ID, LIMIT, 0, filters, (*query.DateFilter)(nil)).
				Return(expectedItems, expectedCount, nil)

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, PAGE, filters, nil)

			assert.NoError(t, err)
			assert.Equal(t, expectedCount, count)
			assert.Equal(t, expectedItems, orderItems)
		})

		t.Run("TenantOrStoreIdIsNotProvided", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)
			// Invalid tenant id
			orderItems, count, err := orderItemService.Get(0, STORE_ID, LIMIT, PAGE, nil, nil)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)

			// Invalid store id
			orderItems, count, err = orderItemService.Get(TENANT_ID, 0, LIMIT, PAGE, nil, nil)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("InvalidLimitAndPageParams", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			// Test invalid limit (0)
			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, 0, PAGE, nil, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Limit could not less then 1")
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)

			// Test invalid page (0)
			orderItems, count, err = orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, 0, nil, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "page could not less then 1")
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)

			// Negative limit
			orderItems, count, err = orderItemService.Get(TENANT_ID, STORE_ID, -5, PAGE, nil, nil)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)

			// negative page
			orderItems, count, err = orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, -5, nil, nil)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("InvalidPage_Zero", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, 0, nil, nil)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "page could not less then 1")
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("InvalidPage_Negative", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, -1, nil, nil)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "page could not less then 1")
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("DateFilter_StartDateAfterEndDate", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			startDate := int64(1700000000)
			endDate := int64(1600000000)
			dateFilter := &query.DateFilter{
				StartDate: &startDate,
				EndDate:   &endDate,
			}

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, PAGE, nil, dateFilter)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Start date")
			assert.Contains(t, err.Error(), "cannot be after end date")
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("DateFilter_NegativeStartDate", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			startDate := int64(-1000)
			dateFilter := &query.DateFilter{
				StartDate: &startDate,
			}

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, PAGE, nil, dateFilter)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Invalid start date timestamp")
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("DateFilter_NegativeEndDate", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			endDate := int64(-1000)
			dateFilter := &query.DateFilter{
				EndDate: &endDate,
			}

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, PAGE, nil, dateFilter)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Invalid emd date timestamp")
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("DateFilter_StartDateTooFarInFuture", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			startDate := int64(5000000000) // 2100+
			dateFilter := &query.DateFilter{
				StartDate: &startDate,
			}

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, PAGE, nil, dateFilter)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Start date is too far in the future")
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("DateFilter_EndDateTooFarInFuture", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			endDate := int64(5000000000) // Way beyond 2100
			dateFilter := &query.DateFilter{
				EndDate: &endDate,
			}

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, PAGE, nil, dateFilter)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "End date is too far in the future")
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("DateFilter_ValidDateRange", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			startDate := int64(1600000000)
			endDate := int64(1700000000)
			dateFilter := &query.DateFilter{
				StartDate: &startDate,
				EndDate:   &endDate,
			}

			expectedItems := []*model.OrderItem{}
			expectedCount := 3

			orderItemRepo.Mock.On("Get", TENANT_ID, STORE_ID, LIMIT, 0, ([]*query.QueryFilter)(nil), dateFilter).
				Return(expectedItems, expectedCount, nil)

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, PAGE, nil, dateFilter)

			assert.NoError(t, err)
			assert.Equal(t, expectedCount, count)
			assert.Equal(t, expectedItems, orderItems)
			orderItemRepo.Mock.AssertExpectations(t)
		})

		t.Run("RepositoryReturnsError", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			expectedError := errors.New("database connection failed")

			orderItemRepo.Mock.On("Get", TENANT_ID, STORE_ID, LIMIT, 0, ([]*query.QueryFilter)(nil), (*query.DateFilter)(nil)).
				Return(([]*model.OrderItem)(nil), 0, expectedError)

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, PAGE, nil, nil)

			assert.Error(t, err)
			assert.Equal(t, expectedError, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
			orderItemRepo.Mock.AssertExpectations(t)
		})
	})

	t.Run("FindById", func(t *testing.T) {
		t.Run("NormalFindById", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			expectedOrderItem := &model.OrderItem{
				Id:             1,
				DiscountAmount: 0,
				Subtotal:       10000,
				TotalAmount:    10000,
				TotalQuantity:  1,
				PurchasedPrice: 10000,
				StoreId:        STORE_ID,
				TenantId:       TENANT_ID,
				CreatedAt:      &now,
			}

			expectedPurchasedList := []*model.PurchasedItem{
				{
					Id:             1,
					TotalAmount:    10000,
					Quantity:       1,
					PurchasedPrice: 10000,
					DiscountAmount: 0,
					ItemId:         1,
					OrderItemId:    expectedOrderItem.Id, // Connect with orderItem id
					CreatedAt:      &now,
				},
				{
					Id:             2,
					TotalAmount:    5000,
					Quantity:       1,
					PurchasedPrice: 5000,
					DiscountAmount: 0,
					ItemId:         2,
					OrderItemId:    expectedOrderItem.Id, // Connect with orderItem id
					CreatedAt:      &now,
				},
			}

			// Page minus by 1 because page will be 0 index from repository
			orderItemRepo.Mock.On("FindById", expectedOrderItem.Id, TENANT_ID).Return(expectedOrderItem, expectedPurchasedList, nil)
			orderItem, purchasedItemList, err := orderItemService.FindById(expectedOrderItem.Id, TENANT_ID)
			assert.NoError(t, err)
			assert.Len(t, purchasedItemList, 2)
			assert.NotNil(t, orderItem)
			for i, purchasedItem := range purchasedItemList {
				assert.Equal(t, expectedPurchasedList[i].Id, purchasedItem.Id)
				assert.Equal(t, expectedPurchasedList[i].ItemId, purchasedItem.ItemId)
				assert.Equal(t, expectedPurchasedList[i].OrderItemId, purchasedItem.OrderItemId)
			}
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo)

			const (
				TENANT_ID     = 1
				ORDER_ITEM_ID = 1
				LIMIT         = 10
				PAGE          = 1
			)

			// tenant id
			orderItem, purchasedItemList, err := orderItemService.FindById(ORDER_ITEM_ID, 0)
			assert.Error(t, err)
			assert.Nil(t, orderItem)
			assert.Nil(t, purchasedItemList)

			// order item id
			orderItem, purchasedItemList, err = orderItemService.FindById(0, TENANT_ID)
			assert.Error(t, err)
			assert.Nil(t, orderItem)
			assert.Nil(t, purchasedItemList)
		})
	})

	t.Run("Transactions", func(t *testing.T) {
		orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
		orderItemService := NewOrderItemServiceImpl(orderItemRepo)

		t.Run("NormalTransactions", func(t *testing.T) {
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    10_000,
				DiscountAmount: 0,
				SubTotal:       10_000,

				Items: []*model.PurchasedItem{
					{
						Quantity:       1,
						PurchasedPrice: 10_000,
						DiscountAmount: 0,
						TotalAmount:    10_000,
						ItemId:         1,
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			returnOrderItemId := 1
			orderItemRepo.Mock.On("Transactions", expectedParams).Return(returnOrderItemId)
			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Nil(t, err)
			assert.Equal(t, returnOrderItemId, orderItemId)
		})

		t.Run("InvalidTenantIdStoreIdUserId", func(t *testing.T) {
			invalidParams := &repository.CreateTransactionParams{
				// UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			orderItemId, err := orderItemService.Transactions(invalidParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)

			invalidParams = &repository.CreateTransactionParams{
				UserId: USER_ID,
				// TenantId: TENANT_ID,
				StoreId: STORE_ID,
			}
			orderItemId, err = orderItemService.Transactions(invalidParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)

			invalidParams = &repository.CreateTransactionParams{
				UserId:   USER_ID,
				TenantId: TENANT_ID,
				// StoreId: STORE_ID,
			}
			orderItemId, err = orderItemService.Transactions(invalidParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
		})

		t.Run("EmptyTransactions", func(t *testing.T) {
			invalidParams := &repository.CreateTransactionParams{
				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,

				Items: nil,
			}
			orderItemId, err := orderItemService.Transactions(invalidParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
			assert.Equal(t, "At least one item is required", err.Error())

			invalidParams = &repository.CreateTransactionParams{
				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,

				Items: []*model.PurchasedItem{},
			}
			orderItemId, err = orderItemService.Transactions(invalidParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
		})

		t.Run("PriceMismatch", func(t *testing.T) {
			/*
				Duplicate order with the same item id is allowed,
				however price mismatch within the same id should not happen
			*/
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    10_000,
				DiscountAmount: 0,
				SubTotal:       10_000,

				Items: []*model.PurchasedItem{
					{
						Quantity:       1,
						PurchasedPrice: 10_000,
						DiscountAmount: 0,
						TotalAmount:    10_000,
						ItemId:         1,
					},
					{
						Quantity:       1,
						PurchasedPrice: 9000, // Should be 10_000
						DiscountAmount: 0,
						TotalAmount:    10_000,
						ItemId:         1, //  It has the same id
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
			assert.ErrorContains(t, err, "Price mismatch")
		})

		t.Run("ItemTotalMismatch", func(t *testing.T) {
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    10_000,
				DiscountAmount: 0,
				SubTotal:       10_000,

				Items: []*model.PurchasedItem{
					{
						Quantity:       1,
						PurchasedPrice: 10_000,
						DiscountAmount: 100,
						TotalAmount:    10_000, // Should be 9900
						ItemId:         1,
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
			assert.ErrorContains(t, err, "total mismatch")
		})

		t.Run("TotalQuantityMismatch", func(t *testing.T) {
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  99, // Should be 1
				TotalAmount:    10_000,
				DiscountAmount: 0,
				SubTotal:       10_000,

				Items: []*model.PurchasedItem{
					{
						Quantity:       1,
						PurchasedPrice: 10_000,
						DiscountAmount: 0,
						TotalAmount:    10_000,
						ItemId:         1,
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
			assert.ErrorContains(t, err, "Total quantity")
		})

		t.Run("SubtotalMismatch", func(t *testing.T) {
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    10_000,
				DiscountAmount: 0,
				SubTotal:       9000, // Should be 10_000

				Items: []*model.PurchasedItem{
					{
						Quantity:       1,
						PurchasedPrice: 10_000,
						DiscountAmount: 0,
						TotalAmount:    10_000,
						ItemId:         1,
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
			assert.ErrorContains(t, err, "Subtotal mismatch")
		})

		t.Run("TotalAmountMismatch", func(t *testing.T) {
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    20_000, // Should be 10_000
				DiscountAmount: 0,
				SubTotal:       10_000,

				Items: []*model.PurchasedItem{
					{
						Quantity:       1,
						PurchasedPrice: 10_000,
						DiscountAmount: 0,
						TotalAmount:    10_000,
						ItemId:         1,
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
			assert.ErrorContains(t, err, "Total amount mismatch")
		})

		t.Run("DiscountAmountMismatch", func(t *testing.T) {
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    10_000,
				DiscountAmount: 100, // Should be 0
				SubTotal:       10_000,

				Items: []*model.PurchasedItem{
					{
						Quantity:       1,
						PurchasedPrice: 10_000,
						DiscountAmount: 0,
						TotalAmount:    10_000,
						ItemId:         1,
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
			assert.ErrorContains(t, err, "Discount amount mismatch")

			expectedParams = &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    9_000,
				DiscountAmount: 0, // Should be 100
				SubTotal:       10_000,

				Items: []*model.PurchasedItem{
					{
						Quantity:       1,
						PurchasedPrice: 10_000,
						DiscountAmount: 1_000,
						TotalAmount:    9_000,
						ItemId:         1,
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			orderItemId, err = orderItemService.Transactions(expectedParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
			assert.ErrorContains(t, err, "Discount amount mismatch")
		})

		t.Run("InsufficientPayment", func(t *testing.T) {
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 9000, // Should be 10_000
				TotalQuantity:  1,
				TotalAmount:    10_000,
				DiscountAmount: 0,
				SubTotal:       10_000,

				Items: []*model.PurchasedItem{
					{
						Quantity:       1,
						PurchasedPrice: 10_000,
						DiscountAmount: 0,
						TotalAmount:    10_000,
						ItemId:         1,
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
			assert.ErrorContains(t, err, "Insufficient payment")
		})

		t.Run("TooManyItems", func(t *testing.T) {
			// Create 1001 items to exceed the limit
			items := make([]*model.PurchasedItem, 1001)
			for i := 0; i < 1001; i++ {
				items[i] = &model.PurchasedItem{
					Quantity:       1,
					PurchasedPrice: 10_000,
					DiscountAmount: 0,
					TotalAmount:    10_000,
					ItemId:         i + 1,
				}
			}

			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_010_000,
				TotalQuantity:  1001,
				TotalAmount:    10_010_000,
				DiscountAmount: 0,
				SubTotal:       10_010_000,
				Items:          items,
				UserId:         USER_ID,
				TenantId:       TENANT_ID,
				StoreId:        STORE_ID,
			}

			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
			assert.ErrorContains(t, err, "Too many items")
		})

		t.Run("ExactlyMaxItems", func(t *testing.T) {
			// Test with exactly 1000 items (boundary test - should succeed)
			items := make([]*model.PurchasedItem, 1000)
			for i := 0; i < 1000; i++ {
				items[i] = &model.PurchasedItem{
					Quantity:       1,
					PurchasedPrice: 10_000,
					DiscountAmount: 0,
					TotalAmount:    10_000,
					ItemId:         i + 1,
				}
			}

			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000_000,
				TotalQuantity:  1000,
				TotalAmount:    10_000_000,
				DiscountAmount: 0,
				SubTotal:       10_000_000,
				Items:          items,
				UserId:         USER_ID,
				TenantId:       TENANT_ID,
				StoreId:        STORE_ID,
			}

			returnOrderItemId := 1
			orderItemRepo.Mock = &mock.Mock{}
			orderItemRepo.Mock.On("Transactions", expectedParams).Return(returnOrderItemId)
			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Nil(t, err)
			assert.Equal(t, returnOrderItemId, orderItemId)
		})

		t.Run("MultipleItemsWithDiscount", func(t *testing.T) {
			// Test with multiple different items and discounts
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 30_000,
				TotalQuantity:  5,
				TotalAmount:    27_700,
				DiscountAmount: 1_300,
				SubTotal:       29_000, // 20_000 + 9_000

				Items: []*model.PurchasedItem{
					{
						Quantity:       2,
						PurchasedPrice: 10_000,
						DiscountAmount: 500,
						TotalAmount:    19_000, // (10_000 * 2) - (500 * 2)
						ItemId:         1,
					},
					{
						Quantity:       3,
						PurchasedPrice: 3_000,
						DiscountAmount: 100,
						TotalAmount:    8_700, // (3_000 * 3) - (100 * 3)
						ItemId:         2,
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			returnOrderItemId := 1
			orderItemRepo.Mock = &mock.Mock{}
			orderItemRepo.Mock.On("Transactions", expectedParams).Return(returnOrderItemId)
			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Nil(t, err)
			assert.Equal(t, returnOrderItemId, orderItemId)
		})

		t.Run("RepositoryError", func(t *testing.T) {
			// Test repository error handling
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    10_000,
				DiscountAmount: 0,
				SubTotal:       10_000,

				Items: []*model.PurchasedItem{
					{
						Quantity:       1,
						PurchasedPrice: 10_000,
						DiscountAmount: 0,
						TotalAmount:    10_000,
						ItemId:         1,
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			orderItemRepo.Mock = &mock.Mock{}
			orderItemRepo.Mock.On("Transactions", expectedParams).Return(0, errors.New("database error"))
			orderItemId, err := orderItemService.Transactions(expectedParams)
			assert.Error(t, err)
			assert.Equal(t, 0, orderItemId)
			assert.ErrorContains(t, err, "Failed to create transaction")
		})
	})
}
