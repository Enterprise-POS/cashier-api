package repository

import "cashier-api/model"

type PurchasedItemListRepository interface {
	CreateList(data []*model.PurchasedItemList, itemId int, orderItemId int, withReturn bool) ([]*model.PurchasedItemList, error)
}
