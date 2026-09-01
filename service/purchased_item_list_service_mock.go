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
	panic("unimplemented")
}

func NewPurchasedItemServiceMock(mock *mock.Mock) PurchasedItemService {
	return &PurchasedItemServiceMock{Mock: mock}
}
