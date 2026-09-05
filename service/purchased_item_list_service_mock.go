package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type PurchasedItemServiceMock struct {
	Mock *mock.Mock
}

// PurchasedItemListLogs implements [PurchasedItemService].
func (p *PurchasedItemServiceMock) PurchasedItemListLogs(tenantId int, storeId int, itemIds []int, limit int, page int, dateFilter *query.DateFilter, filters []query.QueryFilter) ([]*model.PurchasedItem, int, error) {
	args := p.Mock.Called(tenantId, storeId, itemIds, limit, page, dateFilter, filters)

	var result []*model.PurchasedItem
	if args.Get(0) != nil {
		result = args.Get(0).([]*model.PurchasedItem)
	}

	return result, args.Int(1), args.Error(2)
}

func NewPurchasedItemServiceMock(mock *mock.Mock) PurchasedItemService {
	return &PurchasedItemServiceMock{Mock: mock}
}
