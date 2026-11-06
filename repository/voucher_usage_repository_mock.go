package repository

import (
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type VoucherUsageRepositoryMock struct {
	Mock *mock.Mock
}

func NewVoucherUsageRepositoryMock(mockObject *mock.Mock) VoucherUsageRepository {
	return &VoucherUsageRepositoryMock{Mock: mockObject}
}

func (repository *VoucherUsageRepositoryMock) CreateVoucherUsages(usages []*model.VoucherUsage) error {
	args := repository.Mock.Called(usages)
	return args.Error(0)
}

func (repository *VoucherUsageRepositoryMock) FindByOrderItemId(orderItemId int) ([]*model.VoucherUsage, error) {
	args := repository.Mock.Called(orderItemId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.VoucherUsage), args.Error(1)
}
