package repository

import "cashier-api/model"

// Always make sure order_item.id is really available
type PurchasedItemListRepository interface {
	/*
		Before create list, make sure that orderItemId is really available, otherwise DB will by unsynchronized
	*/
	CreateList(data []*model.PurchasedItemList, withReturn bool) ([]*model.PurchasedItemList, error)

	/*
		Get the list, by order_item.id
	*/
	GetByOrderItemId(orderItemId int) ([]*model.PurchasedItemList, error)
}
