package repository

import (
	"cashier-api/helper/query"
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type OrderItemRepositoryMock struct {
	Mock *mock.Mock
}

func NewOrderItemRepositoryMock(mock *mock.Mock) OrderItemRepository {
	return &OrderItemRepositoryMock{mock}
}

// Get implements OrderItemRepository.
func (repository *OrderItemRepositoryMock) Get(
	tenantId int,
	storeId int,
	limit int,
	page int,
	filters []*query.QueryFilter,
	dateFilter *query.DateFilter,
) ([]*model.OrderItem, int, error) {
	args := repository.Mock.Called(tenantId, storeId, limit, page, filters, dateFilter)

	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(1)
	} else {
		return args.Get(0).([]*model.OrderItem), args.Int(1), args.Error(2)
	}
}

// PlaceOrderItem implements OrderItemRepository.
func (repository *OrderItemRepositoryMock) PlaceOrderItem(*model.OrderItem) (*model.OrderItem, error) {
	panic("unimplemented")
}

// Transactions implements OrderItemRepository.
func (repository *OrderItemRepositoryMock) Transactions(params *CreateTransactionParams) (int, error) {
	args := repository.Mock.Called(params)

	if args.Int(0) == 0 {
		return 0, args.Error(1)
	}

	return args.Int(0), nil
}

// FindById implements OrderItemRepository.
func (repository *OrderItemRepositoryMock) FindById(itemId int, tenantId int) (*model.OrderItem, []*model.PurchasedItem, error) {
	args := repository.Mock.Called(itemId, tenantId)

	if args.Get(0) == nil || args.Get(1) == nil {
		return nil, nil, args.Error(2)
	}

	return args.Get(0).(*model.OrderItem), args.Get(1).([]*model.PurchasedItem), nil
}

// GetSalesReport implements OrderItemRepository.
func (repository *OrderItemRepositoryMock) GetSalesReport(tenantId int, storeId int, dateFilter *query.DateFilter) (*SalesReport, error) {
	args := repository.Mock.Called(tenantId, storeId, dateFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*SalesReport), nil
}
