package model

import "time"

type StoreStock struct {
	Id        int        `json:"id,omitempty"         gorm:"primaryKey;autoIncrement;column:id"`
	Stocks    int        `json:"stocks"               gorm:"column:stocks"`
	Price     int        `json:"price"                gorm:"column:price"`
	CreatedAt *time.Time `json:"created_at,omitempty" gorm:"column:created_at;<-:create"`
	ItemId    int        `json:"item_id"              gorm:"column:item_id"`
	TenantId  int        `json:"tenant_id"            gorm:"column:tenant_id"`
	StoreId   int        `json:"store_id"             gorm:"column:store_id"`
}

func (StoreStock) TableName() string {
	return "store_stock"
}

/*
When GET the data from store_stock. Sometimes
we want the name as well
*/
type StoreStockV2 struct {
	Id           int        `json:"id,omitempty"         gorm:"column:id"`
	ItemName     string     `json:"item_name"            gorm:"column:item_name"`
	Stocks       int        `json:"stocks"               gorm:"column:stocks"`
	Price        int        `json:"price"                gorm:"column:price"`
	StockType    StockType  `json:"stock_type"           gorm:"column:stock_type"`
	CreatedAt    *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	ItemId       int        `json:"item_id"              gorm:"column:item_id"`
	TotalCount   int        `json:"total_count"          gorm:"column:total_count"`
	BasePrice    int        `json:"base_price"           gorm:"column:base_price"`
	IsActive     bool       `json:"is_active"            gorm:"column:is_active"`
	CategoryId   int        `json:"category_id"          gorm:"column:category_id"`
	CategoryName string     `json:"category_name"        gorm:"column:category_name"`
}

/*
Special struct for load data at cashier app
*/
type CashierData struct {
	CategoryId   int    `json:"category_id"   gorm:"column:category_id"`
	CategoryName string `json:"category_name" gorm:"column:category_name"`

	ItemId    int       `json:"item_id"    gorm:"column:item_id"`
	ItemName  string    `json:"item_name"  gorm:"column:item_name"`
	Stocks    int       `json:"stocks"     gorm:"column:stocks"`
	StockType StockType `json:"stock_type" gorm:"column:stock_type"`
	IsActive  bool      `json:"is_active"  gorm:"column:is_active"`
	BasePrice int       `json:"base_price" gorm:"column:base_price"`

	StoreStockId     int `json:"store_stock_id"     gorm:"column:store_stock_id"`
	StoreStockStocks int `json:"store_stock_stocks" gorm:"column:store_stock_stocks"`
	StoreStockPrice  int `json:"store_stock_price"  gorm:"column:store_stock_price"`
}
