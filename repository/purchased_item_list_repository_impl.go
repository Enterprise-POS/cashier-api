package repository

import (
	common "cashier-api/helper"
	"cashier-api/helper/query"
	"cashier-api/model"
	"errors"

	"gorm.io/gorm"
)

type PurchasedItemRepositoryImpl struct {
	Client *gorm.DB
}

func NewPurchasedItemRepositoryImpl(client *gorm.DB) PurchasedItemRepository {
	return &PurchasedItemRepositoryImpl{Client: client}
}

/*
CreateList:

	If we want insert 10 row and 1 row data violate / for example un-exist order_item_id
	then the supabase will fail all the 10 row,
	un exist`item_id will result error !
*/
func (repository *PurchasedItemRepositoryImpl) CreateList(data []*model.PurchasedItem, withReturn bool) ([]*model.PurchasedItem, error) {
	result := repository.Client.Create(&data)
	if result.Error != nil {
		return nil, result.Error
	}

	if withReturn {
		// GORM mutates data in-place, so it already contains id, created_at, etc.
		return data, nil
	}

	return nil, nil
}

func (repository *PurchasedItemRepositoryImpl) GetByOrderItemId(orderItemId int) ([]*model.PurchasedItem, error) {
	var result []*model.PurchasedItem
	if err := repository.Client.Where("order_item_id = ?", orderItemId).Find(&result).Error; err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, errors.New("fatal error list of purchased item not available")
	}

	return result, nil
}

// PurchasedItemListLogs implements PurchasedItemRepository.
func (repository *PurchasedItemRepositoryImpl) PurchasedItemListLogs(
	tenantId int,
	storeId int,
	itemIds []int,
	limit int,
	page int,
	dateFilter *query.DateFilter,
) ([]*model.PurchasedItem, int, error) {
	start := page * limit

	db := repository.Client.Table("purchased_item_list pil").
		Select(`
			warehouse.item_id,
			warehouse.item_name,
			warehouse.base_price,
			order_item.id AS order_item_id,
			order_item.created_at AS order_item_created_at,
			pil.id,
			pil.quantity,
			pil.store_price_snapshot,
			pil.item_name_snapshot,
			pil.base_price_snapshot,
			pil.discount_amount,
			pil.total_amount,
			order_item.store_id,
			order_item.tenant_id
		`).
		Joins("INNER JOIN warehouse ON warehouse.item_id = pil.item_id").
		Joins("INNER JOIN order_item ON order_item.id = pil.order_item_id AND order_item.deleted_at IS NULL").
		Where("order_item.tenant_id = ?", tenantId)

	if len(itemIds) > 0 {
		db = db.Where("pil.item_id IN ?", itemIds)
	}

	if storeId > 0 {
		db = db.Where("order_item.store_id = ?", storeId)
	}

	if dateFilter != nil {
		if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
			startDate := common.EpochToRFC3339(*dateFilter.StartDate)
			endDate := common.EpochToRFC3339(*dateFilter.EndDate)
			db = db.Where("order_item.created_at >= ? AND order_item.created_at < ?", startDate, endDate)
		} else if dateFilter.StartDate != nil {
			startDate := common.EpochToRFC3339(*dateFilter.StartDate)
			db = db.Where("order_item.created_at >= ?", startDate)
		} else if dateFilter.EndDate != nil {
			endDate := common.EpochToRFC3339(*dateFilter.EndDate)
			db = db.Where("order_item.created_at < ?", endDate)
		}
	}

	var totalCount int64
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	var rows []*model.PurchasedItem
	err := db.
		Order("order_item.created_at DESC, pil.id DESC").
		Limit(limit).
		Offset(start).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, int(totalCount), nil
}
