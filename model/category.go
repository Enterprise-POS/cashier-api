package model

import "time"

// This is called when only get the list of category from some tenant
type Category struct {
	Id           int        `json:"id,omitempty"`
	CategoryName string     `json:"category_name"`
	CreatedAt    *time.Time `json:"created_at"`
	TenantId     int        `json:"tenant_id"`
}

/*
CategoryWithItem:

	category table many to many warehouse table
	the table between name is category_mtm_warehouse
*/
type CategoryWithItem struct {
	Id           int    `json:"id,omitempty"`
	CategoryName string `json:"category_name"`

	// CreatedAt is Category property
	CreatedAt *time.Time `json:"created_at"`
	TenantId  int        `json:"tenant_id"`

	// Item reference
	ItemId   int    `json:"item_id,omitempty"`
	ItemName string `json:"item_name"`
	Stocks   int    `json:"stocks"`
	IsActive bool   `json:"is_active"`
}
