package repository

import (
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type WarehouseRepositoryMock struct {
	Mock mock.Mock
}

func (repository *WarehouseRepositoryMock) Get(tenantId int, limit int, page int) ([]*model.Item, int, error) {
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

func (repository *WarehouseRepositoryMock) FindById(itemId int, tenantId int) (_ *model.Item, _ error) {
	panic("not implemented") // TODO: Implement
}

func (repository *WarehouseRepositoryMock) CreateItem(items []*model.Item) (_ []*model.Item, _ error) {
	panic("not implemented") // TODO: Implement
}

func (repository *WarehouseRepositoryMock) Edit(quantity int, item *model.Item) (_ error) {
	panic("not implemented") // TODO: Implement
}

/*
Deactivate/Activate item, not delete it from DB
*/
func (repository *WarehouseRepositoryMock) SetActivate(tenantId int, itemId int, setInto bool) (_ error) {
	panic("not implemented") // TODO: Implement
}
