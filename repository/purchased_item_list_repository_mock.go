package repository

import (
	"cashier-api/helper/query"
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type PurchasedItemRepositoryMock struct {
	Mock *mock.Mock
}

func NewPurchasedItemRepositoryMock(mock *mock.Mock) PurchasedItemRepository {
	return &PurchasedItemRepositoryMock{mock}
}

// CreateList implements [PurchasedItemRepository].
func (p *PurchasedItemRepositoryMock) CreateList(data []*model.PurchasedItem, withReturn bool) ([]*model.PurchasedItem, error) {
	panic("unimplemented")
}

// GetByOrderItemId implements [PurchasedItemRepository].
func (p *PurchasedItemRepositoryMock) GetByOrderItemId(orderItemId int) ([]*model.PurchasedItem, error) {
	panic("unimplemented")
}

// PurchasedItemListLogs implements [PurchasedItemRepository].
func (p *PurchasedItemRepositoryMock) PurchasedItemListLogs(tenantId int, storeId int, itemIds []int, limit int, page int, dateFilter *query.DateFilter, filters []query.QueryFilter) ([]*model.PurchasedItem, int, error) {
	args := p.Mock.Called(tenantId, storeId, itemIds, limit, page, dateFilter, filters)

	var result []*model.PurchasedItem
	if args.Get(0) != nil {
		result = args.Get(0).([]*model.PurchasedItem)
	}

	return result, args.Int(1), args.Error(2)
}
