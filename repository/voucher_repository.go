package repository

import "cashier-api/model"

type VoucherRepository interface {
	CreateVoucher(voucher *model.Voucher) (*model.Voucher, error)
	DeleteVoucher(voucherId, tenantId int) error
	FindVoucherByCode(code string, tenantId int) (*model.Voucher, error)
	IncrementVoucherUsedCount(voucherId, tenantId int) error
}
