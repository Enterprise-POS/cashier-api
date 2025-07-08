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
	// Caution ! Because there are ItemId and CategoryId, it's different from Category
	// We don't specify omitempty because we only hope to get the data from this struct not inserting nor updating
	CategoryId int `json:"category_id"`

	CategoryName string `json:"category_name"`

	// CreatedAt is Category property
	// CreatedAt *time.Time `json:"created_at"`

	// Item reference
	ItemId   int    `json:"item_id"`
	ItemName string `json:"item_name"`
	Stocks   int    `json:"stocks"`
}
