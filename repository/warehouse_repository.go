package repository

import "cashier-api/model"

type WarehouseRepository interface {
	Get(tenantId int, limit int, page int) ([]*model.Item, int, error) // 2nd return is the count of all data
	FindById(itemId int, tenantId int) *model.Item
	CreateItem(item []*model.Item) error
	Edit(quantity int, item *model.Item) error
}
