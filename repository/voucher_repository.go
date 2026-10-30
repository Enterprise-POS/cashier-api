package repository

import (
	"cashier-api/helper/query"
	"cashier-api/model"
)

type VoucherRepository interface {
	GetVouchers(tenantId, limit, page int, searchQuery string, queryFilters []*query.QueryFilter) ([]*model.Voucher, int, error)
	CreateVoucher(voucher *model.Voucher) (*model.Voucher, error)
	FindVoucherById(voucherId, tenantId int) (*model.Voucher, error)
	UpdateVoucher(voucherId, tenantId int, updates *model.VoucherUpdate) (*model.Voucher, error)
	DeleteVoucher(voucherId, tenantId int) error
	FindVoucherByCode(code string, tenantId int) (*model.Voucher, error)
	IncrementVoucherUsedCount(voucherId, tenantId int) error
}
