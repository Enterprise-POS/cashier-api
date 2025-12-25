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
	Get(tenantId int, storeId int, limit int, page int, filters []*query.QueryFilter, dateFilter *query.DateFilter) ([]*model.OrderItem, int, error)

	/*

	 */
	FindById(orderItemid int, tenantId int) (*model.OrderItem, []*model.PurchasedItem, error)

	/*
		This method will insert into 2 table
	*/
	Transactions(params *CreateTransactionParams) (int, error)

	// Edit(quantity int, item *model.Item) error
}

type CreateTransactionParams struct {
	// Order summary
	PurchasedPrice int `json:"purchased_price"`
	TotalQuantity  int `json:"total_quantity"`
	TotalAmount    int `json:"total_amount"`
	DiscountAmount int `json:"discount_amount"`
	SubTotal       int `json:"sub_total"`

	// Items
	Items []*model.PurchasedItem `json:"items"`

	// Validation/Context
	UserId   int `json:"user_id"`
	TenantId int `json:"tenant_id"`
	StoreId  int `json:"store_id"`
}
