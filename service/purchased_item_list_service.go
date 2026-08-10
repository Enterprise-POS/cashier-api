package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
)

type PurchasedItemService interface {
	/*
		Get purchased_item_list logs for one or more item IDs.
	*/
	PurchasedItemListLogs(tenantId int, storeId int, itemIds []int, limit int, page int, dateFilter *query.DateFilter) ([]*model.PurchasedItem, int, error)
}
