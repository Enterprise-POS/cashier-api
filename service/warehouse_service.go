package service

import "cashier-api/model"

/*
package service

	typically handle the logic by user inputs
*/
type WarehouseService interface {
	/*
		Return all items from current requested tenantId
	*/
	GetWarehouseItems(tenantId, limit, page int) ([]*model.Item, int, error)

	/*
		Create new item for current tenantId.
		Once the item created, will never be erased from DB, only soft delete is allowed
	*/
	CreateItem(items []*model.Item) ([]*model.Item, error)

	/*
		Return detailed item information
	*/
	FindById(itemId int, tenantId int) (*model.Item, error)

	/*
		Edit/update some specific item quantities
	*/
	Edit(quantity int, item *model.Item) error

	/*
		Deactivate/Activate item, not delete it from DB
	*/
	SetActivate(tenantId, itemId int, setInto bool) error
}
