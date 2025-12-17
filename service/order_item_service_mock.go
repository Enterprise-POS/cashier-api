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
func (service *OrderItemServiceMock) Get(tenantId int, limit int, page int, filters []*query.QueryFilter) ([]*model.OrderItem, int, error) {
	panic("unimplemented")
}

// PlaceOrderItem implements OrderItemService.
func (service *OrderItemServiceMock) PlaceOrderItem(*model.OrderItem) (*model.OrderItem, error) {
	panic("unimplemented")
}

// Transactions implements OrderItemService.
func (service *OrderItemServiceMock) Transactions(params *repository.CreateTransactionParams) (int, error) {
	args := service.Mock.Called(params)
	if args.Int(0) == 0 {
		return 0, args.Error(1)
	}

	return args.Int(0), nil
}
