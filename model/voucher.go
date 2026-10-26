package model

import (
	"time"

	"gorm.io/gorm"
)

type VoucherType = string

const (
	FROM_TOTAL_LIMITED_PERCENTAGE       VoucherType = "FROM_TOTAL_LIMITED_PERCENTAGE"
	FROM_TOTAL_LIMITED_STATIC_DECREMENT VoucherType = "FROM_TOTAL_LIMITED_STATIC_DECREMENT"
	ITEM_COMBINATIONS                   VoucherType = "ITEM_COMBINATIONS"
	ITEM_QUANTITIES                     VoucherType = "ITEM_QUANTITIES"
)

type Voucher struct {
	Id          int            `json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id"`
	Code        string         `json:"code" gorm:"column:code;uniqueIndex:unique_voucher_tenant_code"`
	Name        string         `json:"name" gorm:"column:name;not null"`
	TenantId    int            `json:"tenant_id" gorm:"column:tenant_id"`
	CreatedAt   time.Time      `json:"created_at,omitempty" gorm:"column:created_at;autoCreateTime;<-:create"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index"`
	VoucherType VoucherType    `json:"voucher_type" gorm:"column:voucher_type"`
	UsedCount   int            `json:"used_count" gorm:"column:used_count"`
	UsageLimit  int            `json:"usage_limit" gorm:"column:usage_limit"`
	StartAt     *time.Time     `json:"start_at" gorm:"column:start_at"`
	EndAt       *time.Time     `json:"end_at" gorm:"column:end_at"`
	// UpdatedAt    time.Time      `json:"updated_at,omitempty" gorm:"column:updated_at"`

	// No.1 'FROM_TOTAL_LIMITED_PERCENTAGE'
	Percentage       int    `json:"percentage" gorm:"column:percentage"`
	MaximumDecrement int    `json:"maximum_decrement" gorm:"column:maximum_decrement"`
	RoundingPolicy   string `json:"rounding_policy" gorm:"column:rounding_policy"`

	// No.2 'FROM_TOTAL_LIMITED_STATIC_DECREMENT'
	Decrement int

	// No.3 'ITEM_COMBINATIONS'
	ItemIdCombinations []int `json:"item_id_combinations" gorm:"column:item_id_combinations;type:jsonb;serializer:json"`

	// No.4 'ITEM_QUANTITIES'
	ItemQuantities int `json:"item_quantities" gorm:"column:item_quantities"`
}

func (voucher *Voucher) TableName() string {
	return "voucher"
}

func IsValidVoucherType(voucherType VoucherType) bool {
	switch voucherType {
	case FROM_TOTAL_LIMITED_PERCENTAGE, FROM_TOTAL_LIMITED_STATIC_DECREMENT, ITEM_COMBINATIONS, ITEM_QUANTITIES:
		return true
	default:
		return false
	}
}
