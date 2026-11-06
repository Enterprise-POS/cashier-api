package model

import "time"

type VoucherUsage struct {
	Id          int       `json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id"`
	VoucherId   int       `json:"voucher_id" gorm:"column:voucher_id;not null"`
	OrderItemId int       `json:"order_item_id" gorm:"column:order_item_id;not null"`
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"column:created_at;autoCreateTime;<-:create"`

	Voucher   Voucher   `json:"-" gorm:"foreignKey:VoucherId;references:Id"`
	OrderItem OrderItem `json:"-" gorm:"foreignKey:OrderItemId;references:Id"`
}

func (voucherUsage *VoucherUsage) TableName() string {
	return "voucher_usage"
}
