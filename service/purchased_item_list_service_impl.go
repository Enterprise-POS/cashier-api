package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
)

type PurchasedItemServiceImpl struct {
	Repository repository.PurchasedItemRepository
}

func NewPurchasedItemServiceImpl(repository repository.PurchasedItemRepository) PurchasedItemService {
	return &PurchasedItemServiceImpl{Repository: repository}
}

// PurchasedItemListLogs implements PurchasedItemService.
func (service *PurchasedItemServiceImpl) PurchasedItemListLogs(
	tenantId int,
	storeId int,
	itemIds []int,
	limit int,
	page int,
	dateFilter *query.DateFilter,
) ([]*model.PurchasedItem, int, error) {
	if tenantId <= 0 {
		return nil, 0, errors.New("Tenant id is Required !")
	}
	if storeId < 0 {
		return nil, 0, fmt.Errorf("Given store id value is not allowed. storeId: %d", storeId)
	}
	// Will let user see logs more flexible
	// if len(itemIds) == 0 {
	// 	return nil, 0, errors.New("At least one item_id is required")
	// }
	if len(itemIds) > 100 {
		return nil, 0, errors.New("Too many item_id values (max 100)")
	}
	for _, itemId := range itemIds {
		if itemId <= 0 {
			return nil, 0, fmt.Errorf("invalid item_id: %d", itemId)
		}
	}

	if limit < 0 {
		return nil, 0, fmt.Errorf("Limit could not less then 1 (limit >= 1). Given limit %d", limit)
	} else if limit > 1000 {
		return nil, 0, fmt.Errorf("Limit could not greater then 1000. Given limit %d", limit)
	} else if limit == 0 {
		limit = 10
	}

	if page < 1 {
		return nil, 0, fmt.Errorf("page could not less then 1 (page >= 1). Given page %d", page)
	} else if page == 0 {
		page = 1
	}

	if dateFilter != nil {
		if dateFilter.StartDate != nil && dateFilter.EndDate != nil && *dateFilter.StartDate > *dateFilter.EndDate {
			return nil, 0, fmt.Errorf("Start date (%d) cannot be after end date (%d)", *dateFilter.StartDate, *dateFilter.EndDate)
		}
		if dateFilter.StartDate != nil && *dateFilter.StartDate < 0 {
			return nil, 0, fmt.Errorf("Invalid start date timestamp: %d", *dateFilter.StartDate)
		}
		if dateFilter.EndDate != nil && *dateFilter.EndDate < 0 {
			return nil, 0, fmt.Errorf("Invalid emd date timestamp: %d", *dateFilter.EndDate)
		}

		maxTimestamp := int64(4102444800) // 2100-01-01 00:00:00 UTC
		if dateFilter.StartDate != nil && *dateFilter.StartDate > maxTimestamp {
			return nil, 0, fmt.Errorf("Start date is too far in the future: %d", *dateFilter.StartDate)
		}
		if dateFilter.EndDate != nil && *dateFilter.EndDate > maxTimestamp {
			return nil, 0, fmt.Errorf("End date is too far in the future: %d", *dateFilter.EndDate)
		}
	}

	return service.Repository.PurchasedItemListLogs(tenantId, storeId, itemIds, limit, page-1, dateFilter)
}
