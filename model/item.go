package model

import "time"

/*
Item (Warehouse Row)

	tag: json:"name" is required because supabase properties return json with snake_case

	omitempty: tell supabase that this is auto increment id so don't need to specify.
	if don't that it will generate id with 0, then duplicate key will occurred

	CreatedAt: *time.Time, pointer data type is a must, otherwise it will insert as 0 UTC
*/
type StockType string

const (
	StockTypeUnlimited StockType = "UNLIMITED"
	StockTypeTracked   StockType = "TRACKED"
)

type Item struct {
	ItemId    int       `json:"item_id,omitempty" gorm:"primaryKey;autoIncrement;column:item_id"`
	ItemName  string    `json:"item_name" gorm:"column:item_name"`
	Stocks    int       `json:"stocks" gorm:"column:stocks"`
	StockType StockType `json:"stock_type" gorm:"column:stock_type"`
	BasePrice int       `json:"base_price" gorm:"column:base_price"`
	TenantId  int       `json:"tenant_id" gorm:"column:tenant_id"`
	IsActive  bool      `json:"is_active" gorm:"column:is_active"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"column:created_at;<-:create"`
}

func (Item) TableName() string {
	return "warehouse"
}
