package repository

import (
	"cashier-api/model"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrVoucherNotFound         = errors.New("VOUCHER_NOT_FOUND")
	ErrVoucherUsageUnavailable = errors.New("VOUCHER_USAGE_LIMIT_REACHED_OR_UNAVAILABLE")
)

func NewVoucherRepositoryImpl(client *gorm.DB) VoucherRepository {
	return &VoucherRepositoryImpl{
		Client: client,
	}
}

type VoucherRepositoryImpl struct {
	Client *gorm.DB
}

func (repository *VoucherRepositoryImpl) CreateVoucher(voucher *model.Voucher) (*model.Voucher, error) {
	if err := repository.Client.Create(voucher).Error; err != nil {
		return nil, err
	}

	return voucher, nil
}

func (repository *VoucherRepositoryImpl) DeleteVoucher(voucherId, tenantId int) error {
	result := repository.Client.
		Where("id = ? AND tenant_id = ?", voucherId, tenantId).
		Delete(&model.Voucher{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrVoucherNotFound
	}

	return nil
}

func (repository *VoucherRepositoryImpl) FindVoucherByCode(code string, tenantId int) (*model.Voucher, error) {
	var voucher model.Voucher
	if err := repository.Client.
		Where("code = ? AND tenant_id = ?", code, tenantId).
		First(&voucher).Error; err != nil {
		return nil, err
	}

	return &voucher, nil
}

func (repository *VoucherRepositoryImpl) IncrementVoucherUsedCount(voucherId, tenantId int) error {
	result := repository.Client.Model(&model.Voucher{}).
		Where("id = ? AND tenant_id = ? AND (usage_limit = 0 OR used_count < usage_limit)", voucherId, tenantId).
		UpdateColumn("used_count", gorm.Expr("used_count + ?", 1))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrVoucherUsageUnavailable
	}

	return nil
}
