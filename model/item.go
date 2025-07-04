package model

import "time"

/*
Item (Warehouse Row)

	tag: json:"name" is required because supabase properties return json with snake_case

	omitempty: tell supabase that this is auto increment id so don't need to specify.
	if don't that it will generate id with 0, then duplicate key will occurred

	CreatedAt: *time.Time, pointer data type is a must, otherwise it will insert as 0 UTC
*/
type Item struct {
	ItemId    int        `json:"item_id,omitempty"`
	ItemName  string     `json:"item_name"`
	Stocks    int        `json:"stocks"`
	TenantId  int        `json:"tenant_id"`
	IsActive  bool       `json:"is_active"`
	Category  int        `json:"category,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}
