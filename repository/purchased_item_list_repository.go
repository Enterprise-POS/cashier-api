package repository

import (
	"cashier-api/helper/query"
	"cashier-api/model"
)

// Always make sure order_item.id is really available
type PurchasedItemRepository interface {
	/*
		Before create list, make sure that orderItemId is really available, otherwise DB will by unsynchronized
	*/
	CreateList(data []*model.PurchasedItem, withReturn bool) ([]*model.PurchasedItem, error)

	/*
		Get the list, by order_item.id
	*/
	GetByOrderItemId(orderItemId int) ([]*model.PurchasedItem, error)

	/*
		Get purchased_item_list rows for one or more item IDs, including current warehouse
		item data and the order date.
	*/
	PurchasedItemListLogs(tenantId int, storeId int, itemIds []int, limit int, page int, dateFilter *query.DateFilter, filters []query.QueryFilter) ([]*model.PurchasedItem, int, error)
}
