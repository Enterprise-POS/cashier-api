package model

import "time"

/*
	You can't add row to purchased_item_list
	unless you create order_item first
*/

type PurchasedItem struct {
	Id               int        `json:"id,omitempty"`
	Quantity         int        `json:"quantity"`
	PurchasedPrice   int        `json:"purchased_price"`
	DiscountAmount   int        `json:"discount_amount"`
	TotalAmount      int        `json:"total_amount"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	ItemId           int        `json:"item_id"`
	ItemNameSnapshot string     `json:"item_name_snapshot"`
	OrderItemId      int        `json:"order_item_id"`
}
