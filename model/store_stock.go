package model

import "time"

type StoreStock struct {
	Id        int        `json:"id,omitempty"`
	Stocks    int        `json:"stocks"`
	Price     int        `json:"price"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	ItemId    int        `json:"item_id"`
	TenantId  int        `json:"tenant_id"`
	StoreId   int        `json:"store_id"`
}

/*
When GET the data from store_stock. Sometimes
we want the name as well
*/
type StoreStockV2 struct {
	Id         int        `json:"id,omitempty"`
	ItemName   string     `json:"item_name"`
	Stocks     int        `json:"stocks"` // StoreStock Stock
	Price      int        `json:"price"`
	StockType  StockType  `json:"stock_type"`
	CreatedAt  *time.Time `json:"created_at,omitempty"` // Warehouse Item created_at
	ItemId     int        `json:"item_id"`
	TotalCount int        `json:"total_count"`
}
