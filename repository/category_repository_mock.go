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
func (c *CategoryRepositoryMock) Delete(*model.Category) error {
	panic("unimplemented")
}

// Get implements CategoryRepository.
func (c *CategoryRepositoryMock) Get(tenantId int, page int, limit int) ([]*model.Category, int, error) {
	panic("unimplemented")
}

// GetCategoryWithItems implements CategoryRepository.
func (c *CategoryRepositoryMock) GetCategoryWithItems(tenantId int, page int, limit int, doCount bool) ([]*model.CategoryWithItem, int, error) {
	panic("unimplemented")
}

// GetItemsByCategoryId implements CategoryRepository.
func (c *CategoryRepositoryMock) GetItemsByCategoryId(tenantId int, categoryId int, limit int, page int) ([]*model.CategoryWithItem, error) {
	panic("unimplemented")
}

// Register implements CategoryRepository.
func (c *CategoryRepositoryMock) Register(tobeRegisters []*model.CategoryMtmWarehouse) error {
	panic("unimplemented")
}

// Unregister implements CategoryRepository.
func (c *CategoryRepositoryMock) Unregister(toUnregister *model.CategoryMtmWarehouse) error {
	panic("unimplemented")
}

// Update implements CategoryRepository.
func (c *CategoryRepositoryMock) Update(tenantId int, categoryId int, tobeChangeCategoryName string) (*model.Category, error) {
	panic("unimplemented")
}
