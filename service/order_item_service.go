package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/repository"
)

type OrderItemService interface {
	/*
		When cashier app press the button, then
		this will called
	*/
	PlaceOrderItem(*model.OrderItem) (*model.OrderItem, error)

	/*
		Get the list of order_item, purchased_item_list will not included
		2nd params return is the count of all data
	*/
	Get(tenantId, storeId, limit, page int, filters []*query.QueryFilter, dateFilter *query.DateFilter) ([]*model.OrderItem, int, error)

	/*
		Always minus page by 1 because PostgreSQL start index from 0
	*/
	FindById(orderItemid int, tenantId int) (*model.OrderItem, []*model.PurchasedItem, error)

	// CreateItem(item []*model.Item) error
	// Edit(quantity int, item *model.Item) error

	/*
		This method will insert into 2 table
	*/
	Transactions(params *repository.CreateTransactionParams) (int, error)
}
