package repository

import (
	"cashier-api/helper/query"
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

func (repository *VoucherRepositoryImpl) GetVouchers(
	tenantId int,
	limit int,
	page int,
	searchQuery string,
	queryFilters []*query.QueryFilter,
) ([]*model.Voucher, int, error) {
	offset := page * limit
	baseQuery := func() *gorm.DB {
		q := repository.Client.Model(&model.Voucher{}).Where("tenant_id = ?", tenantId)
		if searchQuery != "" {
			nameSearchPattern := "%" + searchQuery + "%"
			codeSearchPattern := "%" + searchQuery + "%"
			q = q.Where(
				"(LOWER(name) LIKE LOWER(?) OR LOWER(code) LIKE LOWER(?))",
				nameSearchPattern,
				codeSearchPattern,
			)
		}

		hasCustomOrder := false
		for _, filter := range queryFilters {
			if filter == nil {
				continue
			}
			switch filter.Column {
			case query.CreatedAtColumn:
				hasCustomOrder = true
				if filter.Ascending {
					q = q.Order("voucher.created_at ASC")
				} else {
					q = q.Order("voucher.created_at DESC")
				}
			case query.VoucherNameColumn:
				hasCustomOrder = true
				if filter.Ascending {
					q = q.Order("voucher.name ASC")
				} else {
					q = q.Order("voucher.name DESC")
				}
			case query.VoucherCodeColumn:
				hasCustomOrder = true
				if filter.Ascending {
					q = q.Order("voucher.code ASC")
				} else {
					q = q.Order("voucher.code DESC")
				}
			}
			if hasCustomOrder {
				break
			}
		}
		if !hasCustomOrder {
			q = q.Order("voucher.created_at DESC")
		}

		return q
	}

	var totalCount int64
	if err := baseQuery().Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}
	if totalCount == 0 {
		return []*model.Voucher{}, 0, nil
	}

	var vouchers []*model.Voucher
	if err := baseQuery().Limit(limit).Offset(offset).Find(&vouchers).Error; err != nil {
		return nil, 0, err
	}

	return vouchers, int(totalCount), nil
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
