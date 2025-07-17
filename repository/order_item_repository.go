package repository

import (
	"cashier-api/helper/query"
	"cashier-api/model"
)

type OrderItemRepository interface {
	/*
		When cashier app press the button, then
		this will called
	*/
	PlaceOrderItem(*model.OrderItem) (*model.OrderItem, error)

	/*
		Get the list of order_item, purchased_item_list will not included
		2nd params return is the count of all data
	*/
	Get(tenantId int, limit int, page int, filters []*query.QueryFilter) ([]*model.OrderItem, int, error)

	// FindById(itemId int, tenantId int) *model.Item
	// CreateItem(item []*model.Item) error
	// Edit(quantity int, item *model.Item) error
}
