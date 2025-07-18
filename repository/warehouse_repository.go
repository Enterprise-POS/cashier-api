package repository

import "cashier-api/model"

/*
package repository

	handle query logic
*/
type WarehouseRepository interface {
	Get(tenantId int, limit int, page int) ([]*model.Item, int, error) // 2nd return is the count of all data
	FindById(itemId int, tenantId int) (*model.Item, error)
	CreateItem(items []*model.Item) ([]*model.Item, error)
	Edit(quantity int, item *model.Item) error

	/*
		Deactivate/Activate item, not delete it from DB
	*/
	SetActivate(tenantId, itemId int, setInto bool) error
}
