package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestOrderItemServiceImpl(t *testing.T) {
	now := time.Now()
	const USER_ID = 1
	const TENANT_ID = 1
	const STORE_ID = 1
	const LIMIT = 10
	const PAGE = 1

	t.Run("Get", func(t *testing.T) {
		paymentProvider := NewPaymentProviderImpl()

		t.Run("NormalGet", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)
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
					CreatedAt:      now,
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

		t.Run("TenantIdIsNotProvided", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)
			// Invalid tenant id
			orderItems, count, err := orderItemService.Get(0, STORE_ID, LIMIT, PAGE, nil, nil)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("InvalidLimitAndPageParams", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

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
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, 0, nil, nil)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "page could not less then 1")
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("InvalidPage_Negative", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

			orderItems, count, err := orderItemService.Get(TENANT_ID, STORE_ID, LIMIT, -1, nil, nil)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "page could not less then 1")
			assert.Equal(t, 0, count)
			assert.Nil(t, orderItems)
		})

		t.Run("DateFilter_StartDateAfterEndDate", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

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
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

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
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

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
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

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
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

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
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

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
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

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
		paymentProvider := NewPaymentProviderImpl()

		t.Run("NormalFindById", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

			expectedOrderItem := &model.OrderItemWithStore{
				Id:             1,
				DiscountAmount: 0,
				Subtotal:       10000,
				TotalAmount:    10000,
				TotalQuantity:  1,
				PurchasedPrice: 10000,
				StoreId:        STORE_ID,
				TenantId:       TENANT_ID,
				CreatedAt:      now,
				StoreName:      "Test Store Name",
				PaymentType:    model.PaymentTypeQRIS,
			}

			expectedPurchasedList := []*model.PurchasedItem{
				{
					Id:                 1,
					TotalAmount:        10000,
					Quantity:           1,
					StorePriceSnapshot: 10000,
					DiscountAmount:     0,
					ItemId:             1,
					OrderItemId:        expectedOrderItem.Id, // Connect with orderItem id
					CreatedAt:          &now,
				},
				{
					Id:                 2,
					TotalAmount:        5000,
					Quantity:           1,
					StorePriceSnapshot: 5000,
					DiscountAmount:     0,
					ItemId:             2,
					OrderItemId:        expectedOrderItem.Id, // Connect with orderItem id
					CreatedAt:          &now,
				},
			}

			// Page minus by 1 because page will be 0 index from repository
			orderItemRepo.Mock.On("FindById", expectedOrderItem.Id, TENANT_ID).Return(expectedOrderItem, expectedPurchasedList, nil)
			orderItem, purchasedItemList, err := orderItemService.FindById(expectedOrderItem.Id, TENANT_ID)
			assert.NoError(t, err)
			assert.Len(t, purchasedItemList, 2)
			assert.NotNil(t, orderItem)
			assert.Equal(t, expectedOrderItem.StoreName, orderItem.StoreName)
			assert.Equal(t, expectedOrderItem.PaymentType, orderItem.PaymentType)
			for i, purchasedItem := range purchasedItemList {
				assert.Equal(t, expectedPurchasedList[i].Id, purchasedItem.Id)
				assert.Equal(t, expectedPurchasedList[i].ItemId, purchasedItem.ItemId)
				assert.Equal(t, expectedPurchasedList[i].OrderItemId, purchasedItem.OrderItemId)
			}
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

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
		cashServerKey := "NOT_SERVER_KEY_THIS_KEY_WILL_BE_USE_FOR_CASH_PAYMENT_METHOD" // Will not payment gateway for this case
		paymentProvider := NewPaymentProviderImpl()

		orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
		orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

		baseValidParams := func() *repository.CreateTransactionParams {
			return &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    10_000,
				DiscountAmount: 0,
				SubTotal:       10_000,
				Items: []*model.PurchasedItem{
					{
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     0,
						TotalAmount:        10_000,
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},
				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}
		}

		t.Run("NormalTransactions", func(t *testing.T) {
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    10_000,
				DiscountAmount: 0,
				SubTotal:       10_000,
				PaymentType:    model.PaymentTypeCash,

				Items: []*model.PurchasedItem{
					{
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     0,
						TotalAmount:        10_000,
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			now := time.Now()
			expectedTransactionDataReturn := &repository.TransactionDataReturn{
				CreatedOrderItemId: 1,
				CreatedAt:          &now,
			}
			orderItemRepo.Mock.On("Transactions", expectedParams).Return(expectedTransactionDataReturn, nil)
			transactionDataReturn, err := orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Nil(t, err)
			assert.Equal(t, expectedTransactionDataReturn.CreatedOrderItemId, transactionDataReturn.CreatedOrderItemId)
			assert.NotNil(t, transactionDataReturn.CreatedAt)
		})

		t.Run("InvalidTenantIdStoreIdUserId", func(t *testing.T) {
			invalidParams := &repository.CreateTransactionParams{
				// UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			transactionDataReturn, err := orderItemService.Transactions(invalidParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)

			invalidParams = &repository.CreateTransactionParams{
				UserId: USER_ID,
				// TenantId: TENANT_ID,
				StoreId: STORE_ID,
			}
			transactionDataReturn, err = orderItemService.Transactions(invalidParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)

			invalidParams = &repository.CreateTransactionParams{
				UserId:   USER_ID,
				TenantId: TENANT_ID,
				// StoreId: STORE_ID,
			}
			transactionDataReturn, err = orderItemService.Transactions(invalidParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
		})

		t.Run("EmptyTransactions", func(t *testing.T) {
			invalidParams := &repository.CreateTransactionParams{
				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,

				Items: nil,
			}
			transactionDataReturn, err := orderItemService.Transactions(invalidParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
			assert.Equal(t, "At least one item is required", err.Error())

			invalidParams = &repository.CreateTransactionParams{
				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,

				Items: []*model.PurchasedItem{},
			}
			transactionDataReturn, err = orderItemService.Transactions(invalidParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
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
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     0,
						TotalAmount:        10_000,
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
					{
						Quantity:           1,
						StorePriceSnapshot: 9000, // Should be 10_000
						DiscountAmount:     0,
						TotalAmount:        10_000,
						ItemId:             1, //  It has the same id
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			transactionDataReturn, err := orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
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
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     100,
						TotalAmount:        10_000, // Should be 9900
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			transactionDataReturn, err := orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
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
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     0,
						TotalAmount:        10_000,
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			transactionDataReturn, err := orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
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
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     0,
						TotalAmount:        10_000,
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			transactionDataReturn, err := orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
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
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     0,
						TotalAmount:        10_000,
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			transactionDataReturn, err := orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
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
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     0,
						TotalAmount:        10_000,
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			transactionDataReturn, err := orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
			assert.ErrorContains(t, err, "Discount amount mismatch")

			expectedParams = &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    9_000,
				DiscountAmount: 0, // Should be 100
				SubTotal:       10_000,

				Items: []*model.PurchasedItem{
					{
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     1_000,
						TotalAmount:        9_000,
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			transactionDataReturn, err = orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)

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
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     0,
						TotalAmount:        10_000,
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			transactionDataReturn, err := orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
			assert.ErrorContains(t, err, "Insufficient payment")
		})

		t.Run("InvalidPaymentType", func(t *testing.T) {
			invalidParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    10_000,
				DiscountAmount: 0,
				SubTotal:       10_000,
				PaymentType:    "VISA",

				Items: []*model.PurchasedItem{
					{
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     0,
						TotalAmount:        10_000,
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			transactionDataReturn, err := orderItemService.Transactions(invalidParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
			assert.ErrorContains(t, err, "invalid payment_type")
		})

		t.Run("TooManyItems", func(t *testing.T) {
			// Create 1001 items to exceed the limit
			items := make([]*model.PurchasedItem, 1001)
			for i := 0; i < 1001; i++ {
				items[i] = &model.PurchasedItem{
					Quantity:           1,
					StorePriceSnapshot: 10_000,
					DiscountAmount:     0,
					TotalAmount:        10_000,
					ItemId:             i + 1,
					ItemNameSnapshot:   fmt.Sprintf("Item Name snapshot %d", i),
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

			transactionDataReturn, err := orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
			assert.ErrorContains(t, err, "Too many items")
		})

		t.Run("ExactlyMaxItems", func(t *testing.T) {
			// Test with exactly 1000 items (boundary test - should succeed)
			items := make([]*model.PurchasedItem, 1000)
			for i := 0; i < 1000; i++ {
				items[i] = &model.PurchasedItem{
					Quantity:           1,
					StorePriceSnapshot: 10_000,
					DiscountAmount:     0,
					TotalAmount:        10_000,
					ItemId:             i + 1,
					ItemNameSnapshot:   fmt.Sprintf("Item Name snapshot %d", i),
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

			now := time.Now()
			expectedTransactionDataReturn := &repository.TransactionDataReturn{
				CreatedAt:          &now,
				CreatedOrderItemId: 1,
			}
			orderItemRepo.Mock = &mock.Mock{}
			orderItemRepo.Mock.On("Transactions", expectedParams).Return(expectedTransactionDataReturn)
			transactionReturnData, err := orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Nil(t, err)
			assert.Equal(t, expectedTransactionDataReturn.CreatedOrderItemId, transactionReturnData.CreatedOrderItemId)
			assert.NotNil(t, transactionReturnData.CreatedAt)
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
						Quantity:           2,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     500,
						TotalAmount:        19_000, // (10_000 * 2) - (500 * 2)
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot 1",
					},
					{
						Quantity:           3,
						StorePriceSnapshot: 3_000,
						DiscountAmount:     100,
						TotalAmount:        8_700, // (3_000 * 3) - (100 * 3)
						ItemId:             2,
						ItemNameSnapshot:   "Item Name Snapshot 2",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			now := time.Now()
			expectedTransactionDataReturn := &repository.TransactionDataReturn{
				CreatedAt:          &now,
				CreatedOrderItemId: 1,
			}
			orderItemRepo.Mock = &mock.Mock{}
			orderItemRepo.Mock.On("Transactions", expectedParams).Return(expectedTransactionDataReturn)
			transactionDataReturn, err := orderItemService.Transactions(expectedParams, cashServerKey)
			assert.Nil(t, err)
			assert.Equal(t, expectedTransactionDataReturn.CreatedOrderItemId, transactionDataReturn.CreatedOrderItemId)
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
						Quantity:           1,
						StorePriceSnapshot: 10_000,
						DiscountAmount:     0,
						TotalAmount:        10_000,
						ItemId:             1,
						ItemNameSnapshot:   "Item Name Snapshot",
					},
				},

				UserId:   USER_ID,
				TenantId: TENANT_ID,
				StoreId:  STORE_ID,
			}

			orderItemRepo.Mock = &mock.Mock{}
			orderItemRepo.Mock.On("Transactions", expectedParams).Return(nil, errors.New("database error"))
			transactionDataReturn, err := orderItemService.Transactions(expectedParams, "")
			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
			assert.ErrorContains(t, err, "Failed to create transaction")
		})

		t.Run("NegativePrice", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			params := baseValidParams()
			params.Items[0].StorePriceSnapshot = -10_000
			params.Items[0].TotalAmount = -10_000
			params.PurchasedPrice = -10_000
			params.TotalAmount = -10_000
			params.SubTotal = -10_000

			transactionDataReturn, err := orderItemService.Transactions(params, cashServerKey)

			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
			assert.ErrorContains(t, err, "price cannot be negative")
			paymentProviderMock.Mock.AssertNotCalled(t, "CreateTransaction", mock.Anything, mock.Anything)
			orderItemRepo.Mock.AssertNotCalled(t, "Transactions", mock.Anything)
		})

		t.Run("NegativeDiscountAmount", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			params := baseValidParams()
			params.Items[0].DiscountAmount = -100

			transactionDataReturn, err := orderItemService.Transactions(params, cashServerKey)

			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
			assert.ErrorContains(t, err, "discount amount is invalid")
		})

		t.Run("DiscountExceedsPrice", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			params := baseValidParams()
			params.Items[0].DiscountAmount = 10_001 // exceeds StorePriceSnapshot of 10_000
			params.Items[0].TotalAmount = -1

			transactionDataReturn, err := orderItemService.Transactions(params, cashServerKey)

			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
			assert.ErrorContains(t, err, "discount amount is invalid")
		})

		t.Run("InvalidItemNameSnapshot", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			params := baseValidParams()
			params.Items[0].ItemNameSnapshot = "<script>alert(1)</script>"

			transactionDataReturn, err := orderItemService.Transactions(params, cashServerKey)

			assert.Error(t, err)
			assert.Nil(t, transactionDataReturn)
			assert.ErrorContains(t, err, "Illegal input from item name snapshot")
		})

		t.Run("DefaultPaymentTypeWhenEmpty", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			params := baseValidParams()
			params.PaymentType = "" // should default to Cash

			expectedRepoParams := *params
			expectedRepoParams.PaymentType = model.PaymentTypeCash
			expectedRepoParams.PaymentToken = ""
			expectedRepoParams.PaymentURL = ""

			expectedReturn := &repository.TransactionDataReturn{CreatedOrderItemId: 1}
			orderItemRepo.Mock.On("Transactions", &expectedRepoParams).Return(expectedReturn, nil)

			transactionDataReturn, err := orderItemService.Transactions(params, cashServerKey)

			assert.NoError(t, err)
			assert.Equal(t, model.PaymentTypeCash, params.PaymentType)
			assert.Equal(t, expectedReturn.CreatedOrderItemId, transactionDataReturn.CreatedOrderItemId)
			paymentProviderMock.Mock.AssertNotCalled(t, "CreateTransaction", mock.Anything, mock.Anything)
		})

		t.Run("QRISPaymentType_CreatesGatewayTransactionAndPersistsToken", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			params := baseValidParams()
			params.PaymentType = model.PaymentTypeQRIS

			gatewayResponse := &model.PaymentProviderResponse{
				Token:       "snap-token-123",
				RedirectURL: "https://app.midtrans.com/snap/v3/redirection/snap-token-123",
			}
			paymentProviderMock.Mock.On("CreateTransaction", mock.Anything, params).Return(gatewayResponse, nil)

			expectedRepoParams := *params
			expectedRepoParams.PaymentToken = gatewayResponse.Token
			expectedRepoParams.PaymentURL = gatewayResponse.RedirectURL

			expectedReturn := &repository.TransactionDataReturn{CreatedOrderItemId: 1}
			orderItemRepo.Mock.On("Transactions", &expectedRepoParams).Return(expectedReturn, nil)

			transactionDataReturn, err := orderItemService.Transactions(params, "some-server-key")

			assert.NoError(t, err)
			assert.Equal(t, gatewayResponse.Token, transactionDataReturn.PaymentToken)
			assert.Equal(t, gatewayResponse.RedirectURL, transactionDataReturn.PaymentURL)
			paymentProviderMock.Mock.AssertExpectations(t)
			orderItemRepo.Mock.AssertExpectations(t)
		})

		t.Run("CardPaymentType_CreatesGatewayTransaction", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			params := baseValidParams()
			params.PaymentType = model.PaymentTypeCard

			gatewayResponse := &model.PaymentProviderResponse{
				Token:       "snap-token-card",
				RedirectURL: "https://app.midtrans.com/snap/v3/redirection/snap-token-card",
			}
			paymentProviderMock.Mock.On("CreateTransaction", mock.Anything, params).Return(gatewayResponse, nil)

			expectedRepoParams := *params
			expectedRepoParams.PaymentToken = gatewayResponse.Token
			expectedRepoParams.PaymentURL = gatewayResponse.RedirectURL

			expectedReturn := &repository.TransactionDataReturn{CreatedOrderItemId: 2}
			orderItemRepo.Mock.On("Transactions", &expectedRepoParams).Return(expectedReturn, nil)

			transactionDataReturn, err := orderItemService.Transactions(params, "some-server-key")

			assert.NoError(t, err)
			assert.Equal(t, gatewayResponse.Token, transactionDataReturn.PaymentToken)
			paymentProviderMock.Mock.AssertExpectations(t)
		})

		t.Run("EWalletPaymentType_CreatesGatewayTransaction", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			params := baseValidParams()
			params.PaymentType = model.PaymentTypeEWallet

			gatewayResponse := &model.PaymentProviderResponse{
				Token:       "snap-token-ewallet",
				RedirectURL: "https://app.midtrans.com/snap/v3/redirection/snap-token-ewallet",
			}
			paymentProviderMock.Mock.On("CreateTransaction", mock.Anything, params).Return(gatewayResponse, nil)

			expectedRepoParams := *params
			expectedRepoParams.PaymentToken = gatewayResponse.Token
			expectedRepoParams.PaymentURL = gatewayResponse.RedirectURL

			expectedReturn := &repository.TransactionDataReturn{CreatedOrderItemId: 3}
			orderItemRepo.Mock.On("Transactions", &expectedRepoParams).Return(expectedReturn, nil)

			transactionDataReturn, err := orderItemService.Transactions(params, "some-server-key")

			assert.NoError(t, err)
			assert.Equal(t, gatewayResponse.Token, transactionDataReturn.PaymentToken)
			paymentProviderMock.Mock.AssertExpectations(t)
		})

		t.Run("OtherPaymentType_NoGatewayCall", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			params := baseValidParams()
			params.PaymentType = model.PaymentTypeOther

			expectedRepoParams := *params
			expectedRepoParams.PaymentToken = ""
			expectedRepoParams.PaymentURL = ""

			expectedReturn := &repository.TransactionDataReturn{CreatedOrderItemId: 4}
			orderItemRepo.Mock.On("Transactions", &expectedRepoParams).Return(expectedReturn, nil)

			transactionDataReturn, err := orderItemService.Transactions(params, "some-server-key")

			assert.NoError(t, err)
			assert.Empty(t, transactionDataReturn.PaymentToken)
			assert.Empty(t, transactionDataReturn.PaymentURL)
			paymentProviderMock.Mock.AssertNotCalled(t, "CreateTransaction", mock.Anything, mock.Anything)
		})

		t.Run("GatewayCreateTransactionFails", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			params := baseValidParams()
			params.PaymentType = model.PaymentTypeQRIS

			gatewayErr := errors.New("gateway unreachable")
			paymentProviderMock.Mock.On("CreateTransaction", mock.Anything, params).Return(nil, gatewayErr)

			transactionDataReturn, err := orderItemService.Transactions(params, "some-server-key")

			assert.Error(t, err)
			assert.Equal(t, gatewayErr, err)
			assert.Nil(t, transactionDataReturn)
			// Repository must never be touched if the gateway call itself failed.
			orderItemRepo.Mock.AssertNotCalled(t, "Transactions", mock.Anything)
		})

		/*
			t.Run("RepositoryError_NoCompensatingCancelCalled", func(t *testing.T) {
				orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
				paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
				orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

				params := baseValidParams()
				params.PaymentType = model.PaymentTypeQRIS

				gatewayResponse := &model.PaymentProviderResponse{
					Token:       "snap-token-orphan",
					RedirectURL: "https://app.midtrans.com/snap/v3/redirection/snap-token-orphan",
				}
				paymentProviderMock.Mock.On("CreateTransaction", mock.Anything, params).Return(gatewayResponse, nil)

				expectedRepoParams := *params
				expectedRepoParams.PaymentToken = gatewayResponse.Token
				expectedRepoParams.PaymentURL = gatewayResponse.RedirectURL

				dbErr := errors.New("database error")
				orderItemRepo.Mock.On("Transactions", &expectedRepoParams).Return(nil, dbErr)

				transactionDataReturn, err := orderItemService.Transactions(params, "some-server-key")

				assert.Error(t, err)
				assert.Nil(t, transactionDataReturn)
				assert.ErrorContains(t, err, "Failed to create transaction")
				paymentProviderMock.Mock.AssertNotCalled(t, "CancelTransaction", mock.Anything, mock.Anything)
			})
		*/
	})

	t.Run("CheckTransaction", func(t *testing.T) {
		const TRANSACTION_ID = "TEST_TRANSACTION_ID_NOT_REAL"
		const SERVER_KEY = "dummy-server-key"
		const ORDER_ITEM_ID = 501

		t.Run("EmptyTransactionId", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			response, err := orderItemService.CheckTransaction("", SERVER_KEY)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "transaction id is required")
			assert.Equal(t, model.PaymentStatusResponse{}, response)
			paymentProviderMock.Mock.AssertNotCalled(t, "CheckTransaction", mock.Anything, mock.Anything)
		})

		// Non-success statuses: persist and return immediately. The stock-sync
		// branch must never be touched for any of these.
		t.Run("NonSuccessStatuses", func(t *testing.T) {
			statuses := []model.PaymentStatus{
				model.PaymentStatusPending,
				model.PaymentStatusRefunded,
				model.PaymentStatusFailed,
				model.PaymentStatusExpired,
				model.PaymentStatusCancelled,
				model.PaymentStatusPartiallyRefunded,
			}

			for _, status := range statuses {
				t.Run(string(status), func(t *testing.T) {
					orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
					paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
					orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

					expectedResponse := model.PaymentStatusResponse{
						PaymentStatus: status,
					}

					paymentProviderMock.Mock.On("CheckTransaction", mock.Anything, TRANSACTION_ID).
						Return(expectedResponse, nil)
					orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, status).
						Return(nil)

					response, err := orderItemService.CheckTransaction(TRANSACTION_ID, SERVER_KEY)

					assert.NoError(t, err)
					assert.Equal(t, status, response.PaymentStatus)
					assert.Empty(t, response.Message)
					paymentProviderMock.Mock.AssertExpectations(t)
					orderItemRepo.Mock.AssertExpectations(t)
					orderItemRepo.Mock.AssertNotCalled(t, "GetOrderItemByTransactionId", mock.Anything)
					orderItemRepo.Mock.AssertNotCalled(t, "SyncDataStock", mock.Anything)
				})
			}
		})

		t.Run("SetPaymentStatus_LocalPersistFails", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			expectedResponse := model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatusSuccess,
			}

			paymentProviderMock.Mock.On("CheckTransaction", mock.Anything, TRANSACTION_ID).
				Return(expectedResponse, nil)
			orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, model.PaymentStatusSuccess).
				Return(errors.New("db write failed"))

			response, err := orderItemService.CheckTransaction(TRANSACTION_ID, SERVER_KEY)

			// Gateway confirmed status, so this is NOT a hard error — only a message.
			assert.NoError(t, err)
			assert.Equal(t, model.PaymentStatusSuccess, response.PaymentStatus)
			assert.Contains(t, response.Message, "failed to persist locally")
			assert.Contains(t, response.Message, "db write failed")
			// Must return before ever reaching the stock-sync branch.
			orderItemRepo.Mock.AssertNotCalled(t, "GetOrderItemByTransactionId", mock.Anything)
			orderItemRepo.Mock.AssertNotCalled(t, "SyncDataStock", mock.Anything)
		})

		// Success: gateway confirmed, local status persisted. From here the
		// four sub-branches of the stock-sync logic are tested individually.
		t.Run("Success_OrderLookupFails", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			expectedResponse := model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatusSuccess,
			}

			paymentProviderMock.Mock.On("CheckTransaction", mock.Anything, TRANSACTION_ID).
				Return(expectedResponse, nil)
			orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, model.PaymentStatusSuccess).
				Return(nil)
			orderItemRepo.Mock.On("GetOrderItemByTransactionId", TRANSACTION_ID).
				Return(model.OrderItem{}, errors.New("record not found"))

			response, err := orderItemService.CheckTransaction(TRANSACTION_ID, SERVER_KEY)

			// Payment is still confirmed — a lookup failure must not surface as a payment failure.
			assert.NoError(t, err)
			assert.Equal(t, model.PaymentStatusSuccess, response.PaymentStatus)
			assert.Contains(t, response.Message, "stock could not be synced")
			orderItemRepo.Mock.AssertExpectations(t)
			orderItemRepo.Mock.AssertNotCalled(t, "SyncDataStock", mock.Anything)
		})

		t.Run("Success_AlreadySynced", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			expectedResponse := model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatusSuccess,
			}
			orderItem := model.OrderItem{Id: ORDER_ITEM_ID, IsDataStockSync: true}

			paymentProviderMock.Mock.On("CheckTransaction", mock.Anything, TRANSACTION_ID).
				Return(expectedResponse, nil)
			orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, model.PaymentStatusSuccess).
				Return(nil)
			orderItemRepo.Mock.On("GetOrderItemByTransactionId", TRANSACTION_ID).
				Return(orderItem, nil)

			response, err := orderItemService.CheckTransaction(TRANSACTION_ID, SERVER_KEY)

			assert.NoError(t, err)
			assert.Equal(t, model.PaymentStatusSuccess, response.PaymentStatus)
			assert.Equal(t, "Payment completed", response.Message)
			orderItemRepo.Mock.AssertExpectations(t)
			// Already synced — SyncDataStock must be skipped entirely.
			orderItemRepo.Mock.AssertNotCalled(t, "SyncDataStock", mock.Anything)
		})

		t.Run("Success_SyncFails", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			expectedResponse := model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatusSuccess,
			}
			orderItem := model.OrderItem{Id: ORDER_ITEM_ID, IsDataStockSync: false}

			paymentProviderMock.Mock.On("CheckTransaction", mock.Anything, TRANSACTION_ID).
				Return(expectedResponse, nil)
			orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, model.PaymentStatusSuccess).
				Return(nil)
			orderItemRepo.Mock.On("GetOrderItemByTransactionId", TRANSACTION_ID).
				Return(orderItem, nil)
			orderItemRepo.Mock.On("SyncDataStock", ORDER_ITEM_ID).
				Return(errors.New("stock decrement failed"))

			response, err := orderItemService.CheckTransaction(TRANSACTION_ID, SERVER_KEY)

			// Payment is still confirmed — sync failure must not look like a payment failure.
			assert.NoError(t, err)
			assert.Equal(t, model.PaymentStatusSuccess, response.PaymentStatus)
			assert.Contains(t, response.Message, "stock sync failed")
			orderItemRepo.Mock.AssertExpectations(t)
		})

		t.Run("Success_SyncSucceeds", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			expectedResponse := model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatusSuccess,
			}
			orderItem := model.OrderItem{Id: ORDER_ITEM_ID, IsDataStockSync: false}

			paymentProviderMock.Mock.On("CheckTransaction", mock.Anything, TRANSACTION_ID).
				Return(expectedResponse, nil)
			orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, model.PaymentStatusSuccess).
				Return(nil)
			orderItemRepo.Mock.On("GetOrderItemByTransactionId", TRANSACTION_ID).
				Return(orderItem, nil)
			orderItemRepo.Mock.On("SyncDataStock", ORDER_ITEM_ID).
				Return(nil)

			response, err := orderItemService.CheckTransaction(TRANSACTION_ID, SERVER_KEY)

			assert.NoError(t, err)
			assert.Equal(t, model.PaymentStatusSuccess, response.PaymentStatus)
			assert.Equal(t, "Payment completed and data synced successfully", response.Message)
			orderItemRepo.Mock.AssertExpectations(t)
		})

		t.Run("UnknownPaymentStatus", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			expectedResponse := model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatus("SOME_UNKNOWN_STATUS"),
			}

			paymentProviderMock.Mock.On("CheckTransaction", mock.Anything, TRANSACTION_ID).
				Return(expectedResponse, nil)

			response, err := orderItemService.CheckTransaction(TRANSACTION_ID, SERVER_KEY)

			assert.NoError(t, err)
			assert.Equal(t, expectedResponse.PaymentStatus, response.PaymentStatus)
			orderItemRepo.Mock.AssertNotCalled(t, "SetPaymentStatus", mock.Anything, mock.Anything, mock.Anything)
			orderItemRepo.Mock.AssertNotCalled(t, "GetOrderItemByTransactionId", mock.Anything)
		})

		t.Run("GatewayError_NonNotFound", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			gatewayErr := errors.New("500 internal server error")

			paymentProviderMock.Mock.On("CheckTransaction", mock.Anything, TRANSACTION_ID).
				Return(nil, gatewayErr)

			response, err := orderItemService.CheckTransaction(TRANSACTION_ID, SERVER_KEY)

			assert.Error(t, err)
			assert.Equal(t, gatewayErr, err)
			assert.Equal(t, model.PaymentStatusResponse{}, response)
			orderItemRepo.Mock.AssertNotCalled(t, "SetPaymentStatus", mock.Anything, mock.Anything, mock.Anything)
		})

		t.Run("GatewayError_404_CancelSucceeds", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			gatewayErr := errors.New("404 transaction not found")

			paymentProviderMock.Mock.On("CheckTransaction", mock.Anything, TRANSACTION_ID).
				Return(nil, gatewayErr)
			orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, model.PaymentStatusCancelled).
				Return(nil)

			response, err := orderItemService.CheckTransaction(TRANSACTION_ID, SERVER_KEY)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "never completed at the payment gateway")
			assert.Contains(t, err.Error(), "cancelled")
			assert.Equal(t, model.PaymentStatusResponse{}, response)
			orderItemRepo.Mock.AssertExpectations(t)
		})

		t.Run("GatewayError_404_NoLocalRecord", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			gatewayErr := errors.New("404 transaction not found")

			paymentProviderMock.Mock.On("CheckTransaction", mock.Anything, TRANSACTION_ID).
				Return(nil, gatewayErr)
			orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, model.PaymentStatusCancelled).
				Return(gorm.ErrRecordNotFound)

			response, err := orderItemService.CheckTransaction(TRANSACTION_ID, SERVER_KEY)

			assert.Error(t, err)
			assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			assert.Contains(t, err.Error(), "not found")
			assert.Equal(t, model.PaymentStatusResponse{}, response)
			orderItemRepo.Mock.AssertExpectations(t)
		})

		t.Run("GatewayError_404_CancelFailsWithInfraError", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			gatewayErr := errors.New("404 transaction not found")
			dbErr := errors.New("connection refused")

			paymentProviderMock.Mock.On("CheckTransaction", mock.Anything, TRANSACTION_ID).
				Return(nil, gatewayErr)
			orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, model.PaymentStatusCancelled).
				Return(dbErr)

			response, err := orderItemService.CheckTransaction(TRANSACTION_ID, SERVER_KEY)

			assert.Error(t, err)
			assert.ErrorIs(t, err, dbErr)
			assert.Contains(t, err.Error(), "failed to update transaction")
			assert.Equal(t, model.PaymentStatusResponse{}, response)
			orderItemRepo.Mock.AssertExpectations(t)
		})
	})

	t.Run("CancelTransaction", func(t *testing.T) {
		const SERVER_KEY = "dummy-server-key"

		t.Run("NoOrderItemIdOrTransactionId", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			response, err := orderItemService.CancelTransaction(0, "", TENANT_ID, SERVER_KEY)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "either order_item id or transaction_id is required")
			assert.Equal(t, model.PaymentStatusResponse{}, response)
			orderItemRepo.Mock.AssertNotCalled(t, "FindById", mock.Anything, mock.Anything)
			orderItemRepo.Mock.AssertNotCalled(t, "SetPaymentStatus", mock.Anything, mock.Anything, mock.Anything)
			paymentProviderMock.Mock.AssertNotCalled(t, "CancelTransaction", mock.Anything, mock.Anything)
		})

		t.Run("NegativeOrderItemIdNoTransactionId", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			response, err := orderItemService.CancelTransaction(-5, "", TENANT_ID, SERVER_KEY)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "either order_item id or transaction_id is required")
			assert.Equal(t, model.PaymentStatusResponse{}, response)
			orderItemRepo.Mock.AssertNotCalled(t, "FindById", mock.Anything, mock.Anything)
		})

		t.Run("TransactionIdTakesPriority_WithZeroOrderItemId", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			const TRANSACTION_ID = "TRX-DIRECT-1"

			expectedResponse := model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatusCancelled,
			}

			paymentProviderMock.Mock.On("CancelTransaction", mock.Anything, TRANSACTION_ID).
				Return(expectedResponse, nil)
			orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, model.PaymentStatusCancelled).
				Return(nil)

			response, err := orderItemService.CancelTransaction(0, TRANSACTION_ID, TENANT_ID, SERVER_KEY)

			assert.NoError(t, err)
			assert.Equal(t, model.PaymentStatusCancelled, response.PaymentStatus)
			assert.Empty(t, response.Message)
			orderItemRepo.Mock.AssertNotCalled(t, "FindById", mock.Anything, mock.Anything)
			paymentProviderMock.Mock.AssertExpectations(t)
			orderItemRepo.Mock.AssertExpectations(t)
		})

		t.Run("TransactionIdTakesPriority_OverPositiveOrderItemId", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			const ORDER_ITEM_ID = 5
			const TRANSACTION_ID = "TRX-DIRECT-2"

			expectedResponse := model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatusCancelled,
			}

			paymentProviderMock.Mock.On("CancelTransaction", mock.Anything, TRANSACTION_ID).
				Return(expectedResponse, nil)
			orderItemRepo.Mock.On("SetPaymentStatus", ORDER_ITEM_ID, TRANSACTION_ID, model.PaymentStatusCancelled).
				Return(nil)

			response, err := orderItemService.CancelTransaction(ORDER_ITEM_ID, TRANSACTION_ID, TENANT_ID, SERVER_KEY)

			assert.NoError(t, err)
			assert.Equal(t, model.PaymentStatusCancelled, response.PaymentStatus)
			// FindById should be skipped entirely — transactionId wins when both are supplied.
			orderItemRepo.Mock.AssertNotCalled(t, "FindById", mock.Anything, mock.Anything)
			paymentProviderMock.Mock.AssertExpectations(t)
			orderItemRepo.Mock.AssertExpectations(t)
		})

		t.Run("ResolvesTransactionIdViaFindById", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			const ORDER_ITEM_ID = 7

			foundOrderItem := &model.OrderItemWithStore{
				Id:            ORDER_ITEM_ID,
				TransactionId: "TRX-FOUND-1",
			}

			expectedResponse := model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatusCancelled,
			}

			// The FindById mock returns nil for BOTH values if EITHER argument passed
			// to .Return(...) is nil — so an empty (non-nil) slice is used here rather
			// than literal nil, or orderItem itself would come back nil too.
			orderItemRepo.Mock.On("FindById", ORDER_ITEM_ID, TENANT_ID).
				Return(foundOrderItem, []*model.PurchasedItem{}, nil)
			paymentProviderMock.Mock.On("CancelTransaction", mock.Anything, "TRX-FOUND-1").
				Return(expectedResponse, nil)
			orderItemRepo.Mock.On("SetPaymentStatus", ORDER_ITEM_ID, "TRX-FOUND-1", model.PaymentStatusCancelled).
				Return(nil)

			response, err := orderItemService.CancelTransaction(ORDER_ITEM_ID, "", TENANT_ID, SERVER_KEY)

			assert.NoError(t, err)
			assert.Equal(t, model.PaymentStatusCancelled, response.PaymentStatus)
			orderItemRepo.Mock.AssertExpectations(t)
			paymentProviderMock.Mock.AssertExpectations(t)
		})

		t.Run("FindByIdFails", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			const ORDER_ITEM_ID = 999

			findErr := errors.New("order item not found")

			orderItemRepo.Mock.On("FindById", ORDER_ITEM_ID, TENANT_ID).
				Return(nil, nil, findErr)

			response, err := orderItemService.CancelTransaction(ORDER_ITEM_ID, "", TENANT_ID, SERVER_KEY)

			assert.Error(t, err)
			assert.Equal(t, findErr, err)
			assert.Equal(t, model.PaymentStatusResponse{}, response)
			paymentProviderMock.Mock.AssertNotCalled(t, "CancelTransaction", mock.Anything, mock.Anything)
			orderItemRepo.Mock.AssertNotCalled(t, "SetPaymentStatus", mock.Anything, mock.Anything, mock.Anything)
		})

		t.Run("PaymentGatewayRejectsCancellation", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			const TRANSACTION_ID = "TRX-REJECT-1"

			gatewayErr := errors.New("412 already settled, cannot cancel")

			paymentProviderMock.Mock.On("CancelTransaction", mock.Anything, TRANSACTION_ID).
				Return(nil, gatewayErr)

			response, err := orderItemService.CancelTransaction(0, TRANSACTION_ID, TENANT_ID, SERVER_KEY)

			assert.Error(t, err)
			assert.Equal(t, gatewayErr, err)
			assert.Equal(t, model.PaymentStatusResponse{}, response)
			orderItemRepo.Mock.AssertNotCalled(t, "SetPaymentStatus", mock.Anything, mock.Anything, mock.Anything)
		})

		t.Run("SetPaymentStatus_LocalPersistFails", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			const TRANSACTION_ID = "TRX-PERSIST-FAIL"

			expectedResponse := model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatusCancelled,
			}

			paymentProviderMock.Mock.On("CancelTransaction", mock.Anything, TRANSACTION_ID).
				Return(expectedResponse, nil)
			orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, model.PaymentStatusCancelled).
				Return(errors.New("db write failed"))

			response, err := orderItemService.CancelTransaction(0, TRANSACTION_ID, TENANT_ID, SERVER_KEY)

			// Cancellation succeeded at the gateway, so this stays a soft failure — no hard error.
			assert.NoError(t, err)
			assert.Equal(t, model.PaymentStatusCancelled, response.PaymentStatus)
			assert.Contains(t, response.Message, "payment cancelled successfully")
			assert.Contains(t, response.Message, "db write failed")
		})

		t.Run("Success_NoMessage", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			paymentProviderMock := NewPaymentProviderMock(&mock.Mock{}).(*PaymentProviderMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProviderMock)

			const TRANSACTION_ID = "TRX-SUCCESS-1"

			expectedResponse := model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatusCancelled,
				StatusCode:    "200",
			}

			paymentProviderMock.Mock.On("CancelTransaction", mock.Anything, TRANSACTION_ID).
				Return(expectedResponse, nil)
			orderItemRepo.Mock.On("SetPaymentStatus", 0, TRANSACTION_ID, model.PaymentStatusCancelled).
				Return(nil)

			response, err := orderItemService.CancelTransaction(0, TRANSACTION_ID, TENANT_ID, SERVER_KEY)

			assert.NoError(t, err)
			assert.Equal(t, model.PaymentStatusCancelled, response.PaymentStatus)
			assert.Equal(t, "200", response.StatusCode)
			assert.Empty(t, response.Message)
			orderItemRepo.Mock.AssertExpectations(t)
			paymentProviderMock.Mock.AssertExpectations(t)
		})
	})

	t.Run("DeleteInvoice", func(t *testing.T) {
		paymentProvider := NewPaymentProviderImpl()

		t.Run("SuccessCase", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

			const ORDER_ITEM_ID = 1

			orderItemRepo.Mock.On("DeleteInvoice", ORDER_ITEM_ID, TENANT_ID).Return(nil)

			err := orderItemService.DeleteInvoice(ORDER_ITEM_ID, TENANT_ID)

			assert.NoError(t, err)
			orderItemRepo.Mock.AssertExpectations(t)
		})

		t.Run("InvalidOrderItemId_Zero", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

			err := orderItemService.DeleteInvoice(0, TENANT_ID)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid order item id")
			// Repository should never be called
			orderItemRepo.Mock.AssertNotCalled(t, "DeleteInvoice", mock.Anything, mock.Anything)
		})

		t.Run("InvalidOrderItemId_Negative", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

			err := orderItemService.DeleteInvoice(-1, TENANT_ID)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid order item id")
			orderItemRepo.Mock.AssertNotCalled(t, "DeleteInvoice", mock.Anything, mock.Anything)
		})

		t.Run("InvalidTenantId_Zero", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

			err := orderItemService.DeleteInvoice(1, 0)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid tenant id")
			orderItemRepo.Mock.AssertNotCalled(t, "DeleteInvoice", mock.Anything, mock.Anything)
		})

		t.Run("InvalidTenantId_Negative", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

			err := orderItemService.DeleteInvoice(1, -1)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid tenant id")
			orderItemRepo.Mock.AssertNotCalled(t, "DeleteInvoice", mock.Anything, mock.Anything)
		})

		t.Run("NotFound", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

			const ORDER_ITEM_ID = 999999

			orderItemRepo.Mock.On("DeleteInvoice", ORDER_ITEM_ID, TENANT_ID).
				Return(fmt.Errorf("order item %d not found", ORDER_ITEM_ID))

			err := orderItemService.DeleteInvoice(ORDER_ITEM_ID, TENANT_ID)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
			orderItemRepo.Mock.AssertExpectations(t)
		})

		t.Run("RepositoryError", func(t *testing.T) {
			orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
			orderItemService := NewOrderItemServiceImpl(orderItemRepo, paymentProvider)

			const ORDER_ITEM_ID = 1

			orderItemRepo.Mock.On("DeleteInvoice", ORDER_ITEM_ID, TENANT_ID).
				Return(errors.New("database connection failed"))

			err := orderItemService.DeleteInvoice(ORDER_ITEM_ID, TENANT_ID)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "database connection failed")
			orderItemRepo.Mock.AssertExpectations(t)
		})
	})
}
