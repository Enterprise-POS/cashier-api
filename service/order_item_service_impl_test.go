package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderItemServiceImpl(t *testing.T) {
	orderItemRepo := repository.NewOrderItemRepositoryMock(&mock.Mock{}).(*repository.OrderItemRepositoryMock)
	orderItemService := NewOrderItemServiceImpl(orderItemRepo)

	const USER_ID = 1
	const TENANT_ID = 1
	const STORE_ID = 1

	t.Run("Transactions", func(t *testing.T) {
		t.Run("NormalTransactions", func(t *testing.T) {
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 10_000,
				TotalQuantity:  1,
				TotalAmount:    10_000,
				DiscountAmount: 0,
				SubTotal:       10_000,

				Items: []*model.PurchasedItemList{
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

				Items: []*model.PurchasedItemList{},
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

				Items: []*model.PurchasedItemList{
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

				Items: []*model.PurchasedItemList{
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

				Items: []*model.PurchasedItemList{
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

				Items: []*model.PurchasedItemList{
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

				Items: []*model.PurchasedItemList{
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

				Items: []*model.PurchasedItemList{
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

				Items: []*model.PurchasedItemList{
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

				Items: []*model.PurchasedItemList{
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
			items := make([]*model.PurchasedItemList, 1001)
			for i := 0; i < 1001; i++ {
				items[i] = &model.PurchasedItemList{
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
			items := make([]*model.PurchasedItemList, 1000)
			for i := 0; i < 1000; i++ {
				items[i] = &model.PurchasedItemList{
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

				Items: []*model.PurchasedItemList{
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

				Items: []*model.PurchasedItemList{
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
