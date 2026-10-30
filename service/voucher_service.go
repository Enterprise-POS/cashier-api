package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
)

type VoucherService interface {
	GetVouchers(tenantId, limit, page int, searchQuery string, queryFilters []*query.QueryFilter) ([]*model.Voucher, int, error)
	// CreateVoucher validates and creates a voucher. Every voucher requires
	// TenantId,
	// a non-empty Code,
	// and a valid VoucherType.
	// Name may be empty, but cannot exceed 255 characters.
	//
	// UsedCount and CreatedAt must be zero;
	//
	// UsageLimit must be non-negative (zero means unlimited).
	// StartAt and EndAt are optional independently, but EndAt must be later than StartAt when both are set.
	//
	// Type-specific properties:
	//   - FROM_TOTAL_LIMITED_PERCENTAGE:
	// 			1. Percentage (1-100),
	// 			2. RoundingPolicy ("CEIL", "ROUND", or "FLOOR"),
	//      3. MaximumDecrement (0 or greater; zero means no maximum).
	//   - FROM_TOTAL_LIMITED_STATIC_DECREMENT:
	// 			1. Decrement (greater than zero).
	//   - ITEM_COMBINATIONS:
	// 			1. Decrement (greater than zero).
	// 			2. ItemIdCombinations must be non-empty and contain positive item IDs.
	//     			Repeated IDs specify required quantities (for example, [1, 1, 1]
	//     			requires three units of item 1). The cart must match the configured
	//     			item IDs and quantities exactly; the decrement applies once.
	//   - ITEM_QUANTITIES:
	// 			1. Decrement (greater than zero)
	// 			2. ItemQuantities (greater than zero), and exactly one positive ID in ItemIdCombinations.
	// Returns the created voucher with database-generated fields populated.
	CreateVoucher(voucher *model.Voucher) (*model.Voucher, error)
	DeleteVoucher(voucherId, tenantId int) error
	PreviewVoucher(tenantId int, code string, items []*model.Item) (beforeDiscountPrice, discountPrice, afterDiscountPrice, voucherId int, eligible bool, err error)
	CommitVoucherUsage(voucherId, tenantId int) error
	UseVoucher(tenantId int, code string, items []*model.Item) (beforeDiscountPrice, discountPrice, afterDiscountPrice int, err error)
}
