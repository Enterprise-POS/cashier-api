package repository

import "cashier-api/model"

type WarehouseRepository interface {
	FindById(itemId int, tenantId int) *model.Item
	CreateItem(item *model.Item) error
	Edit(quantity int, item *model.Item) error
}
