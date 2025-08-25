package repository

import "cashier-api/model"

/*
package repository

	handle query logic
*/
type WarehouseRepository interface {
	/*
		Return all items from current requested tenantId
	*/
	Get(tenantId int, limit int, page int) ([]*model.Item, int, error) // 2nd return is the count of all data
	FindById(itemId int, tenantId int) (*model.Item, error)

	/*
		Create new item for current tenantId.
		Once the item created, will never be erased from DB, only soft delete is allowed
	*/
	CreateItem(items []*model.Item) ([]*model.Item, error)
	Edit(quantity int, item *model.Item) error

	/*
		Deactivate/Activate item, not delete it from DB
	*/
	SetActivate(tenantId, itemId int, setInto bool) error
}
