package repository

import (
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type CategoryRepositoryMock struct {
	Mock *mock.Mock
}

func NewCategoryRepositoryMock(mock *mock.Mock) CategoryRepository {
	return &CategoryRepositoryMock{
		Mock: mock,
	}
}

// Create implements CategoryRepository.
func (repository *CategoryRepositoryMock) Create(tenantId int, categories []*model.Category) ([]*model.Category, error) {
	args := repository.Mock.Called(tenantId, categories)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*model.Category), nil
}

// Delete implements CategoryRepository.
func (repository *CategoryRepositoryMock) Delete(category *model.Category) error {
	args := repository.Mock.Called(category)

	if args.Get(0) == nil {
		return nil
	} else {
		return args.Error(0)
	}
}

// Get implements CategoryRepository.
func (repository *CategoryRepositoryMock) Get(tenantId int, page int, limit int, nameQuery string) ([]*model.Category, int, error) {
	args := repository.Mock.Called(tenantId, page, limit, nameQuery)

	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}

	return args.Get(0).([]*model.Category), args.Int(1), nil
}

// GetCategoryWithItems implements CategoryRepository.
func (repository *CategoryRepositoryMock) GetCategoryWithItems(tenantId int, page int, limit int) ([]*model.CategoryWithItem, int, error) {
	args := repository.Mock.Called(tenantId, page, limit)

	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}

	return args.Get(0).([]*model.CategoryWithItem), args.Int(1), nil
}

// GetItemsByCategoryId implements CategoryRepository.
func (repository *CategoryRepositoryMock) GetItemsByCategoryId(tenantId int, categoryId int, limit int, page int) ([]*model.CategoryWithItem, int, error) {
	args := repository.Mock.Called(tenantId, categoryId, limit, page)

	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(1)
	} else {
		return args.Get(0).([]*model.CategoryWithItem), args.Int(1), nil
	}
}

// Register implements CategoryRepository.
func (repository *CategoryRepositoryMock) Register(tobeRegisters []*model.CategoryMtmWarehouse) error {
	args := repository.Mock.Called(tobeRegisters)

	if args.Get(0) != nil {
		return args.Error(0)
	} else {
		return nil
	}
}

// Unregister implements CategoryRepository.
func (repository *CategoryRepositoryMock) Unregister(toUnregister *model.CategoryMtmWarehouse) error {
	args := repository.Mock.Called(toUnregister)

	if args.Get(0) != nil {
		return args.Error(0)
	} else {
		return nil
	}
}

// Update implements CategoryRepository.
func (repository *CategoryRepositoryMock) Update(tenantId int, categoryId int, tobeChangeCategoryName string) (*model.Category, error) {
	args := repository.Mock.Called(tenantId, categoryId, tobeChangeCategoryName)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Category), nil
}
