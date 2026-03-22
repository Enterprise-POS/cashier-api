package model

import "time"

/*
You can't add row to purchased_item_list
unless you create order_item first
*/
type PurchasedItem struct {
	Id                 int        `json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id"`
	Quantity           int        `json:"quantity" gorm:"column:quantity"`
	StorePriceSnapshot int        `json:"store_price_snapshot" gorm:"column:store_price_snapshot"`
	BasePriceSnapshot  int        `json:"base_price_snapshot" gorm:"column:base_price_snapshot"`
	DiscountAmount     int        `json:"discount_amount" gorm:"column:discount_amount"`
	TotalAmount        int        `json:"total_amount" gorm:"column:total_amount"`
	CreatedAt          *time.Time `json:"created_at,omitempty" gorm:"column:created_at;<-:create"`
	ItemId             int        `json:"item_id" gorm:"column:item_id"`
	ItemNameSnapshot   string     `json:"item_name_snapshot" gorm:"column:item_name_snapshot"`
	OrderItemId        int        `json:"order_item_id" gorm:"column:order_item_id"`

	// ! DEPRECATED, by default if this property is not defined then
	// ! the default value given by GO is 0 (if it's int)
	// PurchasedPrice int `json:"purchased_price"`
}

func (p *PurchasedItem) TableName() string {
	return "purchased_item_list"
}
