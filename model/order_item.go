package model

import (
	"time"
)

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

/*
Go automatically checks if OrderItem has a custom UnmarshalJSON method.
If it does, that method is called instead of the default unmarshaling.
https://pkg.go.dev/encoding/json#example-package-CustomMarshalJSON
*/
/*
func (orderItem *OrderItem) UnmarshalJSON(data []byte) error {
	var temp struct {
		Id          int `json:"id"`
		OrderItemId int `json:"order_item_id"`

		PurchasedPrice          int `json:"purchased_price"`
		OrderItemPurchasedPrice int `json:"order_item_purchased_price"`

		Subtotal          int `json:"subtotal"`
		OrderItemSubtotal int `json:"order_item_subtotal"`

		TotalQuantity          int `json:"total_quantity"`
		OrderItemTotalQuantity int `json:"order_item_total_quantity"`

		TotalAmount          int `json:"total_amount"`
		OrderItemTotalAmount int `json:"order_item_total_amount"`

		DiscountAmount          int `json:"discount_amount"`
		OrderItemDiscountAmount int `json:"order_item_discount_amount"`

		StoreId          int `json:"store_id"`
		OrderItemStoreId int `json:"order_item_store_id"`

		TenantId          int `json:"tenant_id"`
		OrderItemTenantId int `json:"order_item_tenant_id"`

		CreatedAt          *time.Time `json:"created_at,omitempty"`
		OrderItemCreatedAt *time.Time `json:"order_item_created_at,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Prefer order_item_* versions if non-zero, otherwise use non-prefixed versions
	if temp.OrderItemId != 0 {
		orderItem.Id = temp.OrderItemId
	} else {
		orderItem.Id = temp.Id
	}

	if temp.OrderItemPurchasedPrice != 0 {
		orderItem.PurchasedPrice = temp.OrderItemPurchasedPrice
	} else {
		orderItem.PurchasedPrice = temp.PurchasedPrice
	}

	if temp.OrderItemSubtotal != 0 {
		orderItem.Subtotal = temp.OrderItemSubtotal
	} else {
		orderItem.Subtotal = temp.Subtotal
	}

	if temp.OrderItemTotalQuantity != 0 {
		orderItem.TotalQuantity = temp.OrderItemTotalQuantity
	} else {
		orderItem.TotalQuantity = temp.TotalQuantity
	}

	if temp.OrderItemTotalAmount != 0 {
		orderItem.TotalAmount = temp.OrderItemTotalAmount
	} else {
		orderItem.TotalAmount = temp.TotalAmount
	}

	if temp.OrderItemDiscountAmount != 0 {
		orderItem.DiscountAmount = temp.OrderItemDiscountAmount
	} else {
		orderItem.DiscountAmount = temp.DiscountAmount
	}

	if temp.OrderItemStoreId != 0 {
		orderItem.StoreId = temp.OrderItemStoreId
	} else {
		orderItem.StoreId = temp.StoreId
	}

	if temp.OrderItemTenantId != 0 {
		orderItem.TenantId = temp.OrderItemTenantId
	} else {
		orderItem.TenantId = temp.TenantId
	}

	if temp.OrderItemCreatedAt != nil {
		orderItem.CreatedAt = temp.OrderItemCreatedAt
	} else if temp.CreatedAt != nil {
		orderItem.CreatedAt = temp.CreatedAt
	}

	return nil
}
*/
