package repository

import (
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type StoreStockRepositoryMock struct {
	Mock *mock.Mock
}

func NewStoreStockRepositoryMock(mock *mock.Mock) StoreStockRepository {
	return &StoreStockRepositoryMock{Mock: mock}
}

// Get implements StoreStockRepository.
func (repository *StoreStockRepositoryMock) Get(tenantId int, storeId int, limit int, page int) ([]*model.StoreStock, int, error) {
	args := repository.Mock.Called(tenantId, storeId, limit, page)

	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}

	return args.Get(0).([]*model.StoreStock), args.Int(1), nil
}

// GetV2 implements StoreStockRepository.
func (repository *StoreStockRepositoryMock) GetV2(tenantId int, storeId int, limit int, page int, nameQuery string) ([]*model.StoreStockV2, int, error) {
	args := repository.Mock.Called(tenantId, storeId, limit, page)

	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}

	return args.Get(0).([]*model.StoreStockV2), args.Int(1), nil
}

// TransferStockToStoreStock implements StoreStockRepository.
func (repository *StoreStockRepositoryMock) TransferStockToStoreStock(quantity int, itemId int, storeId int, tenantId int) error {
	args := repository.Mock.Called(quantity, itemId, storeId, tenantId)

	if args.Get(0) == nil {
		return nil
	} else {
		return args.Error(0)
	}
}

// TransferStockToWarehouse implements StoreStockRepository.
func (repository *StoreStockRepositoryMock) TransferStockToWarehouse(quantity int, itemId int, storeId int, tenantId int) error {
	args := repository.Mock.Called(quantity, itemId, storeId, tenantId)

	if args.Get(0) == nil {
		return nil
	} else {
		return args.Error(0)
	}
}
