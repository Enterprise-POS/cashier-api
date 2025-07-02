package repository

import "cashier-api/model"

type OrderItemRepository interface {
	PlaceOrderItem(*model.OrderItem) (*model.OrderItem, error)
	// Get(tenantId int, limit int, page int) ([]*model.Item, int, error) // 2nd params return is the count of all data
	// FindById(itemId int, tenantId int) *model.Item
	// CreateItem(item []*model.Item) error
	// Edit(quantity int, item *model.Item) error
}
