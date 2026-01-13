package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/repository"

	"github.com/stretchr/testify/mock"
)

type OrderItemServiceMock struct {
	Mock *mock.Mock
}

func NewOrderItemServiceMock(mock *mock.Mock) OrderItemService {
	return &OrderItemServiceMock{Mock: mock}
}

// Get implements OrderItemService.
func (service *OrderItemServiceMock) Get(tenantId int, storeId int, limit int, page int, filters []*query.QueryFilter, dateFilter *query.DateFilter) ([]*model.OrderItem, int, error) {
	args := service.Mock.Called(tenantId, storeId, limit, page, filters, dateFilter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}

	return args.Get(0).([]*model.OrderItem), args.Int(1), nil
}

// PlaceOrderItem implements OrderItemService.
func (service *OrderItemServiceMock) PlaceOrderItem(*model.OrderItem) (*model.OrderItem, error) {
	panic("unimplemented")
}

// FindById implements OrderItemService.
func (service *OrderItemServiceMock) FindById(orderItemid int, tenantId int) (*model.OrderItem, []*model.PurchasedItem, error) {
	args := service.Mock.Called(orderItemid, tenantId)
	if args.Get(0) == nil || args.Get(1) == nil {
		return nil, nil, args.Error(2)
	}

	return args.Get(0).(*model.OrderItem), args.Get(1).([]*model.PurchasedItem), nil
}

// Transactions implements OrderItemService.
func (service *OrderItemServiceMock) Transactions(params *repository.CreateTransactionParams) (int, error) {
	args := service.Mock.Called(params)
	if args.Int(0) == 0 {
		return 0, args.Error(1)
	}

	return args.Int(0), nil
}

// GetSalesReport implements OrderItemService.
func (service *OrderItemServiceMock) GetSalesReport(tenantId int, storeId int, dateFilter *query.DateFilter) (*repository.SalesReport, error) {
	args := service.Mock.Called(tenantId, storeId, dateFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*repository.SalesReport), nil
}
