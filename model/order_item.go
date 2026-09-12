package model

import (
	"time"

	"gorm.io/gorm"
)

type PaymentType string

const (
	PaymentTypeCash    PaymentType = "CASH"
	PaymentTypeQRIS    PaymentType = "QRIS"
	PaymentTypeCard    PaymentType = "CARD"
	PaymentTypeEWallet PaymentType = "EWALLET"
	PaymentTypeOther   PaymentType = "OTHER"
)

type PaymentStatus string

const (
	PaymentStatusSuccess           PaymentStatus = "SUCCESS"
	PaymentStatusPending           PaymentStatus = "PENDING"
	PaymentStatusRefunded          PaymentStatus = "REFUNDED"
	PaymentStatusFailed            PaymentStatus = "FAILED"
	PaymentStatusExpired           PaymentStatus = "EXPIRED"
	PaymentStatusCancelled         PaymentStatus = "CANCELLED"
	PaymentStatusPartiallyRefunded PaymentStatus = "PARTIALLY_REFUNDED"
)

type OrderItem struct {
	Id             int            `json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id"`
	PurchasedPrice int            `json:"purchased_price" gorm:"column:purchased_price"`
	CreatedAt      time.Time      `json:"created_at,omitempty" gorm:"column:created_at"`
	TotalQuantity  int            `json:"total_quantity" gorm:"column:total_quantity"`
	TotalAmount    int            `json:"total_amount" gorm:"column:total_amount"`
	DiscountAmount int            `json:"discount_amount" gorm:"column:discount_amount"`
	Subtotal       int            `json:"subtotal" gorm:"column:subtotal"`
	StoreId        int            `json:"store_id" gorm:"column:store_id"`
	TenantId       int            `json:"tenant_id" gorm:"column:tenant_id"`
	PaymentType    PaymentType    `json:"payment_type" gorm:"column:payment_type"`
	PaymentStatus  PaymentStatus  `json:"payment_status" gorm:"column:payment_status"`
	TransactionId  string         `json:"transaction_id" gorm:"column:transaction_id"`
	PaymentURL     *string        `json:"payment_url"`
	PaymentToken   *string        `json:"payment_token"`
	DeletedAt      gorm.DeletedAt `json:"-"` // Soft delete
}

func (orderItem *OrderItem) TableName() string {
	return "order_item"
}

type OrderItemWithStore struct {
	Id             int           `json:"id"`
	PurchasedPrice int           `json:"purchased_price"`
	CreatedAt      time.Time     `json:"created_at"`
	TotalQuantity  int           `json:"total_quantity"`
	TotalAmount    int           `json:"total_amount"`
	DiscountAmount int           `json:"discount_amount"`
	Subtotal       int           `json:"subtotal"`
	StoreId        int           `json:"store_id"`
	TenantId       int           `json:"tenant_id"`
	PaymentType    PaymentType   `json:"payment_type"`
	PaymentStatus  PaymentStatus `json:"payment_status"`
	TransactionId  string        `json:"transaction_id"`
	PaymentURL     *string       `json:"payment_url"`
	PaymentToken   *string       `json:"payment_token"`

	StoreName        string `json:"store_name"` // Joined field
	StoreAddress     string `json:"store_address"`
	StorePhoneNumber string `json:"store_phone_number"`
}

type PaymentProviderResponse struct {
	Token         string   `json:"token"`
	RedirectURL   string   `json:"redirect_url"`
	StatusCode    string   `json:"status_code,omitempty"`
	ErrorMessages []string `json:"error_messages,omitempty"`
}

type PaymentStatusResponse struct {
	PaymentStatus PaymentStatus `json:"payment_status"`
	StatusCode    string        `json:"status_code"`
	Message       string        `json:"message"`
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
