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
