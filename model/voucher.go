package model

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
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
	Code        string         `json:"code" gorm:"column:code;uniqueIndex:unique_voucher_tenant_code,where:deleted_at IS NULL"`
	Name        string         `json:"name" gorm:"column:name;not null"`
	TenantId    int            `json:"tenant_id" gorm:"column:tenant_id;uniqueIndex:unique_voucher_tenant_code,where:deleted_at IS NULL"`
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
	RoundingPolicy   string `json:"rounding_policy" gorm:"column:rounding_policy"` // By default will be ROUND

	// No.2 'FROM_TOTAL_LIMITED_STATIC_DECREMENT'
	Decrement int

	// No.3 'ITEM_COMBINATIONS'
	ItemIdCombinations VoucherItemIDs `json:"item_id_combinations" gorm:"column:item_id_combinations;type:integer[]"`

	// No.4 'ITEM_QUANTITIES'
	ItemQuantities int `json:"item_quantities" gorm:"column:item_quantities"`
}

type VoucherItemIDs []int

func (itemIds VoucherItemIDs) Value() (driver.Value, error) {
	if itemIds == nil {
		return nil, nil
	}

	values := make([]string, len(itemIds))
	for index, itemId := range itemIds {
		values[index] = strconv.Itoa(itemId)
	}
	return "{" + strings.Join(values, ",") + "}", nil
}

func (itemIds *VoucherItemIDs) Scan(value interface{}) error {
	if value == nil {
		*itemIds = nil
		return nil
	}

	var array string
	switch typedValue := value.(type) {
	case string:
		array = typedValue
	case []byte:
		array = string(typedValue)
	default:
		return fmt.Errorf("cannot scan PostgreSQL integer array from %T", value)
	}

	if len(array) < 2 || array[0] != '{' || array[len(array)-1] != '}' {
		return fmt.Errorf("invalid PostgreSQL integer array: %q", array)
	}

	contents := array[1 : len(array)-1]
	if contents == "" {
		*itemIds = VoucherItemIDs{}
		return nil
	}

	result := make(VoucherItemIDs, 0)
	for _, item := range strings.Split(contents, ",") {
		itemId, err := strconv.Atoi(strings.TrimSpace(item))
		if err != nil {
			return fmt.Errorf("invalid PostgreSQL integer array item %q: %w", item, err)
		}
		result = append(result, itemId)
	}
	*itemIds = result
	return nil
}

type VoucherUpdate struct {
	Name       *string
	Code       *string
	StartAt    *time.Time
	StartAtSet bool
	EndAt      *time.Time
	EndAtSet   bool
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
