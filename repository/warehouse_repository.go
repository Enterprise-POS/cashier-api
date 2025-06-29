package repository

import "cashier-api/model"

type WarehouseRepository interface {
	Get(limit int, page int) []*model.Item
	FindById(itemId int, tenantId int) *model.Item
	CreateItem(item []*model.Item) error
	Edit(quantity int, item *model.Item) error
}
