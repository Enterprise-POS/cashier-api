package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type VoucherServiceMock struct {
	Mock *mock.Mock
}

func NewVoucherServiceMock(mockObject *mock.Mock) VoucherService {
	return &VoucherServiceMock{Mock: mockObject}
}

func (service *VoucherServiceMock) GetVouchers(tenantId, limit, page int, searchQuery string, queryFilters []*query.QueryFilter) ([]*model.Voucher, int, error) {
	args := service.Mock.Called(tenantId, limit, page, searchQuery, queryFilters)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}

	return args.Get(0).([]*model.Voucher), args.Int(1), args.Error(2)
}

func (service *VoucherServiceMock) CreateVoucher(voucher *model.Voucher) (*model.Voucher, error) {
	args := service.Mock.Called(voucher)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Voucher), args.Error(1)
}

func (service *VoucherServiceMock) DeleteVoucher(voucherId, tenantId int) error {
	args := service.Mock.Called(voucherId, tenantId)
	return args.Error(0)
}

func (service *VoucherServiceMock) PreviewVoucher(tenantId int, code string, items []*model.Item) (int, int, int, int, bool, error) {
	args := service.Mock.Called(tenantId, code, items)
	return args.Int(0), args.Int(1), args.Int(2), args.Int(3), args.Bool(4), args.Error(5)
}

func (service *VoucherServiceMock) CommitVoucherUsage(voucherId, tenantId int) error {
	args := service.Mock.Called(voucherId, tenantId)
	return args.Error(0)
}

func (service *VoucherServiceMock) UseVoucher(tenantId int, code string, items []*model.Item) (int, int, int, error) {
	args := service.Mock.Called(tenantId, code, items)
	return args.Int(0), args.Int(1), args.Int(2), args.Error(3)
}
