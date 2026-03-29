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

	/*
		Using aggregate function from SQL to get report
	*/
	GetSalesReport(tenantId int, storeId int, dateFilter *query.DateFilter) (*SalesReport, error)

	/*
		Get per-item profit data for Excel export
	*/
	GetProfitReport(tenantId int, storeId int, dateFilter *query.DateFilter) ([]*ProfitReportRow, error)

	/*
		Get tenant name and store name for display purposes.
		If storeId is 0, storeName will be "All Stores".
	*/
	GetTenantAndStoreName(tenantId int, storeId int) (tenantName string, storeName string, err error)
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

type SalesReport struct {
	SumPurchasedPrice int `json:"sum_purchased_price"`
	SumTotalQuantity  int `json:"sum_total_quantity"`
	SumTotalAmount    int `json:"sum_total_amount"`
	SumDiscountAmount int `json:"sum_discount_amount"`
	SumSubtotal       int `json:"sum_subtotal"`
	SumTransactions   int `json:"sum_transactions"`
	SumProfit         int `json:"sum_profit"`
}

type ProfitReportRow struct {
	ItemId        int    `json:"item_id"        gorm:"column:item_id"`
	ItemName      string `json:"item_name"      gorm:"column:item_name"`
	TotalQuantity int    `json:"total_quantity" gorm:"column:total_quantity"`
	TotalRevenue  int    `json:"total_revenue"  gorm:"column:total_revenue"`
	TotalCogs     int    `json:"total_cogs"     gorm:"column:total_cogs"`
	TotalDiscount int    `json:"total_discount" gorm:"column:total_discount"`
	TotalProfit   int    `json:"total_profit"   gorm:"column:total_profit"`
}

type ProfitReportRow struct {
	ItemId        int    `json:"item_id"        gorm:"column:item_id"`
	ItemName      string `json:"item_name"      gorm:"column:item_name"`
	TotalQuantity int    `json:"total_quantity" gorm:"column:total_quantity"`
	TotalRevenue  int    `json:"total_revenue"  gorm:"column:total_revenue"`
	TotalCogs     int    `json:"total_cogs"     gorm:"column:total_cogs"`
	TotalDiscount int    `json:"total_discount" gorm:"column:total_discount"`
	TotalProfit   int    `json:"total_profit"   gorm:"column:total_profit"`
}
