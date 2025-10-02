package repository

import (
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type WarehouseRepositoryMock struct {
	Mock mock.Mock
}

func (repository *WarehouseRepositoryMock) Get(tenantId int, limit int, page int, queryName string) ([]*model.Item, int, error) {
	// .Called doesn't mean equal the parameter here
	// so when for example testing tenantId not exist, tenantId = 0
	// then we hope the return from 'repository' is 'nil'
	// in that case args.Get(0) return 'nil'
	args := repository.Mock.Called(tenantId, limit, page)

	// if tenant id is not exist;
	// 0: return nil
	// 1: 0
	// 2: "(PGRST103) Requested range not satisfiable"
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}

	// Normal situation
	return args.Get(0).([]*model.Item), args.Int(1), nil
}

func (repository *WarehouseRepositoryMock) GetActiveItem(tenantId int, limit int, page int, nameQuery string) ([]*model.Item, int, error) {
	args := repository.Mock.Called(tenantId, limit, tenantId, nameQuery)

	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}

	// Normal situation
	return args.Get(0).([]*model.Item), args.Int(1), nil
}

func (repository *WarehouseRepositoryMock) FindById(itemId int, tenantId int) (*model.Item, error) {
	args := repository.Mock.Called(itemId, tenantId)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Item), nil
}

func (repository *WarehouseRepositoryMock) CreateItem(items []*model.Item) ([]*model.Item, error) {
	args := repository.Mock.Called(items)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*model.Item), nil
}

func (repository *WarehouseRepositoryMock) Edit(quantity int, item *model.Item) (_ error) {
	args := repository.Mock.Called(quantity, item)
	if args.Get(0) == nil {
		return nil
	}

	return args.Error(0)
}

/*
Deactivate/Activate item, not delete it from DB
*/
func (repository *WarehouseRepositoryMock) SetActivate(tenantId int, itemId int, setInto bool) (_ error) {
	args := repository.Mock.Called(tenantId, itemId, setInto)
	if args.Get(0) == nil {
		return nil
	}

	return args.Error(0)
}

func (repository *WarehouseRepositoryMock) FindCompleteById(itemId int, tenantId int) (*model.CategoryWithItem, error) {
	args := repository.Mock.Called(itemId, tenantId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.CategoryWithItem), nil
}
