package repository

import (
	"cashier-api/model"
	"errors"

	"gorm.io/gorm"
)

type VoucherUsageRepository interface {
	CreateVoucherUsages(usages []*model.VoucherUsage) error
	FindByOrderItemId(orderItemId int) ([]*model.VoucherUsage, error)
}

type VoucherUsageRepositoryImpl struct {
	Client *gorm.DB
}

func NewVoucherUsageRepositoryImpl(client *gorm.DB) VoucherUsageRepository {
	return &VoucherUsageRepositoryImpl{Client: client}
}

func (repository *VoucherUsageRepositoryImpl) CreateVoucherUsages(usages []*model.VoucherUsage) error {
	if len(usages) == 0 {
		return nil
	}
	for _, usage := range usages {
		if usage == nil || usage.VoucherId < 1 || usage.OrderItemId < 1 {
			return errors.New("voucher ID and order item ID are required")
		}
	}

	return repository.Client.Create(&usages).Error
}

func (repository *VoucherUsageRepositoryImpl) FindByOrderItemId(orderItemId int) ([]*model.VoucherUsage, error) {
	if orderItemId < 1 {
		return nil, errors.New("order item ID is required")
	}

	var usages []*model.VoucherUsage
	if err := repository.Client.Where("order_item_id = ?", orderItemId).Find(&usages).Error; err != nil {
		return nil, err
	}
	return usages, nil
}
