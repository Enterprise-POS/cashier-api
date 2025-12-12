package repository

import (
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type StoreRepositoryMock struct {
	Mock *mock.Mock
}

func NewStoreRepositoryMock(mock *mock.Mock) StoreRepository {
	return &StoreRepositoryMock{Mock: mock}
}

// Create implements StoreRepository.
func (repository *StoreRepositoryMock) Create(tenantId int, name string) (*model.Store, error) {
	args := repository.Mock.Called(tenantId, name)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Store), nil
}

// GetAll implements StoreRepository.
func (repository *StoreRepositoryMock) GetAll(tenantId int, page int, limit int, includeNonActive bool) ([]*model.Store, int, error) {
	args := repository.Mock.Called(tenantId, page, limit, includeNonActive)

	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}

	return args.Get(0).([]*model.Store), args.Int(1), nil
}

// SetActivate implements StoreRepository.
func (repository *StoreRepositoryMock) SetActivate(tenantId int, storeId int, setInto bool) error {
	args := repository.Mock.Called(tenantId, storeId, setInto)

	if args.Get(0) == nil {
		return nil
	}

	return args.Error(0)
}

// Edit implements StoreRepository.
func (repository *StoreRepositoryMock) Edit(tobeEditStore *model.Store) (*model.Store, error) {
	args := repository.Mock.Called(tobeEditStore)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Store), args.Error(1)
}
