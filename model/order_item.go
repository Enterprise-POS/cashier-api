package model

import "time"

type OrderItem struct {
	Id             int        `json:"id,omitempty"`
	PurchasedPrice int        `json:"purchased_price"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	TotalQuantity  int        `json:"total_quantity"`
	TotalAmount    int        `json:"total_amount"`
	DiscountAmount int        `json:"discount_amount"`
	Subtotal       int        `json:"subtotal"`
	StoreId        int        `json:"store_id"`
	TenantId       int        `json:"tenant_id"`
}
