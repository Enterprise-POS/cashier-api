package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
)

type OrderItemServiceImpl struct {
	Repository repository.OrderItemRepository
}

func NewOrderItemServiceImpl(repository repository.OrderItemRepository) OrderItemService {
	return &OrderItemServiceImpl{
		Repository: repository,
	}
}

// Get implements OrderItemService.
func (service *OrderItemServiceImpl) Get(
	tenantId int,
	storeId int,
	limit int,
	page int,
	filters []*query.QueryFilter,
	dateFilter *query.DateFilter,
) ([]*model.OrderItem, int, error) {
	if tenantId <= 0 || storeId <= 0 {
		return nil, 0, errors.New("Tenant id, Store id, User id is Required !")
	}

	if limit < 1 {
		return nil, 0, fmt.Errorf("Limit could not less then 1 (limit >= 1). Given limit %d", limit)
	}

	if page < 1 {
		return nil, 0, fmt.Errorf("page could not less then 1 (page >= 1). Given page %d", page)
	}

	// Date filter validation
	if dateFilter != nil {
		// Check if both dates are provided and start is after end
		if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
			if *dateFilter.StartDate > *dateFilter.EndDate {
				return nil, 0, fmt.Errorf("Start date (%d) cannot be after end date (%d)", *dateFilter.StartDate, *dateFilter.EndDate)
			}
		}

		// Check for negative timestamps (dates before 1970)
		if dateFilter.StartDate != nil && *dateFilter.StartDate < 0 {
			return nil, 0, fmt.Errorf("Invalid start date timestamp: %d", *dateFilter.StartDate)
		}
		if dateFilter.EndDate != nil && *dateFilter.EndDate < 0 {
			return nil, 0, fmt.Errorf("Invalid emd date timestamp: %d", *dateFilter.EndDate)
		}

		// Check for unreasonably far future dates (e.g., year 2100+)
		maxTimestamp := int64(4102444800) // 2100-01-01 00:00:00 UTC
		if dateFilter.StartDate != nil && *dateFilter.StartDate > maxTimestamp {
			return nil, 0, fmt.Errorf("Start date is too far in the future: %d", *dateFilter.StartDate)
		}
		if dateFilter.EndDate != nil && *dateFilter.EndDate > maxTimestamp {
			return nil, 0, fmt.Errorf("End date is too far in the future: %d", *dateFilter.EndDate)
		}

		// Check for user intentionally specify endDate bigger than startDate
		if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
			if *dateFilter.StartDate > *dateFilter.EndDate {
				return nil, 0, fmt.Errorf("Start date (%d) cannot be after end date (%d)", *dateFilter.StartDate, *dateFilter.EndDate)
			}
		}
	}

	orderItems, count, err := service.Repository.Get(tenantId, storeId, limit, page-1, filters, dateFilter)
	if err != nil {
		return nil, 0, err
	}

	return orderItems, count, nil
}

// PlaceOrderItem implements OrderItemService.
func (service *OrderItemServiceImpl) PlaceOrderItem(*model.OrderItem) (*model.OrderItem, error) {
	panic("unimplemented")
}

// Transactions implements OrderItemService.
func (service *OrderItemServiceImpl) Transactions(params *repository.CreateTransactionParams) (int, error) {
	if params.TenantId <= 0 || params.StoreId <= 0 || params.UserId <= 0 {
		return 0, errors.New("Tenant id, Store id, User id is Required !")
	}

	if len(params.Items) == 0 {
		return 0, errors.New("At least one item is required")
	}

	if len(params.Items) > 1000 {
		return 0, errors.New("Too many items (max 1000)")
	}

	var (
		calculatedSubTotal    = 0                 // Sum before discount
		calculatedDiscount    = 0                 // Total discount
		calculatedTotal       = 0                 // Sum after discount
		calculatedQuantity    = 0                 // Total quantity
		priceConsistencyCheck = make(map[int]int) // item_id -> price
	)
	for _, item := range params.Items {
		// Check price consistency for same item
		if existingPrice, exists := priceConsistencyCheck[item.ItemId]; exists {
			if existingPrice != item.PurchasedPrice {
				return 0, fmt.Errorf("Price mismatch for item_id %d: expected %d, got %d",
					item.ItemId, existingPrice, item.PurchasedPrice)
			}
		} else {
			priceConsistencyCheck[item.ItemId] = item.PurchasedPrice
		}

		if item.Quantity < 1 {
			return 0, fmt.Errorf("Given quantity %d, from item_id: %d. Quantity should never be <= 0", item.Quantity, item.Id)
		}

		// Calculate totals
		itemSubTotal := item.PurchasedPrice * item.Quantity
		itemDiscount := item.DiscountAmount * item.Quantity
		itemTotal := itemSubTotal - itemDiscount

		calculatedSubTotal += itemSubTotal
		calculatedDiscount += itemDiscount
		calculatedTotal += itemTotal
		calculatedQuantity += item.Quantity

		// Validate individual item total
		if item.TotalAmount != itemTotal {
			return 0, fmt.Errorf("Item %d total mismatch: expected %d, got %d",
				item.ItemId, itemTotal, item.TotalAmount)
		}
	}

	// Validate against provided totals
	if calculatedQuantity != params.TotalQuantity {
		return 0, fmt.Errorf("Total quantity mismatch: calculated %d, provided %d",
			calculatedQuantity, params.TotalQuantity)
	}

	if calculatedSubTotal != params.SubTotal {
		return 0, fmt.Errorf("Subtotal mismatch: calculated %d, provided %d",
			calculatedSubTotal, params.SubTotal)
	}

	if calculatedTotal != params.TotalAmount {
		return 0, fmt.Errorf("Total amount mismatch: calculated %d, provided %d",
			calculatedTotal, params.TotalAmount)
	}

	if calculatedDiscount != params.DiscountAmount {
		return 0, fmt.Errorf("Discount amount mismatch: calculated %d, provided %d",
			calculatedDiscount, params.DiscountAmount)
	}

	// Validate payment (if you track cash given)
	// Remove this if PurchasedPrice is just another name for TotalAmount
	if params.PurchasedPrice < params.TotalAmount {
		return 0, fmt.Errorf("Insufficient payment: need %d, got %d",
			params.TotalAmount, params.PurchasedPrice)
	}

	orderId, err := service.Repository.Transactions(params)

	if err != nil {
		return 0, fmt.Errorf("Failed to create transaction: %w", err)
	}

	return orderId, nil
}
